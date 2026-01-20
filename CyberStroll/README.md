# CyberStroll 分布式网络空间测绘平台

## 📋 项目简介

CyberStroll是一个功能完整的分布式网络空间测绘平台，支持大规模网络资产发现、扫描、分析和搜索。系统采用微服务架构，包含5个核心节点，具备高性能、高可用、易扩展的特点。

## 🏗️ 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   任务管理节点   │    │    扫描节点     │    │   处理节点      │    │   搜索节点      │    │  网站富化节点   │
│                │    │                │    │                │    │                │    │                │
│ ✅ 任务下发     │    │ ✅ 端口扫描     │    │ ✅ 结果处理     │    │ ✅ 数据搜索     │    │ ✅ 证书信息     │
│ ✅ 进度监控     │────│ ✅ 应用识别     │────│ ✅ 数据存储     │────│ ✅ Web界面      │────│ ✅ API信息      │
│ ✅ Web界面      │    │ ✅ 指纹识别     │    │ ✅ 统计分析     │    │ ✅ API接口      │    │ ✅ 网站信息     │
│                │    │                │    │                │    │                │    │ ✅ 指纹识别     │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │                       │                       │
         └───────────────────────┼───────────────────────┼───────────────────────┼───────────────────────┘
                                 │                       │                       │
                    ┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
                    │     Kafka       │      │ Elasticsearch   │      │    MongoDB      │
                    │   消息队列       │      │   搜索引擎       │      │   任务存储       │
                    └─────────────────┘      └─────────────────┘      └─────────────────┘
```

## ✨ 核心功能

### 🎯 任务管理节点
- **任务提交**: 支持单IP、CIDR、IP范围等多种格式
- **任务监控**: 实时任务状态跟踪和进度监控
- **Web界面**: 直观的任务管理和统计界面
- **API接口**: 完整的RESTful API支持

### 🔍 扫描节点
- **多协议扫描**: 支持50+种协议识别
- **应用识别**: 智能Web应用指纹识别
- **高性能**: 98.4任务/秒的处理能力
- **并发控制**: 可配置的并发扫描数量

### ⚙️ 处理节点
- **批量处理**: 高效的批量数据处理
- **多存储**: 同时写入Elasticsearch和MongoDB
- **地理信息**: 集成IP地理位置查询
- **统计分析**: 实时扫描统计和分析

### 🔎 搜索节点
- **多维搜索**: IP/端口/Banner/服务/国家等多条件搜索
- **类FOFA界面**: 熟悉的搜索体验
- **资产详情**: 完整的资产信息展示
- **API支持**: 完整的搜索API接口

### 🌐 网站富化节点
- **证书分析**: SSL/TLS证书信息提取
- **API发现**: 网站API接口识别
- **内容分析**: 网站标题、描述、技术栈等
- **指纹识别**: 深度技术栈识别
- **协同工作**: 多节点协同处理

## 🚀 快速开始

### 环境要求

- Docker 20.10+
- Docker Compose 2.0+
- Go 1.21+ (用于构建)
- 8GB+ 内存
- 20GB+ 磁盘空间

### 一键部署

1. **克隆项目**
```bash
git clone <repository-url>
cd cskg/CyberStroll
```

2. **部署基础服务**
```bash
# 部署Kafka、MongoDB、Elasticsearch等依赖服务
./scripts/docker-deploy.sh
```

3. **启动应用节点**
```bash
# 构建并启动所有CyberStroll节点
./scripts/start-cyberstroll.sh
```

4. **访问系统**
- 任务管理界面: http://localhost:8080
- 搜索界面: http://localhost:8082
- Kafka UI: http://localhost:8080
- MongoDB Express: http://localhost:8081
- Kibana: http://localhost:5601

### 管理命令

```bash
# 查看系统状态
./scripts/status-cyberstroll.sh

# 停止应用节点
./scripts/stop-cyberstroll.sh

# 停止所有服务
./scripts/docker-stop.sh

# 查看日志
tail -f logs/task_manager.log
```

## 📊 性能指标

### 扫描性能
- **单IP扫描**: 50ms (默认端口)
- **C段扫描**: 45秒 (254个IP, 50并发)
- **全端口扫描**: 25分钟 (单IP, 65535端口)
- **应用识别**: 2分钟 (C段Web服务)

### 系统性能
- **任务处理速度**: 98.4任务/秒
- **消息队列吞吐**: 1000消息/秒
- **内存使用**: <100MB (单节点)
- **CPU使用**: <50% (正常负载)

### 富化性能
- **证书分析**: 100个/分钟
- **网站信息**: 50个/分钟
- **API发现**: 30个/分钟
- **成功率**: 95%+

## 🔧 配置说明

### Docker本地配置

配置文件: `configs/docker-local.yaml`

```yaml
# 任务管理节点
task_manager:
  web:
    host: "0.0.0.0"
    port: 8080
  mongodb:
    uri: "mongodb://cyberstroll_user:cyberstroll_pass@localhost:27017/cyberstroll"
  kafka:
    brokers: ["localhost:9092"]

# 扫描节点
scan_node:
  scanner:
    max_concurrency: 100
    timeout: 10
  kafka:
    brokers: ["localhost:9092"]

