package scan

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cskg/CyberStroll/internal/config"
	"github.com/cskg/CyberStroll/pkg/models"
)

// ProbeStats 单个探测包的统计信息
type ProbeStats struct {
	Name          string `json:"name"`          // 探测包名称
	Sent          int64  `json:"sent"`          // 发送次数
	Successful    int64  `json:"successful"`    // 成功匹配次数
	HitRate       float64 `json:"hit_rate"`     // 命中率
	LastUpdated   time.Time `json:"last_updated"` // 最后更新时间
}

// ProbeStatistics 探测包统计数据管理
type ProbeStatistics struct {
	Stats    map[string]*ProbeStats `json:"stats"`    // 探测包统计信息
	LastSave time.Time             `json:"last_save"` // 最后保存时间
}

// Scanner 扫描器结构体
type Scanner struct {
	config        *config.ScanConfig
	probeStats    *ProbeStatistics
	statsMutex    sync.RWMutex
	statsFilePath string
}

// NewScanner 创建扫描器实例
func NewScanner(cfg *config.ScanConfig) *Scanner {
	scanner := &Scanner{
		config:        cfg,
		statsFilePath: "./probe_stats.json",
	}
	
	// 初始化探测包统计
	scanner.initProbeStats()
	
	// 加载已有的统计数据
	scanner.loadProbeStats()
	
	return scanner
}

// initProbeStats 初始化探测包统计数据
func (s *Scanner) initProbeStats() {
	s.probeStats = &ProbeStatistics{
		Stats: make(map[string]*ProbeStats),
	}
	
	// 定义基础探测包名称列表，避免循环依赖
	probeNames := []string{"http", "ssh", "ftp", "smtp", "pop3", "imap", "mysql", "postgresql", "mssql", "mqtt"}
	
	// 为所有探测包初始化统计数据
	for _, name := range probeNames {
		s.probeStats.Stats[name] = &ProbeStats{
			Name:        name,
			Sent:        0,
			Successful:  0,
			HitRate:     0.0,
			LastUpdated: time.Now(),
		}
	}
}

// loadProbeStats 从文件加载探测包统计数据
func (s *Scanner) loadProbeStats() {
	data, err := os.ReadFile(s.statsFilePath)
	if err != nil {
		// 文件不存在，使用默认统计
		return
	}
	
	var savedStats ProbeStatistics
	err = json.Unmarshal(data, &savedStats)
	if err != nil {
		fmt.Printf("警告: 加载探测包统计数据失败: %v，使用默认统计\n", err)
		return
	}
	
	// 合并已保存的统计数据
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	
	for name, stats := range savedStats.Stats {
		if _, exists := s.probeStats.Stats[name]; exists {
			// 保留已有统计，只更新非零值
			if stats.Sent > 0 {
				s.probeStats.Stats[name].Sent += stats.Sent
			}
			if stats.Successful > 0 {
				s.probeStats.Stats[name].Successful += stats.Successful
			}
			if stats.HitRate > 0 {
				// 重新计算命中率
				total := s.probeStats.Stats[name].Sent
				if total > 0 {
					s.probeStats.Stats[name].HitRate = float64(s.probeStats.Stats[name].Successful) / float64(total)
				}
			}
		}
	}
	
	fmt.Printf("成功加载探测包统计数据，共 %d 个探测包\n", len(savedStats.Stats))
}

// saveProbeStats 将探测包统计数据保存到文件
func (s *Scanner) saveProbeStats() {
	s.statsMutex.RLock()
	stats := s.probeStats
	s.statsMutex.RUnlock()
	
	// 更新最后保存时间
	stats.LastSave = time.Now()
	
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		fmt.Printf("警告: 序列化探测包统计数据失败: %v\n", err)
		return
	}
	
	err = os.WriteFile(s.statsFilePath, data, 0644)
	if err != nil {
		fmt.Printf("警告: 保存探测包统计数据失败: %v\n", err)
		return
	}
	
	fmt.Printf("成功保存探测包统计数据到 %s\n", s.statsFilePath)
}

