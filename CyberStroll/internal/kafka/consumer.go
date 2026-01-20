package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// TaskConsumer Kafka任务消费者
type TaskConsumer struct {
	systemReader  *kafka.Reader
	regularReader *kafka.Reader
	resultReader  *kafka.Reader
	config        *KafkaConfig
	logger        *log.Logger
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers           []string `yaml:"brokers" mapstructure:"brokers"`
	SystemTaskTopic   string   `yaml:"system_task_topic" mapstructure:"system_task_topic"`
	RegularTaskTopic  string   `yaml:"regular_task_topic" mapstructure:"regular_task_topic"`
	ResultTopic       string   `yaml:"result_topic" mapstructure:"result_topic"`
	GroupID           string   `yaml:"group_id" mapstructure:"group_id"`
	AutoOffsetReset   string   `yaml:"auto_offset_reset" mapstructure:"auto_offset_reset"`
	SessionTimeout    int      `yaml:"session_timeout" mapstructure:"session_timeout"`
	HeartbeatInterval int      `yaml:"heartbeat_interval" mapstructure:"heartbeat_interval"`
}

// Task 任务结构
type Task struct {
	TaskID    string                 `json:"task_id"`
	IP        string                 `json:"ip"`
	TaskType  string                 `json:"task_type"`
	Priority  int                    `json:"priority"`
	User      string                 `json:"user,omitempty"`
	Config    map[string]interface{} `json:"config"`
	Timestamp int64                  `json:"timestamp"`
}

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
	Brokers     []string `yaml:"brokers"`
	GroupID     string   `yaml:"group_id"`
	Topics      []string `yaml:"topics"`
	MaxRetries  int      `yaml:"max_retries"`
	EnableDebug bool     `yaml:"enable_debug"`
}

// NewTaskConsumer 创建任务消费者 (兼容新接口)
func NewTaskConsumerWithConfig(config *ConsumerConfig, logger *log.Logger) (*TaskConsumer, error) {
	kafkaConfig := &KafkaConfig{
		Brokers:          config.Brokers,
		SystemTaskTopic:  "system_tasks",
		RegularTaskTopic: "regular_tasks",
		ResultTopic:      "scan_results",
		GroupID:          config.GroupID,
	}
	
	return NewTaskConsumer(kafkaConfig, logger), nil
}
func NewTaskConsumer(config *KafkaConfig, logger *log.Logger) *TaskConsumer {
	if logger == nil {
		logger = log.Default()
	}

	// 系统任务Reader
	systemReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.Brokers,
		Topic:   config.SystemTaskTopic,
		GroupID: config.GroupID + "-system",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset: kafka.LastOffset,
	})

	// 常规任务Reader
	regularReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.Brokers,
		Topic:   config.RegularTaskTopic,
		GroupID: config.GroupID + "-regular",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset: kafka.LastOffset,
	})

	// 扫描结果Reader
	resultReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.Brokers,
		Topic:   config.ResultTopic,
		GroupID: config.GroupID + "-results",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset: kafka.LastOffset,
	})

	return &TaskConsumer{
		systemReader:  systemReader,
		regularReader: regularReader,
		resultReader:  resultReader,
		config:        config,
		logger:        logger,
	}
}

// ConsumeTask 消费任务 (优先处理系统任务)
func (tc *TaskConsumer) ConsumeTask(ctx context.Context) (*Task, error) {
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 优先尝试消费系统任务
	select {
	case <-timeoutCtx.Done():
		// 超时，尝试消费常规任务
		return tc.consumeRegularTask(ctx)
	default:
		if task, err := tc.consumeSystemTask(timeoutCtx); err == nil {
			tc.logger.Printf("消费到系统任务: %s", task.TaskID)
			return task, nil
		}
		// 系统任务队列为空，消费常规任务
		return tc.consumeRegularTask(ctx)
	}
}

// consumeSystemTask 消费系统任务
func (tc *TaskConsumer) consumeSystemTask(ctx context.Context) (*Task, error) {
	message, err := tc.systemReader.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取系统任务失败: %v", err)
	}

	var task Task
	if err := json.Unmarshal(message.Value, &task); err != nil {
		tc.logger.Printf("解析系统任务失败: %v", err)
		return nil, fmt.Errorf("解析系统任务失败: %v", err)
	}

	// 标记为系统任务
	task.Priority = 10 // 系统任务最高优先级

	tc.logger.Printf("消费系统任务: TaskID=%s, IP=%s, Type=%s", 
		task.TaskID, task.IP, task.TaskType)

	return &task, nil
}

