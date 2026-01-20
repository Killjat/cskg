#!/bin/bash

# CyberStroll Web界面演示启动脚本

echo "🚀 启动 CyberStroll Web界面演示..."

# 检查可执行文件
if [ ! -f "bin/task_manager" ]; then
    echo "❌ task_manager 可执行文件不存在，请先运行构建"
    exit 1
fi

if [ ! -f "bin/search_node" ]; then
    echo "❌ search_node 可执行文件不存在，请先运行构建"
    exit 1
fi

# 创建日志目录
mkdir -p logs

# 启动任务管理节点 (端口 8080)
echo "🌐 启动任务管理节点 (http://localhost:8080)..."
nohup ./bin/task_manager --config configs/task_manager.yaml > logs/task_manager_demo.log 2>&1 &
TASK_MANAGER_PID=$!
echo $TASK_MANAGER_PID > logs/task_manager_demo.pid

# 等待服务启动
sleep 3

# 启动搜索节点 (端口 8081)
echo "🔍 启动搜索节点 (http://localhost:8081)..."
nohup ./bin/search_node --config configs/search_node.yaml > logs/search_node_demo.log 2>&1 &
SEARCH_NODE_PID=$!
echo $SEARCH_NODE_PID > logs/search_node_demo.pid

# 等待服务完全启动
echo "⏳ 等待服务启动..."
sleep 5

# 检查服务状态
echo "📊 检查服务状态..."

if curl -s http://localhost:8080 > /dev/null; then
    echo "✅ 任务管理界面正常: http://localhost:8080"
else
    echo "❌ 任务管理界面启动失败"
fi

if curl -s http://localhost:8081 > /dev/null; then
    echo "✅ 搜索界面正常: http://localhost:8081"
else
    echo "❌ 搜索界面启动失败"
fi

echo ""
echo "🎉 Web界面演示已启动！"
echo ""
echo "📱 访问地址:"
echo "  任务管理界面: http://localhost:8080"
echo "  搜索界面:     http://localhost:8081"
echo ""
echo "📋 管理命令:"
echo "  查看日志:     tail -f logs/task_manager_demo.log"
echo "  查看日志:     tail -f logs/search_node_demo.log"
echo "  停止服务:     ./stop_web_demo.sh"
echo ""
echo "⚠️  注意: 这是演示模式，某些功能可能需要 Kafka、MongoDB、Elasticsearch 支持"