// updateProbeStats 更新单个探测包的统计数据
func (s *Scanner) updateProbeStats(probeName string, successful bool) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	
	stats, exists := s.probeStats.Stats[probeName]
	if !exists {
		// 如果统计不存在，创建新的
		stats = &ProbeStats{
			Name:        probeName,
			Sent:        0,
			Successful:  0,
			HitRate:     0.0,
			LastUpdated: time.Now(),
		}
		s.probeStats.Stats[probeName] = stats
	}
	
	// 更新统计数据
	stats.Sent++
	if successful {
		stats.Successful++
	}
	
	// 重新计算命中率
	stats.HitRate = float64(stats.Successful) / float64(stats.Sent)
	stats.LastUpdated = time.Now()
}

// getAllProbes 获取所有默认探测包，按命中率排序
func (s *Scanner) getAllProbes() []ProbeInfo {
	// 基础探测包列表
	probes := []ProbeInfo{
		{Name: "http", Probe: []byte("GET / HTTP/1.1\r\nHost: {{host}}\r\n\r\n"), App: "Apache, Nginx, Microsoft IIS, PHP"},
		{Name: "ssh", Probe: []byte("SSH-2.0-OpenSSH_7.0\r\n"), App: "OpenSSH"},
		{Name: "ftp", Probe: []byte("USER anonymous\r\n"), App: "FTP Server"},
		{Name: "smtp", Probe: []byte("EHLO example.com\r\n"), App: "SMTP Server"},
		{Name: "pop3", Probe: []byte("USER test\r\n"), App: "POP3 Server"},
		{Name: "imap", Probe: []byte("A001 LOGIN test test\r\n"), App: "IMAP Server"},
		{Name: "mysql", Probe: []byte{0x05, 0x00, 0x00, 0x01, 0x85, 0xa6, 0x3f, 0x00, 0x00, 0x00, 0x00, 0x01, 0x21, 0x00}, App: "MySQL"},
		{Name: "postgresql", Probe: []byte{0x00, 0x00, 0x00, 0x32, 0x00, 0x03, 0x00, 0x00, 0x75, 0x73, 0x65, 0x72, 0x3d, 0x70, 0x6f, 0x73, 0x74, 0x67, 0x72, 0x65, 0x73, 0x00, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x3d, 0x70, 0x6f, 0x73, 0x74, 0x67, 0x72, 0x65, 0x73, 0x00, 0x00}, App: "PostgreSQL"},
		{Name: "mssql", Probe: []byte{0x02, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1a, 0x00, 0x06, 0x01, 0x00, 0x1b, 0x00, 0x01, 0x02, 0x00, 0x1c, 0x00, 0x0c, 0x03, 0x00, 0x28, 0x00, 0x04, 0x04, 0x00, 0x24, 0x00, 0x00, 0x05, 0x00, 0x26, 0x00, 0x08, 0x06, 0x01, 0x57, 0x00, 0x00, 0x00, 0x00}, App: "Microsoft SQL Server"},
		{Name: "mqtt", Probe: []byte{0x10, 0x0c, 0x00, 0x04, 0x4d, 0x51, 0x54, 0x54, 0x04, 0x02, 0x00, 0x3c, 0x00, 0x00}, App: "MQTT Broker"},
	}
	
	// 按命中率排序探测包
	s.sortProbesByHitRate(&probes)
	
	return probes
}

// sortProbesByHitRate 按命中率排序探测包
func (s *Scanner) sortProbesByHitRate(probes *[]ProbeInfo) {
	// 获取当前的命中率统计
	s.statsMutex.RLock()
	statsMap := make(map[string]*ProbeStats)
	for name, stats := range s.probeStats.Stats {
		// 创建副本，避免并发问题
		statsCopy := *stats
		statsMap[name] = &statsCopy
	}
	s.statsMutex.RUnlock()
	
	// 按命中率排序
	sort.Slice(*probes, func(i, j int) bool {
		// 获取两个探测包的命中率
		hitRateI := 0.0
		if stats, exists := statsMap[(*probes)[i].Name]; exists {
			hitRateI = stats.HitRate
		}
		
		hitRateJ := 0.0
		if stats, exists := statsMap[(*probes)[j].Name]; exists {
			hitRateJ = stats.HitRate
		}
		
		// 命中率高的排在前面
		return hitRateI > hitRateJ
	})
	
	// 打印排序后的探测包
	fmt.Println("按命中率排序的探测包:")
	for i, probe := range *probes {
		if stats, exists := statsMap[probe.Name]; exists {
			fmt.Printf("%d. %s (命中率: %.2f%%)\n", i+1, probe.Name, stats.HitRate*100)
		} else {
			fmt.Printf("%d. %s (命中率: 0.00%%)\n", i+1, probe.Name)
		}
	}
}

