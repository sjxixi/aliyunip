package redis

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/r_kvstore"

	"aliyun-ip-manager/internal/aliyun"
	"aliyun-ip-manager/internal/logger"
)

type Client struct {
	baseClient *aliyun.Client
}

type DBInstance struct {
	InstanceId     string
	InstanceStatus string
	Engine         string
	EngineVersion  string
	InstanceName   string
}

type SecurityIPGroup struct {
	SecurityIPGroupName string
	SecurityIPList      string
}

func NewClient(baseClient *aliyun.Client) *Client {
	return &Client{
		baseClient: baseClient,
	}
}

func (c *Client) ListInstances(ctx context.Context) ([]DBInstance, error) {
	request := r_kvstore.CreateDescribeInstancesRequest()
	request.Scheme = "https"
	request.PageSize = "100"
	request.PageNumber = "1"

	response := r_kvstore.CreateDescribeInstancesResponse()
	if err := c.baseClient.DoRequest(ctx, request, response); err != nil {
		logger.Error("failed to list redis instances", "error", err)
		return nil, fmt.Errorf("failed to list redis instances: %w", err)
	}

	instances := make([]DBInstance, 0, len(response.Instances.KVStoreInstance))
	for _, instance := range response.Instances.KVStoreInstance {
		instances = append(instances, DBInstance{
			InstanceId:     instance.InstanceId,
			InstanceStatus: instance.InstanceStatus,
			Engine:         "Redis",
			EngineVersion:  instance.EngineVersion,
			InstanceName:   instance.InstanceName,
		})
	}

	logger.Info("list redis instances success", "count", len(instances))
	return instances, nil
}

func (c *Client) GetSecurityIPs(ctx context.Context, instanceId string) ([]SecurityIPGroup, error) {
	request := r_kvstore.CreateDescribeSecurityIpsRequest()
	request.Scheme = "https"
	request.InstanceId = instanceId

	response := r_kvstore.CreateDescribeSecurityIpsResponse()
	if err := c.baseClient.DoRequest(ctx, request, response); err != nil {
		logger.Error("failed to get security ips for redis instance", "error", err, "instance_id", instanceId)
		return nil, fmt.Errorf("failed to get security ips for redis instance %s: %w", instanceId, err)
	}

	groups := make([]SecurityIPGroup, 0, len(response.SecurityIpGroups.SecurityIpGroup))
	for _, group := range response.SecurityIpGroups.SecurityIpGroup {
		groups = append(groups, SecurityIPGroup{
			SecurityIPGroupName: group.SecurityIpGroupName,
			SecurityIPList:      group.SecurityIpList,
		})
	}

	logger.Info("get security ips success", "instance_id", instanceId, "count", len(groups))
	return groups, nil
}

func (c *Client) AddIPToWhitelist(ctx context.Context, instanceId, ipGroupName, newIP string) error {
	ips, err := c.GetSecurityIPs(ctx, instanceId)
	if err != nil {
		return err
	}

	var targetGroup *SecurityIPGroup
	for _, group := range ips {
		if group.SecurityIPGroupName == ipGroupName {
			targetGroup = &group
			break
		}
	}

	if targetGroup == nil {
		return fmt.Errorf("security ip group not found: %s", ipGroupName)
	}

	updatedIPs := targetGroup.SecurityIPList + "," + newIP

	request := r_kvstore.CreateModifySecurityIpsRequest()
	request.Scheme = "https"
	request.InstanceId = instanceId
	request.SecurityIps = updatedIPs
	request.SecurityIpGroupName = ipGroupName

	response := r_kvstore.CreateModifySecurityIpsResponse()
	if err := c.baseClient.DoRequest(ctx, request, response); err != nil {
		return fmt.Errorf("failed to add ip %s to whitelist for redis instance %s: %w", newIP, instanceId, err)
	}

	logger.Info("add ip to redis whitelist success", "instance_id", instanceId, "ip_group_name", ipGroupName, "new_ip", newIP)
	return nil
}
