package search

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cskg/CyberStroll/internal/elasticsearch"
)

// Service 搜索服务结构体
type Service struct {
	esClient *elasticsearch.Client
}

// NewService 创建搜索服务实例
func NewService(esClient *elasticsearch.Client) *Service {
	return &Service{
		esClient: esClient,
	}
}

// SearchScanResults 搜索扫描结果 - 支持单字符串查询
func (s *Service) SearchScanResults(ctx context.Context, query string) (*elasticsearch.ScanResultsResponse, error) {
	// 构建Elasticsearch查询
	esQuery := buildQuery(query)

	// 执行查询
	results, err := s.esClient.SearchScanResults(ctx, esQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to search scan results: %w", err)
	}

	return results, nil
}

// SearchScanResultsWithFilters 搜索扫描结果 - 支持多字段过滤查询
func (s *Service) SearchScanResultsWithFilters(ctx context.Context, filters map[string]interface{}) (*elasticsearch.ScanResultsResponse, error) {
	// 构建Elasticsearch查询
	esQuery := buildFilterQuery(filters)

	// 执行查询
	results, err := s.esClient.SearchScanResults(ctx, esQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to search scan results with filters: %w", err)
	}

	return results, nil
}

// buildQuery 构建单字符串查询
func buildQuery(query string) map[string]interface{} {
	// 如果查询为空，返回所有结果
	if query == "" {
		return map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
			"size": 100,
		}
	}

	// 构建查询
	queryMap := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"multi_match": map[string]interface{}{
							"query":    query,
							"fields":   []string{"ip", "protocol", "service", "app", "banner"},
							"operator": "or",
						},
					},
				},
			},
		},
		"size": 1000,
	}

	// 只有当query可以转换为数字时，才添加port字段的term查询
	if _, err := strconv.Atoi(query); err == nil {
		// 如果query是数字，添加port字段的term查询
		boolQuery := queryMap["query"].(map[string]interface{})["bool"].(map[string]interface{})
		shouldClauses := boolQuery["should"].([]map[string]interface{})
		shouldClauses = append(shouldClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"port": query,
			},
		})
		boolQuery["should"] = shouldClauses
	}

	return queryMap
}

// buildFilterQuery 构建多字段过滤查询
func buildFilterQuery(filters map[string]interface{}) map[string]interface{} {
	// 构建bool查询
	boolQuery := make(map[string]interface{})
	mustClauses := make([]map[string]interface{}, 0)

	// 添加IP过滤
	if ip, ok := filters["ip"].(string); ok && ip != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"ip": ip,
			},
		})
	}

	// 添加端口过滤
	if port, ok := filters["port"].(int); ok {
		mustClauses = append(mustClauses, map[string]interface{}{
			"term": map[string]interface{}{
				"port": port,
			},
		})
	}

	// 添加协议过滤
	if protocol, ok := filters["protocol"].(string); ok && protocol != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"protocol": protocol,
			},
		})
	}

	// 添加服务过滤
	if service, ok := filters["service"].(string); ok && service != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"service": service,
			},
		})
	}

	// 添加Banner过滤
	if banner, ok := filters["banner"].(string); ok && banner != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"banner": banner,
			},
		})
	}

	// 添加应用过滤
	if app, ok := filters["app"].(string); ok && app != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"app": app,
			},
		})
	}

	// 添加状态过滤
	if status, ok := filters["status"].(string); ok && status != "" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"status": status,
			},
		})
	}

	// 构建查询
	searchQuery := make(map[string]interface{})
	if len(mustClauses) > 0 {
		boolQuery["must"] = mustClauses
		searchQuery["query"] = map[string]interface{}{
			"bool": boolQuery,
		}
	} else {
		// 空查询，返回所有结果
		searchQuery["query"] = map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	}

	// 设置结果大小
	searchQuery["size"] = 1000

	return searchQuery
}

// FormatResults 格式化搜索结果
func (s *Service) FormatResults(results *elasticsearch.ScanResultsResponse) string {
	var output string

	// 添加结果统计
	output += fmt.Sprintf("共找到 %d 个结果\n", results.Hits.Total.Value)
	output += fmt.Sprintf("========================================\n\n")

	// 格式化每个结果
	for i, hit := range results.Hits.Hits {
		result := hit.Source
		output += fmt.Sprintf("结果 %d:\n", i+1)
		output += fmt.Sprintf("  IP: %s\n", result.IP)
		output += fmt.Sprintf("  端口: %d\n", result.Port)
		output += fmt.Sprintf("  协议: %s\n", result.Protocol)
		output += fmt.Sprintf("  服务: %s\n", result.Service)
		output += fmt.Sprintf("  应用: %s\n", result.App)
		output += fmt.Sprintf("  状态: %s\n", result.Status)
		output += fmt.Sprintf("  扫描时间: %s\n", result.CreatedAt.Format("2006-01-02 15:04:05"))
		if result.Banner != "" {
			output += fmt.Sprintf("  Banner: %s\n", result.Banner)
		}
		output += fmt.Sprintf("  任务ID: %s\n", result.TaskID)
		output += fmt.Sprintf("========================================\n\n")
	}

	return output
}

// GetSummary 获取搜索结果摘要
func (s *Service) GetSummary(results *elasticsearch.ScanResultsResponse) string {
	return fmt.Sprintf("共找到 %d 个扫描结果", results.Hits.Total.Value)
}
