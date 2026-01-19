package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// runDemoAnalysis 运行演示分析
func runDemoAnalysis() {
	fmt.Println("=== SSL Certificate Analyzer Demo ===")
	
	// 创建配置
	config := &Config{
		Timeout:         5 * time.Second,
		SkipVerify:      false,
		FollowRedirects: true,
		MaxRedirects:    5,
		UserAgent:       "cert-analyzer-demo/1.0",
		Verbose:         true,
	}

	// 创建分析器
	analyzer := NewCertificateAnalyzer(config)

	// 测试URL列表
	testURLs := []string{
		"https://httpbin.org",
		"https://jsonplaceholder.typicode.com",
		"https://api.github.com",
	}

	fmt.Printf("Testing %d URLs...\n\n", len(testURLs))

	for i, url := range testURLs {
		fmt.Printf("Test %d: Analyzing %s\n", i+1, url)
		
		result := analyzer.AnalyzeURL(url)
		
		if result.Status == "success" {
			fmt.Printf("✅ Success - Certificate for: %s\n", result.Certificate.Subject.CommonName)
			fmt.Printf("   Issuer: %s\n", result.Certificate.Issuer.CommonName)
			fmt.Printf("   Valid until: %s\n", result.Certificate.Validity.NotAfter.Format("2006-01-02"))
			fmt.Printf("   Days remaining: %d\n", result.Certificate.Validity.DaysRemaining)
			fmt.Printf("   Security score: %d/100\n", result.SecurityAnalysis.SecurityScore)
			
			if len(result.SecurityAnalysis.Warnings) > 0 {
				fmt.Printf("   Warnings: %v\n", result.SecurityAnalysis.Warnings)
			}
		} else {
			fmt.Printf("❌ Failed - %s\n", result.Error)
		}
		
		fmt.Println()
	}

	// 演示批量分析
	fmt.Println("=== Batch Analysis Demo ===")
	
	batchAnalyzer := NewBatchAnalyzer(analyzer, 2)
	batchResult := batchAnalyzer.AnalyzeURLs(testURLs)
	
	fmt.Printf("Batch Results:\n")
	fmt.Printf("- Total URLs: %d\n", batchResult.TotalURLs)
	fmt.Printf("- Successful: %d\n", batchResult.SuccessCount)
	fmt.Printf("- Failed: %d\n", batchResult.FailureCount)
	
	if batchResult.Summary != nil {
		fmt.Printf("- Average security score: %.1f\n", batchResult.Summary.AverageScore)
		fmt.Printf("- Expired certificates: %d\n", batchResult.Summary.ExpiredCerts)
		fmt.Printf("- Expiring soon: %d\n", batchResult.Summary.ExpiringSoon)
	}

	// 输出完整的JSON结果示例
	if len(batchResult.Results) > 0 && batchResult.Results[0].Status == "success" {
		fmt.Println("\n=== Sample JSON Output ===")
		jsonData, err := json.MarshalIndent(batchResult.Results[0], "", "  ")
		if err != nil {
			log.Printf("JSON marshal error: %v", err)
		} else {
			fmt.Println(string(jsonData))
		}
	}
}