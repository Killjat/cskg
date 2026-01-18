#!/bin/bash

# 工业协议蜜罐系统启动脚本

# 设置颜色
green='\033[0;32m'
red='\033[0;31m'
yellow='\033[1;33m'
nc='\033[0m' # No Color

echo -e "${green}=== 工业协议蜜罐系统 ===${nc}"
echo -e "${green}执行时间：$(date)${nc}"
echo ""

# 检查是否已安装Go
if ! command -v go &> /dev/null; then
    echo -e "${red}错误：未安装Go语言环境${nc}"
    echo -e "${yellow}请先安装Go 1.21或更高版本${nc}"
    exit 1
fi

go_version=$(go version | awk '{print $3}' | sed 's/go//')
echo -e "${green}Go版本：${go_version}${nc}"

# 检查Go模块
echo -e "${green}检查Go模块...${nc}"
go mod tidy
if [ $? -ne 0 ]; then
    echo -e "${red}错误：Go模块初始化失败${nc}"
    exit 1
fi

# 创建必要的目录
echo -e "${green}创建必要的目录...${nc}"
mkdir -p logs data/fingerprints data/pcap

# 编译项目
echo -e "${green}编译项目...${nc}"
go build -o honeypot-server ./cmd/api
if [ $? -ne 0 ]; then
    echo -e "${red}错误：项目编译失败${nc}"
    exit 1
fi

# 启动服务
echo -e "${green}启动蜜罐系统...${nc}"
echo -e "${yellow}Web界面将在 http://0.0.0.0:8080 上运行${nc}"
echo -e "${yellow}按 Ctrl+C 停止服务${nc}"
echo ""

# 运行服务
./honeypot-server
