#!/usr/bin/env python3

import os
import math
from PIL import Image, ImageDraw
import piexif

# 创建测试图片的目录
TEST_IMAGES_DIR = './test_images'
os.makedirs(TEST_IMAGES_DIR, exist_ok=True)

# GPS坐标数据（纬度，经度，地点名称）
gps_coordinates = [
    (39.9042, 116.4074, "北京天安门"),
    (31.2304, 121.4737, "上海外滩"),
    (23.1291, 113.2644, "广州塔"),
    (22.5431, 114.0579, "深圳市民中心"),
    (30.2741, 120.1551, "杭州西湖"),
    (32.0603, 118.7969, "南京夫子庙"),
    (31.8612, 117.2830, "合肥包公园"),
    (24.4798, 118.0894, "厦门鼓浪屿"),
    (25.0329, 102.7120, "昆明滇池"),
    (34.2632, 108.9511, "西安大雁塔")
]

def dms(degrees):
    """将十进制度数转换为度分秒格式"""
    d = int(degrees)
    m = int((degrees - d) * 60)
    s = (degrees - d - m / 60) * 3600
    return ((d, 1), (m, 1), (int(s * 10000), 10000))

def create_test_image(index, latitude, longitude, location_name):
    """创建带有GPS信息的测试图片"""
    # 创建一个200x200的红色背景图片
    img = Image.new('RGB', (200, 200), color='red')
    draw = ImageDraw.Draw(img)
    
    # 在图片上添加文字
    draw.text((10, 10), f"GPS Test Image {index+1}", fill='white')
    draw.text((10, 30), f"Location: {location_name}", fill='white')
    draw.text((10, 50), f"Lat: {latitude:.6f}", fill='white')
    draw.text((10, 70), f"Lon: {longitude:.6f}", fill='white')
    
    # 准备EXIF数据
    exif = {
        "0th": {
            piexif.ImageIFD.Make: "TestCamera",
            piexif.ImageIFD.Model: "TestModel",
        },
        "Exif": {
            piexif.ExifIFD.DateTimeOriginal: "2024:01:01 12:00:00",
        },
    }
    
    # 添加GPS信息
    lat_ref = "N" if latitude >= 0 else "S"
    lon_ref = "E" if longitude >= 0 else "W"
    
    gps_ifd = {
        piexif.GPSIFD.GPSVersionID: (2, 3, 0, 0),
        piexif.GPSIFD.GPSLatitudeRef: lat_ref,
        piexif.GPSIFD.GPSLatitude: dms(abs(latitude)),
        piexif.GPSIFD.GPSLongitudeRef: lon_ref,
        piexif.GPSIFD.GPSLongitude: dms(abs(longitude)),
        piexif.GPSIFD.GPSAltitudeRef: 0,
        piexif.GPSIFD.GPSAltitude: (100, 1),
    }
    
    exif["GPS"] = gps_ifd
    
    # 转换为EXIF二进制数据
    exif_bytes = piexif.dump(exif)
    
    # 保存图片
    filename = os.path.join(TEST_IMAGES_DIR, f"test_gps_{index+1}.jpg")
    img.save(filename, exif=exif_bytes, format='JPEG')
    
    print(f"创建测试图片: {filename}")
    print(f"  位置: {location_name}")
    print(f"  GPS: {latitude:.6f}, {longitude:.6f}")
    print()
    
    return filename

def main():
    """主函数"""
    print("=== 创建带有GPS信息的测试图片 ===")
    print(f"计划创建 {len(gps_coordinates)} 张测试图片")
    print(f"保存目录: {TEST_IMAGES_DIR}")
    print("\n" + "="*60)
    
    created_files = []
    
    for i, (lat, lon, name) in enumerate(gps_coordinates):
        filename = create_test_image(i, lat, lon, name)
        created_files.append(filename)
    
    print("="*60)
    print(f"已成功创建 {len(created_files)} 张测试图片")
    print("测试图片列表:")
    for filename in created_files:
        print(f"  - {os.path.basename(filename)}")
    
    return created_files

if __name__ == '__main__':
    main()
