#!/bin/bash

# FOFA API测试运行脚本

echo "🚀 FOFA API测试工具"
echo "=================="

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go"
    exit 1
fi

# 检查配置文件
if [ ! -f "fofa_config.json" ]; then
    echo "❌ 配置文件 fofa_config.json 不存在"
    echo ""
    if [ -f "fofa_config.json.example" ]; then
        echo "📋 发现示例配置文件，正在复制..."
        cp fofa_config.json.example fofa_config.json
        echo "✅ 已创建 fofa_config.json 文件"
        echo ""
        echo "⚠️  请编辑 fofa_config.json 文件，填入你的FOFA凭据:"
        echo "   - email: 你的FOFA邮箱"
        echo "   - key: 你的FOFA API Key"
        echo ""
        echo "然后重新运行此脚本"
        exit 1
    else
        echo "请创建配置文件 fofa_config.json，内容如下:"
        echo '{'
        echo '  "email": "your_email@example.com",'
        echo '  "key": "your_fofa_api_key",'
        echo '  "base_url": "https://fofa.info/api/v1/search/all"'
        echo '}'
        exit 1
    fi
fi

# 检查配置文件是否为示例内容
if grep -q "your_email@example.com" fofa_config.json; then
    echo "⚠️  检测到配置文件使用示例内容"
    echo "请编辑 fofa_config.json 文件，填入真实的FOFA凭据"
    exit 1
fi

# 初始化Go模块
if [ ! -f "go.mod" ]; then
    echo "📦 初始化Go模块..."
    go mod init fofa-api-test
fi

# 下载依赖
echo "📦 下载依赖..."
go mod tidy

# 运行测试
echo "🏃 开始运行测试..."
echo ""

go run main.go

echo ""
echo "✅ 测试完成！"
echo "📁 查看当前目录下的 fofa_api_test_result_*.json 文件获取详细结果"