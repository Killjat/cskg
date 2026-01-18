#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
测试脚本：测试uploads目录中的所有图片文件
"""

import os
import requests
import json
import time

def test_uploads_directory():
    """测试uploads目录中的所有图片文件"""
    print("=== 测试uploads目录中的图片GPS位置信息提取 ===")
    
    # API地址
    api_url = "http://localhost:8080/api/upload"
    
    # 上传文件目录
    upload_dir = "./uploads"
    
    # 检查目录是否存在
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
    failed_files = []
    
    # 详细结果
    detailed_results = []
    
    # 按文件大小排序，先测试大文件
    image_files.sort(key=lambda x: -os.path.getsize(os.path.join(upload_dir, x)))
    
    # 逐个测试文件
    for i, filename in enumerate(image_files, 1):
        print(f"============================================================")
        print(f"测试图片 {i}/{total_files}: {filename}")
        
        # 获取文件大小
        file_path = os.path.join(upload_dir, filename)
        file_size = os.path.getsize(file_path)
        print(f"  文件大小: {file_size / 1024:.2f} KB")
        
        try:
            # 读取文件
            with open(file_path, "rb") as f:
                files = {"image": (filename, f, "image/jpeg")}
                
                # 记录开始时间
                start_time = time.time()
                
                # 发送请求
                response = requests.post(api_url, files=files, timeout=30)
                
                # 记录结束时间
                end_time = time.time()
                
            # 解析响应
            result = response.json()
            
            # 打印响应时间
            print(f"  响应时间: {end_time - start_time:.2f} 秒")
            
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
                    "location": f"{latitude:.6f}, {longitude:.6f}",
                    "time": end_time - start_time
                })
            else:
                message = result.get("message", "未知错误")
                print(f"  ✗ 处理失败: {message}")
                
                detailed_results.append({
                    "filename": filename,
                    "success": False,
                    "has_gps": False,
                    "location": "",
                    "time": end_time - start_time
                })
                
                failed_files.append(filename)
        except Exception as e:
            print(f"  ✗ 发生异常: {e}")
            
            detailed_results.append({
                "filename": filename,
                "success": False,
                "has_gps": False,
                "location": "",
                "time": 0
            })
            
            failed_files.append(filename)
    
    # 打印总结
    print()
    print("============================================================")
    print("测试结果总结:")
    print(f"总测试图片数: {total_files}")
    print(f"成功处理图片数: {success_count}")
    print(f"包含GPS信息的图片数: {gps_count}")
    print(f"失败图片数: {len(failed_files)}")
    print()
    
    if failed_files:
        print("失败的图片:")
        for filename in failed_files:
            print(f"  - {filename}")
    
    print()
    print("============================================================")
    print("详细测试结果:")
    print(f"{'序号':<4} {'文件名':<40} {'是否成功':<8} {'是否有GPS':<8} {'响应时间':<10} {'位置':<30}")
    print("-" * 120)
    
    for i, result in enumerate(detailed_results, 1):
        success = "✓" if result["success"] else "✗"
        has_gps = "✓" if result["has_gps"] else "✗"
        time_str = f"{result['time']:.2f}s"
        print(f"{i:<4} {result['filename']:<40} {success:<8} {has_gps:<8} {time_str:<10} {result['location']:<30}")
    
    print()
    print("============================================================")
    print("测试完成！")

if __name__ == "__main__":
    test_uploads_directory()
