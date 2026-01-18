package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cskg/CyberStroll/internal/config"
	"github.com/cskg/CyberStroll/pkg/models"
	"github.com/segmentio/kafka-go"
)

// Producer Kafka生产者结构体
type Producer struct {
	writer *kafka.Writer
}

// NewProducer 创建Kafka生产者实例
func NewProducer(cfg *config.KafkaConfig) (*Producer, error) {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		// 简化配置，移除可能不兼容的字段
	}

	return &Producer{
		writer: writer,
	}, nil
}

// SendTask 发送任务到Kafka
func (p *Producer) SendTask(ctx context.Context, topic string, task *models.Task) error {
	// 将任务转换为JSON字节
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// 创建Kafka消息
	message := kafka.Message{
		Topic: topic,
		Key:   []byte(task.ID),
		Value: taskBytes,
	}

	// 发送消息
	if err := p.writer.WriteMessages(ctx, message); err != nil {
		return fmt.Errorf("failed to send task to kafka: %w", err)
	}

	return nil
}

// Close 关闭Kafka生产者
func (p *Producer) Close() error {
	return p.writer.Close()
}
