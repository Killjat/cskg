package main

import (
	"time"
)

// ProbeType 探测类型
type ProbeType string

const (
	ProbeTypeTCP ProbeType = "TCP"
	ProbeTypeUDP ProbeType = "UDP"
)

// Probe 探测定义
type Probe struct {
	Name        string    `json:"name"`        // 探测名称
	Type        ProbeType `json:"type"`        // 探测类型 (TCP/UDP)
	Payload     []byte    `json:"payload"`     // 探测载荷
	PayloadHex  string    `json:"payload_hex"` // 十六进制表示
	Ports       []int     `json:"ports"`       // 目标端口
	Protocol    string    `json:"protocol"`    // 协议名称
	Description string    `json:"description"` // 描述
	Timeout     int       `json:"timeout"`     // 超时时间(秒)
	Rarity      int       `json:"rarity"`      // 稀有度 (1-9)
}

// ProbeResult 探测结果
type ProbeResult struct {
	Target      string        `json:"target"`       // 目标地址
	Port        int           `json:"port"`         // 端口
	ProbeName   string        `json:"probe_name"`   // 探测名称
	Protocol    string        `json:"protocol"`     // 协议
	Success     bool          `json:"success"`      // 是否成功
	Response    []byte        `json:"response"`     // 原始响应
	ResponseHex string        `json:"response_hex"` // 十六进制响应
	Banner      string        `json:"banner"`       // 解析后的Banner
	ParsedInfo  *ParsedInfo   `json:"parsed_info"`  // 协议解析信息
	Duration    time.Duration `json:"duration"`     // 探测耗时
	Error       string        `json:"error"`        // 错误信息
	Timestamp   time.Time     `json:"timestamp"`    // 时间戳
}

// ParsedInfo 协议解析信息
type ParsedInfo struct {
	Protocol    string            `json:"protocol"`     // 协议名称
	Service     string            `json:"service"`      // 服务名称
	Version     string            `json:"version"`      // 版本信息
	Product     string            `json:"product"`      // 产品名称
	ExtraInfo   string            `json:"extra_info"`   // 额外信息
	Hostname    string            `json:"hostname"`     // 主机名
	OS          string            `json:"os"`           // 操作系统
	DeviceType  string            `json:"device_type"`  // 设备类型
	CPE         string            `json:"cpe"`          // CPE标识
	Fields      map[string]string `json:"fields"`       // 协议字段
	Confidence  int               `json:"confidence"`   // 置信度
}

// Target 探测目标
type Target struct {
	Host string `json:"host"` // 主机地址
	Port int    `json:"port"` // 端口
}

// ProbeConfig 探测配置
type ProbeConfig struct {
	Timeout         time.Duration `json:"timeout"`          // 默认超时时间
	MaxConcurrency  int           `json:"max_concurrency"`  // 最大并发数
	RetryCount      int           `json:"retry_count"`      // 重试次数
	ReadTimeout     time.Duration `json:"read_timeout"`     // 读取超时
	ConnectTimeout  time.Duration `json:"connect_timeout"`  // 连接超时
	MaxResponseSize int           `json:"max_response_size"` // 最大响应大小
	EnableLogging   bool          `json:"enable_logging"`   // 是否启用日志
	ProbeDelay      time.Duration `json:"probe_delay"`      // 探测间隔
	AdaptiveTimeout bool          `json:"adaptive_timeout"` // 自适应超时
	NetworkDiag     bool          `json:"network_diag"`     // 网络诊断
}

// DefaultProbeConfig 默认配置
func DefaultProbeConfig() *ProbeConfig {
	return &ProbeConfig{
		Timeout:         time.Second * 10,
		MaxConcurrency:  50,
		RetryCount:      2,
		ReadTimeout:     time.Second * 5,
		ConnectTimeout:  time.Second * 3,
		MaxResponseSize: 8192,
		EnableLogging:   true,
	}
}

// ProbeStats 探测统计
type ProbeStats struct {
	TotalProbes    int           `json:"total_probes"`    // 总探测数
	SuccessProbes  int           `json:"success_probes"`  // 成功探测数
	FailedProbes   int           `json:"failed_probes"`   // 失败探测数
	AvgDuration    time.Duration `json:"avg_duration"`    // 平均耗时
	TotalDuration  time.Duration `json:"total_duration"`  // 总耗时
	ProtocolCounts map[string]int `json:"protocol_counts"` // 协议统计
}

// ProbeEngine 探测引擎
type ProbeEngine struct {
	config    *ProbeConfig
	probes    map[string]*Probe // 探测库
	parsers   map[string]ProtocolParser // 协议解析器
	stats     *ProbeStats
}

// ProtocolParser 协议解析器接口
type ProtocolParser interface {
	Parse(data []byte) (*ParsedInfo, error)
	GetProtocol() string
	GetConfidence(data []byte) int
}