package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cskg/CyberStroll/internal/scanner"
)

// BatchScanner 批量扫描器
type BatchScanner struct {
	engine   *scanner.EnhancedProbeEngine
	logger   *log.Logger
	results  []*ScanResult
	mutex    sync.Mutex
}

// ScanResult 扫描结果
type ScanResult struct {
	IP           string
	Status       string
	OpenPorts    []PortResult
	ScanTime     time.Duration
	ErrorMessage string
}

// PortResult 端口结果
type PortResult struct {
	Port    int
	Service string
	Version string
	Banner  string
}

// NewBatchScanner 创建批量扫描器
func NewBatchScanner() *BatchScanner {
	logger := log.New(os.Stdout, "[BatchScanner] ", log.LstdFlags)
	
	config := &scanner.ScannerConfig{
		MaxConcurrency: 20,
		Timeout:        5 * time.Second,
		RetryCount:     2,
		ProbeDelay:     100 * time.Millisecond,
		EnableLogging:  false,
	}
	
	engine := scanner.NewEnhancedProbeEngine(config)
	
	return &BatchScanner{
		engine:  engine,
		logger:  logger,
		results: make([]*ScanResult, 0),
	}
}

// LoadTargetsFromFile 从文件加载目标IP
func (bs *BatchScanner) LoadTargetsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	var ips []string
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		ip := strings.TrimSpace(scanner.Text())
		if ip != "" && !strings.HasPrefix(ip, "#") {
			ips = append(ips, ip)
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}
	
	return ips, nil
}

// ScanTargets 批量扫描目标
func (bs *BatchScanner) ScanTargets(ips []string, ports []int) {
	fmt.Printf("🎯 开始批量扫描 %d 个目标...\n", len(ips))
	fmt.Printf("📋 扫描端口: %v\n", ports)
	fmt.Printf("⚙️  并发数: %d, 超时: %v\n", 20, 5*time.Second)
	fmt.Println()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // 限制并发数

	startTime := time.Now()
	
	for i, ip := range ips {
		wg.Add(1)
		go func(index int, targetIP string) {
			defer wg.Done()
			
			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			bs.scanSingleTarget(index+1, targetIP, ports)
		}(i, ip)
		
		// 添加小延迟避免过于激进
		time.Sleep(50 * time.Millisecond)
	}
	
	wg.Wait()
	totalDuration := time.Since(startTime)
	
	fmt.Printf("\n🎉 批量扫描完成! 总耗时: %v\n", totalDuration)
	bs.generateReport()
}

// scanSingleTarget 扫描单个目标
func (bs *BatchScanner) scanSingleTarget(index int, ip string, ports []int) {
	fmt.Printf("[%d] 🔍 扫描 %s", index, ip)
	
	startTime := time.Now()
	
	// 创建扫描任务
	task := &scanner.ScanTask{
		TaskID:   fmt.Sprintf("batch-scan-%d", index),
		IP:       ip,
		TaskType: "port_scan_specified",
		Config: scanner.ScanConfig{
			Ports:      ports,
			Timeout:    5,
			ScanDepth:  "basic",
			EnableApps: false,
		},
		Timestamp: time.Now().Unix(),
	}
	
	// 执行扫描
	result, err := bs.engine.ScanTarget(task)
	duration := time.Since(startTime)
	
	// 处理结果
	scanResult := &ScanResult{
		IP:       ip,
		ScanTime: duration,
	}
	
	if err != nil {
		scanResult.Status = "failed"
		scanResult.ErrorMessage = err.Error()
		fmt.Printf(" ❌ 失败 (%v) - %v\n", duration, err)
	} else if result.ScanStatus == "success" {
		scanResult.Status = "success"
		
		// 转换端口结果
		for _, port := range result.Results.OpenPorts {
			scanResult.OpenPorts = append(scanResult.OpenPorts, PortResult{
				Port:    port.Port,
				Service: port.Service,
				Version: port.Version,
				Banner:  port.Banner,
			})
		}
		
		if len(scanResult.OpenPorts) > 0 {
			fmt.Printf(" ✅ 成功 (%v) - 发现 %d 个开放端口\n", 
				duration, len(scanResult.OpenPorts))
		} else {
			fmt.Printf(" 🔒 成功 (%v) - 无开放端口\n", duration)
		}
	} else {
		scanResult.Status = "failed"
		scanResult.ErrorMessage = "扫描状态异常"
		fmt.Printf(" ⚠️  异常 (%v) - 状态: %s\n", duration, result.ScanStatus)
	}
	
	// 保存结果
	bs.mutex.Lock()
	bs.results = append(bs.results, scanResult)
	bs.mutex.Unlock()
}

