# 工业协议蜜罐系统程序设计文档

## 1. 设计概述

### 1.1 设计目标

本设计文档基于需求分析文档，旨在详细描述工业协议蜜罐系统的程序设计方案，包括系统架构、模块设计、数据库设计、接口设计等内容，为开发人员提供明确的开发指导，确保系统能够满足需求分析文档中提出的所有功能和非功能需求。

### 1.2 设计原则

1. **模块化设计**：系统采用模块化架构，各模块之间低耦合、高内聚，便于维护和扩展
2. **可扩展性**：支持通过插件方式添加新的协议支持和功能扩展
3. **安全性**：确保蜜罐系统本身的安全性，防止被用作攻击跳板
4. **性能优先**：优化系统性能，支持高并发连接和低延迟响应
5. **可靠性**：系统具有容错能力，确保7x24小时稳定运行
6. **易用性**：提供简洁的配置方式和友好的Web界面

### 1.3 技术栈选择

| 技术类别         | 技术选型               | 版本要求   | 选型理由                                                     |
|----------------|--------------------|--------|----------------------------------------------------------|
| 开发语言         | Go语言              | 1.21+  | 高性能、并发支持好、编译型语言，适合网络服务开发                               |
| Web框架          | Gin                | 1.9.1  | 轻量级、高性能、易用的Go Web框架，适合构建RESTful API和Web界面                    |
| 数据包捕获         | gopacket           | 1.1.19 | 纯Go实现的数据包捕获库，跨平台支持好，无需外部依赖（避免使用pcap减少系统依赖）               |
| 配置管理          | Viper              | 1.16.0 | 功能强大的配置管理库，支持多种配置格式，便于系统配置管理                              |
| 日志管理          | zap                | 1.24.0 | 高性能、结构化日志库，支持多种输出格式和日志级别                                      |
| 数据库           | SQLite             | 3.40+  | 轻量级、嵌入式数据库，无需独立部署，适合存储设备指纹和连接记录等结构化数据                     |
| 设备指纹识别        | 自定义算法+JA3指纹库     | -      | 结合自定义指纹识别算法和JA3 TLS指纹库，提高设备识别准确性                                |
| HTML模板引擎       | html/template      | 标准库    | Go标准库中的HTML模板引擎，安全、高效，适合构建Web界面                                   |
| 前端UI框架         | 原生HTML/CSS/JavaScript | -      | 减少外部依赖，提高系统安全性和加载速度                                        |

## 2. 系统架构设计

### 2.1 整体架构

系统采用分层架构设计，从下到上分为数据采集层、数据处理层、数据存储层和应用层，各层之间通过明确的接口进行通信，实现低耦合、高内聚的设计目标。

```
┌─────────────────────────────────────────────────────────────────┐
│                       应用层 (Web界面)                           │
│  - Web服务模块                                                │
│  - 实时数据展示                                                │
│  - 设备指纹管理                                                │
│  - 连接记录查询                                                │
│  - 会话流量分析                                                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       数据存储层                                │
│  - SQLite数据库                                                │
│  - 设备指纹存储                                                │
│  - 连接记录存储                                                │
│  - 会话流量存储                                                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       数据处理层                                │
│  - 设备指纹识别模块                                              │
│  - 连接记录管理模块                                              │
│  - 会话分析模块                                                │
│  - 异常检测模块                                                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       数据采集层                                │
│  - 数据包捕获模块                                              │
│  - 流量解析模块                                                │
│  - 协议识别模块                                                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       操作系统层                                │
│  - 网络接口                                                    │
│  - 系统资源                                                    │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 模块关系图

各模块之间的依赖关系如下所示：

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   配置管理模块     │────▶   日志管理模块     │────▶   Web服务模块    │
└─────────────────┘     └─────────────────┘     └─────────────────┘
        ▲                        ▲                        │
        │                        │                        ▼
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  数据包捕获模块    │────▶  设备指纹识别模块   │────▶  数据存储模块    │
└─────────────────┘     └─────────────────┘     └─────────────────┘
        ▲                        │                        │
        │                        ▼                        ▼
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   流量解析模块     │────▶   连接记录管理模块   │────▶   会话管理模块    │
└─────────────────┘     └─────────────────┘     └─────────────────┘
        ▲                        │
        │                        ▼
┌─────────────────┐     ┌─────────────────┐
│   协议识别模块     │────▶   异常检测模块     │
└─────────────────┘     └─────────────────┘
```

## 3. 模块详细设计

### 3.1 配置管理模块

#### 3.1.1 模块功能

- 读取和解析配置文件
- 提供配置信息的访问接口
- 支持配置热加载（可选）

#### 3.1.2 模块设计

