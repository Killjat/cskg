package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cskg/CyberStroll/internal/enrichment"
	"github.com/cskg/CyberStroll/internal/storage"
	"github.com/cskg/CyberStroll/pkg/config"
)

func main() {
	// 解析命令行参数
	var configPath string
	flag.StringVar(&configPath, "config", "configs/docker-local.yaml", "配置文件路径")
	flag.Parse()

	// 创建日志记录器
	logger := log.New(os.Stdout, "[ENRICHMENT] ", log.LstdFlags|log.Lshortfile)
	logger.Println("启动网站数据富化节点...")

	// 使用简化配置
	cfg := &config.EnrichmentNodeConfig{
		Node: config.NodeConfig{
			ID:     "enrichment-node-001",
			Name:   "网站数据富化节点1",
			Region: "docker-local",
		},
		Elasticsearch: config.ESConfig{
			URLs:     []string{"http://localhost:9200"},
			Index:    "cyberstroll_ip_scan",
			Username: "",
			Password: "",
			Timeout:  30,
		},
		Enrichment: config.EnrichmentConfig{
			BatchSize:         50,
			WorkerCount:       5,
			ScanInterval:      30 * time.Second,
			RequestTimeout:    30 * time.Second,
			MaxRetries:        3,
			EnableCert:        true,
			EnableAPI:         true,
			EnableWebInfo:     true,
			EnableFingerprint: true,
			EnableContent:     true,
		},
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
		logger.Printf("创建Elasticsearch客户端失败: %v", err)
		logger.Println("将使用模拟富化器")
		esClient = nil
	}

	// 创建富化配置
	enrichConfig := &enrichment.EnrichmentConfig{
		BatchSize:         cfg.Enrichment.BatchSize,
		WorkerCount:       cfg.Enrichment.WorkerCount,
		ScanInterval:      cfg.Enrichment.ScanInterval,
		RequestTimeout:    cfg.Enrichment.RequestTimeout,
		MaxRetries:        cfg.Enrichment.MaxRetries,
		EnableCert:        cfg.Enrichment.EnableCert,
		EnableAPI:         cfg.Enrichment.EnableAPI,
		EnableWebInfo:     cfg.Enrichment.EnableWebInfo,
		EnableFingerprint: cfg.Enrichment.EnableFingerprint,
		EnableContent:     cfg.Enrichment.EnableContent,
	}

	// 创建网站数据富化器
	webEnricher := enrichment.NewWebEnricher(esClient, enrichConfig, logger)

	// 启动富化器
	if err := webEnricher.Start(); err != nil {
		log.Fatalf("启动网站数据富化器失败: %v", err)
	}

	logger.Printf("网站数据富化节点启动成功 (节点ID: %s)", cfg.Node.ID)
	logger.Printf("工作协程数: %d", enrichConfig.WorkerCount)
	logger.Printf("扫描间隔: %v", enrichConfig.ScanInterval)
	logger.Printf("启用功能: 证书=%v, API=%v, 网站信息=%v, 指纹=%v, 内容=%v",
		enrichConfig.EnableCert, enrichConfig.EnableAPI, enrichConfig.EnableWebInfo,
		enrichConfig.EnableFingerprint, enrichConfig.EnableContent)

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号
	<-sigChan
	logger.Println("收到停止信号，正在关闭网站数据富化节点...")

	// 停止富化器
	webEnricher.Stop()

	// 关闭Elasticsearch客户端
	if esClient != nil {
		esClient.Close()
	}

	logger.Println("网站数据富化节点已停止")
}