#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
测试脚本：使用保存的上传文件测试GPS提取功能
"""

import os
import requests
import json

def test_saved_uploads():
    """测试保存的上传文件"""
    print("=== 测试保存的上传文件GPS位置信息提取 ===")
    
    # API地址
    api_url = "http://localhost:8080/api/upload"
    
    # 上传文件目录
    upload_dir = "./uploads"
    
    # 获取所有上传的文件
    if not os.path.exists(upload_dir):
        print(f"上传目录 {upload_dir} 不存在")
        return
    
    # 获取所有图片文件
    image_files = []
    for filename in os.listdir(upload_dir):
        if filename.lower().endswith((".jpg", ".jpeg", ".png", ".tiff", ".tif")):
            image_files.append(filename)
    
    if not image_files:
        print(f"上传目录 {upload_dir} 中没有图片文件")
        return
    
    print(f"测试图片数量: {len(image_files)}")
    print(f"API地址: {api_url}")
    print()
    
    # 测试结果
    total_files = len(image_files)
    success_count = 0
    gps_count = 0
    
    # 详细结果
    detailed_results = []
    
    # 逐个测试文件
    for i, filename in enumerate(image_files, 1):
        print(f"============================================================")
        print(f"测试图片 {i}/{total_files}: {filename}")
        
        # 构建文件路径
        file_path = os.path.join(upload_dir, filename)
        
        try:
            # 读取文件
            with open(file_path, "rb") as f:
                files = {"image": (filename, f, "image/jpeg")}
                
                # 发送请求
                response = requests.post(api_url, files=files, timeout=30)
                
            # 解析响应
            result = response.json()
            
            if result.get("success"):
                success_count += 1
                gps_count += 1
                
                gps_data = result.get("data", {})
                latitude = gps_data.get("latitude")
                longitude = gps_data.get("longitude")
                datetime = gps_data.get("datetime", "")
                make = gps_data.get("make", "")
                model = gps_data.get("model", "")
                
                print(f"  ✓ 成功提取GPS: {latitude:.6f}, {longitude:.6f}")
                if datetime:
                    print(f"  📅 拍摄时间: {datetime}")
                if make or model:
                    print(f"  📷 相机: {make} {model}")
                
                detailed_results.append({
                    "filename": filename,
                    "success": True,
                    "has_gps": True,
                    "location": f"{latitude:.6f}, {longitude:.6f}"
                })
            else:
                message = result.get("message", "未知错误")
                print(f"  ✗ 处理失败: {message}")
                
                detailed_results.append({
                    "filename": filename,
                    "success": False,
                    "has_gps": False,
                    "location": ""
                })
        except Exception as e:
            print(f"  ✗ 发生异常: {e}")
            
            detailed_results.append({
                "filename": filename,
                "success": False,
                "has_gps": False,
                "location": ""
            })
    
    # 打印总结
    print()
    print("============================================================")
    print("测试结果总结:")
    print(f"总测试图片数: {total_files}")
    print(f"成功处理图片数: {success_count}")
    print(f"包含GPS信息的图片数: {gps_count}")
    print()
    print("============================================================")
    print("详细测试结果:")
    print(f"{'序号':<4} {'文件名':<40} {'是否成功':<8} {'是否有GPS':<8} {'位置':<30}")
    print("-" * 100)
    
    for i, result in enumerate(detailed_results, 1):
        success = "✓" if result["success"] else "✗"
        has_gps = "✓" if result["has_gps"] else "✗"
        print(f"{i:<4} {result['filename']:<40} {success:<8} {has_gps:<8} {result['location']:<30}")
    
    print()
    print("============================================================")
    print("测试完成！")

if __name__ == "__main__":
    test_saved_uploads()
