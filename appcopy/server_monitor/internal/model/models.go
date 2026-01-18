package model

import (
	"time"
)

// LoginData 登录数据模型
type LoginData struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username"`
	LoginTime    time.Time `json:"login_time"`
	SourceIP     string    `json:"source_ip"`
	Terminal     string    `json:"terminal"`
	LoginType    string    `json:"login_type"` // local, ssh, etc.
	Duration     int       `json:"duration"`    // in seconds
	LogoutTime   time.Time `json:"logout_time"`
	Status       string    `json:"status"`     // active, logout
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// FileOperationData 文件操作数据模型
type FileOperationData struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Operation     string    `json:"operation"` // create, read, write, delete, copy, move
	FilePath      string    `json:"file_path"`
	OperationTime time.Time `json:"operation_time"`
	Username      string    `json:"username"`
	PID           int       `json:"pid"`
	ProcessName   string    `json:"process_name"`
	Result        string    `json:"result"` // success, failed
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ProcessData 进程数据模型
type ProcessData struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	PID          int       `json:"pid"`
	PPID         int       `json:"ppid"`
	Name         string    `json:"name"`
	Command      string    `json:"command"`
	Username     string    `json:"username"`
	StartTime    time.Time `json:"start_time"`
	CPUUsage     float64   `json:"cpu_usage"`
	MemoryUsage  float64   `json:"memory_usage"`
	Status       string    `json:"status"` // running, sleeping, zombie, etc.
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CommandData 命令数据模型
type CommandData struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Command      string    `json:"command"`
	PID          int       `json:"pid"`
	PPID         int       `json:"ppid"`
	Username     string    `json:"username"`
	StartTime    time.Time `json:"start_time"`
	Duration     int       `json:"duration"`    // in seconds
	CPUUsage     float64   `json:"cpu_usage"`
	MemoryUsage  float64   `json:"memory_usage"`
	Status       string    `json:"status"`     // running, completed
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AlertData 告警数据模型
type AlertData struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	AlertType        string    `json:"alert_type"` // login, file, process, command
	AlertLevel       string    `json:"alert_level"` // info, warning, error, critical
	Message          string    `json:"message"`
	Detail           string    `json:"detail"`
	Status           string    `json:"status"`     // pending, acknowledged, resolved
	AcknowledgedBy   string    `json:"acknowledged_by"`
	AcknowledgedTime time.Time `json:"acknowledged_time"`
	ResolvedTime     time.Time `json:"resolved_time"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// SystemStats 系统统计数据模型
type SystemStats struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	Timestamp        time.Time `json:"timestamp"`
	TotalProcesses   int       `json:"total_processes"`
	RunningProcesses int       `json:"running_processes"`
	CPUUsage         float64   `json:"cpu_usage"`
	MemoryUsage      float64   `json:"memory_usage"`
	DiskUsage        float64   `json:"disk_usage"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}