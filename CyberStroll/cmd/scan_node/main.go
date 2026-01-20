package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cskg/CyberStroll/internal/kafka"
	"github.com/cskg/CyberStroll/internal/scanner"
	"github.com/cskg/CyberStroll/pkg/config"
)

// ScanNode 扫描节点
type ScanNode struct {
	config        *config.ScanNodeConfig
	consumer      *kafka.TaskConsumer
	producer      *kafka.ResultProducer
	probeEngine   *scanner.ProbeEngine
	logger        *log.Logger
	taskChan      chan *kafka.Task
	resultChan    chan *kafka.ScanResult
	workerPool    *WorkerPool
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// WorkerPool 工作池
type WorkerPool struct {
	workers    int
	taskChan   chan *kafka.Task
	resultChan chan *kafka.ScanResult
	engine     *scanner.ProbeEngine
	logger     *log.Logger
	wg         sync.WaitGroup
}

func main() {
	var (
		configFile = flag.String("config", "configs/scan_node.yaml", "配置文件路径")
		nodeID     = flag.String("node-id", "", "节点ID")
		workers    = flag.Int("workers", 10, "工作线程数")
		testMode   = flag.Bool("test", false, "测试模式")
	)
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadScanNodeConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 设置节点ID
	if *nodeID != "" {
		cfg.Node.ID = *nodeID
	}
	if *workers > 0 {
		cfg.Scanner.MaxConcurrency = *workers
	}

	// 创建日志器
	logger := log.New(os.Stdout, fmt.Sprintf("[ScanNode-%s] ", cfg.Node.ID), log.LstdFlags)

	// 测试模式
	if *testMode {
		runTestMode(cfg, logger)
		return
	}

	// 创建扫描节点
	node, err := NewScanNode(cfg, logger)
	if err != nil {
		log.Fatalf("创建扫描节点失败: %v", err)
	}

	// 启动节点
	logger.Println("启动扫描节点...")
	if err := node.Start(); err != nil {
		log.Fatalf("启动扫描节点失败: %v", err)
	}

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Println("收到退出信号，正在关闭...")
	node.Stop()
}

// NewScanNode 创建扫描节点
func NewScanNode(cfg *config.ScanNodeConfig, logger *log.Logger) (*ScanNode, error) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 创建Kafka消费者
	consumer := kafka.NewTaskConsumer(&cfg.Kafka, logger)

	// 创建Kafka生产者
	producer := kafka.NewResultProducer(&cfg.Kafka, logger)

	// 创建探测引擎
	probeEngine := scanner.NewProbeEngine(&cfg.Scanner)

	// 创建通道
	taskChan := make(chan *kafka.Task, 100)
	resultChan := make(chan *kafka.ScanResult, 100)

	// 创建工作池
	workerPool := &WorkerPool{
		workers:    cfg.Scanner.MaxConcurrency,
		taskChan:   taskChan,
		resultChan: resultChan,
		engine:     probeEngine,
		logger:     logger,
	}

	return &ScanNode{
		config:      cfg,
		consumer:    consumer,
		producer:    producer,
		probeEngine: probeEngine,
		logger:      logger,
		taskChan:    taskChan,
		resultChan:  resultChan,
		workerPool:  workerPool,
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// Start 启动扫描节点
func (sn *ScanNode) Start() error {
	sn.logger.Printf("扫描节点启动: NodeID=%s, Workers=%d", 
		sn.config.Node.ID, sn.config.Scanner.MaxConcurrency)

	// 启动工作池
	sn.wg.Add(1)
	go sn.workerPool.Start(sn.ctx, &sn.wg)

	// 启动任务消费协程
	sn.wg.Add(1)
	go sn.consumeTasks(&sn.wg)

	// 启动结果发送协程
	sn.wg.Add(1)
	go sn.sendResults(&sn.wg)

	// 启动统计协程
	sn.wg.Add(1)
	go sn.printStats(&sn.wg)

	return nil
}

// Stop 停止扫描节点
func (sn *ScanNode) Stop() {
	sn.logger.Println("正在停止扫描节点...")
	
	// 取消上下文
	sn.cancel()
	
	// 等待所有协程结束
	sn.wg.Wait()
	
	// 关闭资源
	sn.consumer.Close()
	sn.producer.Close()
	
	sn.logger.Println("扫描节点已停止")
}

// consumeTasks 消费任务
func (sn *ScanNode) consumeTasks(wg *sync.WaitGroup) {
	defer wg.Done()
	
	sn.logger.Println("开始消费任务...")
	
	for {
		select {
		case <-sn.ctx.Done():
			sn.logger.Println("停止消费任务")
			return
		default:
			// 消费任务 (优先处理系统任务)
			task, err := sn.consumer.ConsumeTask(sn.ctx)
			if err != nil {
				if err != context.DeadlineExceeded {
					sn.logger.Printf("消费任务失败: %v", err)
				}
				time.Sleep(time.Second)
				continue
			}
			
			// 发送到工作池
			select {
			case sn.taskChan <- task:
				// 任务发送成功
			case <-sn.ctx.Done():
				return
			}
		}
	}
}

// sendResults 发送结果
func (sn *ScanNode) sendResults(wg *sync.WaitGroup) {
	defer wg.Done()
	
	sn.logger.Println("开始发送结果...")
	
	for {
		select {
		case <-sn.ctx.Done():
			sn.logger.Println("停止发送结果")
			return
		case result := <-sn.resultChan:
			// 发送结果到Kafka
			if err := sn.producer.SendResult(sn.ctx, result); err != nil {
				sn.logger.Printf("发送结果失败: %v", err)
			}
		}
	}
}

// printStats 打印统计信息
func (sn *ScanNode) printStats(wg *sync.WaitGroup) {
	defer wg.Done()
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-sn.ctx.Done():
			return
		case <-ticker.C:
			stats := sn.probeEngine.GetStats()
			sn.logger.Printf("统计信息: 总扫描=%d, 成功=%d, 失败=%d, 平均时间=%dms", 
				stats.TotalScans, stats.SuccessScans, stats.FailedScans, stats.AverageTime)
		}
	}
}

