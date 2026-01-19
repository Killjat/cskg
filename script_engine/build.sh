#!/bin/bash

# Script Engine 构建脚本

echo "🚀 开始构建 Script Engine..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 未找到Go环境，请先安装Go"
    exit 1
fi

echo "✅ Go环境检查通过"

# 清理旧的构建文件
echo "🧹 清理旧的构建文件..."
rm -f script_engine script_engine.exe

# 下载依赖
echo "📦 下载依赖包..."
go mod tidy

# 构建项目
echo "🔨 编译项目..."
go build -o script_engine .

if [ $? -eq 0 ]; then
    echo "✅ 构建成功！"
    echo "📁 可执行文件: ./script_engine"
    echo ""
    echo "🎯 使用示例:"
    echo "  ./script_engine -help"
    echo "  ./script_engine -list-scripts"
    echo "  ./script_engine -target 192.168.1.100:502 -protocol modbus"
    echo ""
else
    echo "❌ 构建失败"
    exit 1
fi