package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// 命令行参数
var (
	target     = flag.String("target", "", "目标地址 (格式: host:port)")
	host       = flag.String("host", "", "目标主机")
	port       = flag.Int("port", 0, "目标端口")
	timeout    = flag.Duration("timeout", 10*time.Second, "探测超时时间")
	concurrent = flag.Int("concurrent", 10, "并发数")
	output     = flag.String("output", "text", "输出格式 (text/json)")
	verbose    = flag.Bool("verbose", false, "详细输出")
	probeList  = flag.Bool("list-probes", false, "列出所有可用探测")
	stats      = flag.Bool("stats", false, "显示统计信息")
	probeMode  = flag.String("probe-mode", "all", "探测模式 (port/all/smart)")
	protocolStats = flag.Bool("protocol-stats", false, "显示协议支持统计")
)

func main() {
	flag.Parse()
	
	fmt.Println("🔍 网络探测引擎")
	fmt.Println("=" + strings.Repeat("=", 30))
	
	// 创建探测引擎
	config := DefaultProbeConfig()
	config.Timeout = *timeout
	config.MaxConcurrency = *concurrent
	config.EnableLogging = *verbose
	
	engine := NewProbeEngine(config)
	
	// 列出探测
	if *probeList {
		listProbes(engine)
		return
	}
	
	// 解析目标
	targets, err := parseTargets()
	if err != nil {
		fmt.Printf("❌ 目标解析错误: %v\n", err)
		printUsage()
		return
	}
	
	if len(targets) == 0 {
		fmt.Println("❌ 未指定目标")
		printUsage()
		return
	}
	
	fmt.Printf("🎯 开始探测 %d 个目标...\n\n", len(targets))
	
	// 执行探测
	start := time.Now()
	
	if len(targets) == 1 {
		// 单目标探测
		results, err := engine.ProbeTargetWithMode(targets[0], *probeMode)
		if err != nil {
			fmt.Printf("❌ 探测失败: %v\n", err)
			return
		}
		
		outputResults(map[string][]*ProbeResult{
			fmt.Sprintf("%s:%d", targets[0].Host, targets[0].Port): results,
		})
	} else {
		// 多目标探测
		allResults, err := engine.ProbeMultipleTargetsWithMode(targets, *probeMode)
		if err != nil {
			fmt.Printf("❌ 探测失败: %v\n", err)
			return
		}
		
		outputResults(allResults)
	}
	
	duration := time.Since(start)
	
	// 显示统计信息
	if *stats {
		fmt.Println("\n📊 探测统计:")
		fmt.Println(strings.Repeat("-", 40))
		engineStats := engine.GetStats()
		fmt.Printf("总探测数: %d\n", engineStats.TotalProbes)
		fmt.Printf("成功探测: %d\n", engineStats.SuccessProbes)
		fmt.Printf("失败探测: %d\n", engineStats.FailedProbes)
		fmt.Printf("成功率: %.1f%%\n", float64(engineStats.SuccessProbes)/float64(engineStats.TotalProbes)*100)
		fmt.Printf("平均耗时: %v\n", engineStats.AvgDuration)
		fmt.Printf("总耗时: %v\n", duration)
		
		if len(engineStats.ProtocolCounts) > 0 {
			fmt.Println("\n协议分布:")
			for protocol, count := range engineStats.ProtocolCounts {
				fmt.Printf("  %s: %d\n", protocol, count)
			}
		}
	}
}

// parseTargets 解析目标参数
func parseTargets() ([]Target, error) {
	var targets []Target
	
	// 优先使用 -target 参数
	if *target != "" {
		parts := strings.Split(*target, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("目标格式错误，应为 host:port")
		}
		
		portNum, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("端口格式错误: %v", err)
		}
		
		targets = append(targets, Target{
			Host: parts[0],
			Port: portNum,
		})
	} else if *host != "" && *port != 0 {
		// 使用 -host 和 -port 参数
		targets = append(targets, Target{
			Host: *host,
			Port: *port,
		})
	}
	
	return targets, nil
}

// listProbes 列出所有可用探测
func listProbes(engine *ProbeEngine) {
	loader := NewProbeLoader()
	probes := loader.LoadBuiltinProbes()
	
	fmt.Printf("📚 可用探测 (%d 个):\n", len(probes))
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-15s %-8s %-12s %-8s %s\n", "名称", "类型", "协议", "稀有度", "描述")
	fmt.Println(strings.Repeat("-", 80))
	
	for _, probe := range probes {
		fmt.Printf("%-15s %-8s %-12s %-8d %s\n",
			probe.Name,
			probe.Type,
			probe.Protocol,
			probe.Rarity,
			probe.Description)
	}
	
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("总计: %d 个探测\n", len(probes))
}

