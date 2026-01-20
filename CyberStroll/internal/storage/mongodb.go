package storage

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoClient MongoDB客户端
type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
	config   *MongoConfig
}

// MongoConfig MongoDB配置
type MongoConfig struct {
	URI      string `yaml:"uri"`
	Database string `yaml:"database"`
	Timeout  int    `yaml:"timeout"`
}

// Task 任务结构
type Task struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TaskID          string             `bson:"task_id" json:"task_id"`
	TaskInitiator   string             `bson:"task_initiator" json:"task_initiator"`
	TaskTarget      string             `bson:"task_target" json:"task_target"`
	TaskType        string             `bson:"task_type" json:"task_type"`
	TaskCategory    string             `bson:"task_category" json:"task_category"`
	TaskStatus      string             `bson:"task_status" json:"task_status"`
	TargetCount     int                `bson:"target_count" json:"target_count"`
	CompletedCount  int                `bson:"completed_count" json:"completed_count"`
	FailedCount     int                `bson:"failed_count" json:"failed_count"`
	CreatedTime     time.Time          `bson:"created_time" json:"created_time"`
	StartedTime     *time.Time         `bson:"started_time,omitempty" json:"started_time,omitempty"`
	CompletedTime   *time.Time         `bson:"completed_time,omitempty" json:"completed_time,omitempty"`
	Progress        float64            `bson:"progress" json:"progress"`
	Config          TaskConfig         `bson:"config" json:"config"`
	StatusHistory   []StatusChange     `bson:"status_history" json:"status_history"`
}

// TaskConfig 任务配置
type TaskConfig struct {
	Ports     []int `bson:"ports" json:"ports"`
	Timeout   int   `bson:"timeout" json:"timeout"`
	Threads   int   `bson:"threads" json:"threads"`
	MaxRetries int  `bson:"max_retries" json:"max_retries"`
}

