#!/bin/bash

# 工业协议蜜罐系统CentOS部署脚本

# 配置颜色
green='\033[0;32m'
red='\033[0;31m'
yellow='\033[1;33m'
nc='\033[0m' # No Color

echo -e "${green}=== 工业协议蜜罐系统 - CentOS部署脚本 ===${nc}"
echo -e "${green}执行时间：$(date)${nc}"
echo ""

# 检查操作系统
if [ ! -f /etc/centos-release ] && [ ! -f /etc/redhat-release ]; then
    echo -e "${red}错误：该脚本仅支持CentOS/RHEL系统${nc}"
    exit 1
fi

echo -e "${green}1. 检查Go语言环境...${nc}"

# 检查Go是否已安装
if ! command -v go &> /dev/null; then
    echo -e "${yellow}未安装Go语言，正在安装...${nc}"
    
    # 安装依赖
    sudo yum install -y gcc gcc-c++ make wget libpcap-devel
    
    # 下载并安装Go 1.21
    wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
    rm go1.21.5.linux-amd64.tar.gz
    
    # 设置环境变量
    echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc
    echo "export GOPATH=$HOME/go" >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH=$HOME/go
    echo -e "${green}Go 1.21.5 安装完成${nc}"
fi

go_version=$(go version | awk '{print $3}' | sed 's/go//')
echo -e "${green}Go版本：${go_version}${nc}"

# 创建项目目录
echo -e "${green}2. 创建项目目录...${nc}"
PROJECT_DIR="/opt/honeypot_system"
mkdir -p $PROJECT_DIR

# 复制项目文件
echo -e "${green}3. 复制项目文件...${nc}"
# 注意：在实际部署中，这里应该是从代码仓库克隆或上传项目文件
# 由于我们是在开发环境中，直接复制本地文件
cp -r honeypot_system_go/* $PROJECT_DIR/

# 进入项目目录
cd $PROJECT_DIR

echo -e "${green}4. 设置Go模块代理...${nc}"
export GOPROXY=https://goproxy.cn,direct

echo -e "${green}5. 下载依赖...${nc}"
go mod tidy
if [ $? -ne 0 ]; then
    echo -e "${red}错误：Go模块依赖下载失败${nc}"
    exit 1
fi

echo -e "${green}6. 编译项目...${nc}"
go build -o honeypot-server ./cmd/api
if [ $? -ne 0 ]; then
    echo -e "${red}错误：项目编译失败${nc}"
    exit 1
fi

echo -e "${green}7. 创建必要的目录...${nc}"
mkdir -p logs data/fingerprints data/pcap

echo -e "${green}8. 设置服务启动脚本...${nc}"
# 创建systemd服务文件
SERVICE_FILE="/etc/systemd/system/honeypot.service"
sudo cat > $SERVICE_FILE << EOF
[Unit]
Description=Industrial Protocol Honeypot System
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$PROJECT_DIR
ExecStart=$PROJECT_DIR/honeypot-server
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF

# 重新加载systemd配置
sudo systemctl daemon-reload

# 启用服务
sudo systemctl enable honeypot.service

echo -e "${green}9. 启动蜜罐系统...${nc}"
# 启动服务
sudo systemctl start honeypot.service

# 检查服务状态
sleep 3
echo -e "${green}10. 检查服务状态...${nc}"
sudo systemctl status honeypot.service --no-pager

echo -e "${green}11. 获取访问地址...${nc}"
# 获取服务器IP
SERVER_IP=$(curl -s icanhazip.com 2>/dev/null || echo "无法获取公网IP")
INTERNAL_IP=$(hostname -I | awk '{print $1}')

echo ""
echo -e "${green}=== 部署完成 ===${nc}"
echo -e "${green}蜜罐系统已成功部署并启动${nc}"
echo ""
echo -e "${yellow}访问地址：${nc}"
echo -e "- 公网访问：http://$SERVER_IP:8080"
echo -e "- 内网访问：http://$INTERNAL_IP:8080"
echo -e "- API接口：http://$SERVER_IP:8080/api/fingerprints"
echo ""
echo -e "${yellow}服务管理命令：${nc}"
echo -e "- 启动服务：sudo systemctl start honeypot.service"
echo -e "- 停止服务：sudo systemctl stop honeypot.service"
echo -e "- 重启服务：sudo systemctl restart honeypot.service"
echo -e "- 查看状态：sudo systemctl status honeypot.service"
echo -e "- 查看日志：sudo journalctl -u honeypot.service -f"
echo ""
echo -e "${green}蜜罐系统部署完成！${nc}"
