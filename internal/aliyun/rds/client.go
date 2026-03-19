package rds

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"

	"aliyun-ip-manager/internal/aliyun"
	"aliyun-ip-manager/internal/logger"
)

type Client struct {
	baseClient *aliyun.Client
}

type DBInstance struct {
	DBInstanceId          string
	DBInstanceStatus      string
	Engine                string
	EngineVersion         string
	DBInstanceNetType     string
	DBInstanceType        string
	InstanceNetworkType   string
	DBInstanceDescription string
}

type SecurityIPGroup struct {
	SecurityIPListName string
	SecurityIPList     string
}

func NewClient(baseClient *aliyun.Client) *Client {
	return &Client{
		baseClient: baseClient,
	}
}

func (c *Client) ListInstances(ctx context.Context) ([]DBInstance, error) {
	request := rds.CreateDescribeDBInstancesRequest()
	request.Scheme = "https"
	request.PageSize = "100"
	request.PageNumber = "1"

	response := rds.CreateDescribeDBInstancesResponse()
	if err := c.baseClient.DoRequest(ctx, request, response); err != nil {
		logger.Error("failed to list rds instances", "error", err)
		return nil, fmt.Errorf("failed to list rds instances: %w", err)
	}

	instances := make([]DBInstance, 0, len(response.Items.DBInstance))
	for _, instance := range response.Items.DBInstance {
		instances = append(instances, DBInstance{
			DBInstanceId:          instance.DBInstanceId,
			DBInstanceStatus:      instance.DBInstanceStatus,
			Engine:                instance.Engine,
			EngineVersion:         instance.EngineVersion,
			DBInstanceNetType:     instance.DBInstanceNetType,
			DBInstanceType:        instance.DBInstanceType,
			InstanceNetworkType:   instance.InstanceNetworkType,
			DBInstanceDescription: instance.DBInstanceDescription,
		})
	}

	logger.Info("list rds instances success", "count", len(instances))
	return instances, nil
}

func (c *Client) GetSecurityIPs(ctx context.Context, instanceId string) ([]SecurityIPGroup, error) {
	request := rds.CreateDescribeDBInstanceIPArrayListRequest()
	request.Scheme = "https"
	request.DBInstanceId = instanceId

	response := rds.CreateDescribeDBInstanceIPArrayListResponse()
	if err := c.baseClient.DoRequest(ctx, request, response); err != nil {
		logger.Error("failed to get security ips for rds instance", "error", err, "instance_id", instanceId)
		return nil, fmt.Errorf("failed to get security ips for rds instance %s: %w", instanceId, err)
	}

	groups := make([]SecurityIPGroup, 0, len(response.Items.DBInstanceIPArray))
	for _, group := range response.Items.DBInstanceIPArray {
		groups = append(groups, SecurityIPGroup{
			SecurityIPListName: group.DBInstanceIPArrayName,
			SecurityIPList:     group.SecurityIPList,
		})
	}

	logger.Info("get security ips success", "instance_id", instanceId, "count", len(groups))
	return groups, nil
}

func (c *Client) AddIPToWhitelist(ctx context.Context, instanceId, ipListName, newIP string) error {
	ips, err := c.GetSecurityIPs(ctx, instanceId)
	if err != nil {
		return err
	}

	var targetGroup *SecurityIPGroup
	for _, group := range ips {
		if group.SecurityIPListName == ipListName {
			targetGroup = &group
			break
		}
	}

	if targetGroup == nil {
		return fmt.Errorf("security ip group not found: %s", ipListName)
	}

	updatedIPs := targetGroup.SecurityIPList + "," + newIP

	request := rds.CreateModifySecurityIpsRequest()
	request.Scheme = "https"
	request.DBInstanceId = instanceId
	request.SecurityIps = updatedIPs
	request.DBInstanceIPArrayName = ipListName

	response := rds.CreateModifySecurityIpsResponse()
	if err := c.baseClient.DoRequest(ctx, request, response); err != nil {
		return fmt.Errorf("failed to add ip %s to whitelist for rds instance %s: %w", newIP, instanceId, err)
	}

	logger.Info("add ip to rds whitelist success", "instance_id", instanceId, "ip_list_name", ipListName, "new_ip", newIP)
	return nil
}
