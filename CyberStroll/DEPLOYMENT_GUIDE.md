# CyberStroll 部署指南

## 🚀 Docker 本地部署 (推荐)

### 环境要求

- Docker 20.10+
- Docker Compose 2.0+
- Go 1.21+ (用于构建)
- 8GB+ 内存
- 20GB+ 磁盘空间

### 一键部署

```bash
# 1. 进入项目目录
cd cskg/CyberStroll

# 2. 部署基础服务 (Kafka、MongoDB、Elasticsearch等)
./scripts/docker-deploy.sh

# 3. 启动CyberStroll应用节点
./scripts/start-cyberstroll.sh

# 4. 检查系统状态
./scripts/status-cyberstroll.sh
```

### 服务访问地址

- **任务管理界面**: http://localhost:8080
- **搜索界面**: http://localhost:8082
- **Kafka UI**: http://localhost:8080
- **MongoDB Express**: http://localhost:8081 (admin/admin123)
- **Kibana**: http://localhost:5601

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

### 服务连接信息

| 服务 | 地址 | 认证信息 |
|------|------|----------|
| MongoDB | localhost:27017 | cyberstroll_user/cyberstroll_pass |
| Elasticsearch | localhost:9200 | 无认证 |
| Kafka | localhost:9092 | 无认证 |
| Redis | localhost:6379 | 密码: cyberstroll123 |

### Kafka主题

- `system_tasks` - 系统任务
- `regular_tasks` - 常规任务  
- `scan_results` - 扫描结果
- `enrichment_tasks` - 富化任务

---

## 📋 手动部署 (高级用户)

### 系统要求

### 硬件要求
- **CPU**: 4核心以上 (推荐8核心)
- **内存**: 8GB以上 (推荐16GB)
- **存储**: 100GB以上可用空间
- **网络**: 稳定的网络连接

### 软件要求
- **操作系统**: Linux (Ubuntu 20.04+, CentOS 7+) 或 macOS
- **Go**: 1.21或更高版本
- **Docker**: 20.10或更高版本 (可选)
- **Python**: 3.8或更高版本 (用于指纹识别)

## 🔧 依赖服务安装

### 1. Kafka 安装

```bash
# 下载Kafka
wget https://downloads.apache.org/kafka/2.13-3.6.1/kafka_2.13-3.6.1.tgz
tar -xzf kafka_2.13-3.6.1.tgz
cd kafka_2.13-3.6.1

# 启动Zookeeper
bin/zookeeper-server-start.sh config/zookeeper.properties &

# 启动Kafka
bin/kafka-server-start.sh config/server.properties &

# 创建主题
bin/kafka-topics.sh --create --topic system_tasks --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
bin/kafka-topics.sh --create --topic regular_tasks --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
bin/kafka-topics.sh --create --topic scan_results --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
```

### 2. MongoDB 安装

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y mongodb

# CentOS/RHEL
sudo yum install -y mongodb-server

# 启动MongoDB
sudo systemctl start mongod
sudo systemctl enable mongod

# 创建数据库和用户
mongo
> use cyberstroll
> db.createUser({
    user: "cyberstroll",
    pwd: "password",
    roles: ["readWrite"]
})
```

### 3. Elasticsearch 安装

```bash
# 下载Elasticsearch
wget https://artifacts.elastic.co/downloads/elasticsearch/elasticsearch-8.15.0-linux-x86_64.tar.gz
tar -xzf elasticsearch-8.15.0-linux-x86_64.tar.gz
cd elasticsearch-8.15.0

# 配置Elasticsearch
echo "xpack.security.enabled: false" >> config/elasticsearch.yml
echo "discovery.type: single-node" >> config/elasticsearch.yml

# 启动Elasticsearch
./bin/elasticsearch &
```

## 🚀 CyberStroll 部署

### 1. 获取源代码

```bash
git clone <repository-url>
cd CyberStroll
```

### 2. 构建项目

```bash
# 赋予执行权限
chmod +x scripts/build.sh

# 执行构建
./scripts/build.sh
```

### 3. 配置文件设置

#### 扫描节点配置 (configs/scan_node.yaml)
```yaml
service:
  name: "scan-node"
  node_id: "scan-001"

