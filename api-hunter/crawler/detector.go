package crawler

import (
	"regexp"
	"strings"
	"time"

	"api-hunter/storage"
)

// Detector API检测器
type Detector struct {
	patterns []*APIPattern
}

// APIPattern API模式
type APIPattern struct {
	Name        string
	Pattern     *regexp.Regexp
	Method      string
	Type        string
	Description string
}

// NewDetector 创建检测器
func NewDetector() *Detector {
	detector := &Detector{
		patterns: []*APIPattern{},
	}
	
	detector.initPatterns()
	return detector
}

// initPatterns 初始化检测模式
func (d *Detector) initPatterns() {
	patterns := []struct {
		name        string
		pattern     string
		method      string
		apiType     string
		description string
	}{
		{
			name:        "Fetch API",
			pattern:     `fetch\s*\(\s*["']([^"']+)["']`,
			method:      "GET",
			apiType:     "REST",
			description: "Fetch API调用",
		},
		{
			name:        "Axios GET",
			pattern:     `axios\.get\s*\(\s*["']([^"']+)["']`,
			method:      "GET",
			apiType:     "REST",
			description: "Axios GET请求",
		},
		{
			name:        "Axios POST",
			pattern:     `axios\.post\s*\(\s*["']([^"']+)["']`,
			method:      "POST",
			apiType:     "REST",
			description: "Axios POST请求",
		},
		{
			name:        "jQuery AJAX",
			pattern:     `\$\.ajax\s*\(\s*{\s*[^}]*url\s*:\s*["']([^"']+)["']`,
			method:      "GET",
			apiType:     "REST",
			description: "jQuery AJAX调用",
		},
		{
			name:        "XMLHttpRequest",
			pattern:     `\.open\s*\(\s*["']([^"']+)["']\s*,\s*["']([^"']+)["']`,
			method:      "",
			apiType:     "REST",
			description: "XMLHttpRequest调用",
		},
		{
			name:        "WebSocket",
			pattern:     `new\s+WebSocket\s*\(\s*["']([^"']+)["']`,
			method:      "WEBSOCKET",
			apiType:     "WebSocket",
			description: "WebSocket连接",
		},
		{
			name:        "GraphQL",
			pattern:     `["']([^"']*graphql[^"']*)["']`,
			method:      "POST",
			apiType:     "GraphQL",
			description: "GraphQL端点",
		},
		{
			name:        "REST API Path",
			pattern:     `["']([^"']*/api/[^"']*)["']`,
			method:      "GET",
			apiType:     "REST",
			description: "REST API路径",
		},
		{
			name:        "JSON Endpoint",
			pattern:     `["']([^"']*\.json[^"']*)["']`,
			method:      "GET",
			apiType:     "REST",
			description: "JSON端点",
		},
	}

	for _, p := range patterns {
		compiled, err := regexp.Compile(p.pattern)
		if err != nil {
			continue
		}

		d.patterns = append(d.patterns, &APIPattern{
			Name:        p.name,
			Pattern:     compiled,
			Method:      p.method,
			Type:        p.apiType,
			Description: p.description,
		})
	}
}

// DetectAPIsInJS 在JavaScript代码中检测API
func (d *Detector) DetectAPIsInJS(jsContent string) []storage.APIEndpoint {
	var apis []storage.APIEndpoint
	seen := make(map[string]bool)

	for _, pattern := range d.patterns {
		matches := pattern.Pattern.FindAllStringSubmatch(jsContent, -1)
		
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			var url, method string
			
			// 处理XMLHttpRequest特殊情况
			if pattern.Name == "XMLHttpRequest" && len(match) > 2 {
				method = strings.ToUpper(match[1])
				url = match[2]
			} else {
				url = match[1]
				method = pattern.Method
			}

			// 过滤无效URL
			if !d.isValidAPIURL(url) {
				continue
			}

			// 去重
			key := method + ":" + url
			if seen[key] {
				continue
			}
			seen[key] = true

			api := storage.APIEndpoint{
				URL:         url,
				Method:      method,
				Type:        pattern.Type,
				Source:      "javascript",
				Domain:      d.extractDomain(url),
				Path:        d.extractPath(url),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			apis = append(apis, api)
		}
	}

	return apis
}

