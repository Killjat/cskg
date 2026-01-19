package analyzer

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"api-hunter/storage"
)

// JSAnalyzer JavaScript分析器
type JSAnalyzer struct {
	db     *storage.Database
	client *http.Client
}

// NewJSAnalyzer 创建JavaScript分析器
func NewJSAnalyzer(db *storage.Database) *JSAnalyzer {
	return &JSAnalyzer{
		db: db,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// AnalyzeJSFiles 分析JavaScript文件
func (ja *JSAnalyzer) AnalyzeJSFiles(sessionID string) error {
	// 获取未分析的JS文件
	jsFiles, err := ja.db.GetUnanalyzedJSFiles(sessionID, 100)
	if err != nil {
		return fmt.Errorf("获取JS文件失败: %v", err)
	}

	log.Printf("开始分析 %d 个JavaScript文件", len(jsFiles))

	for _, jsFile := range jsFiles {
		if err := ja.analyzeJSFile(&jsFile); err != nil {
			log.Printf("分析JS文件失败 %s: %v", jsFile.URL, err)
			continue
		}

		// 标记为已分析
		if err := ja.db.MarkJSFileAnalyzed(jsFile.ID); err != nil {
			log.Printf("标记JS文件已分析失败: %v", err)
		}
	}

	return nil
}

// analyzeJSFile 分析单个JavaScript文件
func (ja *JSAnalyzer) analyzeJSFile(jsFile *storage.JSFile) error {
	var content string
	var err error

	// 获取JS文件内容
	if jsFile.Content != "" {
		content = jsFile.Content
	} else {
		content, err = ja.fetchJSContent(jsFile.URL)
		if err != nil {
			return fmt.Errorf("获取JS内容失败: %v", err)
		}
		
		// 更新文件大小
		jsFile.Size = int64(len(content))
		jsFile.Content = content
	}

	// 分析API调用
	apis := ja.extractAPIsFromJS(content, jsFile.URL)
	
	// 保存发现的API
	for _, api := range apis {
		if err := ja.db.SaveAPIEndpoint(&api); err != nil {
			log.Printf("保存API失败: %v", err)
		}
	}

	jsFile.APIs = len(apis)
	log.Printf("从 %s 发现 %d 个API", jsFile.URL, len(apis))

	return nil
}

// fetchJSContent 获取JavaScript文件内容
func (ja *JSAnalyzer) fetchJSContent(url string) (string, error) {
	resp, err := ja.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	// 读取内容
	buf := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

// extractAPIsFromJS 从JavaScript中提取API
func (ja *JSAnalyzer) extractAPIsFromJS(content, baseURL string) []storage.APIEndpoint {
	var apis []storage.APIEndpoint
	seen := make(map[string]bool)

	// 提取各种类型的API调用
	fetchAPIs := ja.extractFetchAPIs(content)
	axiosAPIs := ja.extractAxiosAPIs(content)
	jqueryAPIs := ja.extractJQueryAPIs(content)
	xhrAPIs := ja.extractXHRAPIs(content)
	wsAPIs := ja.extractWebSocketAPIs(content)

	// 合并所有API
	allAPIs := append(fetchAPIs, axiosAPIs...)
	allAPIs = append(allAPIs, jqueryAPIs...)
	allAPIs = append(allAPIs, xhrAPIs...)
	allAPIs = append(allAPIs, wsAPIs...)

	// 去重并完善API信息
	for _, api := range allAPIs {
		key := api.Method + ":" + api.URL
		if !seen[key] {
			api.Source = "javascript"
			api.CreatedAt = time.Now()
			api.UpdatedAt = time.Now()
			
			// 设置域名和路径
			if strings.HasPrefix(api.URL, "/") {
				// 相对路径，需要补充域名
				api.Path = api.URL
			} else {
				api.Domain = ja.extractDomain(api.URL)
				api.Path = ja.extractPath(api.URL)
			}
			
			api.Type = ja.detectAPIType(api.URL)
			
			apis = append(apis, api)
			seen[key] = true
		}
	}

	return apis
}

// extractFetchAPIs 提取fetch API调用
func (ja *JSAnalyzer) extractFetchAPIs(content string) []storage.APIEndpoint {
	var apis []storage.APIEndpoint

	// 匹配简单的fetch调用
	simplePattern := regexp.MustCompile(`fetch\s*\(\s*["']([^"']+)["']`)
	matches := simplePattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			apis = append(apis, storage.APIEndpoint{
				URL:    match[1],
				Method: "GET",
			})
		}
	}

	// 匹配带选项的fetch调用
	optionsPattern := regexp.MustCompile(`fetch\s*\(\s*["']([^"']+)["']\s*,\s*{\s*[^}]*method\s*:\s*["']([^"']+)["']`)
	matches = optionsPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 2 {
			apis = append(apis, storage.APIEndpoint{
				URL:    match[1],
				Method: strings.ToUpper(match[2]),
			})
		}
	}

	return apis
}

