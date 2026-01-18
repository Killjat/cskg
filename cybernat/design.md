# 节点管理系统设计文档

## 1. 项目概述

本项目是一个基于Go语言的节点管理系统，采用Server-Client架构，用于管理远程节点并下发任务。系统支持通过配置文件管理节点信息，远程安装和删除客户端，以及向客户端下发任务并接收执行结果。

## 2. 技术栈

- **Go 1.21**：主要开发语言
- **NATS消息队列**：用于Server和Client之间的可靠通信
- **Cobra CLI框架**：用于构建命令行接口
- **YAML配置**：用于存储节点信息和系统配置
- **SSH**：用于远程安装和删除客户端

## 3. 核心功能模块

### 3.1 配置管理
- 加载和解析YAML配置文件
- 管理节点信息
- 提供节点查询功能

### 3.2 NATS通信
- Server发布任务到NATS
- Client订阅任务
- Client发布结果到NATS
- Server订阅结果

### 3.3 远程安装
- 通过SSH连接到目标节点
- 上传客户端二进制文件
- 创建systemd服务
- 启动和管理服务

### 3.4 任务执行
- 执行系统命令
- 处理超时
- 收集执行结果

### 3.5 节点管理
- 安装客户端到节点
- 删除节点上的客户端
- 检查节点状态

## 4. 数据模型

```go
// Node 节点信息
type Node struct {
	ID          string    // 节点唯一标识
	Name        string    // 节点名称
	Host        string    // 节点IP地址
	Port        int       // SSH端口
	User        string    // SSH用户名
	Password    string    // SSH密码
	Status      string    // 节点状态
	LastContact time.Time // 最后联系时间
}

// Task 任务信息
type Task struct {
	ID        string    // 任务唯一标识
	NodeID    string    // 目标节点ID
	Command   string    // 要执行的命令
	Args      []string  // 命令参数
	Timeout   int       // 执行超时时间（秒）
	CreatedAt time.Time // 任务创建时间
}

// TaskResult 任务结果
type TaskResult struct {
	TaskID    string    // 任务ID
	NodeID    string    // 节点ID
	Output    string    // 命令输出
	Error     string    // 错误信息
	ExitCode  int       // 退出码
	Completed bool      // 是否完成
	Timestamp time.Time // 结果时间
}

// Config 配置信息
type Config struct {
	Server   ServerConfig // 服务端配置
	NATS     NATSConfig   // NATS配置
	Nodes    []Node       // 节点列表
}

// ServerConfig 服务端配置
type ServerConfig struct {
	Port int // 服务端端口
}

// NATSConfig NATS配置
type NATSConfig struct {
	URL            string // NATS服务器URL
	TaskSubject    string // 任务主题
	ResultSubject  string // 结果主题
}
```

## 5. 主要函数设计

### 5.1 配置管理
```go
// LoadConfig 加载配置文件
func LoadConfig(filePath string) (*Config, error)

// GetNodeByID 根据ID获取节点信息
func GetNodeByID(config *Config, nodeID string) (*Node, error)
```

### 5.2 NATS通信
```go
// NATSClient NATS客户端结构体
type NATSClient struct {
	Conn          *nats.Conn
	TaskSubject   string
	ResultSubject string
}

// NewNATSClient 创建NATS客户端
func NewNATSClient(url, taskSubject, resultSubject string) (*NATSClient, error)

// Close 关闭NATS连接
func (c *NATSClient) Close()

// PublishTask 发布任务
func (c *NATSClient) PublishTask(task *Task) error

// SubscribeResults 订阅任务结果
func (c *NATSClient) SubscribeResults(callback func(*TaskResult)) error

// PublishResult 发布任务结果
func (c *NATSClient) PublishResult(result *TaskResult) error

// SubscribeTasks 订阅任务
func (c *NATSClient) SubscribeTasks(callback func(*Task)) error
```

