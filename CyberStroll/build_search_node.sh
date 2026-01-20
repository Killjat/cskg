#!/bin/bash

# 构建搜索节点

echo "🔍 构建CyberStroll搜索节点..."

# 设置Go环境
export GO111MODULE=on
export GOPROXY=https://goproxy.cn,direct

# 创建日志目录
mkdir -p logs

# 构建搜索节点
echo "构建搜索节点..."
cd cmd/search_node
go build -o ../../search_node .
cd ../..

if [ -f "search_node" ]; then
    echo "✅ 搜索节点构建成功: search_node"
    echo ""
    echo "使用方法:"
    echo "  ./search_node -config configs/search_node.yaml"
    echo "  ./search_node -port 8082"
    echo "  ./search_node -test  # 测试模式"
    echo ""
    echo "Web界面: http://localhost:8082"
    echo "API接口:"
    echo "  GET /api/search?query=apache&ip=192.168.1.1&port=80"
    echo "  GET /api/stats"
    echo "  GET /api/export?format=json"
else
    echo "❌ 搜索节点构建失败"
    exit 1
fi