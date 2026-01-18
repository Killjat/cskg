package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rwcarlsen/goexif/exif"
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

	fmt.Println("=== 测试goexif库读取GPS信息 ===")

	// 尝试获取GPSLatitude
	lat, err := x.Get("GPSLatitude")
	if err != nil {
		fmt.Printf("获取GPSLatitude失败: %v\n", err)
	} else {
		fmt.Printf("GPSLatitude: %v\n", lat)
	}

	// 尝试获取GPSLatitudeRef
	latRef, err := x.Get("GPSLatitudeRef")
	if err != nil {
		fmt.Printf("获取GPSLatitudeRef失败: %v\n", err)
	} else {
		fmt.Printf("GPSLatitudeRef: %v\n", latRef)
	}

	// 尝试获取GPSLongitude
	lon, err := x.Get("GPSLongitude")
	if err != nil {
		fmt.Printf("获取GPSLongitude失败: %v\n", err)
	} else {
		fmt.Printf("GPSLongitude: %v\n", lon)
	}

	// 尝试获取GPSLongitudeRef
	lonRef, err := x.Get("GPSLongitudeRef")
	if err != nil {
		fmt.Printf("获取GPSLongitudeRef失败: %v\n", err)
	} else {
		fmt.Printf("GPSLongitudeRef: %v\n", lonRef)
	}

	// 使用exif包提供的常量获取GPS标签
	fmt.Println("\n=== 使用exif包常量获取GPS标签 ===")

	// 使用exif.GPSLatitude常量
	latConst, err := x.Get(exif.GPSLatitude)
	if err != nil {
		fmt.Printf("使用exif.GPSLatitude常量获取失败: %v\n", err)
	} else {
		fmt.Printf("exif.GPSLatitude: %v\n", latConst)
	}

	// 尝试获取DateTime
	dateTime, err := x.DateTime()
	if err != nil {
		fmt.Printf("获取DateTime失败: %v\n", err)
	} else {
		fmt.Printf("DateTime: %v\n", dateTime)
	}

	// 尝试获取Make
	make, err := x.Get("Make")
	if err != nil {
		fmt.Printf("获取Make失败: %v\n", err)
	} else {
		fmt.Printf("Make: %v\n", make)
	}

	// 尝试获取Model
	model, err := x.Get("Model")
	if err != nil {
		fmt.Printf("获取Model失败: %v\n", err)
	} else {
		fmt.Printf("Model: %v\n", model)
	}
}
