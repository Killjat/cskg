package apnic

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"ip-discovery/storage"
)

// Parser APNIC数据解析器
type Parser struct {
	CountryCode string // 目标国家代码，如 "TW"
}

// NewParser 创建新的APNIC数据解析器
func NewParser(countryCode string) *Parser {
	return &Parser{
		CountryCode: countryCode,
	}
}

// ParseFile 解析APNIC文件
func (p *Parser) ParseFile(filename string) ([]*storage.IPSegment, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	var segments []*storage.IPSegment
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 跳过注释和空行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析行
		segment, err := p.parseLine(line)
		if err != nil {
			fmt.Printf("解析第%d行失败: %v\n", lineNum, err)
			continue
		}

		if segment != nil {
			segments = append(segments, segment)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	fmt.Printf("解析完成，共找到 %d 个%s的IP段\n", len(segments), p.CountryCode)
	return segments, nil
}

// parseLine 解析单行数据
func (p *Parser) parseLine(line string) (*storage.IPSegment, error) {
	// APNIC格式: registry|cc|type|start|value|date|status[|extensions...]
	fields := strings.Split(line, "|")
	if len(fields) < 7 {
		return nil, nil // 跳过格式不正确的行
	}

	registry := fields[0]
	cc := fields[1]
	recordType := fields[2]
	start := fields[3]
	value := fields[4]
	date := fields[5]
	status := fields[6]

	// 只处理目标国家的IPv4记录
	if cc != p.CountryCode || recordType != "ipv4" {
		return nil, nil
	}

	// 解析IP数量
	ipCount, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("解析IP数量失败: %v", err)
	}

	// 计算CIDR和结束IP
	cidr, endIP, err := p.calculateCIDRAndEndIP(start, uint32(ipCount))
	if err != nil {
		return nil, fmt.Errorf("计算CIDR失败: %v", err)
	}

	segment := &storage.IPSegment{
		CIDR:      cidr,
		StartIP:   start,
		EndIP:     endIP,
		Country:   cc,
		Type:      recordType,
		Status:    status,
		Date:      date,
		Registry:  registry,
		IPCount:   uint32(ipCount),
		CreatedAt: time.Now(),
	}

	return segment, nil
}

// calculateCIDRAndEndIP 计算CIDR和结束IP
func (p *Parser) calculateCIDRAndEndIP(startIP string, ipCount uint32) (string, string, error) {
	ip := net.ParseIP(startIP)
	if ip == nil {
		return "", "", fmt.Errorf("无效的IP地址: %s", startIP)
	}

	// 转换为IPv4
	ipv4 := ip.To4()
	if ipv4 == nil {
		return "", "", fmt.Errorf("不是IPv4地址: %s", startIP)
	}

	// 计算子网掩码长度
	prefixLen := 32 - int(log2(ipCount))
	if prefixLen < 0 || prefixLen > 32 {
		return "", "", fmt.Errorf("无效的IP数量: %d", ipCount)
	}

	// 创建CIDR
	_, ipNet, err := net.ParseCIDR(fmt.Sprintf("%s/%d", startIP, prefixLen))
	if err != nil {
		return "", "", fmt.Errorf("创建CIDR失败: %v", err)
	}

	// 计算结束IP
	endIP := p.calculateEndIP(ipv4, ipCount)

	return ipNet.String(), endIP.String(), nil
}

// calculateEndIP 计算结束IP
func (p *Parser) calculateEndIP(startIP net.IP, ipCount uint32) net.IP {
	// 将IP转换为uint32
	ipInt := uint32(startIP[0])<<24 + uint32(startIP[1])<<16 + uint32(startIP[2])<<8 + uint32(startIP[3])
	
	// 加上IP数量减1（因为包含起始IP）
	endInt := ipInt + ipCount - 1
	
	// 转换回IP
	return net.IPv4(
		byte(endInt>>24),
		byte(endInt>>16),
		byte(endInt>>8),
		byte(endInt),
	)
}

// log2 计算以2为底的对数
func log2(n uint32) uint32 {
	if n == 0 {
		return 0
	}
	
	var result uint32
	for n > 1 {
		n >>= 1
		result++
	}
	return result
}

// SplitToCSegments 将IP段拆分为C段
func (p *Parser) SplitToCSegments(segments []*storage.IPSegment) []*storage.IPSegment {
	var cSegments []*storage.IPSegment

	for _, segment := range segments {
		// 解析CIDR
		_, ipNet, err := net.ParseCIDR(segment.CIDR)
		if err != nil {
			fmt.Printf("解析CIDR失败: %s, %v\n", segment.CIDR, err)
			continue
		}

		// 获取网络地址和掩码长度
		maskLen, _ := ipNet.Mask.Size()

		// 如果已经是/24或更小，直接添加
		if maskLen >= 24 {
			cSegments = append(cSegments, segment)
			continue
		}

		// 拆分为/24段
		subSegments := p.splitToC24(segment, ipNet)
		cSegments = append(cSegments, subSegments...)
	}

	fmt.Printf("拆分完成，共生成 %d 个C段\n", len(cSegments))
	return cSegments
}

// splitToC24 拆分为/24段
func (p *Parser) splitToC24(segment *storage.IPSegment, ipNet *net.IPNet) []*storage.IPSegment {
	var segments []*storage.IPSegment

	// 获取网络地址
	networkIP := ipNet.IP.To4()
	if networkIP == nil {
		return segments
	}

	// 获取掩码长度
	maskLen, _ := ipNet.Mask.Size()

	// 计算需要拆分的C段数量
	hostBits := 24 - maskLen
	if hostBits <= 0 {
		return segments
	}

	cSegmentCount := 1 << uint(hostBits)

	// 生成每个C段
	for i := 0; i < cSegmentCount; i++ {
		// 计算C段的网络地址
		cNetworkIP := make(net.IP, 4)
		copy(cNetworkIP, networkIP)
		
		// 修改第三个字节
		cNetworkIP[2] = networkIP[2] + byte(i)

		// 创建C段CIDR
		cCIDR := fmt.Sprintf("%s/24", cNetworkIP.String())

		// 计算结束IP (xxx.xxx.xxx.255)
		endIP := make(net.IP, 4)
		copy(endIP, cNetworkIP)
		endIP[3] = 255

		cSegment := &storage.IPSegment{
			CIDR:      cCIDR,
			StartIP:   cNetworkIP.String(),
			EndIP:     endIP.String(),
			Country:   segment.Country,
			Type:      segment.Type,
			Status:    segment.Status,
			Date:      segment.Date,
			Registry:  segment.Registry,
			IPCount:   256,
			CreatedAt: time.Now(),
		}

		segments = append(segments, cSegment)
	}

	return segments
}