// DetectAPIsInHTML 在HTML中检测API
func (d *Detector) DetectAPIsInHTML(htmlContent string) []storage.APIEndpoint {
	var apis []storage.APIEndpoint

	// 检测表单action
	formPattern := regexp.MustCompile(`<form[^>]+action\s*=\s*["']([^"']+)["']`)
	matches := formPattern.FindAllStringSubmatch(htmlContent, -1)
	
	for _, match := range matches {
		if len(match) > 1 && d.isValidAPIURL(match[1]) {
			apis = append(apis, storage.APIEndpoint{
				URL:       match[1],
				Method:    "POST",
				Type:      "REST",
				Source:    "form",
				Domain:    d.extractDomain(match[1]),
				Path:      d.extractPath(match[1]),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}
	}

	// 检测data属性中的API
	dataAPIPattern := regexp.MustCompile(`data-(?:api|url|endpoint)\s*=\s*["']([^"']+)["']`)
	matches = dataAPIPattern.FindAllStringSubmatch(htmlContent, -1)
	
	for _, match := range matches {
		if len(match) > 1 && d.isValidAPIURL(match[1]) {
			apis = append(apis, storage.APIEndpoint{
				URL:       match[1],
				Method:    "GET",
				Type:      "REST",
				Source:    "data-attribute",
				Domain:    d.extractDomain(match[1]),
				Path:      d.extractPath(match[1]),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}
	}

	return apis
}

// isValidAPIURL 检查是否为有效的API URL
func (d *Detector) isValidAPIURL(url string) bool {
	if url == "" {
		return false
	}

	// 跳过无效的URL
	invalidPrefixes := []string{
		"javascript:",
		"mailto:",
		"tel:",
		"#",
		"data:",
		"blob:",
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
		".zip", ".rar", ".tar", ".gz",
	}

	for _, ext := range staticExtensions {
		if strings.HasSuffix(urlLower, ext) {
			return false
		}
	}

	// API相关的关键词
	apiKeywords := []string{
		"/api/", "/rest/", "/graphql", "/v1/", "/v2/", "/v3/",
		".json", "/ajax/", "/service/", "/endpoint/",
	}

	for _, keyword := range apiKeywords {
		if strings.Contains(urlLower, keyword) {
			return true
		}
	}

	// 检查是否为相对路径且可能是API
	if strings.HasPrefix(url, "/") && !strings.Contains(url, ".") {
		return true
	}

	return false
}

// extractDomain 提取域名
func (d *Detector) extractDomain(url string) string {
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	} else if strings.HasPrefix(url, "/") {
		return "" // 相对路径没有域名
	}

	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[0]
	}

	return ""
}

// extractPath 提取路径
func (d *Detector) extractPath(url string) string {
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

// DetectAPIType 检测API类型
func (d *Detector) DetectAPIType(url, content string) string {
	urlLower := strings.ToLower(url)
	contentLower := strings.ToLower(content)

	// GraphQL检测
	if strings.Contains(urlLower, "graphql") || 
	   strings.Contains(contentLower, "query") && strings.Contains(contentLower, "mutation") {
		return "GraphQL"
	}

	// WebSocket检测
	if strings.HasPrefix(urlLower, "ws://") || strings.HasPrefix(urlLower, "wss://") {
		return "WebSocket"
	}

	// REST API检测
	restKeywords := []string{"/api/", "/rest/", "/v1/", "/v2/", "/v3/"}
	for _, keyword := range restKeywords {
		if strings.Contains(urlLower, keyword) {
			return "REST"
		}
	}

	return "REST" // 默认为REST
}

// AddCustomPattern 添加自定义检测模式
func (d *Detector) AddCustomPattern(name, pattern, method, apiType, description string) error {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	d.patterns = append(d.patterns, &APIPattern{
		Name:        name,
		Pattern:     compiled,
		Method:      method,
		Type:        apiType,
		Description: description,
	})

	return nil
}

// GetPatterns 获取所有检测模式
func (d *Detector) GetPatterns() []*APIPattern {
	return d.patterns
}