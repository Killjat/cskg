package storage

import "time"

// IPSegment IP段信息
type IPSegment struct {
	CIDR        string    `json:"cidr"`         // CIDR格式的IP段，如 "1.2.3.0/24"
	StartIP     string    `json:"start_ip"`     // 起始IP
	EndIP       string    `json:"end_ip"`       // 结束IP
	Country     string    `json:"country"`      // 国家代码
	Type        string    `json:"type"`         // 类型（ipv4/ipv6）
	Status      string    `json:"status"`       // 状态（allocated/assigned）
	Date        string    `json:"date"`         // 分配日期
	Registry    string    `json:"registry"`     // 注册机构
	IPCount     uint32    `json:"ip_count"`     // IP数量
	CreatedAt   time.Time `json:"created_at"`   // 创建时间
}

// AliveResult IP探活结果
type AliveResult struct {
	IP            string        `json:"ip"`              // IP地址
	CIDR          string        `json:"cidr"`            // 所属C段
	IsAlive       bool          `json:"is_alive"`        // 是否存活
	ResponseTime  time.Duration `json:"response_time"`   // 响应时间
	PacketLoss    float64       `json:"packet_loss"`     // 丢包率
	ScanTime      time.Time     `json:"scan_time"`       // 扫描时间
	ErrorMessage  string        `json:"error_message"`   // 错误信息
}

// ScanTask 扫描任务
type ScanTask struct {
	CIDR      string    `json:"cidr"`       // 要扫描的C段
	StartTime time.Time `json:"start_time"` // 开始时间
	Status    string    `json:"status"`     // 状态：pending, running, completed, failed
}

// ScanStats 扫描统计
type ScanStats struct {
	TotalSegments     int       `json:"total_segments"`     // 总IP段数
	ProcessedSegments int       `json:"processed_segments"` // 已处理段数
	TotalIPs          int       `json:"total_ips"`          // 总IP数
	AliveIPs          int       `json:"alive_ips"`          // 存活IP数
	StartTime         time.Time `json:"start_time"`         // 开始时间
	EndTime           time.Time `json:"end_time"`           // 结束时间
	Duration          string    `json:"duration"`           // 持续时间
}