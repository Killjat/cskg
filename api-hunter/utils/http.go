package utils

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPClient HTTP客户端工具
type HTTPClient struct {
	client     *http.Client
	userAgents []string
	headers    map[string]string
}

// NewHTTPClient 创建HTTP客户端
func NewHTTPClient(timeout time.Duration, userAgents []string) *HTTPClient {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 跳过SSL验证
			},
		},
	}

	return &HTTPClient{
		client:     client,
		userAgents: userAgents,
		headers:    make(map[string]string),
	}
}

// SetProxy 设置代理
func (hc *HTTPClient) SetProxy(proxyURL string) error {
	if proxyURL == "" {
		return nil
	}

	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("解析代理URL失败: %v", err)
	}

	transport := hc.client.Transport.(*http.Transport)
	transport.Proxy = http.ProxyURL(proxy)

	return nil
}

// SetHeaders 设置默认headers
func (hc *HTTPClient) SetHeaders(headers map[string]string) {
	hc.headers = headers
}

// Get 发送GET请求
func (hc *HTTPClient) Get(url string) (*http.Response, error) {
	return hc.Request("GET", url, nil)
}

// Post 发送POST请求
func (hc *HTTPClient) Post(url string, body io.Reader) (*http.Response, error) {
	return hc.Request("POST", url, body)
}

// Request 发送HTTP请求
func (hc *HTTPClient) Request(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// 设置User-Agent
	if len(hc.userAgents) > 0 {
		req.Header.Set("User-Agent", hc.getRandomUserAgent())
	}

	// 设置默认headers
	for key, value := range hc.headers {
		req.Header.Set(key, value)
	}

	// 设置常用headers
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")

	return hc.client.Do(req)
}

// getRandomUserAgent 获取随机User-Agent
func (hc *HTTPClient) getRandomUserAgent() string {
	if len(hc.userAgents) == 0 {
		return "Mozilla/5.0 (compatible; API-Hunter/1.0)"
	}
	// 简化版本，返回第一个
	return hc.userAgents[0]
}

// CheckURL 检查URL可访问性
func (hc *HTTPClient) CheckURL(url string) (*URLCheckResult, error) {
	start := time.Now()
	resp, err := hc.Get(url)
	duration := time.Since(start)

	result := &URLCheckResult{
		URL:          url,
		ResponseTime: duration,
		Timestamp:    time.Now(),
	}

	if err != nil {
		result.Error = err.Error()
		result.Accessible = false
		return result, nil
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.ContentType = resp.Header.Get("Content-Type")
	result.ContentLength = resp.ContentLength
	result.Accessible = resp.StatusCode < 400

	// 读取部分内容
	if resp.ContentLength > 0 && resp.ContentLength < 1024*1024 { // 小于1MB
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			result.Content = string(body)
		}
	}

	return result, nil
}

// URLCheckResult URL检查结果
type URLCheckResult struct {
	URL           string        `json:"url"`
	Accessible    bool          `json:"accessible"`
	StatusCode    int           `json:"status_code"`
	ContentType   string        `json:"content_type"`
	ContentLength int64         `json:"content_length"`
	ResponseTime  time.Duration `json:"response_time"`
	Content       string        `json:"content"`
	Error         string        `json:"error"`
	Timestamp     time.Time     `json:"timestamp"`
}

// DetectTechnology 检测网站技术栈
func (hc *HTTPClient) DetectTechnology(url string) (*TechStack, error) {
	resp, err := hc.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content := string(body)
	headers := resp.Header

	tech := &TechStack{
		URL: url,
	}

	// 检测服务器
	if server := headers.Get("Server"); server != "" {
		tech.Server = server
	}

	// 检测框架
	tech.Framework = hc.detectFramework(content, headers)

	// 检测JavaScript库
	tech.JSLibraries = hc.detectJSLibraries(content)

	// 检测CMS
	tech.CMS = hc.detectCMS(content, headers)

	return tech, nil
}

// TechStack 技术栈信息
type TechStack struct {
	URL         string   `json:"url"`
	Server      string   `json:"server"`
	Framework   string   `json:"framework"`
	JSLibraries []string `json:"js_libraries"`
	CMS         string   `json:"cms"`
}

