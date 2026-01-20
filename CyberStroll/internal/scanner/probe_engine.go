package scanner

import (
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ProbeEngine 探测引擎
type ProbeEngine struct {
	config *ScannerConfig
	stats  *ScannerStats
	mutex  sync.RWMutex
}

// NewProbeEngine 创建探测引擎
func NewProbeEngine(config *ScannerConfig) *ProbeEngine {
	if config == nil {
		config = &ScannerConfig{
			MaxConcurrency: 100,
			Timeout:        10 * time.Second,
			RetryCount:     3,
			ProbeDelay:     100 * time.Millisecond,
			EnableLogging:  true,
		}
	}

	return &ProbeEngine{
		config: config,
		stats: &ScannerStats{
			TotalScans:   0,
			SuccessScans: 0,
			FailedScans:  0,
		},
	}
}

// ScanTarget 扫描目标
func (pe *ProbeEngine) ScanTarget(task *ScanTask) (*ScanResult, error) {
	startTime := time.Now()
	
	result := &ScanResult{
		TaskID:    task.TaskID,
		IP:        task.IP,
		ScanType:  task.TaskType,
		ScanTime:  startTime.Format(time.RFC3339),
		NodeID:    "scan-node-001", // TODO: 从配置获取
		Timestamp: startTime.Unix(),
		Results: ScanDetails{
			OpenPorts:    make([]PortInfo, 0),
			Applications: make([]ApplicationInfo, 0),
		},
	}

	// 确定扫描端口
	ports := pe.determinePorts(task)
	
	// 并发扫描端口
	openPorts, err := pe.scanPorts(task.IP, ports)
	if err != nil {
		result.ScanStatus = "failed"
		result.ErrorMessage = err.Error()
		pe.updateStats(false, time.Since(startTime))
		return result, err
	}

	result.Results.OpenPorts = openPorts
	result.ResponseTime = time.Since(startTime).Milliseconds()
	result.ScanStatus = "success"
	
	pe.updateStats(true, time.Since(startTime))
	return result, nil
}

// determinePorts 确定扫描端口
func (pe *ProbeEngine) determinePorts(task *ScanTask) []int {
	if len(task.Config.Ports) > 0 {
		return task.Config.Ports
	}

	// 根据任务类型确定端口
	switch task.TaskType {
	case "port_scan_specified":
		return task.Config.Ports
	case "port_scan_default":
		return getDefaultPorts()
	case "port_scan_full":
		return getFullPortRange()
	case "app_identification":
		return getWebPorts()
	default:
		return getDefaultPorts()
	}
}

// scanPorts 并发扫描端口
func (pe *ProbeEngine) scanPorts(ip string, ports []int) ([]PortInfo, error) {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	openPorts := make([]PortInfo, 0)
	
	// 创建工作池
	semaphore := make(chan struct{}, pe.config.MaxConcurrency)
	
	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			
			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			// 扫描单个端口
			if portInfo := pe.scanSinglePort(ip, p); portInfo != nil {
				mutex.Lock()
				openPorts = append(openPorts, *portInfo)
				mutex.Unlock()
			}
			
			// 添加延迟避免过于激进
			time.Sleep(pe.config.ProbeDelay)
		}(port)
	}
	
	wg.Wait()
	return openPorts, nil
}

// scanSinglePort 扫描单个端口
func (pe *ProbeEngine) scanSinglePort(ip string, port int) *PortInfo {
	address := net.JoinHostPort(ip, strconv.Itoa(port))
	
	// TCP连接测试
	conn, err := net.DialTimeout("tcp", address, pe.config.Timeout)
	if err != nil {
		return nil // 端口关闭
	}
	defer conn.Close()
	
	// 获取Banner
	banner := pe.grabBanner(conn, port)
	
	// 识别服务
	service, version := pe.identifyService(port, banner)
	
	return &PortInfo{
		Port:     port,
		Protocol: "tcp",
		Service:  service,
		Version:  version,
		Banner:   banner,
		State:    "open",
	}
}

// grabBanner 获取服务Banner
func (pe *ProbeEngine) grabBanner(conn net.Conn, port int) string {
	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
	// 根据端口发送探测包
	probe := pe.getProbeForPort(port)
	if probe != "" {
		conn.Write([]byte(probe))
	}
	
	// 读取响应
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return ""
	}
	
	return strings.TrimSpace(string(buffer[:n]))
}

