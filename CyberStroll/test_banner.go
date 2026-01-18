package main

import (
	"fmt"
	"log"

	"github.com/cskg/CyberStroll/internal/config"
	"github.com/cskg/CyberStroll/internal/scan"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建扫描器
	scanner := scan.NewScanner(&cfg.Scan)

	// 测试grabBanner函数
	ip := "158.101.131.206"
	port := 80
	banner := scanner.GrabBanner(ip, port)

	fmt.Printf("IP: %s\n", ip)
	fmt.Printf("Port: %d\n", port)
	fmt.Printf("Banner: %q\n", banner)
	if banner == "" {
		fmt.Println("Banner is empty")
	} else {
		fmt.Println("Banner successfully captured!")
	}
}
