#!/bin/bash

# CyberStroll Docker停止脚本
# 停止所有服务并清理资源

set -e

echo "🛑 停止CyberStroll系统..."

# 进入项目目录
cd "$(dirname "$0")/.."

echo "📁 当前工作目录: $(pwd)"

# 显示当前运行的容器
echo "📊 当前运行的容器:"
docker-compose ps

echo ""
echo "🛑 停止所有服务..."

# 停止所有服务
docker-compose down

echo "✅ 所有服务已停止"

# 询问是否清理数据
echo ""
read -p "是否清理所有数据卷? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🧹 清理数据卷..."
    docker-compose down -v
    echo "✅ 数据卷已清理"
else
    echo "📦 数据卷已保留"
fi

# 询问是否清理镜像
echo ""
read -p "是否清理Docker镜像? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🧹 清理Docker镜像..."
    docker-compose down --rmi all
    echo "✅ Docker镜像已清理"
else
    echo "🖼️ Docker镜像已保留"
fi

echo ""
echo "🎉 CyberStroll系统停止完成!"