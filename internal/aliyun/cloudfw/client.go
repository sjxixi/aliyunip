package cloudfw

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cloudfw"

	"aliyun-ip-manager/internal/aliyun"
	"aliyun-ip-manager/internal/logger"
)

type Client struct {
	baseClient *aliyun.Client
}

func NewClient(baseClient *aliyun.Client) *Client {
	return &Client{
		baseClient: baseClient,
	}
}

func (c *Client) ListAddressBooks(ctx context.Context, req *ListAddressBooksRequest) (*ListAddressBooksResponse, error) {
	logger.Info("listing address books",
		"page_number", req.PageNumber,
		"page_size", req.PageSize,
		"original_region", c.baseClient.GetRegion())

	request := cloudfw.CreateDescribeAddressBookRequest()
	request.Scheme = "https"
	request.Domain = "cloudfw.aliyuncs.com"

	if req.PageNumber > 0 {
		request.CurrentPage = fmt.Sprintf("%d", req.PageNumber)
	}
	if req.PageSize > 0 {
		request.PageSize = fmt.Sprintf("%d", req.PageSize)
	}

	logger.Info("cloudfw request created",
		"current_page", request.CurrentPage,
		"page_size", request.PageSize,
		"domain", request.Domain)

	response := cloudfw.CreateDescribeAddressBookResponse()
	logger.Info("calling cloudfw API...")
	if err := c.baseClient.DoRequest(ctx, request, response); err != nil {
		logger.Error("failed to list address books", "error", err)
		return nil, fmt.Errorf("failed to list address books: %w", err)
	}

	logger.Info("cloudfw API call completed",
		"request_id", response.RequestId,
		"page_no", response.PageNo,
		"total_count_raw", response.TotalCount)

	totalCount := 0
	if response.TotalCount != "" {
		if tc, err := strconv.Atoi(response.TotalCount); err == nil {
			totalCount = tc
		} else {
			logger.Warn("failed to parse total count",
				"total_count_raw", response.TotalCount,
				"error", err)
		}
	}

	logger.Info("cloudfw response received",
		"request_id", response.RequestId,
		"total_count", totalCount,
		"acls_count", len(response.Acls))

	addressBooks := make([]AddressBook, 0, len(response.Acls))
	for i, acl := range response.Acls {
		logger.Debug("processing acl",
			"index", i,
			"group_name", acl.GroupName,
			"group_uuid", acl.GroupUuid,
			"group_type", acl.GroupType)
		addressBooks = append(addressBooks, AddressBook{
			AddressBookName: acl.GroupName,
			AddressBookId:   acl.GroupUuid,
			Description:     acl.Description,
			AddressList:     acl.AddressList,
			AutoAddTagEcs:   acl.AutoAddTagEcs == 1,
			GroupType:       acl.GroupType,
		})

		logger.Debug("found address book",
			"name", acl.GroupName,
			"id", acl.GroupUuid)
	}

	result := &ListAddressBooksResponse{
		RequestId:    response.RequestId,
		TotalCount:   totalCount,
		PageNumber:   req.PageNumber,
		PageSize:     req.PageSize,
		AddressBooks: addressBooks,
	}

	logger.Info("address books listed successfully",
		"request_id", response.RequestId,
		"total_count", result.TotalCount,
		"address_book_count", len(result.AddressBooks))

	return result, nil
}

func (c *Client) AddIpToAddressBook(ctx context.Context, req *AddIpToAddressBookRequest) (*AddIpToAddressBookResponse, error) {
	logger.Info("adding ip to address book",
		"address_book_name", req.AddressBookName,
		"address_book_id", req.AddressBookId,
		"ip_count", len(req.IpList))

	if req.AddressBookName == "" && req.AddressBookId == "" {
		return nil, fmt.Errorf("either address book name or id is required")
	}
	if len(req.IpList) == 0 {
		return nil, fmt.Errorf("ip list is required")
	}

	logger.Info("looking up address book",
		"name", req.AddressBookName,
		"id", req.AddressBookId)

	listReq := &ListAddressBooksRequest{PageNumber: 1, PageSize: 100}
	listResp, err := c.ListAddressBooks(ctx, listReq)
	if err != nil {
		logger.Error("failed to list address books", "error", err)
		return nil, fmt.Errorf("failed to list address books: %w", err)
	}

	var targetBook *AddressBook
	for _, book := range listResp.AddressBooks {
		if req.AddressBookId != "" && book.AddressBookId == req.AddressBookId {
			targetBook = &book
			break
		}
		if req.AddressBookName != "" && book.AddressBookName == req.AddressBookName {
			targetBook = &book
			break
		}
	}

	if targetBook == nil {
		logger.Error("address book not found",
			"name", req.AddressBookName,
			"id", req.AddressBookId)
		return nil, fmt.Errorf("address book not found: name=%s, id=%s", req.AddressBookName, req.AddressBookId)
	}

	logger.Info("found address book",
		"name", targetBook.AddressBookName,
		"uuid", targetBook.AddressBookId,
		"group_type", targetBook.GroupType,
		"existing_ips", len(targetBook.AddressList))

	formatIP := func(ip string) string {
		if ip == "" {
			return ""
		}
		if !strings.Contains(ip, "/") {
			return ip + "/32"
		}
		return ip
	}

	ipSet := make(map[string]bool)
	for _, ip := range targetBook.AddressList {
		if ip != "" {
			ipSet[ip] = true
		}
	}
	for _, ip := range req.IpList {
		if ip != "" {
			formattedIP := formatIP(ip)
			ipSet[formattedIP] = true
		}
	}

	ipList := make([]string, 0, len(ipSet))
	for ip := range ipSet {
		ipList = append(ipList, ip)
	}

	addressListStr := ""
	for i, ip := range ipList {
		if i > 0 {
			addressListStr += ","
		}
		addressListStr += ip
	}

	logger.Info("modifying address book",
		"group_uuid", targetBook.AddressBookId,
		"group_name", targetBook.AddressBookName,
		"total_ips", len(ipList),
		"address_list", addressListStr)

	request := cloudfw.CreateModifyAddressBookRequest()
	request.Scheme = "https"
	request.Domain = "cloudfw.aliyuncs.com"
	request.GroupUuid = targetBook.AddressBookId
	request.GroupName = targetBook.AddressBookName
	request.Description = targetBook.Description
	request.AddressList = addressListStr

	if targetBook.AutoAddTagEcs {
		request.AutoAddTagEcs = "1"
	} else {
		request.AutoAddTagEcs = "0"
	}

	response := cloudfw.CreateModifyAddressBookResponse()
	if err := c.baseClient.DoRequest(ctx, request, response); err != nil {
		logger.Error("failed to modify address book", "error", err)
		return nil, fmt.Errorf("failed to modify address book: %w", err)
	}

	logger.Info("ip added to address book successfully",
		"address_book_name", req.AddressBookName,
		"request_id", response.RequestId)

	return &AddIpToAddressBookResponse{
		RequestId: response.RequestId,
	}, nil
}
