package processor

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/cskg/CyberStroll/internal/kafka"
	"github.com/cskg/CyberStroll/internal/storage"
)

// ResultProcessor 结果处理器
type ResultProcessor struct {
	consumer      *kafka.TaskConsumer
	esClient      *storage.ElasticsearchClient
	mongoClient   *storage.MongoClient
	config        *ProcessorConfig
	logger        *log.Logger
	stats         *ProcessorStats
	batchBuffer   []*storage.ScanDocument
	bufferMutex   sync.Mutex
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// ProcessorConfig 处理器配置
type ProcessorConfig struct {
	BatchSize       int           `yaml:"batch_size"`
	BatchTimeout    time.Duration `yaml:"batch_timeout"`
	MaxConcurrency  int           `yaml:"max_concurrency"`
	RetryCount      int           `yaml:"retry_count"`
	EnableGeoLookup bool          `yaml:"enable_geo_lookup"`
}

// ProcessorStats 处理器统计
type ProcessorStats struct {
	TotalProcessed   int64 `json:"total_processed"`
	SuccessProcessed int64 `json:"success_processed"`
	FailedProcessed  int64 `json:"failed_processed"`
	LastProcessTime  int64 `json:"last_process_time"`
	BatchesProcessed int64 `json:"batches_processed"`
	mutex            sync.RWMutex
}

// NewResultProcessor 创建结果处理器
func NewResultProcessor(
	consumer *kafka.TaskConsumer,
	esClient *storage.ElasticsearchClient,
	mongoClient *storage.MongoClient,
	config *ProcessorConfig,
	logger *log.Logger,
) *ResultProcessor {
	if config == nil {
		config = &ProcessorConfig{
			BatchSize:       100,
			BatchTimeout:    5 * time.Second,
			MaxConcurrency:  10,
			RetryCount:      3,
			EnableGeoLookup: false,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ResultProcessor{
		consumer:    consumer,
		esClient:    esClient,
		mongoClient: mongoClient,
		config:      config,
		logger:      logger,
		stats:       &ProcessorStats{},
		batchBuffer: make([]*storage.ScanDocument, 0, config.BatchSize),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start 启动结果处理器
func (rp *ResultProcessor) Start() error {
	rp.logger.Println("启动结果处理器...")

	// 启动结果消费协程
	rp.wg.Add(1)
	go rp.consumeResults(&rp.wg)

	// 启动批量处理协程
	rp.wg.Add(1)
	go rp.batchProcessor(&rp.wg)

	// 启动统计打印协程
	rp.wg.Add(1)
	go rp.printStats(&rp.wg)

	return nil
}

// Stop 停止结果处理器
func (rp *ResultProcessor) Stop() {
	rp.logger.Println("正在停止结果处理器...")

	// 取消上下文
	rp.cancel()

	// 等待所有协程结束
	rp.wg.Wait()

	// 处理剩余的批量数据
	rp.flushBatch()

	rp.logger.Println("结果处理器已停止")
}

// consumeResults 消费扫描结果
func (rp *ResultProcessor) consumeResults(wg *sync.WaitGroup) {
	defer wg.Done()

	rp.logger.Println("开始消费扫描结果...")

	for {
		select {
		case <-rp.ctx.Done():
			rp.logger.Println("停止消费扫描结果")
			return
		default:
			// 从Kafka消费扫描结果
			result, err := rp.consumer.ConsumeResult(rp.ctx)
			if err != nil {
				if err != context.DeadlineExceeded {
					rp.logger.Printf("消费扫描结果失败: %v", err)
				}
				time.Sleep(time.Second)
				continue
			}

			// 处理扫描结果
			rp.processResult(result)
		}
	}
}

// processResult 处理单个扫描结果
func (rp *ResultProcessor) processResult(result *kafka.ScanResult) {
	rp.logger.Printf("处理扫描结果: TaskID=%s, IP=%s", result.TaskID, result.IP)

	// 转换为Elasticsearch文档
	docs := rp.convertToESDocuments(result)

	// 添加到批量缓冲区
	rp.bufferMutex.Lock()
	rp.batchBuffer = append(rp.batchBuffer, docs...)
	rp.bufferMutex.Unlock()

	// 更新MongoDB统计
	rp.updateTaskStatistics(result)

	// 更新统计
	rp.updateStats(true)
}

// convertToESDocuments 转换为Elasticsearch文档
func (rp *ResultProcessor) convertToESDocuments(result *kafka.ScanResult) []*storage.ScanDocument {
	var docs []*storage.ScanDocument

	// 解析扫描时间
	scanTime, _ := time.Parse(time.RFC3339, result.ScanTime)

	// 处理开放端口
	if results, ok := result.Results.(map[string]interface{}); ok {
		if openPorts, ok := results["open_ports"].([]interface{}); ok {
			for _, portData := range openPorts {
				if portMap, ok := portData.(map[string]interface{}); ok {
					doc := &storage.ScanDocument{
						IP:             result.IP,
						Port:           int(portMap["port"].(float64)),
						Protocol:       getString(portMap, "protocol"),
						Service:        getString(portMap, "service"),
						ServiceVersion: getString(portMap, "version"),
						Banner:         getString(portMap, "banner"),
						State:          getString(portMap, "state"),
						ScanTime:       scanTime,
						LastUpdate:     time.Now(),
						TaskID:         result.TaskID,
						NodeID:         result.NodeID,
						Applications:   []storage.ApplicationDoc{},
						Metadata: map[string]interface{}{
							"scan_type":   result.ScanType,
							"scan_status": result.ScanStatus,
						},
					}

					// 添加地理信息 (如果启用)
					if rp.config.EnableGeoLookup {
						doc.GeoInfo = rp.lookupGeoInfo(result.IP)
					}

					docs = append(docs, doc)
				}
			}
		}
	}

	// 如果没有开放端口，创建一个基础文档
	if len(docs) == 0 {
		doc := &storage.ScanDocument{
			IP:         result.IP,
			Port:       0,
			Protocol:   "tcp",
			Service:    "none",
			State:      "closed",
			ScanTime:   scanTime,
			LastUpdate: time.Now(),
			TaskID:     result.TaskID,
			NodeID:     result.NodeID,
			Metadata: map[string]interface{}{
				"scan_type":   result.ScanType,
				"scan_status": result.ScanStatus,
			},
		}

		if rp.config.EnableGeoLookup {
			doc.GeoInfo = rp.lookupGeoInfo(result.IP)
		}

		docs = append(docs, doc)
	}

	return docs
}

// batchProcessor 批量处理器
func (rp *ResultProcessor) batchProcessor(wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(rp.config.BatchTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-rp.ctx.Done():
			rp.logger.Println("停止批量处理器")
			return
		case <-ticker.C:
			rp.processBatch()
		}
	}
}

// processBatch 处理批量数据
func (rp *ResultProcessor) processBatch() {
	rp.bufferMutex.Lock()
	if len(rp.batchBuffer) == 0 {
		rp.bufferMutex.Unlock()
		return
	}

	// 复制缓冲区数据
	batch := make([]*storage.ScanDocument, len(rp.batchBuffer))
	copy(batch, rp.batchBuffer)
	rp.batchBuffer = rp.batchBuffer[:0] // 清空缓冲区
	rp.bufferMutex.Unlock()

	rp.logger.Printf("处理批量数据: %d 个文档", len(batch))

	// 批量索引到Elasticsearch
	if err := rp.esClient.BulkIndexDocuments(batch); err != nil {
		rp.logger.Printf("批量索引失败: %v", err)
		rp.updateStats(false)
		return
	}

	rp.logger.Printf("批量索引成功: %d 个文档", len(batch))
	rp.stats.mutex.Lock()
	rp.stats.BatchesProcessed++
	rp.stats.mutex.Unlock()
}

// flushBatch 刷新批量数据
func (rp *ResultProcessor) flushBatch() {
	rp.bufferMutex.Lock()
	defer rp.bufferMutex.Unlock()

	if len(rp.batchBuffer) > 0 {
		rp.logger.Printf("刷新剩余批量数据: %d 个文档", len(rp.batchBuffer))
		if err := rp.esClient.BulkIndexDocuments(rp.batchBuffer); err != nil {
			rp.logger.Printf("刷新批量数据失败: %v", err)
		}
		rp.batchBuffer = rp.batchBuffer[:0]
	}
}

// updateTaskStatistics 更新任务统计
func (rp *ResultProcessor) updateTaskStatistics(result *kafka.ScanResult) {
	// 创建任务统计记录
	stats := &storage.TaskStatistics{
		TaskID:       result.TaskID,
		IP:           result.IP,
		ScanStatus:   result.ScanStatus,
		OpenPorts:    []int{},
		Services:     []string{},
		ResponseTime: 0, // 从结果中提取
		ErrorMessage: "",
	}

	// 提取开放端口和服务
	if results, ok := result.Results.(map[string]interface{}); ok {
		if openPorts, ok := results["open_ports"].([]interface{}); ok {
			for _, portData := range openPorts {
				if portMap, ok := portData.(map[string]interface{}); ok {
					port := int(portMap["port"].(float64))
					service := getString(portMap, "service")
					
					stats.OpenPorts = append(stats.OpenPorts, port)
					if service != "" && service != "unknown" {
						stats.Services = append(stats.Services, service)
					}
				}
			}
		}
	}

	// 保存到MongoDB (这里使用简化版本)
	// TODO: 实现完整的MongoDB统计保存
	rp.logger.Printf("更新任务统计: TaskID=%s, IP=%s, OpenPorts=%d", 
		stats.TaskID, stats.IP, len(stats.OpenPorts))
}

// lookupGeoInfo 查找地理信息
func (rp *ResultProcessor) lookupGeoInfo(ip string) *storage.GeoInfo {
	// 这里应该集成真实的地理位置查询服务
	// 暂时返回模拟数据
	return &storage.GeoInfo{
		Country:     "Unknown",
		CountryCode: "XX",
		Region:      "Unknown",
		City:        "Unknown",
		ISP:         "Unknown",
	}
}

// updateStats 更新统计信息
func (rp *ResultProcessor) updateStats(success bool) {
	rp.stats.mutex.Lock()
	defer rp.stats.mutex.Unlock()

	rp.stats.TotalProcessed++
	if success {
		rp.stats.SuccessProcessed++
	} else {
		rp.stats.FailedProcessed++
	}
	rp.stats.LastProcessTime = time.Now().Unix()
}

// GetStats 获取统计信息
func (rp *ResultProcessor) GetStats() *ProcessorStats {
	rp.stats.mutex.RLock()
	defer rp.stats.mutex.RUnlock()

	return &ProcessorStats{
		TotalProcessed:   rp.stats.TotalProcessed,
		SuccessProcessed: rp.stats.SuccessProcessed,
		FailedProcessed:  rp.stats.FailedProcessed,
		LastProcessTime:  rp.stats.LastProcessTime,
		BatchesProcessed: rp.stats.BatchesProcessed,
	}
}

// printStats 打印统计信息
func (rp *ResultProcessor) printStats(wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-rp.ctx.Done():
			return
		case <-ticker.C:
			stats := rp.GetStats()
			rp.logger.Printf("处理器统计: 总处理=%d, 成功=%d, 失败=%d, 批次=%d",
				stats.TotalProcessed, stats.SuccessProcessed, stats.FailedProcessed, stats.BatchesProcessed)
		}
	}
}

// 辅助函数
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}