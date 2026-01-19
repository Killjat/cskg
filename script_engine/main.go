package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// 版本信息
const (
	Version = "1.0.0"
	Author  = "Script Engine Team"
)

// 命令行参数
var (
	target     = flag.String("target", "", "目标地址 (格式: host:port)")
	targets    = flag.String("targets", "", "目标文件路径")
	protocol   = flag.String("protocol", "", "协议类型 (modbus, redis, mqtt等)")
	scripts    = flag.String("scripts", "all", "要执行的脚本 (all, info, vuln, auth)")
	category   = flag.String("category", "", "脚本类别 (discovery, vulnerability, authentication, exploitation)")
	output     = flag.String("output", "text", "输出格式 (text, json, xml)")
	outputFile = flag.String("output-file", "", "输出文件路径")
	verbose    = flag.Bool("verbose", false, "详细输出")
	timeout    = flag.Duration("timeout", 30*time.Second, "脚本执行超时时间")
	concurrent = flag.Int("concurrent", 10, "并发执行数")
	listScripts = flag.Bool("list-scripts", false, "列出所有可用脚本")
	autoDetect = flag.Bool("auto-detect", false, "自动检测协议并执行相应脚本")
	help       = flag.Bool("help", false, "显示帮助信息")
	version    = flag.Bool("version", false, "显示版本信息")
)

func main() {
	flag.Parse()

	// 显示版本信息
	if *version {
		fmt.Printf("Script Engine v%s\n", Version)
		fmt.Printf("Author: %s\n", Author)
		fmt.Printf("Build: %s\n", time.Now().Format("2006-01-02"))
		return
	}

	// 显示帮助信息
	if *help {
		printHelp()
		return
	}

	fmt.Println("🚀 Script Engine - 深度协议探测脚本系统")
	fmt.Println(strings.Repeat("=", 50))

	// 创建脚本引擎
	engine := NewScriptEngine(&ScriptConfig{
		Timeout:     *timeout,
		Concurrent:  *concurrent,
		Verbose:     *verbose,
		OutputFormat: *output,
	})

	// 列出所有脚本
	if *listScripts {
		listAllScripts(engine)
		return
	}

	// 解析目标
	targetList, err := parseTargets()
	if err != nil {
		fmt.Printf("❌ 目标解析错误: %v\n", err)
		printUsage()
		return
	}

	if len(targetList) == 0 {
		fmt.Println("❌ 未指定目标")
		printUsage()
		return
	}

	fmt.Printf("🎯 开始深度探测 %d 个目标...\n\n", len(targetList))

	// 执行脚本
	start := time.Now()
	results := make(map[string]*TargetResult)

	for _, t := range targetList {
		fmt.Printf("📡 探测目标: %s\n", t.String())
		
		result, err := engine.ExecuteScripts(t, *protocol, *scripts, *category)
		if err != nil {
			fmt.Printf("❌ 探测失败: %v\n", err)
			continue
		}
		
		results[t.String()] = result
		
		// 输出结果
		if *output == "text" {
			printTextResult(result)
		}
	}

	duration := time.Since(start)
	fmt.Printf("\n⏱️  总耗时: %v\n", duration)

	// 保存结果到文件
	if *outputFile != "" {
		err := saveResults(results, *outputFile, *output)
		if err != nil {
			fmt.Printf("⚠️  保存结果失败: %v\n", err)
		} else {
			fmt.Printf("💾 结果已保存到: %s\n", *outputFile)
		}
	}

	// 显示统计信息
	printStatistics(results)
}

// parseTargets 解析目标列表
func parseTargets() ([]Target, error) {
	var targetList []Target

	// 单个目标
	if *target != "" {
		t, err := ParseTarget(*target)
		if err != nil {
			return nil, err
		}
		targetList = append(targetList, t)
	}

	// 目标文件
	if *targets != "" {
		fileTargets, err := LoadTargetsFromFile(*targets)
		if err != nil {
			return nil, err
		}
		targetList = append(targetList, fileTargets...)
	}

	return targetList, nil
}

