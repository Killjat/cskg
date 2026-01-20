package search

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cskg/CyberStroll/internal/storage"
)

// WebHandler Web处理器
type WebHandler struct {
	searchEngine *SearchEngine
	templates    *template.Template
	logger       *log.Logger
}

// NewWebHandler 创建Web处理器
func NewWebHandler(searchEngine *SearchEngine, logger *log.Logger) *WebHandler {
	return &WebHandler{
		searchEngine: searchEngine,
		logger:       logger,
	}
}

// LoadTemplates 加载模板
func (wh *WebHandler) LoadTemplates(templateDir string) error {
	// 创建带有自定义函数的模板
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
	}

	pattern := filepath.Join(templateDir, "*.html")
	templates, err := template.New("").Funcs(funcMap).ParseGlob(pattern)
	if err != nil {
		return fmt.Errorf("加载模板失败: %v", err)
	}
	wh.templates = templates
	return nil
}

// SetupRoutes 设置路由
func (wh *WebHandler) SetupRoutes(mux *http.ServeMux) {
	// 静态文件
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Web页面
	mux.HandleFunc("/", wh.handleIndex)
	mux.HandleFunc("/search", wh.handleSearch)
	mux.HandleFunc("/asset/", wh.handleAssetDetail)

	// API接口
	mux.HandleFunc("/api/search", wh.handleAPISearch)
	mux.HandleFunc("/api/asset/", wh.handleAPIAssetInfo)
	mux.HandleFunc("/api/recent", wh.handleAPIRecentScans)
	mux.HandleFunc("/api/stats", wh.handleAPIStats)
}

