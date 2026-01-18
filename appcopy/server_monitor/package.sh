#!/bin/bash

# 服务器监控系统打包脚本
# 生成适用于Linux系统的安装包

set -e

# 配置变量
APP_NAME="server-monitor"
VERSION="1.0.0"
BUILD_DIR="build"
PACKAGE_DIR="$BUILD_DIR/$APP_NAME-$VERSION"
ARCH="$(uname -m)"
DATE="$(date +%Y%m%d)"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}===== 服务器监控系统打包脚本 =====${NC}"

# 检查Go环境
if ! command -v go >/dev/null 2>&1; then
    echo -e "${RED}错误：未找到Go环境，请先安装Go 1.21.0+${NC}"
    exit 1
fi

# 创建构建目录
echo -e "${GREEN}创建构建目录...${NC}"
rm -rf "$BUILD_DIR"
mkdir -p "$PACKAGE_DIR"
mkdir -p "$PACKAGE_DIR/static"
mkdir -p "$PACKAGE_DIR/internal/collector"
mkdir -p "$PACKAGE_DIR/internal/config"
mkdir -p "$PACKAGE_DIR/internal/model"

# 复制源代码文件
echo -e "${GREEN}复制源代码文件...${NC}"
cp main.go "$PACKAGE_DIR/"
cp go.mod go.sum "$PACKAGE_DIR/"
cp config.yaml "$PACKAGE_DIR/"
cp deploy.sh package.sh README.md "$PACKAGE_DIR/"
cp -r static/* "$PACKAGE_DIR/static/"
cp -r internal/* "$PACKAGE_DIR/internal/"

# 创建配置文件示例
echo -e "${GREEN}创建配置文件示例...${NC}"
cp config.yaml "$PACKAGE_DIR/config.yaml.example"

# 设置执行权限
echo -e "${GREEN}设置执行权限...${NC}"
chmod +x "$PACKAGE_DIR/deploy.sh"
chmod +x "$PACKAGE_DIR/package.sh"

# 编译不同架构的二进制文件
echo -e "${GREEN}编译二进制文件...${NC}"

# 编译当前架构
echo -e "${YELLOW}编译当前架构 ($ARCH)...${NC}"
cd "$PACKAGE_DIR"
go build -o "$APP_NAME" main.go
cd - > /dev/null

# 编译x86_64架构
echo -e "${YELLOW}编译x86_64架构...${NC}"
cd "$PACKAGE_DIR"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$APP_NAME-amd64" main.go
cd - > /dev/null

# 编译ARM64架构
echo -e "${YELLOW}编译ARM64架构...${NC}"
cd "$PACKAGE_DIR"
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o "$APP_NAME-arm64" main.go
cd - > /dev/null

# 创建压缩包
echo -e "${GREEN}创建压缩包...${NC}"

# 创建tar.gz压缩包
tar -czf "$BUILD_DIR/$APP_NAME-$VERSION-$ARCH-$DATE.tar.gz" -C "$BUILD_DIR" "$APP_NAME-$VERSION"

# 创建zip压缩包（可选）
if command -v zip >/dev/null 2>&1; then
    zip -r "$BUILD_DIR/$APP_NAME-$VERSION-$ARCH-$DATE.zip" "$PACKAGE_DIR" > /dev/null
fi

# 清理临时文件
echo -e "${GREEN}清理临时文件...${NC}"
rm -rf "$PACKAGE_DIR"

# 显示打包结果
echo -e "${GREEN}===== 打包完成 =====${NC}"
echo -e "${YELLOW}生成的安装包：${NC}"
ls -la "$BUILD_DIR/"

echo -e "${GREEN}安装包使用说明：${NC}"
echo "1. 解压安装包：tar -xzf $APP_NAME-$VERSION-$ARCH-$DATE.tar.gz"
echo "2. 进入安装目录：cd $APP_NAME-$VERSION"
echo "3. 执行部署脚本：sudo ./deploy.sh deploy"
echo "4. 访问Web界面：http://<服务器IP>:8081/"
