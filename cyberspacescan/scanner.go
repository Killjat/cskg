package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// Scanner 扫描器
type Scanner struct {
	config  *Config
	workers int
	timeout time.Duration
}

// ScanResult 扫描结果
type ScanResult struct {
	IP       string
	IsAlive  bool
	TCPPorts []PortResult
	UDPPorts []PortResult
}

// PortResult 端口扫描结果
type PortResult struct {
	Port     int
	Protocol string // "tcp" or "udp"
	State    string // "open", "closed", "filtered"
	Service  string
	Banner   string
	Response []byte
}

// NewScanner 创建扫描器
func NewScanner(config *Config) *Scanner {
	return &Scanner{
		config:  config,
		workers: config.Scan.Workers,
		timeout: time.Duration(config.Scan.Timeout) * time.Millisecond,
	}
}

// Scan 执行扫描
func (s *Scanner) Scan(targets []string) []*ScanResult {
	var results []*ScanResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 创建任务队列
	jobs := make(chan string, len(targets))
	for _, target := range targets {
		jobs <- target
	}
	close(jobs)

	// 启动工作协程
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range jobs {
				result := s.scanHost(ip)
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	return results
}

// scanHost 扫描单个主机
func (s *Scanner) scanHost(ip string) *ScanResult {
	result := &ScanResult{
		IP:       ip,
		IsAlive:  false,
		TCPPorts: []PortResult{},
		UDPPorts: []PortResult{},
	}

	// ICMP探活
	if s.isAlive(ip) {
		result.IsAlive = true
		fmt.Printf("[+] %s 存活\n", ip)

		// TCP端口扫描
		result.TCPPorts = s.scanTCPPorts(ip, s.config.Ports.TCP)

		// UDP端口扫描
		result.UDPPorts = s.scanUDPPorts(ip, s.config.Ports.UDP)
	} else {
		fmt.Printf("[-] %s 不可达\n", ip)
	}

	return result
}

// isAlive ICMP探活
func (s *Scanner) isAlive(ip string) bool {
	// 尝试TCP连接常见端口来判断存活
	// 因为Go标准库不支持直接发送ICMP包（需要root权限）
	commonPorts := []int{80, 443, 22, 21, 3389}

	for _, port := range commonPorts {
		address := fmt.Sprintf("%s:%d", ip, port)
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)
		if err == nil {
			conn.Close()
			return true
		}
	}

	return false
}

// scanTCPPorts TCP端口扫描
func (s *Scanner) scanTCPPorts(ip string, ports []int) []PortResult {
	var results []PortResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, 50) // 限制并发数

	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := s.scanTCPPort(ip, p)
			if result.State == "open" {
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
				fmt.Printf("  [TCP] %s:%d 开放 - %s\n", ip, p, result.Service)
			}
		}(port)
	}

	wg.Wait()
	return results
}

// scanTCPPort 扫描单个TCP端口
func (s *Scanner) scanTCPPort(ip string, port int) PortResult {
	result := PortResult{
		Port:     port,
		Protocol: "tcp",
		State:    "closed",
		Service:  identifyService(port, "tcp"),
	}

	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, s.timeout)
	if err != nil {
		return result
	}
	defer conn.Close()

	result.State = "open"

	// 获取Banner
	if s.config.Output.SaveResponse {
		banner, response := grabBanner(conn, port)
		result.Banner = banner
		result.Response = response
	}

	return result
}

// scanUDPPorts UDP端口扫描
func (s *Scanner) scanUDPPorts(ip string, ports []int) []PortResult {
	var results []PortResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, 50)

	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := s.scanUDPPort(ip, p)
			if result.State == "open" {
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
				fmt.Printf("  [UDP] %s:%d 开放 - %s\n", ip, p, result.Service)
			}
		}(port)
	}

	wg.Wait()
	return results
}

// scanUDPPort 扫描单个UDP端口
func (s *Scanner) scanUDPPort(ip string, port int) PortResult {
	result := PortResult{
		Port:     port,
		Protocol: "udp",
		State:    "open|filtered",
		Service:  identifyService(port, "udp"),
	}

	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("udp", address, s.timeout)
	if err != nil {
		return result
	}
	defer conn.Close()

	// 发送探测包
	probe := getUDPProbe(port)
	conn.Write([]byte(probe))

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(s.timeout))

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err == nil && n > 0 {
		result.State = "open"
		if s.config.Output.SaveResponse {
			result.Response = buffer[:n]
		}
	}

	return result
}

// grabBanner 获取服务Banner
func grabBanner(conn net.Conn, port int) (string, []byte) {
	// 发送探测请求
	probe := getProbe(port)
	if probe != "" {
		conn.Write([]byte(probe))
	}

	// 读取响应
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil || n == 0 {
		return "", nil
	}

	response := buffer[:n]
	banner := string(response)
	if len(banner) > 200 {
		banner = banner[:200]
	}

	return banner, response
}

// getProbe 获取端口探测请求
func getProbe(port int) string {
	probes := map[int]string{
		80:   "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n",
		8080: "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n",
		443:  "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n",
		21:   "",     // FTP会自动发送banner
		22:   "",     // SSH会自动发送banner
		25:   "EHLO test\r\n",
		3306: "",     // MySQL需要特殊处理
	}

	if probe, ok := probes[port]; ok {
		return probe
	}
	return ""
}

// getUDPProbe 获取UDP探测包
func getUDPProbe(port int) string {
	probes := map[int]string{
		53:  "\x00\x00\x10\x00\x00\x00\x00\x00\x00\x00\x00\x00", // DNS查询
		123: "\x1b" + string(make([]byte, 47)),                   // NTP请求
		161: "\x30\x26\x02\x01\x00",                              // SNMP请求
	}

	if probe, ok := probes[port]; ok {
		return probe
	}
	return ""
}

// identifyService 识别服务
func identifyService(port int, protocol string) string {
	services := map[int]string{
		21:    "FTP",
		22:    "SSH",
		23:    "Telnet",
		25:    "SMTP",
		53:    "DNS",
		80:    "HTTP",
		110:   "POP3",
		143:   "IMAP",
		443:   "HTTPS",
		445:   "SMB",
		1433:  "MSSQL",
		3306:  "MySQL",
		3389:  "RDP",
		5432:  "PostgreSQL",
		5900:  "VNC",
		6379:  "Redis",
		8080:  "HTTP-Proxy",
		8443:  "HTTPS-Alt",
		9200:  "Elasticsearch",
		27017: "MongoDB",
		123:   "NTP",
		161:   "SNMP",
		500:   "IKE",
		1900:  "SSDP",
	}

	if service, ok := services[port]; ok {
		return service
	}
	return "Unknown"
}
