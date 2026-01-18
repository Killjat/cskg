package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

func main() {
	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法: go run check_exif.go <图片文件路径>")
		os.Exit(1)
	}

	// 获取图片文件路径
	filePath := os.Args[1]

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("无法打开文件: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// 读取文件内容
	buffer, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		os.Exit(1)
	}

	// 解析EXIF数据
	x, err := exif.Decode(bytes.NewReader(buffer))
	if err != nil {
		fmt.Printf("无法解析EXIF数据: %v\n", err)
		os.Exit(1)
	}

	// 打印所有EXIF标签
	fmt.Println("=== 所有EXIF标签 ===")
	err = x.Walk(func(name exif.FieldName, tag *tiff.Tag) error {
		fmt.Printf("%s: %v\n", name, tag)
		return nil
	})

	if err != nil {
		fmt.Printf("遍历EXIF标签失败: %v\n", err)
	}

	// 专门检查GPS相关标签
	fmt.Println("\n=== GPS相关标签 ===")

	// GPSLatitude
	lat, err := x.Get(exif.GPSLatitude)
	if err != nil {
		fmt.Printf("GPSLatitude: 不存在\n")
	} else {
		fmt.Printf("GPSLatitude: %v\n", lat)
	}

	// GPSLatitudeRef
	latRef, err := x.Get(exif.GPSLatitudeRef)
	if err != nil {
		fmt.Printf("GPSLatitudeRef: 不存在\n")
	} else {
		fmt.Printf("GPSLatitudeRef: %v\n", latRef)
	}

	// GPSLongitude
	lon, err := x.Get(exif.GPSLongitude)
	if err != nil {
		fmt.Printf("GPSLongitude: 不存在\n")
	} else {
		fmt.Printf("GPSLongitude: %v\n", lon)
	}

	// GPSLongitudeRef
	lonRef, err := x.Get(exif.GPSLongitudeRef)
	if err != nil {
		fmt.Printf("GPSLongitudeRef: 不存在\n")
	} else {
		fmt.Printf("GPSLongitudeRef: %v\n", lonRef)
	}

	// GPSAltitude
	alt, err := x.Get(exif.GPSAltitude)
	if err != nil {
		fmt.Printf("GPSAltitude: 不存在\n")
	} else {
		fmt.Printf("GPSAltitude: %v\n", alt)
	}

	// GPSAltitudeRef
	altRef, err := x.Get(exif.GPSAltitudeRef)
	if err != nil {
		fmt.Printf("GPSAltitudeRef: 不存在\n")
	} else {
		fmt.Printf("GPSAltitudeRef: %v\n", altRef)
	}

	// GPSDateTime
	gpsTime, err := x.Get(exif.GPSDateTime)
	if err != nil {
		fmt.Printf("GPSDateTime: 不存在\n")
	} else {
		fmt.Printf("GPSDateTime: %v\n", gpsTime)
	}
}
