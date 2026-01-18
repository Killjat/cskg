package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cskg/CyberStroll/internal/config"
	"github.com/cskg/CyberStroll/internal/kafka"
	"github.com/cskg/CyberStroll/internal/task"
	"github.com/cskg/CyberStroll/pkg/models"
)

func main() {
	fmt.Println("=== CyberStroll Task Manager ===")
	fmt.Println("任务管理模块 - 负责下发扫描任务到Kafka")
	fmt.Println()

	// 解析命令行参数
	var (
		configPath  string
		creator     string
		taskType    string
		protocol    string
		targets     string
		portRange   string
		systemTask  bool
	)

	flag.StringVar(&configPath, "config", "./config", "配置文件路径")
	flag.StringVar(&creator, "creator", "admin", "任务发起人")
	flag.StringVar(&taskType, "type", "port_scan", "任务类型: ip_alive, port_scan, service_scan, web_scan, banner_grab")
	flag.StringVar(&protocol, "protocol", "tcp", "协议类型: tcp, udp, all")
	flag.StringVar(&targets, "targets", "", "目标IP列表，用逗号分隔")
	flag.StringVar(&portRange, "port", "80,443,8080", "端口范围，如: 80,443,8080 或 1-1000")
	flag.BoolVar(&systemTask, "system", false, "是否为系统任务")

	help := flag.Bool("help", false, "显示帮助信息")

	flag.Parse()

	if *help {
		showHelp()
		os.Exit(0)
	}

	if targets == "" {
		fmt.Println("错误: 目标IP列表不能为空")
		showHelp()
		os.Exit(1)
	}

	// 验证参数
	if !isValidTaskType(taskType) {
		fmt.Printf("错误: 无效的任务类型: %s\n", taskType)
		showHelp()
		os.Exit(1)
	}

	if !isValidProtocol(protocol) {
		fmt.Printf("错误: 无效的协议类型: %s\n", protocol)
		showHelp()
		os.Exit(1)
	}

	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("错误: 加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 创建Kafka生产者
	producer, err := kafka.NewProducer(&cfg.Kafka)
	if err != nil {
		fmt.Printf("错误: 创建Kafka生产者失败: %v\n", err)
		os.Exit(1)
	}
	defer producer.Close()

	// 创建任务管理器
	taskManager := task.NewManager(cfg, producer)

	// 创建上下文
	ctx := context.Background()

	// 解析目标IP列表
	targetList := strings.Split(targets, ",")

	// 创建任务
	createdTask, err := taskManager.CreateTask(
		ctx,
		creator,
		models.TaskType(taskType),
		models.ProtocolType(protocol),
		targetList,
		portRange,
		systemTask,
	)

	if err != nil {
		fmt.Printf("错误: 创建任务失败: %v\n", err)
		os.Exit(1)
	}

	// 输出任务创建结果
	fmt.Printf("✓ 任务创建成功!\n")
	fmt.Printf("任务ID: %s\n", createdTask.ID)
	fmt.Printf("任务类型: %s\n", createdTask.Type)
	fmt.Printf("协议类型: %s\n", createdTask.Protocol)
	fmt.Printf("目标IP: %v\n", createdTask.Targets)
	fmt.Printf("端口范围: %s\n", createdTask.PortRange)
	fmt.Printf("任务状态: %s\n", createdTask.Status)
	fmt.Printf("创建时间: %s\n", createdTask.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("是否系统任务: %v\n", systemTask)
	fmt.Printf("任务已下发到Kafka主题: %s\n", func() string {
		if systemTask {
			return cfg.Kafka.Topics.SystemTask
		}
		return cfg.Kafka.Topics.NormalTask
	}())
}

// isValidTaskType 验证任务类型是否有效
func isValidTaskType(taskType string) bool {
	validTypes := []string{
		string(models.TaskTypeIPAlive),
		string(models.TaskTypePortScan),
		string(models.TaskTypeServiceScan),
		string(models.TaskTypeWebScan),
		string(models.TaskTypeBannerGrab),
	}

	for _, validType := range validTypes {
		if taskType == validType {
			return true
		}
	}
	return false
}

// isValidProtocol 验证协议类型是否有效
func isValidProtocol(protocol string) bool {
	validProtocols := []string{
		string(models.ProtocolTCP),
		string(models.ProtocolUDP),
		string(models.ProtocolAll),
	}

	for _, validProtocol := range validProtocols {
		if protocol == validProtocol {
			return true
		}
	}
	return false
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Println("使用方法:")
	fmt.Println("  task_manager [选项]")
	fmt.Println()
	fmt.Println("选项:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  task_manager -targets 192.168.1.1,192.168.1.2 -type port_scan -protocol tcp -port 80,443")
	fmt.Println("  task_manager -targets 10.0.0.1 -type ip_alive -system")
	fmt.Println("  task_manager -targets example.com -type web_scan -port 80,443")
}
