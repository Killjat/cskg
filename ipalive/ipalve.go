package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Result 定义检测结果结构体
type Result struct {
	IP        string    // IP地址
	Port      int       // 检测端口
	IsAlive   bool      // 是否存活
	CheckTime time.Time // 检测时间
}

// 全局配置参数
var (
	concurrency = flag.Int("c", 50, "并发检测协程数")
	timeout     = flag.Duration("t", 3*time.Second, "TCP连接超时时间")
	ports       = flag.String("p", "80,443,22,21,3389", "检测端口，多个用逗号分隔")
	inputFile   = flag.String("f", "", "待检测IP/IP段的文件路径（每行一个条目）")
	outputFile  = flag.String("o", "ip_alive_result.csv", "存活IP结果输出CSV路径")
)

// 全局锁保护CSV并发写入
var csvMu sync.Mutex

// readIPFile 读取IP文件，返回去重后的条目列表
func readIPFile(filePath string) ([]string, error) {
	var entries []string
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entries = append(entries, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件出错: %v", err)
	}

	// 去重
	uniqueEntries := make([]string, 0, len(entries))
	entryMap := make(map[string]bool)
	for _, entry := range entries {
		if !entryMap[entry] {
			entryMap[entry] = true
			uniqueEntries = append(uniqueEntries, entry)
		}
	}

	if len(uniqueEntries) == 0 {
		return nil, fmt.Errorf("文件无有效IP/IP段条目")
	}

	fmt.Printf("从文件 %s 读取到 %d 个有效条目（去重后）\n", filePath, len(uniqueEntries))
	return uniqueEntries, nil
}

// parseIPEntries 解析所有IP条目为待检测的IP列表
func parseIPEntries(entries []string) ([]string, error) {
	var allIPs []string
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if strings.Contains(entry, "/") {
			ips, err := parseCIDR(entry)
			if err != nil {
				return nil, fmt.Errorf("解析CIDR %s 失败: %v", entry, err)
			}
			allIPs = append(allIPs, ips...)
		} else if strings.Contains(entry, "-") {
			ips, err := parseIPRange(entry)
			if err != nil {
				return nil, fmt.Errorf("解析IP段 %s 失败: %v", entry, err)
			}
			allIPs = append(allIPs, ips...)
		} else if net.ParseIP(entry) != nil {
			allIPs = append(allIPs, entry)
		} else {
			return nil, fmt.Errorf("无效格式: %s", entry)
		}
	}

	// 最终去重
	uniqueIPs := make([]string, 0, len(allIPs))
	ipMap := make(map[string]bool)
	for _, ip := range allIPs {
		if !ipMap[ip] {
			ipMap[ip] = true
			uniqueIPs = append(uniqueIPs, ip)
		}
	}

	return uniqueIPs, nil
}

// parseCIDR 解析CIDR格式，返回可用主机IP列表
func parseCIDR(cidr string) ([]string, error) {
	var ips []string
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	ip := ip2int(ipNet.IP)
	mask := ip2int(net.IP(ipNet.Mask))
	network := ip & mask
	broadcast := network | ^mask

	// 跳过网络地址和广播地址
	for currentIP := network + 1; currentIP < broadcast; currentIP++ {
		ips = append(ips, int2ip(currentIP).String())
	}
	return ips, nil
}

// parseIPRange 解析起始-结束IP段
func parseIPRange(rangeStr string) ([]string, error) {
	var ips []string
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("格式应为 起始IP-结束IP")
	}

	startIP := net.ParseIP(strings.TrimSpace(parts[0]))
	endIP := net.ParseIP(strings.TrimSpace(parts[1]))
	if startIP == nil || endIP == nil {
		return nil, fmt.Errorf("IP格式错误")
	}

	start := ip2int(startIP)
	end := ip2int(endIP)
	if start > end {
		return nil, fmt.Errorf("起始IP大于结束IP")
	}

	for currentIP := start; currentIP <= end; currentIP++ {
		ips = append(ips, int2ip(currentIP).String())
	}
	return ips, nil
}

