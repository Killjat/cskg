#!/bin/bash

# API Hunter 构建脚本

set -e

echo "🔨 开始构建 API Hunter..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go 1.19 或更高版本"
    exit 1
fi

# 检查Go版本
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.19"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "❌ Go 版本过低，需要 $REQUIRED_VERSION 或更高版本，当前版本: $GO_VERSION"
    exit 1
fi

echo "✅ Go 版本检查通过: $GO_VERSION"

# 创建必要的目录
echo "📁 创建目录结构..."
mkdir -p data
mkdir -p logs
mkdir -p exports
mkdir -p web/static
mkdir -p web/templates

# 下载依赖
echo "📦 下载依赖包..."
go mod tidy

# 构建应用
echo "🔨 构建应用..."
CGO_ENABLED=1 go build -ldflags="-s -w" -o api-hunter .

# 检查构建结果
if [ -f "api-hunter" ]; then
    echo "✅ 构建成功！"
    echo "📊 文件大小: $(du -h api-hunter | cut -f1)"
else
    echo "❌ 构建失败！"
    exit 1
fi

# 构建不同平台版本
echo "🌍 构建多平台版本..."

# Linux AMD64
echo "  📦 构建 Linux AMD64..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o api-hunter-linux-amd64 .

# Windows AMD64
echo "  📦 构建 Windows AMD64..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o api-hunter-windows-amd64.exe .

# macOS AMD64
echo "  📦 构建 macOS AMD64..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o api-hunter-darwin-amd64 .

# macOS ARM64 (Apple Silicon)
echo "  📦 构建 macOS ARM64..."
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o api-hunter-darwin-arm64 .

echo "✅ 多平台构建完成！"

# 显示构建结果
echo ""
echo "📋 构建结果:"
ls -lh api-hunter*

# 运行测试
echo ""
echo "🧪 运行测试..."
if go test ./... -v; then
    echo "✅ 所有测试通过！"
else
    echo "⚠️  部分测试失败，请检查代码"
fi

# 创建发布包
echo ""
echo "📦 创建发布包..."
VERSION=$(date +%Y%m%d_%H%M%S)
PACKAGE_NAME="api-hunter-${VERSION}"

mkdir -p "${PACKAGE_NAME}"
cp api-hunter "${PACKAGE_NAME}/"
cp config.yaml "${PACKAGE_NAME}/"
cp README.md "${PACKAGE_NAME}/"
cp -r web "${PACKAGE_NAME}/"

# 创建启动脚本
cat > "${PACKAGE_NAME}/start.sh" << 'EOF'
#!/bin/bash
echo "🚀 启动 API Hunter..."
./api-hunter web --port 8080
EOF

chmod +x "${PACKAGE_NAME}/start.sh"

# 创建tar包
tar -czf "${PACKAGE_NAME}.tar.gz" "${PACKAGE_NAME}"
rm -rf "${PACKAGE_NAME}"

echo "✅ 发布包创建完成: ${PACKAGE_NAME}.tar.gz"

# 显示使用说明
echo ""
echo "🎉 构建完成！"
echo ""
echo "📖 使用说明:"
echo "  1. 扫描网站: ./api-hunter scan -u https://example.com"
echo "  2. 启动Web界面: ./api-hunter web"
echo "  3. 查看帮助: ./api-hunter --help"
echo ""
echo "🌐 Web界面地址: http://localhost:8080"
echo "📁 数据存储位置: ./data/api_hunter.db"
echo "📝 日志文件位置: ./logs/api_hunter.log"
echo ""
echo "🚀 开始使用 API Hunter 发现网页中的API接口吧！"