// StatusChange 状态变更记录
type StatusChange struct {
	Status    string    `bson:"status" json:"status"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
	Reason    string    `bson:"reason" json:"reason"`
	Operator  string    `bson:"operator" json:"operator"`
}

// TaskStatistics 任务统计
type TaskStatistics struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TaskID       string             `bson:"task_id" json:"task_id"`
	IP           string             `bson:"ip" json:"ip"`
	ScanStatus   string             `bson:"scan_status" json:"scan_status"`
	OpenPorts    []int              `bson:"open_ports" json:"open_ports"`
	Services     []string           `bson:"services" json:"services"`
	ScanTime     time.Time          `bson:"scan_time" json:"scan_time"`
	ResponseTime int64              `bson:"response_time" json:"response_time"`
	ErrorMessage string             `bson:"error_message,omitempty" json:"error_message,omitempty"`
}

// SystemIPPool 系统IP池
type SystemIPPool struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	IPRange        string             `bson:"ip_range" json:"ip_range"`
	IP             string             `bson:"ip" json:"ip"`
	Priority       int                `bson:"priority" json:"priority"`
	LastScanTime   *time.Time         `bson:"last_scan_time,omitempty" json:"last_scan_time,omitempty"`
	ScanFrequency  int                `bson:"scan_frequency" json:"scan_frequency"`
	Status         string             `bson:"status" json:"status"`
	CreatedTime    time.Time          `bson:"created_time" json:"created_time"`
}

// NewMongoClient 创建MongoDB客户端
func NewMongoClient(config *MongoConfig) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.URI))
	if err != nil {
		return nil, fmt.Errorf("连接MongoDB失败: %v", err)
	}

	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("MongoDB连接测试失败: %v", err)
	}

	database := client.Database(config.Database)

	mongoClient := &MongoClient{
		client:   client,
		database: database,
		config:   config,
	}

	// 创建索引
	if err := mongoClient.createIndexes(); err != nil {
		return nil, fmt.Errorf("创建索引失败: %v", err)
	}

	return mongoClient, nil
}

// createIndexes 创建索引
func (mc *MongoClient) createIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 任务表索引
	tasksCollection := mc.database.Collection("tasks")
	taskIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "task_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "task_initiator", Value: 1}}},
		{Keys: bson.D{{Key: "task_status", Value: 1}}},
		{Keys: bson.D{{Key: "created_time", Value: -1}}},
		{Keys: bson.D{{Key: "task_category", Value: 1}, {Key: "task_status", Value: 1}}},
	}

	if _, err := tasksCollection.Indexes().CreateMany(ctx, taskIndexes); err != nil {
		return fmt.Errorf("创建任务表索引失败: %v", err)
	}

	// 任务统计表索引
	statsCollection := mc.database.Collection("task_statistics")
	statsIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "task_id", Value: 1}}},
		{Keys: bson.D{{Key: "ip", Value: 1}}},
		{Keys: bson.D{{Key: "scan_time", Value: -1}}},
		{Keys: bson.D{{Key: "task_id", Value: 1}, {Key: "ip", Value: 1}}},
	}

	if _, err := statsCollection.Indexes().CreateMany(ctx, statsIndexes); err != nil {
		return fmt.Errorf("创建统计表索引失败: %v", err)
	}

	// 系统IP池索引
	ipPoolCollection := mc.database.Collection("system_ip_pool")
	ipPoolIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "ip", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "priority", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "last_scan_time", Value: 1}}},
	}

	if _, err := ipPoolCollection.Indexes().CreateMany(ctx, ipPoolIndexes); err != nil {
		return fmt.Errorf("创建IP池索引失败: %v", err)
	}

	return nil
}

// CreateTask 创建任务
func (mc *MongoClient) CreateTask(task *Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	task.CreatedTime = time.Now()
	task.StatusHistory = []StatusChange{
		{
			Status:    task.TaskStatus,
			Timestamp: task.CreatedTime,
			Reason:    "任务创建",
			Operator:  task.TaskInitiator,
		},
	}

	collection := mc.database.Collection("tasks")
	result, err := collection.InsertOne(ctx, task)
	if err != nil {
		return fmt.Errorf("创建任务失败: %v", err)
	}

	task.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetTask 获取任务
func (mc *MongoClient) GetTask(taskID string) (*Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := mc.database.Collection("tasks")
	var task Task

	err := collection.FindOne(ctx, bson.M{"task_id": taskID}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("任务不存在: %s", taskID)
		}
		return nil, fmt.Errorf("获取任务失败: %v", err)
	}

	return &task, nil
}

// UpdateTaskStatus 更新任务状态
func (mc *MongoClient) UpdateTaskStatus(taskID, status, reason, operator string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := mc.database.Collection("tasks")

	// 添加状态变更记录
	statusChange := StatusChange{
		Status:    status,
		Timestamp: time.Now(),
		Reason:    reason,
		Operator:  operator,
	}

	update := bson.M{
		"$set": bson.M{
			"task_status": status,
		},
		"$push": bson.M{
			"status_history": statusChange,
		},
	}

	// 如果是开始状态，设置开始时间
	if status == "running" {
		update["$set"].(bson.M)["started_time"] = time.Now()
	}

	// 如果是完成状态，设置完成时间
	if status == "completed" || status == "failed" {
		update["$set"].(bson.M)["completed_time"] = time.Now()
	}

	result, err := collection.UpdateOne(ctx, bson.M{"task_id": taskID}, update)
	if err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	return nil
}

// UpdateTaskProgress 更新任务进度
func (mc *MongoClient) UpdateTaskProgress(taskID string, completed, failed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 先获取任务信息计算进度
	task, err := mc.GetTask(taskID)
	if err != nil {
		return err
	}

	progress := float64(completed+failed) / float64(task.TargetCount) * 100
	if progress > 100 {
		progress = 100
	}

	collection := mc.database.Collection("tasks")
	update := bson.M{
		"$set": bson.M{
			"completed_count": completed,
			"failed_count":    failed,
			"progress":        progress,
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"task_id": taskID}, update)
	if err != nil {
		return fmt.Errorf("更新任务进度失败: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	return nil
}

// ListTasks 列出任务
func (mc *MongoClient) ListTasks(initiator string, status string, limit int64) ([]*Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := mc.database.Collection("tasks")

	// 构建查询条件
	filter := bson.M{}
	if initiator != "" {
		filter["task_initiator"] = initiator
	}
	if status != "" {
		filter["task_status"] = status
	}

	// 设置查询选项
	opts := options.Find().SetSort(bson.D{{Key: "created_time", Value: -1}})
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %v", err)
	}
	defer cursor.Close(ctx)

	var tasks []*Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("解析任务数据失败: %v", err)
	}

	return tasks, nil
}

// CreateTaskStatistics 创建任务统计
func (mc *MongoClient) CreateTaskStatistics(stats *TaskStatistics) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := mc.database.Collection("task_statistics")
	_, err := collection.InsertOne(ctx, stats)
	if err != nil {
		return fmt.Errorf("创建任务统计失败: %v", err)
	}

	return nil
}

// GetTaskStatistics 获取任务统计
func (mc *MongoClient) GetTaskStatistics(taskID string) ([]*TaskStatistics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := mc.database.Collection("task_statistics")
	cursor, err := collection.Find(ctx, bson.M{"task_id": taskID})
	if err != nil {
		return nil, fmt.Errorf("查询任务统计失败: %v", err)
	}
	defer cursor.Close(ctx)

	var stats []*TaskStatistics
	if err := cursor.All(ctx, &stats); err != nil {
		return nil, fmt.Errorf("解析统计数据失败: %v", err)
	}

	return stats, nil
}

// AddSystemIP 添加系统IP
func (mc *MongoClient) AddSystemIP(ipPool *SystemIPPool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ipPool.CreatedTime = time.Now()
	ipPool.Status = "active"

	collection := mc.database.Collection("system_ip_pool")
	_, err := collection.InsertOne(ctx, ipPool)
	if err != nil {
		return fmt.Errorf("添加系统IP失败: %v", err)
	}

	return nil
}

// GetSystemIPs 获取系统IP列表
func (mc *MongoClient) GetSystemIPs(limit int64) ([]*SystemIPPool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := mc.database.Collection("system_ip_pool")

	// 按优先级和最后扫描时间排序
	opts := options.Find().
		SetSort(bson.D{{Key: "priority", Value: -1}, {Key: "last_scan_time", Value: 1}}).
		SetLimit(limit)

	cursor, err := collection.Find(ctx, bson.M{"status": "active"}, opts)
	if err != nil {
		return nil, fmt.Errorf("查询系统IP失败: %v", err)
	}
	defer cursor.Close(ctx)

	var ips []*SystemIPPool
	if err := cursor.All(ctx, &ips); err != nil {
		return nil, fmt.Errorf("解析IP数据失败: %v", err)
	}

	return ips, nil
}

// UpdateSystemIPScanTime 更新系统IP扫描时间
func (mc *MongoClient) UpdateSystemIPScanTime(ip string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := mc.database.Collection("system_ip_pool")
	now := time.Now()

	result, err := collection.UpdateOne(
		ctx,
		bson.M{"ip": ip},
		bson.M{"$set": bson.M{"last_scan_time": now}},
	)
	if err != nil {
		return fmt.Errorf("更新IP扫描时间失败: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("IP不存在: %s", ip)
	}

	return nil
}

// Close 关闭连接
func (mc *MongoClient) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return mc.client.Disconnect(ctx)
}