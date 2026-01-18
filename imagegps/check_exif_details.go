package main

import (
	"fmt"
	"os"

	"github.com/rwcarlsen/goexif/exif"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run check_exif_details.go <image_file>")
		os.Exit(1)
	}

	filePath := os.Args[1]

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("无法打开文件: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// 解析EXIF数据
	x, err := exif.Decode(file)
	if err != nil {
		fmt.Printf("无法解析EXIF数据: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== 基本EXIF数据 ===")

	// 检查基本相机信息
	make, _ := x.Get(exif.Make)
	model, _ := x.Get(exif.Model)
	dateTime, _ := x.DateTime()

	fmt.Printf("制造商: %v\n", make)
	fmt.Printf("型号: %v\n", model)
	fmt.Printf("拍摄时间: %v\n", dateTime)

	// 专门检查GPS标签
	fmt.Println("\n=== GPS标签详细信息 ===")
	
	// 检查所有可能的GPS标签
	fmt.Println("检查GPSLatitude...")
	lat, latErr := x.Get(exif.GPSLatitude)
	if latErr == nil {
		fmt.Printf("GPSLatitude: %v\n", lat)
	} else {
		fmt.Printf("GPSLatitude: 不存在\n")
	}

	fmt.Println("检查GPSLatitudeRef...")
	latRef, latRefErr := x.Get(exif.GPSLatitudeRef)
	if latRefErr == nil {
		fmt.Printf("GPSLatitudeRef: %v\n", latRef)
	} else {
		fmt.Printf("GPSLatitudeRef: 不存在\n")
	}

	fmt.Println("检查GPSLongitude...")
	lon, lonErr := x.Get(exif.GPSLongitude)
	if lonErr == nil {
		fmt.Printf("GPSLongitude: %v\n", lon)
	} else {
		fmt.Printf("GPSLongitude: 不存在\n")
	}

	fmt.Println("检查GPSLongitudeRef...")
	lonRef, lonRefErr := x.Get(exif.GPSLongitudeRef)
	if lonRefErr == nil {
		fmt.Printf("GPSLongitudeRef: %v\n", lonRef)
	} else {
		fmt.Printf("GPSLongitudeRef: 不存在\n")
	}

	fmt.Println("检查GPSAltitude...")
	alt, altErr := x.Get(exif.GPSAltitude)
	if altErr == nil {
		fmt.Printf("GPSAltitude: %v\n", alt)
	} else {
		fmt.Printf("GPSAltitude: 不存在\n")
	}

	// 检查是否有任何GPS数据
	if latErr == nil && lonErr == nil {
		fmt.Println("\n=== 发现GPS数据 ===")
		fmt.Printf("纬度: %v\n", lat)
		fmt.Printf("经度: %v\n", lon)
	} else {
		fmt.Println("\n=== 未发现GPS数据 ===")
	}
}
