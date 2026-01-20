package search

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/cskg/CyberStroll/internal/storage"
)

// SearchEngine 搜索引擎
type SearchEngine struct {
	esClient *storage.ElasticsearchClient
	logger   *log.Logger
}

// SearchRequest 搜索请求
type SearchRequest struct {
	IP       string `json:"ip" form:"ip"`
	Port     string `json:"port" form:"port"`
	Banner   string `json:"banner" form:"banner"`
	Service  string `json:"service" form:"service"`
	Country  string `json:"country" form:"country"`
	Page     int    `json:"page" form:"page"`
	Size     int    `json:"size" form:"size"`
	SortBy   string `json:"sort_by" form:"sort_by"`
	SortDesc bool   `json:"sort_desc" form:"sort_desc"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Total   int64                    `json:"total"`
	Page    int                      `json:"page"`
	Size    int                      `json:"size"`
	Results []storage.ScanDocument   `json:"results"`
	Stats   *SearchStats             `json:"stats"`
}

// SearchStats 搜索统计
type SearchStats struct {
	TotalHosts    int64            `json:"total_hosts"`
	TotalPorts    int64            `json:"total_ports"`
	TopServices   []ServiceStat    `json:"top_services"`
	TopCountries  []CountryStat    `json:"top_countries"`
	PortDistrib   []PortStat       `json:"port_distribution"`
}

// ServiceStat 服务统计
type ServiceStat struct {
	Service string `json:"service"`
	Count   int64  `json:"count"`
}

// CountryStat 国家统计
type CountryStat struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}

// PortStat 端口统计
type PortStat struct {
	Port  int   `json:"port"`
	Count int64 `json:"count"`
}

// NewSearchEngine 创建搜索引擎
func NewSearchEngine(esClient *storage.ElasticsearchClient, logger *log.Logger) *SearchEngine {
	return &SearchEngine{
		esClient: esClient,
		logger:   logger,
	}
}

// Search 执行搜索
func (se *SearchEngine) Search(req *SearchRequest) (*SearchResponse, error) {
	// 验证和设置默认值
	if err := se.validateRequest(req); err != nil {
		return nil, fmt.Errorf("请求验证失败: %v", err)
	}

	// 构建Elasticsearch查询
	query := se.buildQuery(req)

	// 执行搜索
	docs, err := se.esClient.SearchDocuments(query)
	if err != nil {
		return nil, fmt.Errorf("搜索执行失败: %v", err)
	}

	// 获取总数
	total := int64(len(docs))

	// 分页处理
	start := (req.Page - 1) * req.Size
	end := start + req.Size
	if start > len(docs) {
		start = len(docs)
	}
	if end > len(docs) {
		end = len(docs)
	}

	pagedDocs := docs[start:end]

	// 获取统计信息
	stats, err := se.getSearchStats(req)
	if err != nil {
		se.logger.Printf("获取统计信息失败: %v", err)
		stats = &SearchStats{}
	}

	return &SearchResponse{
		Total:   total,
		Page:    req.Page,
		Size:    req.Size,
		Results: pagedDocs,
		Stats:   stats,
	}, nil
}

// validateRequest 验证请求
func (se *SearchEngine) validateRequest(req *SearchRequest) error {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	if req.Size > 1000 {
		req.Size = 1000
	}
	if req.SortBy == "" {
		req.SortBy = "scan_time"
	}

	// 验证IP格式
	if req.IP != "" {
		if strings.Contains(req.IP, "/") {
			// CIDR格式
			_, _, err := net.ParseCIDR(req.IP)
			if err != nil {
				return fmt.Errorf("无效的CIDR格式: %s", req.IP)
			}
		} else if strings.Contains(req.IP, "-") {
			// IP范围格式
			parts := strings.Split(req.IP, "-")
			if len(parts) != 2 {
				return fmt.Errorf("无效的IP范围格式: %s", req.IP)
			}
			if net.ParseIP(strings.TrimSpace(parts[0])) == nil ||
				net.ParseIP(strings.TrimSpace(parts[1])) == nil {
				return fmt.Errorf("无效的IP范围: %s", req.IP)
			}
		} else {
			// 单个IP
			if net.ParseIP(req.IP) == nil {
				return fmt.Errorf("无效的IP地址: %s", req.IP)
			}
		}
	}

	// 验证端口
	if req.Port != "" {
		if strings.Contains(req.Port, "-") {
			// 端口范围
			parts := strings.Split(req.Port, "-")
			if len(parts) != 2 {
				return fmt.Errorf("无效的端口范围格式: %s", req.Port)
			}
			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 != nil || err2 != nil || start < 1 || end > 65535 || start > end {
				return fmt.Errorf("无效的端口范围: %s", req.Port)
			}
		} else {
			// 单个端口
			port, err := strconv.Atoi(req.Port)
			if err != nil || port < 1 || port > 65535 {
				return fmt.Errorf("无效的端口号: %s", req.Port)
			}
		}
	}

	return nil
}

// buildQuery 构建Elasticsearch查询
func (se *SearchEngine) buildQuery(req *SearchRequest) map[string]interface{} {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{},
			},
		},
		"sort": []map[string]interface{}{},
		"from": (req.Page - 1) * req.Size,
		"size": req.Size,
	}

	must := query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})

	// IP条件
	if req.IP != "" {
		if strings.Contains(req.IP, "/") {
			// CIDR查询
			must = append(must, map[string]interface{}{
				"term": map[string]interface{}{
					"ip": req.IP,
				},
			})
		} else if strings.Contains(req.IP, "-") {
			// IP范围查询
			parts := strings.Split(req.IP, "-")
			must = append(must, map[string]interface{}{
				"range": map[string]interface{}{
					"ip": map[string]interface{}{
						"gte": strings.TrimSpace(parts[0]),
						"lte": strings.TrimSpace(parts[1]),
					},
				},
			})
		} else {
			// 精确IP查询
			must = append(must, map[string]interface{}{
				"term": map[string]interface{}{
					"ip": req.IP,
				},
			})
		}
	}

	// 端口条件
	if req.Port != "" {
		if strings.Contains(req.Port, "-") {
			// 端口范围查询
			parts := strings.Split(req.Port, "-")
			start, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
			must = append(must, map[string]interface{}{
				"range": map[string]interface{}{
					"port": map[string]interface{}{
						"gte": start,
						"lte": end,
					},
				},
			})
		} else {
			// 精确端口查询
			port, _ := strconv.Atoi(req.Port)
			must = append(must, map[string]interface{}{
				"term": map[string]interface{}{
					"port": port,
				},
			})
		}
	}

	// Banner条件
	if req.Banner != "" {
		must = append(must, map[string]interface{}{
			"match": map[string]interface{}{
				"banner": req.Banner,
			},
		})
	}

	// 服务条件
	if req.Service != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"service": req.Service,
			},
		})
	}

	// 国家条件
	if req.Country != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"geo_info.country": req.Country,
			},
		})
	}

	// 只查询开放端口
	must = append(must, map[string]interface{}{
		"term": map[string]interface{}{
			"state": "open",
		},
	})

	// 更新查询
	query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = must

	// 排序
	sortOrder := "desc"
	if !req.SortDesc {
		sortOrder = "asc"
	}

	query["sort"] = []map[string]interface{}{
		{
			req.SortBy: map[string]interface{}{
				"order": sortOrder,
			},
		},
	}

	return query
}

// getSearchStats 获取搜索统计
func (se *SearchEngine) getSearchStats(req *SearchRequest) (*SearchStats, error) {
	// 构建聚合查询
	aggQuery := map[string]interface{}{
		"query": se.buildQuery(req)["query"],
		"size":  0,
		"aggs": map[string]interface{}{
			"total_hosts": map[string]interface{}{
				"cardinality": map[string]interface{}{
					"field": "ip",
				},
			},
			"total_ports": map[string]interface{}{
				"cardinality": map[string]interface{}{
					"field": "port",
				},
			},
			"top_services": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "service",
					"size":  10,
				},
			},
			"top_countries": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "geo_info.country",
					"size":  10,
				},
			},
			"port_distribution": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "port",
					"size":  20,
				},
			},
		},
	}

	// 执行聚合查询
	docs, err := se.esClient.SearchDocuments(aggQuery)
	if err != nil {
		return nil, err
	}

	// 这里应该解析聚合结果，但由于当前的SearchDocuments方法不支持聚合
	// 我们返回基础统计信息
	stats := &SearchStats{
		TotalHosts:   int64(len(docs)),
		TotalPorts:   int64(len(docs)),
		TopServices:  []ServiceStat{},
		TopCountries: []CountryStat{},
		PortDistrib:  []PortStat{},
	}

	// 简单统计
	serviceMap := make(map[string]int64)
	countryMap := make(map[string]int64)
	portMap := make(map[int]int64)

	for _, doc := range docs {
		if doc.Service != "" {
			serviceMap[doc.Service]++
		}
		if doc.GeoInfo != nil && doc.GeoInfo.Country != "" {
			countryMap[doc.GeoInfo.Country]++
		}
		if doc.Port > 0 {
			portMap[doc.Port]++
		}
	}

	// 转换为统计结构
	for service, count := range serviceMap {
		stats.TopServices = append(stats.TopServices, ServiceStat{
			Service: service,
			Count:   count,
		})
	}

	for country, count := range countryMap {
		stats.TopCountries = append(stats.TopCountries, CountryStat{
			Country: country,
			Count:   count,
		})
	}

	for port, count := range portMap {
		stats.PortDistrib = append(stats.PortDistrib, PortStat{
			Port:  port,
			Count: count,
		})
	}

	return stats, nil
}

// GetAssetInfo 获取资产信息
func (se *SearchEngine) GetAssetInfo(ip string) ([]storage.ScanDocument, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"ip": ip,
						},
					},
					{
						"term": map[string]interface{}{
							"state": "open",
						},
					},
				},
			},
		},
		"sort": []map[string]interface{}{
			{
				"port": map[string]interface{}{
					"order": "asc",
				},
			},
		},
		"size": 1000,
	}

	return se.esClient.SearchDocuments(query)
}

// GetRecentScans 获取最近扫描
func (se *SearchEngine) GetRecentScans(limit int) ([]storage.ScanDocument, error) {
	if limit <= 0 {
		limit = 100
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"sort": []map[string]interface{}{
			{
				"scan_time": map[string]interface{}{
					"order": "desc",
				},
			},
		},
		"size": limit,
	}

	return se.esClient.SearchDocuments(query)
}