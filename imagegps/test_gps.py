#!/usr/bin/env python3

import os
import requests
import glob

# API endpoint for image GPS extraction
API_URL = 'http://localhost:8080/api/upload'

# Directory containing test images (created with GPS info)
TEST_IMAGES_DIR = './test_images'

# Get all image files
image_files = glob.glob(os.path.join(TEST_IMAGES_DIR, '*.jpg'))

# Limit to 10 images for testing
image_files = image_files[:10]

def test_image_gps(image_path):
    """Test GPS extraction for a single image"""
    try:
        with open(image_path, 'rb') as f:
            files = {'image': (os.path.basename(image_path), f, 'image/jpeg')}
            response = requests.post(API_URL, files=files, timeout=10)
            
        if response.status_code == 200:
            result = response.json()
            data = result.get('data', {})
            return {
                'file': os.path.basename(image_path),
                'success': result.get('success', False),
                'has_gps': data.get('has_gps', False),
                'latitude': data.get('latitude', 0),
                'longitude': data.get('longitude', 0),
                'location': f"{data.get('latitude', 0):.6f}, {data.get('longitude', 0):.6f}",
                'datetime': data.get('datetime', ''),
                'make': data.get('make', ''),
                'model': data.get('model', ''),
                'google_map_url': data.get('google_map_url', '')
            }
        else:
            return {
                'file': os.path.basename(image_path),
                'success': False,
                'error': f"HTTP {response.status_code}: {response.text}"
            }
    except Exception as e:
        return {
            'file': os.path.basename(image_path),
            'success': False,
            'error': str(e)
        }

def main():
    """Main test function"""
    print("=== 测试图片GPS位置信息提取 ===")
    print(f"测试图片数量: {len(image_files)}")
    print(f"API地址: {API_URL}")
    print("\n" + "="*60)
    
    results = []
    success_count = 0
    gps_count = 0
    
    for i, image_path in enumerate(image_files, 1):
        print(f"测试图片 {i}/{len(image_files)}: {os.path.basename(image_path)}")
        result = test_image_gps(image_path)
        results.append(result)
        
        if result['success']:
            success_count += 1
            if result['has_gps']:
                gps_count += 1
                print(f"  ✓ 成功提取GPS: {result['location']}")
                print(f"  📅 拍摄时间: {result['datetime']}")
                print(f"  📷 相机: {result['make']} {result['model']}")
            else:
                print(f"  ✓ 成功处理，但图片中没有GPS信息")
        else:
            print(f"  ✗ 处理失败: {result['error']}")
        print()
    
    # Summary
    print("="*60)
    print("测试结果总结:")
    print(f"总测试图片数: {len(image_files)}")
    print(f"成功处理图片数: {success_count}")
    print(f"包含GPS信息的图片数: {gps_count}")
    
    # Detailed results table
    print("\n" + "="*60)
    print("详细测试结果:")
    print(f"{'序号':<5} {'文件名':<40} {'是否成功':<8} {'是否有GPS':<8} {'位置':<25}")
    print("-"*88)
    
    for i, result in enumerate(results, 1):
        success = "✓" if result['success'] else "✗"
        has_gps = "✓" if result.get('has_gps', False) else "✗"
        location = result.get('location', '') if result.get('has_gps', False) else "-"
        print(f"{i:<5} {result['file']:<40} {success:<8} {has_gps:<8} {location:<25}")
    
    # Print Google Map URLs for images with GPS
    print("\n" + "="*60)
    print("包含GPS信息的图片地图链接:")
    for result in results:
        if result['success'] and result['has_gps']:
            print(f"{result['file']}: {result['google_map_url']}")

if __name__ == '__main__':
    main()
