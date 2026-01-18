package main

import (
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/cskg/assetdiscovery/common"
)

// ScanExecutor 扫描执行器结构体
type ScanExecutor struct {
	client      *Client
	webScanner  *WebScanner
}

// NewScanExecutor 创建新的扫描执行器
func NewScanExecutor(client *Client) *ScanExecutor {
	return &ScanExecutor{
		client: client,
	}
}

// Execute 执行扫描任务
func (se *ScanExecutor) Execute(task *common.Task) ([]*common.Result, error) {
	var results []*common.Result
	var err error

	// 根据任务类型执行不同的扫描
	switch task.TaskType {
	case common.TaskTypeScanIP:
		results, err = se.ScanIP(task)
	case common.TaskTypeScanPort:
		results, err = se.ScanPort(task)
	case common.TaskTypeScanService:
		results, err = se.ScanService(task)
	case common.TaskTypeScanWeb:
		results, err = se.ScanWeb(task)
	default:
		log.Printf("Unknown task type: %s", task.TaskType)
		// 默认执行完整扫描
		results, err = se.ScanIP(task)
	}

	return results, err
}

// ScanIP 执行IP存活检测
func (se *ScanExecutor) ScanIP(task *common.Task) ([]*common.Result, error) {
	log.Printf("Scanning IP %s...", task.Target)

	// 检测IP是否存活
	isAlive := se.isIPAlive(task.Target)
	if !isAlive {
		log.Printf("IP %s is not alive", task.Target)
		return []*common.Result{}, nil
	}

	// 如果IP存活，执行端口扫描
	return se.ScanPort(task)
}

// ScanPort 执行端口扫描
func (se *ScanExecutor) ScanPort(task *common.Task) ([]*common.Result, error) {
	log.Printf("Scanning ports for %s...", task.Target)

	// 解析端口范围
	portRange := task.PortRange
	if portRange == "" {
		portRange = se.client.config.Scan.PortRange
	}

	// 获取要扫描的端口列表
	ports := se.parsePortRange(portRange)
	if len(ports) == 0 {
		log.Printf("No ports to scan for %s", task.Target)
		return []*common.Result{}, nil
	}

	// 限制并发扫描数量
	concurrentScans := se.client.config.Scan.ConcurrentScans
	if concurrentScans <= 0 {
		concurrentScans = 100
	}

	// 创建结果通道和信号量
	resultChan := make(chan *common.Result, len(ports))
	semaphore := make(chan struct{}, concurrentScans)

	// 并发扫描每个端口
	for _, port := range ports {
		go func(p int) {
			semaphore <- struct{}{} // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			// 检测端口是否开放
			isOpen, protocol := se.isPortOpen(task.Target, p)
			if isOpen {
				// 创建端口扫描结果
				result := &common.Result{
					TaskID:    task.TaskID,
					Target:    task.Target,
					Port:      p,
					Protocol:  protocol,
					Service:   "unknown",
					Status:    "open",
				}

				// 识别服务
				se.identifyService(result)

				// 如果是HTTP/HTTPS服务，进行Web扫描
				if strings.HasPrefix(result.Service, "http") {
					se.scanWebService(result)
				}

				resultChan <- result
			}
		}(port)
	}

	// 收集结果
	var results []*common.Result
	for range ports {
		if result := <-resultChan; result != nil {
			results = append(results, result)
		}
	}

	close(resultChan)
	close(semaphore)

	return results, nil
}

// ScanService 执行服务识别
func (se *ScanExecutor) ScanService(task *common.Task) ([]*common.Result, error) {
	// 简化实现，直接调用端口扫描
	return se.ScanPort(task)
}

// ScanWeb 执行Web站点识别
func (se *ScanExecutor) ScanWeb(task *common.Task) ([]*common.Result, error) {
	// 简化实现，直接调用端口扫描
	return se.ScanPort(task)
}

// isIPAlive 检测IP是否存活
func (se *ScanExecutor) isIPAlive(ip string) bool {
	// 使用ICMP ping检测IP是否存活
	// 简化实现，使用TCP端口80检测
	conn, err := net.DialTimeout("tcp", ip+":80", time.Duration(se.client.config.Scan.Timeout)*time.Second)
	if err != nil {
		conn, err = net.DialTimeout("tcp", ip+":443", time.Duration(se.client.config.Scan.Timeout)*time.Second)
		if err != nil {
			return false
		}
	}
	defer conn.Close()

	return true
}

// isPortOpen 检测端口是否开放
func (se *ScanExecutor) isPortOpen(ip string, port int) (bool, string) {
	// 尝试TCP连接
	target := net.JoinHostPort(ip, string(rune(port)))
	conn, err := net.DialTimeout("tcp", target, time.Duration(se.client.config.Scan.Timeout)*time.Second)
	if err != nil {
		return false, ""
	}
	defer conn.Close()

	return true, "tcp"
}

// parsePortRange 解析端口范围字符串
func (se *ScanExecutor) parsePortRange(portRange string) []int {
	// 简化实现，支持格式：1-1000 或 80,443,8080
	var ports []int

	// 处理逗号分隔的端口列表
	if strings.Contains(portRange, ",") {
		portStrs := strings.Split(portRange, ",")
		for _, portStr := range portStrs {
			portStr = strings.TrimSpace(portStr)
			if port, err := strconv.Atoi(portStr); err == nil {
				ports = append(ports, port)
			}
		}
		return ports
	}

	// 处理范围格式
	if strings.Contains(portRange, "-") {
		parts := strings.Split(portRange, "-")
		if len(parts) == 2 {
			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil && start <= end {
				for port := start; port <= end; port++ {
					ports = append(ports, port)
				}
			}
		}
		return ports
	}

	// 处理单个端口
	if port, err := strconv.Atoi(portRange); err == nil {
		return []int{port}
	}

	// 默认返回常用端口
	return []int{21, 22, 23, 25, 53, 80, 443, 8080, 8443, 3306, 3389, 5432}
}

// identifyService 识别端口上的服务
func (se *ScanExecutor) identifyService(result *common.Result) {
	// 根据端口号识别常见服务
	portServiceMap := map[int]string{
		21:  "ftp",
		22:  "ssh",
		23:  "telnet",
		25:  "smtp",
		53:  "dns",
		80:  "http",
		443: "https",
		3306: "mysql",
		8080: "http",
		8443: "https",
	}

	if service, exists := portServiceMap[result.Port]; exists {
		result.Service = service
	}
}

// scanWebService 扫描Web服务
func (se *ScanExecutor) scanWebService(result *common.Result) {
	if se.webScanner == nil {
		// 如果WebScanner未初始化，创建一个
		se.webScanner = NewWebScanner(se.client)
	}

	// 使用WebScanner扫描Web服务
	se.webScanner.ScanWebService(result)
}