// getProbeForPort 获取端口探测包
func (pe *ProbeEngine) getProbeForPort(port int) string {
	probes := map[int]string{
		21:   "",                                    // FTP - 等待Banner
		22:   "",                                    // SSH - 等待Banner
		23:   "",                                    // Telnet - 等待Banner
		25:   "",                                    // SMTP - 等待Banner
		53:   "",                                    // DNS
		80:   "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n", // HTTP
		110:  "",                                    // POP3 - 等待Banner
		143:  "",                                    // IMAP - 等待Banner
		443:  "",                                    // HTTPS
		993:  "",                                    // IMAPS
		995:  "",                                    // POP3S
		1433: "",                                    // SQL Server
		3306: "",                                    // MySQL
		3389: "",                                    // RDP
		5432: "",                                    // PostgreSQL
		6379: "INFO\r\n",                           // Redis
		8080: "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n", // HTTP Alt
	}
	
	if probe, exists := probes[port]; exists {
		return probe
	}
	return ""
}

// identifyService 识别服务
func (pe *ProbeEngine) identifyService(port int, banner string) (string, string) {
	// 基于端口的服务识别
	portServices := map[int]string{
		21:   "ftp",
		22:   "ssh",
		23:   "telnet",
		25:   "smtp",
		53:   "dns",
		80:   "http",
		110:  "pop3",
		143:  "imap",
		443:  "https",
		993:  "imaps",
		995:  "pop3s",
		1433: "ms-sql-s",
		3306: "mysql",
		3389: "ms-wbt-server",
		5432: "postgresql",
		6379: "redis",
		8080: "http-proxy",
	}
	
	service := "unknown"
	if s, exists := portServices[port]; exists {
		service = s
	}
	
	// 基于Banner的服务和版本识别
	version := pe.extractVersionFromBanner(banner, service)
	
	return service, version
}

// extractVersionFromBanner 从Banner提取版本信息
func (pe *ProbeEngine) extractVersionFromBanner(banner, service string) string {
	if banner == "" {
		return ""
	}
	
	banner = strings.ToLower(banner)
	
	// SSH版本提取
	if service == "ssh" && strings.Contains(banner, "ssh-") {
		parts := strings.Fields(banner)
		if len(parts) > 0 {
			return parts[0]
		}
	}
	
	// HTTP服务器版本提取
	if (service == "http" || service == "https") && strings.Contains(banner, "server:") {
		lines := strings.Split(banner, "\n")
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "server:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) > 1 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}
	
	// MySQL版本提取
	if service == "mysql" && strings.Contains(banner, "mysql") {
		// MySQL banner通常包含版本信息
		return "mysql"
	}
	
	// Redis版本提取
	if service == "redis" && strings.Contains(banner, "redis_version") {
		lines := strings.Split(banner, "\n")
		for _, line := range lines {
			if strings.Contains(line, "redis_version:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) > 1 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}
	
	return ""
}

// updateStats 更新统计信息
func (pe *ProbeEngine) updateStats(success bool, duration time.Duration) {
	pe.mutex.Lock()
	defer pe.mutex.Unlock()
	
	pe.stats.TotalScans++
	if success {
		pe.stats.SuccessScans++
	} else {
		pe.stats.FailedScans++
	}
	
	// 更新平均时间
	if pe.stats.TotalScans > 0 {
		pe.stats.AverageTime = (pe.stats.AverageTime*(pe.stats.TotalScans-1) + duration.Milliseconds()) / pe.stats.TotalScans
	}
	
	pe.stats.LastScanTime = time.Now().Unix()
}

// GetStats 获取统计信息
func (pe *ProbeEngine) GetStats() *ScannerStats {
	pe.mutex.RLock()
	defer pe.mutex.RUnlock()
	
	return &ScannerStats{
		TotalScans:   pe.stats.TotalScans,
		SuccessScans: pe.stats.SuccessScans,
		FailedScans:  pe.stats.FailedScans,
		AverageTime:  pe.stats.AverageTime,
		LastScanTime: pe.stats.LastScanTime,
	}
}

// 预定义端口列表
func getDefaultPorts() []int {
	return []int{
		21, 22, 23, 25, 53, 80, 110, 135, 139, 143, 443, 445, 993, 995,
		1723, 3306, 3389, 5432, 6379, 8080, 8443, 9200, 27017,
	}
}

func getWebPorts() []int {
	return []int{
		80, 443, 8080, 8443, 8000, 8888, 9000, 9090, 3000, 5000,
	}
}

func getFullPortRange() []int {
	ports := make([]int, 65535)
	for i := 1; i <= 65535; i++ {
		ports[i-1] = i
	}
	return ports
}