
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

// TaiwanIPRange 台湾IP范围结构
type TaiwanIPRange struct {
	Network string
	Country string
	Region  string
}

func main() {
	fmt.Println("获取台湾IP段程序启动...")
	
	// 创建输出文件
	file, err := os.Create("taiwan_ip_ranges.txt")
	if err != nil {
		fmt.Printf("创建文件失败: %v\n", err)
		return
	}
	defer file.Close()
	
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	
	// 获取台湾IP段
	taiwanRanges := getTaiwanIPRanges()
	
	if len(taiwanRanges) > 0 {
		fmt.Printf("获取到 %d 个台湾IP段:\n", len(taiwanRanges))
		for _, ipRange := range taiwanRanges {
			fmt.Printf("网络段: %s, 地区: %s\n", ipRange.Network, ipRange.Region)
			writer.WriteString(fmt.Sprintf("%s\t%s\t%s\n", ipRange.Network, ipRange.Country, ipRange.Region))
		}
	} else {
		fmt.Println("未获取到台湾IP段")
	}
	
	// 验证示例IP地址
	fmt.Println("\n验证示例IP地址:")
	testIPs := []string{"210.242.0.0", "210.242.255.255", "1.160.0.0", "1.179.255.255"}
	for _, ip := range testIPs {
		if isTaiwanIP(ip) {
			fmt.Printf("IP %s 是台湾IP\n", ip)
		} else {
			fmt.Printf("IP %s 不是台湾IP\n", ip)
		}
	}
	
	fmt.Println("\n台湾IP段已保存到 taiwan_ip_ranges.txt 文件中")
}

// getTaiwanIPRanges 获取台湾IP段列表
func getTaiwanIPRanges() []TaiwanIPRange {
	// 台湾常见的IP段
	taiwanNetworks := []string{
		"1.160.0.0/11",
		"210.242.0.0/16",
		"210.243.0.0/16",
		"211.23.0.0/16",
		"211.75.0.0/16",
		"218.32.0.0/12",
		"220.130.0.0/16",
		"220.131.0.0/16",
	}
	
	var ranges []TaiwanIPRange
	for _, network := range taiwanNetworks {
		ranges = append(ranges, TaiwanIPRange{
			Network: network,
			Country: "Taiwan",
			Region:  "台湾地区",
		})
	}
	
	return ranges
}

// isTaiwanIP 验证IP是否为台湾IP
func isTaiwanIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	
	// 台湾IP范围检查
	taiwanRanges := []string{
		"1.160.0.0/11",
		"210.242.0.0/16",
		"211.23.0.0/16",
		"218.32.0.0/12",
		"220.130.0.0/15",
	}
	
	for _, cidr := range taiwanRanges {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if ipNet.Contains(ip) {
			return true
		}
	}
	
	return false
}
