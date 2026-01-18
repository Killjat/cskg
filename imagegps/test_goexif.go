package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

func main() {
	// 测试图片路径
	imagePath := "./test_images/test_gps_1.jpg"

	// 打开图片文件
	f, err := os.Open(imagePath)
	if err != nil {
		log.Fatalf("无法打开文件: %v", err)
	}
	defer f.Close()

	// 解析EXIF数据
	x, err := exif.Decode(f)
	if err != nil {
		log.Fatalf("无法解析EXIF数据: %v", err)
	}

	// 打印所有EXIF标签
	fmt.Println("=== 所有EXIF标签 ===")
	// 使用Get方法获取一些基本标签
	basicTags := []exif.FieldName{
		"Make",
		"Model",
		"DateTime",
		"GPSLatitude",
		"GPSLatitudeRef",
		"GPSLongitude",
		"GPSLongitudeRef",
		"GPSAltitude",
		"GPSAltitudeRef",
		"GPSVersionID",
	}

	for _, tagName := range basicTags {
		tag, err := x.Get(tagName)
		if err != nil {
			fmt.Printf("%s: 获取失败 - %v\n", tagName, err)
		} else {
			fmt.Printf("%s: %v\n", tagName, tag)
		}
	}

	// 尝试直接获取GPS相关标签
	fmt.Println("\n=== 直接获取GPS标签 ===")

	// GPSLatitude
	lat, err := x.Get(exif.GPSLatitude)
	if err != nil {
		fmt.Printf("获取GPSLatitude失败: %v\n", err)
	} else {
		fmt.Printf("GPSLatitude: %v\n", lat)
		// 尝试获取Rat2值
		deg, degDenom, err := lat.Rat2(0)
		if err != nil {
			fmt.Printf("获取GPSLatitude度失败: %v\n", err)
		} else {
			fmt.Printf("  度: %d/%d\n", deg, degDenom)
		}

		min, minDenom, err := lat.Rat2(1)
		if err != nil {
			fmt.Printf("获取GPSLatitude分失败: %v\n", err)
		} else {
			fmt.Printf("  分: %d/%d\n", min, minDenom)
		}

		sec, secDenom, err := lat.Rat2(2)
		if err != nil {
			fmt.Printf("获取GPSLatitude秒失败: %v\n", err)
		} else {
			fmt.Printf("  秒: %d/%d\n", sec, secDenom)
		}
	}

	// GPSLongitude
	lon, err := x.Get(exif.GPSLongitude)
	if err != nil {
		fmt.Printf("获取GPSLongitude失败: %v\n", err)
	} else {
		fmt.Printf("GPSLongitude: %v\n", lon)
	}

	// GPSLatitudeRef
	latRef, err := x.Get(exif.GPSLatitudeRef)
	if err != nil {
		fmt.Printf("获取GPSLatitudeRef失败: %v\n", err)
	} else {
		latRefStr, _ := latRef.StringVal()
		fmt.Printf("GPSLatitudeRef: %s\n", latRefStr)
	}

	// GPSLongitudeRef
	lonRef, err := x.Get(exif.GPSLongitudeRef)
	if err != nil {
		fmt.Printf("获取GPSLongitudeRef失败: %v\n", err)
	} else {
		lonRefStr, _ := lonRef.StringVal()
		fmt.Printf("GPSLongitudeRef: %s\n", lonRefStr)
	}

	// 尝试获取其他GPS标签
	fmt.Println("\n=== 其他GPS标签 ===")

	// GPSVersionID
	ver, err := x.Get(exif.GPSVersionID)
	if err != nil {
		fmt.Printf("获取GPSVersionID失败: %v\n", err)
	} else {
		fmt.Printf("GPSVersionID: %v\n", ver)
	}

	// GPSAltitude
	alt, err := x.Get(exif.GPSAltitude)
	if err != nil {
		fmt.Printf("获取GPSAltitude失败: %v\n", err)
	} else {
		fmt.Printf("GPSAltitude: %v\n", alt)
	}

	// GPSAltitudeRef
	altRef, err := x.Get(exif.GPSAltitudeRef)
	if err != nil {
		fmt.Printf("获取GPSAltitudeRef失败: %v\n", err)
	} else {
		fmt.Printf("GPSAltitudeRef: %v\n", altRef)
	}
}
