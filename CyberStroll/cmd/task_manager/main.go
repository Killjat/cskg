package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/cskg/CyberStroll/internal/kafka"
	"github.com/cskg/CyberStroll/internal/storage"
	"github.com/cskg/CyberStroll/internal/taskmanager"
	"github.com/cskg/CyberStroll/pkg/config"
)

// TaskManagerNode 任务管理节点
type TaskManagerNode struct {
	config      *config.TaskManagerConfig
	storage     *storage.MongoClient
	taskManager *taskmanager.TaskManager
	httpServer  *http.Server
	logger      *log.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

func main() {
	var (
		configFile = flag.String("config", "configs/task_manager.yaml", "配置文件路径")
		port       = flag.Int("port", 8080, "HTTP服务端口")
		testMode   = flag.Bool("test", false, "测试模式")
	)
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadTaskManagerConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 设置端口
	if *port != 8080 {
		cfg.Web.Port = *port
	}

	// 创建日志器
	logger := log.New(os.Stdout, "[TaskManager] ", log.LstdFlags)

	// 测试模式
	if *testMode {
		runTestMode(cfg, logger)
		return
	}

	// 创建任务管理节点
	node, err := NewTaskManagerNode(cfg, logger)
	if err != nil {
		log.Fatalf("创建任务管理节点失败: %v", err)
	}

	// 启动节点
	logger.Println("启动任务管理节点...")
	if err := node.Start(); err != nil {
		log.Fatalf("启动任务管理节点失败: %v", err)
	}

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Println("收到退出信号，正在关闭...")
	node.Stop()
}

// NewTaskManagerNode 创建任务管理节点
func NewTaskManagerNode(cfg *config.TaskManagerConfig, logger *log.Logger) (*TaskManagerNode, error) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 创建MongoDB客户端
	mongoClient, err := storage.NewMongoClient(&storage.MongoConfig{
		URI:      cfg.Storage.MongoDB.URI,
		Database: cfg.Storage.MongoDB.Database,
		Timeout:  cfg.Storage.MongoDB.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("创建MongoDB客户端失败: %v", err)
	}

	// 创建Kafka生产者
	taskProducer := kafka.NewTaskProducer(&cfg.Kafka, logger)

	// 创建Kafka消费者 (用于接收结果)
	resultConsumer := kafka.NewTaskConsumer(&cfg.Kafka, logger)

	// 创建任务管理器
	tmConfig := &taskmanager.TaskManagerConfig{
		MaxTasksPerUser:    10,
		MaxIPsPerTask:      3000,
		SystemTaskInterval: 300,
		EnableSystemTasks:  true,
	}
	taskMgr := taskmanager.NewTaskManager(mongoClient, taskProducer, resultConsumer, tmConfig, logger)

	// 创建HTTP服务器
	mux := http.NewServeMux()
	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Web.Host, cfg.Web.Port),
		Handler: mux,
	}

	node := &TaskManagerNode{
		config:      cfg,
		storage:     mongoClient,
		taskManager: taskMgr,
		httpServer:  httpServer,
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
	}

	// 设置HTTP路由
	node.setupRoutes(mux)

	return node, nil
}

// Start 启动任务管理节点
func (tmn *TaskManagerNode) Start() error {
	tmn.logger.Printf("任务管理节点启动: HTTP=%s", tmn.httpServer.Addr)

	// 启动系统任务生成器
	tmn.wg.Add(1)
	go func() {
		defer tmn.wg.Done()
		tmn.taskManager.StartSystemTaskGenerator()
	}()

	// 启动HTTP服务器
	tmn.wg.Add(1)
	go func() {
		defer tmn.wg.Done()
		if err := tmn.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			tmn.logger.Printf("HTTP服务器错误: %v", err)
		}
	}()

	// 启动统计打印
	tmn.wg.Add(1)
	go tmn.printStats(&tmn.wg)

	return nil
}

// Stop 停止任务管理节点
func (tmn *TaskManagerNode) Stop() {
	tmn.logger.Println("正在停止任务管理节点...")

	// 取消上下文
	tmn.cancel()

	// 关闭HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tmn.httpServer.Shutdown(ctx)

	// 等待所有协程结束
	tmn.wg.Wait()

	// 关闭资源
	tmn.storage.Close()

	tmn.logger.Println("任务管理节点已停止")
}

// setupRoutes 设置HTTP路由
func (tmn *TaskManagerNode) setupRoutes(mux *http.ServeMux) {
	// 静态文件
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// API路由
	mux.HandleFunc("/api/tasks/submit", tmn.handleSubmitTask)
	mux.HandleFunc("/api/tasks/status", tmn.handleTaskStatus)
	mux.HandleFunc("/api/tasks/list", tmn.handleListTasks)
	mux.HandleFunc("/api/stats", tmn.handleStats)

	// Web界面
	mux.HandleFunc("/", tmn.handleIndex)
	mux.HandleFunc("/tasks", tmn.handleTasksPage)
	mux.HandleFunc("/stats", tmn.handleStatsPage)
}

