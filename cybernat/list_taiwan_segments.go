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

// ListTaiwanSegments 列出所有台湾IP段
func ListTaiwanSegments() {
	log.Println("=== 列出所有台湾IP段 ===")

	// 打开本地APNIC文件
	localFile := "delegated-apnic-latest"
	file, err := os.Open(localFile)
	if err != nil {
		log.Fatalf("打开本地APNIC文件失败: %v", err)
	}
	defer file.Close()

	// 初始化IP段管理器
	manager := NewIPSegmentManager(1 * time.Hour)

	// 获取台湾IP段
	segments, err := manager.GetTaiwanIPSegments()
	if err != nil {
		log.Fatalf("获取台湾IP段失败: %v", err)
	}

	log.Printf("拆分后共 %d 个C段", len(segments))

	// 打印所有C段
	fmt.Println("\n=== 台湾IP段列表 ===")
	for i, segment := range segments {
		fmt.Printf("%d. %s\n", i+1, segment.CIDR)
	}
}

// parseAPNICFile 解析APNIC文件内容，提取台湾省IP段
func parseAPNICFile(reader io.Reader) ([]*IPSegment, error) {
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

		// 计算CIDR前缀长度
		prefixLen := 32
		for i := value; i > 1; i /= 2 {
			prefixLen--
		}

		cidr := fmt.Sprintf("%s/%d", startIP, prefixLen)

		// 创建IP段对象
		segment := &IPSegment{
			CIDR:    cidr,
			Country: "Taiwan",
			Region:  "",
			City:    "",
			ISP:     "",
			ASN:     0,
		}

		segments = append(segments, segment)
	}

	// 如果没有获取到台湾IP段，返回错误
	if len(segments) == 0 {
		return nil, fmt.Errorf("未获取到台湾省IP段")
	}

	return segments, nil
}

// splitSegmentsIntoCSegments 将IP段拆分为多个C段（/24）
func splitSegmentsIntoCSegments(segments []*IPSegment) []*IPSegment {
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
		// 计算需要拆分的C段数量
		splitCount := 1 << (24 - ones)

		// 复制原始IP，用于递增
		currentIP := make(net.IP, len(ip))
		copy(currentIP, ip)

		// 拆分IP段为C段
		for i := 0; i < splitCount; i++ {
			// 创建新的IP段对象
			cSegment := &IPSegment{
				CIDR:    fmt.Sprintf("%d.%d.%d.0/24", currentIP[12], currentIP[13], currentIP[14]),
				Country: segment.Country,
				Region:  segment.Region,
				City:    segment.City,
				ISP:     segment.ISP,
				ASN:     segment.ASN,
			}

			cSegments = append(cSegments, cSegment)

			// 递增IP到下一个C段
			currentIP[14]++
			if currentIP[14] == 0 {
				currentIP[13]++
				if currentIP[13] == 0 {
					currentIP[12]++
				}
			}
		}
	}

	return cSegments
}