// consumeRegularTask 消费常规任务
func (tc *TaskConsumer) consumeRegularTask(ctx context.Context) (*Task, error) {
	message, err := tc.regularReader.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取常规任务失败: %v", err)
	}

	var task Task
	if err := json.Unmarshal(message.Value, &task); err != nil {
		tc.logger.Printf("解析常规任务失败: %v", err)
		return nil, fmt.Errorf("解析常规任务失败: %v", err)
	}

	tc.logger.Printf("消费常规任务: TaskID=%s, IP=%s, Type=%s, User=%s", 
		task.TaskID, task.IP, task.TaskType, task.User)

	return &task, nil
}

// ConsumeSystemTasks 持续消费系统任务
func (tc *TaskConsumer) ConsumeSystemTasks(ctx context.Context, taskChan chan<- *Task) error {
	tc.logger.Println("开始消费系统任务...")
	
	for {
		select {
		case <-ctx.Done():
			tc.logger.Println("停止消费系统任务")
			return ctx.Err()
		default:
			task, err := tc.consumeSystemTask(ctx)
			if err != nil {
				// 如果是超时错误，继续循环
				if err == context.DeadlineExceeded {
					continue
				}
				tc.logger.Printf("消费系统任务错误: %v", err)
				time.Sleep(time.Second) // 错误后等待1秒
				continue
			}
			
			select {
			case taskChan <- task:
				// 任务发送成功
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

// ConsumeRegularTasks 持续消费常规任务
func (tc *TaskConsumer) ConsumeRegularTasks(ctx context.Context, taskChan chan<- *Task) error {
	tc.logger.Println("开始消费常规任务...")
	
	for {
		select {
		case <-ctx.Done():
			tc.logger.Println("停止消费常规任务")
			return ctx.Err()
		default:
			task, err := tc.consumeRegularTask(ctx)
			if err != nil {
				// 如果是超时错误，继续循环
				if err == context.DeadlineExceeded {
					continue
				}
				tc.logger.Printf("消费常规任务错误: %v", err)
				time.Sleep(time.Second) // 错误后等待1秒
				continue
			}
			
			select {
			case taskChan <- task:
				// 任务发送成功
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

// GetConsumerStats 获取消费者统计信息
func (tc *TaskConsumer) GetConsumerStats() map[string]interface{} {
	systemStats := tc.systemReader.Stats()
	regularStats := tc.regularReader.Stats()
	
	return map[string]interface{}{
		"system_consumer": map[string]interface{}{
			"topic":     systemStats.Topic,
			"partition": systemStats.Partition,
			"offset":    systemStats.Offset,
			"lag":       systemStats.Lag,
		},
		"regular_consumer": map[string]interface{}{
			"topic":     regularStats.Topic,
			"partition": regularStats.Partition,
			"offset":    regularStats.Offset,
			"lag":       regularStats.Lag,
		},
	}
}

// ConsumeResult 消费扫描结果
func (tc *TaskConsumer) ConsumeResult(ctx context.Context) (*ScanResult, error) {
	message, err := tc.resultReader.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取扫描结果失败: %v", err)
	}

	var result ScanResult
	if err := json.Unmarshal(message.Value, &result); err != nil {
		tc.logger.Printf("解析扫描结果失败: %v", err)
		return nil, fmt.Errorf("解析扫描结果失败: %v", err)
	}

	return &result, nil
}

// Close 关闭消费者
func (tc *TaskConsumer) Close() error {
	tc.logger.Println("关闭Kafka消费者...")
	
	if err := tc.systemReader.Close(); err != nil {
		tc.logger.Printf("关闭系统任务Reader失败: %v", err)
	}
	
	if err := tc.regularReader.Close(); err != nil {
		tc.logger.Printf("关闭常规任务Reader失败: %v", err)
	}
	
	if err := tc.resultReader.Close(); err != nil {
		tc.logger.Printf("关闭结果Reader失败: %v", err)
	}
	
	return nil
}

// CommitMessage 手动提交消息
func (tc *TaskConsumer) CommitMessage(ctx context.Context, message kafka.Message) error {
	// 根据topic确定使用哪个reader
	if message.Topic == tc.config.SystemTaskTopic {
		return tc.systemReader.CommitMessages(ctx, message)
	} else if message.Topic == tc.config.RegularTaskTopic {
		return tc.regularReader.CommitMessages(ctx, message)
	}
	
	return fmt.Errorf("未知的topic: %s", message.Topic)
}