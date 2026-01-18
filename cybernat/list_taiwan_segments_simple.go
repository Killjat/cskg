package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

// SimpleSegment 简单的IP段结构体
type SimpleSegment struct {
	CIDR string
}

// ListTaiwanSegmentsSimple 列出所有台湾IP段的简单版本
func ListTaiwanSegmentsSimple() {
	log.Println("=== 列出所有台湾IP段 ===")

	// 打开本地APNIC文件
	localFile := "delegated-apnic-latest"
	file, err := os.Open(localFile)
	if err != nil {
		log.Fatalf("打开本地APNIC文件失败: %v", err)
	}
	defer file.Close()

	// 解析文件内容，提取台湾IP段
	segments, err := parseAPNICFileSimple(file)
	if err != nil {
		log.Fatalf("解析APNIC文件失败: %v", err)
	}

	log.Printf("从APNIC文件中提取到 %d 个台湾IP段", len(segments))

	// 打印所有IP段
	fmt.Println("\n=== 台湾IP段列表 ===")
	for i, segment := range segments {
		fmt.Printf("%d. %s\n", i+1, segment.CIDR)
	}

	fmt.Printf("\n共 %d 个台湾IP段\n", len(segments))
}

// parseAPNICFileSimple 简单解析APNIC文件内容，提取台湾省IP段
func parseAPNICFileSimple(reader io.Reader) ([]*SimpleSegment, error) {
	// 解析内容
	segments := make([]*SimpleSegment, 0)
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
		segment := &SimpleSegment{
			CIDR: cidr,
		}

		segments = append(segments, segment)
	}

	// 如果没有获取到台湾IP段，返回错误
	if len(segments) == 0 {
		return nil, fmt.Errorf("未获取到台湾省IP段")
	}

	return segments, nil
}
