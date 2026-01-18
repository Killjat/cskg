package fingerprint

import (
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CSVRecord CSV记录结构
type CSVRecord struct {
	IP          string
	Port        string
	Protocol    string
	State       string
	Service     string
	Banner      string
	Response    string
	Product     string // 新增：产品名称
	Version     string // 新增：版本号
	Category    string // 新增：类别
	Vendor      string // 新增：厂商
	Confidence  string // 新增：置信度
	OS          string // 新增：操作系统
	Tags        string // 新增：标签
}

// ProcessCSV 处理CSV文件并添加指纹识别结果
func ProcessCSV(inputFile string, outputFile string) error {
	// 读取CSV文件
	records, err := readCSV(inputFile)
	if err != nil {
		return fmt.Errorf("读取CSV失败: %v", err)
	}

	fmt.Printf("📖 读取CSV文件: %s\n", inputFile)
	fmt.Printf("📊 共读取 %d 条记录\n\n", len(records))

	// 对每条记录进行指纹识别
	processed := 0
	identified := 0
	
	for i, record := range records {
		fmt.Printf("[%d/%d] 处理 IP: %s 端口: %s ... ", i+1, len(records), record.IP, record.Port)
		
		// 跳过空Banner
		if record.Banner == "" && record.Response == "" {
			fmt.Println("跳过（无Banner）")
			continue
		}
		
		processed++
		
		// 解码Response（如果是Base64编码）
		var response []byte
		if record.Response != "" {
			decoded, err := base64.StdEncoding.DecodeString(record.Response)
			if err == nil {
				response = decoded
			} else {
				response = []byte(record.Response)
			}
		}
		
		// 进行指纹识别
		fp := GetTopFingerprint(record.Banner, response)
		
		if fp != nil {
			record.Product = fp.Product
			record.Version = fp.Version
			record.Category = fp.Category
			record.Vendor = fp.Vendor
			record.Confidence = fmt.Sprintf("%d%%", fp.Confidence)
			record.OS = fp.OS
			record.Tags = strings.Join(fp.Tags, ", ")
			
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

	// 写入新的CSV文件
	if err := writeCSV(outputFile, records); err != nil {
		return fmt.Errorf("写入CSV失败: %v", err)
	}

	fmt.Printf("\n✅ 处理完成！\n")
	fmt.Printf("📊 统计信息:\n")
	fmt.Printf("   总记录数: %d\n", len(records))
	fmt.Printf("   处理记录: %d\n", processed)
	fmt.Printf("   成功识别: %d\n", identified)
	fmt.Printf("   识别率: %.1f%%\n", float64(identified)*100/float64(processed))
	fmt.Printf("💾 输出文件: %s\n", outputFile)

	return nil
}

// readCSV 读取CSV文件
func readCSV(filename string) ([]*CSVRecord, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // 允许字段数量不一致
	
	// 读取所有行
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("CSV文件为空或只有标题行")
	}

	// 解析标题行
	header := rows[0]
	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[col] = i
	}

	// 解析数据行
	var records []*CSVRecord
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

	return records, nil
}

// writeCSV 写入CSV文件
func writeCSV(filename string, records []*CSVRecord) error {
	// 创建输出目录
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入BOM以支持Excel正确显示中文
	file.WriteString("\xEF\xBB\xBF")

	// 写入标题行（包含新增的指纹识别字段）
	header := []string{
		"IP地址", "端口", "协议", "状态", "服务", "Banner",
		"产品", "版本", "类别", "厂商", "置信度", "操作系统", "标签",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// 写入数据行
	for _, record := range records {
		row := []string{
			record.IP,
			record.Port,
			record.Protocol,
			record.State,
			record.Service,
			truncateString(record.Banner, 100), // 截断过长的Banner
			record.Product,
			record.Version,
			record.Category,
			record.Vendor,
			record.Confidence,
			record.OS,
			record.Tags,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	// 移除换行符
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ProcessCSVWithStats 处理CSV并返回统计信息
func ProcessCSVWithStats(inputFile string, outputFile string) (*ProcessStats, error) {
	stats := &ProcessStats{
		StartTime: time.Now(),
	}

	// 读取CSV
	records, err := readCSV(inputFile)
	if err != nil {
		return nil, err
	}

	stats.TotalRecords = len(records)
	
	// 统计各类别
	categoryCount := make(map[string]int)
	productCount := make(map[string]int)

	// 处理每条记录
	for _, record := range records {
		if record.Banner == "" && record.Response == "" {
			continue
		}
		
		stats.ProcessedRecords++
		
		// 识别指纹
		var response []byte
		if record.Response != "" {
			decoded, _ := base64.StdEncoding.DecodeString(record.Response)
			response = decoded
		}
		
		fp := GetTopFingerprint(record.Banner, response)
		
		if fp != nil {
			stats.IdentifiedRecords++
			
			record.Product = fp.Product
			record.Version = fp.Version
			record.Category = fp.Category
			record.Vendor = fp.Vendor
			record.Confidence = fmt.Sprintf("%d%%", fp.Confidence)
			record.OS = fp.OS
			record.Tags = strings.Join(fp.Tags, ", ")
			
			// 统计
			categoryCount[fp.Category]++
			productKey := fp.Product
			if fp.Version != "" {
				productKey += " " + fp.Version
			}
			productCount[productKey]++
		}
	}

	// 写入结果
	if err := writeCSV(outputFile, records); err != nil {
		return nil, err
	}

	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime)
	stats.CategoryStats = categoryCount
	stats.ProductStats = productCount
	stats.OutputFile = outputFile

	return stats, nil
}

// ProcessStats 处理统计信息
type ProcessStats struct {
	TotalRecords      int
	ProcessedRecords  int
	IdentifiedRecords int
	CategoryStats     map[string]int
	ProductStats      map[string]int
	StartTime         time.Time
	EndTime           time.Time
	Duration          time.Duration
	OutputFile        string
}

// PrintStats 打印统计信息
func (s *ProcessStats) PrintStats() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 指纹识别统计报告")
	fmt.Println(strings.Repeat("=", 60))
	
	fmt.Printf("\n⏱️  处理时间: %v\n", s.Duration.Round(time.Millisecond))
	
	fmt.Println("\n📈 记录统计:")
	fmt.Printf("   总记录数: %d\n", s.TotalRecords)
	fmt.Printf("   处理记录: %d\n", s.ProcessedRecords)
	fmt.Printf("   成功识别: %d\n", s.IdentifiedRecords)
	if s.ProcessedRecords > 0 {
		fmt.Printf("   识别率: %.1f%%\n", float64(s.IdentifiedRecords)*100/float64(s.ProcessedRecords))
	}
	
	if len(s.CategoryStats) > 0 {
		fmt.Println("\n📦 类别分布:")
		for category, count := range s.CategoryStats {
			fmt.Printf("   %-15s : %d\n", category, count)
		}
	}
	
	if len(s.ProductStats) > 0 {
		fmt.Println("\n🔧 产品分布 (Top 10):")
		// 排序并显示前10个
		type kv struct {
			Key   string
			Value int
		}
		var sorted []kv
		for k, v := range s.ProductStats {
			sorted = append(sorted, kv{k, v})
		}
		// 简单冒泡排序
		for i := 0; i < len(sorted); i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[j].Value > sorted[i].Value {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
		
		limit := 10
		if len(sorted) < limit {
			limit = len(sorted)
		}
		for i := 0; i < limit; i++ {
			fmt.Printf("   %-30s : %d\n", sorted[i].Key, sorted[i].Value)
		}
	}
	
	fmt.Printf("\n💾 输出文件: %s\n", s.OutputFile)
	fmt.Println(strings.Repeat("=", 60))
}

// BatchProcessCSV 批量处理多个CSV文件
func BatchProcessCSV(inputDir string, outputDir string) error {
	// 查找所有CSV文件
	files, err := filepath.Glob(filepath.Join(inputDir, "*.csv"))
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("未找到CSV文件")
	}

	fmt.Printf("🔍 找到 %d 个CSV文件\n\n", len(files))

	// 处理每个文件
	for i, file := range files {
		fmt.Printf("[%d/%d] 处理文件: %s\n", i+1, len(files), filepath.Base(file))
		
		// 生成输出文件名
		baseName := filepath.Base(file)
		outputFile := filepath.Join(outputDir, "fingerprint_"+baseName)
		
		// 处理文件
		if err := ProcessCSV(file, outputFile); err != nil {
			fmt.Printf("❌ 处理失败: %v\n\n", err)
			continue
		}
		
		fmt.Println()
	}

	return nil
}
