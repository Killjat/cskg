package utils

import (
	"regexp"
	"strings"
	"time"
)

// Filter 过滤器
type Filter struct {
	allowedDomains []string
	blockedDomains []string
	blockedPaths   []string
	allowedExts    []string
	blockedExts    []string
	patterns       []*regexp.Regexp
}

// NewFilter 创建过滤器
func NewFilter() *Filter {
	return &Filter{
		allowedDomains: make([]string, 0),
		blockedDomains: make([]string, 0),
		blockedPaths:   make([]string, 0),
		allowedExts:    make([]string, 0),
		blockedExts:    make([]string, 0),
		patterns:       make([]*regexp.Regexp, 0),
	}
}

// SetAllowedDomains 设置允许的域名
func (f *Filter) SetAllowedDomains(domains []string) {
	f.allowedDomains = domains
}

// SetBlockedDomains 设置阻止的域名
func (f *Filter) SetBlockedDomains(domains []string) {
	f.blockedDomains = domains
}

// SetBlockedPaths 设置阻止的路径
func (f *Filter) SetBlockedPaths(paths []string) {
	f.blockedPaths = paths
}

// SetAllowedExtensions 设置允许的文件扩展名
func (f *Filter) SetAllowedExtensions(exts []string) {
	f.allowedExts = exts
}

// SetBlockedExtensions 设置阻止的文件扩展名
func (f *Filter) SetBlockedExtensions(exts []string) {
	f.blockedExts = exts
}

// AddPattern 添加正则表达式模式
func (f *Filter) AddPattern(pattern string) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	f.patterns = append(f.patterns, regex)
	return nil
}

