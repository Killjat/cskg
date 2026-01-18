package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"github.com/cskg/assetdiscovery/common"
)

// Server 服务端结构体
type Server struct {
	config          *common.Config
	taskManager     *TaskManager
	resultManager   *ResultManager
	clientManager   *ClientManager
	router          *Router
	kafkaProducer   *kafka.Writer
	kafkaConsumer   *kafka.Reader
	webServer       *WebServer
}

// NewServer 创建新的服务端实例
func NewServer(config *common.Config) (*Server, error) {
	// 初始化Kafka生产者
	kafkaProducer := &kafka.Writer{
		Addr:     kafka.TCP(config.Kafka.Brokers...),
		Topic:    config.Kafka.TaskTopic,
		Balancer: &kafka.LeastBytes{},
	}

	// 初始化Kafka消费者
	kafkaConsumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     config.Kafka.Brokers,
		Topic:       config.Kafka.ResultTopic,
		GroupID:     config.Kafka.GroupID,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     1 * time.Second,
	})

	// 创建服务器实例
	server := &Server{
		config:        config,
		kafkaProducer: kafkaProducer,
		kafkaConsumer: kafkaConsumer,
	}

	// 初始化各个组件
	server.taskManager = NewTaskManager(server)
	server.resultManager = NewResultManager(server)
	server.clientManager = NewClientManager(server)
	server.router = NewRouter(server)
	server.webServer = NewWebServer(server)

	return server, nil
}

// Start 启动服务端
func (s *Server) Start() error {
	log.Println("Starting asset discovery server...")

	// 启动结果处理协程
	go s.resultManager.Start()

	// 启动客户端管理器
	go s.clientManager.Start()

	// 启动Web服务器
	go func() {
		if err := s.webServer.Start(); err != nil {
			log.Fatalf("Failed to start web server: %v", err)
		}
	}()

	// 启动scantaskip.txt文件读取协程
	go s.readScanTaskIPFile()

	log.Printf("Server started successfully on %s:%d", s.config.Server.Host, s.config.Server.Port)

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 优雅关闭
	log.Println("Shutting down server...")
	return s.Shutdown()
}

// readScanTaskIPFile 读取scantaskip.txt文件并下发任务
func (s *Server) readScanTaskIPFile() {
	log.Println("Starting to read scantaskip.txt file...")
	
	// 周期性读取文件（每30秒）
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// 读取文件
			content, err := os.ReadFile("scantaskip.txt")
			if err != nil {
				// 文件不存在时不报错，继续下一次读取
				if !os.IsNotExist(err) {
					log.Printf("Error reading scantaskip.txt: %v", err)
				}
				continue
			}
			
			// 解析文件内容，每行一个IP或网段
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					// 跳过空行和注释行
					continue
				}
				
				// 为每个IP/网段创建扫描任务
				params := map[string]interface{}{
					"port_range": s.config.Scan.PortRange,
				}
				s.CreateTask(common.TaskTypeScanIP, line, params)
			}
			
			// 读取完成后清空文件
			if err := os.Truncate("scantaskip.txt", 0); err != nil {
				log.Printf("Error truncating scantaskip.txt: %v", err)
			}
		}
	}
}

// Shutdown 关闭服务端
func (s *Server) Shutdown() error {
	// 关闭Kafka连接
	if err := s.kafkaProducer.Close(); err != nil {
		log.Printf("Error closing Kafka producer: %v", err)
	}

	if err := s.kafkaConsumer.Close(); err != nil {
		log.Printf("Error closing Kafka consumer: %v", err)
	}

	// 关闭Web服务器
	if err := s.webServer.Shutdown(); err != nil {
		log.Printf("Error shutting down web server: %v", err)
	}

	log.Println("Server shutdown successfully")
	return nil
}

// CreateTask 创建新任务
func (s *Server) CreateTask(taskType common.TaskType, target string, params map[string]interface{}) string {
	taskID := uuid.New().String()
	task := common.Task{
		TaskID:     taskID,
		TaskType:   taskType,
		Target:     target,
		Parameters: params,
		Timestamp:  time.Now().Unix(),
	}

	// 路由分配任务
	s.router.RouteTask(&task)

	return taskID
}

func main() {
	// 加载配置
	configPath := "./config/config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	config, err := common.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建并启动服务器
	server, err := NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
