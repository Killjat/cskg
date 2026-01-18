package main

import (
	"bufio"
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
	fmt.Println("=== CyberStroll MQTT 扫描工具 ===")
	fmt.Println("读取 mqttip.txt 文件并下发应用识别扫描任务")
	fmt.Println()

	// 解析命令行参数
	var (
		configPath  string
		creator     string
		protocol    string
		portRange   string
		filePath    string
		systemTask  bool
	)

	flag.StringVar(&configPath, "config", "./config", "配置文件路径")
	flag.StringVar(&creator, "creator", "admin", "任务发起人")
	flag.StringVar(&protocol, "protocol", "tcp", "协议类型: tcp, udp, all")
	flag.StringVar(&portRange, "port", "1883,8883", "MQTT默认端口范围")
	flag.StringVar(&filePath, "file", "mqttip.txt", "MQTT IP地址文件路径")
	flag.BoolVar(&systemTask, "system", false, "是否为系统任务")

	help := flag.Bool("help", false, "显示帮助信息")

	flag.Parse()

	if *help {
		showHelp()
		os.Exit(0)
	}

	// 读取mqttip.txt文件
	ips, err := readIPsFromFile(filePath)
	if err != nil {
		fmt.Printf("错误: 读取文件失败: %v\n", err)
		os.Exit(1)
	}

	if len(ips) == 0 {
		fmt.Println("错误: 文件中没有找到IP地址")
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

	// 为每个IP地址创建应用识别扫描任务
	taskCount := 0
	for _, ip := range ips {
		// 去除IP地址前后的空格
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}

		// 创建应用识别扫描任务
		createdTask, err := taskManager.CreateTask(
			ctx,
			creator,
			models.TaskTypeServiceScan, // 应用识别扫描任务类型
			models.ProtocolType(protocol),
			[]string{ip}, // 单个IP地址
			portRange,
			systemTask,
		)

		if err != nil {
			fmt.Printf("错误: 创建任务失败 (IP: %s): %v\n", ip, err)
			continue
		}

		taskCount++
		fmt.Printf("✓ 任务创建成功! (IP: %s, 任务ID: %s)\n", ip, createdTask.ID)
	}

	// 输出任务创建结果统计
	fmt.Println()
	fmt.Printf("=== 任务创建完成 ===\n")
	fmt.Printf("总IP地址数: %d\n", len(ips))
	fmt.Printf("成功创建任务数: %d\n", taskCount)
	fmt.Printf("失败创建任务数: %d\n", len(ips)-taskCount)
}

// readIPsFromFile 从文件中读取IP地址列表
func readIPsFromFile(filePath string) ([]string, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 读取文件内容
	var ips []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ip := scanner.Text()
		if ip != "" {
			ips = append(ips, ip)
		}
	}

	// 检查扫描过程中是否有错误
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ips, nil
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Println("使用方法:")
	fmt.Println("  mqtt_scan [选项]")
	fmt.Println()
	fmt.Println("选项:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  mqtt_scan")
	fmt.Println("  mqtt_scan -file ./mqttip.txt -port 1883,8883")
	fmt.Println("  mqtt_scan -config ./config -creator user1 -protocol tcp -system")
}
