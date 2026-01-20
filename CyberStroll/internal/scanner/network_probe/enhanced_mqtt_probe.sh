#!/bin/bash

echo "🔍 增强型MQTT探测工具"
echo "=========================="

# 测试IP列表
IPS=(
    "59.106.209.190"
    "18.176.255.164" 
    "27.231.209.9"
    "116.91.193.85"
    "110.160.202.123"
    "104.41.184.83"
)

# MQTT相关端口
MQTT_PORTS=(1883 8883 1884 8884 8080 9001)

echo "📋 测试目标: ${#IPS[@]} 个IP"
echo "🔌 测试端口: ${MQTT_PORTS[*]}"
echo ""

success_count=0
total_tests=0

for ip in "${IPS[@]}"; do
    echo "🎯 测试IP: $ip"
    
    # 首先进行ping测试
    if ping -c 1 -W 3 "$ip" >/dev/null 2>&1; then
        echo "   ✅ Ping成功"
        
        # 测试各个MQTT端口
        for port in "${MQTT_PORTS[@]}"; do
            echo -n "   🔍 端口 $port: "
            total_tests=$((total_tests + 1))
            
            # 使用nc进行端口测试
            if timeout 3 nc -z "$ip" "$port" 2>/dev/null; then
                echo "开放 ✅"
                success_count=$((success_count + 1))
                
                # 如果端口开放，使用我们的工具进行详细探测
                echo "      🔬 详细探测:"
                ./network_probe -target "$ip:$port" -probe-mode smart -timeout 5s 2>/dev/null | grep -E "(✅|📄|🏷️)" | sed 's/^/         /'
                
            else
                echo "关闭 ❌"
            fi
        done
    else
        echo "   ❌ Ping失败 - 主机不可达"
        # 即使ping失败也测试端口（有些主机禁ping）
        for port in "${MQTT_PORTS[@]}"; do
            echo -n "   🔍 端口 $port (无ping): "
            total_tests=$((total_tests + 1))
            
            if timeout 5 nc -z "$ip" "$port" 2>/dev/null; then
                echo "开放 ✅"
                success_count=$((success_count + 1))
            else
                echo "关闭 ❌"
            fi
        done
    fi
    
    echo ""
done

echo "📊 测试结果统计:"
echo "=================="
echo "总测试数: $total_tests"
echo "开放端口: $success_count"
echo "成功率: $(( success_count * 100 / total_tests ))%"

if [ $success_count -gt 0 ]; then
    echo ""
    echo "✅ 发现 $success_count 个开放的端口!"
    echo "💡 建议使用以下命令进行详细探测:"
    echo "   ./network_probe -target IP:PORT -probe-mode all -verbose"
else
    echo ""
    echo "⚠️  未发现开放的MQTT端口"
    echo "🔧 可能的解决方案:"
    echo "   1. 检查网络连接"
    echo "   2. 尝试使用VPN"
    echo "   3. 测试本地MQTT服务器"
    echo "   4. 检查防火墙设置"
fi