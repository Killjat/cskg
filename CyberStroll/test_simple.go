package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cskg/CyberStroll/internal/storage"
)

func main() {
	// 测试MongoDB连接
	fmt.Println("测试MongoDB连接...")
	
	mongoClient, err := storage.NewMongoClient(&storage.MongoConfig{
		URI:      "mongodb://cyberstroll:cyberstroll123@localhost:27017/cyberstroll?authSource=admin",
		Database: "cyberstroll",
		Timeout:  10,
	})
	if err != nil {
		log.Fatalf("MongoDB连接失败: %v", err)
	}
	defer mongoClient.Close()

	fmt.Println("✅ MongoDB连接成功!")

	// 创建测试任务
	task := &storage.Task{
		TaskID:        "test-" + fmt.Sprintf("%d", time.Now().Unix()),
		TaskInitiator: "test_user",
		TaskTarget:    "8.8.8.8,1.1.1.1",
		TaskType:      "port_scan_default",
		TaskCategory:  "regular_task",
		TaskStatus:    "pending",
		TargetCount:   2,
		Config: storage.TaskConfig{
			Timeout: 5,
		},
	}

	// 保存任务
	err = mongoClient.CreateTask(task)
	if err != nil {
		log.Fatalf("创建任务失败: %v", err)
	}

	fmt.Printf("✅ 任务创建成功: %s\n", task.TaskID)

	// 查询任务
	retrievedTask, err := mongoClient.GetTask(task.TaskID)
	if err != nil {
		log.Fatalf("查询任务失败: %v", err)
	}

	fmt.Printf("✅ 任务查询成功: %s - %s\n", retrievedTask.TaskID, retrievedTask.TaskStatus)

	// 启动简单的HTTP服务器
	http.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":    "success",
			"message":   "CyberStroll系统运行正常",
			"timestamp": time.Now().Unix(),
			"task_id":   task.TaskID,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	fmt.Println("🚀 启动HTTP服务器: http://localhost:8088")
	fmt.Println("测试URL: http://localhost:8088/api/test")
	
	log.Fatal(http.ListenAndServe(":8088", nil))
}