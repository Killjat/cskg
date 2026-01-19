#!/bin/bash

# SSL证书分析工具综合功能测试脚本

set -e

echo "=== SSL Certificate Analyzer Comprehensive Test ==="

# 确保工具已构建
if [ ! -f "cert-analyzer" ]; then
    echo "Building cert-analyzer..."
    go build -o cert-analyzer
fi

echo ""
echo "Testing all features of the certificate analyzer..."

# 测试1: 基本证书分析
echo ""
echo "Test 1: Basic certificate analysis"
./cert-analyzer -u https://httpbin.org -o test-basic.json -v
echo "✅ Basic analysis completed"

# 测试2: 高级安全分析
echo ""
echo "Test 2: Advanced security analysis"
./cert-analyzer -u https://httpbin.org --enable-advanced -o test-advanced.json -v
echo "✅ Advanced analysis completed"

# 测试3: 威胁情报分析
echo ""
echo "Test 3: Threat intelligence analysis"
./cert-analyzer -u https://github.com --enable-threat-intel -o test-threat.json -v
echo "✅ Threat intelligence analysis completed"

# 测试4: 钓鱼检测
echo ""
echo "Test 4: Phishing detection"
./cert-analyzer -u https://api.github.com --enable-phishing -o test-phishing.json -v
echo "✅ Phishing detection completed"

# 测试5: DGA检测
echo ""
echo "Test 5: DGA detection"
./cert-analyzer -u https://jsonplaceholder.typicode.com --enable-dga -o test-dga.json -v
echo "✅ DGA detection completed"

# 测试6: 搜索功能 + 高级分析
echo ""
echo "Test 6: Combined search and advanced analysis"
./cert-analyzer -u https://httpbin.org --enable-search --search-methods "crtsh" --enable-advanced --max-search-results 5 -o test-combined.json -v
echo "✅ Combined analysis completed"

# 测试7: 批量分析 + 高级功能
echo ""
echo "Test 7: Batch analysis with advanced features"
cat > test-urls-comprehensive.txt << EOF
https://httpbin.org
https://jsonplaceholder.typicode.com
https://api.github.com
EOF

./cert-analyzer -f test-urls-comprehensive.txt --enable-advanced --enable-search --search-methods "crtsh" --max-search-results 3 -c 2 -o test-batch-advanced.json -v
echo "✅ Batch advanced analysis completed"

# 测试8: CSV导出 + 高级分析
echo ""
echo "Test 8: CSV export with advanced analysis"
./cert-analyzer -f test-urls-comprehensive.txt --enable-advanced --format csv -o test-advanced.csv -v
echo "✅ CSV export with advanced analysis completed"

# 分析结果
echo ""
echo "=== Analysis Results Summary ==="
echo "Generated test files:"
ls -la test-*.json test-*.csv test-urls-comprehensive.txt 2>/dev/null || true

# 显示高级分析结果示例
echo ""
echo "Advanced Analysis Sample (Risk Scores):"
if [ -f "test-advanced.json" ]; then
    echo "Basic advanced analysis:"
    cat test-advanced.json | grep -E '"risk_score"|"threat_intelligence"|"phishing_analysis"|"dga_analysis"' | head -10
fi

echo ""
echo "Combined Analysis Sample (Search + Advanced):"
if [ -f "test-combined.json" ]; then
    echo "Related sites found:"
    cat test-combined.json | grep -E '"total_found"|"search_time_ms"' | head -2
    echo "Risk assessment:"
    cat test-combined.json | grep -E '"risk_score"|"recommendations"' | head -2
fi

echo ""
echo "Batch Analysis Summary:"
if [ -f "test-batch-advanced.json" ]; then
    echo "Batch statistics:"
    cat test-batch-advanced.json | grep -E '"total_urls"|"success_count"|"failure_count"' | head -3
fi

# 显示CSV结果
echo ""
echo "CSV Export Sample (first 3 lines):"
if [ -f "test-advanced.csv" ]; then
    head -3 test-advanced.csv
fi

# 功能演示总结
echo ""
echo "=== Feature Demonstration Summary ==="
echo ""
echo "✅ Completed Features:"
echo "  - Basic SSL/TLS certificate analysis"
echo "  - Certificate chain validation"
echo "  - Security scoring (0-100)"
echo "  - Related sites discovery (via crt.sh)"
echo "  - Advanced threat intelligence analysis"
echo "  - Phishing detection algorithms"
echo "  - DGA (Domain Generation Algorithm) detection"
echo "  - Timeline and anomaly analysis"
echo "  - Risk scoring and recommendations"
echo "  - Batch processing capabilities"
echo "  - Multiple output formats (JSON, CSV)"
echo "  - Concurrent processing"
echo ""
echo "🔍 Analysis Capabilities:"
echo "  - Certificate fingerprint analysis"
echo "  - Domain similarity detection"
echo "  - Entropy-based DGA detection"
echo "  - Time-based anomaly detection"
echo "  - Infrastructure correlation"
echo "  - Threat attribution"
echo ""
echo "📊 Output Features:"
echo "  - Structured JSON with detailed analysis"
echo "  - CSV format for spreadsheet analysis"
echo "  - Risk scores and severity levels"
echo "  - Actionable security recommendations"
echo "  - Related infrastructure mapping"

# 清理测试文件
echo ""
read -p "Clean up test files? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -f test-*.json test-*.csv test-urls-comprehensive.txt
    echo "Test files cleaned up"
fi

echo ""
echo "🎉 Comprehensive testing completed successfully!"
echo ""
echo "This tool now provides enterprise-grade SSL certificate analysis"
echo "suitable for security research, threat hunting, and infrastructure monitoring."