package alb

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alb"

	"aliyun-ip-manager/internal/aliyun"
	"aliyun-ip-manager/internal/logger"
)

type Client struct {
	aliyunClient *aliyun.Client
}

func NewClient(aliyunClient *aliyun.Client) *Client {
	return &Client{
		aliyunClient: aliyunClient,
	}
}

func (c *Client) ListAclPolicies(ctx context.Context, req *ListAclPoliciesRequest) (*ListAclPoliciesResponse, error) {
	logger.Info("listing alb acl policies",
		"page_number", req.PageNumber,
		"page_size", req.PageSize)

	request := alb.CreateListAclsRequest()
	request.Scheme = "https"

	if req.PageSize > 0 {
		request.MaxResults = requests.NewInteger(req.PageSize)
	}

	response := alb.CreateListAclsResponse()
	if err := c.aliyunClient.DoRequest(ctx, request, response); err != nil {
		logger.Error("failed to list alb policies", "error", err)
		return nil, fmt.Errorf("failed to list alb policies: %w", err)
	}

	aclPolicies := make([]AclPolicy, 0, len(response.Acls))
	for _, acl := range response.Acls {
		aclPolicies = append(aclPolicies, AclPolicy{
			AclId:            acl.AclId,
			AclName:          acl.AclName,
			AddressIPVersion: acl.AddressIPVersion,
			ResourceGroupId:  acl.ResourceGroupId,
		})
	}

	result := &ListAclPoliciesResponse{
		RequestId:   response.RequestId,
		TotalCount:  response.TotalCount,
		PageNumber:  req.PageNumber,
		PageSize:    req.PageSize,
		AclPolicies: aclPolicies,
		NextToken:   response.NextToken,
	}

	logger.Info("alb acl policies listed successfully",
		"total_count", result.TotalCount,
		"acl_policy_count", len(result.AclPolicies))

	return result, nil
}

func (c *Client) ListAclEntries(ctx context.Context, req *ListAclEntriesRequest) (*ListAclEntriesResponse, error) {
	logger.Info("listing alb acl entries",
		"acl_id", req.AclId,
		"page_number", req.PageNumber,
		"page_size", req.PageSize)

	if req.AclId == "" {
		return nil, fmt.Errorf("acl id is required")
	}

	request := alb.CreateListAclEntriesRequest()
	request.Scheme = "https"
	request.AclId = req.AclId

	if req.PageSize > 0 {
		request.MaxResults = requests.NewInteger(req.PageSize)
	}

	response := alb.CreateListAclEntriesResponse()
	if err := c.aliyunClient.DoRequest(ctx, request, response); err != nil {
		logger.Error("failed to list acl entries", "error", err)
		return nil, fmt.Errorf("failed to list acl entries: %w", err)
	}

	aclEntries := make([]AclEntry, 0, len(response.AclEntries))
	for _, entry := range response.AclEntries {
		aclEntries = append(aclEntries, AclEntry{
			Entry:            entry.Entry,
			EntryDescription: entry.Description,
		})
	}

	result := &ListAclEntriesResponse{
		RequestId:  response.RequestId,
		TotalCount: response.TotalCount,
		PageNumber: req.PageNumber,
		PageSize:   req.PageSize,
		AclEntries: aclEntries,
		NextToken:  response.NextToken,
	}

	logger.Info("alb acl entries listed successfully",
		"acl_id", req.AclId,
		"total_count", result.TotalCount,
		"acl_entry_count", len(result.AclEntries))

	return result, nil
}

func (c *Client) AddAclEntries(ctx context.Context, req *AddAclEntriesRequest) (*AddAclEntriesResponse, error) {
	logger.Info("adding entries to alb acl",
		"acl_id", req.AclId,
		"entry_count", len(req.Entries))

	if req.AclId == "" {
		return nil, fmt.Errorf("acl id is required")
	}
	if len(req.Entries) == 0 {
		return nil, fmt.Errorf("entries are required")
	}

	request := alb.CreateAddEntriesToAclRequest()
	request.Scheme = "https"
	request.AclId = req.AclId

	entries := make([]alb.AddEntriesToAclAclEntries, 0, len(req.Entries))
	for _, entry := range req.Entries {
		entries = append(entries, alb.AddEntriesToAclAclEntries{
			Entry:       entry.Entry,
			Description: entry.EntryDescription,
		})
	}
	request.AclEntries = &entries

	response := alb.CreateAddEntriesToAclResponse()
	if err := c.aliyunClient.DoRequest(ctx, request, response); err != nil {
		logger.Error("failed to add entries to acl", "error", err)
		return nil, fmt.Errorf("failed to add entries to acl: %w", err)
	}

	result := &AddAclEntriesResponse{
		RequestId: response.RequestId,
	}

	logger.Info("entries added to alb acl successfully",
		"acl_id", req.AclId,
		"request_id", response.RequestId)

	return result, nil
}
