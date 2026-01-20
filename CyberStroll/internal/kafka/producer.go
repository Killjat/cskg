package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// ResultProducer Kafka结果生产者
type ResultProducer struct {
	writer *kafka.Writer
	config *KafkaConfig
	logger *log.Logger
}

// ScanResult 扫描结果
type ScanResult struct {
	TaskID       string      `json:"task_id"`
	IP           string      `json:"ip"`
	ScanType     string      `json:"scan_type"`
	ScanStatus   string      `json:"scan_status"`
	ScanTime     string      `json:"scan_time"`
	ResponseTime int64       `json:"response_time"`
	Results      interface{} `json:"results"`
	ErrorMessage string      `json:"error_message,omitempty"`
	NodeID       string      `json:"node_id"`
	Timestamp    int64       `json:"timestamp"`
}

// NewResultProducer 创建结果生产者
func NewResultProducer(config *KafkaConfig, logger *log.Logger) *ResultProducer {
	if logger == nil {
		logger = log.Default()
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: config.Brokers,
		Topic:   config.ResultTopic,
		Balancer: &kafka.LeastBytes{}, // 负载均衡策略
		
		// 性能配置
		BatchSize:    100,                // 批量大小
		BatchTimeout: 10 * time.Millisecond, // 批量超时
		
		// 可靠性配置
		RequiredAcks: 1, // 至少一个副本确认
		Async:        false,            // 同步发送
		
		// 重试配置
		MaxAttempts: 3,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	})

	return &ResultProducer{
		writer: writer,
		config: config,
		logger: logger,
	}
}

// SendResult 发送扫描结果
func (rp *ResultProducer) SendResult(ctx context.Context, result *ScanResult) error {
	// 序列化结果
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("序列化扫描结果失败: %v", err)
	}

	// 创建消息
	message := kafka.Message{
		Key:   []byte(result.IP),           // 使用IP作为分区键
		Value: data,
		Time:  time.Now(),
		Headers: []kafka.Header{
			{Key: "task_id", Value: []byte(result.TaskID)},
			{Key: "scan_type", Value: []byte(result.ScanType)},
			{Key: "node_id", Value: []byte(result.NodeID)},
		},
	}

	// 发送消息
	err = rp.writer.WriteMessages(ctx, message)
	if err != nil {
		rp.logger.Printf("发送扫描结果失败: TaskID=%s, IP=%s, Error=%v", 
			result.TaskID, result.IP, err)
		return fmt.Errorf("发送扫描结果失败: %v", err)
	}

	rp.logger.Printf("发送扫描结果成功: TaskID=%s, IP=%s, Status=%s", 
		result.TaskID, result.IP, result.ScanStatus)

	return nil
}

// SendResultAsync 异步发送扫描结果
func (rp *ResultProducer) SendResultAsync(result *ScanResult, callback func(error)) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		err := rp.SendResult(ctx, result)
		if callback != nil {
			callback(err)
		}
	}()
}

// SendBatchResults 批量发送扫描结果
func (rp *ResultProducer) SendBatchResults(ctx context.Context, results []*ScanResult) error {
	if len(results) == 0 {
		return nil
	}

	messages := make([]kafka.Message, len(results))
	
	for i, result := range results {
		data, err := json.Marshal(result)
		if err != nil {
			rp.logger.Printf("序列化结果失败: TaskID=%s, Error=%v", result.TaskID, err)
			continue
		}

		messages[i] = kafka.Message{
			Key:   []byte(result.IP),
			Value: data,
			Time:  time.Now(),
			Headers: []kafka.Header{
				{Key: "task_id", Value: []byte(result.TaskID)},
				{Key: "scan_type", Value: []byte(result.ScanType)},
				{Key: "node_id", Value: []byte(result.NodeID)},
			},
		}
	}

	// 批量发送
	err := rp.writer.WriteMessages(ctx, messages...)
	if err != nil {
		rp.logger.Printf("批量发送扫描结果失败: Count=%d, Error=%v", len(results), err)
		return fmt.Errorf("批量发送扫描结果失败: %v", err)
	}

	rp.logger.Printf("批量发送扫描结果成功: Count=%d", len(results))
	return nil
}

// GetProducerStats 获取生产者统计信息
func (rp *ResultProducer) GetProducerStats() kafka.WriterStats {
	return rp.writer.Stats()
}

// Close 关闭生产者
func (rp *ResultProducer) Close() error {
	rp.logger.Println("关闭Kafka生产者...")
	return rp.writer.Close()
}

// TaskProducer 任务生产者 (用于任务管理节点)
type TaskProducer struct {
	systemWriter  *kafka.Writer
	regularWriter *kafka.Writer
	config        *KafkaConfig
	logger        *log.Logger
}

