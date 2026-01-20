#!/bin/bash

# FOFA协议检测能力测试脚本

echo "🔍 FOFA协议检测能力测试工具"
echo "================================"

# 检查配置文件
if [ ! -f "fofa_config.json" ]; then
    echo "⚠️  未找到配置文件 fofa_config.json"
    echo "📝 正在创建示例配置文件..."
    cp fofa_config.json.example fofa_config.json
    echo "✅ 已创建 fofa_config.json"
    echo ""
    echo "请编辑 fofa_config.json 文件，填入您的FOFA凭据:"
    echo "  - email: 您的FOFA邮箱"
    echo "  - key: 您的FOFA API Key"
    echo ""
    echo "然后重新运行此脚本"
    exit 1
fi

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 未找到Go环境，请先安装Go"
    exit 1
fi

echo "✅ 环境检查通过"
echo ""

# 显示菜单
echo "请选择测试模式:"
echo "1. 测试所有协议 (推荐)"
echo "2. 测试单个协议"
echo "3. 快速测试 (仅测试常见协议)"
echo "4. 显示支持的协议列表"
echo "5. 退出"
echo ""

read -p "请输入选择 (1-5): " choice

case $choice in
    1)
        echo "🚀 开始测试所有协议..."
        go run fofa_test_main.go fofa_tester.go -verbose
        ;;
    2)
        echo ""
        echo "支持的协议:"
        echo "工控: modbus, dnp3, bacnet, opcua, s7"
        echo "数据库: mysql, postgresql, redis, sqlserver, oracle, mongodb"
        echo "IoT: mqtt, coap, lorawan, amqp"
        echo "网络: http, https, ssh, ftp, smtp"
        echo ""
        read -p "请输入要测试的协议名称: " protocol
        echo "🎯 测试协议: $protocol"
        go run fofa_test_main.go fofa_tester.go -protocol "$protocol" -verbose
        ;;
    3)
        echo "🚀 快速测试常见协议..."
        protocols=("http" "https" "ssh" "mysql" "redis" "mongodb")
        for protocol in "${protocols[@]}"; do
            echo "测试 $protocol..."
            go run fofa_test_main.go fofa_tester.go -protocol "$protocol"
            echo ""
        done
        ;;
    4)
        echo "📋 支持的协议列表:"
        go run fofa_test_main.go fofa_tester.go -help | grep -A 20 "支持的协议:"
        ;;
    5)
        echo "👋 退出"
        exit 0
        ;;
    *)
        echo "❌ 无效选择"
        exit 1
        ;;
esac

echo ""
echo "✅ 测试完成！"
echo "📊 查看详细报告: ls -la fofa_test_report_*.json"