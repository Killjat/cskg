package device

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DeviceFingerprint 设备指纹结构体
type DeviceFingerprint struct {
	// 设备ID
	ID string `json:"id"`
	// 客户端IP
	ClientIP string `json:"client_ip"`
	// 客户端端口
	ClientPort int `json:"client_port"`
	// 服务器IP
	ServerIP string `json:"server_ip"`
	// 服务器端口
	ServerPort int `json:"server_port"`
	// 协议类型
	Protocol string `json:"protocol"`
	// 设备信息
	DeviceInfo DeviceInfo `json:"device_info"`
	// 指纹识别分数
	Score float64 `json:"score"`
	// 首次发现时间
	FirstSeen time.Time `json:"first_seen"`
	// 最后发现时间
	LastSeen time.Time `json:"last_seen"`
	// 连接次数
	ConnectionCount int `json:"connection_count"`
}

// DeviceInfo 设备详细信息
type DeviceInfo struct {
	// 操作系统猜测
	OS string `json:"os"`
	// 操作系统版本
	OSVersion string `json:"os_version"`
	// 设备类型
	DeviceType string `json:"device_type"`
	// 浏览器信息
	Browser string `json:"browser"`
	// 浏览器版本
	BrowserVersion string `json:"browser_version"`
	// 设备型号
	DeviceModel string `json:"device_model"`
	// 制造商
	Manufacturer string `json:"manufacturer"`
	// JA3指纹
	JA3Fingerprint string `json:"ja3_fingerprint"`
	// TCP窗口缩放
	TCPWindowScale int `json:"tcp_window_scale"`
	// TLS扩展
	TLSExtensions []string `json:"tls_extensions"`
}

// FingerprintManager 设备指纹管理器
type FingerprintManager struct {
	// 指纹数据库路径
	dbPath string
	// 启用用户代理分析
	userAgentAnalysis bool
	// 启用JA3指纹识别
	ja3Fingerprinting bool
	// 启用TCP窗口缩放分析
	tcpWindowScaling bool
	// 启用TLS扩展分析
	tlsExtensions bool
	// 指纹存储
	fingerprints map[string]*DeviceFingerprint
	// 互斥锁
	mutex sync.RWMutex
}

// NewFingerprintManager 创建新的指纹管理器
func NewFingerprintManager(dbPath string, userAgentAnalysis, ja3Fingerprinting, tcpWindowScaling, tlsExtensions bool) *FingerprintManager {
	// 创建数据目录
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	return &FingerprintManager{
		dbPath:             dbPath,
		userAgentAnalysis:  userAgentAnalysis,
		ja3Fingerprinting:  ja3Fingerprinting,
		tcpWindowScaling:   tcpWindowScaling,
		tlsExtensions:      tlsExtensions,
		fingerprints:       make(map[string]*DeviceFingerprint),
	}
}

// GenerateID 生成设备指纹ID
func (fm *FingerprintManager) GenerateID(clientIP string, clientPort int, serverIP string, serverPort int, protocol string) string {
	// 生成唯一ID
	data := fmt.Sprintf("%s:%d:%s:%d:%s", clientIP, clientPort, serverIP, serverPort, protocol)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// CaptureFingerprint 捕获设备指纹
func (fm *FingerprintManager) CaptureFingerprint(clientIP string, clientPort int, serverIP string, serverPort int, protocol string, rawData []byte) *DeviceFingerprint {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	// 生成指纹ID
	fingerprintID := fm.GenerateID(clientIP, clientPort, serverIP, serverPort, protocol)

	// 检查是否已存在指纹
	if fingerprint, exists := fm.fingerprints[fingerprintID]; exists {
		// 更新最后发现时间和连接次数
		fingerprint.LastSeen = time.Now()
		fingerprint.ConnectionCount++
		return fingerprint
	}

	// 创建新的指纹
	fingerprint := &DeviceFingerprint{
		ID:              fingerprintID,
		ClientIP:        clientIP,
		ClientPort:      clientPort,
		ServerIP:        serverIP,
		ServerPort:      serverPort,
		Protocol:        protocol,
		DeviceInfo:      fm.analyzeDeviceInfo(rawData),
		Score:           0.85, // 默认分数
		FirstSeen:       time.Now(),
		LastSeen:        time.Now(),
		ConnectionCount: 1,
	}

	// 存储指纹
	fm.fingerprints[fingerprintID] = fingerprint

	return fingerprint
}

// analyzeDeviceInfo 分析设备信息
func (fm *FingerprintManager) analyzeDeviceInfo(rawData []byte) DeviceInfo {
	deviceInfo := DeviceInfo{
		OS:               "Unknown",
		OSVersion:        "Unknown",
		DeviceType:       "Unknown",
		Browser:          "Unknown",
		BrowserVersion:   "Unknown",
		DeviceModel:      "Unknown",
		Manufacturer:     "Unknown",
		JA3Fingerprint:   "",
		TCPWindowScale:   0,
		TLSExtensions:    []string{},
	}

	// 简单的设备信息分析示例
	dataStr := string(rawData)

	// 根据协议特征识别设备
	if len(dataStr) > 0 {
		// Modbus设备识别
		if strings.Contains(dataStr, "Modbus") || strings.HasPrefix(dataStr, "\x00") {
			deviceInfo.DeviceType = "Industrial Controller"
			deviceInfo.Manufacturer = "Siemens"
			deviceInfo.DeviceModel = "S7-1200"
			deviceInfo.OS = "RTOS"
			deviceInfo.OSVersion = "Unknown"
		} else if strings.HasPrefix(dataStr, "\x50") {
			// MySQL设备识别
			deviceInfo.DeviceType = "Database Client"
			deviceInfo.Manufacturer = "Oracle"
			deviceInfo.DeviceModel = "MySQL Client"
			deviceInfo.OS = "Linux"
			deviceInfo.OSVersion = "Unknown"
		} else if strings.HasPrefix(dataStr, "*") || strings.HasPrefix(dataStr, "$") {
			// Redis设备识别
			deviceInfo.DeviceType = "Cache Client"
			deviceInfo.Manufacturer = "Redis Labs"
			deviceInfo.DeviceModel = "Redis Client"
			deviceInfo.OS = "Linux"
			deviceInfo.OSVersion = "Unknown"
		} else if len(dataStr) > 4 {
			// Kafka设备识别
			deviceInfo.DeviceType = "Message Client"
			deviceInfo.Manufacturer = "Apache"
			deviceInfo.DeviceModel = "Kafka Client"
			deviceInfo.OS = "Linux"
			deviceInfo.OSVersion = "Unknown"
		} else {
			// 其他设备
			deviceInfo.DeviceType = "Unknown Device"
			deviceInfo.Manufacturer = "Unknown"
			deviceInfo.DeviceModel = "Unknown"
		}
	}

	return deviceInfo
}

// GetFingerprint 获取设备指纹
func (fm *FingerprintManager) GetFingerprint(id string) (*DeviceFingerprint, bool) {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	fingerprint, exists := fm.fingerprints[id]
	return fingerprint, exists
}

// GetAllFingerprints 获取所有设备指纹
func (fm *FingerprintManager) GetAllFingerprints() []*DeviceFingerprint {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	fingerprints := make([]*DeviceFingerprint, 0, len(fm.fingerprints))
	for _, fp := range fm.fingerprints {
		fingerprints = append(fingerprints, fp)
	}

	return fingerprints
}

// Close 关闭指纹管理器
func (fm *FingerprintManager) Close() {
	// 保存指纹数据到文件
	// 实际应用中可以使用数据库存储
}