```go
// config/config.go
package config

import (
    "github.com/spf13/viper"
)

// Config 系统配置结构体
type Config struct {
    Honeypot      HoneypotConfig      `mapstructure:"honeypot"`
    PacketCapture PacketCaptureConfig `mapstructure:"packet_capture"`
    DeviceFP      DeviceFPConfig      `mapstructure:"device_fingerprint"`
    Web           WebConfig           `mapstructure:"web"`
    Database      DatabaseConfig      `mapstructure:"database"`
}

// HoneypotConfig 蜜罐配置
type HoneypotConfig struct {
    Name     string `mapstructure:"name"`
    Version  string `mapstructure:"version"`
    LogPath  string `mapstructure:"log_path"`
    LogLevel string `mapstructure:"log_level"`
}

// PacketCaptureConfig 数据包捕获配置
type PacketCaptureConfig struct {
    Enabled      bool     `mapstructure:"enabled"`
    Interfaces   []string `mapstructure:"interfaces"`
    Ports        []int    `mapstructure:"ports"`
    FullCapture  bool     `mapstructure:"full_capture"`
    SavePath     string   `mapstructure:"save_path"`
}

// DeviceFPConfig 设备指纹配置
type DeviceFPConfig struct {
    Enabled bool          `mapstructure:"enabled"`
    DBPath  string        `mapstructure:"db_path"`
    Rules   FPRulesConfig `mapstructure:"rules"`
}

// FPRulesConfig 指纹识别规则配置
type FPRulesConfig struct {
    UserAgentAnalysis bool `mapstructure:"user_agent_analysis"`
    JA3Fingerprinting bool `mapstructure:"ja3_fingerprinting"`
    TCPWindowScaling  bool `mapstructure:"tcp_window_scaling"`
    TLSExtensions     bool `mapstructure:"tls_extensions"`
}

// WebConfig Web配置
type WebConfig struct {
    Enabled         bool `mapstructure:"enabled"`
    Host            string `mapstructure:"host"`
    Port            int    `mapstructure:"port"`
    HTTPS           bool   `mapstructure:"https"`
    SessionTimeout  int    `mapstructure:"session_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
    Type     string `mapstructure:"type"`
    Path     string `mapstructure:"path"`
    MaxOpenConns int `mapstructure:"max_open_conns"`
    MaxIdleConns int `mapstructure:"max_idle_conns"`
    ConnMaxLifetime int `mapstructure:"conn_max_lifetime"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(configPath)
    viper.AddConfigPath(".")
    viper.AddConfigPath("./config")
    
    // 设置默认值
    setDefaults()
    
    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }
    
    return &config, nil
}

// setDefaults 设置默认配置
func setDefaults() {
    viper.SetDefault("honeypot.name", "Industrial Protocol Honeypot")
    viper.SetDefault("honeypot.version", "1.0.0")
    viper.SetDefault("honeypot.log_path", "./logs")
    viper.SetDefault("honeypot.log_level", "info")
    
    viper.SetDefault("packet_capture.enabled", true)
    viper.SetDefault("packet_capture.interfaces", []string{"any"})
    viper.SetDefault("packet_capture.ports", []int{502, 3306, 6379, 9092, 80, 9999})
    viper.SetDefault("packet_capture.full_capture", true)
    viper.SetDefault("packet_capture.save_path", "./data/pcap")
    
    viper.SetDefault("device_fingerprint.enabled", true)
    viper.SetDefault("device_fingerprint.db_path", "./data/fingerprints.db")
    viper.SetDefault("device_fingerprint.rules.user_agent_analysis", true)
    viper.SetDefault("device_fingerprint.rules.ja3_fingerprinting", true)
    viper.SetDefault("device_fingerprint.rules.tcp_window_scaling", true)
    viper.SetDefault("device_fingerprint.rules.tls_extensions", true)
    
    viper.SetDefault("web.enabled", true)
    viper.SetDefault("web.host", "0.0.0.0")
    viper.SetDefault("web.port", 8080)
    viper.SetDefault("web.https", false)
    viper.SetDefault("web.session_timeout", 3600)
    
    viper.SetDefault("database.type", "sqlite")
    viper.SetDefault("database.path", "./data/honeypot.db")
    viper.SetDefault("database.max_open_conns", 10)
    viper.SetDefault("database.max_idle_conns", 5)
    viper.SetDefault("database.conn_max_lifetime", 3600)
}
```

### 3.2 日志管理模块

#### 3.2.1 模块功能

- 生成结构化日志
- 支持多种日志级别（debug, info, warn, error）
- 支持日志文件滚动和归档
- 支持日志输出到控制台和文件

#### 3.2.2 模块设计

```go
// internal/logger/logger.go
package logger

import (
    "os"
    "path/filepath"
    "time"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "gopkg.in/natefinch/lumberjack.v2"
)

// Logger 日志接口
type Logger interface {
    Debug(msg string, fields ...zap.Field)
    Info(msg string, fields ...zap.Field)
    Warn(msg string, fields ...zap.Field)
    Error(msg string, fields ...zap.Field)
    Fatal(msg string, fields ...zap.Field)
    Close() error
}

// logger 日志实现结构体
type logger struct {
    zapLogger *zap.Logger
    sugar     *zap.SugaredLogger
}

// NewLogger 创建新的日志实例
func NewLogger(logPath string, logLevel string) Logger {
    // 创建日志目录
    if err := os.MkdirAll(logPath, 0755); err != nil {
        panic(err)
    }

    // 配置日志文件滚动
    lumberjackLogger := &lumberjack.Logger{
        Filename:   filepath.Join(logPath, time.Now().Format("2006-01-02")+"_honeypot.log"),
        MaxSize:    100, // 100MB
        MaxBackups: 30,  // 保留30个备份
        MaxAge:     7,   // 保留7天
        Compress:   true, // 启用压缩
    }

    // 配置日志编码器
    encoderConfig := zap.NewProductionEncoderConfig()
    encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

    // 配置日志级别
    level := zapcore.InfoLevel
    switch logLevel {
    case "debug":
        level = zapcore.DebugLevel
    case "info":
        level = zapcore.InfoLevel
    case "warn":
        level = zapcore.WarnLevel
    case "error":
        level = zapcore.ErrorLevel
    }

    // 创建核心日志器
    core := zapcore.NewCore(
        zapcore.NewJSONEncoder(encoderConfig),
        zapcore.AddSync(lumberjackLogger),
        level,
    )

    // 添加控制台输出
    consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
    consoleCore := zapcore.NewCore(
        consoleEncoder,
        zapcore.AddSync(os.Stdout),
        level,
    )

    // 创建混合核心
    combinedCore := zapcore.NewTee(core, consoleCore)

    // 创建日志器
    zapLogger := zap.New(combinedCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
    sugar := zapLogger.Sugar()

    return &logger{
        zapLogger: zapLogger,
        sugar:     sugar,
    }
}

// Debug 输出debug级别日志
func (l *logger) Debug(msg string, fields ...zap.Field) {
    l.zapLogger.Debug(msg, fields...)
}

// Info 输出info级别日志
func (l *logger) Info(msg string, fields ...zap.Field) {
    l.zapLogger.Info(msg, fields...)
}

// Warn 输出warn级别日志
func (l *logger) Warn(msg string, fields ...zap.Field) {
    l.zapLogger.Warn(msg, fields...)
}

// Error 输出error级别日志
func (l *logger) Error(msg string, fields ...zap.Field) {
    l.zapLogger.Error(msg, fields...)
}

// Fatal 输出fatal级别日志
func (l *logger) Fatal(msg string, fields ...zap.Field) {
    l.zapLogger.Fatal(msg, fields...)
}

// Close 关闭日志器
func (l *logger) Close() error {
    return l.zapLogger.Sync()
}
```