// detectFramework 检测框架
func (hc *HTTPClient) detectFramework(content string, headers http.Header) string {
	contentLower := strings.ToLower(content)

	// 检测常见框架
	frameworks := map[string][]string{
		"React":    {"react", "_react", "react-dom"},
		"Vue.js":   {"vue.js", "vue.min.js", "__vue__"},
		"Angular":  {"angular", "ng-app", "ng-controller"},
		"jQuery":   {"jquery", "jquery.min.js"},
		"Bootstrap": {"bootstrap", "bootstrap.min.css"},
		"Laravel":  {"laravel_session", "laravel_token"},
		"Django":   {"csrfmiddlewaretoken", "django"},
		"Rails":    {"rails", "authenticity_token"},
		"Express":  {"express", "x-powered-by: express"},
	}

	for framework, signatures := range frameworks {
		for _, signature := range signatures {
			if strings.Contains(contentLower, signature) {
				return framework
			}
		}
	}

	// 检查headers
	if xPoweredBy := headers.Get("X-Powered-By"); xPoweredBy != "" {
		return xPoweredBy
	}

	return ""
}

// detectJSLibraries 检测JavaScript库
func (hc *HTTPClient) detectJSLibraries(content string) []string {
	var libraries []string
	contentLower := strings.ToLower(content)

	jsLibs := map[string][]string{
		"jQuery":     {"jquery", "jquery.min.js"},
		"React":      {"react.js", "react.min.js"},
		"Vue.js":     {"vue.js", "vue.min.js"},
		"Angular":    {"angular.js", "angular.min.js"},
		"Lodash":     {"lodash", "underscore"},
		"Moment.js":  {"moment.js", "moment.min.js"},
		"Chart.js":   {"chart.js", "chart.min.js"},
		"D3.js":      {"d3.js", "d3.min.js"},
		"Axios":      {"axios", "axios.min.js"},
	}

	for lib, signatures := range jsLibs {
		for _, signature := range signatures {
			if strings.Contains(contentLower, signature) {
				libraries = append(libraries, lib)
				break
			}
		}
	}

	return libraries
}

// detectCMS 检测CMS
func (hc *HTTPClient) detectCMS(content string, headers http.Header) string {
	contentLower := strings.ToLower(content)

	cmsSignatures := map[string][]string{
		"WordPress": {"wp-content", "wp-includes", "wordpress"},
		"Drupal":    {"drupal", "sites/default", "misc/drupal.js"},
		"Joomla":    {"joomla", "option=com_", "joomla!"},
		"Magento":   {"magento", "mage/cookies.js", "var/magento"},
		"Shopify":   {"shopify", "cdn.shopify.com", "shopify-analytics"},
	}

	for cms, signatures := range cmsSignatures {
		for _, signature := range signatures {
			if strings.Contains(contentLower, signature) {
				return cms
			}
		}
	}

	return ""
}

// TestAPIEndpoint 测试API端点
func (hc *HTTPClient) TestAPIEndpoint(url, method string) (*APITestResult, error) {
	start := time.Now()
	
	var resp *http.Response
	var err error

	switch strings.ToUpper(method) {
	case "GET":
		resp, err = hc.Get(url)
	case "POST":
		resp, err = hc.Post(url, nil)
	default:
		resp, err = hc.Request(method, url, nil)
	}

	duration := time.Since(start)

	result := &APITestResult{
		URL:          url,
		Method:       method,
		ResponseTime: duration,
		Timestamp:    time.Now(),
	}

	if err != nil {
		result.Error = err.Error()
		return result, nil
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.ContentType = resp.Header.Get("Content-Type")
	result.Headers = make(map[string]string)

	// 复制响应headers
	for key, values := range resp.Header {
		if len(values) > 0 {
			result.Headers[key] = values[0]
		}
	}

	// 读取响应体
	if resp.ContentLength > 0 && resp.ContentLength < 1024*10 { // 小于10KB
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			result.Response = string(body)
		}
	}

	return result, nil
}

// APITestResult API测试结果
type APITestResult struct {
	URL          string            `json:"url"`
	Method       string            `json:"method"`
	StatusCode   int               `json:"status_code"`
	ContentType  string            `json:"content_type"`
	Headers      map[string]string `json:"headers"`
	Response     string            `json:"response"`
	ResponseTime time.Duration     `json:"response_time"`
	Error        string            `json:"error"`
	Timestamp    time.Time         `json:"timestamp"`
}

// RandomChoice 从切片中随机选择一个元素
func RandomChoice(choices []string) string {
	if len(choices) == 0 {
		return ""
	}
	// 简化版本，返回第一个
	// 在实际应用中可以使用 math/rand 来随机选择
	return choices[0]
}