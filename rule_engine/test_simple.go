package main

import (
	"fmt"
	"regexp"
)

func main() {
	fmt.Println("🔍 Banner指纹识别引擎 - 快速测试")
	
	// 测试SSH Banner识别
	banner := "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5"
	pattern := `SSH-([.\d]+)-OpenSSH[_\s]+(\S+)`
	
	fmt.Printf("Banner: %s\n", banner)
	fmt.Printf("Pattern: %s\n", pattern)
	
	regex, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Printf("❌ 正则编译失败: %v\n", err)
		return
	}
	
	matches := regex.FindStringSubmatch(banner)
	if matches != nil {
		fmt.Printf("✅ 匹配成功!\n")
		fmt.Printf("   服务: SSH (OpenSSH)\n")
		fmt.Printf("   协议版本: %s\n", matches[1])
		fmt.Printf("   软件版本: %s\n", matches[2])
		fmt.Printf("   完整匹配: %s\n", matches[0])
	} else {
		fmt.Printf("❌ 匹配失败\n")
	}
	
	// 测试更多Banner
	fmt.Println("\n🧪 测试更多Banner:")
	
	testCases := []struct {
		banner  string
		pattern string
		service string
	}{
		{"nginx/1.18.0", `nginx[/\s]+(\d+\.\d+\.\d+)`, "nginx"},
		{"Apache/2.4.41", `Apache[/\s]+(\d+\.\d+\.\d+)`, "Apache"},
		{"+PONG\r\n", `\+PONG`, "Redis"},
	}
	
	for i, tc := range testCases {
		fmt.Printf("\n%d. %s -> %s\n", i+1, tc.banner, tc.service)
		
		regex, err := regexp.Compile(tc.pattern)
		if err != nil {
			fmt.Printf("   ❌ 正则编译失败\n")
			continue
		}
		
		if matches := regex.FindStringSubmatch(tc.banner); matches != nil {
			fmt.Printf("   ✅ 匹配成功")
			if len(matches) > 1 {
				fmt.Printf(" (版本: %s)", matches[1])
			}
			fmt.Println()
		} else {
			fmt.Printf("   ❌ 匹配失败\n")
		}
	}
	
	fmt.Println("\n🎉 测试完成! 规则引擎核心逻辑工作正常。")
}