// +build ignore

package main

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

// 简化版指纹结构
type Fingerprint struct {
	Product    string
	Version    string
	Category   string
	OS         string
	Vendor     string
	Tags       []string
	Confidence int
	CPE        string
}

type Rule struct {
	Name       string
	Category   string
	Vendor     string
	Pattern    *regexp.Regexp
	Version    *regexp.Regexp
	Confidence int
	Tags       []string
}

var rules = []*Rule{
	{
		Name:       "Nginx",
		Category:   "Web服务器",
		Vendor:     "Nginx Inc.",
		Pattern:    regexp.MustCompile(`(?i)nginx`),
		Version:    regexp.MustCompile(`nginx[/\s]+(\d+\.\d+\.\d+)`),
		Confidence: 95,
		Tags:       []string{"web", "http"},
	},
	{
		Name:       "Apache",
		Category:   "Web服务器",
		Vendor:     "Apache Software Foundation",
		Pattern:    regexp.MustCompile(`(?i)apache`),
		Version:    regexp.MustCompile(`Apache[/\s]+(\d+\.\d+\.\d+)`),
		Confidence: 95,
		Tags:       []string{"web", "http"},
	},
	{
		Name:       "GHost",
		Category:   "Web服务器",
		Vendor:     "Unknown",
		Pattern:    regexp.MustCompile(`(?i)GHost`),
		Confidence: 90,
		Tags:       []string{"web"},
	},
	{
		Name:       "PHP",
		Category:   "编程语言",
		Vendor:     "PHP Group",
		Pattern:    regexp.MustCompile(`(?i)PHP[/\s]+|X-Powered-By.*PHP`),
		Version:    regexp.MustCompile(`PHP[/\s]+(\d+\.\d+\.\d+)`),
		Confidence: 90,
		Tags:       []string{"php"},
	},
	{
		Name:       "OpenSSH",
		Category:   "SSH服务",
		Vendor:     "OpenBSD",
		Pattern:    regexp.MustCompile(`(?i)SSH-.*OpenSSH`),
		Version:    regexp.MustCompile(`OpenSSH[_\s]+(\d+\.\d+)`),
		Confidence: 95,
		Tags:       []string{"ssh"},
	},
}

func identify(banner string, response []byte) []*Fingerprint {
	var results []*Fingerprint
	
	var decodedResponse string
	if response != nil && len(response) > 0 {
		decoded, err := base64.StdEncoding.DecodeString(string(response))
		if err == nil {
			decodedResponse = string(decoded)
		}
	}
	
	fullContent := banner
	if decodedResponse != "" {
		fullContent = banner + "\n" + decodedResponse
	}
	
	if fullContent == "" {
		return results
	}
	
	for _, rule := range rules {
		if rule.Pattern.MatchString(fullContent) {
			fp := &Fingerprint{
				Product:    rule.Name,
				Category:   rule.Category,
				Vendor:     rule.Vendor,
				Tags:       rule.Tags,
				Confidence: rule.Confidence,
			}
			
			if rule.Version != nil {
				if matches := rule.Version.FindStringSubmatch(fullContent); len(matches) > 1 {
					fp.Version = matches[1]
				}
			}
			
			fp.OS = inferOS(fullContent)
			fp.CPE = generateCPE(fp)
			
			results = append(results, fp)
		}
	}
	
	return results
}

func inferOS(content string) string {
	lowerContent := strings.ToLower(content)
	
	if strings.Contains(lowerContent, "ubuntu") {
		return "Linux/Ubuntu"
	}
	if strings.Contains(lowerContent, "centos") {
		return "Linux/CentOS"
	}
	if strings.Contains(lowerContent, "debian") {
		return "Linux/Debian"
	}
	
	return "Unknown"
}

func generateCPE(fp *Fingerprint) string {
	if fp.Product == "" {
		return ""
	}
	
	vendor := strings.ToLower(strings.ReplaceAll(fp.Vendor, " ", "_"))
	product := strings.ToLower(strings.ReplaceAll(fp.Product, " ", "_"))
	version := fp.Version
	
	if vendor == "" {
		vendor = "*"
	}
	if version == "" {
		version = "*"
	}
	
	return "cpe:/a:" + vendor + ":" + product + ":" + version
}

func main() {
	fmt.Println("=== 🔍 指纹识别演示 ===\n")
	
	// 示例1
	fmt.Println("📌 示例1: Nginx服务器")
	banner1 := "HTTP/1.1 403 Forbidden\r\nServer: nginx\r\n"
	fps1 := identify(banner1, nil)
	printFingerprints(fps1)
	
	// 示例2
	fmt.Println("\n📌 示例2: Apache + PHP")
	banner2 := "HTTP/1.1 200 OK\r\nServer: Apache/2.4.41 (Ubuntu)\r\nX-Powered-By: PHP/7.4.3\r\n"
	fps2 := identify(banner2, nil)
	printFingerprints(fps2)
	
	// 示例3
	fmt.Println("\n📌 示例3: GHost服务器（台湾网站）")
	banner3 := "HTTP/1.0 400 Bad Request\r\nServer: GHost\r\n"
	fps3 := identify(banner3, nil)
	printFingerprints(fps3)
	
	// 示例4
	fmt.Println("\n📌 示例4: OpenSSH")
	banner4 := "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.3"
	fps4 := identify(banner4, nil)
	printFingerprints(fps4)
	
	fmt.Println("\n✅ 指纹识别演示完成！")
}

func printFingerprints(fps []*Fingerprint) {
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
