package scanner

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// EnhancedProbeEngine 增强版探测引擎
// 集成了network_probe、rule_engine和servicefingerprint的功能
type EnhancedProbeEngine struct {
	config           *ScannerConfig
	stats            *ScannerStats
	bannerRules      map[string]*BannerRule
	fingerprintCache *sync.Map
	mutex            sync.RWMutex
}

// BannerRule Banner匹配规则
type BannerRule struct {
	Service     string `json:"service"`
	Pattern     string `json:"pattern"`
	Version     string `json:"version"`
	Confidence  int    `json:"confidence"`
	Description string `json:"description"`
}

// FingerprintResult 指纹识别结果
type FingerprintResult struct {
	Applications []ApplicationInfo `json:"applications"`
	Technologies []string          `json:"technologies"`
	Server       string            `json:"server"`
	Framework    string            `json:"framework"`
}

// NewEnhancedProbeEngine 创建增强版探测引擎
func NewEnhancedProbeEngine(config *ScannerConfig) *EnhancedProbeEngine {
	if config == nil {
		config = &ScannerConfig{
			MaxConcurrency: 100,
			Timeout:        10 * time.Second,
			RetryCount:     3,
			ProbeDelay:     100 * time.Millisecond,
			EnableLogging:  true,
		}
	}

	engine := &EnhancedProbeEngine{
		config:           config,
		stats:            &ScannerStats{},
		bannerRules:      make(map[string]*BannerRule),
		fingerprintCache: &sync.Map{},
	}

	// 加载内置Banner规则
	engine.loadBuiltinBannerRules()

	return engine
}

// ScanTarget 扫描目标（增强版）
func (epe *EnhancedProbeEngine) ScanTarget(task *ScanTask) (*ScanResult, error) {
	startTime := time.Now()

	result := &ScanResult{
		TaskID:    task.TaskID,
		IP:        task.IP,
		ScanType:  task.TaskType,
		ScanTime:  startTime.Format(time.RFC3339),
		NodeID:    "enhanced-scan-node-001",
		Timestamp: startTime.Unix(),
		Results: ScanDetails{
			OpenPorts:    make([]PortInfo, 0),
			Applications: make([]ApplicationInfo, 0),
		},
	}

	// 确定扫描端口
	ports := epe.determinePorts(task)

	// 并发扫描端口
	openPorts, err := epe.scanPortsEnhanced(task.IP, ports)
	if err != nil {
		result.ScanStatus = "failed"
		result.ErrorMessage = err.Error()
		epe.updateStats(false, time.Since(startTime))
		return result, err
	}

	result.Results.OpenPorts = openPorts

	// 如果启用应用识别，进行Web指纹识别
	if task.Config.EnableApps {
		webPorts := epe.filterWebPorts(openPorts)
		if len(webPorts) > 0 {
			apps := epe.identifyWebApplications(task.IP, webPorts)
			result.Results.Applications = apps
		}
	}

	result.ResponseTime = time.Since(startTime).Milliseconds()
	result.ScanStatus = "success"

	epe.updateStats(true, time.Since(startTime))
	return result, nil
}

// scanPortsEnhanced 增强版端口扫描
func (epe *EnhancedProbeEngine) scanPortsEnhanced(ip string, ports []int) ([]PortInfo, error) {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	openPorts := make([]PortInfo, 0)

	// 创建工作池
	semaphore := make(chan struct{}, epe.config.MaxConcurrency)

	for _, port := range ports {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 扫描单个端口（增强版）
			if portInfo := epe.scanSinglePortEnhanced(ip, p); portInfo != nil {
				mutex.Lock()
				openPorts = append(openPorts, *portInfo)
				mutex.Unlock()
			}

			// 添加延迟
			time.Sleep(epe.config.ProbeDelay)
		}(port)
	}

	wg.Wait()
	return openPorts, nil
}

