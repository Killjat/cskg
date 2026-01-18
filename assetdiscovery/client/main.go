package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/cskg/assetdiscovery/common"
)

// Client 客户端结构体
type Client struct {
	config          *common.Config
	taskConsumer    *TaskConsumer
	resultProducer  *ResultProducer
	scanExecutor    *ScanExecutor
	clientManager   *ClientManager
	kafkaConsumer   *kafka.Reader
	kafkaProducer   *kafka.Writer
}

// NewClient 创建新的客户端实例
func NewClient(config *common.Config) (*Client, error) {
	// 初始化Kafka消费者
	kafkaConsumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     config.Kafka.Brokers,
		Topic:       config.Kafka.TaskTopic,
		GroupID:     config.Kafka.GroupID + "_" + config.Client.ClientID,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     1 * time.Second,
	})

	// 初始化Kafka生产者
	kafkaProducer := &kafka.Writer{
		Addr:     kafka.TCP(config.Kafka.Brokers...),
		Topic:    config.Kafka.ResultTopic,
		Balancer: &kafka.LeastBytes{},
	}

	// 创建客户端实例
	client := &Client{
		config:        config,
		kafkaConsumer: kafkaConsumer,
		kafkaProducer: kafkaProducer,
	}

	// 初始化各个组件
	client.resultProducer = NewResultProducer(client)
	client.scanExecutor = NewScanExecutor(client)
	client.scanExecutor.webScanner = NewWebScanner(client)
	client.taskConsumer = NewTaskConsumer(client)
	client.clientManager = NewClientManager(client)

	return client, nil
}

// Start 启动客户端
func (c *Client) Start() error {
	log.Printf("Starting asset discovery client %s...", c.config.Client.ClientID)

	// 注册客户端
	c.clientManager.Register()

	// 启动心跳协程
	go c.clientManager.SendHeartbeats()

	// 启动任务消费者
	go c.taskConsumer.Start()

	log.Printf("Client started successfully. Client ID: %s", c.config.Client.ClientID)

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 优雅关闭
	log.Println("Shutting down client...")
	return c.Shutdown()
}

// Shutdown 关闭客户端
func (c *Client) Shutdown() error {
	// 关闭Kafka连接
	if err := c.kafkaConsumer.Close(); err != nil {
		log.Printf("Error closing Kafka consumer: %v", err)
	}

	if err := c.kafkaProducer.Close(); err != nil {
		log.Printf("Error closing Kafka producer: %v", err)
	}

	log.Println("Client shutdown successfully")
	return nil
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

	// 创建并启动客户端
	client, err := NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	if err := client.Start(); err != nil {
		log.Fatalf("Client error: %v", err)
	}
}
