# CyberStroll 分布式网络空间测绘系统 - 开发文档

## 📋 目录

- [系统概述](#系统概述)
- [架构设计](#架构设计)
- [核心组件](#核心组件)
- [数据流程](#数据流程)
- [开发环境搭建](#开发环境搭建)
- [配置说明](#配置说明)
- [API接口](#api接口)
- [数据结构](#数据结构)
- [部署指南](#部署指南)
- [故障排除](#故障排除)
- [开发规范](#开发规范)

## 🎯 系统概述

CyberStroll是一个分布式网络空间测绘系统，用于大规模网络资产发现、端口扫描、服务识别和数据富化。系统采用微服务架构，支持水平扩展，能够处理大量并发扫描任务。

### 核心功能

- **网络资产发现**: 支持IP、CIDR、IP范围、URL等多种目标格式
- **端口扫描**: 高性能并发端口扫描，支持TCP/UDP协议
- **服务识别**: 自动识别服务类型、版本信息和应用指纹
- **数据富化**: 对Web服务进行深度分析，提取网站信息、证书、API等
- **任务管理**: 支持系统任务和用户任务，具有优先级调度
- **实时监控**: 提供任务状态监控和统计信息

### 技术栈

- **后端**: Go 1.19+
- **消息队列**: Apache Kafka
- **数据存储**: Elasticsearch, MongoDB
- **容器化**: Docker, Docker Compose
- **Web界面**: HTML5, JavaScript, CSS3

## 🏗️ 架构设计

### 系统架构图

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web界面       │    │   任务管理器     │    │   扫描节点       │
│   (HTTP API)    │◄──►│ Task Manager    │◄──►│  Scan Node      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │                        │
                              ▼                        ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │     Kafka       │    │     Kafka       │
                       │  (消息队列)      │    │  (扫描结果)      │
                       └─────────────────┘    └─────────────────┘
                              │                        │
                              ▼                        ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   富化节点       │    │   处理节点       │    │  Elasticsearch  │
│Enrichment Node │◄──►│ Processor Node  │◄──►│   (数据存储)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                      │
        ▼                      ▼
┌─────────────────┐    ┌─────────────────┐
│  Elasticsearch  │    │    MongoDB      │
│   (富化数据)     │    │   (任务状态)     │
└─────────────────┘    └─────────────────┘
```

### 微服务组件

1. **任务管理器 (Task Manager)**: 任务调度和分发
2. **扫描节点 (Scan Node)**: 执行网络扫描任务
3. **处理节点 (Processor Node)**: 处理扫描结果
4. **富化节点 (Enrichment Node)**: Web服务数据富化
5. **搜索节点 (Search Node)**: 网络空间搜索界面

## 🔧 核心组件

### 1. 任务管理器 (Task Manager)

**职责**: 
- 接收用户提交的扫描任务
- URL解析和域名解析
- 任务优先级管理（系统任务 > 用户任务）
- 任务分发到Kafka队列

**关键特性**:
- 支持多种目标格式：IP、CIDR、IP范围、URL
- 自动URL解析：`https://www.baidu.com` → `180.101.49.44`
- 任务分类：URL任务自动归类为系统任务
- 限流控制：每个用户最大任务数、每个任务最大IP数

**配置文件**: `configs/task_manager.yaml`

```yaml
# 任务管理节点配置
node:
  id: "task-manager-001"
  name: "任务管理节点1"

kafka:
  brokers: ["localhost:9092"]
  system_task_topic: "system_tasks"
  regular_task_topic: "regular_tasks"

web:
  host: "0.0.0.0"
  port: 8088

storage:
  mongodb:
    uri: "mongodb://cyberstroll:cyberstroll123@localhost:27017/cyberstroll?authSource=admin"
```

### 2. 扫描节点 (Scan Node)

**职责**:
- 从Kafka消费扫描任务
- 执行端口扫描和服务识别
- 发送扫描结果到Kafka

**关键特性**:
- 高并发扫描：支持20个并发工作协程
- 优先级处理：优先处理系统任务
- 多种扫描类型：端口扫描、应用识别
- 自动重试机制

**配置文件**: `configs/scan_node.yaml`

```yaml
# 扫描节点配置
node:
  id: "scan-node-001"

kafka:
  brokers: ["localhost:9092"]
  system_task_topic: "system_tasks"
  regular_task_topic: "regular_tasks"
  result_topic: "scan_results"

scanner:
  max_concurrency: 20
  timeout: "3s"
  retry_count: 2
```

### 3. 处理节点 (Processor Node)

**职责**:
- 从Kafka消费扫描结果
- 数据格式化和标准化
- 存储到Elasticsearch
- 更新任务状态到MongoDB

**关键特性**:
- 批量处理：提高写入性能
- 数据验证：确保数据完整性
- 地理位置信息：可选的IP地理位置查询
- 错误处理：失败重试机制

**配置文件**: `configs/processor_node.yaml`

```yaml
# 处理节点配置
kafka:
  brokers: ["localhost:9092"]
  result_topic: "scan_results"

elasticsearch:
  urls: ["http://localhost:9200"]
  index: "cyberstroll_ip_scan"

processing:
  batch_size: 100
  batch_timeout: "5s"
  max_concurrency: 10
```

### 4. 富化节点 (Enrichment Node)

**职责**:
- 识别Web服务（HTTP/HTTPS）
- 提取网站信息、证书信息
- 技术指纹识别
- API发现和分析

**关键特性**:
- 智能识别：通过service字段识别Web资产
- 多维度富化：内容、证书、指纹、API
- 增量更新：在现有文档基础上添加富化数据
- 可配置功能：可选择启用/禁用特定富化功能

**配置文件**: `configs/enrichment_node.yaml`

```yaml
# 富化节点配置
node:
  id: "enrichment-node-001"

elasticsearch:
  urls: ["http://localhost:9200"]
  index: "cyberstroll_ip_scan"

enrichment:
  batch_size: 50
  worker_count: 5
  scan_interval: "30s"
  enable_cert: true
  enable_api: true
  enable_web_info: true
  enable_fingerprint: true
  enable_content: true
```

### 5. 搜索节点 (Search Node)

**职责**:
- 提供网络空间搜索功能
- 类似FOFA的搜索界面
- 多维度数据查询和过滤
- 搜索结果导出

**关键特性**:
- 多条件搜索：IP、端口、Banner、服务、协议、国家等
- 分页浏览：支持大量结果的分页显示
- 实时统计：显示搜索结果统计信息
- 数据导出：支持JSON和CSV格式导出
- 响应式界面：支持桌面和移动设备

**配置文件**: `configs/search_node.yaml`

```yaml
# 搜索节点配置
node:
  id: "search-node-001"

elasticsearch:
  urls: ["http://localhost:9200"]
  index: "cyberstroll_ip_scan"

web:
  host: "0.0.0.0"
  port: 8082
```

**API接口**:
- `GET /api/search` - 搜索数据
- `GET /api/stats` - 获取统计
- `GET /api/export` - 导出结果

**Web界面**: http://localhost:8082

详细使用说明请参考 `SEARCH_INTERFACE_GUIDE.md`。

## 📊 数据流程

### 完整数据流程

```
1. 用户提交任务
   ↓
2. 任务管理器解析目标
   ↓
3. 任务分发到Kafka
   ↓
4. 扫描节点消费任务
   ↓
5. 执行网络扫描
   ↓
6. 扫描结果发送到Kafka
   ↓
7. 处理节点消费结果
   ↓
8. 数据存储到Elasticsearch
   ↓
9. 富化节点识别Web服务
   ↓
10. 执行数据富化
    ↓
11. 富化数据更新到Elasticsearch
```

### 任务优先级

- **系统任务 (Priority: 10)**: URL任务、定时任务
- **用户任务 (Priority: 1)**: 用户手动提交的IP扫描任务

### 数据存储

**Elasticsearch索引结构**:
```json
{
  "ip": "98.88.224.123",
  "port": 80,
  "protocol": "tcp",
  "service": "http",
  "service_version": "gunicorn/19.9.0",
  "banner": "HTTP/1.1 200 OK...",
  "state": "open",
  "scan_time": "2026-01-20T20:54:12+08:00",
  "task_id": "9acccda7-73a2-4b5a-8d55-4978a9fcad4c",
  "node_id": "scan-node-001",
  "metadata": {
    "enrichment_data": {
      "web_info": {...},
      "cert_info": {...},
      "fingerprints": [...],
      "content_info": {...},
      "api_info": {...}
    }
  }
}
```

## 🚀 开发环境搭建

### 前置要求

- Go 1.19+
- Docker & Docker Compose
- Git

### 快速启动

1. **克隆项目**
```bash
git clone <repository-url>
cd CyberStroll
```

2. **启动基础服务**
```bash
# 启动Kafka、Elasticsearch、MongoDB
docker-compose -f docker-compose-simple.yaml up -d
```

3. **启动各个节点**

```bash
# 终端1: 启动任务管理器
go run cmd/task_manager/main.go -port 8088

# 终端2: 启动扫描节点
go run cmd/scan_node/main.go

# 终端3: 启动处理节点
go run cmd/processor_node/main.go

# 终端4: 启动富化节点
go run cmd/enrichment_node/main.go
```

4. **验证系统**
```bash
# 提交测试任务
curl -X POST "http://localhost:8088/api/tasks/submit" \
  -H "Content-Type: application/json" \
  -d '{
    "initiator": "admin",
    "targets": ["https://httpbin.org"],
    "task_type": "app_identification",
    "timeout": 30
  }'
```

### 开发工具

推荐使用以下工具进行开发：

- **IDE**: VS Code, GoLand
- **API测试**: Postman, curl
- **数据查看**: Elasticsearch Head, MongoDB Compass
- **日志查看**: Docker logs, tail

## ⚙️ 配置说明

### 环境变量

系统支持通过环境变量覆盖配置：

```bash
export KAFKA_BROKERS="localhost:9092"
export ELASTICSEARCH_URL="http://localhost:9200"
export MONGODB_URI="mongodb://localhost:27017/cyberstroll"
```

### 配置文件优先级

1. 命令行参数
2. 环境变量
3. 配置文件
4. 默认值

### 关键配置项

**Kafka配置**:
- `brokers`: Kafka集群地址
- `system_task_topic`: 系统任务队列
- `regular_task_topic`: 用户任务队列
- `result_topic`: 扫描结果队列

**扫描配置**:
- `max_concurrency`: 最大并发数
- `timeout`: 扫描超时时间
- `retry_count`: 重试次数

**存储配置**:
- `elasticsearch.urls`: ES集群地址
- `elasticsearch.index`: 索引名称
- `mongodb.uri`: MongoDB连接字符串

## 🔌 API接口

### 任务管理API

**提交扫描任务**
```http
POST /api/tasks/submit
Content-Type: application/json

{
  "initiator": "admin",
  "targets": ["https://www.baidu.com", "192.168.1.0/24"],
  "task_type": "app_identification",
  "ports": [80, 443, 8080],
  "timeout": 30
}
```

**查询任务状态**
```http
GET /api/tasks/{task_id}/status
```

**获取用户任务列表**
```http
GET /api/tasks?initiator=admin&limit=10
```

**获取系统统计**
```http
GET /api/stats
```

### 响应格式

**成功响应**:
```json
{
  "task_id": "a0d62b75-f036-4c33-9d32-a12c2a3279a7",
  "status": "success",
  "message": "任务提交成功",
  "target_count": 2
}
```

**错误响应**:
```json
{
  "status": "error",
  "message": "解析目标失败: 无效的IP地址",
  "target_count": 0
}
```

## 📋 数据结构

### 任务结构

```go
type Task struct {
    TaskID    string                 `json:"task_id"`
    IP        string                 `json:"ip"`
    TaskType  string                 `json:"task_type"`
    Priority  int                    `json:"priority"`
    User      string                 `json:"user,omitempty"`
    Config    map[string]interface{} `json:"config"`
    Timestamp int64                  `json:"timestamp"`
}
```

### 扫描结果结构

```go
type ScanResult struct {
    TaskID       string      `json:"task_id"`
    IP           string      `json:"ip"`
    ScanType     string      `json:"scan_type"`
    ScanStatus   string      `json:"scan_status"`
    ScanTime     string      `json:"scan_time"`
    ResponseTime int64       `json:"response_time"`
    Results      interface{} `json:"results"`
    ErrorMessage string      `json:"error_message,omitempty"`
    NodeID       string      `json:"node_id"`
    Timestamp    int64       `json:"timestamp"`
}
```

### 富化数据结构

```go
type EnrichmentData struct {
    CertInfo      *CertificateInfo `json:"cert_info,omitempty"`
    APIInfo       *APIInfo         `json:"api_info,omitempty"`
    WebInfo       *WebsiteInfo     `json:"web_info,omitempty"`
    Fingerprints  []Fingerprint    `json:"fingerprints,omitempty"`
    ContentInfo   *ContentInfo     `json:"content_info,omitempty"`
    EnrichTime    time.Time        `json:"enrich_time"`
}
```

## 🚀 部署指南

### Docker部署

1. **构建镜像**
```bash
# 构建任务管理器镜像
docker build -t cyberstroll/task-manager -f docker/Dockerfile.task-manager .

# 构建扫描节点镜像
docker build -t cyberstroll/scan-node -f docker/Dockerfile.scan-node .

# 构建处理节点镜像
docker build -t cyberstroll/processor-node -f docker/Dockerfile.processor-node .

# 构建富化节点镜像
docker build -t cyberstroll/enrichment-node -f docker/Dockerfile.enrichment-node .
```

2. **部署集群**
```bash
# 使用Docker Compose部署完整集群
docker-compose -f docker-compose.yaml up -d
```

### Kubernetes部署

```yaml
# k8s/cyberstroll-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cyberstroll-task-manager
spec:
  replicas: 2
  selector:
    matchLabels:
      app: task-manager
  template:
    metadata:
      labels:
        app: task-manager
    spec:
      containers:
      - name: task-manager
        image: cyberstroll/task-manager:latest
        ports:
        - containerPort: 8088
        env:
        - name: KAFKA_BROKERS
          value: "kafka:9092"
        - name: MONGODB_URI
          value: "mongodb://mongodb:27017/cyberstroll"
```

### 生产环境配置

**性能优化**:
- 增加扫描节点数量以提高扫描速度
- 调整Kafka分区数以提高并发处理能力
- 配置Elasticsearch集群以提高存储性能

**安全配置**:
- 启用Kafka SASL认证
- 配置Elasticsearch访问控制
- 使用TLS加密通信

**监控配置**:
- 集成Prometheus监控
- 配置Grafana仪表板
- 设置告警规则

## 🔍 故障排除

### 常见问题

**1. 任务管理器无法启动**
```bash
# 检查MongoDB连接
mongo mongodb://cyberstroll:cyberstroll123@localhost:27017/cyberstroll

# 检查端口占用
netstat -tlnp | grep 8088
```

**2. 扫描节点无法消费任务**
```bash
# 检查Kafka连接
kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic system_tasks

# 检查消费者组状态
kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group cyberstroll_scan_group
```

**3. 富化节点无法找到Web资产**
```bash
# 检查Elasticsearch中的数据
curl "http://localhost:9200/cyberstroll_ip_scan/_search?q=service:http&size=5"

# 检查富化节点配置
grep -A 5 "enrichment:" configs/enrichment_node.yaml
```

**4. 数据未写入Elasticsearch**
```bash
# 检查ES集群状态
curl "http://localhost:9200/_cluster/health"

# 检查索引状态
curl "http://localhost:9200/_cat/indices?v"
```

### 日志分析

**查看组件日志**:
```bash
# 任务管理器日志
tail -f logs/task_manager.log

# 扫描节点日志
tail -f logs/scan_node.log

# 处理节点日志
tail -f logs/processor_node.log

# 富化节点日志
tail -f logs/enrichment_node.log
```

**关键日志模式**:
- `任务提交成功`: 任务管理器成功接收任务
- `Worker-X 完成任务`: 扫描节点完成扫描
- `批量索引成功`: 处理节点成功写入ES
- `Web资产富化完成`: 富化节点完成富化

### 性能调优

**扫描性能**:
- 调整`max_concurrency`参数
- 优化`timeout`设置
- 增加扫描节点数量

**存储性能**:
- 调整ES批量写入大小
- 优化索引映射
- 配置ES集群分片

**网络优化**:
- 调整Kafka批量大小
- 优化网络超时设置
- 使用本地网络部署

## 📝 开发规范

### 代码规范

**Go代码规范**:
- 使用`gofmt`格式化代码
- 遵循Go命名约定
- 添加必要的注释
- 使用错误处理最佳实践

**项目结构**:
```
CyberStroll/
├── cmd/                    # 主程序入口
│   ├── task_manager/
│   ├── scan_node/
│   ├── processor_node/
│   └── enrichment_node/
├── internal/               # 内部包
│   ├── kafka/
│   ├── storage/
│   ├── scanner/
│   ├── taskmanager/
│   ├── processor/
│   └── enrichment/
├── pkg/                    # 公共包
│   └── config/
├── configs/                # 配置文件
├── docker/                 # Docker文件
├── scripts/                # 脚本文件
└── docs/                   # 文档
```

### Git工作流

**分支策略**:
- `main`: 主分支，稳定版本
- `develop`: 开发分支
- `feature/*`: 功能分支
- `hotfix/*`: 热修复分支

**提交规范**:
```
feat: 添加URL解析功能
fix: 修复富化节点数据更新问题
docs: 更新开发文档
refactor: 重构扫描引擎
test: 添加单元测试
```

### 测试规范

**单元测试**:
```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./internal/taskmanager

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**集成测试**:
```bash
# 运行集成测试
go test -tags=integration ./tests/integration
```

### 文档规范

- API文档使用OpenAPI 3.0规范
- 代码注释使用godoc格式
- 配置文件包含详细说明
- 部署文档包含完整步骤

## 📈 监控和运维

### 监控指标

**系统指标**:
- CPU使用率
- 内存使用率
- 磁盘I/O
- 网络流量

**业务指标**:
- 任务提交速率
- 扫描完成速率
- 富化成功率
- 错误率

**存储指标**:
- Elasticsearch索引大小
- MongoDB连接数
- Kafka消息积压

### 告警规则

```yaml
# prometheus/alerts.yml
groups:
- name: cyberstroll
  rules:
  - alert: TaskManagerDown
    expr: up{job="task-manager"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "任务管理器服务不可用"
      
  - alert: HighErrorRate
    expr: rate(scan_errors_total[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "扫描错误率过高"
```

### 备份策略

**数据备份**:
- Elasticsearch快照备份
- MongoDB定期备份
- 配置文件版本控制

**恢复流程**:
1. 停止相关服务
2. 恢复数据库数据
3. 恢复配置文件
4. 重启服务
5. 验证数据完整性

## 🔮 未来规划

### 功能扩展

- **多协议支持**: 支持更多网络协议扫描
- **漏洞检测**: 集成漏洞扫描功能
- **威胁情报**: 集成威胁情报数据
- **机器学习**: 智能资产分类和异常检测

### 性能优化

- **分布式扫描**: 支持跨地域分布式扫描
- **缓存优化**: 减少重复扫描
- **流式处理**: 实时数据处理
- **智能调度**: 基于负载的任务调度

### 运维增强

- **自动扩缩容**: 基于负载自动调整节点数量
- **故障自愈**: 自动检测和恢复故障节点
- **配置管理**: 集中化配置管理
- **审计日志**: 完整的操作审计

---

## 📞 联系方式

如有问题或建议，请联系开发团队：

- **项目地址**: [GitHub Repository]
- **文档地址**: [Documentation Site]
- **问题反馈**: [Issue Tracker]

---

*最后更新时间: 2026-01-20*