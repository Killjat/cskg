package common

// TaskType 任务类型
type TaskType string

const (
	// TaskTypeScanIP IP扫描任务
	TaskTypeScanIP TaskType = "scan_ip"
	// TaskTypeScanPort 端口扫描任务
	TaskTypeScanPort TaskType = "scan_port"
	// TaskTypeScanService 服务识别任务
	TaskTypeScanService TaskType = "scan_service"
	// TaskTypeScanWeb Web站点识别任务
	TaskTypeScanWeb TaskType = "scan_web"
)

// Task 任务结构体
type Task struct {
	TaskID     string   `json:"task_id"`
	TaskType   TaskType `json:"task_type"`
	ClientID   string   `json:"client_id,omitempty"`
	Target     string   `json:"target"`
	PortRange  string   `json:"port_range,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Timestamp  int64    `json:"timestamp"`
}

// Result 扫描结果结构体
type Result struct {
	TaskID     string      `json:"task_id"`
	ClientID   string      `json:"client_id"`
	Target     string      `json:"target"`
	Port       int         `json:"port"`
	Protocol   string      `json:"protocol"`
	Service    string      `json:"service"`
	Version    string      `json:"version,omitempty"`
	WebInfo    *WebInfo    `json:"web_info,omitempty"`
	Status     string      `json:"status"`
	Error      string      `json:"error,omitempty"`
	Timestamp  int64       `json:"timestamp"`
}

// WebInfo Web站点信息结构体
type WebInfo struct {
	Title       string            `json:"title"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body,omitempty"`
	Fingerprint []string          `json:"fingerprint"`
	HasLogin    bool              `json:"has_login"`
	ICPInfo     *ICPInfo          `json:"icp_info,omitempty"`
	StatusCode  int               `json:"status_code"`
	URL         string            `json:"url"`
}

// ICPInfo ICP备案信息结构体
type ICPInfo struct {
	Domain      string `json:"domain"`
	ICP         string `json:"icp"`
	CompanyName string `json:"company_name"`
	Valid       bool   `json:"valid"`
}

// ClientStatus 客户端状态结构体
type ClientStatus struct {
	ClientID     string `json:"client_id"`
	Status       string `json:"status"`
	LastSeen     int64  `json:"last_seen"`
	ActiveTasks  int    `json:"active_tasks"`
	CompletedTasks int  `json:"completed_tasks"`
}
