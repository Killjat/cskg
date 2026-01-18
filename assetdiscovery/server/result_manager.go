package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/segmentio/kafka-go"

	"github.com/cskg/assetdiscovery/common"
)

// ResultManager 结果管理器结构体
type ResultManager struct {
	server       *Server
	db           *sql.DB
	results      map[string][]*common.Result
	resultsLock  sync.RWMutex
}

// NewResultManager 创建新的结果管理器
func NewResultManager(server *Server) *ResultManager {
	// 初始化数据库
	db, err := sql.Open("sqlite3", server.config.Database.Path)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// 创建结果表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS scan_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id TEXT,
		client_id TEXT,
		target TEXT,
		port INTEGER,
		protocol TEXT,
		service TEXT,
		version TEXT,
		web_info TEXT,
		status TEXT,
		err TEXT,
		timestamp INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	return &ResultManager{
		server:  server,
		db:      db,
		results: make(map[string][]*common.Result),
	}
}

// Start 启动结果处理协程
func (rm *ResultManager) Start() {
	log.Println("Starting result manager...")

	// 启动Kafka结果消费者
	go rm.consumeResults()
}

// consumeResults 消费Kafka结果消息
func (rm *ResultManager) consumeResults() {
	for {
		// 从Kafka读取消息
		msg, err := rm.server.kafkaConsumer.ReadMessage(nil)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		// 解析结果消息
		var result common.Result
		if err := json.Unmarshal(msg.Value, &result); err != nil {
			log.Printf("Error unmarshaling result: %v", err)
			continue
		}

		// 处理结果
		rm.processResult(&result)
	}
}

// processResult 处理扫描结果
func (rm *ResultManager) processResult(result *common.Result) {
	log.Printf("Processing result from client %s: %s:%d - %s", 
		result.ClientID, result.Target, result.Port, result.Service)

	// 保存到内存中
	rm.resultsLock.Lock()
	if _, exists := rm.results[result.Target]; !exists {
		rm.results[result.Target] = make([]*common.Result, 0)
	}
	rm.results[result.Target] = append(rm.results[result.Target], result)
	rm.resultsLock.Unlock()

	// 保存到数据库
	rm.saveResultToDB(result)
}

// saveResultToDB 将结果保存到数据库
func (rm *ResultManager) saveResultToDB(result *common.Result) {
	// 序列化WebInfo
	webInfoJSON, err := json.Marshal(result.WebInfo)
	if err != nil {
		log.Printf("Error marshaling web info: %v", err)
		webInfoJSON = []byte("null")
	}

	// 插入数据库
	insertSQL := `
	INSERT INTO scan_results (task_id, client_id, target, port, protocol, service, version, web_info, status, err, timestamp)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = rm.db.Exec(insertSQL, 
		result.TaskID, result.ClientID, result.Target, result.Port, 
		result.Protocol, result.Service, result.Version, string(webInfoJSON), 
		result.Status, result.Error, result.Timestamp)

	if err != nil {
		log.Printf("Error saving result to DB: %v", err)
	}
}

// GetResultsByTarget 根据目标获取结果
func (rm *ResultManager) GetResultsByTarget(target string) []*common.Result {
	rm.resultsLock.RLock()
	defer rm.resultsLock.RUnlock()

	if results, exists := rm.results[target]; exists {
		return results
	}
	return []*common.Result{}
}

// GetAllResults 获取所有结果
func (rm *ResultManager) GetAllResults() []*common.Result {
	rm.resultsLock.RLock()
	defer rm.resultsLock.RUnlock()

	allResults := make([]*common.Result, 0)
	for _, results := range rm.results {
		allResults = append(allResults, results...)
	}

	return allResults
}

// GetResultsCount 获取结果数量
func (rm *ResultManager) GetResultsCount() int {
	rm.resultsLock.RLock()
	defer rm.resultsLock.RUnlock()

	count := 0
	for _, results := range rm.results {
		count += len(results)
	}

	return count
}
