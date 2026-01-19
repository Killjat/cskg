package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// FOFAConfig FOFA配置
type FOFAConfig struct {
	Email   string `json:"email"`
	Key     string `json:"key"`
	BaseURL string `json:"base_url"`
}

// FOFAResponse FOFA API响应
type FOFAResponse struct {
	Error   bool     `json:"error"`
	ErrMsg  string   `json:"errmsg"`
	Query   string   `json:"query"`
	Page    int      `json:"page"`
	Mode    string   `json:"mode"`
	Size    int      `json:"size"`
	Results [][]string `json:"results"`
}

// APIEndpoint 发现的API端点
type APIEndpoint struct {
	URL         string    `json:"url"`
	Method      string    `json:"method"`
	Type        string    `json:"type"`
	Source      string    `json:"source"`
	Domain      string    `json:"domain"`
	Path        string    `json:"path"`
	StatusCode  int       `json:"status_code"`
	ContentType string    `json:"content_type"`
	Response    string    `json:"response"`
	Timestamp   time.Time `json:"timestamp"`
}

// TestResult 测试结果
type TestResult struct {
	TargetURL    string        `json:"target_url"`
	Success      bool          `json:"success"`
	Error        string        `json:"error"`
	APIs         []APIEndpoint `json:"apis"`
	ResponseTime time.Duration `json:"response_time"`
	StatusCode   int           `json:"status_code"`
	ContentType  string        `json:"content_type"`
	PageSize     int           `json:"page_size"`
	Timestamp    time.Time     `json:"timestamp"`
}

// FOFATester FOFA测试器
type FOFATester struct {
	config *FOFAConfig
	client *http.Client
}

// NewFOFATester 创建FOFA测试器
func NewFOFATester(configFile string) (*FOFATester, error) {
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &FOFATester{
		config: config,
		client: client,
	}, nil
}

