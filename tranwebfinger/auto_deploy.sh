#!/bin/bash

# 自动化部署脚本，用于在Linux服务器上部署自进化Wappalyzer系统

# 配置参数
PROJECT_NAME="self_evolving_wappalyzer"
VERSION="$(date +%Y%m%d)"
INSTALL_DIR="/opt/$PROJECT_NAME"
USER="root"
GROUP="root"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== 自进化Wappalyzer系统自动化部署脚本 ===${NC}"

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}错误：请以root用户身份运行此脚本${NC}"
    exit 1
fi

# 检查操作系统
echo -e "${YELLOW}1. 检查操作系统...${NC}"
if [ -f /etc/redhat-release ]; then
    OS="CentOS/RHEL"
    PKG_MANAGER="yum"
elif [ -f /etc/debian_version ]; then
    OS="Debian/Ubuntu"
    PKG_MANAGER="apt-get"
else
    echo -e "${RED}错误：不支持的操作系统${NC}"
    exit 1
fi
echo -e "${GREEN}   操作系统：$OS${NC}"

# 更新系统包
echo -e "${YELLOW}2. 更新系统包...${NC}"
if [ "$PKG_MANAGER" = "yum" ]; then
    yum update -y
elif [ "$PKG_MANAGER" = "apt-get" ]; then
    apt-get update -y
fi
echo -e "${GREEN}   系统包更新完成${NC}"

# 安装必要的系统依赖
echo -e "${YELLOW}3. 安装系统依赖...${NC}"
if [ "$PKG_MANAGER" = "yum" ]; then
    yum install -y python3 python3-pip python3-devel gcc
elif [ "$PKG_MANAGER" = "apt-get" ]; then
    apt-get install -y python3 python3-pip python3-dev gcc
fi
echo -e "${GREEN}   系统依赖安装完成${NC}"

# 创建安装目录
echo -e "${YELLOW}4. 创建安装目录...${NC}"
mkdir -p $INSTALL_DIR
echo -e "${GREEN}   安装目录：$INSTALL_DIR${NC}"

# 检查是否存在压缩包
if [ -f "$PROJECT_NAME-$VERSION.tar.gz" ]; then
    echo -e "${YELLOW}5. 使用本地压缩包进行安装...${NC}"
    cp "$PROJECT_NAME-$VERSION.tar.gz" $INSTALL_DIR/
    cd $INSTALL_DIR
    tar -xzf "$PROJECT_NAME-$VERSION.tar.gz"
    mv "$PROJECT_NAME-$VERSION"/* .
    rm -rf "$PROJECT_NAME-$VERSION" "$PROJECT_NAME-$VERSION.tar.gz"
elif [ -n "$1" ]; then
    echo -e "${YELLOW}5. 从URL下载压缩包...${NC}"
    cd $INSTALL_DIR
    wget -O "$PROJECT_NAME-$VERSION.tar.gz" "$1"
    tar -xzf "$PROJECT_NAME-$VERSION.tar.gz"
    mv "$PROJECT_NAME-$VERSION"/* .
    rm -rf "$PROJECT_NAME-$VERSION" "$PROJECT_NAME-$VERSION.tar.gz"
else
    echo -e "${RED}错误：未提供安装包，请将$PROJECT_NAME-$VERSION.tar.gz放在当前目录或提供下载URL${NC}"
    exit 1
fi
echo -e "${GREEN}   安装包解压完成${NC}"

# 设置文件权限
echo -e "${YELLOW}6. 设置文件权限...${NC}"
chown -R $USER:$GROUP $INSTALL_DIR
chmod +x $INSTALL_DIR/*.sh
echo -e "${GREEN}   文件权限设置完成${NC}"

# 安装Python依赖
echo -e "${YELLOW}7. 安装Python依赖...${NC}"
cd $INSTALL_DIR
if [ -f "install_deps.sh" ]; then
    ./install_deps.sh
else
    pip3 install requests scikit-learn numpy flask
fi
echo -e "${GREEN}   Python依赖安装完成${NC}"

# 创建系统服务
echo -e "${YELLOW}8. 创建系统服务...${NC}"

# 创建Web仪表盘系统服务
cat > /etc/systemd/system/wappalyzer-dashboard.service << 'EOF'
[Unit]
Description=Self Evolving Wappalyzer Dashboard
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/self_evolving_wappalyzer
ExecStart=/usr/bin/python3 /opt/self_evolving_wappalyzer/web_dashboard.py
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 重新加载系统服务
systemctl daemon-reload

# 启用服务（可选）
echo -e "${YELLOW}9. 配置服务...${NC}"
echo -e "${GREEN}   系统服务已创建，但未自动启用${NC}"
echo -e "${GREEN}   要手动启动Web仪表盘服务：systemctl start wappalyzer-dashboard${NC}"
echo -e "${GREEN}   要设置Web仪表盘服务开机自启：systemctl enable wappalyzer-dashboard${NC}"

# 显示安装完成信息
echo -e "${GREEN}=== 自进化Wappalyzer系统部署完成 ===${NC}"
echo -e "${GREEN}安装目录：${NC}$INSTALL_DIR"
echo -e "${GREEN}Web仪表盘：${NC}http://localhost:5001"
echo -e "${GREEN}启动命令：${NC}cd $INSTALL_DIR && ./run.sh"
echo -e "${GREEN}服务管理：${NC}"
echo -e "${GREEN}   启动Web仪表盘：systemctl start wappalyzer-dashboard${NC}"
echo -e "${GREEN}   停止Web仪表盘：systemctl stop wappalyzer-dashboard${NC}"
echo -e "${GREEN}   查看Web仪表盘状态：systemctl status wappalyzer-dashboard${NC}"
echo -e "${GREEN}   设置Web仪表盘开机自启：systemctl enable wappalyzer-dashboard${NC}"
echo -e "${GREEN}=== 部署脚本执行完成 ===${NC}"
