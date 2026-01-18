#!/bin/bash

# ImageGPS API测试脚本

echo "=== ImageGPS API测试 ==="
echo ""

# 检查服务是否运行
echo "1. 检查服务健康状态..."
curl -s http://localhost:8080/api/health | jq .
echo ""

# 测试图片上传（需要提供一张包含GPS信息的图片）
echo "2. 上传图片提取GPS信息..."
echo "   使用方法: curl -X POST -F 'image=@/path/to/your/image.jpg' http://localhost:8080/api/upload"
echo ""
echo "   如果你有包含GPS信息的图片，可以运行："
echo "   curl -X POST -F 'image=@你的图片路径.jpg' http://localhost:8080/api/upload | jq ."
echo ""

echo "提示：可以使用iPhone、Android手机拍摄的原始照片进行测试"
echo "     （需确保拍摄时GPS定位已开启）"
