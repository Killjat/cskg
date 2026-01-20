#!/bin/bash

# CyberStroll 构建脚本

set -e

echo "🚀 开始构建 CyberStroll..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go 1.21+"
    exit 1
fi

# 检查Go版本
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.21"

if ! printf '%s\n%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V -C; then
    echo "❌ Go版本过低，需要 $REQUIRED_VERSION+，当前版本: $GO_VERSION"
    exit 1
fi

echo "✅ Go版本检查通过: $GO_VERSION"

# 创建必要的目录
echo "📁 创建目录结构..."
mkdir -p bin
mkdir -p logs
mkdir -p configs
mkdir -p internal/scanner
mkdir -p internal/kafka
mkdir -p internal/storage
mkdir -p internal/state
mkdir -p pkg/config
mkdir -p pkg/models
mkdir -p pkg/utils
mkdir -p web/static
mkdir -p web/templates

# 复制现有模块 (如果存在)
echo "📦 复制依赖模块..."

# 复制network_probe模块
if [ -d "../network_probe" ]; then
    echo "  复制 network_probe 模块..."
    mkdir -p internal/scanner/network_probe
    cp -r ../network_probe/* internal/scanner/network_probe/ 2>/dev/null || echo "    network_probe 复制完成"
else
    echo "  ⚠️  network_probe 模块不存在，跳过"
fi

# 复制rule_engine模块
if [ -d "../rule_engine" ]; then
    echo "  复制 rule_engine 模块..."
    mkdir -p internal/rules
    cp -r ../rule_engine/* internal/rules/ 2>/dev/null || echo "    rule_engine 复制完成"
else
    echo "  ⚠️  rule_engine 模块不存在，跳过"
fi

# 复制script_engine模块
if [ -d "../script_engine" ]; then
    echo "  复制 script_engine 模块..."
    mkdir -p internal/scripts
    cp -r ../script_engine/* internal/scripts/ 2>/dev/null || echo "    script_engine 复制完成"
else
    echo "  ⚠️  script_engine 模块不存在，跳过"
fi

# 复制servicefingerprint模块
if [ -d "../servicefingerprint" ]; then
    echo "  复制 servicefingerprint 模块..."
    mkdir -p internal/fingerprint
    cp -r ../servicefingerprint/* internal/fingerprint/ 2>/dev/null || echo "    servicefingerprint 复制完成"
else
    echo "  ⚠️  servicefingerprint 模块不存在，跳过"
fi

# 下载依赖
echo "📥 下载Go模块依赖..."
go mod tidy
go mod download

# 构建各个组件
echo "🔨 构建扫描节点..."
go build -ldflags="-w -s" -o bin/scan_node cmd/scan_node/main.go
if [ $? -eq 0 ]; then
    echo "✅ 扫描节点构建成功"
else
    echo "❌ 扫描节点构建失败"
    exit 1
fi

echo "🔨 构建任务管理节点..."
if [ -f "cmd/task_manager/main.go" ]; then
    go build -ldflags="-w -s" -o bin/task_manager cmd/task_manager/main.go
    if [ $? -eq 0 ]; then
        echo "✅ 任务管理节点构建成功"
    else
        echo "❌ 任务管理节点构建失败"
    fi
else
    echo "⚠️  任务管理节点代码未找到，跳过构建"
fi

echo "🔨 构建处理节点..."
if [ -f "cmd/processor_node/main.go" ]; then
    go build -ldflags="-w -s" -o bin/processor_node cmd/processor_node/main.go
    if [ $? -eq 0 ]; then
        echo "✅ 处理节点构建成功"
    else
        echo "❌ 处理节点构建失败"
    fi
else
    echo "⚠️  处理节点代码未找到，跳过构建"
fi

echo "🔨 构建搜索节点..."
if [ -f "cmd/search_node/main.go" ]; then
    go build -ldflags="-w -s" -o bin/search_node cmd/search_node/main.go
    if [ $? -eq 0 ]; then
        echo "✅ 搜索节点构建成功"
    else
        echo "❌ 搜索节点构建失败"
    fi
else
    echo "⚠️  搜索节点代码未找到，跳过构建"
fi

echo "🔨 构建网站数据富化节点..."
if [ -f "cmd/enrichment_node/main.go" ]; then
    go build -ldflags="-w -s" -o bin/enrichment_node cmd/enrichment_node/main.go
    if [ $? -eq 0 ]; then
        echo "✅ 网站数据富化节点构建成功"
    else
        echo "❌ 网站数据富化节点构建失败"
    fi
else
    echo "⚠️  网站数据富化节点代码未找到，跳过构建"
fi

# 设置执行权限
echo "🔐 设置执行权限..."
chmod +x bin/*

# 显示构建结果
echo ""
echo "🎉 构建完成！"
echo "📋 构建产物:"
ls -la bin/

echo ""
echo "📖 使用说明:"
echo "  启动扫描节点:       ./bin/scan_node --config configs/scan_node.yaml"
echo "  启动任务管理节点:   ./bin/task_manager --config configs/task_manager.yaml"
echo "  启动处理节点:       ./bin/processor_node --config configs/processor_node.yaml"
echo "  启动搜索节点:       ./bin/search_node --config configs/search_node.yaml"
echo "  启动网站富化节点:   ./bin/enrichment_node --config configs/enrichment_node.yaml"
echo ""
echo "  测试模式:         ./bin/scan_node --test"
echo "  查看帮助:         ./bin/scan_node --help"

# 检查配置文件
echo ""
echo "📝 配置文件检查:"
if [ -f "configs/scan_node.yaml" ]; then
    echo "  ✅ configs/scan_node.yaml"
else
    echo "  ❌ configs/scan_node.yaml (缺失)"
fi

# 检查依赖服务
echo ""
echo "🔍 依赖服务检查:"
echo "  请确保以下服务已启动:"
echo "    - Kafka (localhost:9092)"
echo "    - MongoDB (localhost:27017)"
echo "    - Elasticsearch (localhost:9200)"

echo ""
echo "✨ 构建脚本执行完成！"