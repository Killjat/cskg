#!/bin/bash

echo "=== ImageGPS 图片GPS地理位置提取系统 ==="
echo ""
echo "正在安装依赖..."
go mod tidy

if [ $? -ne 0 ]; then
    echo "依赖安装失败，请检查网络连接或Go环境"
    exit 1
fi

echo ""
echo "依赖安装完成，正在启动服务..."
echo ""

go run main.go