// Start 启动工作池
func (wp *WorkerPool) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	
	wp.logger.Printf("启动工作池: Workers=%d", wp.workers)
	
	// 启动工作协程
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
	
	// 等待所有工作协程结束
	wp.wg.Wait()
	wp.logger.Println("工作池已停止")
}

// worker 工作协程
func (wp *WorkerPool) worker(ctx context.Context, workerID int) {
	defer wp.wg.Done()
	
	wp.logger.Printf("启动工作协程: WorkerID=%d", workerID)
	
	for {
		select {
		case <-ctx.Done():
			wp.logger.Printf("停止工作协程: WorkerID=%d", workerID)
			return
		case task := <-wp.taskChan:
			// 处理任务
			wp.processTask(task, workerID)
		}
	}
}

// processTask 处理任务
func (wp *WorkerPool) processTask(task *kafka.Task, workerID int) {
	startTime := time.Now()
	
	wp.logger.Printf("Worker-%d 开始处理任务: TaskID=%s, IP=%s, Type=%s", 
		workerID, task.TaskID, task.IP, task.TaskType)
	
	// 转换任务格式
	scanTask := &scanner.ScanTask{
		TaskID:   task.TaskID,
		IP:       task.IP,
		TaskType: task.TaskType,
		Priority: task.Priority,
		Config: scanner.ScanConfig{
			Timeout:    10, // 默认10秒超时
			ScanDepth:  "basic",
			EnableApps: task.TaskType == "app_identification",
		},
		Timestamp: task.Timestamp,
	}
	
	// 从任务配置中提取端口
	if ports, ok := task.Config["ports"].([]interface{}); ok {
		scanTask.Config.Ports = make([]int, len(ports))
		for i, port := range ports {
			if p, ok := port.(float64); ok {
				scanTask.Config.Ports[i] = int(p)
			}
		}
	}
	
	// 执行扫描
	result, err := wp.engine.ScanTarget(scanTask)
	if err != nil {
		wp.logger.Printf("Worker-%d 扫描失败: TaskID=%s, Error=%v", 
			workerID, task.TaskID, err)
		
		// 创建失败结果
		result = &scanner.ScanResult{
			TaskID:       task.TaskID,
			IP:           task.IP,
			ScanType:     task.TaskType,
			ScanStatus:   "failed",
			ScanTime:     time.Now().Format(time.RFC3339),
			ErrorMessage: err.Error(),
			NodeID:       "scan-node-001", // TODO: 从配置获取
			Timestamp:    time.Now().Unix(),
		}
	}
	
	// 转换为Kafka结果格式
	kafkaResult := &kafka.ScanResult{
		TaskID:       result.TaskID,
		IP:           result.IP,
		ScanType:     result.ScanType,
		ScanStatus:   result.ScanStatus,
		ScanTime:     result.ScanTime,
		ResponseTime: result.ResponseTime,
		Results:      result.Results,
		ErrorMessage: result.ErrorMessage,
		NodeID:       result.NodeID,
		Timestamp:    result.Timestamp,
	}
	
	// 发送结果
	select {
	case wp.resultChan <- kafkaResult:
		duration := time.Since(startTime)
		wp.logger.Printf("Worker-%d 完成任务: TaskID=%s, Status=%s, Duration=%v", 
			workerID, task.TaskID, result.ScanStatus, duration)
	default:
		wp.logger.Printf("Worker-%d 结果通道满，丢弃结果: TaskID=%s", workerID, task.TaskID)
	}
}

// runTestMode 运行测试模式
func runTestMode(cfg *config.ScanNodeConfig, logger *log.Logger) {
	logger.Println("运行测试模式...")
	
	// 创建探测引擎
	probeEngine := scanner.NewProbeEngine(&cfg.Scanner)
	
	// 测试扫描
	testTask := &scanner.ScanTask{
		TaskID:   "test-001",
		IP:       "127.0.0.1",
		TaskType: "port_scan_default",
		Priority: 1,
		Config: scanner.ScanConfig{
			Ports:     []int{22, 80, 443},
			Timeout:   5,
			ScanDepth: "basic",
		},
		Timestamp: time.Now().Unix(),
	}
	
	result, err := probeEngine.ScanTarget(testTask)
	if err != nil {
		logger.Printf("测试扫描失败: %v", err)
		return
	}
	
	logger.Printf("测试扫描成功: IP=%s, OpenPorts=%d, Status=%s", 
		result.IP, len(result.Results.OpenPorts), result.ScanStatus)
	
	for _, port := range result.Results.OpenPorts {
		logger.Printf("  端口: %d/%s, 服务: %s, 版本: %s", 
			port.Port, port.Protocol, port.Service, port.Version)
	}
}