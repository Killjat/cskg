#!/bin/bash

# 读取mqttip.txt文件的IP地址
IP_LIST=($(cat mqttip.txt))

# 检查IP列表是否为空
if [ ${#IP_LIST[@]} -eq 0 ]; then
    echo "错误：mqttip.txt文件中没有IP地址"
    exit 1
fi

echo "=== MQTT 扫描任务管理 ==="
echo "总共 ${#IP_LIST[@]} 个IP地址"

# 配置参数
GROUP_COUNT=20
PROTOCOL="tcp"
PORT_RANGE="1883,8883,8080,8081"
SCAN_TYPE="port_scan"

# 计算每组IP数量
totalIPs=${#IP_LIST[@]}
sizePerGroup=$((totalIPs / GROUP_COUNT))
remainder=$((totalIPs % GROUP_COUNT))

echo "分成 $GROUP_COUNT 个任务组"
echo "每组大约 $sizePerGroup 个IP，剩余 $remainder 个IP"
echo

# 创建任务目录
mkdir -p mqtt_scan_results

# 分组并下发任务
startIndex=0
for ((i=0; i<GROUP_COUNT; i++)); do
    # 计算当前组的IP数量
    currentSize=$sizePerGroup
    if [ $i -lt $remainder ]; then
        currentSize=$((currentSize + 1))
    fi

    # 获取当前组的IP列表
    endIndex=$((startIndex + currentSize))
    if [ $endIndex -gt $totalIPs ]; then
        endIndex=$totalIPs
    fi

    groupIPs=(${IP_LIST[@]:$startIndex:$currentSize})
    startIndex=$endIndex

    # 将IP列表转换为逗号分隔的字符串
    ipStr=$(IFS=,; echo "${groupIPs[*]}")

    echo "正在下发任务 $((i+1))/$GROUP_COUNT，包含 ${#groupIPs[@]} 个IP..."
    echo "IP范围：${groupIPs[0]} - ${groupIPs[-1]}"
    
    # 构建命令
    cmd="go run cmd/task_manager/main.go -targets $ipStr -type $SCAN_TYPE -protocol $PROTOCOL -port $PORT_RANGE -system"
    
    # 执行命令并保存结果
    $cmd > mqtt_scan_results/task_$((i+1)).log 2>&1
    
    # 检查命令执行结果
    if [ $? -eq 0 ]; then
        echo "任务 $((i+1)) 下发成功！"
    else
        echo "任务 $((i+1)) 下发失败！查看 mqtt_scan_results/task_$((i+1)).log 获取详细信息"
    fi
    
    echo
    # 短暂延迟，避免系统过载
    sleep 1
done

echo "所有任务下发完成！"
echo "任务结果保存在 mqtt_scan_results/ 目录下"
echo
