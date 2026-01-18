#!/bin/bash

# Banner引擎测试脚本

echo "🔍 Banner指纹识别引擎测试"
echo "=========================="

# 编译程序
echo "📦 编译程序..."
go build -o banner_engine .

if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi

echo "✅ 编译成功"
echo ""

# 测试基本功能
echo "🧪 测试基本功能..."

echo "1. 测试Nginx识别:"
./banner_engine -banner "nginx/1.18.0"
echo ""

echo "2. 测试Apache识别:"
./banner_engine -banner "Apache/2.4.41 (Ubuntu)"
echo ""

echo "3. 测试SSH识别:"
./banner_engine -banner "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5"
echo ""

echo "4. 测试MySQL识别:"
./banner_engine -banner "5.7.34-0ubuntu0.18.04.1-log mysql_native_password"
echo ""

echo "5. 测试Redis识别:"
./banner_engine -banner "+PONG"
echo ""

echo "6. 测试JSON输出:"
./banner_engine -banner "nginx/1.18.0" -output json
echo ""

# 测试自定义规则
echo "🔧 测试自定义规则..."

# 创建规则目录
mkdir -p rules

# 复制示例规则
cp examples/custom_rules.json rules/

echo "7. 测试自定义规则加载:"
./banner_engine -banner "MyWebApp v2.1" -rules-dir rules
echo ""

echo "8. 测试未知Banner:"
./banner_engine -banner "UnknownService/1.0"
echo ""

echo "✅ 测试完成!"
echo ""
echo "💡 提示:"
echo "  - 使用 ./banner_engine -interactive 进入交互模式"
echo "  - 使用 ./banner_engine -help 查看所有选项"
echo "  - 查看 README.md 了解详细使用方法"