package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/cskg/CyberStroll/internal/config"
	"github.com/cskg/CyberStroll/internal/elasticsearch"
	"github.com/cskg/CyberStroll/internal/fingerprint"
	"github.com/cskg/CyberStroll/internal/search"
)

var (
	searchService     *search.Service
	fingerprintService *fingerprint.Service
	tmpl              *template.Template
)

func main() {
	fmt.Println("=== CyberStroll Search ===")
	fmt.Println("一键搜索模块 - 负责从Elasticsearch查询扫描结果")
	fmt.Println()

	// 解析命令行参数
	var (
		configPath string
		query      string
		limit      int
		httpAddr   string
		webMode    bool
	)

	flag.StringVar(&configPath, "config", "./config", "配置文件路径")
	flag.StringVar(&query, "query", "", "搜索查询，支持IP、端口、协议、服务、Banner等字段的模糊匹配")
	flag.IntVar(&limit, "limit", 100, "结果数量限制")
	flag.StringVar(&httpAddr, "addr", ":8080", "Web服务地址")
	flag.BoolVar(&webMode, "web", false, "启动Web服务模式")

	help := flag.Bool("help", false, "显示帮助信息")

	flag.Parse()

	if *help {
		showHelp()
		os.Exit(0)
	}

	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("错误: 加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 创建上下文
	ctx := context.Background()

	// 创建Elasticsearch客户端
	esClient, err := elasticsearch.NewClient(&cfg.Elasticsearch)
	if err != nil {
		fmt.Printf("错误: 创建Elasticsearch客户端失败: %v\n", err)
		os.Exit(1)
	}

	// 创建搜索服务
	searchService = search.NewService(esClient)

	// 创建指纹分析服务
	fingerprintService = fingerprint.NewService(searchService)

	// 根据模式执行不同的逻辑
	if webMode {
		startWebServer(httpAddr)
	} else {
		runCLI(ctx, query, limit)
	}
}

// runCLI 运行命令行模式
func runCLI(ctx context.Context, query string, limit int) {
	// 执行搜索
	fmt.Println("正在搜索扫描结果...")
	results, err := searchService.SearchScanResults(ctx, query)
	if err != nil {
		fmt.Printf("错误: 搜索扫描结果失败: %v\n", err)
		os.Exit(1)
	}

	// 显示搜索结果
	fmt.Println(searchService.GetSummary(results))
	fmt.Println()

	// 格式化并输出结果
	formattedResults := searchService.FormatResults(results)
	fmt.Println(formattedResults)

	fmt.Println("搜索完成!")
}

// startWebServer 启动Web服务
func startWebServer(httpAddr string) {
	// 加载模板
	tmpl = template.Must(template.ParseFiles("cmd/search/index.html"))

	// 设置路由
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/fingerprint", fingerprintHandler)

	// 获取本地IP地址
	localIP, err := getLocalIP()
	if err != nil {
		fmt.Printf("警告: 获取本地IP失败: %v，使用默认192.168.1.11\n", err)
		localIP = "192.168.1.11"
	}

	// 启动HTTP服务
	fmt.Printf("Web服务正在启动，监听地址: %s\n", httpAddr)
	fmt.Printf("访问地址: http://%s%s\n", localIP, httpAddr)
	if err := http.ListenAndServe(httpAddr, nil); err != nil {
		fmt.Printf("错误: 启动Web服务失败: %v\n", err)
		os.Exit(1)
	}
}

// getLocalIP 获取本地IP地址
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	
	for _, addr := range addrs {
		// 检查是否是IP地址且不是回环地址
		ipnet, ok := addr.(*net.IPNet)
		if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String(), nil
		}
	}
	
	return "", fmt.Errorf("no valid IP address found")
}

// indexHandler 处理首页请求
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// 渲染首页模板
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// searchHandler 处理搜索请求
func searchHandler(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	r.ParseForm()
	
	// 获取自定义搜索参数
	customQuery := r.FormValue("custom_query")
	
	// 传统表单参数
	ip := r.FormValue("ip")
	portStr := r.FormValue("port")
	protocol := r.FormValue("protocol")
	service := r.FormValue("service")
	app := r.FormValue("app")
	banner := r.FormValue("banner")

	// 构建过滤条件
	filters := make(map[string]interface{})
	
	// 1. 处理自定义搜索格式（ip=xxx port=xxx）
	if customQuery != "" {
		// 解析自定义搜索字符串
		customFilters := parseCustomQuery(customQuery)
		// 添加到主过滤条件
		for k, v := range customFilters {
			filters[k] = v
		}
	} else {
		// 2. 处理传统表单参数
		// 添加IP过滤
		if ip != "" {
			filters["ip"] = ip
		}
		
		// 添加端口过滤
		if portStr != "" {
			if port, err := strconv.Atoi(portStr); err == nil {
				filters["port"] = port
			}
		}
		
		// 添加协议过滤
		if protocol != "" {
			filters["protocol"] = protocol
		}
		
		// 添加服务过滤
		if service != "" {
			filters["service"] = service
		}
		
		// 添加Banner过滤
		if banner != "" {
			filters["banner"] = banner
		}

		// 添加应用过滤
		if app != "" {
			filters["app"] = app
		}
		
		// 添加状态过滤
		status := r.FormValue("status")
		if status != "" {
			filters["status"] = status
		}
	}

	// 执行搜索
	ctx := r.Context()
	results, err := searchService.SearchScanResultsWithFilters(ctx, filters)
	if err != nil {
		http.Error(w, fmt.Sprintf("搜索失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 渲染搜索结果
	if err := tmpl.Execute(w, struct {
		Results *elasticsearch.ScanResultsResponse
		IP      string
		Port    string
		Protocol string
		Service string
		App     string
		Banner  string
	}{
		Results:  results,
		IP:       ip,
		Port:     portStr,
		Protocol: protocol,
		Service:  service,
		App:      app,
		Banner:   banner,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// parseCustomQuery 解析自定义搜索字符串（格式：ip=192.168.1.1 port=80 protocol=tcp）
func parseCustomQuery(query string) map[string]interface{} {
	filters := make(map[string]interface{})
	
	// 分割搜索条件
	parts := strings.Fields(query)
	for _, part := range parts {
		// 分割字段名和值
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue // 跳过格式错误的条件
		}
		
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		
		if key == "" || value == "" {
			continue // 跳过空字段或值
		}
		
		// 尝试将值转换为数字（用于端口等字段）
		if port, err := strconv.Atoi(value); err == nil {
			filters[key] = port
		} else {
			// 字符串值
			filters[key] = value
		}
	}
	
	return filters
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Println("使用方法:")
	fmt.Println("  search [选项]")
	fmt.Println()
	fmt.Println("选项:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  search -query http")
	fmt.Println("  search -query 192.168.1.1")
	fmt.Println("  search -query 80 -limit 50")
	fmt.Println("  search -web -addr :8080")
	fmt.Println("  search -web -config ./custom_config")
}

// fingerprintHandler 处理指纹分析请求
func fingerprintHandler(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	r.ParseForm()
	query := r.FormValue("query")
	field := r.FormValue("field")

	// 执行指纹分析
	ctx := r.Context()
	analysisResults, err := fingerprintService.AnalyzeBannersByQuery(ctx, query, field)
	if err != nil {
		http.Error(w, fmt.Sprintf("指纹分析失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 渲染结果
	if err := tmpl.Execute(w, struct {
		FingerprintResults map[string][]string
		Query              string
		Field              string
		IsFingerprint      bool
	}{
		FingerprintResults: analysisResults,
		Query:              query,
		Field:              field,
		IsFingerprint:      true,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