// scanSinglePortEnhanced 增强版单端口扫描
func (epe *EnhancedProbeEngine) scanSinglePortEnhanced(ip string, port int) *PortInfo {
	address := net.JoinHostPort(ip, strconv.Itoa(port))

	// TCP连接测试
	conn, err := net.DialTimeout("tcp", address, epe.config.Timeout)
	if err != nil {
		return nil // 端口关闭
	}
	defer conn.Close()

	// 获取增强Banner
	banner := epe.grabEnhancedBanner(conn, port)

	// 使用规则引擎识别服务
	service, version, _ := epe.identifyServiceWithRules(port, banner)

	// 获取额外的协议信息
	_ = epe.getProtocolInfo(port, banner)

	return &PortInfo{
		Port:     port,
		Protocol: "tcp",
		Service:  service,
		Version:  version,
		Banner:   banner,
		State:    "open",
		// 可以添加更多字段如置信度、额外信息等
	}
}

// grabEnhancedBanner 获取增强Banner
func (epe *EnhancedProbeEngine) grabEnhancedBanner(conn net.Conn, port int) string {
	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// 根据端口发送特定探测包
	probe := epe.getEnhancedProbeForPort(port)
	if probe != "" {
		conn.Write([]byte(probe))
	}

	// 读取响应
	buffer := make([]byte, 4096) // 增加缓冲区大小
	n, err := conn.Read(buffer)
	if err != nil {
		return ""
	}

	banner := strings.TrimSpace(string(buffer[:n]))

	// 如果是HTTP服务，尝试获取更多信息
	if port == 80 || port == 8080 || port == 443 {
		httpInfo := epe.getHTTPInfo(conn, port)
		if httpInfo != "" {
			banner = httpInfo
		}
	}

	return banner
}

// getEnhancedProbeForPort 获取增强探测包
func (epe *EnhancedProbeEngine) getEnhancedProbeForPort(port int) string {
	probes := map[int]string{
		21:   "",                                                    // FTP
		22:   "",                                                    // SSH
		23:   "",                                                    // Telnet
		25:   "EHLO test\r\n",                                      // SMTP
		53:   "",                                                    // DNS
		80:   "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: CyberStroll/1.0\r\n\r\n", // HTTP
		110:  "",                                                    // POP3
		143:  "",                                                    // IMAP
		443:  "",                                                    // HTTPS
		993:  "",                                                    // IMAPS
		995:  "",                                                    // POP3S
		1433: "",                                                    // SQL Server
		3306: "",                                                    // MySQL
		3389: "",                                                    // RDP
		5432: "",                                                    // PostgreSQL
		6379: "INFO\r\n",                                           // Redis
		8080: "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: CyberStroll/1.0\r\n\r\n", // HTTP Alt
		9200: "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: CyberStroll/1.0\r\n\r\n", // Elasticsearch
		27017: "",                                                   // MongoDB
	}

	if probe, exists := probes[port]; exists {
		return probe
	}
	return ""
}

// identifyServiceWithRules 使用规则引擎识别服务
func (epe *EnhancedProbeEngine) identifyServiceWithRules(port int, banner string) (string, string, int) {
	// 基于端口的初步识别
	service := epe.getServiceByPort(port)
	version := ""
	confidence := 50 // 基于端口的置信度较低

	if banner == "" {
		return service, version, confidence
	}

	// 使用Banner规则进行匹配
	if rule, found := epe.matchBannerRule(banner); found {
		service = rule.Service
		version = rule.Version
		confidence = rule.Confidence
	} else {
		// 使用内置的版本提取逻辑
		version = epe.extractVersionFromBanner(banner, service)
		if version != "" {
			confidence = 80 // 提取到版本信息，置信度较高
		}
	}

	return service, version, confidence
}

// matchBannerRule 匹配Banner规则
func (epe *EnhancedProbeEngine) matchBannerRule(banner string) (*BannerRule, bool) {
	bannerLower := strings.ToLower(banner)

	for _, rule := range epe.bannerRules {
		if strings.Contains(bannerLower, strings.ToLower(rule.Pattern)) {
			return rule, true
		}
	}

	return nil, false
}

