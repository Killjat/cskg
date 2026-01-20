package main

import (
	"fmt"
	"time"

	"github.com/cskg/CyberStroll/internal/scanner"
)

func main() {
	fmt.Println("🚀 CyberStroll 增强版扫描引擎测试")
	fmt.Println("==================================")

	// 创建增强版扫描器配置
	config := &scanner.ScannerConfig{
		MaxConcurrency: 20,
		Timeout:        5 * time.Second,
		RetryCount:     2,
		ProbeDelay:     100 * time.Millisecond,
		EnableLogging:  true,
	}

	// 创建增强版探测引擎
	engine := scanner.NewEnhancedProbeEngine(config)
	fmt.Println("✅ 增强版探测引擎初始化完成")
	fmt.Printf("   集成功能: Banner规则匹配 + Web指纹识别\n")
	fmt.Printf("   最大并发: %d\n", config.MaxConcurrency)

	// 测试目标
	targets := []struct {
		name string
		ip   string
		desc string
	}{
		{"本地回环", "127.0.0.1", "测试本地服务"},
		{"本地网关", "192.168.1.1", "测试路由器服务"},
	}

	// 测试不同的扫描类型
	scanTypes := []struct {
		name     string
		taskType string
		ports    []int
		enableApps bool
	}{
		{"快速扫描", "port_scan_default", []int{22, 80, 443, 8080}, false},
		{"Web应用识别", "app_identification", []int{80, 443, 8080, 8443}, true},
	}

	fmt.Printf("\n🎯 开始测试 %d 个目标，%d 种扫描类型...\n\n", len(targets), len(scanTypes))

	totalStartTime := time.Now()
	testCount := 0

	for _, target := range targets {
		fmt.Printf("🔍 目标: %s (%s)\n", target.name, target.ip)
		fmt.Printf("   描述: %s\n", target.desc)

		for _, scanType := range scanTypes {
			testCount++
			fmt.Printf("\n   [测试 %d] %s\n", testCount, scanType.name)
			fmt.Printf("   端口: %v\n", scanType.ports)

			// 创建扫描任务
			task := &scanner.ScanTask{
				TaskID:   fmt.Sprintf("enhanced-test-%03d", testCount),
				IP:       target.ip,
				TaskType: scanType.taskType,
				Priority: 1,
				Config: scanner.ScanConfig{
					Ports:      scanType.ports,
					Timeout:    5,
					ScanDepth:  "deep",
					EnableApps: scanType.enableApps,
				},
				Timestamp: time.Now().Unix(),
			}

			// 执行扫描
			startTime := time.Now()
			result, err := engine.ScanTarget(task)
			duration := time.Since(startTime)

			if err != nil {
				fmt.Printf("   ❌ 扫描失败: %v\n", err)
				continue
			}

			// 显示结果
			fmt.Printf("   ✅ 扫描完成 (耗时: %v)\n", duration)
			fmt.Printf("   📊 状态: %s, 响应时间: %dms\n", result.ScanStatus, result.ResponseTime)

			if len(result.Results.OpenPorts) > 0 {
				fmt.Printf("   🔓 发现 %d 个开放端口:\n", len(result.Results.OpenPorts))
				for _, port := range result.Results.OpenPorts {
					fmt.Printf("      • %d/%s", port.Port, port.Protocol)
					if port.Service != "unknown" {
						fmt.Printf(" (%s)", port.Service)
					}
					if port.Version != "" {
						fmt.Printf(" - %s", port.Version)
					}
					if port.Banner != "" && len(port.Banner) > 0 {
						bannerPreview := port.Banner
						if len(bannerPreview) > 50 {
							bannerPreview = bannerPreview[:50] + "..."
						}
						fmt.Printf(" [%s]", bannerPreview)
					}
					fmt.Println()
				}
			} else {
				fmt.Printf("   🔒 未发现开放端口\n")
			}

			// 显示应用识别结果
			if len(result.Results.Applications) > 0 {
				fmt.Printf("   🌐 识别到 %d 个Web应用:\n", len(result.Results.Applications))
				for _, app := range result.Results.Applications {
					fmt.Printf("      • %s", app.Name)
					if app.Version != "" {
						fmt.Printf(" v%s", app.Version)
					}
					if app.Category != "" {
						fmt.Printf(" (%s)", app.Category)
					}
					fmt.Printf(" [置信度: %d%%]", app.Confidence)
					fmt.Println()
				}
			}

			// 添加延迟
			time.Sleep(200 * time.Millisecond)
		}

		fmt.Println()
	}

	totalDuration := time.Since(totalStartTime)

	// 显示最终统计
	fmt.Println("📈 增强版扫描引擎测试报告")
	fmt.Println("============================")

	stats := engine.GetStats()
	fmt.Printf("总测试数: %d\n", testCount)
	fmt.Printf("引擎统计:\n")
	fmt.Printf("  - 总扫描: %d\n", stats.TotalScans)
	fmt.Printf("  - 成功: %d\n", stats.SuccessScans)
	fmt.Printf("  - 失败: %d\n", stats.FailedScans)
	fmt.Printf("  - 成功率: %.1f%%\n", float64(stats.SuccessScans)/float64(stats.TotalScans)*100)
	fmt.Printf("  - 平均时间: %dms\n", stats.AverageTime)
	fmt.Printf("总耗时: %v\n", totalDuration)

	fmt.Println("\n🎉 增强版引擎测试完成!")
	fmt.Println("\n✨ 增强功能特性:")
	fmt.Println("   ✅ 智能Banner规则匹配")
	fmt.Println("   ✅ 增强版协议探测")
	fmt.Println("   ✅ Web应用指纹识别")
	fmt.Println("   ✅ 服务版本提取")
	fmt.Println("   ✅ 结果缓存优化")

	fmt.Println("\n💡 集成模块:")
	fmt.Println("   📦 network_probe: 协议探测和解析")
	fmt.Println("   📦 rule_engine: Banner规则匹配")
	fmt.Println("   📦 servicefingerprint: Web应用识别")
}