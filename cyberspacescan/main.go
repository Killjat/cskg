package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	// 命令行参数
	configFile := flag.String("c", "config.yaml", "配置文件路径")
	targetFile := flag.String("t", "targets.txt", "目标文件路径")
	outputDir := flag.String("o", "./results", "输出目录")
	workers := flag.Int("w", 100, "并发协程数")
	timeout := flag.Int("timeout", 2000, "超时时间(毫秒)")
	flag.Parse()

	// 加载配置
	config, err := LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 覆盖命令行参数
	if *targetFile != "targets.txt" {
		config.Targets.File = *targetFile
	}
	if *outputDir != "./results" {
		config.Output.Directory = *outputDir
	}
	if *workers != 100 {
		config.Scan.Workers = *workers
	}
	if *timeout != 2000 {
		config.Scan.Timeout = *timeout
	}

	// 打印配置信息
	fmt.Println("==========================================================")
	fmt.Println("网络空间扫描工具")
	fmt.Println("==========================================================")
	fmt.Printf("配置文件: %s\n", *configFile)
	fmt.Printf("目标文件: %s\n", config.Targets.File)
	fmt.Printf("输出目录: %s\n", config.Output.Directory)
	fmt.Printf("并发数: %d\n", config.Scan.Workers)
	fmt.Printf("超时: %dms\n", config.Scan.Timeout)
	fmt.Println("==========================================================\n")

	// 创建输出目录
	if err := os.MkdirAll(config.Output.Directory, 0755); err != nil {
		log.Fatalf("创建输出目录失败: %v", err)
	}

	// 读取目标列表
	targets, err := LoadTargets(config.Targets.File)
	if err != nil {
		log.Fatalf("读取目标文件失败: %v", err)
	}

	fmt.Printf("加载了 %d 个扫描目标\n\n", len(targets))

	// 创建扫描器
	scanner := NewScanner(config)

	// 开始扫描
	startTime := time.Now()
	results := scanner.Scan(targets)
	duration := time.Since(startTime)

	// 保存结果
	if err := SaveResults(config.Output.Directory, results, config.Output.Format); err != nil {
		log.Printf("保存结果失败: %v", err)
	}

	// 打印统计信息
	fmt.Println("\n==========================================================")
	fmt.Println("扫描完成统计")
	fmt.Println("==========================================================")
	fmt.Printf("总目标数: %d\n", len(targets))
	fmt.Printf("存活主机: %d\n", countAliveHosts(results))
	fmt.Printf("开放端口: %d\n", countOpenPorts(results))
	fmt.Printf("扫描耗时: %s\n", duration)
	fmt.Println("==========================================================")
}

func countAliveHosts(results []*ScanResult) int {
	count := 0
	for _, r := range results {
		if r.IsAlive {
			count++
		}
	}
	return count
}

func countOpenPorts(results []*ScanResult) int {
	count := 0
	for _, r := range results {
		count += len(r.TCPPorts) + len(r.UDPPorts)
	}
	return count
}