// loadConfig 加载配置文件
func loadConfig(configFile string) (*FOFAConfig, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config FOFAConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// SearchTargets 搜索目标URL
func (ft *FOFATester) SearchTargets(query string, size int) ([]string, error) {
	// Base64编码查询语句
	encodedQuery := base64.StdEncoding.EncodeToString([]byte(query))

	// 构建请求URL
	params := url.Values{}
	params.Set("email", ft.config.Email)
	params.Set("key", ft.config.Key)
	params.Set("qbase64", encodedQuery)
	params.Set("size", fmt.Sprintf("%d", size))
	params.Set("page", "1")
	params.Set("fields", "host,port,protocol,title")

	requestURL := ft.config.BaseURL + "?" + params.Encode()

	// 发送请求
	resp, err := ft.client.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("FOFA API请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应
	var fofaResp FOFAResponse
	if err := json.Unmarshal(body, &fofaResp); err != nil {
		return nil, fmt.Errorf("解析FOFA响应失败: %v", err)
	}

	if fofaResp.Error {
		return nil, fmt.Errorf("FOFA API错误: %s", fofaResp.ErrMsg)
	}

	// 提取URL
	var urls []string
	for _, result := range fofaResp.Results {
		if len(result) >= 3 {
			host := result[0]
			port := result[1]
			protocol := result[2]

			var targetURL string
			if port == "80" && protocol == "http" {
				targetURL = fmt.Sprintf("http://%s", host)
			} else if port == "443" && protocol == "https" {
				targetURL = fmt.Sprintf("https://%s", host)
			} else {
				targetURL = fmt.Sprintf("%s://%s:%s", protocol, host, port)
			}

			urls = append(urls, targetURL)
		}
	}

	return urls, nil
}

// TestAPIExtraction 测试API提取
func (ft *FOFATester) TestAPIExtraction(targetURL string) *TestResult {
	result := &TestResult{
		TargetURL: targetURL,
		Timestamp: time.Now(),
		APIs:      []APIEndpoint{},
	}

	start := time.Now()

	// 获取页面内容
	resp, err := ft.client.Get(targetURL)
	if err != nil {
		result.Error = fmt.Sprintf("请求失败: %v", err)
		result.ResponseTime = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.ContentType = resp.Header.Get("Content-Type")
	result.ResponseTime = time.Since(start)

	// 读取页面内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("读取响应失败: %v", err)
		return result
	}

	content := string(body)
	result.PageSize = len(content)

	// 提取API
	apis := ft.extractAPIs(content, targetURL)
	result.APIs = apis
	result.Success = true

	return result
}

// extractAPIs 提取API端点
func (ft *FOFATester) extractAPIs(content, baseURL string) []APIEndpoint {
	var apis []APIEndpoint
	seen := make(map[string]bool)

	// 提取各种类型的API
	fetchAPIs := ft.extractFetchAPIs(content)
	axiosAPIs := ft.extractAxiosAPIs(content)
	jqueryAPIs := ft.extractJQueryAPIs(content)
	xhrAPIs := ft.extractXHRAPIs(content)
	wsAPIs := ft.extractWebSocketAPIs(content)
	restAPIs := ft.extractRESTAPIs(content)
	jsonAPIs := ft.extractJSONAPIs(content)

	// 合并所有API
	allAPIs := append(fetchAPIs, axiosAPIs...)
	allAPIs = append(allAPIs, jqueryAPIs...)
	allAPIs = append(allAPIs, xhrAPIs...)
	allAPIs = append(allAPIs, wsAPIs...)
	allAPIs = append(allAPIs, restAPIs...)
	allAPIs = append(allAPIs, jsonAPIs...)

	// 去重并完善API信息
	for _, api := range allAPIs {
		key := api.Method + ":" + api.URL
		if !seen[key] {
			api.Domain = ft.extractDomain(api.URL)
			api.Path = ft.extractPath(api.URL)
			api.Timestamp = time.Now()
			
			// 解析相对URL
			if strings.HasPrefix(api.URL, "/") {
				api.URL = baseURL + api.URL
				api.Domain = ft.extractDomain(baseURL)
			}
			
			apis = append(apis, api)
			seen[key] = true
		}
	}

	return apis
}

// extractFetchAPIs 提取fetch API调用
func (ft *FOFATester) extractFetchAPIs(content string) []APIEndpoint {
	var apis []APIEndpoint

	// 简单fetch调用
	pattern1 := regexp.MustCompile(`fetch\s*\(\s*["']([^"']+)["']`)
	matches := pattern1.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 && ft.isValidAPIURL(match[1]) {
			apis = append(apis, APIEndpoint{
				URL:    match[1],
				Method: "GET",
				Type:   "REST",
				Source: "fetch",
			})
		}
	}

	// 带选项的fetch调用
	pattern2 := regexp.MustCompile(`fetch\s*\(\s*["']([^"']+)["']\s*,\s*{\s*[^}]*method\s*:\s*["']([^"']+)["']`)
	matches = pattern2.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 2 && ft.isValidAPIURL(match[1]) {
			apis = append(apis, APIEndpoint{
				URL:    match[1],
				Method: strings.ToUpper(match[2]),
				Type:   "REST",
				Source: "fetch",
			})
		}
	}

	return apis
}

// extractAxiosAPIs 提取axios API调用
func (ft *FOFATester) extractAxiosAPIs(content string) []APIEndpoint {
	var apis []APIEndpoint

	methods := []string{"get", "post", "put", "delete", "patch", "head", "options"}
	
	for _, method := range methods {
		pattern := regexp.MustCompile(fmt.Sprintf(`axios\.%s\s*\(\s*["']([^"']+)["']`, method))
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 && ft.isValidAPIURL(match[1]) {
				apis = append(apis, APIEndpoint{
					URL:    match[1],
					Method: strings.ToUpper(method),
					Type:   "REST",
					Source: "axios",
				})
			}
		}
	}

	return apis
}

