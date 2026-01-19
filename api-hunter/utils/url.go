package utils

import (
	"net/url"
	"regexp"
	"strings"
)

// URLUtils URL工具函数
type URLUtils struct{}

// NewURLUtils 创建URL工具实例
func NewURLUtils() *URLUtils {
	return &URLUtils{}
}

// NormalizeURL 标准化URL
func (u *URLUtils) NormalizeURL(rawURL string) (string, error) {
	// 添加协议前缀
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// 移除默认端口
	if (parsedURL.Scheme == "http" && parsedURL.Port() == "80") ||
		(parsedURL.Scheme == "https" && parsedURL.Port() == "443") {
		parsedURL.Host = parsedURL.Hostname()
	}

	// 移除尾部斜杠
	parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/")
	if parsedURL.Path == "" {
		parsedURL.Path = "/"
	}

	return parsedURL.String(), nil
}

// ExtractDomain 提取域名
func (u *URLUtils) ExtractDomain(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsedURL.Host
}

// ExtractPath 提取路径
func (u *URLUtils) ExtractPath(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsedURL.Path
}

// ExtractQuery 提取查询参数
func (u *URLUtils) ExtractQuery(rawURL string) map[string]string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}

	params := make(map[string]string)
	for key, values := range parsedURL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	return params
}

// IsValidURL 检查URL是否有效
func (u *URLUtils) IsValidURL(rawURL string) bool {
	_, err := url.Parse(rawURL)
	return err == nil
}

// IsSameDomain 检查两个URL是否同域
func (u *URLUtils) IsSameDomain(url1, url2 string) bool {
	domain1 := u.ExtractDomain(url1)
	domain2 := u.ExtractDomain(url2)
	return domain1 == domain2
}

// GenerateURLVariations 生成URL变体
func (u *URLUtils) GenerateURLVariations(baseURL string) []string {
	var variations []string

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return variations
	}

	// 基础URL
	variations = append(variations, baseURL)

	// 添加/移除尾部斜杠
	if strings.HasSuffix(parsedURL.Path, "/") {
		newURL := *parsedURL
		newURL.Path = strings.TrimSuffix(newURL.Path, "/")
		variations = append(variations, newURL.String())
	} else {
		newURL := *parsedURL
		newURL.Path = newURL.Path + "/"
		variations = append(variations, newURL.String())
	}

	// HTTP/HTTPS变体
	if parsedURL.Scheme == "http" {
		newURL := *parsedURL
		newURL.Scheme = "https"
		variations = append(variations, newURL.String())
	} else if parsedURL.Scheme == "https" {
		newURL := *parsedURL
		newURL.Scheme = "http"
		variations = append(variations, newURL.String())
	}

	return variations
}

// ExtractAPIEndpoints 从URL中提取可能的API端点
func (u *URLUtils) ExtractAPIEndpoints(baseURL string) []string {
	var endpoints []string

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return endpoints
	}

	// 常见API路径
	apiPaths := []string{
		"/api",
		"/api/v1",
		"/api/v2",
		"/rest",
		"/service",
		"/graphql",
		"/ws",
		"/websocket",
	}

	for _, path := range apiPaths {
		newURL := *parsedURL
		newURL.Path = path
		endpoints = append(endpoints, newURL.String())
	}

	return endpoints
}

// IsAPIURL 检查URL是否可能是API
func (u *URLUtils) IsAPIURL(rawURL string) bool {
	urlLower := strings.ToLower(rawURL)

	// API关键词
	apiKeywords := []string{
		"/api/",
		"/rest/",
		"/service/",
		"/endpoint/",
		"/graphql",
		"/ws/",
		"/websocket",
		".json",
		".xml",
		".api",
	}

	for _, keyword := range apiKeywords {
		if strings.Contains(urlLower, keyword) {
			return true
		}
	}

	// 版本号模式
	versionPattern := regexp.MustCompile(`/v\d+/`)
	if versionPattern.MatchString(urlLower) {
		return true
	}

	return false
}

// ExtractSubdomains 提取子域名
func (u *URLUtils) ExtractSubdomains(rawURL string) []string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}

	host := parsedURL.Hostname()
	parts := strings.Split(host, ".")

	if len(parts) < 3 {
		return nil // 没有子域名
	}

	// 返回子域名部分
	return parts[:len(parts)-2]
}

