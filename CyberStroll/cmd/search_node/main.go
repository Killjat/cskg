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
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cskg/CyberStroll/internal/storage"
	"github.com/cskg/CyberStroll/pkg/config"
)

// SearchNode 搜索节点
type SearchNode struct {
	config   *config.SearchNodeConfig
	esClient *storage.ElasticsearchClient
	server   *http.Server
	logger   *log.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query    string `json:"query"`
	IP       string `json:"ip"`
	Port     string `json:"port"`
	Banner   string `json:"banner"`
	Service  string `json:"service"`
	Protocol string `json:"protocol"`
	Country  string `json:"country"`
	Page     int    `json:"page"`
	Size     int    `json:"size"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Total   int64                    `json:"total"`
	Page    int                      `json:"page"`
	Size    int                      `json:"size"`
	Results []storage.ScanDocument   `json:"results"`
	Stats   map[string]interface{}   `json:"stats"`
}

func main() {
	var (
		configFile = flag.String("config", "configs/search_node.yaml", "配置文件路径")
		port       = flag.Int("port", 8082, "HTTP服务端口")
		testMode   = flag.Bool("test", false, "测试模式")
	)
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadSearchNodeConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 设置端口
	if *port != 8082 {
		cfg.Web.Port = *port
	}

	// 创建日志器
	logger := log.New(os.Stdout, "[SearchNode] ", log.LstdFlags)

	// 测试模式
	if *testMode {
		runTestMode(cfg, logger)
		return
	}

	// 创建搜索节点
	node, err := NewSearchNode(cfg, logger)
	if err != nil {
		log.Fatalf("创建搜索节点失败: %v", err)
	}

	// 启动节点
	logger.Println("启动搜索节点...")
	if err := node.Start(); err != nil {
		log.Fatalf("启动搜索节点失败: %v", err)
	}

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Println("收到退出信号，正在关闭...")
	node.Stop()
}

// NewSearchNode 创建搜索节点
func NewSearchNode(cfg *config.SearchNodeConfig, logger *log.Logger) (*SearchNode, error) {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 创建Elasticsearch客户端
	esClient, err := storage.NewElasticsearchClient(&storage.ESConfig{
		URLs:     cfg.Elasticsearch.URLs,
		Index:    cfg.Elasticsearch.Index,
		Username: cfg.Elasticsearch.Username,
		Password: cfg.Elasticsearch.Password,
		Timeout:  cfg.Elasticsearch.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("创建Elasticsearch客户端失败: %v", err)
	}

	// 创建HTTP服务器
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Web.Host, cfg.Web.Port),
		Handler: mux,
	}

	node := &SearchNode{
		config:   cfg,
		esClient: esClient,
		server:   server,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}

	// 设置HTTP路由
	node.setupRoutes(mux)

	return node, nil
}

// Start 启动搜索节点
func (sn *SearchNode) Start() error {
	sn.logger.Printf("搜索节点启动: HTTP=%s", sn.server.Addr)

	// 启动HTTP服务器
	sn.wg.Add(1)
	go func() {
		defer sn.wg.Done()
		if err := sn.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sn.logger.Printf("HTTP服务器错误: %v", err)
		}
	}()

	return nil
}

// Stop 停止搜索节点
func (sn *SearchNode) Stop() {
	sn.logger.Println("正在停止搜索节点...")

	// 取消上下文
	sn.cancel()

	// 关闭HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sn.server.Shutdown(ctx)

	// 等待所有协程结束
	sn.wg.Wait()

	// 关闭资源
	sn.esClient.Close()

	sn.logger.Println("搜索节点已停止")
}

// setupRoutes 设置HTTP路由
func (sn *SearchNode) setupRoutes(mux *http.ServeMux) {
	// 静态文件
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// API路由
	mux.HandleFunc("/api/search", sn.handleSearch)
	mux.HandleFunc("/api/stats", sn.handleStats)
	mux.HandleFunc("/api/export", sn.handleExport)

	// Web界面
	mux.HandleFunc("/", sn.handleIndex)
	mux.HandleFunc("/search", sn.handleSearchPage)
}

// handleSearch 处理搜索请求
func (sn *SearchNode) handleSearch(w http.ResponseWriter, r *http.Request) {
	// 解析搜索参数
	req := &SearchRequest{
		Query:    r.URL.Query().Get("query"),
		IP:       r.URL.Query().Get("ip"),
		Port:     r.URL.Query().Get("port"),
		Banner:   r.URL.Query().Get("banner"),
		Service:  r.URL.Query().Get("service"),
		Protocol: r.URL.Query().Get("protocol"),
		Country:  r.URL.Query().Get("country"),
		Page:     1,
		Size:     20,
	}

	// 解析分页参数
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if sizeStr := r.URL.Query().Get("size"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 && size <= 100 {
			req.Size = size
		}
	}

	// 构建Elasticsearch查询
	query := sn.buildElasticsearchQuery(req)

	// 执行搜索
	result, err := sn.esClient.SearchDocumentsWithTotal(query)
	if err != nil {
		sn.logger.Printf("搜索失败: %v", err)
		http.Error(w, "搜索失败", http.StatusInternalServerError)
		return
	}

	// 构建响应
	response := &SearchResponse{
		Total:   result.Total,
		Page:    req.Page,
		Size:    req.Size,
		Results: result.Docs,
		Stats:   make(map[string]interface{}),
	}

	// 添加统计信息
	response.Stats = sn.generateStats(result.Docs)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// buildElasticsearchQuery 构建Elasticsearch查询
func (sn *SearchNode) buildElasticsearchQuery(req *SearchRequest) map[string]interface{} {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{},
			},
		},
		"from": (req.Page - 1) * req.Size,
		"size": req.Size,
		"sort": []map[string]interface{}{
			{"scan_time": map[string]string{"order": "desc"}},
		},
	}

	must := query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})

	// 通用查询
	if req.Query != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.Query,
				"fields": []string{"banner", "service", "service_version", "ip"},
			},
		})
	}

	// IP查询
	if req.IP != "" {
		if strings.Contains(req.IP, "/") || strings.Contains(req.IP, "-") {
			// CIDR或范围查询
			must = append(must, map[string]interface{}{
				"range": map[string]interface{}{
					"ip": map[string]interface{}{
						"gte": req.IP,
						"lte": req.IP,
					},
				},
			})
		} else {
			// 精确IP查询
			must = append(must, map[string]interface{}{
				"term": map[string]interface{}{
					"ip": req.IP,
				},
			})
		}
	}

	// 端口查询
	if req.Port != "" {
		if strings.Contains(req.Port, "-") {
			// 端口范围查询
			parts := strings.Split(req.Port, "-")
			if len(parts) == 2 {
				startPort, _ := strconv.Atoi(parts[0])
				endPort, _ := strconv.Atoi(parts[1])
				must = append(must, map[string]interface{}{
					"range": map[string]interface{}{
						"port": map[string]interface{}{
							"gte": startPort,
							"lte": endPort,
						},
					},
				})
			}
		} else {
			// 精确端口查询
			port, _ := strconv.Atoi(req.Port)
			must = append(must, map[string]interface{}{
				"term": map[string]interface{}{
					"port": port,
				},
			})
		}
	}

	// Banner查询
	if req.Banner != "" {
		must = append(must, map[string]interface{}{
			"match": map[string]interface{}{
				"banner": req.Banner,
			},
		})
	}

	// 服务查询
	if req.Service != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"service": req.Service,
			},
		})
	}

	// 协议查询
	if req.Protocol != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"protocol": req.Protocol,
			},
		})
	}

	// 国家查询
	if req.Country != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"geo_info.country": req.Country,
			},
		})
	}

	query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = must

	return query
}

// generateStats 生成统计信息
func (sn *SearchNode) generateStats(docs []storage.ScanDocument) map[string]interface{} {
	stats := map[string]interface{}{
		"total_results": len(docs),
		"services":      make(map[string]int),
		"protocols":     make(map[string]int),
		"countries":     make(map[string]int),
		"ports":         make(map[int]int),
	}

	services := stats["services"].(map[string]int)
	protocols := stats["protocols"].(map[string]int)
	countries := stats["countries"].(map[string]int)
	ports := stats["ports"].(map[int]int)

	for _, doc := range docs {
		if doc.Service != "" {
			services[doc.Service]++
		}
		if doc.Protocol != "" {
			protocols[doc.Protocol]++
		}
		if doc.GeoInfo != nil && doc.GeoInfo.Country != "" {
			countries[doc.GeoInfo.Country]++
		}
		ports[doc.Port]++
	}

	return stats
}

// handleStats 处理统计请求
func (sn *SearchNode) handleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := sn.esClient.GetStats()
	if err != nil {
		http.Error(w, "获取统计失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleExport 处理导出请求
func (sn *SearchNode) handleExport(w http.ResponseWriter, r *http.Request) {
	// 解析搜索参数
	req := &SearchRequest{
		Query:    r.URL.Query().Get("query"),
		IP:       r.URL.Query().Get("ip"),
		Port:     r.URL.Query().Get("port"),
		Banner:   r.URL.Query().Get("banner"),
		Service:  r.URL.Query().Get("service"),
		Protocol: r.URL.Query().Get("protocol"),
		Country:  r.URL.Query().Get("country"),
		Page:     1,
		Size:     1000, // 导出更多数据
	}

	// 构建查询
	query := sn.buildElasticsearchQuery(req)

	// 执行搜索
	docs, err := sn.esClient.SearchDocuments(query)
	if err != nil {
		http.Error(w, "搜索失败", http.StatusInternalServerError)
		return
	}

	// 设置响应头
	format := r.URL.Query().Get("format")
	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=search_results.csv")
		sn.exportCSV(w, docs)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=search_results.json")
		json.NewEncoder(w).Encode(docs)
	}
}

// exportCSV 导出CSV格式
func (sn *SearchNode) exportCSV(w http.ResponseWriter, docs []storage.ScanDocument) {
	// CSV头部
	fmt.Fprintln(w, "IP,Port,Protocol,Service,Banner,Country,ScanTime")

	// 数据行
	for _, doc := range docs {
		country := ""
		if doc.GeoInfo != nil {
			country = doc.GeoInfo.Country
		}

		fmt.Fprintf(w, "%s,%d,%s,%s,\"%s\",%s,%s\n",
			doc.IP,
			doc.Port,
			doc.Protocol,
			doc.Service,
			strings.ReplaceAll(doc.Banner, "\"", "\"\""), // 转义CSV中的引号
			country,
			doc.ScanTime.Format("2006-01-02 15:04:05"),
		)
	}
}

// handleIndex 处理首页
func (sn *SearchNode) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/search", http.StatusFound)
}

// handleSearchPage 处理搜索页面
func (sn *SearchNode) handleSearchPage(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>CyberStroll 网络空间搜索</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: #f5f5f5; }
        
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 20px 0; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header .container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
        .header h1 { font-size: 2.5em; margin-bottom: 10px; }
        .header p { opacity: 0.9; font-size: 1.1em; }
        
        .search-container { max-width: 1200px; margin: 30px auto; padding: 0 20px; }
        .search-box { background: white; padding: 30px; border-radius: 10px; box-shadow: 0 4px 20px rgba(0,0,0,0.1); margin-bottom: 30px; }
        
        .search-input { width: 100%; padding: 15px 20px; font-size: 16px; border: 2px solid #e1e5e9; border-radius: 8px; outline: none; transition: border-color 0.3s; }
        .search-input:focus { border-color: #667eea; }
        
        .filters { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin: 20px 0; }
        .filter-group { display: flex; flex-direction: column; }
        .filter-group label { font-weight: 600; margin-bottom: 5px; color: #333; }
        .filter-group input, .filter-group select { padding: 10px; border: 1px solid #ddd; border-radius: 5px; }
        
        .search-btn { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 15px 30px; border: none; border-radius: 8px; font-size: 16px; cursor: pointer; transition: transform 0.2s; }
        .search-btn:hover { transform: translateY(-2px); }
        
        .results-container { background: white; border-radius: 10px; box-shadow: 0 4px 20px rgba(0,0,0,0.1); overflow: hidden; }
        .results-header { background: #f8f9fa; padding: 20px; border-bottom: 1px solid #e9ecef; display: flex; justify-content: space-between; align-items: center; }
        .results-count { font-weight: 600; color: #333; }
        .export-btn { background: #28a745; color: white; padding: 8px 16px; border: none; border-radius: 5px; cursor: pointer; }
        
        .result-item { padding: 20px; border-bottom: 1px solid #e9ecef; transition: background-color 0.2s; }
        .result-item:hover { background-color: #f8f9fa; }
        .result-item:last-child { border-bottom: none; }
        
        .result-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 10px; }
        .result-ip { font-size: 1.2em; font-weight: 600; color: #667eea; }
        .result-port { background: #e9ecef; padding: 4px 8px; border-radius: 4px; font-size: 0.9em; }
        .result-service { background: #28a745; color: white; padding: 4px 8px; border-radius: 4px; font-size: 0.9em; margin-left: 5px; }
        
        .result-details { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 10px; margin: 10px 0; }
        .detail-item { font-size: 0.9em; }
        .detail-label { font-weight: 600; color: #666; }
        
        .result-banner { background: #f8f9fa; padding: 10px; border-radius: 5px; font-family: monospace; font-size: 0.9em; margin-top: 10px; white-space: pre-wrap; }
        
        .pagination { display: flex; justify-content: center; align-items: center; padding: 20px; gap: 10px; }
        .page-btn { padding: 8px 12px; border: 1px solid #ddd; background: white; cursor: pointer; border-radius: 4px; }
        .page-btn.active { background: #667eea; color: white; border-color: #667eea; }
        
        .stats-panel { background: white; padding: 20px; border-radius: 10px; box-shadow: 0 4px 20px rgba(0,0,0,0.1); margin-bottom: 20px; }
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 15px; }
        .stat-item { text-align: center; padding: 15px; background: #f8f9fa; border-radius: 8px; }
        .stat-number { font-size: 1.5em; font-weight: bold; color: #667eea; }
        .stat-label { color: #666; margin-top: 5px; }
        
        .loading { text-align: center; padding: 40px; color: #666; }
        .error { background: #f8d7da; color: #721c24; padding: 15px; border-radius: 5px; margin: 20px 0; }
        
        @media (max-width: 768px) {
            .filters { grid-template-columns: 1fr; }
            .result-header { flex-direction: column; align-items: flex-start; }
            .result-details { grid-template-columns: 1fr; }
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="container">
            <h1>🔍 CyberStroll</h1>
            <p>网络空间测绘搜索引擎</p>
        </div>
    </div>

    <div class="search-container">
        <div class="search-box">
            <form id="searchForm">
                <input type="text" id="queryInput" class="search-input" placeholder="输入搜索关键词，如：Apache、nginx、SSH等..." autocomplete="off">
                
                <div class="filters">
                    <div class="filter-group">
                        <label>IP地址</label>
                        <input type="text" id="ipInput" placeholder="192.168.1.1 或 192.168.1.0/24">
                    </div>
                    <div class="filter-group">
                        <label>端口</label>
                        <input type="text" id="portInput" placeholder="80 或 80-8080">
                    </div>
                    <div class="filter-group">
                        <label>服务</label>
                        <input type="text" id="serviceInput" placeholder="http, ssh, ftp">
                    </div>
                    <div class="filter-group">
                        <label>协议</label>
                        <select id="protocolInput">
                            <option value="">全部协议</option>
                            <option value="tcp">TCP</option>
                            <option value="udp">UDP</option>
                        </select>
                    </div>
                    <div class="filter-group">
                        <label>国家</label>
                        <input type="text" id="countryInput" placeholder="China, United States">
                    </div>
                    <div class="filter-group">
                        <label>Banner</label>
                        <input type="text" id="bannerInput" placeholder="包含的Banner内容">
                    </div>
                </div>
                
                <button type="submit" class="search-btn">🔍 搜索</button>
            </form>
        </div>

        <div id="statsPanel" class="stats-panel" style="display: none;">
            <div class="stats-grid" id="statsGrid"></div>
        </div>

        <div id="resultsContainer" class="results-container" style="display: none;">
            <div class="results-header">
                <div class="results-count" id="resultsCount">找到 0 条结果</div>
                <div>
                    <button class="export-btn" onclick="exportResults('json')">导出JSON</button>
                    <button class="export-btn" onclick="exportResults('csv')">导出CSV</button>
                </div>
            </div>
            <div id="resultsList"></div>
            <div class="pagination" id="pagination"></div>
        </div>

        <div id="loading" class="loading" style="display: none;">
            <p>🔄 搜索中...</p>
        </div>

        <div id="error" class="error" style="display: none;"></div>
    </div>

    <script>
        let currentPage = 1;
        let currentQuery = {};

        document.getElementById('searchForm').addEventListener('submit', function(e) {
            e.preventDefault();
            currentPage = 1;
            performSearch();
        });

        async function performSearch() {
            const query = {
                query: document.getElementById('queryInput').value,
                ip: document.getElementById('ipInput').value,
                port: document.getElementById('portInput').value,
                service: document.getElementById('serviceInput').value,
                protocol: document.getElementById('protocolInput').value,
                country: document.getElementById('countryInput').value,
                banner: document.getElementById('bannerInput').value,
                page: currentPage,
                size: 20
            };

            currentQuery = query;

            // 显示加载状态
            document.getElementById('loading').style.display = 'block';
            document.getElementById('resultsContainer').style.display = 'none';
            document.getElementById('statsPanel').style.display = 'none';
            document.getElementById('error').style.display = 'none';

            try {
                const params = new URLSearchParams();
                Object.keys(query).forEach(key => {
                    if (query[key]) params.append(key, query[key]);
                });

                const response = await fetch('/api/search?' + params.toString());
                const data = await response.json();

                if (response.ok) {
                    displayResults(data);
                    displayStats(data.stats);
                } else {
                    throw new Error(data.message || '搜索失败');
                }
            } catch (error) {
                document.getElementById('error').textContent = '搜索失败: ' + error.message;
                document.getElementById('error').style.display = 'block';
            } finally {
                document.getElementById('loading').style.display = 'none';
            }
        }

        function displayResults(data) {
            const container = document.getElementById('resultsContainer');
            const countElement = document.getElementById('resultsCount');
            const listElement = document.getElementById('resultsList');

            countElement.textContent = '找到 ' + data.total + ' 条结果';
            
            if (data.results.length === 0) {
                listElement.innerHTML = '<div style="padding: 40px; text-align: center; color: #666;">未找到匹配的结果</div>';
            } else {
                listElement.innerHTML = data.results.map(result => createResultItem(result)).join('');
            }

            displayPagination(data);
            container.style.display = 'block';
        }

        function createResultItem(result) {
            const geoInfo = result.geo_info || {};
            const country = geoInfo.country || '未知';
            const city = geoInfo.city || '';
            const location = city ? country + ', ' + city : country;

            return '<div class="result-item">' +
                '<div class="result-header">' +
                    '<div class="result-ip">' + result.ip + '</div>' +
                    '<div>' +
                        '<span class="result-port">' + result.port + '</span>' +
                        (result.service ? '<span class="result-service">' + result.service + '</span>' : '') +
                    '</div>' +
                '</div>' +
                '<div class="result-details">' +
                    '<div class="detail-item"><span class="detail-label">协议:</span> ' + (result.protocol || '未知') + '</div>' +
                    '<div class="detail-item"><span class="detail-label">状态:</span> ' + (result.state || '未知') + '</div>' +
                    '<div class="detail-item"><span class="detail-label">位置:</span> ' + location + '</div>' +
                    '<div class="detail-item"><span class="detail-label">扫描时间:</span> ' + new Date(result.scan_time).toLocaleString() + '</div>' +
                '</div>' +
                (result.banner ? '<div class="result-banner">' + escapeHtml(result.banner) + '</div>' : '') +
            '</div>';
        }

        function displayStats(stats) {
            if (!stats || stats.total_results === 0) return;

            const panel = document.getElementById('statsPanel');
            const grid = document.getElementById('statsGrid');

            let html = '<div class="stat-item"><div class="stat-number">' + stats.total_results + '</div><div class="stat-label">总结果</div></div>';

            // 显示前5个服务
            const services = Object.entries(stats.services || {}).slice(0, 5);
            services.forEach(([service, count]) => {
                html += '<div class="stat-item"><div class="stat-number">' + count + '</div><div class="stat-label">' + service + '</div></div>';
            });

            grid.innerHTML = html;
            panel.style.display = 'block';
        }

        function displayPagination(data) {
            const pagination = document.getElementById('pagination');
            const totalPages = Math.ceil(data.total / data.size);

            if (totalPages <= 1) {
                pagination.innerHTML = '';
                return;
            }

            let html = '';

            // 上一页
            if (currentPage > 1) {
                html += '<button class="page-btn" onclick="changePage(' + (currentPage - 1) + ')">上一页</button>';
            }

            // 页码
            const startPage = Math.max(1, currentPage - 2);
            const endPage = Math.min(totalPages, currentPage + 2);

            for (let i = startPage; i <= endPage; i++) {
                const activeClass = i === currentPage ? ' active' : '';
                html += '<button class="page-btn' + activeClass + '" onclick="changePage(' + i + ')">' + i + '</button>';
            }

            // 下一页
            if (currentPage < totalPages) {
                html += '<button class="page-btn" onclick="changePage(' + (currentPage + 1) + ')">下一页</button>';
            }

            pagination.innerHTML = html;
        }

        function changePage(page) {
            currentPage = page;
            performSearch();
        }

        function exportResults(format) {
            const params = new URLSearchParams();
            Object.keys(currentQuery).forEach(key => {
                if (currentQuery[key] && key !== 'page') {
                    params.append(key, currentQuery[key]);
                }
            });
            params.append('format', format);

            window.open('/api/export?' + params.toString());
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        // 页面加载时执行一次搜索显示最新数据
        window.addEventListener('load', function() {
            performSearch();
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// runTestMode 运行测试模式
func runTestMode(cfg *config.SearchNodeConfig, logger *log.Logger) {
	logger.Println("运行测试模式...")

	logger.Println("✅ 搜索节点配置加载成功")
	logger.Printf("   Elasticsearch: %v", cfg.Elasticsearch.URLs)
	logger.Printf("   索引: %s", cfg.Elasticsearch.Index)
	logger.Printf("   Web端口: %d", cfg.Web.Port)

	logger.Println("🎉 测试模式完成!")
}