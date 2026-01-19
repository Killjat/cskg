package analyzer

import (
	"fmt"
	"regexp"
	"strings"

	"api-hunter/storage"
)

// PatternAnalyzer 模式分析器
type PatternAnalyzer struct {
	db       *storage.Database
	patterns []*APIPattern
}

// APIPattern API模式
type APIPattern struct {
	Name        string
	Pattern     *regexp.Regexp
	Type        string
	Method      string
	Description string
	Examples    []string
}

// NewPatternAnalyzer 创建模式分析器
func NewPatternAnalyzer(db *storage.Database) *PatternAnalyzer {
	pa := &PatternAnalyzer{
		db: db,
	}
	pa.initializePatterns()
	return pa
}

// initializePatterns 初始化API模式
func (pa *PatternAnalyzer) initializePatterns() {
	pa.patterns = []*APIPattern{
		// RESTful API模式
		{
			Name:        "RESTful Resource",
			Pattern:     regexp.MustCompile(`/api/v?\d*/[a-zA-Z]+(/\d+)?/?$`),
			Type:        "REST",
			Method:      "GET",
			Description: "标准RESTful资源API",
			Examples:    []string{"/api/users", "/api/v1/posts/123"},
		},
		{
			Name:        "RESTful Collection",
			Pattern:     regexp.MustCompile(`/api/v?\d*/[a-zA-Z]+/?$`),
			Type:        "REST",
			Method:      "GET",
			Description: "RESTful集合API",
			Examples:    []string{"/api/users", "/api/v2/products"},
		},
		{
			Name:        "RESTful CRUD",
			Pattern:     regexp.MustCompile(`/api/v?\d*/[a-zA-Z]+/\d+/(edit|delete|update)/?$`),
			Type:        "REST",
			Method:      "POST",
			Description: "RESTful CRUD操作",
			Examples:    []string{"/api/users/123/edit", "/api/posts/456/delete"},
		},
		
		// GraphQL模式
		{
			Name:        "GraphQL Endpoint",
			Pattern:     regexp.MustCompile(`/graphql/?$`),
			Type:        "GraphQL",
			Method:      "POST",
			Description: "GraphQL查询端点",
			Examples:    []string{"/graphql", "/api/graphql"},
		},
		
		// WebSocket模式
		{
			Name:        "WebSocket Endpoint",
			Pattern:     regexp.MustCompile(`ws://|wss://`),
			Type:        "WebSocket",
			Method:      "WEBSOCKET",
			Description: "WebSocket连接端点",
			Examples:    []string{"ws://example.com/ws", "wss://api.example.com/socket"},
		},
		
		// 微服务API模式
		{
			Name:        "Microservice API",
			Pattern:     regexp.MustCompile(`/[a-zA-Z]+-service/`),
			Type:        "REST",
			Method:      "GET",
			Description: "微服务API模式",
			Examples:    []string{"/user-service/api", "/payment-service/v1"},
		},
		
		// 移动API模式
		{
			Name:        "Mobile API",
			Pattern:     regexp.MustCompile(`/(mobile|app)/api/`),
			Type:        "REST",
			Method:      "GET",
			Description: "移动应用API",
			Examples:    []string{"/mobile/api/login", "/app/api/v1/profile"},
		},
		
		// 管理后台API模式
		{
			Name:        "Admin API",
			Pattern:     regexp.MustCompile(`/(admin|manage|dashboard)/api/`),
			Type:        "REST",
			Method:      "GET",
			Description: "管理后台API",
			Examples:    []string{"/admin/api/users", "/dashboard/api/stats"},
		},
		
		// 第三方集成API模式
		{
			Name:        "Webhook",
			Pattern:     regexp.MustCompile(`/(webhook|callback|notify)/`),
			Type:        "REST",
			Method:      "POST",
			Description: "Webhook回调API",
			Examples:    []string{"/webhook/payment", "/callback/oauth"},
		},
		
		// 文件上传API模式
		{
			Name:        "File Upload",
			Pattern:     regexp.MustCompile(`/(upload|file|media)/`),
			Type:        "REST",
			Method:      "POST",
			Description: "文件上传API",
			Examples:    []string{"/api/upload", "/media/upload"},
		},
		
		// 搜索API模式
		{
			Name:        "Search API",
			Pattern:     regexp.MustCompile(`/(search|query|find)/`),
			Type:        "REST",
			Method:      "GET",
			Description: "搜索查询API",
			Examples:    []string{"/api/search", "/query/products"},
		},
		
		// 认证API模式
		{
			Name:        "Authentication",
			Pattern:     regexp.MustCompile(`/(auth|login|logout|register|token)/`),
			Type:        "REST",
			Method:      "POST",
			Description: "认证相关API",
			Examples:    []string{"/auth/login", "/api/token/refresh"},
		},
		
		// 统计API模式
		{
			Name:        "Analytics API",
			Pattern:     regexp.MustCompile(`/(stats|analytics|metrics|report)/`),
			Type:        "REST",
			Method:      "GET",
			Description: "统计分析API",
			Examples:    []string{"/api/stats", "/analytics/report"},
		},
	}
}