// generateReport 生成扫描报告
func (bs *BatchScanner) generateReport() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 批量扫描报告")
	fmt.Println(strings.Repeat("=", 60))
	
	// 统计信息
	totalTargets := len(bs.results)
	successCount := 0
	failedCount := 0
	totalOpenPorts := 0
	totalScanTime := time.Duration(0)
	
	for _, result := range bs.results {
		if result.Status == "success" {
			successCount++
			totalOpenPorts += len(result.OpenPorts)
		} else {
			failedCount++
		}
		totalScanTime += result.ScanTime
	}
	
	fmt.Printf("📈 扫描统计:\n")
	fmt.Printf("   总目标数: %d\n", totalTargets)
	fmt.Printf("   成功扫描: %d\n", successCount)
	fmt.Printf("   失败扫描: %d\n", failedCount)
	fmt.Printf("   成功率: %.1f%%\n", float64(successCount)/float64(totalTargets)*100)
	fmt.Printf("   发现开放端口: %d\n", totalOpenPorts)
	fmt.Printf("   平均扫描时间: %v\n", totalScanTime/time.Duration(totalTargets))
	
	// 详细结果
	fmt.Printf("\n🔍 详细扫描结果:\n")
	fmt.Println(strings.Repeat("-", 60))
	
	for i, result := range bs.results {
		status := "❌"
		if result.Status == "success" {
			if len(result.OpenPorts) > 0 {
				status = "🔓"
			} else {
				status = "🔒"
			}
		}
		
		fmt.Printf("[%2d] %s %-15s (%v)", 
			i+1, status, result.IP, result.ScanTime)
		
		if result.Status == "success" && len(result.OpenPorts) > 0 {
			fmt.Printf(" - 开放端口: ")
			for j, port := range result.OpenPorts {
				if j > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%d", port.Port)
				if port.Service != "unknown" && port.Service != "" {
					fmt.Printf("(%s)", port.Service)
				}
			}
		} else if result.Status == "failed" {
			fmt.Printf(" - 错误: %s", result.ErrorMessage)
		}
		fmt.Println()
	}
	
	// 开放端口汇总
	if totalOpenPorts > 0 {
		fmt.Printf("\n🔓 开放端口汇总:\n")
		fmt.Println(strings.Repeat("-", 60))
		
		portCount := make(map[int]int)
		serviceCount := make(map[string]int)
		
		for _, result := range bs.results {
			for _, port := range result.OpenPorts {
				portCount[port.Port]++
				if port.Service != "unknown" && port.Service != "" {
					serviceCount[port.Service]++
				}
			}
		}
		
		fmt.Printf("端口分布:\n")
		for port, count := range portCount {
			fmt.Printf("   端口 %d: %d 个主机\n", port, count)
		}
		
		if len(serviceCount) > 0 {
			fmt.Printf("\n服务分布:\n")
			for service, count := range serviceCount {
				fmt.Printf("   %s: %d 个主机\n", service, count)
			}
		}
	}
	
	// 保存报告到文件
	bs.saveReportToFile()
}

// saveReportToFile 保存报告到文件
func (bs *BatchScanner) saveReportToFile() {
	filename := fmt.Sprintf("scan_report_%s.txt", time.Now().Format("20060102_150405"))
	
	file, err := os.Create(filename)
	if err != nil {
		bs.logger.Printf("创建报告文件失败: %v", err)
		return
	}
	defer file.Close()
	
	// 写入报告内容
	fmt.Fprintf(file, "CyberStroll 批量扫描报告\n")
	fmt.Fprintf(file, "生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "扫描目标数: %d\n\n", len(bs.results))
	
	for i, result := range bs.results {
		fmt.Fprintf(file, "[%d] %s - %s (%v)\n", 
			i+1, result.IP, result.Status, result.ScanTime)
		
		if len(result.OpenPorts) > 0 {
			fmt.Fprintf(file, "    开放端口: ")
			for j, port := range result.OpenPorts {
				if j > 0 {
					fmt.Fprintf(file, ", ")
				}
				fmt.Fprintf(file, "%d(%s)", port.Port, port.Service)
			}
			fmt.Fprintf(file, "\n")
		}
		
		if result.ErrorMessage != "" {
			fmt.Fprintf(file, "    错误: %s\n", result.ErrorMessage)
		}
		fmt.Fprintf(file, "\n")
	}
	
	fmt.Printf("\n💾 扫描报告已保存到: %s\n", filename)
}

func main() {
	fmt.Println("🚀 CyberStroll 批量IP扫描工具")
	fmt.Println("================================")
	
	// 创建批量扫描器
	scanner := NewBatchScanner()
	
	// 从文件加载目标IP
	ips, err := scanner.LoadTargetsFromFile("target_ips.txt")
	if err != nil {
		log.Fatalf("加载目标IP失败: %v", err)
	}
	
	fmt.Printf("📋 已加载 %d 个目标IP\n", len(ips))
	
	// 定义扫描端口 (常见服务端口)
	ports := []int{
		21,    // FTP
		22,    // SSH
		23,    // Telnet
		25,    // SMTP
		53,    // DNS
		80,    // HTTP
		110,   // POP3
		135,   // RPC
		139,   // NetBIOS
		143,   // IMAP
		443,   // HTTPS
		445,   // SMB
		993,   // IMAPS
		995,   // POP3S
		1433,  // SQL Server
		1521,  // Oracle
		3306,  // MySQL
		3389,  // RDP
		5432,  // PostgreSQL
		6379,  // Redis
		8080,  // HTTP Alt
		8443,  // HTTPS Alt
		9200,  // Elasticsearch
		27017, // MongoDB
	}
	
	// 开始批量扫描
	scanner.ScanTargets(ips, ports)
	
	fmt.Println("\n✨ 扫描完成!")
}