#!/bin/bash

# IP发现系统编译脚本

echo "=== IP发现系统编译脚本 ==="

# 设置变量
APP_NAME="ip-discovery"
VERSION=$(date +%Y%m%d_%H%M%S)

echo "应用名称: $APP_NAME"
echo "版本: $VERSION"

# 清理旧文件
echo "1. 清理旧文件..."
go clean -cache
rm -f $APP_NAME $APP_NAME-* 

# 更新依赖
echo "2. 更新依赖..."
go mod tidy

# 编译当前平台
echo "3. 编译当前平台版本..."
go build -o $APP_NAME main.go
if [ $? -eq 0 ]; then
    echo "✅ 当前平台编译成功"
    ls -lh $APP_NAME
else
    echo "❌ 当前平台编译失败"
    exit 1
fi

# 编译Linux版本
echo "4. 编译Linux版本..."
GOOS=linux GOARCH=amd64 go build -o ${APP_NAME}-linux-amd64 main.go
if [ $? -eq 0 ]; then
    echo "✅ Linux版本编译成功"
    ls -lh ${APP_NAME}-linux-amd64
else
    echo "❌ Linux版本编译失败"
fi

echo ""
echo "=== 编译完成 ==="
echo "生成的文件:"
ls -lh ${APP_NAME}*

echo ""
echo "=== 使用方法 ==="
echo "1. 测试系统: ./$APP_NAME test"
echo "2. 获取数据: ./$APP_NAME fetch"
echo "3. 扫描IP段: ./$APP_NAME scan --cidr \"8.8.8.0/24\""
echo "4. 查看统计: ./$APP_NAME stats"
