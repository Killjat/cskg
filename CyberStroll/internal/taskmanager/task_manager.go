package taskmanager

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/cskg/CyberStroll/internal/kafka"
	"github.com/cskg/CyberStroll/internal/storage"
)

// TaskManager 任务管理器
type TaskManager struct {
	storage       *storage.MongoClient
	taskProducer  *kafka.TaskProducer
	resultConsumer *kafka.TaskConsumer
	logger        *log.Logger
	config        *TaskManagerConfig
	stats         *TaskManagerStats
	mutex         sync.RWMutex
}

// TaskManagerConfig 任务管理器配置
type TaskManagerConfig struct {
	MaxTasksPerUser    int `yaml:"max_tasks_per_user"`
	MaxIPsPerTask      int `yaml:"max_ips_per_task"`
	SystemTaskInterval int `yaml:"system_task_interval"` // 秒
	EnableSystemTasks  bool `yaml:"enable_system_tasks"`
}

// TaskManagerStats 任务管理器统计
type TaskManagerStats struct {
	TotalTasks       int64 `json:"total_tasks"`
	SystemTasks      int64 `json:"system_tasks"`
	RegularTasks     int64 `json:"regular_tasks"`
	CompletedTasks   int64 `json:"completed_tasks"`
	FailedTasks      int64 `json:"failed_tasks"`
	ActiveTasks      int64 `json:"active_tasks"`
	LastTaskTime     int64 `json:"last_task_time"`
}

// TaskRequest 任务请求
type TaskRequest struct {
	Initiator string   `json:"initiator"`
	Targets   []string `json:"targets"`
	TaskType  string   `json:"task_type"`
	Ports     []int    `json:"ports,omitempty"`
	Timeout   int      `json:"timeout,omitempty"`
}

