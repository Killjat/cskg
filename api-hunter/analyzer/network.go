package analyzer

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"api-hunter/storage"
)

// NetworkAnalyzer 网络流量分析器
type NetworkAnalyzer struct {
	db       *storage.Database
	requests []NetworkRequest
	enabled  bool
}

// NetworkRequest 网络请求信息
type NetworkRequest struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	RequestBody string            `json:"request_body"`
	Response    string            `json:"response"`
	StatusCode  int               `json:"status_code"`
	Timestamp   time.Time         `json:"timestamp"`
}

// NewNetworkAnalyzer 创建网络分析器
func NewNetworkAnalyzer(db *storage.Database) *NetworkAnalyzer {
	return &NetworkAnalyzer{
		db:       db,
		requests: make([]NetworkRequest, 0),
		enabled:  false,
	}
}

// StartCapture 开始网络捕获
func (na *NetworkAnalyzer) StartCapture(ctx context.Context) error {
	log.Println("网络捕获功能暂未实现")
	na.enabled = true
	return nil
}

// StopCapture 停止网络捕获
func (na *NetworkAnalyzer) StopCapture() {
	na.enabled = false
	log.Printf("网络捕获已停止，共捕获 %d 个请求", len(na.requests))
}

// AnalyzeTraffic 分析网络流量
func (na *NetworkAnalyzer) AnalyzeTraffic() ([]storage.APIEndpoint, error) {
	var apis []storage.APIEndpoint

	for _, req := range na.requests {
		if na.isAPIRequest(req) {
			api := storage.APIEndpoint{
				URL:         req.URL,
				Method:      req.Method,
				Status:      req.StatusCode,
				Source:      "network",
				Domain:      na.extractDomain(req.URL),
				Path:        na.extractPath(req.URL),
				Type:        na.detectAPIType(req.URL, req.Response),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			// 设置headers和response
			if len(req.Headers) > 0 {
				api.Headers = fmt.Sprintf("%v", req.Headers)
			}
			if req.Response != "" {
				api.Response = req.Response
			}

			apis = append(apis, api)
		}
	}

	return apis, nil
}

// isAPIRequest 判断是否为API请求
func (na *NetworkAnalyzer) isAPIRequest(req NetworkRequest) bool {
	url := strings.ToLower(req.URL)

	// API相关的关键词
	apiKeywords := []string{
		"/api/", "/rest/", "/graphql", "/v1/", "/v2/", "/v3/",
		".json", "/ajax/", "/service/", "/endpoint/",
	}

	for _, keyword := range apiKeywords {
		if strings.Contains(url, keyword) {
			return true
		}
	}

	// 检查Content-Type
	if contentType, exists := req.Headers["Content-Type"]; exists {
		contentTypeLower := strings.ToLower(contentType)
		if strings.Contains(contentTypeLower, "application/json") ||
		   strings.Contains(contentTypeLower, "application/xml") {
			return true
		}
	}

	return false
}

// extractDomain 提取域名
func (na *NetworkAnalyzer) extractDomain(url string) string {
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
func (na *NetworkAnalyzer) extractPath(url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		parts := strings.SplitN(url, "/", 4)
		if len(parts) > 3 {
			return "/" + parts[3]
		}
		return "/"
	}

	return "/"
}

// detectAPIType 检测API类型
func (na *NetworkAnalyzer) detectAPIType(url, response string) string {
	urlLower := strings.ToLower(url)
	responseLower := strings.ToLower(response)

	// GraphQL检测
	if strings.Contains(urlLower, "graphql") ||
	   strings.Contains(responseLower, "\"data\"") && strings.Contains(responseLower, "\"query\"") {
		return "GraphQL"
	}

	// WebSocket检测
	if strings.HasPrefix(urlLower, "ws://") || strings.HasPrefix(urlLower, "wss://") {
		return "WebSocket"
	}

	return "REST"
}

// GetCapturedRequests 获取捕获的请求
func (na *NetworkAnalyzer) GetCapturedRequests() []NetworkRequest {
	return na.requests
}

// AddRequest 添加请求记录（用于测试）
func (na *NetworkAnalyzer) AddRequest(req NetworkRequest) {
	if na.enabled {
		na.requests = append(na.requests, req)
	}
}

// GetStatistics 获取网络分析统计
func (na *NetworkAnalyzer) GetStatistics() map[string]interface{} {
	stats := make(map[string]interface{})
	
	stats["total_requests"] = len(na.requests)
	stats["enabled"] = na.enabled

	// 按方法统计
	methodStats := make(map[string]int)
	for _, req := range na.requests {
		methodStats[req.Method]++
	}
	stats["methods"] = methodStats

	// 按状态码统计
	statusStats := make(map[string]int)
	for _, req := range na.requests {
		statusGroup := fmt.Sprintf("%dxx", req.StatusCode/100)
		statusStats[statusGroup]++
	}
	stats["status_codes"] = statusStats

	return stats
}