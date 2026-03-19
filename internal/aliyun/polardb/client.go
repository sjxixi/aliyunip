package polardb

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/polardb"

	"aliyun-ip-manager/internal/aliyun"
	"aliyun-ip-manager/internal/logger"
)

type Client struct {
	baseClient *aliyun.Client
}

type DBCluster struct {
	DBClusterId          string
	DBClusterStatus      string
	Engine               string
	DBVersion            string
	DBClusterNetworkType string
	DBType               string
	DBClusterDescription string
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

func (c *Client) ListInstances(ctx context.Context) ([]DBCluster, error) {
	request := polardb.CreateDescribeDBClustersRequest()
	request.Scheme = "https"
	request.PageSize = "100"
	request.PageNumber = "1"

	response := polardb.CreateDescribeDBClustersResponse()
	if err := c.baseClient.DoRequest(ctx, request, response); err != nil {
		logger.Error("failed to list polardb clusters", "error", err)
		return nil, fmt.Errorf("failed to list polardb clusters: %w", err)
	}

	clusters := make([]DBCluster, 0, len(response.Items.DBCluster))
	for _, cluster := range response.Items.DBCluster {
		clusters = append(clusters, DBCluster{
			DBClusterId:          cluster.DBClusterId,
			DBClusterStatus:      cluster.DBClusterStatus,
			Engine:               cluster.Engine,
			DBVersion:            cluster.DBVersion,
			DBClusterNetworkType: cluster.DBClusterNetworkType,
			DBType:               cluster.DBType,
			DBClusterDescription: cluster.DBClusterDescription,
		})
	}

	logger.Info("list polardb clusters success", "count", len(clusters))
	return clusters, nil
}

func (c *Client) GetSecurityIPs(ctx context.Context, dbClusterId string) ([]SecurityIPGroup, error) {
	request := polardb.CreateDescribeDBClusterAccessWhitelistRequest()
	request.Scheme = "https"
	request.DBClusterId = dbClusterId

	response := polardb.CreateDescribeDBClusterAccessWhitelistResponse()
	if err := c.baseClient.DoRequest(ctx, request, response); err != nil {
		logger.Error("failed to get security ips for polardb cluster", "error", err, "db_cluster_id", dbClusterId)
		return nil, fmt.Errorf("failed to get security ips for polardb cluster %s: %w", dbClusterId, err)
	}

	groups := make([]SecurityIPGroup, 0, len(response.Items.DBClusterIPArray))
	for _, group := range response.Items.DBClusterIPArray {
		groups = append(groups, SecurityIPGroup{
			SecurityIPListName: group.DBClusterIPArrayName,
			SecurityIPList:     group.SecurityIps,
		})
	}

	logger.Info("get security ips success", "db_cluster_id", dbClusterId, "count", len(groups))
	return groups, nil
}

func (c *Client) AddIPToWhitelist(ctx context.Context, dbClusterId, ipListName, newIP string) error {
	ips, err := c.GetSecurityIPs(ctx, dbClusterId)
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

	request := polardb.CreateModifyDBClusterAccessWhitelistRequest()
	request.Scheme = "https"
	request.DBClusterId = dbClusterId
	request.SecurityIps = updatedIPs
	request.DBClusterIPArrayName = ipListName

	response := polardb.CreateModifyDBClusterAccessWhitelistResponse()
	if err := c.baseClient.DoRequest(ctx, request, response); err != nil {
		return fmt.Errorf("failed to add ip %s to whitelist for polardb cluster %s: %w", newIP, dbClusterId, err)
	}

	logger.Info("add ip to polardb whitelist success", "db_cluster_id", dbClusterId, "ip_list_name", ipListName, "new_ip", newIP)
	return nil
}