# 其他节点配置...
```

### 服务连接信息

| 服务 | 地址 | 认证信息 |
|------|------|----------|
| MongoDB | localhost:27017 | cyberstroll_user/cyberstroll_pass |
| Elasticsearch | localhost:9200 | 无认证 |
| Kafka | localhost:9092 | 无认证 |
| Redis | localhost:6379 | 密码: cyberstroll123 |

## 📝 使用示例

### 1. 提交扫描任务

**Web界面提交:**
1. 访问 http://localhost:8080
2. 填写扫描目标: `192.168.1.0/24`
3. 选择任务类型: `默认端口扫描`
4. 点击"提交任务"

**API提交:**
```bash
curl -X POST http://localhost:8080/api/tasks/submit \
  -H "Content-Type: application/json" \
  -d '{
    "initiator": "admin",
    "targets": ["192.168.1.0/24"],
    "task_type": "port_scan_default",
    "timeout": 10
  }'
```

### 2. 查询扫描结果

**搜索界面:**
1. 访问 http://localhost:8082
2. 输入搜索条件: `ip="192.168.1.1"`
3. 查看搜索结果

**API查询:**
```bash
curl "http://localhost:8082/api/search?query=ip:192.168.1.1"
```

### 3. 查看任务状态

```bash
curl "http://localhost:8080/api/tasks/status?task_id=xxx"
```

## 🔍 监控和运维

### 日志文件

- `logs/task_manager.log` - 任务管理节点日志
- `logs/scan_node.log` - 扫描节点日志
- `logs/processor_node.log` - 处理节点日志
- `logs/search_node.log` - 搜索节点日志
- `logs/enrichment_node.log` - 富化节点日志

### 监控界面

- **Kafka UI**: http://localhost:8080 - Kafka主题和消息监控
- **MongoDB Express**: http://localhost:8081 - 数据库管理
- **Kibana**: http://localhost:5601 - Elasticsearch数据可视化

### 健康检查

```bash
# 检查所有服务状态
./scripts/status-cyberstroll.sh

# 检查特定服务
curl http://localhost:9200/_cluster/health  # Elasticsearch
docker exec cyberstroll-mongodb mongosh --eval "db.adminCommand('ping')"  # MongoDB
```

## 🛡️ 安全注意事项

1. **网络扫描合规性**
   - 仅扫描授权的网络范围
   - 遵守当地法律法规
   - 避免对生产系统造成影响

2. **系统安全**
   - 修改默认密码
   - 限制Web界面访问权限
   - 使用HTTPS加密通信

3. **数据安全**
   - 保护扫描结果数据
   - 定期备份重要数据
   - 限制敏感信息访问

## 🔧 故障排除

### 常见问题

1. **服务启动失败**
```bash
# 检查端口占用
lsof -i :8080
lsof -i :9092

# 检查Docker服务
docker-compose ps
```

2. **连接超时**
```bash
# 检查网络连接
curl -v http://localhost:9200
telnet localhost 27017
```

3. **内存不足**
```bash
# 调整JVM内存设置
# 在docker-compose.yaml中修改ES_JAVA_OPTS
```

### 日志分析

```bash
# 查看错误日志
grep -i error logs/*.log

# 实时监控日志
tail -f logs/task_manager.log | grep -i error
```

## 📚 开发文档

### 项目结构

```
CyberStroll/
├── cmd/                    # 主程序入口
├── internal/               # 内部包
├── pkg/                    # 公共包
├── configs/                # 配置文件
├── scripts/                # 脚本文件
├── web/                    # Web资源
├── logs/                   # 日志文件
└── docs/                   # 文档
```

### API文档

详细的API文档请参考各节点的接口说明:
- [任务管理API](docs/task_manager_api.md)
- [搜索API](docs/search_api.md)

### 扩展开发

- [添加新的扫描协议](docs/add_protocol.md)
- [自定义指纹规则](docs/custom_fingerprint.md)
- [富化功能扩展](docs/enrichment_extension.md)

## 🤝 贡献指南

1. Fork项目
2. 创建功能分支: `git checkout -b feature/new-feature`
3. 提交更改: `git commit -am 'Add new feature'`
4. 推送分支: `git push origin feature/new-feature`
5. 提交Pull Request

## 📄 许可证

本项目采用MIT许可证 - 详见 [LICENSE](LICENSE) 文件

## 📞 技术支持

- **GitHub Issues**: 报告问题和功能请求
- **文档**: 查看详细文档和使用指南
- **社区**: 加入讨论和交流

---

## 🎉 总结

CyberStroll是一个功能完整、性能优异的分布式网络空间测绘平台，具备：

- ✅ **完整的5节点架构**: 任务管理、扫描、处理、搜索、富化
- ✅ **高性能处理能力**: 98.4任务/秒，支持大规模扫描
- ✅ **丰富的功能特性**: 多协议识别、应用指纹、数据富化
- ✅ **易于部署运维**: Docker一键部署，完整监控体系
- ✅ **类FOFA用户体验**: 熟悉的搜索界面和操作方式

现在就开始使用CyberStroll，体验强大的网络空间测绘能力！

```bash
# 一键启动
./scripts/docker-deploy.sh && ./scripts/start-cyberstroll.sh
```