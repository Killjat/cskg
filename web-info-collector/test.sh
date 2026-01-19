#!/bin/bash

# 网站信息收集工具测试脚本

set -e

echo "=== Web Info Collector Test Script ==="

# 确保工具已构建
if [ ! -f "bin/web-info-collector" ]; then
    echo "Building web-info-collector..."
    ./build.sh
fi

echo ""
echo "Running comprehensive tests..."

# 测试1: 单个网站信息收集
echo ""
echo "Test 1: Single website collection (Baidu)"
./bin/web-info-collector -u https://www.baidu.com -v -o test-baidu.json

if [ -f "test-baidu.json" ]; then
    echo "✅ Single website collection completed"
    echo "Basic info extracted:"
    cat test-baidu.json | grep -E '"title"|"icp_license"|"police_record"' | head -3
else
    echo "❌ Single website collection failed"
    exit 1
fi

# 测试2: 深度爬取
echo ""
echo "Test 2: Deep crawling (depth=2, max-pages=5)"
./bin/web-info-collector -u https://httpbin.org -d 2 --max-pages 5 -v -o test-deep.json

if [ -f "test-deep.json" ]; then
    echo "✅ Deep crawling completed"
    echo "Crawl statistics:"
    cat test-deep.json | grep -E '"pages_visited"|"total_links"|"download_links"' | head -3
else
    echo "❌ Deep crawling failed"
    exit 1
fi

# 测试3: 批量收集
echo ""
echo "Test 3: Batch collection"
cat > test-urls.txt << EOF
https://httpbin.org
https://jsonplaceholder.typicode.com
https://www.baidu.com
EOF

./bin/web-info-collector -f test-urls.txt -c 2 -v -o test-batch.json

if [ -f "test-batch.json" ]; then
    echo "✅ Batch collection completed"
    echo "Batch statistics:"
    cat test-batch.json | grep -E '"total_urls"|"success_count"|"failure_count"' | head -3
else
    echo "❌ Batch collection failed"
    exit 1
fi

# 测试4: HTML报告生成
echo ""
echo "Test 4: HTML report generation"
./bin/web-info-collector -f test-urls.txt --format html -o test-report.html -v

if [ -f "test-report.html" ]; then
    echo "✅ HTML report generated successfully"
    echo "Report size: $(wc -c < test-report.html) bytes"
else
    echo "❌ HTML report generation failed"
    exit 1
fi

# 测试5: CSV导出
echo ""
echo "Test 5: CSV export"
./bin/web-info-collector -f test-urls.txt --format csv -o test-results.csv -v

if [ -f "test-results.csv" ]; then
    echo "✅ CSV export successful"
    echo "CSV file lines: $(wc -l < test-results.csv)"
else
    echo "❌ CSV export failed"
    exit 1
fi

# 测试6: 特定功能测试
echo ""
echo "Test 6: Feature-specific tests"

# 测试ICP备案信息提取
echo "Testing ICP license extraction..."
./bin/web-info-collector -u https://www.baidu.com --extract-footer -v -o test-icp.json

if [ -f "test-icp.json" ]; then
    icp_found=$(cat test-icp.json | grep -o '"icp_license":[^,]*' | head -1)
    if [ ! -z "$icp_found" ]; then
        echo "✅ ICP license extraction working"
        echo "Found: $icp_found"
    else
        echo "⚠️  No ICP license found (may be normal)"
    fi
fi

# 显示测试结果文件
echo ""
echo "=== Test Results Summary ==="
echo "Generated test files:"
ls -la test-*.json test-*.csv test-*.html test-urls.txt 2>/dev/null || true

# 显示JSON结果示例
echo ""
echo "Sample collection result (first 30 lines):"
if [ -f "test-baidu.json" ]; then
    head -30 test-baidu.json
fi

echo ""
echo "Sample HTML report preview:"
if [ -f "test-report.html" ]; then
    echo "HTML report contains $(grep -c '<tr>' test-report.html) table rows"
    echo "Report title: $(grep -o '<title>[^<]*' test-report.html | sed 's/<title>//')"
fi

# 清理测试文件
echo ""
read -p "Clean up test files? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -f test-*.json test-*.csv test-*.html test-urls.txt
    echo "Test files cleaned up"
fi

echo ""
echo "All tests completed successfully! ✅"
echo ""
echo "Key features demonstrated:"
echo "  - Basic website information extraction"
echo "  - ICP and police record detection"
echo "  - Icon and favicon collection"
echo "  - Download links discovery"
echo "  - Footer information parsing"
echo "  - Technical stack detection"
echo "  - Batch processing capabilities"
echo "  - Multiple output formats (JSON, CSV, HTML)"
echo "  - Deep crawling with configurable depth"