// extractAxiosAPIs 提取axios API调用
func (ja *JSAnalyzer) extractAxiosAPIs(content string) []storage.APIEndpoint {
	var apis []storage.APIEndpoint

	methods := []string{"get", "post", "put", "delete", "patch", "head", "options"}
	
	for _, method := range methods {
		pattern := regexp.MustCompile(fmt.Sprintf(`axios\.%s\s*\(\s*["']([^"']+)["']`, method))
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				apis = append(apis, storage.APIEndpoint{
					URL:    match[1],
					Method: strings.ToUpper(method),
				})
			}
		}
	}

	// 匹配axios通用调用
	generalPattern := regexp.MustCompile(`axios\s*\(\s*{\s*[^}]*url\s*:\s*["']([^"']+)["'][^}]*method\s*:\s*["']([^"']+)["']`)
	matches := generalPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 2 {
			apis = append(apis, storage.APIEndpoint{
				URL:    match[1],
				Method: strings.ToUpper(match[2]),
			})
		}
	}

	return apis
}

// extractJQueryAPIs 提取jQuery AJAX调用
func (ja *JSAnalyzer) extractJQueryAPIs(content string) []storage.APIEndpoint {
	var apis []storage.APIEndpoint

	// 匹配$.ajax调用
	ajaxPattern := regexp.MustCompile(`\$\.ajax\s*\(\s*{\s*[^}]*url\s*:\s*["']([^"']+)["'][^}]*(?:type|method)\s*:\s*["']([^"']+)["']`)
	matches := ajaxPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 2 {
			apis = append(apis, storage.APIEndpoint{
				URL:    match[1],
				Method: strings.ToUpper(match[2]),
			})
		}
	}

	// 匹配$.get, $.post等快捷方法
	shortcuts := []string{"get", "post", "put", "delete"}
	for _, method := range shortcuts {
		pattern := regexp.MustCompile(fmt.Sprintf(`\$\.%s\s*\(\s*["']([^"']+)["']`, method))
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				apis = append(apis, storage.APIEndpoint{
					URL:    match[1],
					Method: strings.ToUpper(method),
				})
			}
		}
	}

	return apis
}

// extractXHRAPIs 提取XMLHttpRequest调用
func (ja *JSAnalyzer) extractXHRAPIs(content string) []storage.APIEndpoint {
	var apis []storage.APIEndpoint

	// 匹配xhr.open调用
	xhrPattern := regexp.MustCompile(`\.open\s*\(\s*["']([^"']+)["']\s*,\s*["']([^"']+)["']`)
	matches := xhrPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 2 {
			apis = append(apis, storage.APIEndpoint{
				URL:    match[2],
				Method: strings.ToUpper(match[1]),
			})
		}
	}

	return apis
}

// extractWebSocketAPIs 提取WebSocket连接
func (ja *JSAnalyzer) extractWebSocketAPIs(content string) []storage.APIEndpoint {
	var apis []storage.APIEndpoint

	// 匹配WebSocket构造函数
	wsPattern := regexp.MustCompile(`new\s+WebSocket\s*\(\s*["']([^"']+)["']`)
	matches := wsPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			apis = append(apis, storage.APIEndpoint{
				URL:    match[1],
				Method: "WEBSOCKET",
				Type:   "WebSocket",
			})
		}
	}

	return apis
}

// detectAPIType 检测API类型
func (ja *JSAnalyzer) detectAPIType(url string) string {
	urlLower := strings.ToLower(url)
	
	if strings.Contains(urlLower, "graphql") {
		return "GraphQL"
	}
	
	if strings.HasPrefix(urlLower, "ws://") || strings.HasPrefix(urlLower, "wss://") {
		return "WebSocket"
	}
	
	return "REST"
}

// extractDomain 提取域名
func (ja *JSAnalyzer) extractDomain(url string) string {
	// 简化实现，实际应该使用url.Parse
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}
	
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	
	return ""
}

// extractPath 提取路径
func (ja *JSAnalyzer) extractPath(url string) string {
	// 简化实现，实际应该使用url.Parse
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		parts := strings.SplitN(url, "/", 4)
		if len(parts) > 3 {
			return "/" + parts[3]
		}
	} else if strings.HasPrefix(url, "/") {
		return url
	}
	
	return "/"
}