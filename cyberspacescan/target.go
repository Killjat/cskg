package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// Target 扫描目标
type Target struct {
	IP   string
	Type string // "single", "range", "cidr"
}

// LoadTargets 从文件加载目标
func LoadTargets(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var targets []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 移除行内注释
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

		// 再次检查是否为空
		if line == "" {
			continue
		}

		// 解析目标
		parsedTargets, err := parseTarget(line)
		if err != nil {
			fmt.Printf("警告: 解析目标失败 '%s': %v\n", line, err)
			continue
		}

		targets = append(targets, parsedTargets...)
	}

	return targets, scanner.Err()
}

// parseTarget 解析目标格式
func parseTarget(target string) ([]string, error) {
	// CIDR格式: 192.168.1.0/24
	if strings.Contains(target, "/") {
		return parseCIDR(target)
	}

	// IP范围格式: 192.168.1.1-192.168.1.10
	if strings.Contains(target, "-") {
		return parseRange(target)
	}

	// 单个IP
	if net.ParseIP(target) != nil {
		return []string{target}, nil
	}

	return nil, fmt.Errorf("无效的目标格式: %s", target)
}

// parseCIDR 解析CIDR格式
func parseCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// 移除网络地址和广播地址
	if len(ips) > 2 {
		return ips[1 : len(ips)-1], nil
	}

	return ips, nil
}

// parseRange 解析IP范围
func parseRange(ipRange string) ([]string, error) {
	parts := strings.Split(ipRange, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("无效的IP范围格式")
	}

	startIP := net.ParseIP(strings.TrimSpace(parts[0]))
	endIP := net.ParseIP(strings.TrimSpace(parts[1]))

	if startIP == nil || endIP == nil {
		return nil, fmt.Errorf("无效的IP地址")
	}

	var ips []string
	for ip := startIP; !ip.Equal(endIP); inc(ip) {
		ips = append(ips, ip.String())
	}
	ips = append(ips, endIP.String())

	return ips, nil
}

// inc IP地址加1
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// parsePortRange 解析端口范围
func parsePortRange(portStr string) ([]int, error) {
	var ports []int

	// 多个端口: 80,443,8080
	if strings.Contains(portStr, ",") {
		parts := strings.Split(portStr, ",")
		for _, p := range parts {
			port, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil || port < 1 || port > 65535 {
				return nil, fmt.Errorf("无效的端口: %s", p)
			}
			ports = append(ports, port)
		}
		return ports, nil
	}

	// 端口范围: 1-1000
	if strings.Contains(portStr, "-") {
		parts := strings.Split(portStr, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("无效的端口范围格式")
		}

		start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

		if err1 != nil || err2 != nil || start < 1 || end > 65535 || start > end {
			return nil, fmt.Errorf("无效的端口范围")
		}

		for i := start; i <= end; i++ {
			ports = append(ports, i)
		}
		return ports, nil
	}

	// 单个端口
	port, err := strconv.Atoi(strings.TrimSpace(portStr))
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("无效的端口: %s", portStr)
	}

	return []int{port}, nil
}
