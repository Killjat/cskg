#!/bin/bash

# SSL证书分析工具构建脚本

set -e

echo "=== SSL Certificate Analyzer Build Script ==="

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

echo "Go version: $(go version)"

# 创建输出目录
mkdir -p bin

# 获取依赖
echo "Downloading dependencies..."
go mod tidy
go mod download

# 构建不同平台的二进制文件
echo "Building binaries..."

# Linux AMD64
echo "Building for Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/cert-analyzer-linux-amd64

# macOS AMD64
echo "Building for macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/cert-analyzer-darwin-amd64

# macOS ARM64 (Apple Silicon)
echo "Building for macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/cert-analyzer-darwin-arm64

# Windows AMD64
echo "Building for Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/cert-analyzer-windows-amd64.exe

# 本地平台
echo "Building for current platform..."
go build -ldflags="-s -w" -o bin/cert-analyzer

# 设置执行权限
chmod +x bin/cert-analyzer*

echo "Build completed successfully!"
echo "Binaries available in bin/ directory:"
ls -la bin/

# 运行测试
echo ""
echo "Running basic test..."
if ./bin/cert-analyzer --help > /dev/null 2>&1; then
    echo "✅ Basic functionality test passed"
else
    echo "❌ Basic functionality test failed"
    exit 1
fi

echo ""
echo "Build and test completed successfully!"
echo ""
echo "Usage examples:"
echo "  ./bin/cert-analyzer -u https://www.google.com"
echo "  ./bin/cert-analyzer -f examples/urls.txt -o results.json"
echo "  ./bin/cert-analyzer -u https://www.github.com -o github-cert.json -v"