// outputResults 输出探测结果
func outputResults(allResults map[string][]*ProbeResult) {
	if *output == "json" {
		outputJSON(allResults)
	} else {
		outputText(allResults)
	}
}

// outputText 文本格式输出
func outputText(allResults map[string][]*ProbeResult) {
	for target, results := range allResults {
		fmt.Printf("🎯 目标: %s\n", target)
		fmt.Println(strings.Repeat("-", 60))
		
		if len(results) == 0 {
			fmt.Println("❌ 无响应")
			continue
		}
		
		successCount := 0
		for _, result := range results {
			if result.Success {
				successCount++
			}
		}
		
		fmt.Printf("✅ 成功探测: %d/%d\n\n", successCount, len(results))
		
		for i, result := range results {
			if !result.Success {
				if *verbose {
					fmt.Printf("%d. ❌ %s (%s) - %s (耗时: %v)\n",
						i+1, result.ProbeName, result.Protocol, result.Error, result.Duration)
				}
				continue
			}
			
			fmt.Printf("%d. ✅ %s (%s) - 耗时: %v\n",
				i+1, result.ProbeName, result.Protocol, result.Duration)
			
			if result.Banner != "" {
				fmt.Printf("   📄 Banner: %q\n", result.Banner)
			}
			
			if result.ParsedInfo != nil {
				info := result.ParsedInfo
				if info.Product != "" {
					fmt.Printf("   🏷️  产品: %s", info.Product)
					if info.Version != "" {
						fmt.Printf(" v%s", info.Version)
					}
					fmt.Printf(" (置信度: %d%%)\n", info.Confidence)
				}
				
				if info.Service != "" && info.Service != result.Protocol {
					fmt.Printf("   🔧 服务: %s\n", info.Service)
				}
				
				if len(info.Fields) > 0 && *verbose {
					fmt.Printf("   📋 字段:\n")
					for key, value := range info.Fields {
						fmt.Printf("      %s: %s\n", key, value)
					}
				}
			}
			
			if *verbose && len(result.Response) > 0 {
				fmt.Printf("   🔍 原始响应 (%d bytes): %s\n", 
					len(result.Response), result.ResponseHex)
			}
			
			fmt.Println()
		}
		
		fmt.Println()
	}
}

// outputJSON JSON格式输出
func outputJSON(allResults map[string][]*ProbeResult) {
	output := map[string]interface{}{
		"results":   allResults,
		"timestamp": time.Now().Format(time.RFC3339),
		"summary": map[string]interface{}{
			"total_targets": len(allResults),
		},
	}
	
	// 计算总体统计
	totalProbes := 0
	successProbes := 0
	for _, results := range allResults {
		totalProbes += len(results)
		for _, result := range results {
			if result.Success {
				successProbes++
			}
		}
	}
	
	output["summary"].(map[string]interface{})["total_probes"] = totalProbes
	output["summary"].(map[string]interface{})["success_probes"] = successProbes
	if totalProbes > 0 {
		output["summary"].(map[string]interface{})["success_rate"] = float64(successProbes) / float64(totalProbes) * 100
	}
	
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Printf("❌ JSON编码错误: %v\n", err)
		return
	}
	
	fmt.Println(string(jsonData))
}

// printUsage 打印使用说明
func printUsage() {
	fmt.Println("\n📖 使用方法:")
	fmt.Println("  go run . -target host:port")
	fmt.Println("  go run . -host 192.168.1.1 -port 80")
	fmt.Println("  go run . -target 192.168.1.1:22 -verbose")
	fmt.Println("  go run . -list-probes")
	fmt.Println()
	fmt.Println("📋 参数说明:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("🔍 探测模式说明:")
	fmt.Println("  port  - 仅使用端口相关的探测（快速）")
	fmt.Println("  all   - 使用所有探测包试探（全面，默认）")
	fmt.Println("  smart - 智能模式，优先使用常见探测")
	fmt.Println()
	fmt.Println("🌰 示例:")
	fmt.Println("  # 全面探测（发送所有探测包）")
	fmt.Println("  go run . -target 192.168.1.1:22 -probe-mode all")
	fmt.Println("  # 快速探测（仅端口相关）")
	fmt.Println("  go run . -target baidu.com:80 -probe-mode port")
	fmt.Println("  # 智能探测（优先常见协议）")
	fmt.Println("  go run . -target 127.0.0.1:8080 -probe-mode smart")
	fmt.Println("  # 探测非标准端口服务（如22端口的HTTP）")
	fmt.Println("  go run . -target example.com:22 -probe-mode all -verbose")
}