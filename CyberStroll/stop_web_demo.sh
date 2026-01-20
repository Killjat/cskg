#!/bin/bash

# CyberStroll Web界面演示停止脚本

echo "🛑 停止 CyberStroll Web界面演示..."

# 停止任务管理节点
if [ -f "logs/task_manager_demo.pid" ]; then
    PID=$(cat logs/task_manager_demo.pid)
    if kill -0 "$PID" 2>/dev/null; then
        echo "🔴 停止任务管理节点 (PID: $PID)"
        kill "$PID"
        sleep 2
        if kill -0 "$PID" 2>/dev/null; then
            echo "⚠️  强制停止任务管理节点"
            kill -9 "$PID"
        fi
    fi
    rm -f logs/task_manager_demo.pid
fi

# 停止搜索节点
if [ -f "logs/search_node_demo.pid" ]; then
    PID=$(cat logs/search_node_demo.pid)
    if kill -0 "$PID" 2>/dev/null; then
        echo "🔴 停止搜索节点 (PID: $PID)"
        kill "$PID"
        sleep 2
        if kill -0 "$PID" 2>/dev/null; then
            echo "⚠️  强制停止搜索节点"
            kill -9 "$PID"
        fi
    fi
    rm -f logs/search_node_demo.pid
fi

echo "✅ 所有演示服务已停止"