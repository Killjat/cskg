package scanner

import (
	"fmt"
	"net"
	"time"

	"ip-discovery/storage"
)

// PingScanner ping扫描器
type PingScanner struct {
	Timeout time.Duration
	Workers int
}

// NewPingScanner 创建新的ping扫描器
func NewPingScanner(timeout time.Duration, workers int) *PingScanner {
	return &PingScanner{
		Timeout: timeout,
		Workers: workers,
	}
}

// ScanIP 扫描单个IP
func (s *PingScanner) ScanIP(ip string) *storage.AliveResult {
	result := &storage.AliveResult{
		IP:       ip,
		ScanTime: time.Now(),
		IsAlive:  false,
	}

	// 尝试TCP连接测试
	alive, duration, err := s.simplePing(ip)
	
	result.ResponseTime = duration
	result.IsAlive = alive

	if err != nil {
		result.ErrorMessage = err.Error()
	}

	return result
}

// ScanCIDR 扫描CIDR段中的IP
func (s *PingScanner) ScanCIDR(cidr string, maxIPs int) ([]*storage.AliveResult, error) {
	// 解析CIDR
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("解析CIDR失败: %v", err)
	}

	// 生成要扫描的IP列表
	ips := s.generateIPList(ipNet, maxIPs)
	
	// 创建结果通道
	results := make(chan *storage.AliveResult, len(ips))
	jobs := make(chan string, len(ips))

	// 启动工作协程
	for i := 0; i < s.Workers; i++ {
		go func() {
			for ip := range jobs {
				result := s.ScanIP(ip)
				result.CIDR = cidr
				results <- result
			}
		}()
	}

	// 发送任务
	for _, ip := range ips {
		jobs <- ip
	}
	close(jobs)

	// 收集结果
	var scanResults []*storage.AliveResult
	for i := 0; i < len(ips); i++ {
		result := <-results
		scanResults = append(scanResults, result)
	}

	return scanResults, nil
}

// simplePing 简单的连通性测试
func (s *PingScanner) simplePing(ip string) (bool, time.Duration, error) {
	start := time.Now()
	
	// 尝试TCP连接到常见端口
	ports := []string{"80", "443", "22", "21", "25", "53", "110", "143", "993", "995"}
	
	for _, port := range ports {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), s.Timeout)
		if err == nil {
			conn.Close()
			duration := time.Since(start)
			return true, duration, nil
		}
	}
	
	// 如果TCP连接都失败，尝试UDP
	conn, err := net.DialTimeout("udp", net.JoinHostPort(ip, "53"), s.Timeout)
	if err == nil {
		conn.Close()
		duration := time.Since(start)
		return true, duration, nil
	}
	
	duration := time.Since(start)
	return false, duration, fmt.Errorf("无法连接到 %s", ip)
}

// generateIPList 生成要扫描的IP列表
func (s *PingScanner) generateIPList(ipNet *net.IPNet, maxIPs int) []string {
	var ips []string

	// 获取网络地址
	ip := ipNet.IP.To4()
	if ip == nil {
		return ips
	}

	// 获取掩码
	mask := ipNet.Mask
	ones, bits := mask.Size()

	// 计算主机位数
	hostBits := bits - ones
	if hostBits <= 0 {
		return ips
	}

	// 计算可用IP数量（排除网络地址和广播地址）
	maxHosts := (1 << uint(hostBits)) - 2
	if maxHosts <= 0 {
		return ips
	}

	// 限制扫描数量
	scanCount := maxHosts
	if maxIPs > 0 && maxIPs < maxHosts {
		scanCount = maxIPs
	}

	// 生成IP列表（跳过网络地址，从.1开始）
	baseIP := make(net.IP, 4)
	copy(baseIP, ip)

	for i := 1; i <= scanCount; i++ {
		// 计算当前IP
		currentIP := make(net.IP, 4)
		copy(currentIP, baseIP)

		// 根据主机位数修改IP
		if hostBits >= 8 {
			currentIP[3] = byte(i)
		} else {
			// 对于小于/24的网络，需要更复杂的计算
			hostAddr := i
			for j := 3; j >= 0 && hostAddr > 0; j-- {
				if hostBits > 8*(3-j) {
					bitsInByte := hostBits - 8*(3-j)
					if bitsInByte > 8 {
						bitsInByte = 8
					}
					mask := (1 << uint(bitsInByte)) - 1
					currentIP[j] = baseIP[j] + byte(hostAddr&mask)
					hostAddr >>= uint(bitsInByte)
				}
			}
		}

		ips = append(ips, currentIP.String())
	}

	return ips
}
