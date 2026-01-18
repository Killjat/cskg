package fingerprint

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/cskg/CyberStroll/internal/search"
)

// Service 指纹分析服务结构体
type Service struct {
	searchService *search.Service
}

// NewService 创建指纹分析服务实例
func NewService(searchService *search.Service) *Service {
	return &Service{
		searchService: searchService,
	}
}

// ExtractFieldFromBanner 根据关键词从banner中提取特定字段
func ExtractFieldFromBanner(banner string, field string) []string {
	var fields []string

	// 如果字段为空，默认提取服务器信息
	if field == "" {
		return ExtractServerFromBanner(banner)
	}

	// 将字段转换为小写，方便匹配
	field = strings.ToLower(field)

	// 定义提取特定字段的正则表达式
	var patterns []*regexp.Regexp

	// 根据字段选择对应的正则表达式
	switch field {
	case "server":
		// HTTP Server: Server: Apache/2.4.41 (Ubuntu)
		patterns = append(patterns, regexp.MustCompile(`(?i)Server:\s*([^\r\n]+)`))
		// SSH Server: SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5
		patterns = append(patterns, regexp.MustCompile(`(?i)SSH-\d+\.\d+-([^\r\n]+)`))
		// FTP Server: 220 (vsFTPd 3.0.3)
		patterns = append(patterns, regexp.MustCompile(`(?i)220.*\(([^\)]+)\)`))
		// SMTP Server: 220 mail.example.com ESMTP Postfix
		patterns = append(patterns, regexp.MustCompile(`(?i)220.*ESMTP\s+([^\r\n]+)`))
		// MQTT Server: 0x10 0x19 0x00 0x04 MQTT 0x04 0x02 0x00 0x3C 0x00 0x00
		patterns = append(patterns, regexp.MustCompile(`(?i)MQTT\s*([^\s]+)`))
	case "http":
		// HTTP响应头中的Server字段
		patterns = append(patterns, regexp.MustCompile(`(?i)Server:\s*([^\r\n]+)`))
		// HTTP响应头中的X-Powered-By字段
		patterns = append(patterns, regexp.MustCompile(`(?i)X-Powered-By:\s*([^\r\n]+)`))
		// HTTP响应头中的Content-Type字段
		patterns = append(patterns, regexp.MustCompile(`(?i)Content-Type:\s*([^\r\n]+)`))
	case "ssh":
		// SSH版本信息
		patterns = append(patterns, regexp.MustCompile(`(?i)SSH-\d+\.\d+-([^\r\n]+)`))
	case "ftp":
		// FTP服务器信息
		patterns = append(patterns, regexp.MustCompile(`(?i)220.*\(([^\)]+)\)`))
	case "smtp":
		// SMTP服务器信息
		patterns = append(patterns, regexp.MustCompile(`(?i)220.*ESMTP\s+([^\r\n]+)`))
	case "mqtt":
		// MQTT服务器信息
		patterns = append(patterns, regexp.MustCompile(`(?i)MQTT\s*([^\s]+)`))
	default:
		// 通用模式: 提取指定字段后的内容
		patterns = append(patterns, regexp.MustCompile(fmt.Sprintf(`(?i)%s:\s*([^\r\n]+)`, regexp.QuoteMeta(field))))
		// 通用模式: 提取类似 "Name/Version" 的字符串
		patterns = append(patterns, regexp.MustCompile(`(?i)([a-zA-Z0-9\-_]+/[a-zA-Z0-9\.\-_]+)`))
	}

	// 应用所有正则表达式
	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(banner, -1)
		for _, match := range matches {
			if len(match) > 1 {
				fieldValue := strings.TrimSpace(match[1])
				if fieldValue != "" {
					fields = append(fields, fieldValue)
				}
			}
		}
	}

	// 去重
	return removeDuplicates(fields)
}

// ExtractServerFromBanner 从banner中提取服务器信息
func ExtractServerFromBanner(banner string) []string {
	var servers []string

	// 定义提取服务器信息的正则表达式
	patterns := []*regexp.Regexp{
		// HTTP Server: Server: Apache/2.4.41 (Ubuntu)
		regexp.MustCompile(`(?i)Server:\s*([^\r\n]+)`),
		// SSH Server: SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5
		regexp.MustCompile(`(?i)SSH-\d+\.\d+-([^\r\n]+)`),
		// FTP Server: 220 (vsFTPd 3.0.3)
		regexp.MustCompile(`(?i)220.*\(([^\)]+)\)`),
		// SMTP Server: 220 mail.example.com ESMTP Postfix
		regexp.MustCompile(`(?i)220.*ESMTP\s+([^\r\n]+)`),
		// MQTT Server: 0x10 0x19 0x00 0x04 MQTT 0x04 0x02 0x00 0x3C 0x00 0x00
		regexp.MustCompile(`(?i)MQTT\s*([^\s]+)`),
		// 通用模式: 提取类似 "Name/Version" 的字符串
		regexp.MustCompile(`(?i)([a-zA-Z0-9\-_]+/[a-zA-Z0-9\.\-_]+)`),
	}

	// 应用所有正则表达式
	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(banner, -1)
		for _, match := range matches {
			if len(match) > 1 {
				server := strings.TrimSpace(match[1])
				if server != "" {
					servers = append(servers, server)
				}
			}
		}
	}

	// 去重
	return removeDuplicates(servers)
}

// removeDuplicates 去除字符串切片中的重复元素
func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// AnalyzeBanners 分析所有banner数据，提取特定字段信息
func (s *Service) AnalyzeBanners(ctx context.Context, field string) (map[string][]string, error) {
	// 搜索所有包含banner的扫描结果
	results, err := s.searchService.SearchScanResults(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to search scan results: %w", err)
	}

	// 分析每个banner，提取特定字段信息
	fieldStats := make(map[string][]string)
	for _, hit := range results.Hits.Hits {
		result := hit.Source
		if result.Banner != "" {
			fields := ExtractFieldFromBanner(result.Banner, field)
			if len(fields) > 0 {
				// 按IP和端口分组
				key := fmt.Sprintf("%s:%d", result.IP, result.Port)
				fieldStats[key] = fields
			}
		}
	}

	return fieldStats, nil
}

// AnalyzeBannersByQuery 分析符合查询条件的banner数据，提取特定字段信息
func (s *Service) AnalyzeBannersByQuery(ctx context.Context, query string, field string) (map[string][]string, error) {
	// 搜索符合条件的扫描结果
	results, err := s.searchService.SearchScanResults(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search scan results: %w", err)
	}

	// 分析每个banner，提取特定字段信息
	fieldStats := make(map[string][]string)
	for _, hit := range results.Hits.Hits {
		result := hit.Source
		if result.Banner != "" {
			fields := ExtractFieldFromBanner(result.Banner, field)
			if len(fields) > 0 {
				// 按IP和端口分组
				key := fmt.Sprintf("%s:%d", result.IP, result.Port)
				fieldStats[key] = fields
			}
		}
	}

	return fieldStats, nil
}

// FormatAnalysisResults 格式化分析结果
func (s *Service) FormatAnalysisResults(analysisResults map[string][]string) string {
	var output string

	// 添加结果统计
	output += fmt.Sprintf("共分析 %d 个banner，提取到服务器信息\n", len(analysisResults))
	output += fmt.Sprintf("========================================\n\n")

	// 格式化每个结果
	for addr, servers := range analysisResults {
		output += fmt.Sprintf("%s\n", addr)
		for _, server := range servers {
			output += fmt.Sprintf("  - %s\n", server)
		}
		output += fmt.Sprintf("\n")
	}

	return output
}
