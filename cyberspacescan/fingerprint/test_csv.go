// +build ignore

package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// 简化版指纹识别
type Fingerprint struct {
	Product    string
	Version    string
	Category   string
	Confidence int
}

var rules = []struct {
	Name       string
	Category   string
	Pattern    *regexp.Regexp
	Version    *regexp.Regexp
	Confidence int
}{
	{
		Name:       "Nginx",
		Category:   "Web服务器",
		Pattern:    regexp.MustCompile(`(?i)nginx`),
		Version:    regexp.MustCompile(`nginx[/\s]+(\d+\.\d+\.\d+)`),
		Confidence: 95,
	},
	{
		Name:       "Apache",
		Category:   "Web服务器",
		Pattern:    regexp.MustCompile(`(?i)apache`),
		Version:    regexp.MustCompile(`Apache[/\s]+(\d+\.\d+\.\d+)`),
		Confidence: 95,
	},
	{
		Name:       "GHost",
		Category:   "Web服务器",
		Pattern:    regexp.MustCompile(`(?i)GHost`),
		Confidence: 90,
	},
	{
		Name:       "IIS",
		Category:   "Web服务器",
		Pattern:    regexp.MustCompile(`(?i)Microsoft-IIS`),
		Version:    regexp.MustCompile(`Microsoft-IIS[/\s]+(\d+\.\d+)`),
		Confidence: 95,
	},
}

func identify(banner string) *Fingerprint {
	for _, rule := range rules {
		if rule.Pattern.MatchString(banner) {
			fp := &Fingerprint{
				Product:    rule.Name,
				Category:   rule.Category,
				Confidence: rule.Confidence,
			}
			if rule.Version != nil {
				if matches := rule.Version.FindStringSubmatch(banner); len(matches) > 1 {
					fp.Version = matches[1]
				}
			}
			return fp
		}
	}
	return nil
}

type CSVRecord struct {
	IP         string
	Port       string
	Protocol   string
	State      string
	Service    string
	Banner     string
	Product    string
	Version    string
	Category   string
	Confidence string
}

func main() {
	// 使用扫描结果目录中的CSV
	inputFile := "/Users/jatsmith/CodeBuddy/cskg/cyberspacescan/results/scan_result_20260107_185446.csv"
	outputFile := "/Users/jatsmith/CodeBuddy/cskg/cyberspacescan/results/scan_result_20260107_185446_fingerprint.csv"
	
	fmt.Println("🔍 CSV指纹识别测试")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println()
	
	// 检查输入文件
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		fmt.Printf("❌ 输入文件不存在: %s\n", inputFile)
		fmt.Println("提示: 请先运行扫描器生成CSV结果文件")
		return
	}
	
	fmt.Printf("📖 读取文件: %s\n", inputFile)
	
	// 读取CSV
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("❌ 打开文件失败: %v\n", err)
		return
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("❌ 读取CSV失败: %v\n", err)
		return
	}
	
	if len(rows) < 2 {
		fmt.Println("❌ CSV文件为空")
		return
	}
	
	fmt.Printf("📊 共读取 %d 条记录\n\n", len(rows)-1)
	
	// 解析记录
	var records []*CSVRecord
	header := rows[0]
	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[col] = i
	}
	
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue
		}
		
		record := &CSVRecord{}
		if idx, ok := colIndex["IP地址"]; ok && idx < len(row) {
			record.IP = row[idx]
		}
		if idx, ok := colIndex["端口"]; ok && idx < len(row) {
			record.Port = row[idx]
		}
		if idx, ok := colIndex["协议"]; ok && idx < len(row) {
			record.Protocol = row[idx]
		}
		if idx, ok := colIndex["状态"]; ok && idx < len(row) {
			record.State = row[idx]
		}
		if idx, ok := colIndex["服务"]; ok && idx < len(row) {
			record.Service = row[idx]
		}
		if idx, ok := colIndex["Banner"]; ok && idx < len(row) {
			record.Banner = row[idx]
		}
		
		records = append(records, record)
	}
	
	// 处理识别
	identified := 0
	for i, record := range records {
		fmt.Printf("[%d/%d] 处理 %s:%s ... ", i+1, len(records), record.IP, record.Port)
		
		if record.Banner == "" {
			fmt.Println("跳过（无Banner）")
			continue
		}
		
		fp := identify(record.Banner)
		if fp != nil {
			record.Product = fp.Product
			record.Version = fp.Version
			record.Category = fp.Category
			record.Confidence = fmt.Sprintf("%d%%", fp.Confidence)
			identified++
			
			fmt.Printf("✅ %s", fp.Product)
			if fp.Version != "" {
				fmt.Printf(" %s", fp.Version)
			}
			fmt.Println()
		} else {
			fmt.Println("❌ 未识别")
		}
	}
	
	// 写入新CSV
	fmt.Printf("\n💾 写入文件: %s\n", outputFile)
	
	// 确保目录存在
	os.MkdirAll(filepath.Dir(outputFile), 0755)
	
	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("❌ 创建文件失败: %v\n", err)
		return
	}
	defer outFile.Close()
	
	// 写入BOM
	outFile.WriteString("\xEF\xBB\xBF")
	
	writer := csv.NewWriter(outFile)
	defer writer.Flush()
	
	// 写入标题
	newHeader := []string{
		"IP地址", "端口", "协议", "状态", "服务", "Banner",
		"产品", "版本", "类别", "置信度",
	}
	writer.Write(newHeader)
	
	// 写入数据
	for _, record := range records {
		banner := record.Banner
		if len(banner) > 100 {
			banner = banner[:100] + "..."
		}
		banner = strings.ReplaceAll(banner, "\r\n", " ")
		banner = strings.ReplaceAll(banner, "\n", " ")
		
		row := []string{
			record.IP,
			record.Port,
			record.Protocol,
			record.State,
			record.Service,
			banner,
			record.Product,
			record.Version,
			record.Category,
			record.Confidence,
		}
		writer.Write(row)
	}
	
	// 统计
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 处理完成统计")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("总记录数: %d\n", len(records))
	fmt.Printf("成功识别: %d\n", identified)
	if len(records) > 0 {
		fmt.Printf("识别率: %.1f%%\n", float64(identified)*100/float64(len(records)))
	}
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\n✅ 全部完成！")
}