// handleSubmitTask 处理任务提交
func (tmn *TaskManagerNode) handleSubmitTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request taskmanager.TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response, err := tmn.taskManager.SubmitTask(&request)
	if err != nil {
		tmn.logger.Printf("任务提交失败: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleTaskStatus 处理任务状态查询
func (tmn *TaskManagerNode) handleTaskStatus(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		http.Error(w, "Missing task_id parameter", http.StatusBadRequest)
		return
	}

	status, err := tmn.taskManager.GetTaskStatus(taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleListTasks 处理任务列表查询
func (tmn *TaskManagerNode) handleListTasks(w http.ResponseWriter, r *http.Request) {
	initiator := r.URL.Query().Get("initiator")
	limitStr := r.URL.Query().Get("limit")
	
	limit := 50 // 默认限制
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	tasks, err := tmn.taskManager.ListUserTasks(initiator, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// handleStats 处理统计信息查询
func (tmn *TaskManagerNode) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := tmn.taskManager.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleIndex 处理首页
func (tmn *TaskManagerNode) handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>CyberStroll 任务管理</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 5px; }
        .nav { margin: 20px 0; }
        .nav a { margin-right: 20px; text-decoration: none; color: #3498db; }
        .card { border: 1px solid #ddd; padding: 20px; margin: 20px 0; border-radius: 5px; }
        .form-group { margin: 15px 0; }
        .form-group label { display: block; margin-bottom: 5px; font-weight: bold; }
        .form-group input, .form-group select, .form-group textarea { 
            width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 3px; 
        }
        .btn { background: #3498db; color: white; padding: 10px 20px; border: none; border-radius: 3px; cursor: pointer; }
        .btn:hover { background: #2980b9; }
        .result { margin-top: 20px; padding: 15px; border-radius: 3px; }
        .success { background: #d4edda; border: 1px solid #c3e6cb; color: #155724; }
        .error { background: #f8d7da; border: 1px solid #f5c6cb; color: #721c24; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 CyberStroll 任务管理中心</h1>
        <p>分布式网络空间测绘平台</p>
    </div>
    
    <div class="nav">
        <a href="/">首页</a>
        <a href="/tasks">任务管理</a>
        <a href="/stats">统计信息</a>
    </div>

    <div class="card">
        <h2>📋 提交扫描任务</h2>
        <form id="taskForm">
            <div class="form-group">
                <label>任务发起人:</label>
                <input type="text" id="initiator" value="admin" required>
            </div>
            
            <div class="form-group">
                <label>扫描目标 (每行一个IP/CIDR/范围):</label>
                <textarea id="targets" rows="5" placeholder="192.168.1.1&#10;192.168.1.0/24&#10;10.0.0.1-10.0.0.100" required></textarea>
            </div>
            
            <div class="form-group">
                <label>任务类型:</label>
                <select id="taskType">
                    <option value="port_scan_default">默认端口扫描</option>
                    <option value="port_scan_specified">指定端口扫描</option>
                    <option value="port_scan_full">全端口扫描</option>
                    <option value="app_identification">应用识别</option>
                </select>
            </div>
            
            <div class="form-group">
                <label>指定端口 (逗号分隔，仅指定端口扫描时需要):</label>
                <input type="text" id="ports" placeholder="22,80,443,8080">
            </div>
            
            <div class="form-group">
                <label>超时时间 (秒):</label>
                <input type="number" id="timeout" value="10" min="1" max="60">
            </div>
            
            <button type="submit" class="btn">🎯 提交任务</button>
        </form>
        
        <div id="result"></div>
    </div>

    <script>
        document.getElementById('taskForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const resultDiv = document.getElementById('result');
            resultDiv.innerHTML = '<p>正在提交任务...</p>';
            
            const targets = document.getElementById('targets').value.split('\n').filter(t => t.trim());
            const ports = document.getElementById('ports').value.split(',').map(p => parseInt(p.trim())).filter(p => !isNaN(p));
            
            const request = {
                initiator: document.getElementById('initiator').value,
                targets: targets,
                task_type: document.getElementById('taskType').value,
                ports: ports.length > 0 ? ports : null,
                timeout: parseInt(document.getElementById('timeout').value)
            };
            
            try {
                const response = await fetch('/api/tasks/submit', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(request)
                });
                
                const result = await response.json();
                
                if (result.status === 'success') {
                    resultDiv.innerHTML = '<div class="result success"><strong>✅ 任务提交成功!</strong><br>任务ID: ' + result.task_id + '<br>目标数量: ' + result.target_count + '</div>';
                } else {
                    resultDiv.innerHTML = '<div class="result error"><strong>❌ 任务提交失败:</strong><br>' + result.message + '</div>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error"><strong>❌ 网络错误:</strong><br>' + error.message + '</div>';
            }
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// handleTasksPage 处理任务页面
func (tmn *TaskManagerNode) handleTasksPage(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>任务管理 - CyberStroll</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 5px; }
        .nav { margin: 20px 0; }
        .nav a { margin-right: 20px; text-decoration: none; color: #3498db; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #f2f2f2; }
        .status-pending { color: #f39c12; }
        .status-running { color: #3498db; }
        .status-completed { color: #27ae60; }
        .status-failed { color: #e74c3c; }
    </style>
</head>
<body>
    <div class="header">
        <h1>📋 任务管理</h1>
    </div>
    
    <div class="nav">
        <a href="/">首页</a>
        <a href="/tasks">任务管理</a>
        <a href="/stats">统计信息</a>
    </div>

    <div>
        <h2>任务列表</h2>
        <div id="tasks">加载中...</div>
    </div>

    <script>
        async function loadTasks() {
            try {
                const response = await fetch('/api/tasks/list?limit=20');
                const tasks = await response.json();
                
                let html = '<table><tr><th>任务ID</th><th>发起人</th><th>类型</th><th>状态</th><th>进度</th><th>创建时间</th></tr>';
                
                tasks.forEach(task => {
                    const statusClass = 'status-' + task.task_status;
                    html += '<tr>';
                    html += '<td>' + task.task_id + '</td>';
                    html += '<td>' + task.task_initiator + '</td>';
                    html += '<td>' + task.task_type + '</td>';
                    html += '<td class="' + statusClass + '">' + task.task_status + '</td>';
                    html += '<td>' + task.progress.toFixed(1) + '%</td>';
                    html += '<td>' + new Date(task.created_time).toLocaleString() + '</td>';
                    html += '</tr>';
                });
                
                html += '</table>';
                document.getElementById('tasks').innerHTML = html;
            } catch (error) {
                document.getElementById('tasks').innerHTML = '<p>加载失败: ' + error.message + '</p>';
            }
        }
        
        loadTasks();
        setInterval(loadTasks, 5000); // 每5秒刷新
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// handleStatsPage 处理统计页面
func (tmn *TaskManagerNode) handleStatsPage(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>统计信息 - CyberStroll</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 5px; }
        .nav { margin: 20px 0; }
        .nav a { margin-right: 20px; text-decoration: none; color: #3498db; }
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-top: 20px; }
        .stat-card { border: 1px solid #ddd; padding: 20px; border-radius: 5px; text-align: center; }
        .stat-number { font-size: 2em; font-weight: bold; color: #3498db; }
        .stat-label { color: #666; margin-top: 10px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>📊 统计信息</h1>
    </div>
    
    <div class="nav">
        <a href="/">首页</a>
        <a href="/tasks">任务管理</a>
        <a href="/stats">统计信息</a>
    </div>

    <div class="stats-grid" id="stats">
        加载中...
    </div>

    <script>
        async function loadStats() {
            try {
                const response = await fetch('/api/stats');
                const stats = await response.json();
                
                const html = '<div class="stat-card"><div class="stat-number">' + stats.total_tasks + '</div><div class="stat-label">总任务数</div></div>' +
                           '<div class="stat-card"><div class="stat-number">' + stats.system_tasks + '</div><div class="stat-label">系统任务</div></div>' +
                           '<div class="stat-card"><div class="stat-number">' + stats.regular_tasks + '</div><div class="stat-label">常规任务</div></div>' +
                           '<div class="stat-card"><div class="stat-number">' + stats.completed_tasks + '</div><div class="stat-label">已完成</div></div>' +
                           '<div class="stat-card"><div class="stat-number">' + stats.failed_tasks + '</div><div class="stat-label">失败任务</div></div>' +
                           '<div class="stat-card"><div class="stat-number">' + stats.active_tasks + '</div><div class="stat-label">活跃任务</div></div>';
                
                document.getElementById('stats').innerHTML = html;
            } catch (error) {
                document.getElementById('stats').innerHTML = '<p>加载失败: ' + error.message + '</p>';
            }
        }
        
        loadStats();
        setInterval(loadStats, 5000); // 每5秒刷新
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// printStats 打印统计信息
func (tmn *TaskManagerNode) printStats(wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-tmn.ctx.Done():
			return
		case <-ticker.C:
			stats := tmn.taskManager.GetStats()
			tmn.logger.Printf("统计信息: 总任务=%d, 系统任务=%d, 常规任务=%d, 已完成=%d, 失败=%d",
				stats.TotalTasks, stats.SystemTasks, stats.RegularTasks, stats.CompletedTasks, stats.FailedTasks)
		}
	}
}

// runTestMode 运行测试模式
func runTestMode(cfg *config.TaskManagerConfig, logger *log.Logger) {
	logger.Println("运行测试模式...")

	// 创建简单的任务管理器测试
	logger.Println("✅ 任务管理器配置加载成功")
	logger.Printf("   MongoDB: %s", cfg.Storage.MongoDB.URI)
	logger.Printf("   Kafka: %v", cfg.Kafka.Brokers)
	logger.Printf("   Web端口: %d", cfg.Web.Port)

	logger.Println("🎉 测试模式完成!")
}