// extractJQueryAPIs 提取jQuery AJAX调用
func (ft *FOFATester) extractJQueryAPIs(content string) []APIEndpoint {
	var apis []APIEndpoint

	// $.ajax调用
	pattern1 := regexp.MustCompile(`\$\.ajax\s*\(\s*{\s*[^}]*url\s*:\s*["']([^"']+)["']`)
	matches := pattern1.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 && ft.isValidAPIURL(match[1]) {
			apis = append(apis, APIEndpoint{
				URL:    match[1],
				Method: "GET",
				Type:   "REST",
				Source: "jquery",
			})
		}
	}

	// $.get, $.post等
	shortcuts := []string{"get", "post", "put", "delete"}
	for _, method := range shortcuts {
		pattern := regexp.MustCompile(fmt.Sprintf(`\$\.%s\s*\(\s*["']([^"']+)["']`, method))
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 && ft.isValidAPIURL(match[1]) {
				apis = append(apis, APIEndpoint{
					URL:    match[1],
					Method: strings.ToUpper(method),
					Type:   "REST",
					Source: "jquery",
				})
			}
		}
	}

	return apis
}

// extractXHRAPIs 提取XMLHttpRequest调用
func (ft *FOFATester) extractXHRAPIs(content string) []APIEndpoint {
	var apis []APIEndpoint

	pattern := regexp.MustCompile(`\.open\s*\(\s*["']([^"']+)["']\s*,\s*["']([^"']+)["']`)
	matches := pattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 2 && ft.isValidAPIURL(match[2]) {
			apis = append(apis, APIEndpoint{
				URL:    match[2],
				Method: strings.ToUpper(match[1]),
				Type:   "REST",
				Source: "xhr",
			})
		}
	}

	return apis
}

// extractWebSocketAPIs 提取WebSocket连接
func (ft *FOFATester) extractWebSocketAPIs(content string) []APIEndpoint {
	var apis []APIEndpoint

	pattern := regexp.MustCompile(`new\s+WebSocket\s*\(\s*["']([^"']+)["']`)
	matches := pattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			apis = append(apis, APIEndpoint{
				URL:    match[1],
				Method: "WEBSOCKET",
				Type:   "WebSocket",
				Source: "websocket",
			})
		}
	}

	return apis
}

// extractRESTAPIs 提取REST API路径
func (ft *FOFATester) extractRESTAPIs(content string) []APIEndpoint {
	var apis []APIEndpoint

	// API路径模式
	patterns := []string{
		`["']([^"']*/api/[^"']*?)["']`,
		`["']([^"']*/v\d+/[^"']*?)["']`,
		`["']([^"']*/rest/[^"']*?)["']`,
		`["']([^"']*graphql[^"']*?)["']`,
	}

	for _, patternStr := range patterns {
		pattern := regexp.MustCompile(patternStr)
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 && ft.isValidAPIURL(match[1]) {
				apiType := "REST"
				if strings.Contains(strings.ToLower(match[1]), "graphql") {
					apiType = "GraphQL"
				}
				
				apis = append(apis, APIEndpoint{
					URL:    match[1],
					Method: "GET",
					Type:   apiType,
					Source: "pattern",
				})
			}
		}
	}

	return apis
}

// extractJSONAPIs 提取JSON端点
func (ft *FOFATester) extractJSONAPIs(content string) []APIEndpoint {
	var apis []APIEndpoint

	pattern := regexp.MustCompile(`["']([^"']*\.json[^"']*?)["']`)
	matches := pattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 && ft.isValidAPIURL(match[1]) {
			apis = append(apis, APIEndpoint{
				URL:    match[1],
				Method: "GET",
				Type:   "REST",
				Source: "json",
			})
		}
	}

	return apis
}