### 3.3 数据包捕获模块

#### 3.3.1 模块功能

- 监听指定网络接口的流量
- 捕获指定端口的数据包
- 支持全流量捕获和存储
- 提供数据包解析和处理接口

#### 3.3.2 模块设计

```go
// internal/packet/capture.go
package packet

import (
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
)

// PacketCapture 数据包捕获结构体
type PacketCapture struct {
    interfaces         []string
    ports              []int
    fullCapture        bool
    savePath           string
    fingerprintManager *device.FingerprintManager
    logger             logger.Logger
    handles            []*pcap.Handle
    stopChan           chan struct{}
    mutex              sync.Mutex
    running            bool
}

// NewPacketCapture 创建新的数据包捕获实例
func NewPacketCapture(interfaces []string, ports []int, fullCapture bool, savePath string, fingerprintManager *device.FingerprintManager, logger logger.Logger) *PacketCapture {
    os.MkdirAll(savePath, 0755)

    return &PacketCapture{
        interfaces:         interfaces,
        ports:              ports,
        fullCapture:        fullCapture,
        savePath:           savePath,
        fingerprintManager: fingerprintManager,
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
        device := iface
        if iface == "any" {
            device = ""
        }

        handle, err := pcap.OpenLive(device, 1600, true, pcap.BlockForever)
        if err != nil {
            pc.logger.Warn(fmt.Sprintf("Failed to open interface %s: %v", iface, err))
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
            pc.logger.Warn(fmt.Sprintf("Failed to set BPF filter for interface %s: %v", iface, err))
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

    close(pc.stopChan)

    for _, handle := range pc.handles {
        handle.Close()
    }

    pc.handles = nil
    pc.running = false

    pc.logger.Info("Packet capture stopped successfully")
}

// capturePackets 捕获数据包
func (pc *PacketCapture) capturePackets(interfaceIndex int, handle *pcap.Handle) {
    packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
    packets := packetSource.Packets()

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

            if pc.fullCapture && pcapWriter != nil {
                pcapWriter.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
            }

            pc.analyzePacket(packet)

        case <-pc.stopChan:
            pc.logger.Info(fmt.Sprintf("Stopping packet capture on interface %d", interfaceIndex))
            return
        }
    }
}

// analyzePacket 分析数据包
func (pc *PacketCapture) analyzePacket(packet gopacket.Packet) {
    // 解析网络层
    var ipLayer gopacket.Layer
    if ipLayer = packet.Layer(layers.LayerTypeIPv4); ipLayer == nil {
        ipLayer = packet.Layer(layers.LayerTypeIPv6)
        if ipLayer == nil {
            return
        }
    }

    // 解析传输层
    var transportLayer gopacket.Layer
    var protocol string
    if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
        transportLayer = tcpLayer
        protocol = "tcp"
    } else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
        transportLayer = udpLayer
        protocol = "udp"
    } else {
        return
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

    // 提取应用层数据
    applicationLayer := packet.ApplicationLayer()
    var rawData []byte
    if applicationLayer != nil {
        rawData = applicationLayer.Payload()
    }

    // 捕获设备指纹
    pc.fingerprintManager.CaptureFingerprint(srcIP, srcPort, dstIP, dstPort, protocol, rawData)

    // 记录日志
    pc.logger.Debug(fmt.Sprintf("Captured packet: %s:%d -> %s:%d (%s), length: %d", srcIP, srcPort, dstIP, dstPort, protocol, len(packet.Data())))

    // 解析具体协议
    pc.parseProtocol(packet, srcIP, srcPort, dstIP, dstPort, protocol, rawData)
}

// parseProtocol 解析具体协议
func (pc *PacketCapture) parseProtocol(packet gopacket.Packet, srcIP string, srcPort int, dstIP string, dstPort int, protocol string, rawData []byte) {
    switch dstPort {
    case 502:
        pc.parseModbus(rawData)
    case 3306:
        pc.parseMySQL(rawData)
    case 6379:
        pc.parseRedis(rawData)
    case 9092:
        pc.parseKafka(rawData)
    case 80, 443, 9999:
        pc.parseHTTP(rawData)
    default:
        pc.logger.Debug(fmt.Sprintf("Unknown protocol on port %d", dstPort))
    }
}

// parseModbus 解析Modbus协议
func (pc *PacketCapture) parseModbus(rawData []byte) {
    if len(rawData) < 7 {
        return
    }
    // Modbus TCP解析逻辑
    transactionID := uint16(rawData[0])<<8 | uint16(rawData[1])
    protocolID := uint16(rawData[2])<<8 | uint16(rawData[3])
    length := uint16(rawData[4])<<8 | uint16(rawData[5])
    unitID := rawData[6]
    pc.logger.Debug(fmt.Sprintf("Modbus TCP: TransactionID=%d, ProtocolID=%d, Length=%d, UnitID=%d", transactionID, protocolID, length, unitID))
}

// 其他协议解析函数...
```

### 3.4 设备指纹识别模块

#### 3.4.1 模块功能

- 识别访问设备的型号、制造商、操作系统等信息
- 支持多种指纹识别算法（JA3、TCP窗口缩放、TLS扩展等）
- 设备指纹的存储和管理
- 提供设备指纹查询接口

#### 3.4.2 模块设计

