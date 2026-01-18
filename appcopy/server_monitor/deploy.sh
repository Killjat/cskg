#!/bin/bash

# 服务器监控系统部署脚本
# 支持Ubuntu/Debian和CentOS/RHEL系统

set -e

# 配置变量
APP_NAME="server-monitor"
APP_DIR="/opt/$APP_NAME"
BIN_NAME="$APP_NAME"
CONFIG_FILE="config.yaml"
SYSTEMD_SERVICE="$APP_NAME.service"
GIT_REPO=""
GO_VERSION="1.21.0"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}===== 服务器监控系统部署脚本 =====${NC}"

# 检查root权限
if [ "$(id -u)" != "0" ]; then
   echo -e "${RED}错误：脚本需要以root权限运行${NC}"
   exit 1
fi

# 检测Linux发行版
detect_distro() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        DISTRO=$ID
        VERSION=$VERSION_ID
    elif type lsb_release >/dev/null 2>&1; then
        DISTRO=$(lsb_release -si | tr '[:upper:]' '[:lower:]')
        VERSION=$(lsb_release -sr)
    elif [ -f /etc/lsb-release ]; then
        . /etc/lsb-release
        DISTRO=$DISTRIB_ID
        VERSION=$DISTRIB_RELEASE
    elif [ -f /etc/debian_version ]; then
        DISTRO="debian"
        VERSION=$(cat /etc/debian_version)
    elif [ -f /etc/redhat-release ]; then
        DISTRO="rhel"
        VERSION=$(cat /etc/redhat-release | grep -oE '[0-9]+\.[0-9]+')
    else
        DISTRO="unknown"
        VERSION="unknown"
    fi
    echo -e "${YELLOW}检测到系统：$DISTRO $VERSION${NC}"
}

# 安装基础依赖
install_dependencies() {
    echo -e "${GREEN}安装基础依赖...${NC}"
    
    case $DISTRO in
        ubuntu|debian)
            apt-get update
            apt-get install -y curl wget git build-essential procps
            ;;
        centos|rhel|rocky|almalinux)
            yum install -y curl wget git gcc gcc-c++ make procps
            ;;
        *)
            echo -e "${RED}不支持的Linux发行版：$DISTRO${NC}"
            exit 1
            ;;
    esac
}

# 安装Go
install_go() {
    echo -e "${GREEN}安装Go $GO_VERSION...${NC}"
    
    # 检查是否已安装Go
    if command -v go >/dev/null 2>&1; then
        CURRENT_GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+\.[0-9]+' | cut -c 3-)
        if [ "$CURRENT_GO_VERSION" == "$GO_VERSION" ]; then
            echo -e "${YELLOW}Go $GO_VERSION 已安装，跳过...${NC}"
            return
        fi
    fi
    
    # 下载并安装Go
    GO_ARCH=$(uname -m)
    case $GO_ARCH in
        x86_64)
            GO_ARCH="amd64"
            ;;
        aarch64)
            GO_ARCH="arm64"
            ;;
        *)
            echo -e "${RED}不支持的架构：$GO_ARCH${NC}"
            exit 1
            ;;
    esac
    
    GO_URL="https://golang.org/dl/go$GO_VERSION.linux-$GO_ARCH.tar.gz"
    wget -q -O /tmp/go.tar.gz $GO_URL
    tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
    
    # 设置环境变量
    cat << EOF >> /etc/profile
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/root/go
export GOROOT=/usr/local/go
export GO111MODULE=on
export GOPROXY=https://goproxy.io,direct
EOF
    
    # 立即生效
    source /etc/profile
    
    echo -e "${GREEN}Go $GO_VERSION 安装完成${NC}"
}

# 构建应用
build_app() {
    echo -e "${GREEN}构建应用...${NC}"
    
    # 进入应用目录
    cd "$APP_DIR"
    
    # 安装Go依赖
    go mod tidy
    
    # 构建二进制文件
    go build -o "$BIN_NAME" main.go
    
    # 验证构建结果
    if [ -f "$BIN_NAME" ]; then
        echo -e "${GREEN}应用构建成功：$(pwd)/$BIN_NAME${NC}"
    else
        echo -e "${RED}应用构建失败${NC}"
        exit 1
    fi
}

