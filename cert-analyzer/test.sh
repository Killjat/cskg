#!/bin/bash

# SSL证书分析工具测试脚本

set -e

echo "=== SSL Certificate Analyzer Test Script ==="

# 确保工具已构建
if [ ! -f "bin/cert-analyzer" ]; then
    echo "Building cert-analyzer..."
    ./build.sh
fi

echo ""
echo "Running comprehensive tests..."

# 测试1: 单个URL分析
echo ""
echo "Test 1: Single URL analysis (Google)"
./bin/cert-analyzer -u https://www.google.com -v

# 测试2: 输出到文件
echo ""
echo "Test 2: Output to file"
./bin/cert-analyzer -u https://www.github.com -o test-github.json -v
if [ -f "test-github.json" ]; then
    echo "✅ Output file created successfully"
    echo "File size: $(wc -c < test-github.json) bytes"
else
    echo "❌ Output file not created"
    exit 1
fi

# 测试3: 批量分析
echo ""
echo "Test 3: Batch analysis"
./bin/cert-analyzer -f examples/urls.txt -o test-batch.json -c 3 -v
if [ -f "test-batch.json" ]; then
    echo "✅ Batch analysis completed"
    echo "Results file size: $(wc -c < test-batch.json) bytes"
else
    echo "❌ Batch analysis failed"
    exit 1
fi

# 测试4: CSV导出
echo ""
echo "Test 4: CSV export"
./bin/cert-analyzer -f examples/urls.txt -o test-results.csv --format csv -c 2 -v
if [ -f "test-results.csv" ]; then
    echo "✅ CSV export successful"
    echo "CSV file lines: $(wc -l < test-results.csv)"
else
    echo "❌ CSV export failed"
    exit 1
fi

# 测试5: 错误处理
echo ""
echo "Test 5: Error handling (invalid URL)"
if ./bin/cert-analyzer -u https://invalid-domain-that-does-not-exist.com 2>/dev/null; then
    echo "❌ Should have failed for invalid domain"
    exit 1
else
    echo "✅ Correctly handled invalid domain"
fi

# 测试6: 自签名证书
echo ""
echo "Test 6: Self-signed certificate detection"
./bin/cert-analyzer -u https://self-signed.badssl.com --skip-verify -v

# 测试7: 过期证书
echo ""
echo "Test 7: Expired certificate detection"
./bin/cert-analyzer -u https://expired.badssl.com --skip-verify -v

# 显示测试结果文件
echo ""
echo "=== Test Results Summary ==="
echo "Generated test files:"
ls -la test-*.json test-*.csv 2>/dev/null || true

# 显示JSON结果示例
echo ""
echo "Sample JSON output (first 20 lines):"
head -20 test-github.json

echo ""
echo "Sample CSV output (first 5 lines):"
head -5 test-results.csv

# 清理测试文件
echo ""
read -p "Clean up test files? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -f test-*.json test-*.csv
    echo "Test files cleaned up"
fi

echo ""
echo "All tests completed successfully! ✅"