// handleIndex 处理首页
func (wh *WebHandler) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// 获取最近扫描数据
	recentScans, err := wh.searchEngine.GetRecentScans(10)
	if err != nil {
		wh.logger.Printf("获取最近扫描失败: %v", err)
		recentScans = []storage.ScanDocument{}
	}

	data := map[string]interface{}{
		"Title":       "CyberStroll - 网络空间搜索引擎",
		"RecentScans": recentScans,
	}

	if err := wh.renderTemplate(w, "index.html", data); err != nil {
		wh.logger.Printf("渲染模板失败: %v", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// handleSearch 处理搜索页面
func (wh *WebHandler) handleSearch(w http.ResponseWriter, r *http.Request) {
	// 解析搜索参数
	req := &SearchRequest{
		IP:      r.URL.Query().Get("ip"),
		Port:    r.URL.Query().Get("port"),
		Banner:  r.URL.Query().Get("banner"),
		Service: r.URL.Query().Get("service"),
		Country: r.URL.Query().Get("country"),
		Page:    1,
		Size:    20,
	}

	// 解析页码
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	// 解析每页大小
	if sizeStr := r.URL.Query().Get("size"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 && size <= 100 {
			req.Size = size
		}
	}

	// 执行搜索
	var response *SearchResponse
	var err error

	if req.IP != "" || req.Port != "" || req.Banner != "" || req.Service != "" || req.Country != "" {
		response, err = wh.searchEngine.Search(req)
		if err != nil {
			wh.logger.Printf("搜索失败: %v", err)
			http.Error(w, fmt.Sprintf("搜索失败: %v", err), http.StatusBadRequest)
			return
		}
	} else {
		// 空搜索，返回最近扫描
		recentScans, _ := wh.searchEngine.GetRecentScans(req.Size)
		response = &SearchResponse{
			Total:   int64(len(recentScans)),
			Page:    1,
			Size:    req.Size,
			Results: recentScans,
			Stats:   &SearchStats{},
		}
	}

	data := map[string]interface{}{
		"Title":    "搜索结果 - CyberStroll",
		"Request":  req,
		"Response": response,
	}

	if err := wh.renderTemplate(w, "search.html", data); err != nil {
		wh.logger.Printf("渲染模板失败: %v", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// handleAssetDetail 处理资产详情页面
func (wh *WebHandler) handleAssetDetail(w http.ResponseWriter, r *http.Request) {
	// 提取IP地址
	path := strings.TrimPrefix(r.URL.Path, "/asset/")
	ip := strings.Split(path, "/")[0]

	if ip == "" {
		http.Error(w, "无效的IP地址", http.StatusBadRequest)
		return
	}

	// 获取资产信息
	assets, err := wh.searchEngine.GetAssetInfo(ip)
	if err != nil {
		wh.logger.Printf("获取资产信息失败: %v", err)
		http.Error(w, "获取资产信息失败", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":  fmt.Sprintf("资产详情 - %s", ip),
		"IP":     ip,
		"Assets": assets,
	}

	if err := wh.renderTemplate(w, "asset_detail.html", data); err != nil {
		wh.logger.Printf("渲染模板失败: %v", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// handleAPISearch 处理API搜索
func (wh *WebHandler) handleAPISearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 解析搜索参数
	req := &SearchRequest{}
	if err := wh.parseSearchRequest(r, req); err != nil {
		wh.writeJSONError(w, fmt.Sprintf("参数错误: %v", err), http.StatusBadRequest)
		return
	}

	// 执行搜索
	response, err := wh.searchEngine.Search(req)
	if err != nil {
		wh.logger.Printf("API搜索失败: %v", err)
		wh.writeJSONError(w, fmt.Sprintf("搜索失败: %v", err), http.StatusInternalServerError)
		return
	}

	wh.writeJSON(w, response)
}

// handleAPIAssetInfo 处理API资产信息
func (wh *WebHandler) handleAPIAssetInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 提取IP地址
	path := strings.TrimPrefix(r.URL.Path, "/api/asset/")
	ip := strings.Split(path, "/")[0]

	if ip == "" {
		wh.writeJSONError(w, "无效的IP地址", http.StatusBadRequest)
		return
	}

	// 获取资产信息
	assets, err := wh.searchEngine.GetAssetInfo(ip)
	if err != nil {
		wh.logger.Printf("获取资产信息失败: %v", err)
		wh.writeJSONError(w, "获取资产信息失败", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"ip":     ip,
		"assets": assets,
		"total":  len(assets),
	}

	wh.writeJSON(w, response)
}

// handleAPIRecentScans 处理API最近扫描
func (wh *WebHandler) handleAPIRecentScans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 解析限制数量
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	// 获取最近扫描
	scans, err := wh.searchEngine.GetRecentScans(limit)
	if err != nil {
		wh.logger.Printf("获取最近扫描失败: %v", err)
		wh.writeJSONError(w, "获取最近扫描失败", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"scans": scans,
		"total": len(scans),
		"limit": limit,
	}

	wh.writeJSON(w, response)
}

// handleAPIStats 处理API统计信息
func (wh *WebHandler) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 获取Elasticsearch统计
	esStats, err := wh.searchEngine.esClient.GetStats()
	if err != nil {
		wh.logger.Printf("获取ES统计失败: %v", err)
		esStats = map[string]interface{}{}
	}

	response := map[string]interface{}{
		"elasticsearch": esStats,
		"timestamp":     fmt.Sprintf("%d", time.Now().Unix()),
	}

	wh.writeJSON(w, response)
}

// parseSearchRequest 解析搜索请求
func (wh *WebHandler) parseSearchRequest(r *http.Request, req *SearchRequest) error {
	req.IP = r.URL.Query().Get("ip")
	req.Port = r.URL.Query().Get("port")
	req.Banner = r.URL.Query().Get("banner")
	req.Service = r.URL.Query().Get("service")
	req.Country = r.URL.Query().Get("country")
	req.SortBy = r.URL.Query().Get("sort_by")

	// 解析页码
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		} else {
			return fmt.Errorf("无效的页码: %s", pageStr)
		}
	} else {
		req.Page = 1
	}

	// 解析每页大小
	if sizeStr := r.URL.Query().Get("size"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 && size <= 1000 {
			req.Size = size
		} else {
			return fmt.Errorf("无效的每页大小: %s", sizeStr)
		}
	} else {
		req.Size = 20
	}

	// 解析排序方向
	if sortDesc := r.URL.Query().Get("sort_desc"); sortDesc == "true" {
		req.SortDesc = true
	}

	return nil
}

// renderTemplate 渲染模板
func (wh *WebHandler) renderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	if wh.templates == nil {
		return fmt.Errorf("模板未加载")
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return wh.templates.ExecuteTemplate(w, name, data)
}

// writeJSON 写入JSON响应
func (wh *WebHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		wh.logger.Printf("JSON编码失败: %v", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// writeJSONError 写入JSON错误响应
func (wh *WebHandler) writeJSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	
	response := map[string]interface{}{
		"error":   true,
		"message": message,
		"code":    code,
	}
	
	json.NewEncoder(w).Encode(response)
}