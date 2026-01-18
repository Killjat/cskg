#!/bin/bash

echo "=== 节点管理系统测试部署 ==="

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "错误: Docker未运行，请先启动Docker"
    exit 1
fi

# 1. 启动NATS服务器
echo "1. 启动NATS服务器..."
docker run -d --name nats-server -p 4222:4222 -p 8222:8222 nats:latest
if [ $? -ne 0 ]; then
    echo "错误: 无法启动NATS服务器"
    exit 1
fi

# 2. 等待NATS服务器启动
sleep 3

echo "2. 正在启动3个客户端..."

# 3. 启动客户端1
# 使用主机网络访问NATS服务器
echo "   - 启动客户端1 (node-1)..."
./nodemanage client -n node-1 -c config.yaml > client1.log 2>&1 &
CLIENT1_PID=$!

# 4. 启动客户端2
echo "   - 启动客户端2 (node-2)..."
./nodemanage client -n node-2 -c config.yaml > client2.log 2>&1 &
CLIENT2_PID=$!

# 5. 启动客户端3
echo "   - 启动客户端3 (node-3)..."
./nodemanage client -n node-3 -c config.yaml > client3.log 2>&1 &
CLIENT3_PID=$!

# 6. 等待客户端启动
sleep 5

# 7. 检查进程状态
echo "3. 检查进程状态..."
ps aux | grep -E "nodemanage client" | grep -v grep

# 8. 显示日志
echo "4. 客户端1日志片段:"
tail -5 client1.log

echo "5. 客户端2日志片段:"
tail -5 client2.log

echo "6. 客户端3日志片段:"
tail -5 client3.log

echo "7. 启动服务端..."
./nodemanage server -c config.yaml > server.log 2>&1 &
SERVER_PID=$!
sleep 3

echo "8. 服务端日志片段:"
tail -10 server.log

echo "=== 部署完成 ==="
echo "NATS服务器正在运行: docker logs nats-server"
echo "客户端日志: client1.log, client2.log, client3.log"
echo "服务端日志: server.log"
echo "使用Ctrl+C停止所有进程"

# 9. 清理函数
cleanup() {
    echo "\n=== 清理资源 ==="
    kill $CLIENT1_PID $CLIENT2_PID $CLIENT3_PID $SERVER_PID 2>/dev/null
    docker stop nats-server >/dev/null 2>&1
    docker rm nats-server >/dev/null 2>&1
    rm -f client*.log server.log
    echo "清理完成"
}

# 10. 注册清理函数
trap cleanup EXIT

# 11. 保持运行
echo "按Enter键退出并清理资源..."
read
