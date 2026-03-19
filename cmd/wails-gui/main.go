package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"aliyun-ip-manager/internal/aliyun"
	"aliyun-ip-manager/internal/aliyun/alb"
	"aliyun-ip-manager/internal/aliyun/cloudfw"
	"aliyun-ip-manager/internal/aliyun/ecs"
	"aliyun-ip-manager/internal/aliyun/polardb"
	"aliyun-ip-manager/internal/aliyun/rds"
	"aliyun-ip-manager/internal/aliyun/redis"
	"aliyun-ip-manager/internal/config"
	"aliyun-ip-manager/internal/logger"
	"aliyun-ip-manager/pkg/validator"
)

//go:embed all:frontend/dist
var assets embed.FS

type App struct {
	ctx    context.Context
	client *aliyun.Client
}

type Config struct {
	AccessKeyID     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	Region          string `json:"region"`
}

type AclPolicy struct {
	AclId            string `json:"aclId"`
	AclName          string `json:"aclName"`
	AddressIPVersion string `json:"addressIPVersion"`
}

type SecurityGroup struct {
	SecurityGroupId   string `json:"securityGroupId"`
	SecurityGroupName string `json:"securityGroupName"`
	Description       string `json:"description"`
}

type RDSInstance struct {
	DBInstanceId          string   `json:"dbInstanceId"`
	Engine                string   `json:"engine"`
	SecurityIPGroups      []string `json:"securityIpGroups"`
	DBInstanceDescription string   `json:"dbInstanceDescription,omitempty"`
}

type PolarDBCluster struct {
	DBClusterId          string   `json:"dbClusterId"`
	Engine               string   `json:"engine"`
	SecurityIPGroups     []string `json:"securityIpGroups"`
	DBClusterDescription string   `json:"dbClusterDescription,omitempty"`
}

type RedisInstance struct {
	InstanceId       string   `json:"instanceId"`
	Engine           string   `json:"engine"`
	SecurityIPGroups []string `json:"securityIpGroups"`
	InstanceName     string   `json:"instanceName,omitempty"`
}

type AddressBook struct {
	AddressBookName string `json:"addressBookName"`
	AddressBookId   string `json:"addressBookId"`
}

type Resources struct {
	AlbPolicies     []AclPolicy      `json:"albPolicies"`
	SecurityGroups  []SecurityGroup  `json:"securityGroups"`
	RdsInstances    []RDSInstance    `json:"rdsInstances"`
	PolarDBClusters []PolarDBCluster `json:"polarDBClusters"`
	RedisInstances  []RedisInstance  `json:"redisInstances"`
	AddressBooks    []AddressBook    `json:"addressBooks"`
}

type SelectedResource struct {
	Type            string `json:"type"`
	Id              string `json:"id"`
	Name            string `json:"name"`
	SecurityIpGroup string `json:"securityIpGroup,omitempty"`
}

type ExecutionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type ProgressUpdate struct {
	Step      int    `json:"step"`
	Total     int    `json:"total"`
	Message   string `json:"message"`
	Completed bool   `json:"completed"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	logger.Init("")

	logger.Info("app started")
}

func (a *App) ValidateCredentials(accessKeyId, accessKeySecret, region string) *ExecutionResult {
	result := &ExecutionResult{}

	if accessKeyId == "" {
		result.Success = false
		result.Message = "AccessKey ID 不能为空"
		return result
	}
	if accessKeySecret == "" {
		result.Success = false
		result.Message = "AccessKey Secret 不能为空"
		return result
	}
	if region == "" {
		result.Success = false
		result.Message = "Region 不能为空"
		return result
	}

	cfg := config.New()
	cfg.AccessKeyID = accessKeyId
	cfg.AccessKeySecret = accessKeySecret
	cfg.Region = region

	client, err := aliyun.NewClient(cfg)
	if err != nil {
		result.Success = false
		result.Message = "初始化客户端失败"
		result.Error = err.Error()
		return result
	}

	a.client = client

	if err := config.Save(cfg); err != nil {
		logger.Warn("failed to save config", "error", err)
	}

	result.Success = true
	result.Message = "验证成功"
	return result
}

func (a *App) LoadSavedConfig() *Config {
	cfg, err := config.Load()
	if err != nil {
		return nil
	}
	return &Config{
		AccessKeyID:     cfg.AccessKeyID,
		AccessKeySecret: cfg.AccessKeySecret,
		Region:          cfg.Region,
	}
}

func (a *App) LoadResources() *Resources {
	result := &Resources{}

	if a.client == nil {
		return result
	}

	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(6)

	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic in alb list", "panic", r)
			}
		}()
		albClient := alb.NewClient(a.client)
		req := &alb.ListAclPoliciesRequest{PageNumber: 1, PageSize: 100}
		resp, err := albClient.ListAclPolicies(ctx, req)
		if err != nil {
			logger.Error("failed to list alb policies", "error", err)
		} else {
			logger.Info("alb policies loaded", "count", len(resp.AclPolicies))
			mu.Lock()
			for _, policy := range resp.AclPolicies {
				result.AlbPolicies = append(result.AlbPolicies, AclPolicy{
					AclId:            policy.AclId,
					AclName:          policy.AclName,
					AddressIPVersion: policy.AddressIPVersion,
				})
			}
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic in ecs list", "panic", r)
			}
		}()
		sgManager := ecs.NewSecurityGroupManager(a.client)
		groups, err := sgManager.ListSecurityGroups(ctx)
		if err != nil {
			logger.Error("failed to list security groups", "error", err)
		} else {
			logger.Info("security groups loaded", "count", len(groups))
			mu.Lock()
			for _, group := range groups {
				result.SecurityGroups = append(result.SecurityGroups, SecurityGroup{
					SecurityGroupId:   group.SecurityGroupId,
					SecurityGroupName: group.SecurityGroupName,
					Description:       group.Description,
				})
			}
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic in rds list", "panic", r)
			}
		}()
		rdsClient := rds.NewClient(a.client)
		instances, err := rdsClient.ListInstances(ctx)
		if err != nil {
			logger.Error("failed to list rds instances", "error", err)
		} else {
			logger.Info("rds instances loaded", "count", len(instances))
			mu.Lock()
			for _, instance := range instances {
				var groups []string
				ips, ipErr := rdsClient.GetSecurityIPs(ctx, instance.DBInstanceId)
				if ipErr == nil {
					for _, g := range ips {
						groups = append(groups, g.SecurityIPListName)
					}
				}
				result.RdsInstances = append(result.RdsInstances, RDSInstance{
					DBInstanceId:          instance.DBInstanceId,
					Engine:                instance.Engine,
					SecurityIPGroups:      groups,
					DBInstanceDescription: instance.DBInstanceDescription,
				})
			}
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic in polardb list", "panic", r)
			}
		}()
		polardbClient := polardb.NewClient(a.client)
		clusters, err := polardbClient.ListInstances(ctx)
		if err != nil {
			logger.Error("failed to list polardb clusters", "error", err)
		} else {
			logger.Info("polardb clusters loaded", "count", len(clusters))
			mu.Lock()
			for _, cluster := range clusters {
				var groups []string
				ips, ipErr := polardbClient.GetSecurityIPs(ctx, cluster.DBClusterId)
				if ipErr == nil {
					for _, g := range ips {
						groups = append(groups, g.SecurityIPListName)
					}
				}
				result.PolarDBClusters = append(result.PolarDBClusters, PolarDBCluster{
					DBClusterId:          cluster.DBClusterId,
					Engine:               cluster.Engine,
					SecurityIPGroups:     groups,
					DBClusterDescription: cluster.DBClusterDescription,
				})
			}
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic in redis list", "panic", r)
			}
		}()
		redisClient := redis.NewClient(a.client)
		instances, err := redisClient.ListInstances(ctx)
		if err != nil {
			logger.Error("failed to list redis instances", "error", err)
		} else {
			logger.Info("redis instances loaded", "count", len(instances))
			mu.Lock()
			for _, instance := range instances {
				var groups []string
				ips, ipErr := redisClient.GetSecurityIPs(ctx, instance.InstanceId)
				if ipErr == nil {
					for _, g := range ips {
						groups = append(groups, g.SecurityIPGroupName)
					}
				}
				result.RedisInstances = append(result.RedisInstances, RedisInstance{
					InstanceId:       instance.InstanceId,
					Engine:           instance.Engine,
					SecurityIPGroups: groups,
					InstanceName:     instance.InstanceName,
				})
			}
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic in cloudfw list", "panic", r)
			}
		}()
		logger.Info("starting cloudfw address books load")
		cfClient := cloudfw.NewClient(a.client)
		req := &cloudfw.ListAddressBooksRequest{PageNumber: 1, PageSize: 100}
		resp, err := cfClient.ListAddressBooks(ctx, req)
		if err != nil {
			logger.Error("failed to list address books", "error", err)
		} else {
			logger.Info("address books response received",
				"total_count", resp.TotalCount,
				"address_book_count", len(resp.AddressBooks))
			mu.Lock()
			for _, ab := range resp.AddressBooks {
				logger.Debug("adding address book to result",
					"name", ab.AddressBookName,
					"id", ab.AddressBookId)
				result.AddressBooks = append(result.AddressBooks, AddressBook{
					AddressBookName: ab.AddressBookName,
					AddressBookId:   ab.AddressBookId,
				})
			}
			logger.Info("address books added to result", "count", len(result.AddressBooks))
			mu.Unlock()
		}
	}()

	wg.Wait()

	logger.Info("resources loaded",
		"alb_count", len(result.AlbPolicies),
		"sg_count", len(result.SecurityGroups),
		"rds_count", len(result.RdsInstances),
		"polardb_count", len(result.PolarDBClusters),
		"redis_count", len(result.RedisInstances),
		"cloudfw_count", len(result.AddressBooks))

	return result
}

func (a *App) ValidateIP(ip string) *ExecutionResult {
	result := &ExecutionResult{}

	if ip == "" {
		result.Success = false
		result.Message = "IP 地址不能为空"
		return result
	}

	if err := validator.ValidateIPv4(ip); err != nil {
		if err := validator.ValidateCIDR(ip); err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("IP 地址格式无效: %v", err)
			return result
		}
	}

	result.Success = true
	result.Message = "IP 地址格式正确"
	return result
}

func (a *App) ValidatePort(portStr string) *ExecutionResult {
	result := &ExecutionResult{}

	if portStr == "" {
		result.Success = true
		result.Message = ""
		return result
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("端口号格式无效: %v", err)
		return result
	}

	if port < 1 || port > 65535 {
		result.Success = false
		result.Message = "端口号必须在 1-65535 之间"
		return result
	}

	result.Success = true
	result.Message = "端口号格式正确"
	return result
}

func (a *App) ExecuteConfig(ip string, port int, description string, resources []SelectedResource) []ExecutionResult {
	var results []ExecutionResult

	if a.client == nil {
		results = append(results, ExecutionResult{
			Success: false,
			Message: "客户端未初始化",
			Error:   "请先验证凭据",
		})
		return results
	}

	ctx, cancel := context.WithTimeout(a.ctx, 60*time.Second)
	defer cancel()

	for _, resource := range resources {
		var result ExecutionResult
		result.Message = fmt.Sprintf("正在处理 %s: %s", resource.Type, resource.Name)

		var err error
		switch resource.Type {
		case "alb":
			err = a.addToALB(ctx, resource.Id, ip, description)
		case "ecs":
			portRange := fmt.Sprintf("%d/%d", port, port)
			if port == 0 {
				portRange = "1/65535"
			}
			err = a.addToECS(ctx, resource.Id, portRange, "tcp", ip, "")
		case "cloudfw":
			err = a.addToCloudFW(ctx, resource.Id, ip)
		case "rds":
			groupName := "default"
			if resource.SecurityIpGroup != "" {
				groupName = resource.SecurityIpGroup
			}
			err = a.addToRDS(ctx, resource.Id, groupName, ip)
		case "polardb":
			groupName := "default"
			if resource.SecurityIpGroup != "" {
				groupName = resource.SecurityIpGroup
			}
			err = a.addToPolarDB(ctx, resource.Id, groupName, ip)
		case "redis":
			groupName := "default"
			if resource.SecurityIpGroup != "" {
				groupName = resource.SecurityIpGroup
			}
			err = a.addToRedis(ctx, resource.Id, groupName, ip)
		}

		if err != nil {
			result.Success = false
			result.Message = fmt.Sprintf("处理 %s 失败: %s", resource.Name, resource.Type)
			result.Error = err.Error()
			logger.Error("failed to execute config",
				"resource_type", resource.Type,
				"resource_id", resource.Id,
				"error", err)
		} else {
			result.Success = true
			result.Message = fmt.Sprintf("成功添加 IP 到 %s: %s", resource.Type, resource.Name)
			logger.Info("config executed successfully",
				"resource_type", resource.Type,
				"resource_id", resource.Id,
				"ip", ip)
		}

		results = append(results, result)
	}

	return results
}

func (a *App) addToALB(ctx context.Context, aclID, ip, description string) error {
	albClient := alb.NewClient(a.client)

	entryValue := ip
	if !strings.Contains(ip, "/") {
		entryValue = ip + "/32"
	}

	entries := []alb.AclEntry{{
		Entry:            entryValue,
		EntryDescription: description,
	}}
	req := &alb.AddAclEntriesRequest{
		AclId:   aclID,
		Entries: entries,
	}
	_, err := albClient.AddAclEntries(ctx, req)
	return err
}

func (a *App) addToECS(ctx context.Context, sgID, portRange, protocol, ip, description string) error {
	sgManager := ecs.NewSecurityGroupManager(a.client)
	rule := ecs.IngressRule{
		SecurityGroupId: sgID,
		PortRange:       portRange,
		Protocol:        protocol,
		SourceCidrIp:    ip,
		Description:     description,
	}
	return sgManager.AddIngressRule(ctx, rule)
}

func (a *App) addToCloudFW(ctx context.Context, bookId, ip string) error {
	cfClient := cloudfw.NewClient(a.client)
	req := &cloudfw.AddIpToAddressBookRequest{
		AddressBookId: bookId,
		IpList:        []string{ip},
	}
	_, err := cfClient.AddIpToAddressBook(ctx, req)
	return err
}

func (a *App) addToRDS(ctx context.Context, instanceID, listName, ip string) error {
	rdsClient := rds.NewClient(a.client)
	return rdsClient.AddIPToWhitelist(ctx, instanceID, listName, ip)
}

func (a *App) addToPolarDB(ctx context.Context, clusterID, listName, ip string) error {
	polardbClient := polardb.NewClient(a.client)
	return polardbClient.AddIPToWhitelist(ctx, clusterID, listName, ip)
}

func (a *App) addToRedis(ctx context.Context, instanceID, groupName, ip string) error {
	redisClient := redis.NewClient(a.client)
	return redisClient.AddIPToWhitelist(ctx, instanceID, groupName, ip)
}

func (a *App) GetPublicIP() *ExecutionResult {
	result := &ExecutionResult{}

	ipServices := []string{
		"https://api.ipify.org?format=text",
		"https://ifconfig.me/ip",
		"https://checkip.amazonaws.com",
		"https://icanhazip.com",
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, service := range ipServices {
		resp, err := client.Get(service)
		if err != nil {
			logger.Debug("failed to get ip from service", "service", service, "error", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Debug("service returned non-200 status", "service", service, "status", resp.StatusCode)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Debug("failed to read response body", "service", service, "error", err)
			continue
		}

		ip := strings.TrimSpace(string(body))
		if ip != "" {
			result.Success = true
			result.Message = ip
			logger.Info("successfully retrieved public ip", "ip", ip, "service", service)
			return result
		}
	}

	result.Success = false
	result.Message = "无法自动获取公网 IP，请手动输入"
	logger.Warn("failed to retrieve public ip from all services")
	return result
}

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "阿里云 IP 管理工具",
		Width:  1000,
		Height: 750,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		DisableResize: false,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
