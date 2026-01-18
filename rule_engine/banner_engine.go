package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// 简化的类型定义
type ServiceInfo struct {
	Name       string `json:"name"`
	Product    string `json:"product"`
	Version    string `json:"version"`
	Confidence int    `json:"confidence"`
	RuleID     string `json:"rule_id"`
	Matched    string `json:"matched"`
}

type Rule struct {
	ID         string `json:"id"`
	Service    string `json:"service"`
	Pattern    string `json:"pattern"`
	Product    string `json:"product"`
	Version    string `json:"version"`
	Confidence int    `json:"confidence"`
	regex      *regexp.Regexp
}

type Engine struct {
	rules []Rule
	stats struct {
		totalMatches int64
		totalRules   int
	}
}

func NewEngine() *Engine {
	return &Engine{
		rules: make([]Rule, 0),
	}
}

func (e *Engine) LoadBuiltinRules() {
	builtinRules := []Rule{
		{
			ID:         "ssh_openssh",
			Service:    "ssh",
			Pattern:    `SSH-([.\d]+)-OpenSSH[_\s]+(\S+)`,
			Product:    "OpenSSH",
			Version:    "$2",
			Confidence: 95,
		},
		{
			ID:         "http_nginx",
			Service:    "http",
			Pattern:    `(?i)nginx[/\s]+(\d+\.\d+\.\d+)`,
			Product:    "nginx",
			Version:    "$1",
			Confidence: 90,
		},
		{
			ID:         "http_apache",
			Service:    "http",
			Pattern:    `(?i)Apache[/\s]+(\d+\.\d+\.\d+)`,
			Product:    "Apache httpd",
			Version:    "$1",
			Confidence: 90,
		},
		{
			ID:         "mysql",
			Service:    "mysql",
			Pattern:    `(\d+\.\d+\.\d+).*mysql`,
			Product:    "MySQL",
			Version:    "$1",
			Confidence: 90,
		},
		{
			ID:         "redis",
			Service:    "redis",
			Pattern:    `\+PONG`,
			Product:    "Redis",
			Confidence: 95,
		},
		{
			ID:         "ftp_vsftpd",
			Service:    "ftp",
			Pattern:    `220.*vsftpd\s+(\S+)`,
			Product:    "vsftpd",
			Version:    "$1",
			Confidence: 95,
		},
		{
			ID:         "smtp_postfix",
			Service:    "smtp",
			Pattern:    `220.*Postfix`,
			Product:    "Postfix",
			Confidence: 85,
		},
		{
			ID:         "http_iis",
			Service:    "http",
			Pattern:    `Microsoft-IIS[/\s]+(\d+\.\d+)`,
			Product:    "Microsoft IIS",
			Version:    "$1",
			Confidence: 90,
		},
	}
	
	for _, rule := range builtinRules {
		e.AddRule(rule)
	}
}

func (e *Engine) AddRule(rule Rule) error {
	regex, err := regexp.Compile(rule.Pattern)
	if err != nil {
		return fmt.Errorf("正则表达式编译失败: %v", err)
	}
	
	rule.regex = regex
	e.rules = append(e.rules, rule)
	e.stats.totalRules = len(e.rules)
	
	return nil
}

func (e *Engine) Match(banner string) []ServiceInfo {
	var results []ServiceInfo
	e.stats.totalMatches++
	
	for _, rule := range e.rules {
		if matches := rule.regex.FindStringSubmatch(banner); matches != nil {
			service := ServiceInfo{
				Name:       rule.Service,
				Product:    rule.Product,
				Confidence: rule.Confidence,
				RuleID:     rule.ID,
				Matched:    matches[0],
			}
			
			// 提取版本信息
			if rule.Version != "" && len(matches) > 1 {
				version := rule.Version
				for i := 1; i < len(matches); i++ {
					placeholder := fmt.Sprintf("$%d", i)
					version = strings.ReplaceAll(version, placeholder, matches[i])
				}
				service.Version = version
			}
			
			results = append(results, service)
		}
	}
	
	// 按置信度排序
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Confidence < results[j].Confidence {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	
	return results
}

func (e *Engine) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_rules":   e.stats.totalRules,
		"total_matches": e.stats.totalMatches,
	}
}

// 命令行参数
var (
	banner      = flag.String("banner", "", "要匹配的Banner字符串")
	interactive = flag.Bool("interactive", false, "交互模式")
	output      = flag.String("output", "text", "输出格式 (text/json)")
)

func main() {
	flag.Parse()
	
	fmt.Println("🔍 Banner指纹识别引擎")
	fmt.Println("=" + strings.Repeat("=", 30))
	
	// 创建引擎并加载规则
	engine := NewEngine()
	engine.LoadBuiltinRules()
	
	stats := engine.GetStats()
	fmt.Printf("📚 已加载 %d 条内置规则\n\n", stats["total_rules"])
	
	if *interactive {
		runInteractive(engine)
	} else if *banner != "" {
		runSingle(engine, *banner)
	} else {
		fmt.Println("使用方法:")
		fmt.Println("  go run banner_engine.go -banner \"SSH-2.0-OpenSSH_8.2p1\"")
		fmt.Println("  go run banner_engine.go -interactive")
		fmt.Println()
		flag.Usage()
	}
}

func runSingle(engine *Engine, banner string) {
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

func runInteractive(engine *Engine) {
	fmt.Println("🎯 交互模式 - 输入Banner进行匹配，输入 'quit' 退出")
	fmt.Println()
	
	for {
		fmt.Print("banner> ")
		var input string
		fmt.Scanln(&input)
		
		if input == "quit" || input == "exit" {
			fmt.Println("👋 再见!")
			break
		}
		
		if input == "help" {
			fmt.Println("📖 可用命令:")
			fmt.Println("  直接输入Banner进行匹配")
			fmt.Println("  help - 显示帮助")
			fmt.Println("  stats - 显示统计信息")
			fmt.Println("  quit/exit - 退出")
			continue
		}
		
		if input == "stats" {
			stats := engine.GetStats()
			fmt.Printf("📊 统计信息: 规则数=%d, 匹配次数=%d\n", 
				stats["total_rules"], stats["total_matches"])
			continue
		}
		
		if input == "" {
			continue
		}
		
		start := time.Now()
		results := engine.Match(input)
		duration := time.Since(start)
		
		outputText(results, duration)
		fmt.Println()
	}
}

func outputText(results []ServiceInfo, duration time.Duration) {
	if len(results) == 0 {
		fmt.Printf("❌ 未匹配到任何服务 (耗时: %v)\n", duration)
		return
	}
	
	fmt.Printf("✅ 匹配到 %d 个服务 (耗时: %v):\n", len(results), duration)
	
	for i, result := range results {
		fmt.Printf("\n%d. %s", i+1, result.Name)
		if result.Product != "" {
			fmt.Printf(" (%s)", result.Product)
		}
		if result.Version != "" {
			fmt.Printf(" v%s", result.Version)
		}
		fmt.Printf(" - 置信度: %d%%\n", result.Confidence)
		fmt.Printf("   规则ID: %s\n", result.RuleID)
		fmt.Printf("   匹配文本: %q\n", result.Matched)
	}
}

func outputJSON(results []ServiceInfo, duration time.Duration) {
	output := map[string]interface{}{
		"results":     results,
		"count":       len(results),
		"duration_ms": duration.Milliseconds(),
		"timestamp":   time.Now().Format(time.RFC3339),
	}
	
	jsonData, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(jsonData))
}