```go
// internal/device/fingerprint.go
package device

import (
    "crypto/sha256"
    "database/sql"
    "encoding/json"
    "fmt"
    "time"

    _ "github.com/mattn/go-sqlite3"
    "honeypot-system/internal/logger"
)

// DeviceInfo 设备信息结构体
type DeviceInfo struct {
    OS           string `json:"os"`
    DeviceType   string `json:"device_type"`
    Manufacturer string `json:"manufacturer"`
    DeviceModel  string `json:"device_model"`
    Browser      string `json:"browser,omitempty"`
    JA3Hash      string `json:"ja3_hash,omitempty"`
}

// DeviceFingerprint 设备指纹结构体
type DeviceFingerprint struct {
    ID              string      `json:"id"`
    ClientIP        string      `json:"client_ip"`
    ClientPort      int         `json:"client_port"`
    ServerPort      int         `json:"server_port"`
    Protocol        string      `json:"protocol"`
    FirstSeen       time.Time   `json:"first_seen"`
    LastSeen        time.Time   `json:"last_seen"`
    ConnectionCount int         `json:"connection_count"`
    DeviceInfo      DeviceInfo  `json:"device_info"`
    RawData         []byte      `json:"-"`
}

// FingerprintManager 设备指纹管理器
type FingerprintManager struct {
    dbPath              string
    userAgentAnalysis   bool
    ja3Fingerprinting   bool
    tcpWindowScaling    bool
    tlsExtensions       bool
    db                  *sql.DB
    logger              logger.Logger
    fingerprints        map[string]*DeviceFingerprint
    mutex               sync.RWMutex
}

// NewFingerprintManager 创建新的设备指纹管理器
func NewFingerprintManager(dbPath string, userAgentAnalysis bool, ja3Fingerprinting bool, tcpWindowScaling bool, tlsExtensions bool) *FingerprintManager {
    return &FingerprintManager{
        dbPath:              dbPath,
        userAgentAnalysis:   userAgentAnalysis,
        ja3Fingerprinting:   ja3Fingerprinting,
        tcpWindowScaling:    tcpWindowScaling,
        tlsExtensions:       tlsExtensions,
        fingerprints:        make(map[string]*DeviceFingerprint),
    }
}

// Init 初始化设备指纹管理器
func (fm *FingerprintManager) Init(logger logger.Logger) error {
    fm.logger = logger
    
    // 初始化数据库
    if err := fm.initDatabase(); err != nil {
        return err
    }
    
    // 加载设备指纹到内存
    if err := fm.loadFingerprints(); err != nil {
        return err
    }
    
    return nil
}

// initDatabase 初始化数据库
func (fm *FingerprintManager) initDatabase() error {
    var err error
    fm.db, err = sql.Open("sqlite3", fm.dbPath)
    if err != nil {
        return err
    }
    
    // 创建设备指纹表
    createTableSQL := `
    CREATE TABLE IF NOT EXISTS device_fingerprints (
        id TEXT PRIMARY KEY,
        client_ip TEXT NOT NULL,
        client_port INTEGER NOT NULL,
        server_port INTEGER NOT NULL,
        protocol TEXT NOT NULL,
        first_seen TIMESTAMP NOT NULL,
        last_seen TIMESTAMP NOT NULL,
        connection_count INTEGER NOT NULL,
        device_info TEXT NOT NULL,
        raw_data BLOB
    );
    
    CREATE INDEX IF NOT EXISTS idx_client_ip ON device_fingerprints(client_ip);
    CREATE INDEX IF NOT EXISTS idx_server_port ON device_fingerprints(server_port);
    CREATE INDEX IF NOT EXISTS idx_protocol ON device_fingerprints(protocol);
    `
    
    _, err = fm.db.Exec(createTableSQL)
    return err
}

// loadFingerprints 加载设备指纹到内存
func (fm *FingerprintManager) loadFingerprints() error {
    rows, err := fm.db.Query("SELECT * FROM device_fingerprints")
    if err != nil {
        return err
    }
    defer rows.Close()
    
    for rows.Next() {
        var fp DeviceFingerprint
        var deviceInfoJSON string
        
        err := rows.Scan(
            &fp.ID,
            &fp.ClientIP,
            &fp.ClientPort,
            &fp.ServerPort,
            &fp.Protocol,
            &fp.FirstSeen,
            &fp.LastSeen,
            &fp.ConnectionCount,
            &deviceInfoJSON,
            &fp.RawData,
        )
        if err != nil {
            return err
        }
        
        if err := json.Unmarshal([]byte(deviceInfoJSON), &fp.DeviceInfo); err != nil {
            return err
        }
        
        fm.fingerprints[fp.ID] = &fp
    }
    
    return rows.Err()
}

// CaptureFingerprint 捕获设备指纹
func (fm *FingerprintManager) CaptureFingerprint(srcIP string, srcPort int, dstIP string, dstPort int, protocol string, rawData []byte) {
    // 生成设备指纹ID
    fpID := fm.generateFingerprintID(srcIP, srcPort, dstIP, dstPort, protocol, rawData)
    
    fm.mutex.Lock()
    defer fm.mutex.Unlock()
    
    if fp, exists := fm.fingerprints[fpID]; exists {
        // 更新现有指纹
        fp.LastSeen = time.Now()
        fp.ConnectionCount++
        fp.RawData = rawData
        
        // 更新数据库
        fm.updateFingerprint(fp)
    } else {
        // 创建新指纹
        fp := &DeviceFingerprint{
            ID:              fpID,
            ClientIP:        srcIP,
            ClientPort:      srcPort,
            ServerPort:      dstPort,
            Protocol:        protocol,
            FirstSeen:       time.Now(),
            LastSeen:        time.Now(),
            ConnectionCount: 1,
            DeviceInfo:      fm.identifyDevice(rawData),
            RawData:         rawData,
        }
        
        fm.fingerprints[fpID] = fp
        
        // 保存到数据库
        fm.saveFingerprint(fp)
    }
}

// generateFingerprintID 生成设备指纹ID
func (fm *FingerprintManager) generateFingerprintID(srcIP string, srcPort int, dstIP string, dstPort int, protocol string, rawData []byte) string {
    // 基于IP、端口、协议和数据特征生成唯一ID
    data := fmt.Sprintf("%s:%d:%s:%d:%s:%x", srcIP, srcPort, dstIP, dstPort, protocol, sha256.Sum256(rawData))
    return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
}

// identifyDevice 识别设备信息
func (fm *FingerprintManager) identifyDevice(rawData []byte) DeviceInfo {
    // 设备识别逻辑
    // 结合多种指纹识别算法，如JA3、TCP窗口缩放、TLS扩展等
    deviceInfo := DeviceInfo{
        OS:           "Unknown",
        DeviceType:   "Industrial Device",
        Manufacturer: "Unknown",
        DeviceModel:  "Unknown",
    }
    
    // TODO: 实现具体的设备识别算法
    
    return deviceInfo
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

// GetFingerprint 获取指定ID的设备指纹
func (fm *FingerprintManager) GetFingerprint(id string) (*DeviceFingerprint, bool) {
    fm.mutex.RLock()
    defer fm.mutex.RUnlock()
    
    fp, exists := fm.fingerprints[id]
    return fp, exists
}

// saveFingerprint 保存设备指纹到数据库
func (fm *FingerprintManager) saveFingerprint(fp *DeviceFingerprint) error {
    deviceInfoJSON, err := json.Marshal(fp.DeviceInfo)
    if err != nil {
        return err
    }
    
    _, err = fm.db.Exec(
        "INSERT INTO device_fingerprints (id, client_ip, client_port, server_port, protocol, first_seen, last_seen, connection_count, device_info, raw_data) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
        fp.ID, fp.ClientIP, fp.ClientPort, fp.ServerPort, fp.Protocol, fp.FirstSeen, fp.LastSeen, fp.ConnectionCount, deviceInfoJSON, fp.RawData,
    )
    
    return err
}

// updateFingerprint 更新设备指纹到数据库
func (fm *FingerprintManager) updateFingerprint(fp *DeviceFingerprint) error {
    deviceInfoJSON, err := json.Marshal(fp.DeviceInfo)
    if err != nil {
        return err
    }
    
    _, err = fm.db.Exec(
        "UPDATE device_fingerprints SET last_seen = ?, connection_count = ?, device_info = ?, raw_data = ? WHERE id = ?",
        fp.LastSeen, fp.ConnectionCount, deviceInfoJSON, fp.RawData, fp.ID,
    )
    
    return err
}

// Close 关闭设备指纹管理器
func (fm *FingerprintManager) Close() error {
    if fm.db != nil {
        return fm.db.Close()
    }
    return nil
}
```

