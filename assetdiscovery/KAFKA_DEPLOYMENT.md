# Kafka部署和Topic创建指南

## 1. 下载和安装Kafka

### 1.1 下载Kafka

```bash
# 访问Kafka官网下载最新稳定版
# https://kafka.apache.org/downloads

# 或使用curl下载
curl -O https://downloads.apache.org/kafka/3.6.1/kafka_2.13-3.6.1.tgz
```

### 1.2 解压Kafka

```bash
tar -xzf kafka_2.13-3.6.1.tgz
cd kafka_2.13-3.6.1
```

## 2. 启动Kafka服务

Kafka依赖Zookeeper，所以需要先启动Zookeeper，然后再启动Kafka。

### 2.1 启动Zookeeper

```bash
# 使用默认配置启动Zookeeper
bin/zookeeper-server-start.sh config/zookeeper.properties &
```

### 2.2 启动Kafka

```bash
# 使用默认配置启动Kafka
bin/kafka-server-start.sh config/server.properties &
```

## 3. 创建内网资产探查系统所需的Topic

根据系统设计，我们需要创建两个Topic：
1. `assetdiscovery_tasks` - 用于服务端向客户端下发任务
2. `assetdiscovery_results` - 用于客户端向服务端上报结果

### 3.1 创建任务Topic

```bash
bin/kafka-topics.sh --create --topic assetdiscovery_tasks --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
```

### 3.2 创建结果Topic

```bash
bin/kafka-topics.sh --create --topic assetdiscovery_results --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
```

## 4. 验证Topic创建成功

```bash
# 列出所有Topic
bin/kafka-topics.sh --list --bootstrap-server localhost:9092
```

## 5. 测试Topic功能

### 5.1 发送测试消息

```bash
# 向任务Topic发送测试消息
echo "test message" | bin/kafka-console-producer.sh --broker-list localhost:9092 --topic assetdiscovery_tasks
```

### 5.2 接收测试消息

```bash
# 从任务Topic接收测试消息
bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic assetdiscovery_tasks --from-beginning
```

## 6. 配置系统连接Kafka

在系统的`config/config.yaml`文件中，配置Kafka连接信息：

```yaml
kafka:
  brokers: ["localhost:9092"]
  task_topic: "assetdiscovery_tasks"
  result_topic: "assetdiscovery_results"
  group_id: "assetdiscovery_group"
```

## 7. 停止Kafka服务

### 7.1 停止Kafka

```bash
bin/kafka-server-stop.sh
```

### 7.2 停止Zookeeper

```bash
bin/zookeeper-server-stop.sh
```

## 8. 生产环境配置建议

1. **增加副本因子**：生产环境建议设置`replication-factor`为3，提高可靠性
2. **调整分区数**：根据集群规模和吞吐量调整`partitions`数量
3. **配置持久化**：确保Kafka数据持久化到可靠的存储设备
4. **监控**：配置Kafka监控，如Prometheus + Grafana
5. **安全配置**：配置Kafka的认证和授权机制

## 9. 常见问题排查

### 9.1 无法连接到Kafka
- 检查Kafka服务是否正在运行
- 检查防火墙设置
- 检查`advertised.listeners`配置

### 9.2 Topic创建失败
- 检查Zookeeper和Kafka服务是否正常运行
- 检查Topic名称是否符合规范
- 检查权限设置

### 9.3 消息发送失败
- 检查Kafka连接配置
- 检查Topic是否存在
- 检查网络连接

## 10. 自动化脚本

以下是一个简单的自动化脚本，用于启动Kafka服务并创建Topic：

```bash
#!/bin/bash

# 启动Zookeeper
echo "Starting Zookeeper..."
bin/zookeeper-server-start.sh config/zookeeper.properties &
sleep 5

# 启动Kafka
echo "Starting Kafka..."
bin/kafka-server-start.sh config/server.properties &
sleep 10

# 创建Topic
echo "Creating topics..."
bin/kafka-topics.sh --create --topic assetdiscovery_tasks --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
bin/kafka-topics.sh --create --topic assetdiscovery_results --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1

# 验证Topic创建成功
echo "Listing topics..."
bin/kafka-topics.sh --list --bootstrap-server localhost:9092

echo "Kafka deployment completed successfully!"
```

将以上脚本保存为`start-kafka.sh`，然后执行：

```bash
chmod +x start-kafka.sh
./start-kafka.sh
```

---

按照以上步骤，您就可以成功部署Kafka并创建内网资产探查系统所需的Topic了。如果您在部署过程中遇到任何问题，请参考Kafka官方文档或社区支持。