package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cskg/CyberStroll/internal/kafka"
	"github.com/cskg/CyberStroll/internal/processor"
	"github.com/cskg/CyberStroll/internal/storage"
	"github.com/cskg/CyberStroll/pkg/config"
)

func main() {
	// 解析命令行参数
	var configPath string
	flag.StringVar(&configPath, "config", "configs/docker-local.yaml", "配置文件路径")
	flag.Parse()

	// 创建日志记录器
	logger := log.New(os.Stdout, "[PROCESSOR] ", log.LstdFlags|log.Lshortfile)
	logger.Println("启动处理节点...")

	// 使用简化配置
	cfg := &config.ProcessorNodeConfig{
		Node: config.NodeConfig{
			ID:     "processor-node-001",
			Name:   "处理节点1",
			Region: "docker-local",
		},
		Kafka: kafka.KafkaConfig{
			Brokers:     []string{"localhost:9092"},
			ResultTopic: "scan_results",
			GroupID:     "processors",
		},
		Elasticsearch: config.ESConfig{
			URLs:     []string{"http://localhost:9200"},
			Index:    "cyberstroll_ip_scan",
			Username: "",
			Password: "",
			Timeout:  30,
		},
		Storage: config.StorageConfig{
			MongoDB: config.MongoConfig{
				URI:      "mongodb://cyberstroll:cyberstroll123@localhost:27017/cyberstroll?authSource=admin",
				Database: "cyberstroll",
				Timeout:  10,
			},
		},
		Processor: config.ProcessorConfig{
			BatchSize:       100,
			BatchTimeout:    5,
			MaxConcurrency:  10,
			RetryCount:      3,
			EnableGeoLookup: false,
		},
		Debug: false,
	}

	// 创建Kafka消费者
	consumer, err := kafka.NewTaskConsumerWithConfig(&kafka.ConsumerConfig{
		Brokers:     cfg.Kafka.Brokers,
		GroupID:     cfg.Kafka.GroupID,
		Topics:      []string{cfg.Kafka.ResultTopic},
		MaxRetries:  3,
		EnableDebug: cfg.Debug,
	}, logger)
	if err != nil {
		log.Fatalf("创建Kafka消费者失败: %v", err)
	}

	// 创建Elasticsearch客户端
	esClient, err := storage.NewElasticsearchClient(&storage.ESConfig{
		URLs:     cfg.Elasticsearch.URLs,
		Index:    cfg.Elasticsearch.Index,
		Username: cfg.Elasticsearch.Username,
		Password: cfg.Elasticsearch.Password,
		Timeout:  cfg.Elasticsearch.Timeout,
	})
	if err != nil {
		log.Fatalf("创建Elasticsearch客户端失败: %v", err)
	}

	// 创建MongoDB客户端
	mongoClient, err := storage.NewMongoClient(&storage.MongoConfig{
		URI:      cfg.Storage.MongoDB.URI,
		Database: cfg.Storage.MongoDB.Database,
		Timeout:  cfg.Storage.MongoDB.Timeout,
	})
	if err != nil {
		log.Fatalf("创建MongoDB客户端失败: %v", err)
	}

	// 创建结果处理器
	processorConfig := &processor.ProcessorConfig{
		BatchSize:       cfg.Processor.BatchSize,
		BatchTimeout:    time.Duration(cfg.Processor.BatchTimeout) * time.Second,
		MaxConcurrency:  cfg.Processor.MaxConcurrency,
		RetryCount:      cfg.Processor.RetryCount,
		EnableGeoLookup: cfg.Processor.EnableGeoLookup,
	}

	resultProcessor := processor.NewResultProcessor(
		consumer,
		esClient,
		mongoClient,
		processorConfig,
		logger,
	)

	// 启动结果处理器
	if err := resultProcessor.Start(); err != nil {
		log.Fatalf("启动结果处理器失败: %v", err)
	}

	logger.Println("处理节点启动成功")
	logger.Printf("配置: BatchSize=%d, BatchTimeout=%v, MaxConcurrency=%d",
		processorConfig.BatchSize, processorConfig.BatchTimeout, processorConfig.MaxConcurrency)

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号
	<-sigChan
	logger.Println("收到停止信号，正在关闭处理节点...")

	// 停止结果处理器
	resultProcessor.Stop()

	// 关闭客户端连接
	esClient.Close()
	mongoClient.Close()

	logger.Println("处理节点已停止")
}