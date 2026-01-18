package collector

import (
	"github.com/example/server-monitor/internal/model"
)

// Collector 采集器通用接口
type Collector interface {
	Start() error
	Stop() error
}

// LoginCollector 登录信息采集器接口
type LoginCollector interface {
	Collector
	CollectCurrentLogins() ([]model.LoginData, error)
	CollectLoginHistory() ([]model.LoginData, error)
}

// FileCollector 文件操作采集器接口
type FileCollector interface {
	Collector
	CollectFileOperations() ([]model.FileOperationData, error)
	AddWatch(path string, recursive bool) error
	RemoveWatch(path string) error
}

// ProcessCollector 进程信息采集器接口
type ProcessCollector interface {
	Collector
	CollectProcesses() ([]model.ProcessData, error)
	GetProcessByPID(pid int) (model.ProcessData, error)
	GetProcessCount() (int, error)
	GetRunningProcessCount() (int, error)
}

// CommandCollector 命令信息采集器接口
type CommandCollector interface {
	Collector
	CollectCurrentCommands() ([]model.CommandData, error)
	CollectCommandHistory() ([]model.CommandData, error)
}

// SystemStatsCollector 系统统计信息采集器接口
type SystemStatsCollector interface {
	Collector
	CollectSystemStats() (model.SystemStats, error)
}