### 3.5 Web服务模块

#### 3.5.1 模块功能

- 提供RESTful API接口
- 提供Web界面，展示实时数据和统计信息
- 支持设备指纹查询和管理
- 支持连接记录查询和会话流量分析

#### 3.5.2 模块设计

```go
// internal/web/server.go
package web

import (
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"

    "honeypot-system/internal/device"
    "honeypot-system/internal/logger"
)

// Server Web服务器结构体
type Server struct {
    host              string
    port              int
    https             bool
    sessionTimeout    int
    fingerprintManager *device.FingerprintManager
    logger             logger.Logger
    engine             *gin.Engine
    server             *http.Server
}

// NewServer 创建新的Web服务器
func NewServer(host string, port int, https bool, sessionTimeout int, fingerprintManager *device.FingerprintManager, logger logger.Logger) *Server {
    engine := gin.Default()

    // 设置静态文件服务
    engine.Static("/static", "./web/static")
    // 设置模板目录
    engine.LoadHTMLGlob("./web/templates/*")

    s := &Server{
        host:              host,
        port:              port,
        https:             https,
        sessionTimeout:    sessionTimeout,
        fingerprintManager: fingerprintManager,
        logger:             logger,
        engine:            engine,
    }

    // 注册路由
    s.registerRoutes()

    return s
}

// registerRoutes 注册路由
func (s *Server) registerRoutes() {
    // 主页
    s.engine.GET("/", s.handleIndex)
    // 设备指纹列表
    s.engine.GET("/fingerprints", s.handleFingerprints)
    // 设备指纹详情
    s.engine.GET("/fingerprints/:id", s.handleFingerprintDetail)
    // 连接记录
    s.engine.GET("/connections", s.handleConnections)
    // 会话流量
    s.engine.GET("/sessions/:id", s.handleSessionDetail)
    // API - 获取设备指纹列表
    s.engine.GET("/api/fingerprints", s.handleAPIFingerprints)
    // API - 获取设备指纹详情
    s.engine.GET("/api/fingerprints/:id", s.handleAPIFingerprintDetail)
    // API - 获取统计信息
    s.engine.GET("/api/stats", s.handleAPIStats)
    // API - 获取连接记录
    s.engine.GET("/api/connections", s.handleAPIConnections)
}

// handleIndex 处理主页请求
func (s *Server) handleIndex(c *gin.Context) {
    fingerprints := s.fingerprintManager.GetAllFingerprints()
    
    // 计算统计信息
    stats := s.calculateStats(fingerprints)
    
    c.HTML(http.StatusOK, "index.html", gin.H{
        "title":        "Industrial Protocol Honeypot",
        "fingerprints": fingerprints,
        "stats":        stats,
        "timestamp":    time.Now().Format(time.RFC3339),
    })
}

// handleAPIFingerprints 处理API设备指纹列表请求
func (s *Server) handleAPIFingerprints(c *gin.Context) {
    fingerprints := s.fingerprintManager.GetAllFingerprints()
    
    c.JSON(http.StatusOK, gin.H{
        "success":     true,
        "fingerprints": fingerprints,
        "count":       len(fingerprints),
    })
}

// handleAPIStats 处理API统计信息请求
func (s *Server) handleAPIStats(c *gin.Context) {
    fingerprints := s.fingerprintManager.GetAllFingerprints()
    stats := s.calculateStats(fingerprints)
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "stats":   stats,
        "timestamp": time.Now().Format(time.RFC3339),
    })
}

// calculateStats 计算统计信息
func (s *Server) calculateStats(fingerprints []*device.DeviceFingerprint) map[string]interface{} {
    stats := map[string]interface{}{
        "total_connections":     0,
        "total_fingerprints":    len(fingerprints),
        "unique_ips":            make(map[string]bool),
        "os_distribution":       make(map[string]int),
        "device_types":          make(map[string]int),
        "protocol_distribution": make(map[string]int),
    }
    
    for _, fp := range fingerprints {
        stats["total_connections"] = stats["total_connections"].(int) + fp.ConnectionCount
        stats["unique_ips"].(map[string]bool)[fp.ClientIP] = true
        
        os := fp.DeviceInfo.OS
        if os == "" {
            os = "Unknown"
        }
        stats["os_distribution"].(map[string]int)[os]++
        
        deviceType := fp.DeviceInfo.DeviceType
        if deviceType == "" {
            deviceType = "Unknown"
        }
        stats["device_types"].(map[string]int)[deviceType]++
        
        stats["protocol_distribution"].(map[string]int)[fp.Protocol]++
    }
    
    stats["unique_ips_count"] = len(stats["unique_ips"].(map[string]bool))
    delete(stats, "unique_ips")
    
    return stats
}

// 其他路由处理函数...

// Start 启动Web服务器
func (s *Server) Start() error {
    addr := fmt.Sprintf("%s:%d", s.host, s.port)
    
    s.server = &http.Server{
        Addr:    addr,
        Handler: s.engine,
    }
    
    go func() {
        if s.https {
            s.logger.Info(fmt.Sprintf("Web server started on https://%s", addr))
            if err := s.server.ListenAndServeTLS("./web/cert.pem", "./web/key.pem"); err != nil && err != http.ErrServerClosed {
                s.logger.Error(fmt.Sprintf("Failed to start HTTPS server: %v", err))
            }
        } else {
            s.logger.Info(fmt.Sprintf("Web server started on http://%s", addr))
            if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
                s.logger.Error(fmt.Sprintf("Failed to start HTTP server: %v", err))
            }
        }
    }()
    
    return nil
}

// Stop 停止Web服务器
func (s *Server) Stop() {
    if s.server != nil {
        if err := s.server.Close(); err != nil {
            s.logger.Error(fmt.Sprintf("Failed to stop web server: %v", err))
        } else {
            s.logger.Info("Web server stopped successfully")
        }
    }
}
```

