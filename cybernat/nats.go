package main

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

// NATSClient NATS客户端结构体
type NATSClient struct {
	Conn          *nats.Conn
	TaskSubject   string
	ResultSubject string
}

// NewNATSClient 创建NATS客户端
func NewNATSClient(url, taskSubject, resultSubject string) (*NATSClient, error) {
	// 连接到NATS服务器
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("无法连接到NATS服务器: %v", err)
	}

	return &NATSClient{
		Conn:          conn,
		TaskSubject:   taskSubject,
		ResultSubject: resultSubject,
	}, nil
}

// Close 关闭NATS连接
func (c *NATSClient) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}

// PublishTask 发布任务
func (c *NATSClient) PublishTask(task *Task) error {
	// 将任务序列化为JSON
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("无法序列化任务: %v", err)
	}

	// 发布到指定主题
	err = c.Conn.Publish(c.TaskSubject, data)
	if err != nil {
		return fmt.Errorf("无法发布任务: %v", err)
	}

	return nil
}

// SubscribeResults 订阅任务结果
func (c *NATSClient) SubscribeResults(callback func(*TaskResult)) error {
	// 订阅结果主题
	_, err := c.Conn.Subscribe(c.ResultSubject, func(msg *nats.Msg) {
		// 反序列化结果
		var result TaskResult
		err := json.Unmarshal(msg.Data, &result)
		if err != nil {
			fmt.Printf("无法反序列化任务结果: %v\n", err)
			return
		}

		// 调用回调函数
		callback(&result)
	})

	if err != nil {
		return fmt.Errorf("无法订阅任务结果: %v", err)
	}

	return nil
}

// PublishResult 发布任务结果
func (c *NATSClient) PublishResult(result *TaskResult) error {
	// 将结果序列化为JSON
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("无法序列化任务结果: %v", err)
	}

	// 发布到指定主题
	err = c.Conn.Publish(c.ResultSubject, data)
	if err != nil {
		return fmt.Errorf("无法发布任务结果: %v", err)
	}

	return nil
}

// SubscribeTasks 订阅任务
func (c *NATSClient) SubscribeTasks(callback func(*Task)) error {
	// 订阅任务主题
	_, err := c.Conn.Subscribe(c.TaskSubject, func(msg *nats.Msg) {
		// 反序列化任务
		var task Task
		err := json.Unmarshal(msg.Data, &task)
		if err != nil {
			fmt.Printf("无法反序列化任务: %v\n", err)
			return
		}

		// 调用回调函数
		callback(&task)
	})

	if err != nil {
		return fmt.Errorf("无法订阅任务: %v", err)
	}

	return nil
}