// IsAllowed 检查URL是否被允许
func (f *Filter) IsAllowed(url string) bool {
	// 检查域名白名单
	if len(f.allowedDomains) > 0 {
		allowed := false
		for _, domain := range f.allowedDomains {
			if f.containsDomain(url, domain) {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	// 检查域名黑名单
	for _, domain := range f.blockedDomains {
		if f.containsDomain(url, domain) {
			return false
		}
	}

	// 检查路径黑名单
	for _, path := range f.blockedPaths {
		if strings.Contains(strings.ToLower(url), strings.ToLower(path)) {
			return false
		}
	}

	// 检查文件扩展名
	ext := f.extractExtension(url)
	if ext != "" {
		// 检查允许的扩展名
		if len(f.allowedExts) > 0 {
			allowed := false
			for _, allowedExt := range f.allowedExts {
				if strings.EqualFold(ext, allowedExt) {
					allowed = true
					break
				}
			}
			if !allowed {
				return false
			}
		}

		// 检查阻止的扩展名
		for _, blockedExt := range f.blockedExts {
			if strings.EqualFold(ext, blockedExt) {
				return false
			}
		}
	}

	// 检查正则表达式模式
	for _, pattern := range f.patterns {
		if pattern.MatchString(url) {
			return false
		}
	}

	return true
}

// containsDomain 检查URL是否包含指定域名
func (f *Filter) containsDomain(url, domain string) bool {
	return strings.Contains(strings.ToLower(url), strings.ToLower(domain))
}

// extractExtension 提取文件扩展名
func (f *Filter) extractExtension(url string) string {
	// 移除查询参数和fragment
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}
	if idx := strings.Index(url, "#"); idx != -1 {
		url = url[:idx]
	}

	// 提取扩展名
	if idx := strings.LastIndex(url, "."); idx != -1 {
		return url[idx+1:]
	}

	return ""
}

// URLClassifier URL分类器
type URLClassifier struct {
	apiPatterns    []*regexp.Regexp
	staticPatterns []*regexp.Regexp
	adminPatterns  []*regexp.Regexp
}

// NewURLClassifier 创建URL分类器
func NewURLClassifier() *URLClassifier {
	return &URLClassifier{
		apiPatterns: []*regexp.Regexp{
			regexp.MustCompile(`/api/`),
			regexp.MustCompile(`/v\d+/`),
			regexp.MustCompile(`/rest/`),
			regexp.MustCompile(`/service/`),
			regexp.MustCompile(`/graphql`),
			regexp.MustCompile(`\.json$`),
			regexp.MustCompile(`\.xml$`),
		},
		staticPatterns: []*regexp.Regexp{
			regexp.MustCompile(`\.(css|js|jpg|jpeg|png|gif|svg|ico|pdf|zip)$`),
			regexp.MustCompile(`/static/`),
			regexp.MustCompile(`/assets/`),
			regexp.MustCompile(`/public/`),
		},
		adminPatterns: []*regexp.Regexp{
			regexp.MustCompile(`/admin/`),
			regexp.MustCompile(`/manage/`),
			regexp.MustCompile(`/dashboard/`),
			regexp.MustCompile(`/control/`),
		},
	}
}

// ClassifyURL 分类URL
func (uc *URLClassifier) ClassifyURL(url string) string {
	urlLower := strings.ToLower(url)

	// 检查API模式
	for _, pattern := range uc.apiPatterns {
		if pattern.MatchString(urlLower) {
			return "api"
		}
	}

	// 检查静态资源模式
	for _, pattern := range uc.staticPatterns {
		if pattern.MatchString(urlLower) {
			return "static"
		}
	}

	// 检查管理后台模式
	for _, pattern := range uc.adminPatterns {
		if pattern.MatchString(urlLower) {
			return "admin"
		}
	}

	return "page"
}

// DuplicateFilter 去重过滤器
type DuplicateFilter struct {
	seen map[string]bool
}

// NewDuplicateFilter 创建去重过滤器
func NewDuplicateFilter() *DuplicateFilter {
	return &DuplicateFilter{
		seen: make(map[string]bool),
	}
}

// IsDuplicate 检查是否重复
func (df *DuplicateFilter) IsDuplicate(url string) bool {
	normalized := df.normalizeURL(url)
	if df.seen[normalized] {
		return true
	}
	df.seen[normalized] = true
	return false
}

// normalizeURL 标准化URL用于去重
func (df *DuplicateFilter) normalizeURL(url string) string {
	// 转换为小写
	url = strings.ToLower(url)

	// 移除尾部斜杠
	url = strings.TrimSuffix(url, "/")

	// 移除常见的跟踪参数
	trackingParams := []string{
		"utm_source", "utm_medium", "utm_campaign",
		"fbclid", "gclid", "_ga", "_gid",
	}

	for _, param := range trackingParams {
		pattern := regexp.MustCompile(`[?&]` + param + `=[^&]*`)
		url = pattern.ReplaceAllString(url, "")
	}

	// 清理多余的?和&
	url = regexp.MustCompile(`[?&]$`).ReplaceAllString(url, "")
	url = regexp.MustCompile(`\?&`).ReplaceAllString(url, "?")

	return url
}

// Reset 重置去重过滤器
func (df *DuplicateFilter) Reset() {
	df.seen = make(map[string]bool)
}

// Count 获取已处理的URL数量
func (df *DuplicateFilter) Count() int {
	return len(df.seen)
}

// RateLimiter 速率限制器
type RateLimiter struct {
	requests    map[string][]int64
	maxRequests int
	timeWindow  int64 // 时间窗口（秒）
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(maxRequests int, timeWindowSeconds int64) *RateLimiter {
	return &RateLimiter{
		requests:    make(map[string][]int64),
		maxRequests: maxRequests,
		timeWindow:  timeWindowSeconds,
	}
}

// IsAllowed 检查是否允许请求
func (rl *RateLimiter) IsAllowed(domain string) bool {
	now := getCurrentTimestamp()
	
	// 获取域名的请求历史
	if _, exists := rl.requests[domain]; !exists {
		rl.requests[domain] = make([]int64, 0)
	}

	requests := rl.requests[domain]

	// 清理过期的请求记录
	var validRequests []int64
	for _, timestamp := range requests {
		if now-timestamp < rl.timeWindow {
			validRequests = append(validRequests, timestamp)
		}
	}

	// 检查是否超过限制
	if len(validRequests) >= rl.maxRequests {
		return false
	}

	// 记录新请求
	validRequests = append(validRequests, now)
	rl.requests[domain] = validRequests

	return true
}

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// ContentFilter 内容过滤器
type ContentFilter struct {
	minSize int64
	maxSize int64
}

// NewContentFilter 创建内容过滤器
func NewContentFilter(minSize, maxSize int64) *ContentFilter {
	return &ContentFilter{
		minSize: minSize,
		maxSize: maxSize,
	}
}

// IsValidSize 检查内容大小是否有效
func (cf *ContentFilter) IsValidSize(size int64) bool {
	if cf.minSize > 0 && size < cf.minSize {
		return false
	}
	if cf.maxSize > 0 && size > cf.maxSize {
		return false
	}
	return true
}

// IsValidContentType 检查内容类型是否有效
func (cf *ContentFilter) IsValidContentType(contentType string) bool {
	validTypes := []string{
		"text/html",
		"application/json",
		"application/xml",
		"text/xml",
		"application/javascript",
		"text/javascript",
	}

	contentTypeLower := strings.ToLower(contentType)
	for _, validType := range validTypes {
		if strings.Contains(contentTypeLower, validType) {
			return true
		}
	}

	return false
}