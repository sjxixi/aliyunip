<h1 align="center">配置阿里云产品IP白名单工具</h1>

<p align="center">
  <strong>一个用于快速管理阿里云多个服务 IP 白名单的桌面应用程序</strong>
</p>

<p align="center">
  <a href="https://github.com/wailsapp/wails">
    <img src="https://img.shields.io/badge/Built%20with-Wails-blue?style=flat-square" alt="Built with Wails">
  </a>
  <a href="https://go.dev">
    <img src="https://img.shields.io/badge/Go-1.23.4-blue?style=flat-square&logo=go" alt="Go Version">
  </a>
  <a href="https://opensource.org/licenses/MIT">
    <img src="https://img.shields.io/badge/License-MIT-yellow?style=flat-square" alt="License">
  </a>
  <a href="#">
    <img src="https://img.shields.io/badge/Platform-Windows-brightgreen?style=flat-square" alt="Platform">
  </a>
</p>

## 目录

- [功能特性](#功能特性)
- [支持的服务](#支持的服务)
- [快速开始](#快速开始)
- [使用说明](#使用说明)
- [权限配置](#权限配置)
- [项目结构](#项目结构)
- [技术栈](#技术栈)
- [开发指南](#开发指南)
- [常见问题](#常见问题)
- [贡献指南](#贡献指南)
- [许可证](#许可证)

## 功能特性

- 可视化资源选择 - 直观的界面展示所有可管理的阿里云资源
- IP 格式验证 - 自动验证 IP 地址和 CIDR 格式
- 批量配置 - 一次配置多个资源的 IP 白名单
- 实时进度 - 显示配置执行进度和详细结果
- 本地配置 - 自动保存阿里云认证信息（加密存储）
- 友好界面 - 现代化的用户体验设计
- 安全可靠 - 不生成本地日志文件，保护敏感信息
- 多端口支持 - ECS 安全组支持配置多个独立端口

## 支持的服务

| 服务名称 | 功能描述 |
|---------|---------|
| **ALB 访问控制策略** | 管理 ALB 访问控制策略条目 |
| **ECS 安全组规则** | 添加 ECS 安全组入站规则（仅显示非托管安全组） |
| **云防火墙地址簿** | 管理云防火墙地址簿的 IP 地址 |
| **RDS 白名单** | 管理 RDS 实例的 IP 白名单 |
| **PolarDB 白名单** | 管理 PolarDB 集群的 IP 白名单 |
| **Redis 白名单** | 管理 Redis 实例的 IP 白名单 |

## 快速开始

### 前置要求

- Go 1.23.4 或更高版本
- Wails CLI - 用于构建 Wails 应用

### 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 开发模式运行

```bash
cd cmd/wails-gui
wails dev
```

### 编译生产版本

```bash
cd cmd/wails-gui
wails build -tags desktop,production
```

编译后的可执行文件将位于 `cmd/wails-gui/build/bin/` 目录。

## 使用说明

### 1. 配置阿里云认证

1. 启动应用程序
2. 填写阿里云 AccessKey ID 和 AccessKey Secret
3. 选择您的区域（如 `cn-shanghai`、`cn-shanghai-finance-1`）
4. 点击"验证并继续"

### 2. 选择资源

1. 展开需要配置的服务类别
2. 勾选要添加 IP 的资源
3. 对于 RDS/PolarDB/Redis，选择要使用的 IP 分组

### 3. 配置 IP 地址

1. 输入 IP 地址（支持纯 IP 或 CIDR 格式，如 `192.168.1.1` 或 `192.168.1.0/24`）
2. 为每个 ECS 安全组单独配置端口（可选）
3. 为每个 ECS 安全组添加规则描述（可选）
4. 点击"下一步"预览

### 4. 执行配置

1. 确认配置信息
2. 点击"开始执行"
3. 观察进度和结果

## 权限配置

为了正常使用本工具，您需要为阿里云 AccessKey 配置相应的权限。

### 最小权限策略

我们提供了一个预定义的最小权限策略，包含了所有必要的权限：

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecs:DescribeSecurityGroups",
        "ecs:AuthorizeSecurityGroup"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "alb:ListAcls",
        "alb:AddEntriesToAcl"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "rds:DescribeDBInstances",
        "rds:DescribeDBInstanceIPArrayList",
        "rds:ModifySecurityIps"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "polardb:DescribeDBClusters",
        "polardb:DescribeDBClusterAccessWhitelist",
        "polardb:ModifyDBClusterAccessWhitelist"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "kvstore:DescribeInstances",
        "kvstore:DescribeSecurityIps",
        "kvstore:ModifySecurityIps"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "yundun-cloudfirewall:DescribeAddressBook",
        "yundun-cloudfirewall:ModifyAddressBook"
      ],
      "Resource": "*"
    }
  ]
}
```

完整的策略文件请参考 [RECOMMENDED_POLICY.json](./RECOMMENDED_POLICY.json)。

### 如何配置

1. 登录阿里云控制台
2. 进入 RAM 访问控制
3. 创建自定义权限策略，粘贴上述 JSON
4. 创建 RAM 用户并附加该策略
5. 为该用户生成 AccessKey

## 项目结构

```
aliyunip/
├── cmd/
│   └── wails-gui/              # Wails GUI 版本入口
│       ├── main.go             # 主应用程序
│       ├── wails.json          # Wails 配置
│       └── frontend/           # 前端界面
│           └── dist/           # 编译后的前端文件
├── internal/
│   ├── aliyun/                 # 阿里云 API 客户端
│   │   ├── alb/               # ALB 客户端
│   │   ├── cloudfw/           # 云防火墙客户端
│   │   ├── ecs/               # ECS 安全组客户端
│   │   ├── polardb/           # PolarDB 客户端
│   │   ├── rds/               # RDS 客户端
│   │   ├── redis/             # Redis 客户端
│   │   ├── client.go          # 基础客户端
│   │   └── errors.go          # 错误处理
│   ├── config/                 # 配置管理
│   └── logger/                 # 日志记录
├── pkg/
│   └── validator/              # IP 地址验证
├── RECOMMENDED_POLICY.json    # 推荐的阿里云权限策略
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## 技术栈

### 后端

- Go 1.23.4 - 现代化的编程语言
- Wails v2 - 跨平台桌面应用框架
- 阿里云 SDK - 官方阿里云 Go SDK

### 前端

- HTML5 - 页面结构
- CSS3 - 样式和动画
- 原生 JavaScript - 交互逻辑

## 开发指南

### 代码规范

项目遵循 Go 官方代码规范，已通过以下检查：

```bash
# 代码格式化
go fmt ./...

# 静态代码分析
go vet ./...

# 依赖管理
go mod tidy
```

### 运行测试

```bash
go test ./...
```

## 常见问题

### 为什么只能看到部分 ECS 安全组？

本工具只显示非托管安全组（ServiceManaged=false），托管安全组由其他阿里云服务自动管理，不建议手动修改。

### aksk存储路径？

C:\Users\您的用户名\.aliyun-ip-manager\config.json

### 历史记录存储路径？

C:\Users\您的用户名\.aliyun-ip-manager\history.json


## 贡献指南

欢迎贡献！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 致谢

- [Wails](https://wails.io) - 出色的桌面应用框架
- [阿里云 SDK for Go](https://github.com/aliyun/alibaba-cloud-sdk-go) - 官方阿里云 Go SDK

## 支持

如果您有任何问题或建议，请提交 [Issue](../../issues)。


