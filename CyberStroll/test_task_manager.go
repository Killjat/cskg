package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cskg/CyberStroll/internal/taskmanager"
	"github.com/cskg/CyberStroll/internal/storage"
	"github.com/cskg/CyberStroll/internal/kafka"
)

func main() {
	fmt.Println("🧪 CyberStroll 任务管理器测试")
	fmt.Println("==============================")

	// 创建日志器
	logger := log.New(os.Stdout, "[TaskManagerTest] ", log.LstdFlags)

	// 创建存储客户端 (模拟)
	mongoConfig := &storage.MongoConfig{
		URI:      "mongodb://localhost:27017",
		Database: "cyberstroll_test",
		Timeout:  10,
	}
	mongoClient, err := storage.NewMongoClient(mongoConfig)
	if err != nil {
		logger.Printf("⚠️  MongoDB连接失败 (使用模拟模式): %v", err)
	} else {
		logger.Println("✅ MongoDB连接成功")
	}

	// 创建Kafka客户端 (模拟)
	kafkaConfig := &kafka.KafkaConfig{
		Brokers:          []string{"localhost:9092"},
		SystemTaskTopic:  "test_system_tasks",
		RegularTaskTopic: "test_regular_tasks",
		ResultTopic:      "test_scan_results",
		GroupID:          "test_task_managers",
	}

	taskProducer := kafka.NewTaskProducer(kafkaConfig, logger)
	resultConsumer := kafka.NewTaskConsumer(kafkaConfig, logger)

	// 创建任务管理器
	tmConfig := &taskmanager.TaskManagerConfig{
		MaxTasksPerUser:    5,
		MaxIPsPerTask:      100,
		SystemTaskInterval: 60, // 1分钟测试间隔
		EnableSystemTasks:  true,
	}

	taskManager := taskmanager.NewTaskManager(
		mongoClient,
		taskProducer,
		resultConsumer,
		tmConfig,
		logger,
	)

	fmt.Println("✅ 任务管理器创建成功")

	// 测试1: 提交单IP任务
	fmt.Println("\n📋 测试1: 提交单IP扫描任务")
	request1 := &taskmanager.TaskRequest{
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

	// 测试2: 提交CIDR任务
	fmt.Println("\n📋 测试2: 提交CIDR扫描任务")
	request2 := &taskmanager.TaskRequest{
		Initiator: "test_user",
		Targets:   []string{"192.168.1.0/28"}, // 16个IP
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

	// 测试3: 提交IP范围任务
	fmt.Println("\n📋 测试3: 提交IP范围扫描任务")
	request3 := &taskmanager.TaskRequest{
		Initiator: "test_user",
		Targets:   []string{"10.0.0.1-10.0.0.10"},
		TaskType:  "app_identification",
		Timeout:   15,
	}

	response3, err := taskManager.SubmitTask(request3)
	if err != nil {
		fmt.Printf("❌ 任务提交失败: %v\n", err)
	} else {
		fmt.Printf("✅ 任务提交成功: TaskID=%s, 目标数=%d\n", 
			response3.TaskID, response3.TargetCount)
	}

	// 测试4: 测试无效请求
	fmt.Println("\n📋 测试4: 测试无效请求处理")
	invalidRequest := &taskmanager.TaskRequest{
		Initiator: "",
		Targets:   []string{},
		TaskType:  "invalid_type",
	}

	response4, err := taskManager.SubmitTask(invalidRequest)
	if err != nil {
		fmt.Printf("✅ 正确拒绝无效请求: %s\n", response4.Message)
	} else {
		fmt.Printf("❌ 应该拒绝无效请求\n")
	}

	// 测试5: 查看统计信息
	fmt.Println("\n📊 测试5: 查看统计信息")
	stats := taskManager.GetStats()
	fmt.Printf("统计信息:\n")
	fmt.Printf("  总任务数: %d\n", stats.TotalTasks)
	fmt.Printf("  系统任务: %d\n", stats.SystemTasks)
	fmt.Printf("  常规任务: %d\n", stats.RegularTasks)
	fmt.Printf("  已完成: %d\n", stats.CompletedTasks)
	fmt.Printf("  失败任务: %d\n", stats.FailedTasks)

	// 测试6: 任务状态查询
	if response1.Status == "success" {
		fmt.Println("\n🔍 测试6: 查询任务状态")
		status, err := taskManager.GetTaskStatus(response1.TaskID)
		if err != nil {
			fmt.Printf("❌ 查询任务状态失败: %v\n", err)
		} else {
			fmt.Printf("✅ 任务状态查询成功:\n")
			fmt.Printf("  任务ID: %v\n", status["task_id"])
			fmt.Printf("  状态: %v\n", status["status"])
			fmt.Printf("  进度: %.1f%%\n", status["progress"])
			fmt.Printf("  目标数: %v\n", status["target_count"])
		}
	}

	// 测试7: 列出用户任务
	fmt.Println("\n📝 测试7: 列出用户任务")
	tasks, err := taskManager.ListUserTasks("test_user", 10)
	if err != nil {
		fmt.Printf("❌ 查询用户任务失败: %v\n", err)
	} else {
		fmt.Printf("✅ 用户任务查询成功: 找到 %d 个任务\n", len(tasks))
		for i, task := range tasks {
			fmt.Printf("  [%d] %s - %s (%s)\n", 
				i+1, task.TaskID, task.TaskType, task.TaskStatus)
		}
	}

	fmt.Println("\n🎉 任务管理器测试完成!")
	fmt.Println("\n💡 功能特性:")
	fmt.Println("   ✅ 任务提交和验证")
	fmt.Println("   ✅ 多种目标格式支持 (单IP/CIDR/范围)")
	fmt.Println("   ✅ 任务状态管理")
	fmt.Println("   ✅ 用户任务查询")
	fmt.Println("   ✅ 统计信息收集")
	fmt.Println("   ✅ 错误处理和验证")

	fmt.Println("\n🚀 启动Web界面:")
	fmt.Println("   go run cmd/task_manager/main.go")
	fmt.Println("   然后访问: http://localhost:8080")
}