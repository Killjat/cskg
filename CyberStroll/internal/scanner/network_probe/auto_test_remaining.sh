#!/bin/bash

# 自动测试剩余协议脚本
echo "🚀 开始自动测试剩余的25个协议..."
echo "测试时间: $(date)"
echo "========================================"

# 定义剩余需要测试的协议
protocols=(
    # 高优先级 - 工控协议
    "dnp3"
    "bacnet" 
    "opcua"
    "s7"
    
    # 高优先级 - 网络基础协议
    "http"
    "https"
    "ssh"
    "ftp"
    "smtp"
    "dns"
    "snmp"
    
    # 中优先级 - 数据库协议
    "sqlserver"
    "cassandra"
    "neo4j"
    
    # 中优先级 - IoT协议
    "coap"
    "lorawan"
    
    # 中优先级 - 企业协议
    "ntp"
    
    # 中优先级 - 安全协议
    "wireguard"
    
    # 中优先级 - 摄像头协议
    "rtsp"
    "hikvision"
    
    # 低优先级 - 邮件协议
    "pop3"
    "imap"
)

# 创建结果文件
result_file="auto_test_results_$(date +%Y%m%d_%H%M%S).txt"
echo "测试结果将保存到: $result_file"

# 测试每个协议
total=${#protocols[@]}
current=0

for protocol in "${protocols[@]}"; do
    current=$((current + 1))
    echo ""
    echo "[$current/$total] 正在测试协议: $protocol"
    echo "========================================"
    
    # 执行测试并记录结果
    echo "[$current/$total] 测试协议: $protocol - $(date)" >> "$result_file"
    
    timeout 120 ./network_probe -fofa-test -fofa-protocol "$protocol" -verbose >> "$result_file" 2>&1
    
    if [ $? -eq 0 ]; then
        echo "✅ $protocol 测试完成"
        echo "✅ $protocol 测试完成 - $(date)" >> "$result_file"
    elif [ $? -eq 124 ]; then
        echo "⏰ $protocol 测试超时"
        echo "⏰ $protocol 测试超时 - $(date)" >> "$result_file"
    else
        echo "❌ $protocol 测试失败"
        echo "❌ $protocol 测试失败 - $(date)" >> "$result_file"
    fi
    
    echo "----------------------------------------" >> "$result_file"
    
    # 添加短暂延迟避免API限制
    sleep 2
done

echo ""
echo "🎉 所有协议测试完成！"
echo "📊 测试结果已保存到: $result_file"
echo "完成时间: $(date)"

# 生成测试摘要
echo ""
echo "📋 测试摘要:"
echo "总测试协议数: $total"
echo "成功测试: $(grep -c "✅.*测试完成" "$result_file")"
echo "超时测试: $(grep -c "⏰.*测试超时" "$result_file")"
echo "失败测试: $(grep -c "❌.*测试失败" "$result_file")"