// TaskResponse 任务响应
type TaskResponse struct {
	TaskID      string `json:"task_id"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	TargetCount int    `json:"target_count"`
}

// NewTaskManager 创建任务管理器
func NewTaskManager(
	storage *storage.MongoClient,
	taskProducer *kafka.TaskProducer,
	resultConsumer *kafka.TaskConsumer,
	config *TaskManagerConfig,
	logger *log.Logger,
) *TaskManager {
	if config == nil {
		config = &TaskManagerConfig{
			MaxTasksPerUser:    10,
			MaxIPsPerTask:      3000,
			SystemTaskInterval: 300, // 5分钟
			EnableSystemTasks:  true,
		}
	}

	return &TaskManager{
		storage:        storage,
		taskProducer:   taskProducer,
		resultConsumer: resultConsumer,
		logger:         logger,
		config:         config,
		stats:          &TaskManagerStats{},
	}
}

// SubmitTask 提交任务
func (tm *TaskManager) SubmitTask(request *TaskRequest) (*TaskResponse, error) {
	// 验证请求
	if err := tm.validateTaskRequest(request); err != nil {
		return &TaskResponse{
			Status:  "error",
			Message: err.Error(),
		}, err
	}

	// 解析目标IP
	ips, err := tm.parseTargets(request.Targets)
	if err != nil {
		return &TaskResponse{
			Status:  "error",
			Message: fmt.Sprintf("解析目标失败: %v", err),
		}, err
	}

	// 检查IP数量限制
	if len(ips) > tm.config.MaxIPsPerTask {
		return &TaskResponse{
			Status:  "error",
			Message: fmt.Sprintf("目标IP数量超过限制 (%d > %d)", len(ips), tm.config.MaxIPsPerTask),
		}, fmt.Errorf("IP数量超限")
	}

	// 生成任务ID
	taskID := uuid.New().String()

	// 判断任务类别：如果目标包含URL，则作为系统任务
	taskCategory := "regular_task"
	priority := 1 // 常规任务优先级
	for _, target := range request.Targets {
		if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
			taskCategory = "system_task"
			priority = 10 // 系统任务高优先级
			break
		}
	}

	// 创建任务记录
	task := &storage.Task{
		TaskID:        taskID,
		TaskInitiator: request.Initiator,
		TaskTarget:    strings.Join(request.Targets, ","),
		TaskType:      request.TaskType,
		TaskCategory:  taskCategory,
		TaskStatus:    "pending",
		TargetCount:   len(ips),
		Config: storage.TaskConfig{
			Ports:   request.Ports,
			Timeout: request.Timeout,
		},
	}

	// 保存任务到数据库
	if err := tm.storage.CreateTask(task); err != nil {
		tm.logger.Printf("保存任务失败: %v", err)
		return &TaskResponse{
			Status:  "error",
			Message: "保存任务失败",
		}, err
	}

	// 分发任务到Kafka
	if err := tm.distributeTask(taskID, ips, request); err != nil {
		tm.logger.Printf("分发任务失败: %v", err)
		// 更新任务状态为失败
		tm.storage.UpdateTaskStatus(taskID, "failed", "0", "任务分发失败")
		return &TaskResponse{
			Status:  "error",
			Message: "分发任务失败",
		}, err
	}

	// 更新任务状态为运行中
	tm.storage.UpdateTaskStatus(taskID, "running", "0", "任务开始执行")

	// 更新统计
	tm.updateStats("regular_task")

	tm.logger.Printf("任务提交成功: TaskID=%s, Initiator=%s, Targets=%d", 
		taskID, request.Initiator, len(ips))

	return &TaskResponse{
		TaskID:      taskID,
		Status:      "success",
		Message:     "任务提交成功",
		TargetCount: len(ips),
	}, nil
}

// GetTaskStatus 获取任务状态
func (tm *TaskManager) GetTaskStatus(taskID string) (map[string]interface{}, error) {
	task, err := tm.storage.GetTask(taskID)
	if err != nil {
		return nil, err
	}

	status := map[string]interface{}{
		"task_id":        task.TaskID,
		"status":         task.TaskStatus,
		"progress":       task.Progress,
		"target_count":   task.TargetCount,
		"completed_count": task.CompletedCount,
		"failed_count":   task.FailedCount,
		"created_time":   task.CreatedTime,
		"started_time":   task.StartedTime,
		"completed_time": task.CompletedTime,
	}

	return status, nil
}

// ListUserTasks 列出用户任务
func (tm *TaskManager) ListUserTasks(initiator string, limit int) ([]*storage.Task, error) {
	return tm.storage.ListTasks(initiator, "", int64(limit))
}

// validateTaskRequest 验证任务请求
func (tm *TaskManager) validateTaskRequest(request *TaskRequest) error {
	if request.Initiator == "" {
		return fmt.Errorf("任务发起人不能为空")
	}

	if len(request.Targets) == 0 {
		return fmt.Errorf("目标不能为空")
	}

	validTaskTypes := map[string]bool{
		"port_scan_specified": true,
		"port_scan_default":   true,
		"port_scan_full":      true,
		"app_identification":  true,
	}

	if !validTaskTypes[request.TaskType] {
		return fmt.Errorf("无效的任务类型: %s", request.TaskType)
	}

	// 如果是指定端口扫描，必须提供端口列表
	if request.TaskType == "port_scan_specified" && len(request.Ports) == 0 {
		return fmt.Errorf("指定端口扫描必须提供端口列表")
	}

	return nil
}

// parseTargets 解析目标
func (tm *TaskManager) parseTargets(targets []string) ([]string, error) {
	var ips []string

	for _, target := range targets {
		target = strings.TrimSpace(target)
		
		// 检查是否是URL格式
		if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
			resolvedIPs, err := tm.parseURL(target)
			if err != nil {
				return nil, fmt.Errorf("解析URL %s 失败: %v", target, err)
			}
			ips = append(ips, resolvedIPs...)
		} else if strings.Contains(target, "/") {
			// 检查是否是CIDR格式
			cidrIPs, err := tm.parseCIDR(target)
			if err != nil {
				return nil, fmt.Errorf("解析CIDR %s 失败: %v", target, err)
			}
			ips = append(ips, cidrIPs...)
		} else if strings.Contains(target, "-") {
			// IP范围格式 (如: 192.168.1.1-192.168.1.100)
			rangeIPs, err := tm.parseIPRange(target)
			if err != nil {
				return nil, fmt.Errorf("解析IP范围 %s 失败: %v", target, err)
			}
			ips = append(ips, rangeIPs...)
		} else {
			// 单个IP或域名
			if net.ParseIP(target) != nil {
				// 是IP地址
				ips = append(ips, target)
			} else {
				// 可能是域名，尝试解析
				resolvedIPs, err := tm.resolveDomain(target)
				if err != nil {
					return nil, fmt.Errorf("解析域名 %s 失败: %v", target, err)
				}
				ips = append(ips, resolvedIPs...)
			}
		}
	}

	return ips, nil
}

// parseCIDR 解析CIDR
func (tm *TaskManager) parseCIDR(cidr string) ([]string, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); tm.incrementIP(ip) {
		ips = append(ips, ip.String())
		
		// 防止生成过多IP
		if len(ips) > tm.config.MaxIPsPerTask {
			break
		}
	}

	return ips, nil
}

// parseIPRange 解析IP范围
func (tm *TaskManager) parseIPRange(ipRange string) ([]string, error) {
	parts := strings.Split(ipRange, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("无效的IP范围格式")
	}

	startIP := net.ParseIP(strings.TrimSpace(parts[0]))
	endIP := net.ParseIP(strings.TrimSpace(parts[1]))

	if startIP == nil || endIP == nil {
		return nil, fmt.Errorf("无效的IP地址")
	}

	var ips []string
	for ip := startIP; !ip.Equal(endIP); tm.incrementIP(ip) {
		ips = append(ips, ip.String())
		
		// 防止生成过多IP
		if len(ips) > tm.config.MaxIPsPerTask {
			break
		}
	}
	ips = append(ips, endIP.String()) // 包含结束IP

	return ips, nil
}

// incrementIP IP自增
func (tm *TaskManager) incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// parseURL 解析URL并获取IP地址
func (tm *TaskManager) parseURL(urlStr string) ([]string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("无效的URL格式: %v", err)
	}

	host := parsedURL.Hostname()
	if host == "" {
		return nil, fmt.Errorf("无法从URL中提取主机名")
	}

	tm.logger.Printf("从URL %s 提取主机名: %s", urlStr, host)

	// 检查是否已经是IP地址
	if net.ParseIP(host) != nil {
		return []string{host}, nil
	}

	// 解析域名
	return tm.resolveDomain(host)
}

// resolveDomain 解析域名到IP地址
func (tm *TaskManager) resolveDomain(domain string) ([]string, error) {
	tm.logger.Printf("正在解析域名: %s", domain)

	// 使用DNS解析域名
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, fmt.Errorf("DNS解析失败: %v", err)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("域名 %s 没有解析到任何IP地址", domain)
	}

	var ipStrings []string
	for _, ip := range ips {
		// 只取IPv4地址
		if ip.To4() != nil {
			ipStrings = append(ipStrings, ip.String())
		}
	}

	if len(ipStrings) == 0 {
		return nil, fmt.Errorf("域名 %s 没有解析到IPv4地址", domain)
	}

	tm.logger.Printf("域名 %s 解析到IP地址: %v", domain, ipStrings)
	return ipStrings, nil
}

// distributeTask 分发任务
func (tm *TaskManager) distributeTask(taskID string, ips []string, request *TaskRequest) error {
	var tasks []*kafka.Task

	for _, ip := range ips {
		task := &kafka.Task{
			TaskID:   taskID,
			IP:       ip,
			TaskType: request.TaskType,
			Priority: 1, // 常规任务优先级为1
			User:     request.Initiator,
			Config: map[string]interface{}{
				"ports":   request.Ports,
				"timeout": request.Timeout,
			},
			Timestamp: time.Now().Unix(),
		}

		tasks = append(tasks, task)
	}

	// 批量发送任务
	ctx := context.Background()
	return tm.taskProducer.SendBatchTasks(ctx, tasks, false) // false表示常规任务
}

// StartSystemTaskGenerator 启动系统任务生成器
func (tm *TaskManager) StartSystemTaskGenerator() {
	if !tm.config.EnableSystemTasks {
		tm.logger.Println("系统任务生成器已禁用")
		return
	}

	tm.logger.Println("启动系统任务生成器...")
	
	ticker := time.NewTicker(time.Duration(tm.config.SystemTaskInterval) * time.Second)
	go func() {
		for range ticker.C {
			tm.generateSystemTasks()
		}
	}()
}

// generateSystemTasks 生成系统任务
func (tm *TaskManager) generateSystemTasks() {
	tm.logger.Println("生成系统任务...")

	// TODO: 从系统IP池获取IP列表
	// 这里先用示例IP
	systemIPs := []string{
		"8.8.8.8",
		"1.1.1.1",
		"114.114.114.114",
	}

	if len(systemIPs) == 0 {
		tm.logger.Println("系统IP池为空，跳过系统任务生成")
		return
	}

	// 生成任务ID
	taskID := "system-" + uuid.New().String()

	// 创建系统任务记录
	task := &storage.Task{
		TaskID:        taskID,
		TaskInitiator: "system",
		TaskTarget:    strings.Join(systemIPs, ","),
		TaskType:      "port_scan_default",
		TaskCategory:  "system_task",
		TaskStatus:    "pending",
		TargetCount:   len(systemIPs),
		Config: storage.TaskConfig{
			Timeout: 10,
		},
	}

	// 保存任务
	if err := tm.storage.CreateTask(task); err != nil {
		tm.logger.Printf("保存系统任务失败: %v", err)
		return
	}

	// 分发系统任务
	var tasks []*kafka.Task
	for _, ip := range systemIPs {
		kafkaTask := &kafka.Task{
			TaskID:   taskID,
			IP:       ip,
			TaskType: "port_scan_default",
			Priority: 10, // 系统任务高优先级
			Config: map[string]interface{}{
				"timeout": 10,
			},
			Timestamp: time.Now().Unix(),
		}
		tasks = append(tasks, kafkaTask)
	}

	ctx := context.Background()
	if err := tm.taskProducer.SendBatchTasks(ctx, tasks, true); err != nil { // true表示系统任务
		tm.logger.Printf("分发系统任务失败: %v", err)
		tm.storage.UpdateTaskStatus(taskID, "failed", "0", "系统任务分发失败")
		return
	}

	// 更新任务状态
	tm.storage.UpdateTaskStatus(taskID, "running", "0", "系统任务开始执行")

	// 更新统计
	tm.updateStats("system_task")

	tm.logger.Printf("系统任务生成成功: TaskID=%s, IPs=%d", taskID, len(systemIPs))
}

// updateStats 更新统计
func (tm *TaskManager) updateStats(taskType string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tm.stats.TotalTasks++
	if taskType == "system_task" {
		tm.stats.SystemTasks++
	} else {
		tm.stats.RegularTasks++
	}
	tm.stats.LastTaskTime = time.Now().Unix()
}

// GetStats 获取统计信息
func (tm *TaskManager) GetStats() *TaskManagerStats {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	return &TaskManagerStats{
		TotalTasks:     tm.stats.TotalTasks,
		SystemTasks:    tm.stats.SystemTasks,
		RegularTasks:   tm.stats.RegularTasks,
		CompletedTasks: tm.stats.CompletedTasks,
		FailedTasks:    tm.stats.FailedTasks,
		ActiveTasks:    tm.stats.ActiveTasks,
		LastTaskTime:   tm.stats.LastTaskTime,
	}
}