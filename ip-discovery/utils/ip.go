package utils

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// IPUtils IP工具函数
type IPUtils struct{}

// NewIPUtils 创建IP工具实例
func NewIPUtils() *IPUtils {
	return &IPUtils{}
}

// IsValidIP 检查IP地址是否有效
func (u *IPUtils) IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsValidCIDR 检查CIDR是否有效
func (u *IPUtils) IsValidCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

// IPToInt 将IP地址转换为整数
func (u *IPUtils) IPToInt(ip string) (uint32, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return 0, fmt.Errorf("无效的IP地址: %s", ip)
	}

	ipv4 := parsedIP.To4()
	if ipv4 == nil {
		return 0, fmt.Errorf("不是IPv4地址: %s", ip)
	}

	return uint32(ipv4[0])<<24 + uint32(ipv4[1])<<16 + uint32(ipv4[2])<<8 + uint32(ipv4[3]), nil
}

// IntToIP 将整数转换为IP地址
func (u *IPUtils) IntToIP(ipInt uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ipInt>>24),
		byte(ipInt>>16),
		byte(ipInt>>8),
		byte(ipInt))
}

// GetNetworkInfo 获取网络信息
func (u *IPUtils) GetNetworkInfo(cidr string) (*NetworkInfo, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("解析CIDR失败: %v", err)
	}

	maskLen, bits := ipNet.Mask.Size()
	hostBits := bits - maskLen
	totalHosts := uint32(1) << uint(hostBits)
	usableHosts := totalHosts - 2 // 排除网络地址和广播地址

	// 计算网络地址
	networkIP := ipNet.IP.To4()
	
	// 计算广播地址
	broadcastInt, _ := u.IPToInt(networkIP.String())
	broadcastInt += totalHosts - 1
	broadcastIP := u.IntToIP(broadcastInt)

	// 计算第一个和最后一个可用IP
	firstUsableInt := broadcastInt - totalHosts + 2
	lastUsableInt := broadcastInt - 1
	firstUsableIP := u.IntToIP(firstUsableInt)
	lastUsableIP := u.IntToIP(lastUsableInt)

	return &NetworkInfo{
		CIDR:           cidr,
		NetworkIP:      networkIP.String(),
		BroadcastIP:    broadcastIP,
		FirstUsableIP:  firstUsableIP,
		LastUsableIP:   lastUsableIP,
		SubnetMask:     net.IP(ipNet.Mask).String(),
		MaskLength:     maskLen,
		TotalHosts:     totalHosts,
		UsableHosts:    usableHosts,
		HostBits:       hostBits,
	}, nil
}

// NetworkInfo 网络信息结构
type NetworkInfo struct {
	CIDR          string `json:"cidr"`
	NetworkIP     string `json:"network_ip"`
	BroadcastIP   string `json:"broadcast_ip"`
	FirstUsableIP string `json:"first_usable_ip"`
	LastUsableIP  string `json:"last_usable_ip"`
	SubnetMask    string `json:"subnet_mask"`
	MaskLength    int    `json:"mask_length"`
	TotalHosts    uint32 `json:"total_hosts"`
	UsableHosts   uint32 `json:"usable_hosts"`
	HostBits      int    `json:"host_bits"`
}

// SplitCIDR 将大的CIDR拆分为指定大小的子网
func (u *IPUtils) SplitCIDR(cidr string, targetMaskLen int) ([]string, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("解析CIDR失败: %v", err)
	}

	currentMaskLen, _ := ipNet.Mask.Size()
	
	// 检查目标掩码长度是否有效
	if targetMaskLen <= currentMaskLen {
		return []string{cidr}, nil
	}

	if targetMaskLen > 32 {
		return nil, fmt.Errorf("目标掩码长度不能超过32")
	}

	// 计算子网数量
	subnetBits := targetMaskLen - currentMaskLen
	subnetCount := 1 << uint(subnetBits)

	// 获取网络地址
	networkIP := ipNet.IP.To4()
	if networkIP == nil {
		return nil, fmt.Errorf("不是IPv4网络")
	}

	var subnets []string
	
	// 计算每个子网的大小
	hostBits := 32 - targetMaskLen
	subnetSize := uint32(1) << uint(hostBits)

	// 生成子网列表
	baseInt, _ := u.IPToInt(networkIP.String())
	
	for i := 0; i < subnetCount; i++ {
		subnetInt := baseInt + uint32(i)*subnetSize
		subnetIP := u.IntToIP(subnetInt)
		subnetCIDR := fmt.Sprintf("%s/%d", subnetIP, targetMaskLen)
		subnets = append(subnets, subnetCIDR)
	}

	return subnets, nil
}

// IsIPInCIDR 检查IP是否在CIDR范围内
func (u *IPUtils) IsIPInCIDR(ip, cidr string) (bool, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false, fmt.Errorf("无效的IP地址: %s", ip)
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, fmt.Errorf("无效的CIDR: %s", cidr)
	}

	return ipNet.Contains(parsedIP), nil
}

// GetRandomIPsFromCIDR 从CIDR中获取随机IP列表
func (u *IPUtils) GetRandomIPsFromCIDR(cidr string, count int) ([]string, error) {
	info, err := u.GetNetworkInfo(cidr)
	if err != nil {
		return nil, err
	}

	if count > int(info.UsableHosts) {
		count = int(info.UsableHosts)
	}

	var ips []string
	
	// 简单实现：均匀分布选择IP
	step := info.UsableHosts / uint32(count)
	if step == 0 {
		step = 1
	}

	firstInt, _ := u.IPToInt(info.FirstUsableIP)
	
	for i := 0; i < count; i++ {
		ipInt := firstInt + uint32(i)*step
		if ipInt > firstInt+info.UsableHosts-1 {
			break
		}
		ips = append(ips, u.IntToIP(ipInt))
	}

	return ips, nil
}

// ParseIPRange 解析IP范围（如 192.168.1.1-192.168.1.100）
func (u *IPUtils) ParseIPRange(ipRange string) ([]string, error) {
	parts := strings.Split(ipRange, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("无效的IP范围格式: %s", ipRange)
	}

	startIP := strings.TrimSpace(parts[0])
	endIP := strings.TrimSpace(parts[1])

	startInt, err := u.IPToInt(startIP)
	if err != nil {
		return nil, fmt.Errorf("无效的起始IP: %v", err)
	}

	endInt, err := u.IPToInt(endIP)
	if err != nil {
		return nil, fmt.Errorf("无效的结束IP: %v", err)
	}

	if startInt > endInt {
		return nil, fmt.Errorf("起始IP不能大于结束IP")
	}

	var ips []string
	for i := startInt; i <= endInt; i++ {
		ips = append(ips, u.IntToIP(i))
	}

	return ips, nil
}

// FormatBytes 格式化字节数
func (u *IPUtils) FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}