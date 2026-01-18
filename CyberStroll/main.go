package main

import (
	"flag"
	"fmt"
	"os"
)

// CyberStroll 主程序入口
// 用于提供系统的整体入口和帮助信息

func main() {
	fmt.Println("=== CyberStroll - 网络空间安全漫步工具 ===")
	fmt.Println("Cyber Space Security Stroll Tool")
	fmt.Println("Version: 1.0.0")
	fmt.Println()
	fmt.Println("CyberStroll 是一款网络空间安全漫步工具，用于网络空间资产发现、扫描和分析。")
	fmt.Println()

	// 解析命令行参数
	var (
		help    bool
		version bool
	)

	flag.BoolVar(&help, "help", false, "显示帮助信息")
	flag.BoolVar(&version, "version", false, "显示版本信息")

	flag.Parse()

	if version {
		showVersion()
		os.Exit(0)
	}

	// 显示帮助信息
	showHelp()
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Println("使用方法:")
	fmt.Println("  go run main.go [选项]")
	fmt.Println("  ./cyberstroll [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -help    显示帮助信息")
	fmt.Println("  -version 显示版本信息")
	fmt.Println()
	fmt.Println("模块说明:")
	fmt.Println("  1. 任务管理模块 (Task Manager)")
	fmt.Println("     负责向系统下发扫描任务到Kafka")
	fmt.Println("     使用方法: go run cmd/task_manager/main.go -help")
	fmt.Println()
	fmt.Println("  2. 扫描节点模块 (Scan Node)")
	fmt.Println("     负责读取Kafka任务并执行扫描，将结果保存到Elasticsearch")
	fmt.Println("     使用方法: go run cmd/scan_node/main.go -help")
	fmt.Println()
	fmt.Println("  3. 一键搜索模块 (Search)")
	fmt.Println("     负责从Elasticsearch查询扫描结果")
	fmt.Println("     使用方法: go run cmd/search/main.go -help")
	fmt.Println()
	fmt.Println("启动脚本:")
	fmt.Println("  使用 start.sh 脚本可以便捷地启动各个模块")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  # 启动任务管理模块")
	fmt.Println("  go run cmd/task_manager/main.go -targets 192.168.1.1,192.168.1.2 -type port_scan")
	fmt.Println()
	fmt.Println("  # 启动扫描节点模块")
	fmt.Println("  go run cmd/scan_node/main.go")
	fmt.Println()
	fmt.Println("  # 启动一键搜索模块")
	fmt.Println("  go run cmd/search/main.go -query http")
}

// showVersion 显示版本信息
func showVersion() {
	fmt.Println("CyberStroll v1.0.0")
	fmt.Println("网络空间安全漫步工具")
	fmt.Println("Copyright (c) 2026 CSKG Project Team")
}