## 4. 数据库设计

### 4.1 数据库选型

本系统采用SQLite数据库，主要考虑以下因素：

1. **轻量级**：SQLite是嵌入式数据库，无需独立部署，适合蜜罐系统的轻量级部署需求
2. **高性能**：对于设备指纹和连接记录等结构化数据，SQLite的性能足以满足需求
3. **跨平台**：SQLite支持多种操作系统，便于蜜罐系统的跨平台部署
4. **易用性**：SQLite的API简单易用，学习成本低
5. **可靠性**：SQLite具有良好的事务支持和数据完整性保证

### 4.2 数据库表结构

#### 4.2.1 设备指纹表 (device_fingerprints)

| 字段名称         | 数据类型     | 约束          | 描述                     |
|----------------|----------|-------------|------------------------|
| id             | TEXT     | PRIMARY KEY | 设备指纹唯一标识符              |
| client_ip      | TEXT     | NOT NULL    | 客户端IP地址               |
| client_port    | INTEGER  | NOT NULL    | 客户端端口                 |
| server_port    | INTEGER  | NOT NULL    | 服务器端口                 |
| protocol       | TEXT     | NOT NULL    | 协议类型（Modbus、MySQL等）     |
| first_seen     | TIMESTAMP| NOT NULL    | 首次发现时间                |
| last_seen      | TIMESTAMP| NOT NULL    | 最后一次发现时间              |
| connection_count| INTEGER | NOT NULL    | 连接次数                  |
| device_info    | TEXT     | NOT NULL    | 设备信息JSON               |
| raw_data       | BLOB     |             | 原始数据（可选）              |

**索引**：
- idx_client_ip：客户端IP索引
- idx_server_port：服务器端口索引
- idx_protocol：协议类型索引

#### 4.2.2 连接记录表 (connections)

| 字段名称         | 数据类型     | 约束          | 描述                     |
|----------------|----------|-------------|------------------------|
| id             | TEXT     | PRIMARY KEY | 连接记录唯一标识符             |
| client_ip      | TEXT     | NOT NULL    | 客户端IP地址               |
| client_port    | INTEGER  | NOT NULL    | 客户端端口                 |
| server_port    | INTEGER  | NOT NULL    | 服务器端口                 |
| protocol       | TEXT     | NOT NULL    | 协议类型                   |
| timestamp      | TIMESTAMP| NOT NULL    | 连接时间                  |
| duration       | INTEGER  | NOT NULL    | 连接持续时间（秒）             |
| bytes_sent     | INTEGER  | NOT NULL    | 发送的字节数                |
| bytes_received | INTEGER  | NOT NULL    | 接收的字节数                |
| fingerprint_id | TEXT     |             | 关联的设备指纹ID             |

**索引**：
- idx_conn_client_ip：客户端IP索引
- idx_conn_timestamp：连接时间索引
- idx_conn_fingerprint：设备指纹ID索引

#### 4.2.3 会话流量表 (sessions)

| 字段名称         | 数据类型     | 约束          | 描述                     |
|----------------|----------|-------------|------------------------|
| id             | TEXT     | PRIMARY KEY | 会话唯一标识符                |
| client_ip      | TEXT     | NOT NULL    | 客户端IP地址               |
| client_port    | INTEGER  | NOT NULL    | 客户端端口                 |
| server_port    | INTEGER  | NOT NULL    | 服务器端口                 |
| protocol       | TEXT     | NOT NULL    | 协议类型                   |
| start_time     | TIMESTAMP| NOT NULL    | 会话开始时间                |
| end_time       | TIMESTAMP|             | 会话结束时间                |
| total_bytes    | INTEGER  | NOT NULL    | 会话总流量（字节）             |
| packets        | INTEGER  | NOT NULL    | 会话数据包数量               |
| fingerprint_id | TEXT     |             | 关联的设备指纹ID             |

**索引**：
- idx_session_client_ip：客户端IP索引
- idx_session_time：会话时间索引

## 5. 接口设计

### 5.1 RESTful API接口

#### 5.1.1 获取设备指纹列表

- **URL**：`/api/fingerprints`
- **方法**：GET
- **参数**：无
- **响应**：
  ```json
  {
    "success": true,
    "fingerprints": [
      {
        "id": "fingerprint-123",
        "client_ip": "192.168.1.100",
        "client_port": 12345,
        "server_port": 502,
        "protocol": "modbus",
        "first_seen": "2026-01-13T12:00:00Z",
        "last_seen": "2026-01-13T12:30:00Z",
        "connection_count": 5,
        "device_info": {
          "os": "Unknown",
          "device_type": "Industrial Device",
          "manufacturer": "Siemens",
          "device_model": "S7-1200"
        }
      }
    ],
    "count": 1
  }
  ```

#### 5.1.2 获取设备指纹详情

- **URL**：`/api/fingerprints/:id`
- **方法**：GET
- **参数**：
  - `id`：设备指纹ID（路径参数）