// Scan 执行扫描任务
func (s *Scanner) Scan(ctx context.Context, task *models.Task) ([]*models.ScanResult, error) {
	var results []*models.ScanResult

	switch task.Type {
	case models.TaskTypeIPAlive:
		// IP探活
		ipResults, err := s.IPAliveScan(ctx, task)
		if err != nil {
			return nil, fmt.Errorf("ip alive scan failed: %w", err)
		}
		results = append(results, ipResults...)

	case models.TaskTypePortScan:
		// 端口扫描
		portResults, err := s.PortScan(ctx, task)
		if err != nil {
			return nil, fmt.Errorf("port scan failed: %w", err)
		}
		results = append(results, portResults...)

	case models.TaskTypeServiceScan:
		// 服务识别
		serviceResults, err := s.ServiceScan(ctx, task)
		if err != nil {
			return nil, fmt.Errorf("service scan failed: %w", err)
		}
		results = append(results, serviceResults...)

	case models.TaskTypeWebScan:
		// 网站识别
		webResults, err := s.WebScan(ctx, task)
		if err != nil {
			return nil, fmt.Errorf("web scan failed: %w", err)
		}
		results = append(results, webResults...)

	case models.TaskTypeBannerGrab:
		// Banner抓取
		bannerResults, err := s.BannerGrab(ctx, task)
		if err != nil {
			return nil, fmt.Errorf("banner grab failed: %w", err)
		}
		results = append(results, bannerResults...)

	default:
		return nil, fmt.Errorf("unknown task type: %s", task.Type)
	}

	return results, nil
}

