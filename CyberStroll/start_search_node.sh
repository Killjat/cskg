#!/bin/bash

# 启动搜索节点

echo "🔍 启动CyberStroll搜索节点..."

# 检查是否已构建
if [ ! -f "search_node" ]; then
    echo "搜索节点未构建，正在构建..."
    ./build_search_node.sh
fi

# 检查Elasticsearch是否运行
echo "检查Elasticsearch连接..."
if ! curl -s http://localhost:9200/_cluster/health > /dev/null; then
    echo "⚠️  警告: Elasticsearch (localhost:9200) 似乎未运行"
    echo "请确保Elasticsearch已启动，或检查配置文件中的连接地址"
fi

# 创建日志目录
mkdir -p logs

# 启动搜索节点
echo "启动搜索节点..."
echo "Web界面: http://localhost:8082"
echo "按 Ctrl+C 停止服务"
echo ""

./search_node -config configs/search_node.yaml