package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cskg/CyberStroll/internal/config"
	"github.com/cskg/CyberStroll/pkg/models"
	"github.com/segmentio/kafka-go"
)

// Consumer Kafka消费者结构体
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer 创建Kafka消费者实例
func NewConsumer(cfg *config.KafkaConfig, topic string) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.GroupID,
		Topic:    topic,
	})

	return &Consumer{
		reader: reader,
	}, nil
}



// ReadTask 从Kafka读取任务
func (c *Consumer) ReadTask(ctx context.Context) (*models.Task, error) {
	// 读取Kafka消息
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read message from kafka: %w", err)
	}

	// 解析任务
	var task models.Task
	if err := json.Unmarshal(msg.Value, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

// Close 关闭Kafka消费者
func (c *Consumer) Close() error {
	return c.reader.Close()
}