// GenerateSubdomainVariations 生成子域名变体
func (u *URLUtils) GenerateSubdomainVariations(baseURL string) []string {
	var variations []string

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return variations
	}

	host := parsedURL.Hostname()
	parts := strings.Split(host, ".")

	if len(parts) < 2 {
		return variations
	}

	// 常见子域名
	subdomains := []string{
		"api",
		"www",
		"app",
		"mobile",
		"m",
		"admin",
		"dev",
		"test",
		"staging",
		"beta",
		"v1",
		"v2",
		"service",
		"rest",
	}

	baseDomain := strings.Join(parts[len(parts)-2:], ".")

	for _, subdomain := range subdomains {
		newURL := *parsedURL
		newURL.Host = subdomain + "." + baseDomain
		variations = append(variations, newURL.String())
	}

	return variations
}

// CleanURL 清理URL
func (u *URLUtils) CleanURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// 移除fragment
	parsedURL.Fragment = ""

	// 移除常见的跟踪参数
	trackingParams := []string{
		"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"fbclid", "gclid", "msclkid", "_ga", "_gid",
	}

	query := parsedURL.Query()
	for _, param := range trackingParams {
		query.Del(param)
	}
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String()
}

// ExtractFileExtension 提取文件扩展名
func (u *URLUtils) ExtractFileExtension(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	path := parsedURL.Path
	if idx := strings.LastIndex(path, "."); idx != -1 {
		return path[idx+1:]
	}

	return ""
}

// IsStaticResource 检查是否为静态资源
func (u *URLUtils) IsStaticResource(rawURL string) bool {
	ext := u.ExtractFileExtension(rawURL)
	staticExts := []string{
		"css", "js", "jpg", "jpeg", "png", "gif", "svg", "ico",
		"pdf", "zip", "rar", "tar", "gz", "mp3", "mp4", "avi",
		"mov", "wmv", "flv", "swf", "woff", "woff2", "ttf", "eot",
	}

	extLower := strings.ToLower(ext)
	for _, staticExt := range staticExts {
		if extLower == staticExt {
			return true
		}
	}

	return false
}

// BuildURL 构建URL
func (u *URLUtils) BuildURL(scheme, host, path string, params map[string]string) string {
	parsedURL := &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}

	if len(params) > 0 {
		query := url.Values{}
		for key, value := range params {
			query.Set(key, value)
		}
		parsedURL.RawQuery = query.Encode()
	}

	return parsedURL.String()
}

// ParseUserAgent 解析User-Agent
func (u *URLUtils) ParseUserAgent(userAgent string) map[string]string {
	info := make(map[string]string)

	// 简化的User-Agent解析
	if strings.Contains(userAgent, "Chrome") {
		info["browser"] = "Chrome"
	} else if strings.Contains(userAgent, "Firefox") {
		info["browser"] = "Firefox"
	} else if strings.Contains(userAgent, "Safari") {
		info["browser"] = "Safari"
	} else if strings.Contains(userAgent, "Edge") {
		info["browser"] = "Edge"
	}

	if strings.Contains(userAgent, "Windows") {
		info["os"] = "Windows"
	} else if strings.Contains(userAgent, "Mac") {
		info["os"] = "macOS"
	} else if strings.Contains(userAgent, "Linux") {
		info["os"] = "Linux"
	} else if strings.Contains(userAgent, "Android") {
		info["os"] = "Android"
	} else if strings.Contains(userAgent, "iOS") {
		info["os"] = "iOS"
	}

	return info
}

// GenerateCommonPaths 生成常见路径
func (u *URLUtils) GenerateCommonPaths() []string {
	return []string{
		"/",
		"/index.html",
		"/index.php",
		"/home",
		"/about",
		"/contact",
		"/login",
		"/register",
		"/admin",
		"/dashboard",
		"/api",
		"/api/v1",
		"/api/v2",
		"/rest",
		"/service",
		"/graphql",
		"/docs",
		"/documentation",
		"/swagger",
		"/openapi",
		"/health",
		"/status",
		"/ping",
		"/version",
		"/info",
	}
}