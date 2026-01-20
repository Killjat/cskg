package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	// 测试Kafka连接
	fmt.Println("测试Kafka连接...")

	// 创建Writer
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "regular_tasks",
	})
	defer writer.Close()

	// 创建测试消息
	testTask := map[string]interface{}{
		"task_id":   "test-123",
		"ip":        "8.8.8.8",
		"task_type": "port_scan_default",
		"timestamp": time.Now().Unix(),
	}

	data, err := json.Marshal(testTask)
	if err != nil {
		log.Fatalf("序列化失败: %v", err)
	}

	// 发送消息
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message := kafka.Message{
		Key:   []byte("test"),
		Value: data,
	}

	err = writer.WriteMessages(ctx, message)
	if err != nil {
		log.Fatalf("发送消息失败: %v", err)
	}

	fmt.Println("✅ Kafka连接测试成功!")

	// 测试Reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "regular_tasks",
		GroupID: "test-group",
	})
	defer reader.Close()

	fmt.Println("尝试读取消息...")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	msg, err := reader.ReadMessage(ctx2)
	if err != nil {
		log.Printf("读取消息失败: %v", err)
	} else {
		fmt.Printf("✅ 读取到消息: %s\n", string(msg.Value))
	}
}