### 5.3 远程安装
```go
// Installer 远程安装器
type Installer struct {
	ClientBinaryPath string
}

// NewInstaller 创建安装器
func NewInstaller(clientBinaryPath string) *Installer

// Install 安装客户端
func (i *Installer) Install(host string, port int, user, password, nodeID string) error

// Uninstall 卸载客户端
func (i *Installer) Uninstall(host string, port int, user, password, nodeID string) error
```

### 5.4 任务执行
```go
// Executor 任务执行器
type Executor struct {
	ResultCallback func(*TaskResult)
}

// NewExecutor 创建任务执行器
func NewExecutor(resultCallback func(*TaskResult)) *Executor

// ExecuteTask 执行任务
func (e *Executor) ExecuteTask(task *Task)
```

### 5.5 节点管理
```go
// Manager 节点管理器
type Manager struct {
	Installer *Installer
}

// NewManager 创建节点管理器
func NewManager(installer *Installer) *Manager

// InstallNode 安装节点
func (m *Manager) InstallNode(node *Node) error

// DeleteNode 删除节点
func (m *Manager) DeleteNode(node *Node) error

// GetNodeStatus 获取节点状态
func (m *Manager) GetNodeStatus(node *Node) (string, error)
```

## 6. 命令行接口设计

```
nodemanage
├── server      # 启动服务端
├── client      # 启动客户端
├── install     # 安装客户端到节点
├── delete      # 从节点删除客户端
├── task        # 向节点发送任务
├── status      # 查看节点状态
└── list        # 列出所有节点
```

## 7. 工作流程

1. **服务端启动流程**：
   - 加载配置文件
   - 连接到NATS服务器
   - 启动命令行接口
   - 等待用户命令

2. **客户端启动流程**：
   - 解析命令行参数
   - 连接到NATS服务器
   - 订阅任务主题
   - 等待接收任务

3. **安装流程**：
   - 用户执行`install`命令
   - 服务端连接到目标节点
   - 上传客户端二进制文件
   - 创建systemd服务
   - 启动服务

4. **任务执行流程**：
   - 用户执行`task`命令
   - 服务端创建任务并发布到NATS
   - 客户端接收任务
   - 客户端执行任务
   - 客户端发布结果到NATS
   - 服务端接收并展示结果

5. **删除流程**：
   - 用户执行`delete`命令
   - 服务端连接到目标节点
   - 停止并禁用服务
   - 删除客户端文件

## 8. 配置文件格式

```yaml
# 服务端配置
server:
  port: 8080

# NATS配置
nats:
  url: nats://localhost:4222
  task_subject: tasks
  result_subject: results

# 节点列表
nodes:
  - id: node-1
    name: Node 1
    host: 192.168.1.100
    port: 22
    user: root
    password: password
    status: ""
  - id: node-2
    name: Node 2
    host: 192.168.1.101
    port: 22
    user: root
    password: password
    status: ""
```

## 9. 部署说明

1. **安装NATS服务器**：
   - 下载并安装NATS服务器
   - 启动NATS服务器

2. **编译项目**：
   - `go build -o nodemanage .`

3. **启动服务端**：
   - `./nodemanage server`

4. **安装客户端**：
   - `./nodemanage install node-1`

5. **启动客户端**（自动完成）：
   - 安装后自动启动

6. **发送任务**：
   - `./nodemanage task node-1 ls -la`

## 10. 注意事项

1. 确保NATS服务器已正确配置和运行
2. 确保目标节点支持SSH连接
3. 确保目标节点已安装systemd
4. 确保服务端和客户端可以访问NATS服务器
5. 配置文件中的密码为明文，生产环境建议使用密钥认证

## 11. 扩展功能

1. 支持密钥认证
2. 支持任务队列管理
3. 支持节点分组
4. 支持任务模板
5. 支持Web界面
6. 支持监控和告警
7. 支持日志收集

## 12. 故障排查

1. 检查NATS连接是否正常
2. 检查SSH连接是否正常
3. 检查目标节点的systemd服务状态
4. 检查客户端日志
5. 检查服务端日志
