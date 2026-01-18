package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// IPSegment 定义IP段结构体
type IPSegment struct {
	CIDR      string    `json:"cidr"`       // IP段（CIDR格式）
	Country   string    `json:"country"`    // 国家/地区
	Region    string    `json:"region"`     // 省/州
	City      string    `json:"city"`       // 城市
	ISP       string    `json:"isp"`        // ISP
	ASN       uint      `json:"asn"`        // ASN编号
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
}

// IPSegmentManager IP段管理器
type IPSegmentManager struct {
	segments     []*IPSegment
	lastUpdated  time.Time
	cacheTimeout time.Duration
}

// NewIPSegmentManager 创建新的IP段管理器
func NewIPSegmentManager(cacheTimeout time.Duration) *IPSegmentManager {
	return &IPSegmentManager{
		segments:     make([]*IPSegment, 0),
		lastUpdated:  time.Time{},
		cacheTimeout: cacheTimeout,
	}
}

// GetTaiwanIPSegments 获取台湾省IP段列表
func (m *IPSegmentManager) GetTaiwanIPSegments() ([]*IPSegment, error) {
	// 检查缓存是否有效
	if time.Since(m.lastUpdated) < m.cacheTimeout && len(m.segments) > 0 {
		log.Printf("使用缓存的台湾省IP段，共 %d 条", len(m.segments))
		return m.segments, nil
	}

	// 从APNIC获取真实台湾IP段
	segments, err := m.fetchTaiwanIPSegmentsFromAPNIC()
	if err != nil {
		return nil, fmt.Errorf("从APNIC获取台湾IP段失败: %v", err)
	}

	// 将大IP段拆分为多个C段（/24）
	splitSegments := m.splitSegmentsIntoCSegments(segments)

	// 更新缓存
	m.segments = splitSegments
	m.lastUpdated = time.Now()

	log.Printf("获取台湾省IP段成功，共 %d 条，拆分后共 %d 个C段", len(segments), len(splitSegments))
	return splitSegments, nil
}

// splitSegmentsIntoCSegments 将IP段拆分为多个C段（/24）
func (m *IPSegmentManager) splitSegmentsIntoCSegments(segments []*IPSegment) []*IPSegment {
	cSegments := make([]*IPSegment, 0)

	for _, segment := range segments {
		// 解析CIDR
		ip, ipnet, err := net.ParseCIDR(segment.CIDR)
		if err != nil {
			log.Printf("解析CIDR %s 失败，跳过拆分: %v", segment.CIDR, err)
			cSegments = append(cSegments, segment)
			continue
		}

		// 获取CIDR的前缀长度
		ones, _ := ipnet.Mask.Size()

		// 如果已经是C段或更小子网，直接添加
		if ones >= 24 {
			cSegments = append(cSegments, segment)
			continue
		}

		// 否则，拆分为多个C段
		log.Printf("将IP段 %s（/ %d）拆分为C段...", segment.CIDR, ones)

		// 计算需要拆分的C段数量
		// 例如：/16 -> 256个C段，/17 -> 128个C段
		splitCount := 1 << (24 - ones)

		// 复制原始IP，用于递增
		currentIP := make(net.IP, len(ip))
		copy(currentIP, ip)

		// 拆分IP段为C段
		for i := 0; i < splitCount; i++ {
			// 获取IPv4地址的四个字节（Go中IPv4地址以16字节IPv6兼容格式存储）
			// IPv4地址存储在第12-15字节
			ipBytes := currentIP.To4()
			if ipBytes == nil {
				log.Printf("无效的IPv4地址: %s，跳过", currentIP.String())
				break
			}

			// 创建新的C段网络地址
			cSegmentIP := net.IP{ipBytes[0], ipBytes[1], ipBytes[2], 0}
			cSegmentCIDR := cSegmentIP.String() + "/24"

			// 创建新的IP段对象
			cSegment := &IPSegment{
				CIDR:      cSegmentCIDR,
				Country:   segment.Country,
				Region:    segment.Region,
				City:      segment.City,
				ISP:       segment.ISP,
				ASN:       segment.ASN,
				UpdatedAt: segment.UpdatedAt,
			}

			cSegments = append(cSegments, cSegment)

			// 递增IP到下一个C段
			// 将第三个字节加1，超过255则进位
			ipBytes[2]++
			if ipBytes[2] == 0 {
				ipBytes[1]++
				if ipBytes[1] == 0 {
					ipBytes[0]++
				}
			}

			// 将更新后的IPv4字节复制回原始IP（IPv6兼容格式）
			copy(currentIP[12:16], ipBytes[:4])
		}
	}

	return cSegments
}