// NewTaskProducer 创建任务生产者
func NewTaskProducer(config *KafkaConfig, logger *log.Logger) *TaskProducer {
	if logger == nil {
		logger = log.Default()
	}

	// 系统任务Writer
	systemWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  config.Brokers,
		Topic:    config.SystemTaskTopic,
		Balancer: &kafka.Hash{}, // 使用Hash分区确保同一IP的任务在同一分区
		
		BatchSize:    50,
		BatchTimeout: 5 * time.Millisecond,
		RequiredAcks: -1, // 系统任务要求所有副本确认
		MaxAttempts:  5,                // 系统任务重试次数更多
	})

	// 常规任务Writer
	regularWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  config.Brokers,
		Topic:    config.RegularTaskTopic,
		Balancer: &kafka.RoundRobin{}, // 轮询分区
		
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: 1, // 常规任务只需要一个副本确认
		MaxAttempts:  3,
	})

	return &TaskProducer{
		systemWriter:  systemWriter,
		regularWriter: regularWriter,
		config:        config,
		logger:        logger,
	}
}

// SendSystemTask 发送系统任务
func (tp *TaskProducer) SendSystemTask(ctx context.Context, task *Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("序列化系统任务失败: %v", err)
	}

	message := kafka.Message{
		Key:   []byte(task.IP),
		Value: data,
		Time:  time.Now(),
		Headers: []kafka.Header{
			{Key: "task_type", Value: []byte(task.TaskType)},
			{Key: "priority", Value: []byte(fmt.Sprintf("%d", task.Priority))},
		},
	}

	err = tp.systemWriter.WriteMessages(ctx, message)
	if err != nil {
		tp.logger.Printf("发送系统任务失败: TaskID=%s, Error=%v", task.TaskID, err)
		return err
	}

	tp.logger.Printf("发送系统任务成功: TaskID=%s, IP=%s", task.TaskID, task.IP)
	return nil
}

// SendRegularTask 发送常规任务
func (tp *TaskProducer) SendRegularTask(ctx context.Context, task *Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("序列化常规任务失败: %v", err)
	}

	message := kafka.Message{
		Key:   []byte(task.IP),
		Value: data,
		Time:  time.Now(),
		Headers: []kafka.Header{
			{Key: "task_type", Value: []byte(task.TaskType)},
			{Key: "user", Value: []byte(task.User)},
		},
	}

	err = tp.regularWriter.WriteMessages(ctx, message)
	if err != nil {
		tp.logger.Printf("发送常规任务失败: TaskID=%s, Error=%v", task.TaskID, err)
		return err
	}

	tp.logger.Printf("发送常规任务成功: TaskID=%s, IP=%s, User=%s", 
		task.TaskID, task.IP, task.User)
	return nil
}

// SendBatchTasks 批量发送任务
func (tp *TaskProducer) SendBatchTasks(ctx context.Context, tasks []*Task, isSystem bool) error {
	if len(tasks) == 0 {
		return nil
	}

	messages := make([]kafka.Message, len(tasks))
	
	for i, task := range tasks {
		data, err := json.Marshal(task)
		if err != nil {
			tp.logger.Printf("序列化任务失败: TaskID=%s, Error=%v", task.TaskID, err)
			continue
		}

		headers := []kafka.Header{
			{Key: "task_type", Value: []byte(task.TaskType)},
		}
		
		if isSystem {
			headers = append(headers, kafka.Header{Key: "priority", Value: []byte(fmt.Sprintf("%d", task.Priority))})
		} else {
			headers = append(headers, kafka.Header{Key: "user", Value: []byte(task.User)})
		}

		messages[i] = kafka.Message{
			Key:     []byte(task.IP),
			Value:   data,
			Time:    time.Now(),
			Headers: headers,
		}
	}

	var err error
	if isSystem {
		err = tp.systemWriter.WriteMessages(ctx, messages...)
		tp.logger.Printf("批量发送系统任务: Count=%d", len(tasks))
	} else {
		err = tp.regularWriter.WriteMessages(ctx, messages...)
		tp.logger.Printf("批量发送常规任务: Count=%d", len(tasks))
	}

	if err != nil {
		tp.logger.Printf("批量发送任务失败: Count=%d, Error=%v", len(tasks), err)
		return err
	}

	return nil
}

// Close 关闭任务生产者
func (tp *TaskProducer) Close() error {
	tp.logger.Println("关闭任务生产者...")
	
	if err := tp.systemWriter.Close(); err != nil {
		tp.logger.Printf("关闭系统任务Writer失败: %v", err)
	}
	
	if err := tp.regularWriter.Close(); err != nil {
		tp.logger.Printf("关闭常规任务Writer失败: %v", err)
	}
	
	return nil
}