// listAllScripts 列出所有可用脚本
func listAllScripts(engine *ScriptEngine) {
	fmt.Println("📋 可用脚本列表:")
	fmt.Println(strings.Repeat("-", 80))

	scripts := engine.GetAllScripts()
	
	// 按协议分组
	protocolGroups := make(map[string][]*Script)
	for _, script := range scripts {
		protocolGroups[script.Protocol] = append(protocolGroups[script.Protocol], script)
	}

	for protocol, scriptList := range protocolGroups {
		fmt.Printf("\n🔍 %s 协议脚本:\n", strings.ToUpper(protocol))
		for _, script := range scriptList {
			fmt.Printf("  %-20s %-12s %s\n", 
				script.Name, 
				fmt.Sprintf("[%s]", script.Category), 
				script.Description)
		}
	}

	fmt.Printf("\n📊 统计: %d 个协议, %d 个脚本\n", len(protocolGroups), len(scripts))
}

// printTextResult 打印文本格式结果
func printTextResult(result *TargetResult) {
	fmt.Printf("🎯 目标: %s (%s)\n", result.Target, result.Protocol)
	fmt.Printf("📊 执行脚本: %d个\n", len(result.ScriptResults))
	
	successCount := 0
	for _, sr := range result.ScriptResults {
		if sr.Success {
			successCount++
		}
	}
	
	fmt.Printf("✅ 成功: %d个\n", successCount)
	fmt.Printf("❌ 失败: %d个\n", len(result.ScriptResults)-successCount)

	// 显示发现信息
	if len(result.Findings) > 0 {
		fmt.Println("\n📋 发现信息:")
		for key, value := range result.Findings {
			fmt.Printf("  🏷️  %s: %v\n", key, value)
		}
	}

	// 显示漏洞信息
	if len(result.Vulnerabilities) > 0 {
		fmt.Println("\n🚨 安全漏洞:")
		for _, vuln := range result.Vulnerabilities {
			fmt.Printf("  ⚠️  %s (%s)\n", vuln.CVE, vuln.Severity)
			fmt.Printf("      %s\n", vuln.Description)
			if vuln.ExploitAvailable {
				fmt.Printf("      💥 存在可用漏洞利用\n")
			}
		}
	}

	// 显示脚本执行详情
	if *verbose {
		fmt.Println("\n🔍 脚本执行详情:")
		for _, sr := range result.ScriptResults {
			status := "✅"
			if !sr.Success {
				status = "❌"
			}
			fmt.Printf("  %s %-20s [%s] (耗时: %v)\n", 
				status, sr.ScriptName, sr.Category, sr.Duration)
			
			if !sr.Success && sr.Error != "" {
				fmt.Printf("      错误: %s\n", sr.Error)
			}
		}
	}

	fmt.Println()
}

