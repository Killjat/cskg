#!/bin/bash

# CyberStroll应用启动脚本
# 启动所有CyberStroll节点

set -e

echo "🚀 启动CyberStroll应用节点..."

# 进入项目目录
cd "$(dirname "$0")/.."

echo "📁 当前工作目录: $(pwd)"

# 检查依赖服务是否运行
echo "🔍 检查依赖服务状态..."

# 检查MongoDB
if ! docker exec cyberstroll-mongodb mongosh --eval "db.adminCommand('ping')" --quiet > /dev/null 2>&1; then
    echo "❌ MongoDB服务未运行，请先执行: ./scripts/docker-deploy.sh"
    exit 1
fi

# 检查Elasticsearch
if ! curl -s http://localhost:9200/_cluster/health > /dev/null; then
    echo "❌ Elasticsearch服务未运行，请先执行: ./scripts/docker-deploy.sh"
    exit 1
fi

# 检查Kafka
if ! docker exec cyberstroll-kafka kafka-broker-api-versions --bootstrap-server localhost:9092 > /dev/null 2>&1; then
    echo "❌ Kafka服务未运行，请先执行: ./scripts/docker-deploy.sh"
    exit 1
fi

echo "✅ 所有依赖服务正常运行"

# 构建应用
echo "🔨 构建CyberStroll应用..."
if [ ! -f "./scripts/build.sh" ]; then
    echo "❌ 构建脚本不存在，请检查项目结构"
    exit 1
fi

./scripts/build.sh

# 检查可执行文件
if [ ! -f "./bin/task_manager" ]; then
    echo "❌ 任务管理节点可执行文件不存在"
    exit 1
fi

if [ ! -f "./bin/scan_node" ]; then
    echo "❌ 扫描节点可执行文件不存在"
    exit 1
fi

# 创建日志目录
mkdir -p logs

# 启动任务管理节点
echo "🔧 启动任务管理节点..."
nohup ./bin/task_manager --config configs/docker-local.yaml > logs/task_manager.log 2>&1 &
TASK_MANAGER_PID=$!
echo "任务管理节点已启动 (PID: $TASK_MANAGER_PID)"

# 等待任务管理节点启动
sleep 5

# 启动扫描节点
echo "🔧 启动扫描节点..."
nohup ./bin/scan_node --config configs/docker-local.yaml > logs/scan_node.log 2>&1 &
SCAN_NODE_PID=$!
echo "扫描节点已启动 (PID: $SCAN_NODE_PID)"

# 启动处理节点 (如果存在)
if [ -f "./bin/processor_node" ]; then
    echo "🔧 启动处理节点..."
    nohup ./bin/processor_node --config configs/docker-local.yaml > logs/processor_node.log 2>&1 &
    PROCESSOR_PID=$!
    echo "处理节点已启动 (PID: $PROCESSOR_PID)"
fi

# 启动搜索节点 (如果存在)
if [ -f "./bin/search_node" ]; then
    echo "🔧 启动搜索节点..."
    nohup ./bin/search_node --config configs/docker-local.yaml > logs/search_node.log 2>&1 &
    SEARCH_PID=$!
    echo "搜索节点已启动 (PID: $SEARCH_PID)"
fi

# 启动富化节点 (如果存在)
if [ -f "./bin/enrichment_node" ]; then
    echo "🔧 启动富化节点..."
    nohup ./bin/enrichment_node --config configs/docker-local.yaml > logs/enrichment_node.log 2>&1 &
    ENRICHMENT_PID=$!
    echo "富化节点已启动 (PID: $ENRICHMENT_PID)"
fi

# 保存PID到文件
echo $TASK_MANAGER_PID > logs/task_manager.pid
echo $SCAN_NODE_PID > logs/scan_node.pid
[ ! -z "$PROCESSOR_PID" ] && echo $PROCESSOR_PID > logs/processor_node.pid
[ ! -z "$SEARCH_PID" ] && echo $SEARCH_PID > logs/search_node.pid
[ ! -z "$ENRICHMENT_PID" ] && echo $ENRICHMENT_PID > logs/enrichment_node.pid

echo ""
echo "✅ CyberStroll应用节点启动完成!"

echo ""
echo "🌐 应用访问地址:"
echo "  任务管理界面: http://localhost:8080"
[ ! -z "$SEARCH_PID" ] && echo "  搜索界面:     http://localhost:8082"

echo ""
echo "📊 节点状态:"
echo "  任务管理节点: PID $TASK_MANAGER_PID"
echo "  扫描节点:     PID $SCAN_NODE_PID"
[ ! -z "$PROCESSOR_PID" ] && echo "  处理节点:     PID $PROCESSOR_PID"
[ ! -z "$SEARCH_PID" ] && echo "  搜索节点:     PID $SEARCH_PID"
[ ! -z "$ENRICHMENT_PID" ] && echo "  富化节点:     PID $ENRICHMENT_PID"

echo ""
echo "📝 日志文件:"
echo "  任务管理节点: logs/task_manager.log"
echo "  扫描节点:     logs/scan_node.log"
[ ! -z "$PROCESSOR_PID" ] && echo "  处理节点:     logs/processor_node.log"
[ ! -z "$SEARCH_PID" ] && echo "  搜索节点:     logs/search_node.log"
[ ! -z "$ENRICHMENT_PID" ] && echo "  富化节点:     logs/enrichment_node.log"

echo ""
echo "🔧 管理命令:"
echo "  查看状态: ./scripts/status-cyberstroll.sh"
echo "  停止应用: ./scripts/stop-cyberstroll.sh"
echo "  查看日志: tail -f logs/task_manager.log"

echo ""
echo "🎉 CyberStroll系统已完全启动!"