package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

var (
	banner      = flag.String("banner", "", "要匹配的Banner字符串")
	rulesFile   = flag.String("rules", "", "Nmap规则文件路径")
	interactive = flag.Bool("interactive", false, "交互模式")
	output      = flag.String("output", "text", "输出格式 (text/json)")
	rulesDir    = flag.String("rules-dir", "./rules", "用户规则目录")
	confidence  = flag.Int("min-confidence", 0, "最小置信度")
)

func main() {
	flag.Parse()
	
	fmt.Println("🔍 Banner指纹识别引擎")
	fmt.Println("支持Nmap规则库和用户自定义规则")
	fmt.Println("=" + strings.Repeat("=", 40))
	
	// 创建引擎
	config := DefaultConfig()
	config.RulesDir = *rulesDir
	engine := NewBannerEngine(config)
	
	// 创建加载器
	nmapLoader := NewNmapLoader(engine)
	
	// 加载规则
	fmt.Println("📚 正在加载规则...")
	
	// 1. 加载内置规则
	if err := nmapLoader.LoadBuiltinRules(); err != nil {
		fmt.Printf("❌ 加载内置规则失败: %v\n", err)
	} else {
		fmt.Println("✅ 已加载内置规则")
	}
	
	// 2. 加载Nmap规则文件
	if *rulesFile != "" {
		fmt.Printf("📁 加载Nmap规则文件: %s\n", *rulesFile)
		if err := nmapLoader.LoadFromFile(*rulesFile); err != nil {
			fmt.Printf("❌ 加载Nmap规则失败: %v\n", err)
		} else {
			fmt.Println("✅ 已加载Nmap规则")
		}
	}
	
	stats := engine.GetStats()
	fmt.Printf("📊 总共加载了 %d 条规则\n\n", stats.TotalRules)
	
	if *interactive {
		runInteractiveMode(engine)
	} else if *banner != "" {
		runSingleMatch(engine, *banner)
	} else {
		fmt.Println("请使用 -banner 指定要匹配的Banner，或使用 -interactive 进入交互模式")
		flag.Usage()
	}
}

func runSingleMatch(engine *BannerEngine, banner string) {
	fmt.Printf("🔍 匹配Banner: %q\n", banner)
	
	start := time.Now()
	results := engine.Match(banner)
	duration := time.Since(start)
	
	if *output == "json" {
		outputJSON(results, duration)
	} else {
		outputText(results, duration)
	}
}

func runInteractiveMode(engine *BannerEngine) {
	fmt.Println("🎯 交互模式 - 输入 'help' 查看帮助，输入 'quit' 退出")
	fmt.Println()
	
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Print("banner> ")
		if !scanner.Scan() {
			break
		}
		
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		
		switch {
		case input == "quit" || input == "exit":
			fmt.Println("👋 再见!")
			return
		case input == "help":
			showHelp()
		case input == "stats":
			showStats(engine)
		case input == "rules":
			showRules(engine)
		case strings.HasPrefix(input, "match "):
			handleMatch(engine, input[6:])
		case strings.HasPrefix(input, "add "):
			handleAddRule(engine, input[4:])
		default:
			// 直接作为Banner匹配
			handleMatch(engine, input)
		}
		fmt.Println()
	}
}

func showHelp() {
	fmt.Println("📖 可用命令:")
	fmt.Println("  help                    - 显示帮助信息")
	fmt.Println("  stats                   - 显示引擎统计信息")
	fmt.Println("  rules                   - 显示已加载的规则")
	fmt.Println("  match <banner>          - 匹配指定的Banner")
	fmt.Println("  add <json>              - 添加新规则 (JSON格式)")
	fmt.Println("  quit/exit               - 退出程序")
	fmt.Println()
	fmt.Println("📝 示例:")
	fmt.Println("  SSH-2.0-OpenSSH_8.2p1")
	fmt.Println("  match nginx/1.18.0")
	fmt.Println(`  add {"service":"test","pattern":"test.*","product":"Test"}`)
}

func showStats(engine *BannerEngine) {
	engineStats := engine.GetStats()
	
	fmt.Println("📊 引擎统计信息:")
	fmt.Printf("  总规则数: %d\n", engineStats.TotalRules)
	fmt.Printf("  总匹配次数: %d\n", engineStats.TotalMatches)
	fmt.Printf("  缓存命中: %d\n", engineStats.CacheHits)
	fmt.Printf("  缓存未命中: %d\n", engineStats.CacheMisses)
	if engineStats.CacheHits+engineStats.CacheMisses > 0 {
		hitRate := float64(engineStats.CacheHits) / float64(engineStats.CacheHits+engineStats.CacheMisses) * 100
		fmt.Printf("  缓存命中率: %.2f%%\n", hitRate)
	}
	fmt.Printf("  平均匹配时间: %v\n", engineStats.AvgMatchTime)
}

func showRules(engine *BannerEngine) {
	rules := engine.GetRules()
	fmt.Printf("📋 已加载 %d 条规则:\n", len(rules))
	
	for i, rule := range rules {
		if i >= 10 { // 只显示前10条
			fmt.Printf("  ... 还有 %d 条规则\n", len(rules)-10)
			break
		}
		fmt.Printf("  %d. [%s] %s - %s\n", i+1, rule.ID, rule.Service, rule.Product)
	}
}

func handleMatch(engine *BannerEngine, banner string) {
	if banner == "" {
		fmt.Println("❌ Banner不能为空")
		return
	}
	
	start := time.Now()
	results := engine.Match(banner)
	duration := time.Since(start)
	
	outputText(results, duration)
}

func handleAddRule(engine *BannerEngine, jsonStr string) {
	var simpleRule SimpleRule
	if err := json.Unmarshal([]byte(jsonStr), &simpleRule); err != nil {
		fmt.Printf("❌ JSON解析失败: %v\n", err)
		return
	}
	
	if err := engine.AddSimpleRule(&simpleRule); err != nil {
		fmt.Printf("❌ 添加规则失败: %v\n", err)
		return
	}
	
	fmt.Println("✅ 规则添加成功")
}

func outputText(results []*ServiceInfo, duration time.Duration) {
	if len(results) == 0 {
		fmt.Printf("❌ 未匹配到任何服务 (耗时: %v)\n", duration)
		return
	}
	
	fmt.Printf("✅ 匹配到 %d 个服务 (耗时: %v):\n", len(results), duration)
	
	for i, result := range results {
		if *confidence > 0 && result.Confidence < *confidence {
			continue
		}
		
		fmt.Printf("\n%d. %s", i+1, result.Name)
		if result.Product != "" && result.Product != result.Name {
			fmt.Printf(" (%s)", result.Product)
		}
		if result.Version != "" {
			fmt.Printf(" v%s", result.Version)
		}
		fmt.Printf(" - 置信度: %d%%\n", result.Confidence)
		
		if result.Info != "" {
			fmt.Printf("   信息: %s\n", result.Info)
		}
		if result.OS != "" {
			fmt.Printf("   操作系统: %s\n", result.OS)
		}
		fmt.Printf("   规则ID: %s\n", result.RuleID)
		fmt.Printf("   匹配文本: %q\n", result.MatchedText)
	}
}

func outputJSON(results []*ServiceInfo, duration time.Duration) {
	output := map[string]interface{}{
		"results":     results,
		"count":       len(results),
		"duration_ms": duration.Milliseconds(),
		"timestamp":   time.Now().Format(time.RFC3339),
	}
	
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Printf("❌ JSON序列化失败: %v\n", err)
		return
	}
	
	fmt.Println(string(jsonData))
}