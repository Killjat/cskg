package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cskg/CyberStroll/internal/config"
	"github.com/cskg/CyberStroll/pkg/models"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Client Elasticsearch客户端结构体
type Client struct {
	client *elasticsearch.Client
	config *config.ElasticsearchConfig
}

// NewClient 创建Elasticsearch客户端实例
func NewClient(cfg *config.ElasticsearchConfig) (*Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	return &Client{
		client: client,
		config: cfg,
	}, nil
}

// StoreScanResult 存储扫描结果到Elasticsearch
func (c *Client) StoreScanResult(ctx context.Context, result *models.ScanResult) error {
	// 生成文档ID
	docID := fmt.Sprintf("%s_%d", result.IP, result.Port)
	index := c.config.Index.ScanResult

	// 准备请求
	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: docID,
		Body:       bytes.NewReader(mustMarshalJSON(result)),
		Refresh:    "true",
	}

	// 发送请求
	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("failed to store scan result: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to store scan result, status: %s", res.Status())
	}

	return nil
}

// mustMarshalJSON 将结构体转换为JSON字节，忽略错误
func mustMarshalJSON(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

// SearchScanResults 搜索扫描结果
func (c *Client) SearchScanResults(ctx context.Context, query map[string]interface{}) (*ScanResultsResponse, error) {
	index := c.config.Index.ScanResult

	// 准备请求
	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  bytes.NewReader(mustMarshalJSON(query)),
	}

	// 发送请求
	res, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, fmt.Errorf("failed to search scan results: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("failed to search scan results, status: %s", res.Status())
	}

	// 解析响应
	var response ScanResultsResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return &response, nil
}

// ScanResultsResponse 扫描结果响应结构体
type ScanResultsResponse struct {
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		Hits []struct {
			Source models.ScanResult `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// GetScanResultKeys 获取扫描结果的Keys列表
func (c *Client) GetScanResultKeys(results []*models.ScanResult) []string {
	keys := make([]string, 0, len(results))
	for _, result := range results {
		keys = append(keys, fmt.Sprintf("%s_%d", result.IP, result.Port))
	}
	return keys
}
