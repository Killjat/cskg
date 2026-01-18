package packet

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"

	"honeypot-system/internal/device"
	"honeypot-system/internal/logger"
	"honeypot-system/internal/session"
)

// PacketCapture 数据包捕获结构体
type PacketCapture struct {
	// 监听的网络接口
	interfaces []string
	// 监听的端口
	ports []int
	// 全流量捕获开关
	fullCapture bool
	// 流量保存路径
	savePath string
	// 设备指纹管理器
	fingerprintManager *device.FingerprintManager
	// 日志记录器
	logger logger.Logger
	// 捕获句柄
	handles []*pcap.Handle
	// 停止信号
	stopChan chan struct{}
	// 互斥锁
	mutex sync.Mutex
	// 是否运行中
	running bool
	// 会话管理器
	sessionManager interface {
		CreateSession(srcIP string, srcPort int, dstIP string, dstPort int, protocol string) *session.SessionInfo
		AddPacket(sessionID string, packet session.PacketInfo) error
		GetAllSessions() []*session.SessionInfo
	}
}

// NewPacketCapture 创建新的数据包捕获实例
func NewPacketCapture(interfaces []string, ports []int, fullCapture bool, savePath string, fingerprintManager *device.FingerprintManager, sessionManager interface {
	CreateSession(srcIP string, srcPort int, dstIP string, dstPort int, protocol string) *session.SessionInfo
	AddPacket(sessionID string, packet session.PacketInfo) error
	GetAllSessions() []*session.SessionInfo
}, logger logger.Logger) *PacketCapture {
	// 创建流量保存目录
	os.MkdirAll(savePath, 0755)

	return &PacketCapture{
		interfaces:         interfaces,
		ports:              ports,
		fullCapture:        fullCapture,
		savePath:           savePath,
		fingerprintManager: fingerprintManager,
		sessionManager:     sessionManager,
		logger:             logger,
		stopChan:           make(chan struct{}),
		running:            false,
	}
}

// Start 启动数据包捕获
func (pc *PacketCapture) Start() error {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	if pc.running {
		return nil
	}

	// 打开网络接口
	for _, iface := range pc.interfaces {
		// 特殊处理 "any" 接口
		device := iface
		if iface == "any" {
			device = ""
		}

		// 打开网络接口进行捕获
		handle, err := pcap.OpenLive(device, 1600, true, pcap.BlockForever)
		if err != nil {
			pc.logger.Error(fmt.Sprintf("Failed to open interface %s: %v", iface, err))
			continue
		}

		// 设置BPF过滤器
		portFilter := ""
		for i, port := range pc.ports {
			if i > 0 {
				portFilter += " or "
			}
			portFilter += fmt.Sprintf("tcp port %d or udp port %d", port, port)
		}

		if err := handle.SetBPFFilter(portFilter); err != nil {
			pc.logger.Error(fmt.Sprintf("Failed to set BPF filter for interface %s: %v", iface, err))
			handle.Close()
			continue
		}

		pc.handles = append(pc.handles, handle)
	}

	if len(pc.handles) == 0 {
		return fmt.Errorf("failed to open any network interface")
	}

	pc.running = true

	// 启动捕获协程
	for i, handle := range pc.handles {
		go pc.capturePackets(i, handle)
	}

	pc.logger.Info("Packet capture started successfully")
	return nil
}

// Stop 停止数据包捕获
func (pc *PacketCapture) Stop() {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	if !pc.running {
		return
	}

	// 发送停止信号
	close(pc.stopChan)

	// 关闭所有捕获句柄
	for _, handle := range pc.handles {
		handle.Close()
	}

	pc.handles = nil
	pc.running = false

	pc.logger.Info("Packet capture stopped successfully")
}

// capturePackets 捕获数据包
func (pc *PacketCapture) capturePackets(interfaceIndex int, handle *pcap.Handle) {
	// 创建数据包源
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := packetSource.Packets()

	// 创建PCAP文件用于保存流量
	var pcapFile *os.File
	var pcapWriter *pcapgo.Writer

	if pc.fullCapture {
		pcapFileName := filepath.Join(pc.savePath, fmt.Sprintf("capture_%s_%d.pcap", time.Now().Format("2006-01-02_15-04-05"), interfaceIndex))
		var err error
		pcapFile, err = os.Create(pcapFileName)
		if err != nil {
			pc.logger.Error(fmt.Sprintf("Failed to create pcap file: %v", err))
			pcapFile = nil
		} else {
			pcapWriter = pcapgo.NewWriter(pcapFile)
			pcapWriter.WriteFileHeader(1600, handle.LinkType())
			pc.logger.Info(fmt.Sprintf("Started saving packets to %s", pcapFileName))
		}
	}

	defer func() {
		if pcapFile != nil {
			pcapFile.Close()
		}
	}()

	for {
		select {
		case packet := <-packets:
			if packet == nil {
				continue
			}

			// 保存数据包到PCAP文件
			if pc.fullCapture && pcapWriter != nil {
				pcapWriter.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
			}

			// 解析数据包
			pc.analyzePacket(packet)

		case <-pc.stopChan:
			pc.logger.Info(fmt.Sprintf("Stopping packet capture on interface %d", interfaceIndex))
			return
		}
	}
}