// loadIPSegmentsFromFile 从文件加载IP段
func (m *IPSegmentManager) loadIPSegmentsFromFile(filePath string) ([]*IPSegment, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	segments := make([]*IPSegment, 0)
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析行数据
		fields := strings.Split(line, ",")
		if len(fields) != 6 {
			log.Printf("第 %d 行格式错误，跳过: %s", lineNum, line)
			continue
		}

		// 解析ASN
		asn, err := strconv.ParseUint(fields[5], 10, 32)
		if err != nil {
			log.Printf("第 %d 行ASN格式错误，跳过: %s", lineNum, line)
			continue
		}

		// 创建IP段对象
		segment := &IPSegment{
			CIDR:      fields[0],
			Country:   fields[1],
			Region:    fields[2],
			City:      fields[3],
			ISP:       fields[4],
			ASN:       uint(asn),
			UpdatedAt: time.Now(),
		}

		segments = append(segments, segment)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	return segments, nil
}

// fetchTaiwanIPSegmentsFromAPNIC 从本地APNIC文件获取台湾省IP段
func (m *IPSegmentManager) fetchTaiwanIPSegmentsFromAPNIC() ([]*IPSegment, error) {
	// 使用本地的delegated-apnic-latest文件
	localFile := "delegated-apnic-latest"
	log.Printf("使用本地APNIC文件: %s", localFile)

	// 打开本地文件
	file, err := os.Open(localFile)
	if err != nil {
		return nil, fmt.Errorf("打开本地APNIC文件失败: %v", err)
	}
	defer file.Close()

	// 解析文件内容
	return m.parseAPNICFile(file)
}

// parseAPNICFile 解析APNIC文件内容，提取台湾省IP段
func (m *IPSegmentManager) parseAPNICFile(reader io.Reader) ([]*IPSegment, error) {
	// 解析内容
	segments := make([]*IPSegment, 0)
	bufReader := bufio.NewReader(reader)

	for {
		line, err := bufReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("读取内容失败: %v", err)
		}

		// 去除行首行尾空格
		line = strings.TrimSpace(line)

		// 跳过注释行和空行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析APNIC数据格式
		// 格式：registry,cc,type,start,value,date,status,extensions
		fields := strings.Split(line, "|")
		if len(fields) < 5 {
			continue
		}

		// 只处理APNIC注册的台湾(TW)IPv4地址段
		if fields[0] != "apnic" || fields[1] != "TW" || fields[2] != "ipv4" {
			continue
		}

		// 解析IP地址和掩码
		startIP := fields[3]
		value, err := strconv.Atoi(fields[4])
		if err != nil {
			continue
		}

		// 转换为CIDR格式
		// 计算子网掩码位数：32 - log2(value)
		cidr, err := m.ipRangeToCIDR(startIP, value)
		if err != nil {
			continue
		}

		// 创建IP段对象
		segment := &IPSegment{
			CIDR:      cidr,
			Country:   "Taiwan",
			Region:    "",
			City:      "",
			ISP:       "",
			ASN:       0,
			UpdatedAt: time.Now(),
		}

		segments = append(segments, segment)
	}

	// 如果没有获取到台湾IP段，返回错误
	if len(segments) == 0 {
		return nil, fmt.Errorf("未获取到台湾省IP段")
	}

	log.Printf("从APNIC成功获取 %d 条台湾省IP段", len(segments))
	return segments, nil
}

// ipRangeToCIDR 将IP范围转换为CIDR格式
func (m *IPSegmentManager) ipRangeToCIDR(startIP string, value int) (string, error) {
	// 计算CIDR前缀长度
	prefixLen := 32
	for i := value; i > 1; i /= 2 {
		prefixLen--
	}

	return fmt.Sprintf("%s/%d", startIP, prefixLen), nil
}

// ExampleGetIPSegments 获取IP段示例
func ExampleGetIPSegments() {
	// 创建IP段管理器
	manager := NewIPSegmentManager(1 * time.Hour)

	// 获取台湾省IP段
	segments, err := manager.GetTaiwanIPSegments()
	if err != nil {
		log.Fatalf("获取IP段失败: %v", err)
	}

	// 打印IP段信息
	fmt.Println("=== 台湾省IP段列表 ===")
	for i, segment := range segments {
		fmt.Printf("%d. CIDR: %s, Country: %s, Region: %s, City: %s, ISP: %s, ASN: %d\n",
			i+1, segment.CIDR, segment.Country, segment.Region, segment.City, segment.ISP, segment.ASN)
	}

	// 测试缓存功能
	fmt.Println("\n=== 测试缓存功能 ===")
	segments2, err := manager.GetTaiwanIPSegments()
	if err != nil {
		log.Fatalf("获取IP段失败: %v", err)
	}
	fmt.Printf("缓存获取IP段，共 %d 条\n", len(segments2))
}
