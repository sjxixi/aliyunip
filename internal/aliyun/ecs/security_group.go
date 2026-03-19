package ecs

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"

	"aliyun-ip-manager/internal/aliyun"
	"aliyun-ip-manager/internal/logger"
)

type SecurityGroup struct {
	SecurityGroupId   string
	SecurityGroupName string
	Description       string
	VpcId             string
}

type SecurityGroupManager struct {
	client *aliyun.Client
}

type IngressRule struct {
	PortRange       string
	Protocol        string
	SourceCidrIp    string
	Description     string
	Policy          string
	Priority        int
	SecurityGroupId string
}

func NewSecurityGroupManager(client *aliyun.Client) *SecurityGroupManager {
	return &SecurityGroupManager{
		client: client,
	}
}

func (m *SecurityGroupManager) ListSecurityGroups(ctx context.Context) ([]SecurityGroup, error) {
	logger.Info("listing security groups",
		"region", m.client.GetRegion())

	var groups []SecurityGroup
	pageNumber := 1
	pageSize := 100

	for {
		request := ecs.CreateDescribeSecurityGroupsRequest()
		request.Scheme = "https"
		request.PageNumber = requests.NewInteger(pageNumber)
		request.PageSize = requests.NewInteger(pageSize)
		request.ServiceManaged = requests.NewBoolean(false)

		response := ecs.CreateDescribeSecurityGroupsResponse()
		err := m.client.DoRequest(ctx, request, response)
		if err != nil {
			return nil, fmt.Errorf("failed to list security groups (page %d): %w", pageNumber, err)
		}

		for _, sg := range response.SecurityGroups.SecurityGroup {
			groups = append(groups, SecurityGroup{
				SecurityGroupId:   sg.SecurityGroupId,
				SecurityGroupName: sg.SecurityGroupName,
				Description:       sg.Description,
				VpcId:             sg.VpcId,
			})
		}

		logger.Info("loaded security groups page",
			"page", pageNumber,
			"page_count", len(response.SecurityGroups.SecurityGroup),
			"total_count", response.TotalCount)

		if len(groups) >= response.TotalCount {
			break
		}

		pageNumber++
	}

	logger.Info("successfully listed security groups",
		"count", len(groups),
		"region", m.client.GetRegion())

	return groups, nil
}

func (m *SecurityGroupManager) AddIngressRule(ctx context.Context, rule IngressRule) error {
	logger.Info("adding ingress rule",
		"security_group_id", rule.SecurityGroupId,
		"port_range", rule.PortRange,
		"protocol", rule.Protocol,
		"source_cidr", rule.SourceCidrIp,
		"region", m.client.GetRegion())

	request := ecs.CreateAuthorizeSecurityGroupRequest()
	request.Scheme = "https"
	request.SecurityGroupId = rule.SecurityGroupId
	request.PortRange = rule.PortRange
	request.IpProtocol = rule.Protocol
	request.SourceCidrIp = rule.SourceCidrIp
	request.Description = rule.Description

	if rule.Policy != "" {
		request.Policy = rule.Policy
	} else {
		request.Policy = "accept"
	}

	if rule.Priority > 0 {
		request.Priority = strconv.Itoa(rule.Priority)
	} else {
		request.Priority = "1"
	}

	response := ecs.CreateAuthorizeSecurityGroupResponse()
	err := m.client.DoRequest(ctx, request, response)
	if err != nil {
		logger.Error("failed to add ingress rule",
			"security_group_id", rule.SecurityGroupId,
			"port_range", rule.PortRange,
			"protocol", rule.Protocol,
			"error", err)
		return fmt.Errorf("failed to add ingress rule: %w", err)
	}

	logger.Info("successfully added ingress rule",
		"security_group_id", rule.SecurityGroupId,
		"port_range", rule.PortRange,
		"protocol", rule.Protocol,
		"request_id", response.RequestId)

	return nil
}
