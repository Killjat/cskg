package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/cskg/assetdiscovery/common"
)

// WebServer Web服务器结构体
type WebServer struct {
	server     *Server
	templates  *template.Template
}

// NewWebServer 创建新的Web服务器
func NewWebServer(server *Server) *WebServer {
	// 加载模板
	templates, err := template.ParseGlob("../web/templates/*.html")
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	return &WebServer{
		server:    server,
		templates: templates,
	}
}

// Start 启动Web服务器
func (ws *WebServer) Start() error {
	// 设置路由
	mux := http.NewServeMux()

	// 静态文件服务
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../web/static"))))

	// 主页面
	mux.HandleFunc("/", ws.indexHandler)

	// 结果页面
	mux.HandleFunc("/results", ws.resultsHandler)

	// 导出CSV
	mux.HandleFunc("/export/csv", ws.exportCSVHandler)

	// API接口
	mux.HandleFunc("/api/results", ws.apiResultsHandler)
	mux.HandleFunc("/api/tasks", ws.apiTasksHandler)
	mux.HandleFunc("/api/clients", ws.apiClientsHandler)
	mux.HandleFunc("/api/scan", ws.apiScanHandler)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", ws.server.config.Server.Host, ws.server.config.Server.Port)
	log.Printf("Web server starting on %s", addr)

	return http.ListenAndServe(addr, mux)
}

// Shutdown 关闭Web服务器
func (ws *WebServer) Shutdown() error {
	// HTTP服务器没有内置的关闭方法，这里只是一个占位符
	return nil
}

// indexHandler 主页面处理器
func (ws *WebServer) indexHandler(w http.ResponseWriter, r *http.Request) {
	// 准备数据
	data := map[string]interface{}{
		"Title":        "资产探查系统",
		"Version":      "1.0.0",
		"TasksCount":   ws.server.taskManager.GetTasksCount(),
		"ResultsCount": ws.server.resultManager.GetResultsCount(),
		"ClientsCount": len(ws.server.clientManager.GetAllClients()),
	}

	// 渲染模板
	if err := ws.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// resultsHandler 结果页面处理器
func (ws *WebServer) resultsHandler(w http.ResponseWriter, r *http.Request) {
	// 获取所有结果
	results := ws.server.resultManager.GetAllResults()

	// 准备数据
	data := map[string]interface{}{
		"Title":   "扫描结果",
		"Results": results,
		"Count":   len(results),
	}

	// 渲染模板
	if err := ws.templates.ExecuteTemplate(w, "results.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// exportCSVHandler 导出CSV处理器
func (ws *WebServer) exportCSVHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=scan_results.csv")

	// 创建CSV写入器
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// 写入CSV头
	header := []string{
		"Target", "Port", "Protocol", "Service", "Version", 
		"Title", "HasLogin", "ICP", "URL", "Status",
	}
	if err := writer.Write(header); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取所有结果
	results := ws.server.resultManager.GetAllResults()

	// 写入结果数据
	for _, result := range results {
		// 处理Web信息
		var title, icp, url string
		var hasLogin bool

		if result.WebInfo != nil {
			title = result.WebInfo.Title
			hasLogin = result.WebInfo.HasLogin
			url = result.WebInfo.URL

			if result.WebInfo.ICPInfo != nil {
				icp = result.WebInfo.ICPInfo.ICP
			}
		}

		// 写入行
		row := []string{
			result.Target,
			strconv.Itoa(result.Port),
			result.Protocol,
			result.Service,
			result.Version,
			title,
			strconv.FormatBool(hasLogin),
			icp,
			url,
			result.Status,
		}

		if err := writer.Write(row); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	log.Println("Exported scan results to CSV")
}

// apiResultsHandler API结果处理器
func (ws *WebServer) apiResultsHandler(w http.ResponseWriter, r *http.Request) {
	// 获取所有结果
	results := ws.server.resultManager.GetAllResults()

	// 返回JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// apiTasksHandler API任务处理器
func (ws *WebServer) apiTasksHandler(w http.ResponseWriter, r *http.Request) {
	// 获取所有任务
	tasks := ws.server.taskManager.GetAllTasks()

	// 返回JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// apiClientsHandler API客户端处理器
func (ws *WebServer) apiClientsHandler(w http.ResponseWriter, r *http.Request) {
	// 获取所有客户端
	clients := ws.server.clientManager.GetAllClients()

	// 返回JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(clients); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// apiScanHandler API扫描处理器
func (ws *WebServer) apiScanHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求体
	var scanRequest struct {
		Target     string `json:"target"`
		PortRange  string `json:"port_range"`
		TaskType   string `json:"task_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&scanRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 创建扫描任务
	taskType := common.TaskType(scanRequest.TaskType)
	if taskType == "" {
		taskType = common.TaskTypeScanIP
	}

	// 设置默认端口范围
	portRange := scanRequest.PortRange
	if portRange == "" {
		portRange = ws.server.config.Scan.PortRange
	}

	// 创建任务参数
	params := map[string]interface{}{
		"port_range": portRange,
	}

	// 创建任务
	taskID := ws.server.CreateTask(taskType, scanRequest.Target, params)

	// 返回结果
	response := map[string]string{
		"task_id": taskID,
		"status":  "success",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
