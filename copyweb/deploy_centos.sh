#!/bin/bash

# 网页克隆工具 CentOS 部署脚本
# 作者: CodeBuddy
# 日期: $(date +%Y-%m-%d)

# 配置变量
APP_NAME="web-clone-tool"
APP_DIR="/opt/${APP_NAME}"
PYTHON_VERSION="3.10"
PORT_RANGE="8000-9999"
EXCLUDE_PORT="5001"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}====================================="
echo -e "    网页克隆工具 CentOS 部署脚本"
echo -e "=====================================${NC}"

# 检查是否为 root 用户
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}错误: 请使用 root 用户运行此脚本${NC}"
    exit 1
fi

# 1. 更新系统并安装依赖
echo -e "${YELLOW}[1/5] 更新系统并安装必要依赖...${NC}"
yum update -y

# 安装 Python 3.10 和相关工具
yum install -y python3 python3-pip python3-devel gcc make wget curl

# 安装额外的系统依赖
yum install -y epel-release
yum install -y libxml2-devel libxslt-devel openssl-devel

# 2. 创建应用目录
echo -e "${YELLOW}[2/5] 创建应用目录...${NC}"
mkdir -p ${APP_DIR}
mkdir -p ${APP_DIR}/cloned_pages

# 3. 复制应用文件
echo -e "${YELLOW}[3/5] 复制应用文件...${NC}"
# 假设当前目录包含所有应用文件
cp -r *.py ${APP_DIR}/
cp -r requirements.txt ${APP_DIR}/

# 4. 安装 Python 依赖
echo -e "${YELLOW}[4/5] 安装 Python 依赖...${NC}"
cd ${APP_DIR}
pip3 install --upgrade pip
pip3 install -r requirements.txt

# 5. 创建 systemd 服务
echo -e "${YELLOW}[5/5] 创建 systemd 服务...${NC}"
cat > /etc/systemd/system/${APP_NAME}.service << EOF
[Unit]
Description=Web Clone Tool - High Availability Version
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${APP_DIR}
ExecStart=/usr/bin/python3 ${APP_DIR}/app.py
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 6. 重新加载 systemd 并启动服务
systemctl daemon-reload
systemctl enable ${APP_NAME}.service
systemctl start ${APP_NAME}.service

# 7. 检查服务状态
echo -e "${YELLOW}检查服务状态...${NC}"
sleep 3
systemctl status ${APP_NAME}.service --no-pager

# 8. 显示部署结果
echo -e "${GREEN}====================================="
echo -e "    部署完成！"
echo -e "=====================================${NC}"
echo -e "${GREEN}应用目录:${NC} ${APP_DIR}"
echo -e "${GREEN}服务名称:${NC} ${APP_NAME}.service"
echo -e "${GREEN}查看日志:${NC} journalctl -u ${APP_NAME}.service -f"
echo -e "${GREEN}启动服务:${NC} systemctl start ${APP_NAME}.service"
echo -e "${GREEN}停止服务:${NC} systemctl stop ${APP_NAME}.service"
echo -e "${GREEN}重启服务:${NC} systemctl restart ${APP_NAME}.service"
echo -e "${GREEN}查看状态:${NC} systemctl status ${APP_NAME}.service"
echo -e "${GREEN}=====================================${NC}"
echo -e "${YELLOW}注意: 服务将在随机端口 ${PORT_RANGE} 上运行，避免使用 ${EXCLUDE_PORT}${NC}"
echo -e "${YELLOW}可以通过查看日志获取实际运行端口: journalctl -u ${APP_NAME}.service -n 20${NC}"