# 创建配置文件
create_config() {
    echo -e "${GREEN}创建配置文件...${NC}"
    
    if [ ! -f "$APP_DIR/$CONFIG_FILE" ]; then
        cat << EOF > "$APP_DIR/$CONFIG_FILE"
server:
  port: 8081
  host: 0.0.0.0
  read_timeout: 30
  write_timeout: 30
collector:
  login_interval: 5
  process_interval: 2
  command_interval: 2
  file_watch_paths:
  - /var/log
  - /etc
  recursive_watch: false
storage:
  type: sqlite
  file_path: ./server_monitor.db
  host: localhost
  port: 3306
  username: root
  password: ""
  database: server_monitor
alert:
  enabled: false
  levels:
  - warning
  - error
  - critical
  notification:
  - email
  smtp:
    host: localhost
    port: 25
    username: ""
    password: ""
    from: ""
    to: ""
EOF
        echo -e "${GREEN}配置文件创建成功${NC}"
    else
        echo -e "${YELLOW}配置文件已存在，跳过...${NC}"
    fi
}

# 创建systemd服务
create_systemd_service() {
    echo -e "${GREEN}创建systemd服务...${NC}"
    
    cat << EOF > "/etc/systemd/system/$SYSTEMD_SERVICE"
[Unit]
Description=Server Monitor System
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/$BIN_NAME
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
EOF
    
    # 重新加载systemd
    systemctl daemon-reload
    
    echo -e "${GREEN}systemd服务创建成功${NC}"
}

# 启动服务
start_service() {
    echo -e "${GREEN}启动服务...${NC}"
    
    # 检查服务是否已启动
    if systemctl is-active --quiet "$SYSTEMD_SERVICE"; then
        echo -e "${YELLOW}服务已启动，重新启动...${NC}"
        systemctl restart "$SYSTEMD_SERVICE"
    else
        systemctl start "$SYSTEMD_SERVICE"
        systemctl enable "$SYSTEMD_SERVICE"
    fi
    
    # 检查服务状态
    echo -e "${GREEN}检查服务状态...${NC}"
    systemctl status "$SYSTEMD_SERVICE" --no-pager
    
    echo -e "${GREEN}服务启动成功！${NC}"
    echo -e "${YELLOW}访问地址：http://$(curl -s ifconfig.me):8081/${NC}"
}

# 主部署流程
deploy() {
    # 检测系统
    detect_distro
    
    # 安装依赖
    install_dependencies
    
    # 安装Go
    install_go
    
    # 创建应用目录
    mkdir -p "$APP_DIR"
    
    # 复制应用文件到安装目录
    echo -e "${GREEN}复制应用文件到安装目录...${NC}"
    cp -r ./* "$APP_DIR/" 2>/dev/null || true
    
    # 构建应用
    build_app
    
    # 创建配置文件
    create_config
    
    # 创建systemd服务
    create_systemd_service
    
    # 启动服务
    start_service
    
    echo -e "${GREEN}===== 部署完成 =====${NC}"
    echo -e "${YELLOW}服务管理命令：${NC}"
    echo -e "  启动服务：systemctl start $SYSTEMD_SERVICE"
    echo -e "  停止服务：systemctl stop $SYSTEMD_SERVICE"
    echo -e "  重启服务：systemctl restart $SYSTEMD_SERVICE"
    echo -e "  查看状态：systemctl status $SYSTEMD_SERVICE"
    echo -e "  查看日志：journalctl -u $SYSTEMD_SERVICE -f"
    echo -e "  配置文件：$APP_DIR/$CONFIG_FILE"
    echo -e "  访问地址：http://$(curl -s ifconfig.me):8081/"
}

# 卸载服务
uninstall() {
    echo -e "${GREEN}卸载服务器监控系统...${NC}"
    
    # 停止服务
    if systemctl is-active --quiet "$SYSTEMD_SERVICE"; then
        systemctl stop "$SYSTEMD_SERVICE"
        systemctl disable "$SYSTEMD_SERVICE"
    fi
    
    # 删除systemd服务
    if [ -f "/etc/systemd/system/$SYSTEMD_SERVICE" ]; then
        rm -f "/etc/systemd/system/$SYSTEMD_SERVICE"
        systemctl daemon-reload
    fi
    
    # 删除应用目录
    if [ -d "$APP_DIR" ]; then
        rm -rf "$APP_DIR"
    fi
    
    echo -e "${GREEN}卸载完成${NC}"
}

# 显示帮助
show_help() {
    echo -e "${GREEN}服务器监控系统部署脚本${NC}"
    echo "使用方法：$0 [选项]"
    echo "选项："
    echo "  deploy    - 部署服务器监控系统"
    echo "  uninstall - 卸载服务器监控系统"
    echo "  help      - 显示帮助信息"
    echo ""
    echo "示例："
    echo "  $0 deploy    # 部署系统"
    echo "  $0 uninstall # 卸载系统"
}

# 主函数
main() {
    case "$1" in
        deploy)
            deploy
            ;;
        uninstall)
            uninstall
            ;;
        help|-h|--help)
            show_help
            ;;
        *)
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"