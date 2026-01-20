#!/bin/bash

# Kafka主题创建脚本
# 等待Kafka启动完成后创建所需的主题

echo "等待Kafka服务启动..."
sleep 30

# Kafka容器名称
KAFKA_CONTAINER="cyberstroll-kafka"

# 检查Kafka是否就绪
echo "检查Kafka服务状态..."
docker exec $KAFKA_CONTAINER kafka-broker-api-versions --bootstrap-server localhost:9092

if [ $? -eq 0 ]; then
    echo "Kafka服务已就绪，开始创建主题..."
    
    # 创建系统任务主题
    echo "创建系统任务主题: system_tasks"
    docker exec $KAFKA_CONTAINER kafka-topics --create \
        --bootstrap-server localhost:9092 \
        --topic system_tasks \
        --partitions 3 \
        --replication-factor 1 \
        --config retention.ms=604800000
    
    # 创建常规任务主题
    echo "创建常规任务主题: regular_tasks"
    docker exec $KAFKA_CONTAINER kafka-topics --create \
        --bootstrap-server localhost:9092 \
        --topic regular_tasks \
        --partitions 6 \
        --replication-factor 1 \
        --config retention.ms=604800000
    
    # 创建扫描结果主题
    echo "创建扫描结果主题: scan_results"
    docker exec $KAFKA_CONTAINER kafka-topics --create \
        --bootstrap-server localhost:9092 \
        --topic scan_results \
        --partitions 6 \
        --replication-factor 1 \
        --config retention.ms=2592000000
    
    # 创建富化任务主题
    echo "创建富化任务主题: enrichment_tasks"
    docker exec $KAFKA_CONTAINER kafka-topics --create \
        --bootstrap-server localhost:9092 \
        --topic enrichment_tasks \
        --partitions 3 \
        --replication-factor 1 \
        --config retention.ms=604800000
    
    # 列出所有主题
    echo "当前Kafka主题列表:"
    docker exec $KAFKA_CONTAINER kafka-topics --list --bootstrap-server localhost:9092
    
    # 显示主题详情
    echo "主题详细信息:"
    docker exec $KAFKA_CONTAINER kafka-topics --describe --bootstrap-server localhost:9092
    
    echo "Kafka主题创建完成!"
else
    echo "错误: Kafka服务未就绪，请检查服务状态"
    exit 1
fi