- **响应**：
  ```json
  {
    "success": true,
    "fingerprint": {
      "id": "fingerprint-123",
      "client_ip": "192.168.1.100",
      "client_port": 12345,
      "server_port": 502,
      "protocol": "modbus",
      "first_seen": "2026-01-13T12:00:00Z",
      "last_seen": "2026-01-13T12:30:00Z",
      "connection_count": 5,
      "device_info": {
        "os": "Unknown",
        "device_type": "Industrial Device",
        "manufacturer": "Siemens",
        "device_model": "S7-1200"
      }
    }
  }
  ```

#### 5.1.3 获取统计信息

- **URL**：`/api/stats`
- **方法**：GET
- **参数**：无
- **响应**：
  ```json
  {
    "success": true,
    "stats": {
      "total_connections": 100,
      "total_fingerprints": 50,
      "unique_ips_count": 20,
      "os_distribution": {
        "Unknown": 45,
        "Linux": 5
      },
      "device_types": {
        "Industrial Device": 50
      },
      "protocol_distribution": {
        "modbus": 30,
        "mysql": 10,
        "redis": 5,
        "kafka": 5
      }
    },
    "timestamp": "2026-01-13T12:30:00Z"
  }
  ```

#### 5.1.4 获取连接记录

- **URL**：`/api/connections`
- **方法**：GET
- **参数**：
  - `limit`：返回记录数量（可选，默认20）
  - `offset`：偏移量（可选，默认0）
  - `start_time`：开始时间（可选）
  - `end_time`：结束时间（可选）
  - `client_ip`：客户端IP（可选）
  - `protocol`：协议类型（可选）
- **响应**：
  ```json
  {
    "success": true,
    "connections": [
      {
        "id": "conn-123",
        "client_ip": "192.168.1.100",
        "client_port": 12345,
        "server_port": 502,
        "protocol": "modbus",
        "timestamp": "2026-01-13T12:00:00Z",
        "duration": 60,
        "bytes_sent": 1024,
        "bytes_received": 2048,
        "fingerprint_id": "fingerprint-123"
      }
    ],
    "count": 1,
    "total": 100
  }
  ```

### 5.2 模块间接口

#### 5.2.1 数据包捕获模块 → 设备指纹识别模块

- **接口**：`CaptureFingerprint(srcIP string, srcPort int, dstIP string, dstPort int, protocol string, rawData []byte)`
- **功能**：将捕获的数据包传递给设备指纹识别模块，进行设备指纹识别
- **参数**：
  - `srcIP`：源IP地址
  - `srcPort`：源端口
  - `dstIP`：目标IP地址
  - `dstPort`：目标端口
  - `protocol`：协议类型
  - `rawData`：原始数据包数据

#### 5.2.2 设备指纹识别模块 → Web服务模块

- **接口**：`GetAllFingerprints() []*DeviceFingerprint`
- **功能**：获取所有设备指纹列表
- **返回值**：设备指纹列表

- **接口**：`GetFingerprint(id string) (*DeviceFingerprint, bool)`
- **功能**：根据ID获取设备指纹详情
- **参数**：
  - `id`：设备指纹ID
- **返回值**：
  - 设备指纹对象
  - 是否存在

## 6. 部署设计

### 6.1 部署架构

系统采用单节点部署架构，所有组件部署在同一台服务器上，便于维护和管理。对于大规模部署，可以考虑分布式架构，将数据包捕获、数据处理和Web服务分离部署。

### 6.2 部署方式

#### 6.2.1 本地开发环境部署

1. **安装依赖**：
   ```bash
   go mod download
   ```

2. **编译项目**：
   ```bash
   go build -o honeypot_server ./cmd/api
   ```

3. **运行项目**：
   ```bash
   ./honeypot_server
   ```

#### 6.2.2 服务器部署

1. **准备服务器**：
   - 操作系统：CentOS 7/8或Ubuntu 20.04/22.04
   - 硬件要求：4核CPU、8GB内存、100GB存储空间
   - 网络要求：稳定的网络连接，开放需要监听的端口

2. **安装依赖**：
   ```bash
   # CentOS
   yum install -y go git
   
   # Ubuntu
   apt-get update
   apt-get install -y go git
   ```

3. **编译项目**：
   ```bash
   git clone <repository-url>
   cd honeypot-system
   go mod download
   go build -o honeypot_server ./cmd/api
   ```

4. **创建配置文件**：
   ```bash
   cp config/config.yaml.example config/config.yaml
   # 根据实际情况修改配置文件
   ```

5. **创建系统服务**：
   ```bash
   # 创建systemd服务文件
   cat > /etc/systemd/system/honeypot.service << EOF
   [Unit]
   Description=Industrial Protocol Honeypot System
   After=network.target
   
   [Service]
   Type=simple
   User=root
   WorkingDirectory=/path/to/honeypot-system
   ExecStart=/path/to/honeypot-system/honeypot_server
   Restart=on-failure
   
   [Install]
   WantedBy=multi-user.target
   EOF
   
   # 启动服务
   systemctl daemon-reload
   systemctl start honeypot
   systemctl enable honeypot
   ```

#### 6.2.3 Docker容器化部署

1. **创建Dockerfile**：
   ```dockerfile
   FROM golang:1.21-alpine AS builder
   
   WORKDIR /app
   
   COPY go.mod go.sum ./
   RUN go mod download
   
   COPY . .
   RUN CGO_ENABLED=0 GOOS=linux go build -o honeypot_server ./cmd/api
   
   FROM alpine:latest
   
   RUN apk --no-cache add libc6-compat tzdata
   
   WORKDIR /app
   
   COPY --from=builder /app/honeypot_server /app/
   COPY --from=builder /app/config/config.yaml.example /app/config/config.yaml
   COPY --from=builder /app/web /app/web
   
   EXPOSE 8080
   
   CMD ["./honeypot_server"]
   ```

2. **构建Docker镜像**：
   ```bash
   docker build -t honeypot-system .
   ```

3. **运行Docker容器**：
   ```bash
   docker run -d --name honeypot -p 8080:8080 -v ./data:/app/data -v ./logs:/app/logs honeypot-system
   ```

### 6.3 启动脚本

提供跨平台的启动脚本，支持macOS和CentOS：

#### 6.3.1 macOS启动脚本 (start.sh)

