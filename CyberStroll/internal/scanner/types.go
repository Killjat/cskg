package scanner

import (
	"time"
)

// ScanTask 扫描任务
type ScanTask struct {
	TaskID    string    `json:"task_id"`
	IP        string    `json:"ip"`
	TaskType  string    `json:"task_type"`
	Priority  int       `json:"priority"`
	Config    ScanConfig `json:"config"`
	Timestamp int64     `json:"timestamp"`
}

// ScanConfig 扫描配置
type ScanConfig struct {
	Ports      []int  `json:"ports"`
	Timeout    int    `json:"timeout"`
	ScanDepth  string `json:"scan_depth"` // "basic", "deep"
	EnableApps bool   `json:"enable_apps"`
}

// ScanResult 扫描结果
type ScanResult struct {
	TaskID       string        `json:"task_id"`
	IP           string        `json:"ip"`
	ScanType     string        `json:"scan_type"`
	ScanStatus   string        `json:"scan_status"`
	ScanTime     string        `json:"scan_time"`
	ResponseTime int64         `json:"response_time"`
	Results      ScanDetails   `json:"results"`
	ErrorMessage string        `json:"error_message,omitempty"`
	NodeID       string        `json:"node_id"`
	Timestamp    int64         `json:"timestamp"`
}

// ScanDetails 扫描详情
type ScanDetails struct {
	OpenPorts    []PortInfo      `json:"open_ports"`
	Applications []ApplicationInfo `json:"applications"`
	OSInfo       *OSInfo         `json:"os_info,omitempty"`
}

// PortInfo 端口信息
type PortInfo struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
	Version  string `json:"version"`
	Banner   string `json:"banner"`
	State    string `json:"state"`
}

// ApplicationInfo 应用信息
type ApplicationInfo struct {
	Name       string   `json:"name"`
	Version    string   `json:"version"`
	Category   string   `json:"category"`
	Confidence int      `json:"confidence"`
	CPE        string   `json:"cpe"`
	Tags       []string `json:"tags"`
}

// OSInfo 操作系统信息
type OSInfo struct {
	OSName     string `json:"os_name"`
	OSVersion  string `json:"os_version"`
	OSFamily   string `json:"os_family"`
	Confidence int    `json:"confidence"`
}

// ScannerConfig 扫描器配置
type ScannerConfig struct {
	MaxConcurrency int           `yaml:"max_concurrency"`
	Timeout        time.Duration `yaml:"timeout"`
	RetryCount     int           `yaml:"retry_count"`
	ProbeDelay     time.Duration `yaml:"probe_delay"`
	EnableLogging  bool          `yaml:"enable_logging"`
}

// ScannerStats 扫描器统计
type ScannerStats struct {
	TotalScans    int64 `json:"total_scans"`
	SuccessScans  int64 `json:"success_scans"`
	FailedScans   int64 `json:"failed_scans"`
	AverageTime   int64 `json:"average_time"`
	LastScanTime  int64 `json:"last_scan_time"`
}