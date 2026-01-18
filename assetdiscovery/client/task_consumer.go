package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cskg/assetdiscovery/common"
)

// TaskConsumer 任务消费者结构体
type TaskConsumer struct {
	client          *Client
	kafkaActive     bool          // Kafka是否有任务
	localTaskTicker *time.Ticker  // 本地任务执行间隔
}

// NewTaskConsumer 创建新的任务消费者
func NewTaskConsumer(client *Client) *TaskConsumer {
	return &TaskConsumer{
		client:          client,
		kafkaActive:     true,
		localTaskTicker: time.NewTicker(5 * time.Minute), // 每5分钟执行一次本地任务
	}
}

// Start 启动任务消费者
func (tc *TaskConsumer) Start() {
	log.Printf("Task consumer starting for client %s...", tc.client.config.Client.ClientID)

	// 启动本地任务协程
	go tc.runLocalTasks()

	for {
		// 从Kafka读取消息
		msg, err := tc.client.kafkaConsumer.ReadMessage(nil)
		if err != nil {
			// Kafka没有任务，标记为非活跃
			tc.kafkaActive = false
			log.Printf("No Kafka tasks available, will check local tasks: %v", err)
			continue
		}

		// Kafka有任务，标记为活跃
		tc.kafkaActive = true

		// 解析任务消息
		var task common.Task
		if err := json.Unmarshal(msg.Value, &task); err != nil {
			log.Printf("Error unmarshaling task: %v", err)
			continue
		}

		// 检查任务是否属于当前客户端
		if task.ClientID != "" && task.ClientID != tc.client.config.Client.ClientID {
			// 跳过不属于当前客户端的任务
			continue
		}

		// 处理任务
		tc.processTask(&task)
	}
}

// runLocalTasks 运行本地任务
func (tc *TaskConsumer) runLocalTasks() {
	log.Println("Starting local task runner...")

	for {
		select {
		case <-tc.localTaskTicker.C:
			// 只有当Kafka没有任务时，才执行本地任务
			if !tc.kafkaActive {
				tc.readLocalTaskFile()
			}
		}
	}
}

// readLocalTaskFile 读取localtask.txt文件并执行任务
func (tc *TaskConsumer) readLocalTaskFile() {
	log.Println("Reading local task file...")

	// 读取文件
	content, err := os.ReadFile("localtask.txt")
	if err != nil {
		// 文件不存在时不报错，继续下一次读取
		if !os.IsNotExist(err) {
			log.Printf("Error reading localtask.txt: %v", err)
		}
		return
	}

	// 解析文件内容，每行一个IP或网段
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			// 跳过空行和注释行
			continue
		}

		// 创建本地任务
		task := &common.Task{
			TaskID:     "local_" + time.Now().Format("20060102150405"),
			TaskType:   common.TaskTypeScanIP,
			ClientID:   tc.client.config.Client.ClientID,
			Target:     line,
			PortRange:  tc.client.config.Scan.PortRange,
			Parameters: map[string]interface{}{"port_range": tc.client.config.Scan.PortRange},
			Timestamp:  time.Now().Unix(),
		}

		// 处理本地任务
		tc.processTask(task)

		// 较慢的速度执行，每个任务间隔1秒
		time.Sleep(1 * time.Second)
	}
}

// processTask 处理接收到的任务
func (tc *TaskConsumer) processTask(task *common.Task) {
	log.Printf("Received task %s: %s - %s", task.TaskID, task.TaskType, task.Target)

	// 执行扫描任务
	go func() {
		results, err := tc.client.scanExecutor.Execute(task)
		if err != nil {
			log.Printf("Error executing task %s: %v", task.TaskID, err)
			return
		}

		// 发送结果到Kafka
		for _, result := range results {
			tc.client.resultProducer.SendResult(result)
		}

		log.Printf("Completed task %s: %s - %s", task.TaskID, task.TaskType, task.Target)
	}()
}
