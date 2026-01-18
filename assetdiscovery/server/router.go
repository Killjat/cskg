package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/cskg/assetdiscovery/common"
)

// Router 任务路由器结构体
type Router struct {
	server *Server
}

// NewRouter 创建新的路由器
func NewRouter(server *Server) *Router {
	return &Router{
		server: server,
	}
}

// RouteTask 路由任务到客户端
func (r *Router) RouteTask(task *common.Task) {
	// 1. 添加任务到管理器
	r.server.taskManager.AddTask(task)

	// 2. 选择合适的客户端
	clientID := r.selectClient(task)
	if clientID == "" {
		log.Printf("No available clients for task %s", task.TaskID)
		return
	}

	// 3. 设置任务的客户端ID
	task.ClientID = clientID

	// 4. 发送任务到Kafka
	r.sendTaskToKafka(task)

	// 5. 更新客户端状态
	r.server.clientManager.UpdateClientStatus(clientID, 1, 0)

	log.Printf("Routed task %s to client %s: %s - %s", 
		task.TaskID, clientID, task.TaskType, task.Target)
}

// selectClient 选择合适的客户端
func (r *Router) selectClient(task *common.Task) string {
	// 获取可用客户端
	availableClients := r.server.clientManager.GetAvailableClients()
	if len(availableClients) == 0 {
		return ""
	}

	// 简单的负载均衡策略：选择活跃任务最少的客户端
	return r.server.clientManager.GetLeastLoadedClient()
}

// sendTaskToKafka 将任务发送到Kafka
func (r *Router) sendTaskToKafka(task *common.Task) {
	// 序列化任务
	taskJSON, err := json.Marshal(task)
	if err != nil {
		log.Printf("Error marshaling task: %v", err)
		return
	}

	// 创建Kafka消息
	msg := kafka.Message{
		Key:   []byte(task.ClientID),
		Value: taskJSON,
		Time:  time.Now(),
	}

	// 发送消息到Kafka
	if err := r.server.kafkaProducer.WriteMessages(nil, msg); err != nil {
		log.Printf("Error sending task to Kafka: %v", err)
		return
	}

	log.Printf("Task sent to Kafka: %s", task.TaskID)
}

// RouteTaskDirect 直接路由任务到指定客户端
func (r *Router) RouteTaskDirect(task *common.Task, clientID string) {
	// 添加任务到管理器
	r.server.taskManager.AddTask(task)

	// 设置任务的客户端ID
	task.ClientID = clientID

	// 发送任务到Kafka
	r.sendTaskToKafka(task)

	// 更新客户端状态
	r.server.clientManager.UpdateClientStatus(clientID, 1, 0)

	log.Printf("Directly routed task %s to client %s: %s - %s", 
		task.TaskID, clientID, task.TaskType, task.Target)
}