// analyzePacket 分析数据包
func (pc *PacketCapture) analyzePacket(packet gopacket.Packet) {
	// 解析以太网层
	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethernetLayer == nil {
		return
	}

	// 解析IP层
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		ipLayer = packet.Layer(layers.LayerTypeIPv6)
		if ipLayer == nil {
			return
		}
	}

	// 解析TCP或UDP层
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	var transportLayer gopacket.Layer
	var protocol string

	if tcpLayer != nil {
		transportLayer = tcpLayer
		protocol = "tcp"
	} else {
		udpLayer := packet.Layer(layers.LayerTypeUDP)
		if udpLayer != nil {
			transportLayer = udpLayer
			protocol = "udp"
		} else {
			return
		}
	}

	// 提取IP和端口信息
	var srcIP, dstIP string
	var srcPort, dstPort int

	switch ip := ipLayer.(type) {
	case *layers.IPv4:
		srcIP = ip.SrcIP.String()
		dstIP = ip.DstIP.String()
	case *layers.IPv6:
		srcIP = ip.SrcIP.String()
		dstIP = ip.DstIP.String()
	}

	switch transport := transportLayer.(type) {
	case *layers.TCP:
		srcPort = int(transport.SrcPort)
		dstPort = int(transport.DstPort)
	case *layers.UDP:
		srcPort = int(transport.SrcPort)
		dstPort = int(transport.DstPort)
	}

	// 检查是否是我们监听的端口
	isListeningPort := false
	for _, port := range pc.ports {
		if dstPort == port {
			isListeningPort = true
			break
		}
	}

	if !isListeningPort {
		return
	}

	// 获取应用层数据
	applicationLayer := packet.ApplicationLayer()
	var rawData []byte
	if applicationLayer != nil {
		rawData = applicationLayer.Payload()
	}

	// 捕获设备指纹
	pc.fingerprintManager.CaptureFingerprint(srcIP, srcPort, dstIP, dstPort, protocol, rawData)

	// 记录日志
	pc.logger.Debug(fmt.Sprintf("Captured packet: %s:%d -> %s:%d (%s), length: %d", srcIP, srcPort, dstIP, dstPort, protocol, len(packet.Data())))

	// 创建或获取会话
	sessionInfo := pc.sessionManager.CreateSession(srcIP, srcPort, dstIP, dstPort, protocol)

	// 创建数据包信息
	packetType := "request"
	if srcIP == dstIP { // 简单判断，实际应根据协议逻辑判断
		packetType = "response"
	}

	// 解析协议内容
	protocolInfo := make(map[string]interface{})
	pc.parseProtocolWithInfo(packet, srcIP, srcPort, dstIP, dstPort, protocol, rawData, protocolInfo)

	// 创建数据包对象
	packetInfo := session.PacketInfo{
		ID:             fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:           packetType,
		Timestamp:      time.Now(),
		Length:         len(packet.Data()),
		Content:        protocolInfo,
		RawData:        hex.EncodeToString(packet.Data()),
		ProtocolInfo:   protocolInfo,
	}

	// 添加数据包到会话
	if err := pc.sessionManager.AddPacket(sessionInfo.ID, packetInfo); err != nil {
		pc.logger.Error(fmt.Sprintf("Failed to add packet to session: %v", err))
	}

	// 更新会话的设备指纹ID
	// 这里简化处理，实际应在会话创建时或通过更新函数设置

}

