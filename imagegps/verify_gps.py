#!/usr/bin/env python3

import os
import piexif
from PIL import Image

# Directory containing test images
test_dir = './test_images'

# Get all image files
image_files = [f for f in os.listdir(test_dir) if f.endswith('.jpg')]

for filename in image_files:
    filepath = os.path.join(test_dir, filename)
    print(f"=== 验证图片: {filename} ===")
    
    # 1. 使用piexif查看GPS信息
    print("\n1. 使用piexif查看GPS信息:")
    try:
        exif_dict = piexif.load(filepath)
        if 'GPS' in exif_dict and exif_dict['GPS']:
            print(f"   GPS信息存在:")
            for tag, value in exif_dict['GPS'].items():
                tag_name = piexif.TAGS['GPS'][tag]['name']
                print(f"   {tag_name}: {value}")
        else:
            print(f"   未找到GPS信息")
    except Exception as e:
        print(f"   读取失败: {e}")
    
    # 2. 检查PIL是否能读取GPS信息
    print("\n2. 使用PIL查看EXIF信息:")
    try:
        with Image.open(filepath) as img:
            exif_data = img._getexif()
            if exif_data:
                print(f"   EXIF信息存在，共{len(exif_data)}项")
                # 查找GPS相关的EXIF标签
                gps_tags = {}
                for tag, value in exif_data.items():
                    if 0x8825 <= tag <= 0x8834:  # GPS相关标签范围
                        gps_tags[tag] = value
                if gps_tags:
                    print(f"   找到{len(gps_tags)}个GPS相关标签:")
                    for tag, value in gps_tags.items():
                        print(f"   0x{tag:04x}: {value}")
                else:
                    print(f"   未找到GPS相关标签")
            else:
                print(f"   未找到EXIF信息")
    except Exception as e:
        print(f"   读取失败: {e}")
    
    print("\n" + "="*50 + "\n")
