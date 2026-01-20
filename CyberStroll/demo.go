package main

import (
	"fmt"
	"time"

	"github.com/cskg/CyberStroll/internal/scanner"
)

func main() {
	fmt.Println("🚀 CyberStroll 扫描节点演示")
	fmt.Println("============================")

	// 创建扫描器配置
	config := &scanner.ScannerConfig{
		MaxConcurrency: 20,
		Timeout:        3 * time.Second,
		RetryCount:     2,
		ProbeDelay:     50 * time.Millisecond,
		EnableLogging:  true,
	}

	// 创建探测引擎
	engine := scanner.NewProbeEngine(config)
	fmt.Println("✅ 探测引擎初始化完成")
	fmt.Printf("   最大并发: %d\n", config.MaxConcurrency)
	fmt.Printf("   超时时间: %v\n", config.Timeout)

	// 测试目标列表
	targets := []struct {
		name string
		ip   string
		desc string
	}{
		{"本地回环", "127.0.0.1", "测试本地服务"},
		{"本地网络", "192.168.1.1", "测试网关设备"},
		{"公共DNS", "8.8.8.8", "测试Google DNS"},
	}

	// 常见端口列表
	commonPorts := []int{22, 23, 53, 80, 135, 139, 443, 445, 993, 995, 3389, 8080}

	fmt.Printf("\n🎯 开始扫描 %d 个目标...\n", len(targets))
	fmt.Printf("📋 扫描端口: %v\n\n", commonPorts)

	totalStartTime := time.Now()
	
	for i, target := range targets {
		fmt.Printf("[%d/%d] 🔍 扫描目标: %s (%s)\n", i+1, len(targets), target.name, target.ip)
		fmt.Printf("      描述: %s\n", target.desc)

		// 创建扫描任务
		task := &scanner.ScanTask{
			TaskID:   fmt.Sprintf("demo-%03d", i+1),
			IP:       target.ip,
			TaskType: "port_scan_default",
			Priority: 1,
			Config: scanner.ScanConfig{
				Ports:     commonPorts,
				Timeout:   3,
				ScanDepth: "basic",
			},
			Timestamp: time.Now().Unix(),
		}

		// 执行扫描
		startTime := time.Now()
		result, err := engine.ScanTarget(task)
		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("      ❌ 扫描失败: %v\n", err)
			continue
		}

		// 显示结果
		fmt.Printf("      ✅ 扫描完成 (耗时: %v)\n", duration)
		fmt.Printf("      📊 状态: %s, 响应时间: %dms\n", result.ScanStatus, result.ResponseTime)
		
		if len(result.Results.OpenPorts) > 0 {
			fmt.Printf("      🔓 发现 %d 个开放端口:\n", len(result.Results.OpenPorts))
			for _, port := range result.Results.OpenPorts {
				fmt.Printf("         • %d/%s", port.Port, port.Protocol)
				if port.Service != "unknown" {
					fmt.Printf(" (%s)", port.Service)
				}
				if port.Version != "" {
					fmt.Printf(" - %s", port.Version)
				}
				fmt.Println()
			}
		} else {
			fmt.Printf("      🔒 未发现开放端口\n")
		}
		
		fmt.Println()
		
		// 添加延迟避免过于激进
		time.Sleep(500 * time.Millisecond)
	}

	totalDuration := time.Since(totalStartTime)

	// 显示最终统计
	fmt.Println("📈 扫描统计报告")
	fmt.Println("================")
	
	stats := engine.GetStats()
	fmt.Printf("总扫描任务: %d\n", stats.TotalScans)
	fmt.Printf("成功任务: %d\n", stats.SuccessScans)
	fmt.Printf("失败任务: %d\n", stats.FailedScans)
	fmt.Printf("成功率: %.1f%%\n", float64(stats.SuccessScans)/float64(stats.TotalScans)*100)
	fmt.Printf("平均扫描时间: %dms\n", stats.AverageTime)
	fmt.Printf("总耗时: %v\n", totalDuration)

	// 计算扫描速度
	totalPorts := len(targets) * len(commonPorts)
	portsPerSecond := float64(totalPorts) / totalDuration.Seconds()
	fmt.Printf("扫描速度: %.1f 端口/秒\n", portsPerSecond)

	fmt.Println("\n🎉 演示完成!")
	fmt.Println("\n💡 提示:")
	fmt.Println("   - 在生产环境中请调整并发数和超时时间")
	fmt.Println("   - 确保遵守网络扫描的相关法律法规")
	fmt.Println("   - 建议在授权的测试环境中使用")
}