// ip2int IP转整数
func ip2int(ip net.IP) uint32 {
	ipv4 := ip.To4()
	if ipv4 == nil {
		return 0
	}
	return uint32(ipv4[0])<<24 | uint32(ipv4[1])<<16 | uint32(ipv4[2])<<8 | uint32(ipv4[3])
}

// int2ip 整数转IP
func int2ip(n uint32) net.IP {
	return net.IPv4(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}

// initCSV 初始化CSV文件，写入表头
func initCSV(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建CSV失败: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"IP地址", "检测端口", "检测时间"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("写入表头失败: %v", err)
	}
	return nil
}

// writeAliveResultToCSV 仅写入存活的结果到CSV
func writeAliveResultToCSV(res Result, filePath string) error {
	// 过滤不存活的结果
	if !res.IsAlive {
		return nil
	}

	csvMu.Lock()
	defer csvMu.Unlock()

	// 追加模式打开文件
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开CSV失败: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	row := []string{
		res.IP,
		strconv.Itoa(res.Port),
		res.CheckTime.Format("2006-01-02 15:04:05"),
	}

	if err := writer.Write(row); err != nil {
		return fmt.Errorf("写入IP %s:%d 失败: %v", res.IP, res.Port, err)
	}
	return nil
}

// checkIP 检测IP端口是否存活，仅存活结果写入CSV
func checkIP(ip string, port int, outputFile string) {
	result := Result{
		IP:        ip,
		Port:      port,
		IsAlive:   false,
		CheckTime: time.Now(),
	}

	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, *timeout)
	if err == nil {
		_ = conn.Close()
		result.IsAlive = true
		fmt.Printf("[存活] %s:%d\n", ip, port)

		// 仅存活结果写入CSV
		if err := writeAliveResultToCSV(result, outputFile); err != nil {
			fmt.Printf("[错误] 写入存活IP %s:%d 失败: %v\n", ip, port, err)
		}
	}
	// 不存活的结果直接忽略，不写入CSV
}

// parsePorts 解析端口参数
func parsePorts(portStr string) ([]int, error) {
	var ports []int
	parts := strings.Split(portStr, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		port, err := strconv.Atoi(p)
		if err != nil || port < 1 || port > 65535 {
			return nil, fmt.Errorf("无效端口: %s", p)
		}
		ports = append(ports, port)
	}
	return ports, nil
}

func main() {
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("请通过 -f 参数指定IP文件路径")
		flag.Usage()
		os.Exit(1)
	}

	portList, err := parsePorts(*ports)
	if err != nil {
		fmt.Printf("解析端口失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化CSV表头
	if err := initCSV(*outputFile); err != nil {
		fmt.Printf("初始化CSV失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("CSV文件已初始化: %s\n", *outputFile)

	entries, err := readIPFile(*inputFile)
	if err != nil {
		fmt.Printf("读取IP文件失败: %v\n", err)
		os.Exit(1)
	}

	ips, err := parseIPEntries(entries)
	if err != nil {
		fmt.Printf("解析IP条目失败: %v\n", err)
		os.Exit(1)
	}
	if len(ips) == 0 {
		fmt.Println("无有效待检测IP")
		os.Exit(1)
	}

	fmt.Printf("\n=== 检测配置 ===\n")
	fmt.Printf("待检测IP总数: %d\n", len(ips))
	fmt.Printf("检测端口: %v\n", portList)
	fmt.Printf("并发数: %d\n", *concurrency)
	fmt.Printf("超时时间: %v\n", *timeout)
	fmt.Printf("存活结果输出: %s\n", *outputFile)
	fmt.Println("=================\n开始检测...")

	var wg sync.WaitGroup
	sem := make(chan struct{}, *concurrency)

	for _, ip := range ips {
		for _, port := range portList {
			wg.Add(1)
			sem <- struct{}{}

			go func(ip string, port int) {
				defer func() {
					<-sem
					wg.Done()
				}()
				checkIP(ip, port, *outputFile)
			}(ip, port)
		}
	}

	wg.Wait()
	fmt.Printf("\n✅ 检测完成！仅存活IP已写入 %s\n", *outputFile)
}