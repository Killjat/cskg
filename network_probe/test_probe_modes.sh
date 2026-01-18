#!/bin/bash

echo "🔍 网络探测引擎 - 探测模式测试"
echo "=================================="

# 编译程序
echo "📦 编译程序..."
go build -o network_probe .

if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi

echo "✅ 编译成功"
echo ""

# 测试目标
TARGET="httpbin.org:80"
NONSTANDARD_TARGET="httpbin.org:443"  # HTTPS服务

echo "🎯 测试目标: $TARGET"
echo "🎯 非标准端口测试: $NONSTANDARD_TARGET"
echo ""

# 测试1: 端口模式
echo "1️⃣ 测试端口模式 (port) - 仅使用端口相关探测"
echo "----------------------------------------"
./network_probe -target $TARGET -probe-mode port -timeout 5s
echo ""

# 测试2: 全面模式
echo "2️⃣ 测试全面模式 (all) - 使用所有探测包"
echo "----------------------------------------"
./network_probe -target $TARGET -probe-mode all -timeout 5s
echo ""

# 测试3: 智能模式
echo "3️⃣ 测试智能模式 (smart) - 优先常见探测"
echo "----------------------------------------"
./network_probe -target $TARGET -probe-mode smart -timeout 5s
echo ""

# 测试4: 非标准端口服务探测
echo "4️⃣ 测试非标准端口 - 443端口可能运行其他服务"
echo "----------------------------------------"
./network_probe -target $NONSTANDARD_TARGET -probe-mode all -timeout 5s
echo ""

# 测试5: 本地服务探测（如果有的话）
echo "5️⃣ 测试本地服务 - 22端口可能运行HTTP服务"
echo "----------------------------------------"
./network_probe -target localhost:22 -probe-mode all -timeout 3s
echo ""

echo "🏁 测试完成！"
echo ""
echo "💡 说明："
echo "- port模式：快速，仅探测端口相关协议"
echo "- all模式：全面，发送所有探测包，能发现非标准端口服务"
echo "- smart模式：平衡，优先探测常见协议"
echo ""
echo "🔍 非标准端口服务示例："
echo "- 22端口运行HTTP服务"
echo "- 80端口运行SSH服务"  
echo "- 8080端口运行数据库服务"
echo "- 443端口运行FTP服务"