// IPAliveScan IP探活扫描
func (s *Scanner) IPAliveScan(ctx context.Context, task *models.Task) ([]*models.ScanResult, error) {
	var results []*models.ScanResult
	var wg sync.WaitGroup
	resultChan := make(chan *models.ScanResult, len(task.Targets))
	errChan := make(chan error, 1)

	// 限制并发数
	semaphore := make(chan struct{}, s.config.Threads)

	for _, target := range task.Targets {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
			}

			// 使用Ping检查IP是否存活
			isAlive := s.pingIP(ip)

			result := &models.ScanResult{
				IP:        ip,
				Port:      0,
				Protocol:  string(task.Protocol),
				Service:   "",
				Banner:    "",
				Status:    func() string {
					if isAlive {
						return "alive"
					} else {
						return "dead"
					}
				}(),
				CreatedAt: time.Now(),
				TaskID:    task.ID,
			}

			resultChan <- result
		}(target)
	}

	// 等待所有协程完成
	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	// 收集结果
	for {
		select {
		case err := <-errChan:
			if err != nil {
				return nil, err
			}
		case result, ok := <-resultChan:
			if !ok {
				goto done
			}
			results = append(results, result)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

done:
	return results, nil
}

// pingIP 使用TCP连接检查IP是否存活
func (s *Scanner) pingIP(ip string) bool {
	// 使用TCP连接到80端口检查IP是否存活
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:80", ip), time.Duration(s.config.Timeout)*time.Second)
	if err != nil {
		// 尝试连接到443端口
		conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:443", ip), time.Duration(s.config.Timeout)*time.Second)
		if err != nil {
			return false
		}
	}
	defer conn.Close()
	return true
}

// PortScan 端口扫描
func (s *Scanner) PortScan(ctx context.Context, task *models.Task) ([]*models.ScanResult, error) {
	var results []*models.ScanResult
	var wg sync.WaitGroup
	resultChan := make(chan *models.ScanResult, 100)
	errChan := make(chan error, 1)

	// 解析端口范围
	ports := s.parsePortRange(task.PortRange)
	if len(ports) == 0 {
		// 使用配置文件中的默认端口
		ports = s.parsePortRange(s.config.DefaultPorts)
		// 如果配置中没有默认端口或解析失败，使用内置默认端口
		if len(ports) == 0 {
			ports = []int{80, 443, 8080}
		}
	}

	// 限制并发数
	semaphore := make(chan struct{}, s.config.Threads)

	// 根据协议类型决定要扫描的协议列表
	var protocols []models.ProtocolType
	switch task.Protocol {
	case models.ProtocolTCP:
		protocols = []models.ProtocolType{models.ProtocolTCP}
	case models.ProtocolUDP:
		protocols = []models.ProtocolType{models.ProtocolUDP}
	case models.ProtocolAll:
		protocols = []models.ProtocolType{models.ProtocolTCP, models.ProtocolUDP}
	default:
		protocols = []models.ProtocolType{models.ProtocolTCP}
	}

	for _, target := range task.Targets {
		for _, port := range ports {
			for _, protocol := range protocols {
				wg.Add(1)
				go func(ip string, p int, proto models.ProtocolType) {
					defer wg.Done()
					semaphore <- struct{}{}
					defer func() { <-semaphore }()

					// 检查上下文是否已取消
					select {
					case <-ctx.Done():
						errChan <- ctx.Err()
						return
					default:
					}

					// 扫描端口
							isOpen := s.scanPort(ip, p, proto)

							status := "closed"
							if isOpen {
								status = "open"
							}

							result := &models.ScanResult{
								IP:        ip,
								Port:      p,
								Protocol:  string(proto),
								Service:   "",
								App:       "",
								Banner:    "",
								Status:    status,
								CreatedAt: time.Now(),
								TaskID:    task.ID,
							}

							if isOpen {
								// 抓取初始Banner
								initialBanner := s.GrabBanner(ip, p)
								
								// 使用探测包检测服务和应用，获取完整banner信息
								serviceInfo := s.detectService(ip, p, string(proto), initialBanner)
								
								// 更新结果
								result.Service = serviceInfo.Service
								result.App = serviceInfo.App
								result.Banner = serviceInfo.Banner
								resultChan <- result
							}
			}(target, port, protocol)
			}
		}
	}

	// 等待所有协程完成
	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	// 收集结果
	for {
		select {
		case err := <-errChan:
			if err != nil {
				return nil, err
			}
		case result, ok := <-resultChan:
			if !ok {
				goto done
			}
			results = append(results, result)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

done:
	return results, nil
}

// scanPort 扫描单个端口
func (s *Scanner) scanPort(ip string, port int, protocol models.ProtocolType) bool {
	target := fmt.Sprintf("%s:%d", ip, port)
	timeout := time.Duration(s.config.Timeout) * time.Second

	var conn net.Conn
	var err error

	switch protocol {
	case models.ProtocolTCP, models.ProtocolAll:
		conn, err = net.DialTimeout("tcp", target, timeout)
	case models.ProtocolUDP:
		conn, err = net.DialTimeout("udp", target, timeout)
	default:
		conn, err = net.DialTimeout("tcp", target, timeout)
	}

	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}

// parsePortRange 解析端口范围字符串
func (s *Scanner) parsePortRange(portRange string) []int {
	// TODO: 实现端口范围解析逻辑
	// 支持格式：80,443,8080 或 1-1000

	// 目前返回默认端口，包含MQTT默认端口1883
	return []int{80, 443, 8080, 22, 3389, 3306, 5432, 1433, 1883}
}

// ServiceScan 服务识别
func (s *Scanner) ServiceScan(ctx context.Context, task *models.Task) ([]*models.ScanResult, error) {
	// 先执行端口扫描
	portResults, err := s.PortScan(ctx, task)
	if err != nil {
		return nil, err
	}

	// 对开放的端口进行服务识别
	var results []*models.ScanResult
	var wg sync.WaitGroup
	resultChan := make(chan *models.ScanResult, len(portResults))
	errChan := make(chan error, 1)

	// 限制并发数
	semaphore := make(chan struct{}, s.config.Threads)

	for _, result := range portResults {
		if result.Status != "open" {
			continue
		}

		wg.Add(1)
		go func(res *models.ScanResult) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
			}

			// 抓取Banner
			initialBanner := s.GrabBanner(res.IP, res.Port)

			// 使用探测包检测服务和应用，获取完整banner信息
			serviceInfo := s.detectService(res.IP, res.Port, res.Protocol, initialBanner)

			newResult := &models.ScanResult{
				IP:        res.IP,
				Port:      res.Port,
				Protocol:  res.Protocol,
				Service:   serviceInfo.Service,
				App:       serviceInfo.App,
				Banner:    serviceInfo.Banner, // 使用包含所有探测响应的完整banner
				Status:    res.Status,
				CreatedAt: time.Now(),
				TaskID:    task.ID,
			}

			resultChan <- newResult
		}(result)
	}

	// 等待所有协程完成
	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	// 收集结果
	for {
		select {
		case err := <-errChan:
			if err != nil {
				return nil, err
			}
		case result, ok := <-resultChan:
			if !ok {
				goto done
			}
			results = append(results, result)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

done:
	return results, nil
}

// ServiceInfo 服务探测信息
type ServiceInfo struct {
	Service string
	App     string
	Banner  string
}

// ProbeInfo 探测包信息
type ProbeInfo struct {
	Name  string
	Probe []byte
	App   string
}



// getServiceFromApp 根据应用类型获取服务类型
func (s *Scanner) getServiceFromApp(app string) string {
	// 根据应用类型映射到服务类型
	appServiceMap := map[string]string{
		"Apache":              "http",
		"Nginx":               "http",
		"Microsoft IIS":       "http",
		"PHP":                 "http",
		"OpenSSH":             "ssh",
		"FTP Server":          "ftp",
		"SMTP Server":         "smtp",
		"POP3 Server":         "pop3",
		"IMAP Server":         "imap",
		"MySQL":               "mysql",
		"PostgreSQL":          "postgresql",
		"Microsoft SQL Server": "mssql",
		"MQTT Broker":         "mqtt",
	}

	if service, ok := appServiceMap[app]; ok {
		return service
	}

	return "unknown"
}

// detectService 使用探测包检测服务类型，结合所有banner数据
func (s *Scanner) detectService(ip string, port int, protocol string, initialBanner string) ServiceInfo {
	// 完整的banner信息，包含初始banner和所有探测响应
	completeBanner := initialBanner
	app := ""
	var bestMatchConfidence int
	var bestMatchApp string

	// 如果banner已经提供了足够信息，直接使用
	if initialBanner != "" {
		// 基于banner进行更精确的服务识别
		if strings.Contains(initialBanner, "SSH") {
			app = "OpenSSH"
		} else if strings.Contains(initialBanner, "Apache") {
			app = "Apache"
		} else if strings.Contains(initialBanner, "nginx") {
			app = "Nginx"
		} else if strings.Contains(initialBanner, "Microsoft-IIS") {
			app = "Microsoft IIS"
		} else if strings.Contains(initialBanner, "MySQL") {
			app = "MySQL"
		} else if strings.Contains(initialBanner, "PostgreSQL") {
			app = "PostgreSQL"
		} else if strings.Contains(initialBanner, "SMTP") {
			app = "SMTP Server"
		}
	}

	// 发送所有默认探测包，按命中率排序，不依赖于端口猜测
	probes := s.getAllProbes()
	for _, probeInfo := range probes {
		successful := false
		target := fmt.Sprintf("%s:%d", ip, port)
		conn, err := net.DialTimeout("tcp", target, time.Duration(s.config.Timeout)*time.Second)
		if err != nil {
			// 更新探测包统计（失败）
			s.updateProbeStats(probeInfo.Name, false)
			continue
		}
		
		// 设置读写超时
		conn.SetReadDeadline(time.Now().Add(time.Duration(s.config.Timeout) * time.Second))
		conn.SetWriteDeadline(time.Now().Add(time.Duration(s.config.Timeout) * time.Second))
		
		// 发送探测包
		_, err = conn.Write(probeInfo.Probe)
		if err != nil {
			conn.Close()
			// 更新探测包统计（失败）
			s.updateProbeStats(probeInfo.Name, false)
			continue
		}
		
		// 读取响应数据
		buf := make([]byte, 2048) // 增加缓冲区以接收更多二进制数据
		var probeResponse []byte
		
		// 读取所有可用响应数据
		for {
			n, err := conn.Read(buf)
			if err != nil {
				break
			}
			probeResponse = append(probeResponse, buf[:n]...)
			if n < len(buf) {
				// 缓冲区未填满，说明已读取完所有数据
				break
			}
		}
		
		conn.Close()
		
		// 将探测响应转换为字符串
		responseStr := string(probeResponse)
		
		// 将响应添加到完整banner中
		if len(probeResponse) > 0 {
			// 如果有初始banner或之前的响应，添加分隔符
			if completeBanner != "" {
				completeBanner += fmt.Sprintf("\n--- %s PROBE RESPONSE ---", probeInfo.Name)
			}
			completeBanner += "\n" + responseStr
		}
		
		// 分析响应，确定最匹配的应用
			if len(responseStr) > 0 || len(probeResponse) > 0 {
				// 简单的匹配逻辑，可根据实际情况扩展
				confidence := 0
				
				// 检查应用关键词
				if strings.Contains(responseStr, "SSH") {
					bestMatchConfidence = 100
					bestMatchApp = "OpenSSH"
					successful = true
				} else if strings.Contains(responseStr, "HTTP/1.1") {
					if strings.Contains(responseStr, "Apache") {
						confidence = 90
						bestMatchApp = "Apache"
						successful = true
					} else if strings.Contains(responseStr, "nginx") {
						confidence = 90
						bestMatchApp = "Nginx"
						successful = true
					} else if strings.Contains(responseStr, "Microsoft-IIS") {
						confidence = 90
						bestMatchApp = "Microsoft IIS"
						successful = true
					} else if strings.Contains(responseStr, "PHP/") {
						confidence = 80
						bestMatchApp = "PHP"
						successful = true
					} else if strings.Contains(responseStr, "Server:") {
						confidence = 70
						// 提取Server头信息
						lines := strings.Split(responseStr, "\r\n")
						for _, line := range lines {
							if strings.HasPrefix(strings.ToLower(line), "server:") {
								bestMatchApp = strings.TrimSpace(line[7:])
								successful = true
								break
							}
						}
					} else {
						confidence = 60
						bestMatchApp = "HTTP Server"
						successful = true
					}
					if confidence > bestMatchConfidence {
							bestMatchConfidence = confidence
						}
				} else if strings.Contains(responseStr, "220 ") { // FTP banner
					confidence = 85
					if confidence > bestMatchConfidence {
						bestMatchConfidence = confidence
						bestMatchApp = "FTP Server"
						successful = true
					}
				} else if strings.Contains(responseStr, "220") && strings.Contains(responseStr, "SMTP") { // SMTP banner
					confidence = 85
					if confidence > bestMatchConfidence {
						bestMatchConfidence = confidence
						bestMatchApp = "SMTP Server"
						successful = true
					}
				} else if probeInfo.Name == "mqtt" && len(probeResponse) > 2 { // MQTT响应检查
					// MQTT CONNACK响应格式：首字节为0x20，第二个字节为剩余长度
					if probeResponse[0] == 0x20 {
						confidence = 95
						if confidence > bestMatchConfidence {
							bestMatchConfidence = confidence
							bestMatchApp = "MQTT Broker"
							successful = true
						}
					}
				} else if len(responseStr) > 10 || len(probeResponse) > 10 { // 对于二进制响应，检查响应长度
					// 检查数据库响应特征
					if len(probeResponse) == 16 && probeInfo.Name == "mysql" { // MySQL handshake response is typically 16 bytes
						confidence = 95
						if confidence > bestMatchConfidence {
							bestMatchConfidence = confidence
							bestMatchApp = "MySQL"
							successful = true
						}
					} else if len(probeResponse) > 0 && probeInfo.Name == "postgresql" { // PostgreSQL typically returns data
						confidence = 90
						if confidence > bestMatchConfidence {
							bestMatchConfidence = confidence
							bestMatchApp = "PostgreSQL"
							successful = true
						}
					} else if len(probeResponse) > 0 && probeInfo.Name == "mssql" { // MSSQL typically returns data
						confidence = 85
						if confidence > bestMatchConfidence {
							bestMatchConfidence = confidence
							bestMatchApp = "Microsoft SQL Server"
							successful = true
						}
					}
				}
			}
		
		// 更新探测包统计
		s.updateProbeStats(probeInfo.Name, successful)
	}

	// 保存探测包统计
	go s.saveProbeStats()

	// 选择最佳匹配的应用
	if bestMatchConfidence > 0 {
		app = bestMatchApp
	}

	// 如果仍然无法识别，使用默认值
	if app == "" {
		app = "unknown"
	}

	// 根据应用类型获取服务类型
	service := s.getServiceFromApp(app)

	return ServiceInfo{
		Service: service,
		App:     app,
		Banner:  completeBanner,
	}
}

// guessService 根据端口号猜测服务名称
func (s *Scanner) guessService(port int) string {
	serviceMap := map[int]string{
		21:  "ftp",
		22:  "ssh",
		23:  "telnet",
		25:  "smtp",
		53:  "dns",
		80:  "http",
		110: "pop3",
		143: "imap",
		443: "https",
		3306: "mysql",
		5432: "postgresql",
		8080: "http-proxy",
		8443: "https-proxy",
		3389: "rdp",
		1433: "mssql",
	}

	if service, ok := serviceMap[port]; ok {
		return service
	}

	return "unknown"
}

// GrabBanner 抓取端口Banner，支持二进制数据转换
func (s *Scanner) GrabBanner(ip string, port int) string {
	target := fmt.Sprintf("%s:%d", ip, port)
	var banner []byte

	// 尝试连接并获取初始banner
	conn, err := net.DialTimeout("tcp", target, time.Duration(s.config.Timeout)*time.Second)
	if err != nil {
		return ""
	}

	// 设置读写超时
	conn.SetReadDeadline(time.Now().Add(time.Duration(s.config.Timeout) * time.Second))
	conn.SetWriteDeadline(time.Now().Add(time.Duration(s.config.Timeout) * time.Second))

	// 读取初始banner
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err == nil && n > 0 {
		banner = append(banner, buf[:n]...)
	}

	// 如果没有获取到banner，尝试发送HTTP请求
	if len(banner) == 0 {
		httpReq := []byte("GET / HTTP/1.1\r\nHost: " + ip + "\r\n\r\n")
		_, writeErr := conn.Write(httpReq)
		if writeErr == nil {
			// 读取HTTP响应
			for {
				respN, respErr := conn.Read(buf)
				if respErr != nil {
					break
				}
				banner = append(banner, buf[:respN]...)
				if respN < len(buf) {
					break
				}
			}
		}
	}

	conn.Close()

	// 如果还是没有获取到banner，尝试SSH连接
	if len(banner) == 0 {
		sshConn, sshErr := net.DialTimeout("tcp", target, time.Duration(s.config.Timeout)*time.Second)
		if sshErr == nil {
			sshConn.SetReadDeadline(time.Now().Add(time.Duration(s.config.Timeout) * time.Second))
			sshConn.SetWriteDeadline(time.Now().Add(time.Duration(s.config.Timeout) * time.Second))

			// 发送SSH协议标识
			sshReq := []byte("SSH-2.0-OpenSSH_7.0\r\n")
			_, sshWriteErr := sshConn.Write(sshReq)
			if sshWriteErr == nil {
				// 读取SSH响应
				for {
					sshRespN, sshRespErr := sshConn.Read(buf)
					if sshRespErr != nil {
						break
					}
					banner = append(banner, buf[:sshRespN]...)
					if sshRespN < len(buf) {
						break
					}
				}
			}
			sshConn.Close()
		}
	}

	// 将二进制数据转换为字符串
	return string(banner)
}

// WebScan 网站识别（简化版）
func (s *Scanner) WebScan(ctx context.Context, task *models.Task) ([]*models.ScanResult, error) {
	// 对常见Web端口进行扫描
	webTask := *task
	webTask.PortRange = "80,443,8080,8443"

	return s.ServiceScan(ctx, &webTask)
}

// BannerGrab Banner抓取
func (s *Scanner) BannerGrab(ctx context.Context, task *models.Task) ([]*models.ScanResult, error) {
	// 先执行端口扫描
	portResults, err := s.PortScan(ctx, task)
	if err != nil {
		return nil, err
	}

	// 对开放的端口进行Banner抓取和服务识别
	var results []*models.ScanResult
	for _, result := range portResults {
		if result.Status == "open" {
			// 抓取初始Banner
			initialBanner := s.GrabBanner(result.IP, result.Port)
			
			// 使用探测包检测服务和应用，获取完整banner信息
			serviceInfo := s.detectService(result.IP, result.Port, result.Protocol, initialBanner)
			
			// 更新结果
			result.Service = serviceInfo.Service
			result.App = serviceInfo.App
			result.Banner = serviceInfo.Banner
			results = append(results, result)
		}
	}

	return results, nil
}
