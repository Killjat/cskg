package main

import (
	"encoding/base64"
	"fmt"
	
	"cskg/cyberspacescan/fingerprint"
)

func main() {
	fmt.Println("=== 指纹识别演示 ===\n")
	
	// 示例1: 识别Nginx
	fmt.Println("📌 示例1: Nginx服务器")
	banner1 := "HTTP/1.1 403 Forbidden\r\nServer: nginx\r\nDate: Wed, 07 Jan 2026 10:54:41 GMT\r\n"
	fps1 := fingerprint.Identify(banner1, nil)
	printFingerprints(fps1)
	
	// 示例2: 识别Apache + PHP
	fmt.Println("\n📌 示例2: Apache + PHP")
	banner2 := "HTTP/1.1 200 OK\r\nServer: Apache/2.4.41 (Ubuntu)\r\nX-Powered-By: PHP/7.4.3\r\n"
	fps2 := fingerprint.Identify(banner2, nil)
	printFingerprints(fps2)
	
	// 示例3: 识别GHost（台湾网站常用）
	fmt.Println("\n📌 示例3: GHost服务器")
	banner3 := "HTTP/1.0 400 Bad Request\r\nServer: GHost\r\nMime-Version: 1.0\r\n"
	fps3 := fingerprint.Identify(banner3, nil)
	printFingerprints(fps3)
	
	// 示例4: 识别包含WordPress的响应
	fmt.Println("\n📌 示例4: WordPress网站")
	banner4 := "HTTP/1.1 200 OK\r\nServer: nginx/1.18.0\r\n"
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <meta name="generator" content="WordPress 5.8">
    <link rel="stylesheet" href="/wp-content/themes/twentytwenty/style.css">
</head>
<body></body>
</html>`
	response4 := base64.StdEncoding.EncodeToString([]byte(htmlContent))
	fps4 := fingerprint.Identify(banner4, []byte(response4))
	printFingerprints(fps4)
	
	// 示例5: 识别SSH
	fmt.Println("\n📌 示例5: OpenSSH服务")
	banner5 := "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.3"
	fps5 := fingerprint.Identify(banner5, nil)
	printFingerprints(fps5)
	
	// 示例6: 识别Redis
	fmt.Println("\n📌 示例6: Redis数据库")
	banner6 := "$5\r\nredis_version:6.2.5\r\n# Server\r\n"
	fps6 := fingerprint.Identify(banner6, nil)
	printFingerprints(fps6)
	
	// 示例7: 获取最高置信度指纹
	fmt.Println("\n📌 示例7: 获取最高置信度指纹")
	banner7 := "HTTP/1.1 200 OK\r\nServer: Apache/2.4.41\r\nX-Powered-By: PHP/7.4\r\n"
	top := fingerprint.GetTopFingerprint(banner7, nil)
	if top != nil {
		fmt.Printf("最高置信度: %s (置信度: %d%%)\n", top.Product, top.Confidence)
	}
	
	// 示例8: 检查标签
	fmt.Println("\n📌 示例8: 检查是否为Web服务")
	banner8 := "HTTP/1.1 200 OK\r\nServer: nginx\r\n"
	if fingerprint.HasTag(banner8, nil, "web") {
		fmt.Println("✅ 这是一个Web服务")
	}
	if fingerprint.HasTag(banner8, nil, "database") {
		fmt.Println("✅ 这是一个数据库服务")
	} else {
		fmt.Println("❌ 这不是数据库服务")
	}
	
	// 示例9: 获取所有类别
	fmt.Println("\n📌 示例9: 获取识别到的所有类别")
	banner9 := "HTTP/1.1 200 OK\r\nServer: Apache\r\nX-Powered-By: PHP/7.4\r\n"
	categories := fingerprint.GetCategories(banner9, nil)
	fmt.Println("识别到的类别:", categories)
	
	// 示例10: 识别IIS + ASP.NET
	fmt.Println("\n📌 示例10: IIS + ASP.NET")
	banner10 := "HTTP/1.1 200 OK\r\nServer: Microsoft-IIS/10.0\r\nX-Powered-By: ASP.NET\r\nX-AspNet-Version: 4.0.30319\r\n"
	fps10 := fingerprint.Identify(banner10, nil)
	printFingerprints(fps10)
}

func printFingerprints(fps []*fingerprint.Fingerprint) {
	if len(fps) == 0 {
		fmt.Println("  ❌ 未识别出指纹")
		return
	}
	
	for i, fp := range fps {
		fmt.Printf("  [%d] 产品: %s\n", i+1, fp.Product)
		if fp.Version != "" {
			fmt.Printf("      版本: %s\n", fp.Version)
		}
		fmt.Printf("      类别: %s\n", fp.Category)
		if fp.Vendor != "" {
			fmt.Printf("      厂商: %s\n", fp.Vendor)
		}
		if fp.OS != "" && fp.OS != "Unknown" {
			fmt.Printf("      系统: %s\n", fp.OS)
		}
		fmt.Printf("      置信度: %d%%\n", fp.Confidence)
		if len(fp.Tags) > 0 {
			fmt.Printf("      标签: %v\n", fp.Tags)
		}
		if fp.CPE != "" {
			fmt.Printf("      CPE: %s\n", fp.CPE)
		}
	}
}