// isValidAPIURL 检查是否为有效的API URL
func (ft *FOFATester) isValidAPIURL(url string) bool {
	if url == "" || len(url) < 3 {
		return false
	}

	// 跳过无效的URL
	invalidPrefixes := []string{
		"javascript:", "mailto:", "tel:", "#", "data:", "blob:",
	}

	urlLower := strings.ToLower(url)
	for _, prefix := range invalidPrefixes {
		if strings.HasPrefix(urlLower, prefix) {
			return false
		}
	}

	// 跳过静态资源
	staticExtensions := []string{
		".css", ".js", ".jpg", ".jpeg", ".png", ".gif", ".svg",
		".ico", ".woff", ".woff2", ".ttf", ".eot", ".pdf",
	}

	for _, ext := range staticExtensions {
		if strings.HasSuffix(urlLower, ext) {
			return false
		}
	}

	// API相关关键词
	apiKeywords := []string{
		"/api/", "/rest/", "/graphql", "/v1/", "/v2/", "/v3/",
		".json", "/ajax/", "/service/", "/endpoint/",
	}

	for _, keyword := range apiKeywords {
		if strings.Contains(urlLower, keyword) {
			return true
		}
	}

	// 相对路径且可能是API
	if strings.HasPrefix(url, "/") && !strings.Contains(url, ".") {
		return true
	}

	return false
}

// extractDomain 提取域名
func (ft *FOFATester) extractDomain(url string) string {
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	} else if strings.HasPrefix(url, "/") {
		return ""
	}

	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[0]
	}

	return ""
}

// extractPath 提取路径
func (ft *FOFATester) extractPath(url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		parts := strings.SplitN(url, "/", 4)
		if len(parts) > 3 {
			return "/" + parts[3]
		}
		return "/"
	} else if strings.HasPrefix(url, "/") {
		return url
	}

	return "/"
}

// RunTest 运行测试
func (ft *FOFATester) RunTest() {
	fmt.Println("🚀 开始FOFA API测试...")
	fmt.Println(strings.Repeat("=", 60))

	// 搜索目标
	fmt.Println("📡 正在从FOFA获取目标URL...")
	
	// 使用多个查询来获取不同类型的网站
	queries := []string{
		"title=\"API\" && country=\"CN\"",
		"body=\"/api/\" && country=\"CN\"",
		"header=\"application/json\" && country=\"CN\"",
		"body=\"axios\" && country=\"CN\"",
		"body=\"fetch(\" && country=\"CN\"",
	}

	var allURLs []string
	for i, query := range queries {
		fmt.Printf("  查询 %d: %s\n", i+1, query)
		urls, err := ft.SearchTargets(query, 20)
		if err != nil {
			fmt.Printf("  ❌ 查询失败: %v\n", err)
			continue
		}
		fmt.Printf("  ✅ 获取到 %d 个URL\n", len(urls))
		allURLs = append(allURLs, urls...)
	}

	// 去重并限制数量
	uniqueURLs := make(map[string]bool)
	var testURLs []string
	for _, url := range allURLs {
		if !uniqueURLs[url] && len(testURLs) < 100 {
			uniqueURLs[url] = true
			testURLs = append(testURLs, url)
		}
	}

	fmt.Printf("\n📊 准备测试 %d 个唯一URL\n", len(testURLs))
	fmt.Println(strings.Repeat("=", 60))

	// 测试API提取
	var results []TestResult
	successCount := 0
	totalAPIs := 0

	for i, targetURL := range testURLs {
		fmt.Printf("\n[%d/%d] 测试: %s\n", i+1, len(testURLs), targetURL)
		
		result := ft.TestAPIExtraction(targetURL)
		results = append(results, *result)

		if result.Success {
			successCount++
			totalAPIs += len(result.APIs)
			fmt.Printf("  ✅ 成功 | 状态码: %d | API数: %d | 响应时间: %v\n", 
				result.StatusCode, len(result.APIs), result.ResponseTime)
			
			// 显示发现的API
			for j, api := range result.APIs {
				if j < 3 { // 只显示前3个
					fmt.Printf("    - %s %s (%s)\n", api.Method, api.Path, api.Source)
				}
			}
			if len(result.APIs) > 3 {
				fmt.Printf("    ... 还有 %d 个API\n", len(result.APIs)-3)
			}
		} else {
			fmt.Printf("  ❌ 失败: %s\n", result.Error)
		}

		// 添加延迟避免请求过快
		time.Sleep(500 * time.Millisecond)
	}

	// 生成测试报告
	ft.generateReport(results, successCount, totalAPIs)
}

