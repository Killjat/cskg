#!/bin/bash

# CyberStroll应用停止脚本
# 停止所有CyberStroll节点

set -e

echo "🛑 停止CyberStroll应用节点..."

# 进入项目目录
cd "$(dirname "$0")/.."

echo "📁 当前工作目录: $(pwd)"

# 停止函数
stop_node() {
    local node_name=$1
    local pid_file="logs/${node_name}.pid"
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if ps -p $pid > /dev/null 2>&1; then
            echo "🛑 停止${node_name} (PID: $pid)..."
            kill -TERM $pid
            
            # 等待进程停止
            local count=0
            while ps -p $pid > /dev/null 2>&1 && [ $count -lt 10 ]; do
                sleep 1
                count=$((count + 1))
            done
            
            # 如果进程仍在运行，强制杀死
            if ps -p $pid > /dev/null 2>&1; then
                echo "⚠️ 强制停止${node_name}..."
                kill -KILL $pid
            fi
            
            echo "✅ ${node_name}已停止"
        else
            echo "⚠️ ${node_name}进程不存在 (PID: $pid)"
        fi
        rm -f "$pid_file"
    else
        echo "⚠️ ${node_name} PID文件不存在"
    fi
}

# 停止所有节点
stop_node "task_manager"
stop_node "scan_node"
stop_node "processor_node"
stop_node "search_node"
stop_node "enrichment_node"

# 清理其他可能的进程
echo "🧹 清理其他CyberStroll进程..."
pkill -f "task_manager" 2>/dev/null || true
pkill -f "scan_node" 2>/dev/null || true
pkill -f "processor_node" 2>/dev/null || true
pkill -f "search_node" 2>/dev/null || true
pkill -f "enrichment_node" 2>/dev/null || true

echo ""
echo "✅ 所有CyberStroll应用节点已停止"

# 询问是否清理日志
echo ""
read -p "是否清理日志文件? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🧹 清理日志文件..."
    rm -f logs/*.log
    rm -f logs/*.pid
    echo "✅ 日志文件已清理"
else
    echo "📝 日志文件已保留"
fi

echo ""
echo "🎉 CyberStroll应用停止完成!"