// AnalyzeURLPatterns 分析URL模式
func (pa *PatternAnalyzer) AnalyzeURLPatterns(urls []string) []PatternMatch {
	var matches []PatternMatch
	
	for _, url := range urls {
		for _, pattern := range pa.patterns {
			if pattern.Pattern.MatchString(url) {
				matches = append(matches, PatternMatch{
					URL:         url,
					Pattern:     pattern,
					Confidence:  pa.calculateConfidence(url, pattern),
					MatchedPart: pattern.Pattern.FindString(url),
				})
			}
		}
	}
	
	return matches
}

// PatternMatch 模式匹配结果
type PatternMatch struct {
	URL         string      `json:"url"`
	Pattern     *APIPattern `json:"pattern"`
	Confidence  float64     `json:"confidence"`
	MatchedPart string      `json:"matched_part"`
}

// calculateConfidence 计算匹配置信度
func (pa *PatternAnalyzer) calculateConfidence(url string, pattern *APIPattern) float64 {
	confidence := 0.5 // 基础置信度
	
	// 根据URL特征调整置信度
	urlLower := strings.ToLower(url)
	
	// API关键词加分
	apiKeywords := []string{"api", "service", "endpoint", "rest", "graphql"}
	for _, keyword := range apiKeywords {
		if strings.Contains(urlLower, keyword) {
			confidence += 0.1
		}
	}
	
	// 版本号加分
	if regexp.MustCompile(`v\d+`).MatchString(urlLower) {
		confidence += 0.1
	}
	
	// 资源名称加分
	if regexp.MustCompile(`/[a-zA-Z]+s/`).MatchString(url) { // 复数形式
		confidence += 0.1
	}
	
	// HTTP方法匹配加分
	if pattern.Method != "" {
		confidence += 0.1
	}
	
	// 限制在0-1范围内
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return confidence
}

// DetectAPIVersioning 检测API版本控制模式
func (pa *PatternAnalyzer) DetectAPIVersioning(urls []string) map[string][]string {
	versions := make(map[string][]string)
	
	// URL路径版本控制
	pathVersionPattern := regexp.MustCompile(`/(v\d+(\.\d+)?)/`)
	for _, url := range urls {
		matches := pathVersionPattern.FindAllStringSubmatch(url, -1)
		for _, match := range matches {
			if len(match) > 1 {
				versions["path"] = append(versions["path"], match[1])
			}
		}
	}
	
	// 子域名版本控制
	subdomainPattern := regexp.MustCompile(`(v\d+)\.`)
	for _, url := range urls {
		matches := subdomainPattern.FindAllStringSubmatch(url, -1)
		for _, match := range matches {
			if len(match) > 1 {
				versions["subdomain"] = append(versions["subdomain"], match[1])
			}
		}
	}
	
	return versions
}

// AnalyzeAPIStructure 分析API结构
func (pa *PatternAnalyzer) AnalyzeAPIStructure(apis []storage.APIEndpoint) APIStructureAnalysis {
	analysis := APIStructureAnalysis{
		TotalAPIs:    len(apis),
		ByMethod:     make(map[string]int),
		ByType:       make(map[string]int),
		ByDomain:     make(map[string]int),
		PathPatterns: make(map[string]int),
	}
	
	for _, api := range apis {
		// 按方法统计
		analysis.ByMethod[api.Method]++
		
		// 按类型统计
		analysis.ByType[api.Type]++
		
		// 按域名统计
		analysis.ByDomain[api.Domain]++
		
		// 分析路径模式
		pathPattern := pa.extractPathPattern(api.Path)
		analysis.PathPatterns[pathPattern]++
	}
	
	return analysis
}

