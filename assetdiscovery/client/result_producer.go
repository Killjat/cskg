package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/cskg/assetdiscovery/common"
)

// ResultProducer 结果生产者结构体
type ResultProducer struct {
	client *Client
}

// NewResultProducer 创建新的结果生产者
func NewResultProducer(client *Client) *ResultProducer {
	return &ResultProducer{
		client: client,
	}
}

// SendResult 发送扫描结果到Kafka
func (rp *ResultProducer) SendResult(result *common.Result) {
	// 设置结果的客户端ID和时间戳
	result.ClientID = rp.client.config.Client.ClientID
	result.Timestamp = time.Now().Unix()

	// 序列化结果
	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Printf("Error marshaling result: %v", err)
		return
	}

	// 创建Kafka消息
	msg := kafka.Message{
		Key:   []byte(rp.client.config.Client.ClientID),
		Value: resultJSON,
		Time:  time.Now(),
	}

	// 发送消息到Kafka
	if err := rp.client.kafkaProducer.WriteMessages(nil, msg); err != nil {
		log.Printf("Error sending result to Kafka: %v", err)
		return
	}

	log.Printf("Sent result to Kafka: %s:%d - %s", 
		result.Target, result.Port, result.Service)
}

// SendResults 批量发送扫描结果到Kafka
func (rp *ResultProducer) SendResults(results []*common.Result) {
	for _, result := range results {
		rp.SendResult(result)
	}
}
