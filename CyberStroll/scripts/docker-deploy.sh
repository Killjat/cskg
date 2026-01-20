#!/bin/bash

# CyberStroll Docker部署脚本
# 一键部署所有依赖服务

set -e

echo "🚀 开始部署CyberStroll系统..."

# 检查Docker和Docker Compose
echo "检查Docker环境..."
if ! command -v docker &> /dev/null; then
    echo "❌ Docker未安装，请先安装Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose未安装，请先安装Docker Compose"
    exit 1
fi

# 进入项目目录
cd "$(dirname "$0")/.."

echo "📁 当前工作目录: $(pwd)"

# 停止并清理现有容器
echo "🧹 清理现有容器..."
docker-compose down -v 2>/dev/null || true

# 拉取最新镜像
echo "📥 拉取Docker镜像..."
docker-compose pull

# 启动基础服务
echo "🔧 启动基础服务..."
docker-compose up -d zookeeper mongodb elasticsearch redis

# 等待基础服务启动
echo "⏳ 等待基础服务启动 (30秒)..."
sleep 30

# 启动Kafka
echo "🔧 启动Kafka..."
docker-compose up -d kafka

# 等待Kafka启动
echo "⏳ 等待Kafka启动 (30秒)..."
sleep 30

# 创建Kafka主题
echo "📝 创建Kafka主题..."
./scripts/setup-kafka-topics.sh

# 创建Elasticsearch索引
echo "📝 创建Elasticsearch索引..."
./scripts/setup-elasticsearch.sh

# 启动管理界面
echo "🔧 启动管理界面..."
docker-compose up -d kafka-ui mongo-express kibana

echo "✅ 基础服务部署完成!"

# 显示服务状态
echo ""
echo "📊 服务状态:"
docker-compose ps

echo ""
echo "🌐 服务访问地址:"
echo "  Kafka UI:        http://localhost:8080"
echo "  MongoDB Express: http://localhost:8081 (admin/admin123)"
echo "  Kibana:          http://localhost:5601"
echo "  Elasticsearch:   http://localhost:9200"
echo "  MongoDB:         mongodb://localhost:27017"
echo "  Kafka:           localhost:9092"
echo "  Redis:           localhost:6379"

echo ""
echo "🔧 服务连接信息:"
echo "  MongoDB:"
echo "    - 管理员: cyberstroll/cyberstroll123"
echo "    - 应用用户: cyberstroll_user/cyberstroll_pass"
echo "    - 数据库: cyberstroll"
echo ""
echo "  Redis:"
echo "    - 密码: cyberstroll123"
echo ""
echo "  Kafka主题:"
echo "    - system_tasks (系统任务)"
echo "    - regular_tasks (常规任务)"
echo "    - scan_results (扫描结果)"
echo "    - enrichment_tasks (富化任务)"

echo ""
echo "🎉 CyberStroll基础环境部署完成!"
echo "现在可以启动CyberStroll应用节点了。"

# 健康检查
echo ""
echo "🏥 执行健康检查..."
sleep 10

# 检查MongoDB
echo -n "MongoDB: "
if docker exec cyberstroll-mongodb mongosh --eval "db.adminCommand('ping')" --quiet > /dev/null 2>&1; then
    echo "✅ 正常"
else
    echo "❌ 异常"
fi

# 检查Elasticsearch
echo -n "Elasticsearch: "
if curl -s http://localhost:9200/_cluster/health > /dev/null; then
    echo "✅ 正常"
else
    echo "❌ 异常"
fi

# 检查Kafka
echo -n "Kafka: "
if docker exec cyberstroll-kafka kafka-broker-api-versions --bootstrap-server localhost:9092 > /dev/null 2>&1; then
    echo "✅ 正常"
else
    echo "❌ 异常"
fi

# 检查Redis
echo -n "Redis: "
if docker exec cyberstroll-redis redis-cli -a cyberstroll123 ping > /dev/null 2>&1; then
    echo "✅ 正常"
else
    echo "❌ 异常"
fi

echo ""
echo "🚀 部署完成! 可以开始使用CyberStroll系统了。"