// getProtocolInfo 获取协议特定信息
func (epe *EnhancedProbeEngine) getProtocolInfo(port int, banner string) map[string]string {
	info := make(map[string]string)

	switch port {
	case 22: // SSH
		if strings.Contains(banner, "SSH-") {
			parts := strings.Fields(banner)
			if len(parts) > 0 {
				info["ssh_version"] = parts[0]
			}
		}
	case 80, 8080, 443: // HTTP/HTTPS
		if strings.Contains(banner, "Server:") {
			lines := strings.Split(banner, "\n")
			for _, line := range lines {
				if strings.Contains(strings.ToLower(line), "server:") {
					info["server"] = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
				}
				if strings.Contains(strings.ToLower(line), "x-powered-by:") {
					info["powered_by"] = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
				}
			}
		}
	case 6379: // Redis
		if strings.Contains(banner, "redis_version") {
			lines := strings.Split(banner, "\n")
			for _, line := range lines {
				if strings.Contains(line, "redis_version:") {
					info["redis_version"] = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
				}
			}
		}
	}

	return info
}

// identifyWebApplications 识别Web应用
func (epe *EnhancedProbeEngine) identifyWebApplications(ip string, webPorts []PortInfo) []ApplicationInfo {
	var applications []ApplicationInfo

	for _, port := range webPorts {
		// 构造URL
		scheme := "http"
		if port.Port == 443 {
			scheme = "https"
		}
		url := fmt.Sprintf("%s://%s:%d", scheme, ip, port.Port)

		// 检查缓存
		if cached, found := epe.fingerprintCache.Load(url); found {
			if result, ok := cached.(FingerprintResult); ok {
				applications = append(applications, result.Applications...)
				continue
			}
		}

		// 调用Python指纹识别脚本
		apps := epe.callPythonFingerprint(url)
		applications = append(applications, apps...)

		// 缓存结果
		epe.fingerprintCache.Store(url, FingerprintResult{Applications: apps})
	}

	return applications
}