// parseProtocolWithInfo 解析具体协议并返回解析结果
func (pc *PacketCapture) parseProtocolWithInfo(packet gopacket.Packet, srcIP string, srcPort int, dstIP string, dstPort int, protocol string, rawData []byte, protocolInfo map[string]interface{}) {
	// 根据端口和协议类型进行解析
	switch dstPort {
	case 502:
		// Modbus协议解析
		pc.parseModbus(packet, srcIP, srcPort, dstIP, dstPort, rawData, protocolInfo)
	case 3306:
		// MySQL协议解析
		pc.parseMySQL(packet, srcIP, srcPort, dstIP, dstPort, rawData, protocolInfo)
	case 6379:
		// Redis协议解析
		pc.parseRedis(packet, srcIP, srcPort, dstIP, dstPort, rawData, protocolInfo)
	case 9092:
		// Kafka协议解析
		pc.parseKafka(packet, srcIP, srcPort, dstIP, dstPort, rawData, protocolInfo)
	default:
		// 其他协议，记录基本信息
		protocolInfo["protocol"] = "unknown"
		protocolInfo["port"] = dstPort
		pc.logger.Debug(fmt.Sprintf("Unknown protocol on port %d", dstPort))
	}
}

// parseModbus 解析Modbus协议
func (pc *PacketCapture) parseModbus(packet gopacket.Packet, srcIP string, srcPort int, dstIP string, dstPort int, rawData []byte, protocolInfo map[string]interface{}) {
	// 简单的Modbus协议解析示例
	protocolInfo["protocol"] = "modbus"
	protocolInfo["type"] = "tcp"

	if len(rawData) < 7 {
		protocolInfo["error"] = "invalid packet length"
		return
	}

	// Modbus TCP头长度为7字节
	transactionID := uint16(rawData[0])<<8 | uint16(rawData[1])
	protocolID := uint16(rawData[2])<<8 | uint16(rawData[3])
	length := uint16(rawData[4])<<8 | uint16(rawData[5])
	unitID := rawData[6]

	protocolInfo["transaction_id"] = transactionID
	protocolInfo["protocol_id"] = protocolID
	protocolInfo["length"] = length
	protocolInfo["unit_id"] = unitID

	pc.logger.Debug(fmt.Sprintf("Modbus TCP: Transaction ID: %d, Protocol ID: %d, Length: %d, Unit ID: %d", transactionID, protocolID, length, unitID))
}

// parseMySQL 解析MySQL协议
func (pc *PacketCapture) parseMySQL(packet gopacket.Packet, srcIP string, srcPort int, dstIP string, dstPort int, rawData []byte, protocolInfo map[string]interface{}) {
	// 简单的MySQL协议解析示例
	protocolInfo["protocol"] = "mysql"

	if len(rawData) == 0 {
		protocolInfo["error"] = "invalid packet length"
		return
	}

	// MySQL包类型由第一个字节决定
	packetType := rawData[0]
	protocolInfo["packet_type"] = fmt.Sprintf("0x%02X", packetType)
	protocolInfo["length"] = len(rawData)

	pc.logger.Debug(fmt.Sprintf("MySQL Packet: Type: 0x%02X, Length: %d", packetType, len(rawData)))
}

// parseRedis 解析Redis协议
func (pc *PacketCapture) parseRedis(packet gopacket.Packet, srcIP string, srcPort int, dstIP string, dstPort int, rawData []byte, protocolInfo map[string]interface{}) {
	// 简单的Redis协议解析示例
	protocolInfo["protocol"] = "redis"

	if len(rawData) == 0 {
		protocolInfo["error"] = "invalid packet length"
		return
	}

	// Redis协议以特殊字符开头
	commandType := rawData[0]
	protocolInfo["command_type"] = string(commandType)
	protocolInfo["length"] = len(rawData)

	pc.logger.Debug(fmt.Sprintf("Redis Packet: Type: %c, Length: %d", commandType, len(rawData)))
}

// parseKafka 解析Kafka协议
func (pc *PacketCapture) parseKafka(packet gopacket.Packet, srcIP string, srcPort int, dstIP string, dstPort int, rawData []byte, protocolInfo map[string]interface{}) {
	// 简单的Kafka协议解析示例
	protocolInfo["protocol"] = "kafka"

	if len(rawData) < 4 {
		protocolInfo["error"] = "invalid packet length"
		return
	}

	// Kafka包长度为前4字节
	packetLength := uint32(rawData[0])<<24 | uint32(rawData[1])<<16 | uint32(rawData[2])<<8 | uint32(rawData[3])
	protocolInfo["length"] = packetLength

	pc.logger.Debug(fmt.Sprintf("Kafka Packet: Length: %d", packetLength))
}

// GetFingerprints 获取所有设备指纹
func (pc *PacketCapture) GetFingerprints() []*device.DeviceFingerprint {
	return pc.fingerprintManager.GetAllFingerprints()
}
