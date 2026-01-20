package main

import (
	"fmt"
	"log"
)

// 这是一个简单的使用示例，展示如何在代码中使用Banner引擎

func exampleUsage() {
	fmt.Println("🔍 Banner引擎使用示例")
	fmt.Println("===================")
	
	// 1. 创建引擎
	config := DefaultConfig()
	engine := NewBannerEngine(config)
	
	// 2. 加载内置规则
	nmapLoader := NewNmapLoader(engine)
	if err := nmapLoader.LoadBuiltinRules(); err != nil {
		log.Fatalf("加载内置规则失败: %v", err)
	}
	
	fmt.Printf("✅ 已加载 %d 条规则\n\n", engine.GetStats().TotalRules)
	
	// 3. 测试各种Banner
	testBanners := []string{
		"nginx/1.18.0",
		"Apache/2.4.41 (Ubuntu)",
		"SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5",
		"5.7.34-0ubuntu0.18.04.1-log mysql_native_password",
		"+PONG\r\n",
		"220 (vsFTPd 3.0.3)",
		"220 mail.example.com ESMTP Postfix",
		"Microsoft-IIS/10.0",
	}
	
	for i, banner := range testBanners {
		fmt.Printf("%d. 测试Banner: %q\n", i+1, banner)
		
		results := engine.Match(banner)
		if len(results) > 0 {
			best := results[0] // 获取最佳匹配
			fmt.Printf("   ✅ 识别为: %s", best.Name)
			if best.Product != "" {
				fmt.Printf(" (%s)", best.Product)
			}
			if best.Version != "" {
				fmt.Printf(" v%s", best.Version)
			}
			fmt.Printf(" - 置信度: %d%%\n", best.Confidence)
		} else {
			fmt.Printf("   ❌ 未识别\n")
		}
		fmt.Println()
	}
	
	// 4. 添加自定义规则
	fmt.Println("🔧 添加自定义规则...")
	
	customRule := &SimpleRule{
		Service:     "myapp",
		Pattern:     `MyApp[/\s]+v(\d+\.\d+)`,
		Product:     "My Custom Application",
		Version:     "$1",
		Description: "My custom application detection",
		Confidence:  85,
	}
	
	if err := engine.AddSimpleRule(customRule); err != nil {
		fmt.Printf("❌ 添加规则失败: %v\n", err)
	} else {
		fmt.Println("✅ 自定义规则添加成功")
		
		// 测试自定义规则
		testBanner := "MyApp v2.1 Server"
		fmt.Printf("测试自定义Banner: %q\n", testBanner)
		
		results := engine.Match(testBanner)
		if len(results) > 0 {
			best := results[0]
			fmt.Printf("✅ 识别为: %s (%s) v%s - 置信度: %d%%\n", 
				best.Name, best.Product, best.Version, best.Confidence)
		}
	}
	
	// 5. 显示统计信息
	fmt.Println("\n📊 引擎统计:")
	stats := engine.GetStats()
	fmt.Printf("总规则数: %d\n", stats.TotalRules)
	fmt.Printf("总匹配次数: %d\n", stats.TotalMatches)
	fmt.Printf("缓存命中: %d\n", stats.CacheHits)
	fmt.Printf("缓存未命中: %d\n", stats.CacheMisses)
}

// 如果直接运行这个文件，执行示例
func init() {
	// 这个函数可以用来演示API使用
	// 在实际使用中，你会在main函数或其他地方调用这些功能
}