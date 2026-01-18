#!/bin/bash

echo "=== 测试任务发送 ==="

echo "1. 向node-1发送任务..."
./nodemanage task node-1 ls -la

if [ $? -ne 0 ]; then
    echo "错误: 无法发送任务到node-1"
    exit 1
fi

echo "2. 向所有节点广播任务..."
./nodemanage task "" ls -la

if [ $? -ne 0 ]; then
    echo "错误: 无法广播任务"
    exit 1
fi

echo "3. 查看节点状态..."
./nodemanage status node-1
./nodemanage status node-2
./nodemanage status node-3

echo "4. 列出所有节点..."
./nodemanage list

echo "=== 任务发送测试完成 ==="
echo "请查看服务端日志以获取任务执行结果:"
echo "tail -f server.log"