scanner:
  max_concurrency: 100
  timeout: 10s
  retry_count: 3

kafka:
  brokers:
    - "localhost:9092"
  consumer_group: "scan-group"

mongodb:
  uri: "mongodb://localhost:27017"
  database: "cyberstroll"
```

#### 任务管理节点配置 (configs/task_manager.yaml)
```yaml
service:
  name: "task-manager"

web:
  host: "0.0.0.0"
  port: 8080

kafka:
  brokers:
    - "localhost:9092"

mongodb:
  uri: "mongodb://localhost:27017"
  database: "cyberstroll"
```

#### 处理节点配置 (configs/processor_node.yaml)
```yaml
service:
  name: "processor-node"

kafka:
  brokers:
    - "localhost:9092"
  consumer_group: "processor-group"

elasticsearch:
  urls:
    - "http://localhost:9200"
  index: "cyberstroll_ip_scan"

mongodb:
  uri: "mongodb://localhost:27017"
  database: "cyberstroll"

processor:
  batch_size: 100
  batch_timeout: 5s
  max_concurrency: 10
```

#### 搜索节点配置 (configs/search_node.yaml)
```yaml
service:
  name: "search-node"

web:
  host: "0.0.0.0"
  port: 8081

elasticsearch:
  urls:
    - "http://localhost:9200"
  index: "cyberstroll_ip_scan"
```

### 4. 启动服务

#### 方式一: 手动启动

```bash
# 启动任务管理节点
./bin/task_manager --config configs/task_manager.yaml &

# 启动扫描节点
./bin/scan_node --config configs/scan_node.yaml &

# 启动处理节点
./bin/processor_node --config configs/processor_node.yaml &

# 启动搜索节点
./bin/search_node --config configs/search_node.yaml &
```

#### 方式二: 使用systemd服务

创建服务文件:

```bash
# /etc/systemd/system/cyberstroll-task-manager.service
[Unit]
Description=CyberStroll Task Manager
After=network.target

[Service]
Type=simple
User=cyberstroll
WorkingDirectory=/opt/cyberstroll
ExecStart=/opt/cyberstroll/bin/task_manager --config /opt/cyberstroll/configs/task_manager.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启动服务:
```bash
sudo systemctl daemon-reload
sudo systemctl enable cyberstroll-task-manager
sudo systemctl start cyberstroll-task-manager
```

## 🐳 Docker 部署

### 1. 创建 Docker Compose 文件

```yaml
# docker-compose.yml
version: '3.8'

services:
  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000

  kafka:
    image: confluentinc/cp-kafka:latest
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1

  mongodb:
    image: mongo:5.0
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: password
    volumes:
      - mongodb_data:/data/db

  elasticsearch:
    image: elasticsearch:8.15.0
    ports:
      - "9200:9200"
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data

  task-manager:
    build: .
    command: ./bin/task_manager --config configs/task_manager.yaml
    ports:
      - "8080:8080"
    depends_on:
      - kafka
      - mongodb
    volumes:
      - ./configs:/app/configs

  scan-node:
    build: .
    command: ./bin/scan_node --config configs/scan_node.yaml
    depends_on:
      - kafka
      - mongodb
    volumes:
      - ./configs:/app/configs

  processor-node:
    build: .
    command: ./bin/processor_node --config configs/processor_node.yaml
    depends_on:
      - kafka
      - mongodb
      - elasticsearch
    volumes:
      - ./configs:/app/configs

  search-node:
    build: .
    command: ./bin/search_node --config configs/search_node.yaml
    ports:
      - "8081:8081"
    depends_on:
      - elasticsearch
    volumes:
      - ./configs:/app/configs

volumes:
  mongodb_data:
  elasticsearch_data:
```

### 2. 创建 Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN ./scripts/build.sh

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/bin/ ./bin/
COPY --from=builder /app/configs/ ./configs/
COPY --from=builder /app/web/ ./web/

