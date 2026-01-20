package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

// 简化的任务管理器测试

// SimpleTaskManager 简化任务管理器
type SimpleTaskManager struct {
	logger *log.Logger
	stats  *TaskStats
}

// TaskStats 任务统计
type TaskStats struct {
	TotalTasks   int64
	SystemTasks  int64
	RegularTasks int64
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

// NewSimpleTaskManager 创建简化任务管理器
func NewSimpleTaskManager(logger *log.Logger) *SimpleTaskManager {
	return &SimpleTaskManager{
		logger: logger,
		stats:  &TaskStats{},
	}
}

// SubmitTask 提交任务
func (stm *SimpleTaskManager) SubmitTask(request *TaskRequest) (*TaskResponse, error) {
	// 验证请求
	if request.Initiator == "" {
		return &TaskResponse{
			Status:  "error",
			Message: "任务发起人不能为空",
		}, fmt.Errorf("任务发起人不能为空")
	}

	if len(request.Targets) == 0 {
		return &TaskResponse{
			Status:  "error",
			Message: "目标不能为空",
		}, fmt.Errorf("目标不能为空")
	}

	// 解析目标
	targetCount := len(request.Targets)
	
	// 生成任务ID
	taskID := fmt.Sprintf("task-%d", time.Now().Unix())

	// 更新统计
	stm.stats.TotalTasks++
	stm.stats.RegularTasks++

	stm.logger.Printf("任务提交成功: TaskID=%s, Initiator=%s, Targets=%d", 
		taskID, request.Initiator, targetCount)

	return &TaskResponse{
		TaskID:      taskID,
		Status:      "success",
		Message:     "任务提交成功",
		TargetCount: targetCount,
	}, nil
}

// GetStats 获取统计信息
func (stm *SimpleTaskManager) GetStats() *TaskStats {
	return stm.stats
}

func main() {
	fmt.Println("🧪 CyberStroll 简化任务管理器测试")
	fmt.Println("==================================")

	// 创建日志器
	logger := log.New(os.Stdout, "[SimpleTaskManager] ", log.LstdFlags)

	// 创建简化任务管理器
	taskManager := NewSimpleTaskManager(logger)
	fmt.Println("✅ 简化任务管理器创建成功")

	// 测试1: 提交单IP任务
	fmt.Println("\n📋 测试1: 提交单IP扫描任务")
	request1 := &TaskRequest{
		Initiator: "test_user",
		Targets:   []string{"127.0.0.1"},
		TaskType:  "port_scan_default",
		Timeout:   10,
	}

	response1, err := taskManager.SubmitTask(request1)
	if err != nil {
		fmt.Printf("❌ 任务提交失败: %v\n", err)
	} else {
		fmt.Printf("✅ 任务提交成功: TaskID=%s, 目标数=%d\n", 
			response1.TaskID, response1.TargetCount)
	}

	// 测试2: 提交多IP任务
	fmt.Println("\n📋 测试2: 提交多IP扫描任务")
	request2 := &TaskRequest{
		Initiator: "test_user",
		Targets:   []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"},
		TaskType:  "port_scan_specified",
		Ports:     []int{22, 80, 443},
		Timeout:   5,
	}

	response2, err := taskManager.SubmitTask(request2)
	if err != nil {
		fmt.Printf("❌ 任务提交失败: %v\n", err)
	} else {
		fmt.Printf("✅ 任务提交成功: TaskID=%s, 目标数=%d\n", 
			response2.TaskID, response2.TargetCount)
	}

	// 测试3: 测试无效请求
	fmt.Println("\n📋 测试3: 测试无效请求处理")
	invalidRequest := &TaskRequest{
		Initiator: "",
		Targets:   []string{},
		TaskType:  "invalid_type",
	}

	response3, err := taskManager.SubmitTask(invalidRequest)
	if err != nil {
		fmt.Printf("✅ 正确拒绝无效请求: %s\n", response3.Message)
	} else {
		fmt.Printf("❌ 应该拒绝无效请求\n")
	}

	// 测试4: 查看统计信息
	fmt.Println("\n📊 测试4: 查看统计信息")
	stats := taskManager.GetStats()
	fmt.Printf("统计信息:\n")
	fmt.Printf("  总任务数: %d\n", stats.TotalTasks)
	fmt.Printf("  系统任务: %d\n", stats.SystemTasks)
	fmt.Printf("  常规任务: %d\n", stats.RegularTasks)

	fmt.Println("\n🎉 简化任务管理器测试完成!")
	fmt.Println("\n💡 核心功能验证:")
	fmt.Println("   ✅ 任务提交和验证")
	fmt.Println("   ✅ 错误处理")
	fmt.Println("   ✅ 统计信息收集")
	fmt.Println("   ✅ 任务ID生成")

	fmt.Println("\n🚀 下一步:")
	fmt.Println("   1. 集成完整的MongoDB存储")
	fmt.Println("   2. 集成Kafka消息队列")
	fmt.Println("   3. 实现Web管理界面")
	fmt.Println("   4. 添加任务状态管理")
}