```bash
#!/bin/bash

# 工业协议蜜罐系统启动脚本（macOS）

# 设置工作目录
WORKDIR=$(cd "$(dirname "$0")" && pwd)
cd $WORKDIR

# 设置环境变量
export GOPROXY=https://goproxy.cn,direct
export GOMODCACHE=$WORKDIR/.go-mod-cache

# 编译项目
echo "正在编译项目..."
go build -o honeypot_server ./cmd/api

# 启动项目
echo "正在启动工业协议蜜罐系统..."
./honeypot_server
```

#### 6.3.2 CentOS启动脚本 (start_centos.sh)

```bash
#!/bin/bash

# 工业协议蜜罐系统启动脚本（CentOS）

# 设置工作目录
WORKDIR=$(cd "$(dirname "$0")" && pwd)
cd $WORKDIR

# 检查依赖
echo "检查依赖..."
if ! command -v go &> /dev/null; then
    echo "错误：未安装Go语言"
    exit 1
fi

# 设置环境变量
export GOPROXY=https://goproxy.cn,direct
export GOMODCACHE=$WORKDIR/.go-mod-cache

# 编译项目
echo "正在编译项目..."
go build -o honeypot_server ./cmd/api

# 启动项目
echo "正在启动工业协议蜜罐系统..."
./honeypot_server
```

## 7. 安全设计

### 7.1 系统安全

1. **最小权限原则**：
   - 蜜罐系统以普通用户身份运行，仅在必要时使用root权限
   - 限制系统资源访问权限，避免被用作攻击跳板

2. **网络安全**：
   - 使用防火墙限制蜜罐系统的网络访问范围
   - 仅开放必要的端口，如Web服务端口8080
   - 禁止蜜罐系统主动发起网络连接

3. **数据安全**：
   - 设备指纹和连接记录等敏感数据加密存储
   - 定期备份数据，防止数据丢失
   - 实现数据访问控制，防止未授权访问

4. **日志安全**：
   - 系统日志加密存储，防止日志被篡改
   - 实现日志轮转和归档，防止日志过大
   - 定期检查日志，及时发现异常行为

### 7.2 蜜罐安全

1. **蜜罐伪装**：
   - 模拟真实工业设备的行为和响应
   - 避免泄露蜜罐系统的真实身份
   - 生成真实的设备指纹和协议响应

2. **攻击隔离**：
   - 蜜罐系统部署在隔离的网络环境中
   - 限制蜜罐系统对内部网络的访问
   - 实现入侵检测和防御机制

3. **异常检测**：
   - 实时监控蜜罐系统的运行状态
   - 检测异常行为和攻击尝试
   - 实现自动告警机制

## 8. 测试设计

### 8.1 单元测试

针对每个模块编写单元测试，确保模块功能的正确性：

1. **配置管理模块测试**：
   - 测试配置文件加载
   - 测试配置参数获取
   - 测试默认配置设置

2. **日志管理模块测试**：
   - 测试不同日志级别的输出
   - 测试日志文件滚动
   - 测试日志格式正确性

3. **设备指纹识别模块测试**：
   - 测试设备指纹生成
   - 测试设备识别准确性
   - 测试数据库操作

4. **Web服务模块测试**：
   - 测试API接口响应
   - 测试HTML模板渲染
   - 测试统计信息计算

### 8.2 集成测试

测试模块之间的集成是否正常：

1. **数据包捕获与设备指纹识别集成测试**：
   - 测试数据包捕获后是否能正确调用设备指纹识别
   - 测试设备指纹识别结果是否正确

2. **设备指纹识别与Web服务集成测试**：
   - 测试设备指纹是否能正确显示在Web界面
   - 测试API接口是否能正确返回设备指纹数据

3. **完整流程测试**：
   - 测试从数据包捕获到Web界面展示的完整流程
   - 测试系统在不同场景下的表现

### 8.3 性能测试

测试系统在高并发情况下的性能：

1. **并发连接测试**：
   - 模拟大量并发连接，测试系统的处理能力
   - 测试系统的响应时间和资源占用情况

2. **数据包处理测试**：
   - 模拟大量数据包，测试系统的处理速度
   - 测试系统的内存和CPU占用情况

3. **Web界面性能测试**：
   - 测试Web界面在大量数据情况下的加载速度
   - 测试API接口的响应时间

### 8.4 安全性测试

测试系统的安全性：

1. **蜜罐系统安全性测试**：
   - 测试蜜罐系统是否存在漏洞
   - 测试蜜罐系统是否容易被识别为蜜罐

2. **数据安全测试**：
   - 测试敏感数据的加密存储
   - 测试数据访问控制

3. **网络安全测试**：
   - 测试蜜罐系统的网络访问控制
   - 测试蜜罐系统是否能防止被用作攻击跳板

## 9. 维护与监控

### 9.1 系统维护

1. **定期更新**：
   - 定期更新系统和依赖库，修复安全漏洞
   - 定期更新蜜罐系统的协议支持和指纹库

2. **日志管理**：
   - 定期检查系统日志，及时发现异常行为
   - 实现日志轮转和归档，防止日志过大

3. **数据管理**：
   - 定期备份数据，防止数据丢失
   - 定期清理过期数据，优化数据库性能

### 9.2 系统监控

1. **运行状态监控**：
   - 监控系统的CPU、内存、磁盘、网络等资源占用
   - 监控系统的运行状态，如连接数、数据包处理速度等

2. **异常行为监控**：
   - 监控异常连接和攻击尝试
   - 监控系统的异常日志和错误信息

3. **告警机制**：
   - 实现基于阈值的告警机制
   - 支持多种告警方式，如邮件、短信、Webhook等

## 10. 总结

本程序设计文档详细描述了工业协议蜜罐系统的设计方案，包括系统架构、模块设计、数据库设计、接口设计、部署设计、安全设计和测试设计等内容。该文档为开发人员提供了明确的开发指导，确保系统能够满足需求分析文档中提出的所有功能和非功能需求。

系统采用模块化、分层架构设计，具有良好的扩展性和可维护性。同时，系统注重安全性设计，确保蜜罐系统本身的安全性，防止被用作攻击跳板。

通过本设计方案开发的工业协议蜜罐系统，将能够有效地监控和分析工业协议流量，识别设备指纹，提供可视化的Web界面展示实时数据，为工业控制系统的安全防护提供有力支持。