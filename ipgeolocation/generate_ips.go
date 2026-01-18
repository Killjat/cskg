package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

func generateRandomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", 
		rand.Intn(256), 
		rand.Intn(256), 
		rand.Intn(256), 
		rand.Intn(256))
}

func main() {
	rand.Seed(time.Now().UnixNano())
	
	file, err := os.Create("100ips.txt")
	if err != nil {
		fmt.Println("创建文件失败:", err)
		os.Exit(1)
	}
	defer file.Close()
	
	// 添加一些已知的公共IP地址
	knownIPs := []string{
		"8.8.8.8",
		"1.1.1.1",
		"202.108.22.5",
		"114.114.114.114",
		"180.101.49.11",
		"199.91.73.222",
		"156.154.70.1",
		"192.30.255.112",
		"13.107.21.200",
		"204.79.197.200",
	}
	
	for _, ip := range knownIPs {
		file.WriteString(ip + "\n")
	}
	
	// 生成剩余的随机IP地址
	for i := 0; i < 90; i++ {
		ip := generateRandomIP()
		file.WriteString(ip + "\n")
	}
	
	fmt.Println("已生成100个IP地址到100ips.txt文件")
}