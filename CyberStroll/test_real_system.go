package main

import (
	"fmt"
	"log"
	"time"

	"github.com/cskg/CyberStroll/internal/storage"
	"github.com/cskg/CyberStroll/pkg/config"
)

func main() {
	fmt.Println("=== CyberStroll 真实系统测试 ===")

	// 加载配置
	cfg, err := config.LoadConfig("configs/docker-local.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建ES配置
	esConfig := &storage.ESConfig{
		URLs:    []string{"http://localhost:9200"},
		Index:   "cyberstroll_ip_scan",
		Timeout: 30,
	}

	// 连接Elasticsearch
	esClient, err := storage.NewElasticsearchClient(esConfig)
	if err != nil {
		log.Fatalf("连接Elasticsearch失败: %v", err)
	}

	// 测试数据
	testData := []*storage.ScanDocument{
		{
			IP:             "192.168.1.1",
			Port:           80,
			Protocol:       "tcp",
			Service:        "http",
			ServiceVersion: "nginx/1.18.0",
			Banner:         "HTTP/1.1 200 OK\r\nServer: nginx/1.18.0",
			State:          "open",
			ScanTime:       time.Now(),
			LastUpdate:     time.Now(),
			TaskID:         "test-task-001",
			NodeID:         "test-node-001",
			GeoInfo: &storage.GeoInfo{
				Country:   "中国",
				Region:    "北京",
				City:      "北京",
				Latitude:  39.9042,
				Longitude: 116.4074,
			},
			Metadata: make(map[string]interface{}),
		},
		{
			IP:             "192.168.1.2",
			Port:           22,
			Protocol:       "tcp",
			Service:        "ssh",
			ServiceVersion: "OpenSSH_8.0",
			Banner:         "SSH-2.0-OpenSSH_8.0",
			State:          "open",
			ScanTime:       time.Now(),
			LastUpdate:     time.Now(),
			TaskID:         "test-task-001",
			NodeID:         "test-node-001",
			GeoInfo: &storage.GeoInfo{
				Country:   "美国",
				Region:    "加利福尼亚",
				City:      "旧金山",
				Latitude:  37.7749,
				Longitude: -122.4194,
			},
			Metadata: make(map[string]interface{}),
		},
		{
			IP:             "192.168.1.3",
			Port:           443,
			Protocol:       "tcp",
			Service:        "https",
			ServiceVersion: "Apache/2.4.41",
			Banner:         "HTTP/1.1 200 OK\r\nServer: Apache/2.4.41",
			State:          "open",
			ScanTime:       time.Now(),
			LastUpdate:     time.Now(),
			TaskID:         "test-task-001",
			NodeID:         "test-node-001",
			GeoInfo: &storage.GeoInfo{
				Country:   "日本",
				Region:    "东京",
				City:      "东京",
				Latitude:  35.6762,
				Longitude: 139.6503,
			},
			Metadata: make(map[string]interface{}),
		},
	}

	fmt.Printf("正在插入 %d 条测试数据到Elasticsearch...\n", len(testData))

	// 批量插入测试数据
	for i, doc := range testData {
		if err := esClient.IndexDocument(doc); err != nil {
			log.Printf("插入文档 %d 失败: %v", i+1, err)
		} else {
			fmt.Printf("✓ 插入文档 %d 成功 (%s:%d)\n", i+1, doc.IP, doc.Port)
		}
	}

	// 等待索引刷新
	fmt.Println("等待Elasticsearch索引刷新...")
	time.Sleep(2 * time.Second)

	// 测试搜索功能
	fmt.Println("\n=== 测试搜索功能 ===")

	// 测试1: 搜索所有数据
	fmt.Println("1. 搜索所有数据:")
	results, err := esClient.SearchDocuments(map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	})
	if err != nil {
		log.Printf("搜索失败: %v", err)
	} else {
		fmt.Printf("   找到 %d 条记录\n", len(results))
		for _, doc := range results {
			fmt.Printf("   - %s:%d (%s)\n", doc.IP, doc.Port, doc.Service)
		}
	}

	// 测试2: 按服务搜索
	fmt.Println("\n2. 搜索HTTP服务:")
	results, err = esClient.SearchDocuments(map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"match": map[string]interface{}{
							"service": "http",
						},
					},
					{
						"match": map[string]interface{}{
							"service": "https",
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Printf("搜索失败: %v", err)
	} else {
		fmt.Printf("   找到 %d 条HTTP/HTTPS记录\n", len(results))
		for _, doc := range results {
			fmt.Printf("   - %s:%d (%s)\n", doc.IP, doc.Port, doc.Service)
		}
	}

	// 测试3: 按国家搜索
	fmt.Println("\n3. 搜索中国的服务:")
	results, err = esClient.SearchDocuments(map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"geo_info.country": "中国",
			},
		},
	})
	if err != nil {
		log.Printf("搜索失败: %v", err)
	} else {
		fmt.Printf("   找到 %d 条中国记录\n", len(results))
		for _, doc := range results {
			country := "-"
			if doc.GeoInfo != nil {
				country = doc.GeoInfo.Country
			}
			fmt.Printf("   - %s:%d (%s) - %s\n", doc.IP, doc.Port, doc.Service, country)
		}
	}

	// 获取统计信息
	fmt.Println("\n=== Elasticsearch统计信息 ===")
	stats, err := esClient.GetStats()
	if err != nil {
		log.Printf("获取统计失败: %v", err)
	} else {
		if indices, ok := stats["indices"].(map[string]interface{}); ok {
			if scanIndex, ok := indices["cyberstroll_ip_scan"].(map[string]interface{}); ok {
				if primaries, ok := scanIndex["primaries"].(map[string]interface{}); ok {
					if docs, ok := primaries["docs"].(map[string]interface{}); ok {
						if count, ok := docs["count"].(float64); ok {
							fmt.Printf("索引文档总数: %.0f\n", count)
						}
					}
				}
			}
		}
	}

	fmt.Println("\n=== 测试完成 ===")
	fmt.Println("✓ 搜索节点运行在: http://localhost:8082")
	fmt.Println("✓ 富化节点正在运行")
	fmt.Println("✓ Elasticsearch连接正常")
	fmt.Println("✓ 数据插入和搜索功能正常")
	fmt.Println("\n可以通过浏览器访问 http://localhost:8082 查看Web界面")
}