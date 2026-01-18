package models

import (
	"time"
)

// TaskType 任务类型
type TaskType string

// 定义任务类型常量
const (
	TaskTypeIPAlive     TaskType = "ip_alive"      // IP探活
	TaskTypePortScan    TaskType = "port_scan"     // 端口探查
	TaskTypeServiceScan TaskType = "service_scan"  // 服务识别
	TaskTypeWebScan     TaskType = "web_scan"      // 网站识别
	TaskTypeBannerGrab  TaskType = "banner_grab"   // Banner抓取
)

// ProtocolType 协议类型
type ProtocolType string

// 定义协议类型常量
const (
	ProtocolTCP ProtocolType = "tcp"  // TCP协议
	ProtocolUDP ProtocolType = "udp"  // UDP协议
	ProtocolAll ProtocolType = "all"  // 所有协议
)

// TaskStatus 任务状态
type TaskStatus string

// 定义任务状态常量
const (
	TaskStatusPending    TaskStatus = "pending"    // 等待中
	TaskStatusRunning    TaskStatus = "running"    // 运行中
	TaskStatusCompleted  TaskStatus = "completed"  // 已完成
	TaskStatusFailed     TaskStatus = "failed"     // 失败
)

// Task 任务结构体
type Task struct {
	ID               string        `json:"id"`               // 任务ID
	Creator          string        `json:"creator"`          // 任务发起人
	Type             TaskType      `json:"type"`             // 任务类型
	Protocol         ProtocolType  `json:"protocol"`         // 协议类型
	Targets          []string      `json:"targets"`          // 目标IP列表
	PortRange        string        `json:"port_range"`       // 端口范围
	Status           TaskStatus    `json:"status"`           // 任务状态
	CreatedAt        time.Time     `json:"created_at"`       // 任务创建时间
	ScheduledAt      time.Time     `json:"scheduled_at"`     // 任务调度时间
	StartedAt        *time.Time    `json:"started_at,omitempty"` // 任务开始时间
	CompletedAt      *time.Time    `json:"completed_at,omitempty"` // 任务完成时间
	ScanResultIndex  string        `json:"scan_result_index,omitempty"` // 扫描结果索引
	ScanResultKeys   []string      `json:"scan_result_keys,omitempty"` // 扫描结果Keys
	ErrorMessage     string        `json:"error_message,omitempty"` // 错误信息
}

// ScanResult 扫描结果结构体
type ScanResult struct {
	IP       string    `json:"ip"`       // IP地址
	Port     int       `json:"port"`     // 端口号
	Protocol string    `json:"protocol"` // 协议类型
	Service  string    `json:"service"`  // 服务名称
	App      string    `json:"app"`      // 应用服务（通过探测包确定）
	Banner   string    `json:"banner"`   // Banner信息
	Status   string    `json:"status"`   // 端口状态
	CreatedAt time.Time `json:"created_at"` // 扫描时间
	TaskID   string    `json:"task_id"`   // 任务ID
}
