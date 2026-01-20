package main

import (
	"fmt"
	"log"
	"time"

	"github.com/cskg/CyberStroll/internal/scanner"
)

func main() {
	fmt.Println("🧪 CyberStroll 本地测试")
	fmt.Println("========================")

	// 创建扫描器配置
	config := &scanner.ScannerConfig{
		MaxConcurrency: 10,
		Timeout:        5 * time.Second,
		RetryCount:     1,
		ProbeDelay:     100 * time.Millisecond,
		EnableLogging:  true,
	}

	// 创建探测引擎
	engine := scanner.NewProbeEngine(config)
	fmt.Println("✅ 探测引擎创建成功")

	// 创建测试任务
	task := &scanner.ScanTask{
		TaskID:   "test-001",
		IP:       "127.0.0.1",
		TaskType: "port_scan_default",
		Priority: 1,
		Config: scanner.ScanConfig{
			Ports:     []int{22, 80, 443, 8080},
			Timeout:   5,
			ScanDepth: "basic",
		},
		Timestamp: time.Now().Unix(),
	}

	fmt.Printf("🎯 开始扫描目标: %s\n", task.IP)
	fmt.Printf("📋 扫描端口: %v\n", task.Config.Ports)

	// 执行扫描
	startTime := time.Now()
	result, err := engine.ScanTarget(task)
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("❌ 扫描失败: %v", err)
		return
	}

	// 显示结果
	fmt.Printf("\n📊 扫描结果:\n")
	fmt.Printf("   状态: %s\n", result.ScanStatus)
	fmt.Printf("   耗时: %v\n", duration)
	fmt.Printf("   响应时间: %dms\n", result.ResponseTime)
	fmt.Printf("   开放端口数: %d\n", len(result.Results.OpenPorts))

	if len(result.Results.OpenPorts) > 0 {
		fmt.Println("\n🔓 开放端口详情:")
		for _, port := range result.Results.OpenPorts {
			fmt.Printf("   端口 %d/%s: %s", port.Port, port.Protocol, port.Service)
			if port.Version != "" {
				fmt.Printf(" (%s)", port.Version)
			}
			if port.Banner != "" {
				fmt.Printf(" - %s", port.Banner[:min(50, len(port.Banner))])
			}
			fmt.Println()
		}
	}

	// 显示统计信息
	stats := engine.GetStats()
	fmt.Printf("\n📈 引擎统计:\n")
	fmt.Printf("   总扫描: %d\n", stats.TotalScans)
	fmt.Printf("   成功: %d\n", stats.SuccessScans)
	fmt.Printf("   失败: %d\n", stats.FailedScans)
	fmt.Printf("   平均时间: %dms\n", stats.AverageTime)

	fmt.Println("\n✅ 测试完成!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}