// callPythonFingerprint 调用Python指纹识别
func (epe *EnhancedProbeEngine) callPythonFingerprint(url string) []ApplicationInfo {
	// 检查servicefingerprint脚本是否存在
	scriptPath := "../servicefingerprint/main.py"
	
	// 构造命令
	cmd := exec.Command("python3", scriptPath, "--target", url, "--format", "json", "--timeout", "5")
	
	// 执行命令
	output, err := cmd.Output()
	if err != nil {
		// 如果Python脚本执行失败，返回空结果
		return []ApplicationInfo{}
	}

	// 解析JSON结果
	var result struct {
		Applications []struct {
			Name       string `json:"name"`
			Version    string `json:"version"`
			Category   string `json:"category"`
			Confidence int    `json:"confidence"`
		} `json:"applications"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return []ApplicationInfo{}
	}

	// 转换为内部格式
	apps := make([]ApplicationInfo, len(result.Applications))
	for i, app := range result.Applications {
		apps[i] = ApplicationInfo{
			Name:       app.Name,
			Version:    app.Version,
			Category:   app.Category,
			Confidence: app.Confidence,
		}
	}

	return apps
}

// filterWebPorts 过滤Web端口
func (epe *EnhancedProbeEngine) filterWebPorts(ports []PortInfo) []PortInfo {
	var webPorts []PortInfo
	webPortNumbers := map[int]bool{
		80: true, 443: true, 8080: true, 8443: true,
		8000: true, 8888: true, 9000: true, 3000: true,
	}

	for _, port := range ports {
		if webPortNumbers[port.Port] || 
		   strings.Contains(strings.ToLower(port.Service), "http") {
			webPorts = append(webPorts, port)
		}
	}

	return webPorts
}

// getHTTPInfo 获取HTTP详细信息
func (epe *EnhancedProbeEngine) getHTTPInfo(conn net.Conn, port int) string {
	// 发送HTTP请求
	request := "GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: CyberStroll/1.0\r\nConnection: close\r\n\r\n"
	conn.Write([]byte(request))

	// 读取响应
	buffer := make([]byte, 8192)
	n, err := conn.Read(buffer)
	if err != nil {
		return ""
	}

	response := string(buffer[:n])
	
	// 只返回HTTP头部
	if idx := strings.Index(response, "\r\n\r\n"); idx != -1 {
		return response[:idx]
	}

	return response
}

// loadBuiltinBannerRules 加载内置Banner规则
func (epe *EnhancedProbeEngine) loadBuiltinBannerRules() {
	rules := []*BannerRule{
		// SSH规则
		{Service: "ssh", Pattern: "SSH-2.0-OpenSSH", Version: "OpenSSH", Confidence: 90, Description: "OpenSSH服务器"},
		{Service: "ssh", Pattern: "SSH-2.0-libssh", Version: "libssh", Confidence: 85, Description: "libssh服务器"},
		
		// HTTP服务器规则
		{Service: "http", Pattern: "Server: nginx", Version: "nginx", Confidence: 95, Description: "Nginx Web服务器"},
		{Service: "http", Pattern: "Server: Apache", Version: "Apache", Confidence: 95, Description: "Apache Web服务器"},
		{Service: "http", Pattern: "Server: Microsoft-IIS", Version: "IIS", Confidence: 95, Description: "Microsoft IIS"},
		
		// 数据库规则
		{Service: "mysql", Pattern: "mysql_native_password", Version: "MySQL", Confidence: 90, Description: "MySQL数据库"},
		{Service: "redis", Pattern: "redis_version:", Version: "Redis", Confidence: 95, Description: "Redis数据库"},
		{Service: "postgresql", Pattern: "PostgreSQL", Version: "PostgreSQL", Confidence: 90, Description: "PostgreSQL数据库"},
		
		// FTP规则
		{Service: "ftp", Pattern: "220", Version: "FTP", Confidence: 80, Description: "FTP服务器"},
		{Service: "ftp", Pattern: "vsftpd", Version: "vsftpd", Confidence: 90, Description: "vsftpd FTP服务器"},
		
		// SMTP规则
		{Service: "smtp", Pattern: "220", Version: "SMTP", Confidence: 80, Description: "SMTP服务器"},
		{Service: "smtp", Pattern: "Postfix", Version: "Postfix", Confidence: 90, Description: "Postfix邮件服务器"},
	}

	for _, rule := range rules {
		epe.bannerRules[rule.Pattern] = rule
	}
}

// 其他辅助方法...
func (epe *EnhancedProbeEngine) determinePorts(task *ScanTask) []int {
	if len(task.Config.Ports) > 0 {
		return task.Config.Ports
	}

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

func (epe *EnhancedProbeEngine) getServiceByPort(port int) string {
	services := map[int]string{
		21: "ftp", 22: "ssh", 23: "telnet", 25: "smtp", 53: "dns",
		80: "http", 110: "pop3", 143: "imap", 443: "https",
		993: "imaps", 995: "pop3s", 1433: "ms-sql-s", 3306: "mysql",
		3389: "ms-wbt-server", 5432: "postgresql", 6379: "redis",
		8080: "http-proxy", 9200: "elasticsearch", 27017: "mongodb",
	}
	
	if service, exists := services[port]; exists {
		return service
	}
	return "unknown"
}

func (epe *EnhancedProbeEngine) extractVersionFromBanner(banner, service string) string {
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

	return ""
}

func (epe *EnhancedProbeEngine) updateStats(success bool, duration time.Duration) {
	epe.mutex.Lock()
	defer epe.mutex.Unlock()

	epe.stats.TotalScans++
	if success {
		epe.stats.SuccessScans++
	} else {
		epe.stats.FailedScans++
	}

	if epe.stats.TotalScans > 0 {
		epe.stats.AverageTime = (epe.stats.AverageTime*(epe.stats.TotalScans-1) + duration.Milliseconds()) / epe.stats.TotalScans
	}

	epe.stats.LastScanTime = time.Now().Unix()
}

func (epe *EnhancedProbeEngine) GetStats() *ScannerStats {
	epe.mutex.RLock()
	defer epe.mutex.RUnlock()

	return &ScannerStats{
		TotalScans:   epe.stats.TotalScans,
		SuccessScans: epe.stats.SuccessScans,
		FailedScans:  epe.stats.FailedScans,
		AverageTime:  epe.stats.AverageTime,
		LastScanTime: epe.stats.LastScanTime,
	}
}