// saveResults 保存结果到文件
func saveResults(results map[string]*TargetResult, filename, format string) error {
	var data []byte
	var err error

	switch format {
	case "json":
		data, err = json.MarshalIndent(results, "", "  ")
	case "xml":
		// TODO: 实现XML格式
		return fmt.Errorf("XML格式暂未实现")
	default:
		return fmt.Errorf("不支持的输出格式: %s", format)
	}

	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// printStatistics 打印统计信息
func printStatistics(results map[string]*TargetResult) {
	fmt.Println("\n📊 执行统计:")
	fmt.Println(strings.Repeat("-", 40))

	totalTargets := len(results)
	totalScripts := 0
	successfulScripts := 0
	totalVulns := 0

	protocolStats := make(map[string]int)
	categoryStats := make(map[string]int)

	for _, result := range results {
		totalScripts += len(result.ScriptResults)
		totalVulns += len(result.Vulnerabilities)
		
		protocolStats[result.Protocol]++
		
		for _, sr := range result.ScriptResults {
			if sr.Success {
				successfulScripts++
			}
			categoryStats[sr.Category]++
		}
	}

	fmt.Printf("🎯 目标总数: %d\n", totalTargets)
	fmt.Printf("📜 脚本总数: %d\n", totalScripts)
	fmt.Printf("✅ 成功执行: %d (%.1f%%)\n", 
		successfulScripts, 
		float64(successfulScripts)/float64(totalScripts)*100)
	fmt.Printf("🚨 发现漏洞: %d\n", totalVulns)

	if len(protocolStats) > 0 {
		fmt.Println("\n📋 协议分布:")
		for protocol, count := range protocolStats {
			fmt.Printf("  %s: %d\n", protocol, count)
		}
	}

	if len(categoryStats) > 0 {
		fmt.Println("\n📋 脚本类别分布:")
		for category, count := range categoryStats {
			fmt.Printf("  %s: %d\n", category, count)
		}
	}
}

// printUsage 打印使用说明
func printUsage() {
	fmt.Println("\n📖 使用方法:")
	fmt.Println("  script_engine -target host:port -protocol modbus")
	fmt.Println("  script_engine -targets targets.txt -auto-detect")
	fmt.Println("  script_engine -target 192.168.1.100:502 -scripts info,vuln")
	fmt.Println()
	fmt.Println("📋 参数说明:")
	flag.PrintDefaults()
}

// printHelp 打印详细帮助
func printHelp() {
	fmt.Printf("Script Engine v%s - 深度协议探测脚本系统\n\n", Version)
	
	fmt.Println("📖 使用方法:")
	fmt.Println("  script_engine [选项] -target <目标>")
	fmt.Println("  script_engine [选项] -targets <目标文件>")
	fmt.Println()
	
	fmt.Println("🎯 基本示例:")
	fmt.Println("  # 对Modbus设备进行深度探测")
	fmt.Println("  script_engine -target 192.168.1.100:502 -protocol modbus")
	fmt.Println()
	fmt.Println("  # 对Redis服务器进行漏洞扫描")
	fmt.Println("  script_engine -target 192.168.1.100:6379 -protocol redis -category vulnerability")
	fmt.Println()
	fmt.Println("  # 批量扫描并自动检测协议")
	fmt.Println("  script_engine -targets targets.txt -auto-detect")
	fmt.Println()
	fmt.Println("  # 执行特定脚本")
	fmt.Println("  script_engine -target 192.168.1.100:502 -scripts modbus-info,modbus-vuln")
	fmt.Println()
	
	fmt.Println("📋 参数说明:")
	flag.PrintDefaults()
	fmt.Println()
	
	fmt.Println("🔍 支持的协议:")
	fmt.Println("  工控: modbus, dnp3, bacnet, opcua, s7")
	fmt.Println("  数据库: mysql, redis, mongodb, postgresql, oracle")
	fmt.Println("  IoT: mqtt, coap, lorawan, amqp")
	fmt.Println("  企业: kerberos, ldap, radius, ntp")
	fmt.Println("  网络: http, https, ssh, ftp, smtp, dns, snmp")
	fmt.Println()
	
	fmt.Println("📂 脚本类别:")
	fmt.Println("  discovery      - 信息收集和服务发现")
	fmt.Println("  vulnerability  - 漏洞检测和安全评估")
	fmt.Println("  authentication - 认证测试和暴力破解")
	fmt.Println("  exploitation   - 漏洞利用和渗透测试")
	fmt.Println()
	
	fmt.Println("📄 输出格式:")
	fmt.Println("  text - 人类可读的文本格式 (默认)")
	fmt.Println("  json - 结构化JSON格式")
	fmt.Println("  xml  - XML格式 (计划中)")
	fmt.Println()
	
	fmt.Println("🌰 高级用法:")
	fmt.Println("  # 详细输出并保存JSON结果")
	fmt.Println("  script_engine -target 192.168.1.100:502 -protocol modbus -verbose -output json -output-file result.json")
	fmt.Println()
	fmt.Println("  # 列出所有可用脚本")
	fmt.Println("  script_engine -list-scripts")
	fmt.Println()
	fmt.Println("  # 高并发扫描")
	fmt.Println("  script_engine -targets large_targets.txt -concurrent 50 -timeout 10s")
}