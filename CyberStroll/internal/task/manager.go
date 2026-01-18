package task

import (
	"context"
	"fmt"
	"time"

	"github.com/cskg/CyberStroll/internal/config"
	"github.com/cskg/CyberStroll/internal/kafka"
	"github.com/cskg/CyberStroll/pkg/models"
	"github.com/google/uuid"
)

// Manager 任务管理器结构体
type Manager struct {
	kafkaProducer *kafka.Producer
	config        *config.Config
}

// NewManager 创建任务管理器实例
func NewManager(cfg *config.Config, kafkaProducer *kafka.Producer) *Manager {
	return &Manager{
		kafkaProducer: kafkaProducer,
		config:        cfg,
	}
}

// CreateTask 创建新任务
func (m *Manager) CreateTask(ctx context.Context, creator string, taskType models.TaskType, protocol models.ProtocolType, targets []string, portRange string, isSystemTask bool) (*models.Task, error) {
	// 生成任务ID
	taskID := uuid.New().String()
	now := time.Now()

	// 创建任务
	task := &models.Task{
		ID:          taskID,
		Creator:     creator,
		Type:        taskType,
		Protocol:    protocol,
		Targets:     targets,
		PortRange:   portRange,
		Status:      models.TaskStatusPending,
		CreatedAt:   now,
		ScheduledAt: now,
	}

	// 选择Kafka主题
	var topic string
	if isSystemTask {
		topic = m.config.Kafka.Topics.SystemTask
	} else {
		topic = m.config.Kafka.Topics.NormalTask
	}

	// 发送任务到Kafka
	if err := m.kafkaProducer.SendTask(ctx, topic, task); err != nil {
		return nil, fmt.Errorf("failed to send task to kafka: %w", err)
	}

	return task, nil
}

// UpdateTaskStatus 更新任务状态
func (m *Manager) UpdateTaskStatus(ctx context.Context, task *models.Task, status models.TaskStatus) {
	task.Status = status
	now := time.Now()

	switch status {
	case models.TaskStatusRunning:
		task.StartedAt = &now
	case models.TaskStatusCompleted, models.TaskStatusFailed:
		task.CompletedAt = &now
	}

	// TODO: 可以添加任务状态持久化逻辑
}

// AddScanResultInfo 添加扫描结果信息到任务
func (m *Manager) AddScanResultInfo(task *models.Task, index string, keys []string) {
	task.ScanResultIndex = index
	task.ScanResultKeys = keys
}
