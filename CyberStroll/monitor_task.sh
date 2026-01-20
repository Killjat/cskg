#!/bin/bash

TASK_ID="3129d3a4-15d6-44a0-bca9-f3a27b29ac74"

echo "🚀 监控应用识别任务进度..."
echo "任务ID: $TASK_ID"
echo "目标IP数量: 23"
echo "================================"

while true; do
    # 获取任务状态
    STATUS=$(curl -s "http://localhost:8088/api/tasks/status?task_id=$TASK_ID")
    
    # 解析状态信息
    TASK_STATUS=$(echo $STATUS | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
    PROGRESS=$(echo $STATUS | grep -o '"progress":[0-9]*' | cut -d':' -f2)
    COMPLETED=$(echo $STATUS | grep -o '"completed_count":[0-9]*' | cut -d':' -f2)
    FAILED=$(echo $STATUS | grep -o '"failed_count":[0-9]*' | cut -d':' -f2)
    
    # 获取系统统计
    STATS=$(curl -s "http://localhost:8088/api/stats")
    TOTAL_TASKS=$(echo $STATS | grep -o '"total_tasks":[0-9]*' | cut -d':' -f2)
    
    echo "$(date '+%H:%M:%S') - 状态: $TASK_STATUS | 进度: $PROGRESS% | 已完成: $COMPLETED | 失败: $FAILED | 总任务: $TOTAL_TASKS"
    
    if [ "$TASK_STATUS" = "completed" ] || [ "$TASK_STATUS" = "failed" ]; then
        echo "✅ 任务已完成！"
        break
    fi
    
    sleep 5
done