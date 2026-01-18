package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cskg/CyberStroll/internal/config"
	"github.com/cskg/CyberStroll/internal/elasticsearch"
	"github.com/cskg/CyberStroll/internal/kafka"
	"github.com/cskg/CyberStroll/internal/scan"
	"github.com/cskg/CyberStroll/pkg/models"
)

func main() {
	fmt.Println("=== CyberStroll Scan Node ===")
	fmt.Println("扫描节点模块 - 负责读取Kafka任务并执行扫描")
	fmt.Println()

	// 解析命令行参数
	var configPath string
	flag.StringVar(&configPath, "config", "./config", "配置文件路径")

	help := flag.Bool("help", false, "显示帮助信息")

	flag.Parse()

	if *help {
		showHelp()
		os.Exit(0)
	}

	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("错误: 加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 创建上下文，支持优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理信号，实现优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n收到退出信号，正在关闭扫描节点...")
		cancel()
	}()

	// 创建Elasticsearch客户端
	esClient, err := elasticsearch.NewClient(&cfg.Elasticsearch)
	if err != nil {
		fmt.Printf("错误: 创建Elasticsearch客户端失败: %v\n", err)
		os.Exit(1)
	}

	// 创建扫描器
	scanner := scan.NewScanner(&cfg.Scan)

	// 启动任务处理循环
fmt.Println("扫描节点已启动，等待任务...")

// 并发处理系统任务和普通任务
go handleTasks(ctx, cfg, esClient, scanner, cfg.Kafka.Topics.SystemTask)
go handleTasks(ctx, cfg, esClient, scanner, cfg.Kafka.Topics.NormalTask)

// 等待上下文取消
<-ctx.Done()

	fmt.Println("扫描节点已关闭")
}

// handleTasks 处理Kafka任务
func handleTasks(ctx context.Context, cfg *config.Config, esClient *elasticsearch.Client, scanner *scan.Scanner, topic string) {
	// 创建Kafka消费者
	consumer, err := kafka.NewConsumer(&cfg.Kafka, topic)
	if err != nil {
		fmt.Printf("错误: 创建Kafka消费者失败: %v\n", err)
		return
	}
	defer consumer.Close()

	// 任务处理循环
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 读取任务
			task, err := consumer.ReadTask(ctx)
			if err != nil {
				fmt.Printf("错误: 读取Kafka任务失败: %v\n", err)
				continue
			}

			fmt.Printf("\n收到任务:\n")
			fmt.Printf("  任务ID: %s\n", task.ID)
			fmt.Printf("  任务类型: %s\n", task.Type)
			fmt.Printf("  协议类型: %s\n", task.Protocol)
			fmt.Printf("  目标IP: %v\n", task.Targets)
			fmt.Printf("  端口范围: %s\n", task.PortRange)

			// 执行扫描
			results, err := scanner.Scan(ctx, task)
			if err != nil {
				fmt.Printf("错误: 执行扫描任务失败: %v\n", err)
				continue
			}

			fmt.Printf("扫描完成，共获得 %d 个结果\n", len(results))

			// 存储扫描结果到Elasticsearch
			for _, result := range results {
				if err := esClient.StoreScanResult(ctx, result); err != nil {
					fmt.Printf("错误: 存储扫描结果失败: %v\n", err)
					continue
				}
				fmt.Printf("✓ 已存储结果: %s:%d %s %s\n", result.IP, result.Port, result.Protocol, result.Status)
			}

			// 更新任务状态为已完成
			// TODO: 实现任务状态更新逻辑
			task.Status = models.TaskStatusCompleted
			now := time.Now()
			task.CompletedAt = &now

			fmt.Printf("任务 %s 处理完成\n", task.ID)
		}
	}
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Println("使用方法:")
	fmt.Println("  scan_node [选项]")
	fmt.Println()
	fmt.Println("选项:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  scan_node -config ./config")
}
