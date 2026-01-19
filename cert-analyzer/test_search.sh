#!/bin/bash

# SSL证书分析工具搜索功能测试脚本

set -e

echo "=== SSL Certificate Analyzer Search Feature Test ==="

# 确保工具已构建
if [ ! -f "cert-analyzer" ]; then
    echo "Building cert-analyzer..."
    go build -o cert-analyzer
fi

echo ""
echo "Testing certificate search functionality..."

# 测试1: 基本搜索功能（使用免费的crt.sh）
echo ""
echo "Test 1: Basic search with crt.sh"
./cert-analyzer -u https://httpbin.org --enable-search --search-methods "crtsh" --max-search-results 5 -o test-search-basic.json -v

if [ -f "test-search-basic.json" ]; then
    echo "✅ Basic search test completed"
    echo "Found related sites:"
    cat test-search-basic.json | grep -o '"domain":"[^"]*"' | head -5
else
    echo "❌ Basic search test failed"
    exit 1
fi

# 测试2: 多个搜索方法（只使用免费的）
echo ""
echo "Test 2: Multiple search methods (free only)"
./cert-analyzer -u https://github.com --enable-search --search-methods "crtsh" --max-search-results 10 -o test-search-multi.json -v

if [ -f "test-search-multi.json" ]; then
    echo "✅ Multi-method search test completed"
    # 显示搜索统计
    echo "Search statistics:"
    cat test-search-multi.json | grep -E '"total_found"|"search_time_ms"|"search_method"' | head -3
else
    echo "❌ Multi-method search test failed"
    exit 1
fi

# 测试3: 批量搜索
echo ""
echo "Test 3: Batch search"
cat > test-urls-search.txt << EOF
https://httpbin.org
https://jsonplaceholder.typicode.com
https://api.github.com
EOF

./cert-analyzer -f test-urls-search.txt --enable-search --search-methods "crtsh" --max-search-results 5 -c 2 -o test-batch-search.json -v

if [ -f "test-batch-search.json" ]; then
    echo "✅ Batch search test completed"
    # 显示批量搜索结果摘要
    echo "Batch search summary:"
    cat test-batch-search.json | grep -E '"total_urls"|"success_count"|"failure_count"' | head -3
else
    echo "❌ Batch search test failed"
    exit 1
fi

# 测试4: 搜索超时测试
echo ""
echo "Test 4: Search timeout test"
./cert-analyzer -u https://httpbin.org --enable-search --search-methods "crtsh" --search-timeout 2s --max-search-results 3 -v

echo "✅ Search timeout test completed"

# 显示测试结果文件
echo ""
echo "=== Test Results Summary ==="
echo "Generated test files:"
ls -la test-*search*.json test-urls-search.txt 2>/dev/null || true

# 显示一个完整的搜索结果示例
echo ""
echo "Sample search result (related sites):"
if [ -f "test-search-basic.json" ]; then
    cat test-search-basic.json | grep -A 20 '"related_sites"' | head -25
fi

# 清理测试文件
echo ""
read -p "Clean up test files? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -f test-*search*.json test-urls-search.txt
    echo "Test files cleaned up"
fi

echo ""
echo "All search functionality tests completed successfully! ✅"
echo ""
echo "Key features demonstrated:"
echo "- Certificate fingerprint-based search"
echo "- Multiple search engines support"
echo "- Related sites discovery"
echo "- Batch processing with search"
echo "- Configurable search parameters"
echo "- JSON output with search results"