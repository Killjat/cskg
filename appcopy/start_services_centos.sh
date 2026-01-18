#!/bin/bash

# CentOS系统服务启动脚本
# 功能：停止所有相关服务，检查端口，启动Web服务管理器

echo "=== CentOS系统服务启动脚本 ==="
echo "执行时间：$(date)"
echo ""

# 1. 停止所有运行的Python服务进程
echo "1. 停止所有运行的Python服务进程..."
pkill -f "python3.*\.py"
pkill -f "python.*\.py"
sleep 2
echo "   ✓ Python服务进程已停止"
echo ""

# 2. 停止系统服务（Redis、MySQL/MariaDB）
echo "2. 停止系统服务..."

# 检查Redis服务状态并停止
echo "   检查Redis服务..."
if systemctl is-active --quiet redis; then
    echo "   停止Redis服务..."
    sudo systemctl stop redis
    sudo systemctl disable redis 2>/dev/null
    sleep 2
    echo "   ✓ Redis服务已停止"
else
    echo "   Redis服务未运行"
fi

# 检查MySQL/MariaDB服务状态并停止
echo "   检查MySQL/MariaDB服务..."
if systemctl is-active --quiet mysqld; then
    echo "   停止MySQL服务..."
    sudo systemctl stop mysqld
    sudo systemctl disable mysqld 2>/dev/null
    sleep 2
    echo "   ✓ MySQL服务已停止"
elif systemctl is-active --quiet mariadb; then
    echo "   停止MariaDB服务..."
    sudo systemctl stop mariadb
    sudo systemctl disable mariadb 2>/dev/null
    sleep 2
    echo "   ✓ MariaDB服务已停止"
else
    echo "   MySQL/MariaDB服务未运行"
fi
echo ""

# 3. 检查关键端口占用情况
echo "3. 检查关键端口占用情况..."
PORTS="10000 9999 8888 8080 5001 5000 502 3306 6379"
for PORT in $PORTS; do
    PORT_STATUS=$(ss -tuln | grep ":$PORT " | head -1)
    if [ -n "$PORT_STATUS" ]; then
        # 获取占用端口的进程ID
        PID=$(ss -tulnp | grep ":$PORT " | head -1 | awk '{print $7}' | cut -d',' -f2)
        echo "   端口 $PORT 被占用，进程ID：$PID"
        # 尝试杀死占用端口的进程
        if [ -n "$PID" ]; then
            sudo kill $PID 2>/dev/null
            sleep 1
            # 再次检查端口是否释放
            PORT_STATUS=$(ss -tuln | grep ":$PORT " | head -1)
            if [ -n "$PORT_STATUS" ]; then
                echo "   无法释放端口 $PORT"
            else
                echo "   ✓ 端口 $PORT 已释放"
            fi
        else
            echo "   无法获取占用端口的进程ID"
        fi
    else
        echo "   端口 $PORT 可用"
    fi
done
echo ""

# 4. 启动Web服务管理器
echo "4. 启动Web服务管理器..."
# 检查是否已存在web_service_manager.py进程
WS_PID=$(ps aux | grep web_service_manager.py | grep -v grep | awk '{print $2}')
if [ -n "$WS_PID" ]; then
    echo "   杀死现有Web服务管理器进程 $WS_PID..."
    kill $WS_PID
    sleep 2
fi

# 启动Web服务管理器
nohup python3 web_service_manager.py > web_service.log 2>&1 &
sleep 3

# 检查是否启动成功
WS_PID=$(ps aux | grep web_service_manager.py | grep -v grep | awk '{print $2}')
if [ -n "$WS_PID" ]; then
    echo "   ✓ Web服务管理器已成功启动，进程ID：$WS_PID"
    echo "   日志文件：web_service.log"
    echo "   可以使用以下命令查看日志：tail -f web_service.log"
    echo ""
    echo "=== 服务启动完成 ==="
    echo "Web服务管理器访问地址："
    # 获取公网IP
    PUBLIC_IP=$(curl -s icanhazip.com 2>/dev/null || echo "无法获取公网IP")
    # 获取内网IP
    INTERNAL_IP=$(hostname -I | awk '{print $1}')
    echo "- 公网访问：http://$PUBLIC_IP:9999"
    echo "- 内网访问：http://$INTERNAL_IP:9999"
    echo "- 本地访问：http://localhost:9999"
else
    echo "   ✗ Web服务管理器启动失败，请查看日志文件：web_service.log"
fi

echo ""
echo "脚本执行完成！"