// APIStructureAnalysis API结构分析结果
type APIStructureAnalysis struct {
	TotalAPIs    int            `json:"total_apis"`
	ByMethod     map[string]int `json:"by_method"`
	ByType       map[string]int `json:"by_type"`
	ByDomain     map[string]int `json:"by_domain"`
	PathPatterns map[string]int `json:"path_patterns"`
}

// extractPathPattern 提取路径模式
func (pa *PatternAnalyzer) extractPathPattern(path string) string {
	// 将数字ID替换为占位符
	idPattern := regexp.MustCompile(`/\d+`)
	pattern := idPattern.ReplaceAllString(path, "/{id}")
	
	// 将UUID替换为占位符
	uuidPattern := regexp.MustCompile(`/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	pattern = uuidPattern.ReplaceAllString(pattern, "/{uuid}")
	
	return pattern
}

// DetectRESTfulPatterns 检测RESTful模式
func (pa *PatternAnalyzer) DetectRESTfulPatterns(apis []storage.APIEndpoint) []RESTfulResource {
	resources := make(map[string]*RESTfulResource)
	
	for _, api := range apis {
		resourceName := pa.extractResourceName(api.Path)
		if resourceName == "" {
			continue
		}
		
		if _, exists := resources[resourceName]; !exists {
			resources[resourceName] = &RESTfulResource{
				Name:    resourceName,
				Methods: make(map[string]bool),
				Paths:   make(map[string]bool),
			}
		}
		
		resource := resources[resourceName]
		resource.Methods[api.Method] = true
		resource.Paths[api.Path] = true
		resource.Count++
	}
	
	// 转换为切片
	var result []RESTfulResource
	for _, resource := range resources {
		// 检查是否符合RESTful模式
		resource.IsRESTful = pa.isRESTfulResource(resource)
		result = append(result, *resource)
	}
	
	return result
}

// RESTfulResource RESTful资源
type RESTfulResource struct {
	Name      string          `json:"name"`
	Methods   map[string]bool `json:"methods"`
	Paths     map[string]bool `json:"paths"`
	Count     int             `json:"count"`
	IsRESTful bool            `json:"is_restful"`
}

// extractResourceName 提取资源名称
func (pa *PatternAnalyzer) extractResourceName(path string) string {
	// 简化的资源名称提取
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part != "" && part != "api" && !regexp.MustCompile(`v\d+`).MatchString(part) {
			// 移除数字ID
			if !regexp.MustCompile(`^\d+$`).MatchString(part) {
				return part
			}
		}
	}
	return ""
}

// isRESTfulResource 检查是否为RESTful资源
func (pa *PatternAnalyzer) isRESTfulResource(resource *RESTfulResource) bool {
	// 检查是否有基本的CRUD操作
	hasGet := resource.Methods["GET"]
	hasPost := resource.Methods["POST"]
	
	// 至少要有GET或POST
	if !hasGet && !hasPost {
		return false
	}
	
	// 检查路径模式
	hasCollection := false
	hasItem := false
	
	for path := range resource.Paths {
		if regexp.MustCompile(`/` + resource.Name + `/?$`).MatchString(path) {
			hasCollection = true
		}
		if regexp.MustCompile(`/` + resource.Name + `/\d+/?`).MatchString(path) {
			hasItem = true
		}
	}
	
	return hasCollection || hasItem
}

// GenerateAPIDocumentation 生成API文档
func (pa *PatternAnalyzer) GenerateAPIDocumentation(apis []storage.APIEndpoint) string {
	doc := "# API Documentation\n\n"
	
	// 按域名分组
	byDomain := make(map[string][]storage.APIEndpoint)
	for _, api := range apis {
		byDomain[api.Domain] = append(byDomain[api.Domain], api)
	}
	
	for domain, domainAPIs := range byDomain {
		doc += fmt.Sprintf("## %s\n\n", domain)
		
		// 按类型分组
		byType := make(map[string][]storage.APIEndpoint)
		for _, api := range domainAPIs {
			byType[api.Type] = append(byType[api.Type], api)
		}
		
		for apiType, typeAPIs := range byType {
			doc += fmt.Sprintf("### %s APIs\n\n", apiType)
			
			for _, api := range typeAPIs {
				doc += fmt.Sprintf("- **%s** `%s`\n", api.Method, api.Path)
				if api.Status > 0 {
					doc += fmt.Sprintf("  - Status: %d\n", api.Status)
				}
				if api.ContentType != "" {
					doc += fmt.Sprintf("  - Content-Type: %s\n", api.ContentType)
				}
				doc += "\n"
			}
		}
	}
	
	return doc
}