CMD ["./bin/task_manager"]
```

### 3. 启动 Docker 环境

```bash
# 构建和启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f task-manager
```

## 🔍 验证部署

### 1. 检查服务状态

```bash
# 检查端口监听
netstat -tlnp | grep -E "(8080|8081|9092|27017|9200)"

# 检查进程
ps aux | grep -E "(task_manager|scan_node|processor_node|search_node)"
```

### 2. 功能测试

```bash
# 测试任务管理节点API
curl http://localhost:8080/api/stats

# 测试搜索节点API
curl http://localhost:8081/api/recent

# 提交测试任务
curl -X POST http://localhost:8080/api/tasks/submit \
  -H "Content-Type: application/json" \
  -d '{
    "initiator": "test",
    "targets": ["127.0.0.1"],
    "task_type": "port_scan_default"
  }'
```

### 3. Web界面访问

- 任务管理界面: http://localhost:8080
- 搜索界面: http://localhost:8081

## 📊 监控和维护

### 1. 日志管理

```bash
# 查看实时日志
tail -f logs/task_manager.log
tail -f logs/scan_node.log
tail -f logs/processor_node.log
tail -f logs/search_node.log

# 日志轮转配置
# /etc/logrotate.d/cyberstroll
/opt/cyberstroll/logs/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 cyberstroll cyberstroll
}
```

### 2. 性能监控

```bash
# 系统资源监控
htop
iotop
nethogs

# 服务特定监控
curl http://localhost:8080/api/stats
curl http://localhost:8081/api/stats
```

### 3. 数据备份

```bash
# MongoDB备份
mongodump --host localhost:27017 --db cyberstroll --out /backup/mongodb/

# Elasticsearch备份
curl -X PUT "localhost:9200/_snapshot/backup_repo" -H 'Content-Type: application/json' -d'
{
  "type": "fs",
  "settings": {
    "location": "/backup/elasticsearch"
  }
}'
```

## 🚨 故障排除

### 常见问题

1. **Kafka连接失败**
   - 检查Kafka服务状态
   - 验证网络连接
   - 检查防火墙设置

2. **MongoDB连接超时**
   - 检查MongoDB服务状态
   - 验证认证信息
   - 检查网络配置

3. **Elasticsearch索引失败**
   - 检查ES服务状态
   - 验证索引权限
   - 检查磁盘空间

4. **扫描节点无响应**
   - 检查任务队列状态
   - 验证网络权限
   - 检查资源使用情况

### 调试模式

```bash
# 启用调试模式
./bin/scan_node --config configs/scan_node.yaml --debug

# 查看详细日志
export LOG_LEVEL=debug
./bin/task_manager --config configs/task_manager.yaml
```

## 🔒 安全配置

### 1. 网络安全

```bash
# 防火墙配置
sudo ufw allow 8080/tcp  # 任务管理界面
sudo ufw allow 8081/tcp  # 搜索界面
sudo ufw deny 9092/tcp   # Kafka (内部访问)
sudo ufw deny 27017/tcp  # MongoDB (内部访问)
sudo ufw deny 9200/tcp   # Elasticsearch (内部访问)
```

### 2. 认证配置

```yaml
# 添加到配置文件
auth:
  enable: true
  jwt_secret: "your-secret-key"
  token_expire: "24h"
```

### 3. HTTPS配置

```yaml
# Web服务HTTPS配置
web:
  tls:
    enable: true
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
```

## 📈 性能优化

### 1. 系统调优

```bash
# 增加文件描述符限制
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# 网络参数优化
echo "net.core.somaxconn = 65535" >> /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65535" >> /etc/sysctl.conf
sysctl -p
```

### 2. 应用优化

```yaml
# 扫描节点优化
scanner:
  max_concurrency: 200  # 根据硬件调整
  batch_size: 1000      # 批量处理大小
  
# 处理节点优化
processor:
  batch_size: 500       # ES批量索引大小
  max_concurrency: 20   # 并发处理数
```

---

## 📞 技术支持

如有部署问题，请：
1. 查看日志文件获取详细错误信息
2. 检查系统资源使用情况
3. 验证网络连接和权限设置
4. 参考故障排除章节

更多技术支持请联系开发团队或提交GitHub Issue。