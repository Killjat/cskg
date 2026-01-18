package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// ScanResult 扫描结果结构体
type ScanResult struct {
	IP            string    `json:"ip"`             // IP地址
	IsAlive       bool      `json:"is_alive"`       // 是否活跃
	ScanTime      float64   `json:"scan_time"`      // 扫描耗时（毫秒）
	ResponseTime  float64   `json:"response_time"`  // 响应时间（毫秒）
	OpenPorts     []int     `json:"open_ports"`     // 开放端口列表
	ScanTimestamp time.Time `json:"scan_timestamp"` // 扫描时间
	IPSegment     string    `json:"ip_segment"`     // 所属IP段
}

// SYNScanner SYN扫描器
type SYNScanner struct {
	ConcurrentLimit int           // 并发限制
	Timeout         time.Duration // 超时时间
	Ports           []int         // 要扫描的端口列表
}

// NewSYNScanner 创建新的SYN扫描器
func NewSYNScanner(concurrentLimit int, timeout time.Duration, ports []int) *SYNScanner {
	return &SYNScanner{
		ConcurrentLimit: concurrentLimit,
		Timeout:         timeout,
		Ports:           ports,
	}
}

// ScanIPSegment 扫描整个IP段
func (s *SYNScanner) ScanIPSegment(segment *IPSegment) ([]*ScanResult, error) {
	// 解析CIDR，获取IP范围
	ip, ipnet, err := net.ParseCIDR(segment.CIDR)
	if err != nil {
		return nil, fmt.Errorf("解析CIDR失败: %v", err)
	}

	// 计算IP总数
	ones, bits := ipnet.Mask.Size()
	ipCount := 1 << (bits - ones)
	// 减去网络地址和广播地址
	if ipCount > 2 {
		ipCount -= 2
	}

	log.Printf("开始扫描IP段 %s，共 %d 个IP地址", segment.CIDR, ipCount)

	// 创建结果通道和等待组
	results := make([]*ScanResult, 0)
	resultChan := make(chan *ScanResult, s.ConcurrentLimit*2) // 缓冲区不需要太大
	var wg sync.WaitGroup

	// 并发控制
	semaphore := make(chan struct{}, s.ConcurrentLimit)

	// IP生成器协程
	var ipGenWg sync.WaitGroup
	ipGenWg.Add(1)

	// 控制是否停止IP生成
	stopGen := false

	// 启动IP生成和扫描协程
	go func() {
		defer ipGenWg.Done()

		// 创建一个临时IP用于递增
		currentIP := make(net.IP, len(ip))
		copy(currentIP, ip)

		// 跳过网络地址
		if ipCount > 2 {
			inc(currentIP)
		}

		// 生成IP并扫描
		for i := 0; i < ipCount; i++ {
			if stopGen {
				break
			}

			// 获取当前IP的字符串表示
			ipStr := currentIP.String()

			wg.Add(1)
			semaphore <- struct{}{} // 获取信号量

			// 启动扫描协程
			go func(ip string) {
				defer wg.Done()
				defer func() { <-semaphore }() // 释放信号量

				// 扫描单个IP
				result := s.scanIP(ip, segment.CIDR)
				if result != nil {
					resultChan <- result
				}
			}(ipStr)

			// 递增IP
			inc(currentIP)

			// 每生成1000个IP，打印进度
			if i > 0 && (i+1)%1000 == 0 {
				log.Printf("已扫描 %d/%d 个IP...", i+1, ipCount)
			}
		}
	}()

	// 等待所有扫描完成的协程
	go func() {
		ipGenWg.Wait()    // 先等待IP生成完成，确保不再创建新的扫描协程
		wg.Wait()         // 然后等待所有扫描协程完成
		close(resultChan) // 最后关闭结果通道
	}()

	// 收集结果
	scannedCount := 0
	aliveCount := 0
	for result := range resultChan {
		results = append(results, result)
		scannedCount++
		if result.IsAlive {
			aliveCount++
		}

		// 每扫描100个结果，打印进度
		if scannedCount%100 == 0 {
			log.Printf("已完成 %d 个IP扫描，发现 %d 个活跃IP", scannedCount, aliveCount)
		}
	}

	log.Printf("IP段 %s 扫描完成，发现 %d 个活跃IP", segment.CIDR, aliveCount)
	return results, nil
}

// scanIP 扫描单个IP地址
func (s *SYNScanner) scanIP(ip string, ipSegment string) *ScanResult {
	startTime := time.Now()

	// 尝试连接到IP的某个端口（模拟SYN扫描）
	isAlive := s.isIPAlive(ip)

	scanTime := time.Since(startTime).Seconds() * 1000 // 转换为毫秒

	result := &ScanResult{
		IP:            ip,
		IsAlive:       isAlive,
		ScanTime:      scanTime,
		ScanTimestamp: startTime,
		IPSegment:     ipSegment,
	}

	// 如果IP活跃，扫描端口
	if isAlive {
		responseTime := time.Since(startTime).Seconds() * 1000
		result.ResponseTime = responseTime
		result.OpenPorts = s.scanPorts(ip)
	}

	return result
}

// isIPAlive 检测IP是否活跃
// 注意：这里使用TCP连接模拟SYN扫描，实际SYN扫描需要raw socket权限
func (s *SYNScanner) isIPAlive(ip string) bool {
	// 尝试连接到80端口，判断IP是否活跃
	addr := net.JoinHostPort(ip, "80")
	conn, err := net.DialTimeout("tcp", addr, s.Timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// scanPorts 扫描端口
func (s *SYNScanner) scanPorts(ip string) []int {
	openPorts := make([]int, 0)

	for _, port := range s.Ports {
		addr := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
		conn, err := net.DialTimeout("tcp", addr, s.Timeout)
		if err == nil {
			openPorts = append(openPorts, port)
			conn.Close()
		}
	}

	return openPorts
}

// generateIPsFromCIDR 从CIDR生成所有IP地址
func generateIPsFromCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0)
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// 移除网络地址和广播地址
	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}

	return ips, nil
}

// inc 递增IP地址
func inc(ip net.IP) {
	// IPv4地址使用最后4个字节
	for j := len(ip) - 1; j >= len(ip)-4; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// TestSYNScanner 测试SYN扫描器
func TestSYNScanner() {
	// 创建SYN扫描器
	scanner := NewSYNScanner(100, 2*time.Second, []int{80, 443, 22, 21, 3389})

	// 只扫描前5个IP作为示例
	log.Println("=== SYN扫描示例 ===")
	log.Println("注意：实际SYN扫描需要raw socket权限，这里使用TCP连接模拟")
	log.Println("扫描IP段 1.1.1.0/29（8个IP）...")

	// 创建测试IP段
	testSegment := &IPSegment{
		CIDR:    "1.1.1.0/29", // 8个IP地址
		Country: "Taiwan",
		Region:  "Taipei",
		City:    "Taipei",
		ISP:     "Chunghwa Telecom",
		ASN:     3462,
	}

	// 扫描IP段
	results, err := scanner.ScanIPSegment(testSegment)
	if err != nil {
		log.Fatalf("扫描失败: %v", err)
	}

	// 打印结果
	log.Printf("扫描完成，发现 %d 个活跃IP", len(results))
	for _, result := range results {
		if result.IsAlive {
			fmt.Printf("IP: %s, 状态: 活跃, 响应时间: %.2fms, 开放端口: %v\n",
				result.IP, result.ResponseTime, result.OpenPorts)
		} else {
			fmt.Printf("IP: %s, 状态: 不活跃\n", result.IP)
		}
	}
}