// generateReport 生成测试报告
func (ft *FOFATester) generateReport(results []TestResult, successCount, totalAPIs int) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📋 测试报告")
	fmt.Println(strings.Repeat("=", 60))

	// 基本统计
	fmt.Printf("总测试数量: %d\n", len(results))
	fmt.Printf("成功数量: %d\n", successCount)
	fmt.Printf("成功率: %.2f%%\n", float64(successCount)/float64(len(results))*100)
	fmt.Printf("总API数量: %d\n", totalAPIs)
	fmt.Printf("平均每站API数: %.2f\n", float64(totalAPIs)/float64(successCount))

	// 按来源统计API
	sourceStats := make(map[string]int)
	typeStats := make(map[string]int)
	methodStats := make(map[string]int)

	for _, result := range results {
		for _, api := range result.APIs {
			sourceStats[api.Source]++
			typeStats[api.Type]++
			methodStats[api.Method]++
		}
	}

	fmt.Println("\n📊 API来源统计:")
	for source, count := range sourceStats {
		fmt.Printf("  %s: %d\n", source, count)
	}

	fmt.Println("\n📊 API类型统计:")
	for apiType, count := range typeStats {
		fmt.Printf("  %s: %d\n", apiType, count)
	}

	fmt.Println("\n📊 HTTP方法统计:")
	for method, count := range methodStats {
		fmt.Printf("  %s: %d\n", method, count)
	}

	// 保存详细结果到JSON文件
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("fofa_api_test_result_%s.json", timestamp)
	
	reportData := map[string]interface{}{
		"timestamp":     time.Now(),
		"total_tests":   len(results),
		"success_count": successCount,
		"success_rate":  float64(successCount) / float64(len(results)) * 100,
		"total_apis":    totalAPIs,
		"source_stats":  sourceStats,
		"type_stats":    typeStats,
		"method_stats":  methodStats,
		"results":       results,
	}

	if data, err := json.MarshalIndent(reportData, "", "  "); err == nil {
		if err := os.WriteFile(filename, data, 0644); err == nil {
			fmt.Printf("\n💾 详细结果已保存到: %s\n", filename)
		}
	}

	// 显示最佳结果
	fmt.Println("\n🏆 API发现最多的网站:")
	maxAPIs := 0
	var bestResults []TestResult
	
	for _, result := range results {
		if len(result.APIs) > maxAPIs {
			maxAPIs = len(result.APIs)
			bestResults = []TestResult{result}
		} else if len(result.APIs) == maxAPIs && maxAPIs > 0 {
			bestResults = append(bestResults, result)
		}
	}

	for i, result := range bestResults {
		if i < 5 { // 只显示前5个
			fmt.Printf("  %s - %d个API\n", result.TargetURL, len(result.APIs))
		}
	}

	fmt.Println("\n✅ 测试完成!")
}

func main() {
	// 检查配置文件
	configFile := "fofa_config.json"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("❌ 配置文件 %s 不存在\n", configFile)
		fmt.Println("请创建配置文件，内容如下:")
		fmt.Println(`{
  "email": "your_email@example.com",
  "key": "your_fofa_api_key",
  "base_url": "https://fofa.info/api/v1/search/all"
}`)
		return
	}

	// 创建测试器
	tester, err := NewFOFATester(configFile)
	if err != nil {
		log.Fatalf("创建测试器失败: %v", err)
	}

	// 运行测试
	tester.RunTest()
}