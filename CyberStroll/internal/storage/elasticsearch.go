package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ElasticsearchClient Elasticsearch客户端
type ElasticsearchClient struct {
	client *elasticsearch.Client
	config *ESConfig
}

// ESConfig Elasticsearch配置
type ESConfig struct {
	URLs     []string `yaml:"urls"`
	Index    string   `yaml:"index"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	Timeout  int      `yaml:"timeout"`
}

// ScanDocument 扫描文档结构
type ScanDocument struct {
	IP           string                 `json:"ip"`
	Port         int                    `json:"port"`
	Protocol     string                 `json:"protocol"`
	Service      string                 `json:"service"`
	ServiceVersion string               `json:"service_version"`
	Banner       string                 `json:"banner"`
	State        string                 `json:"state"`
	ScanTime     time.Time              `json:"scan_time"`
	LastUpdate   time.Time              `json:"last_update"`
	TaskID       string                 `json:"task_id"`
	NodeID       string                 `json:"node_id"`
	Applications []ApplicationDoc       `json:"applications"`
	GeoInfo      *GeoInfo               `json:"geo_info,omitempty"`
	OSInfo       *OSInfo                `json:"os_info,omitempty"`
	SecurityInfo *SecurityInfo          `json:"security_info,omitempty"`
	NetworkInfo  *NetworkInfo           `json:"network_info,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ApplicationDoc 应用文档
type ApplicationDoc struct {
	Name       string   `json:"name"`
	Version    string   `json:"version"`
	Category   string   `json:"category"`
	Confidence int      `json:"confidence"`
	CPE        string   `json:"cpe"`
	Tags       []string `json:"tags"`
}

// GeoInfo 地理信息
type GeoInfo struct {
	Country      string  `json:"country"`
	CountryCode  string  `json:"country_code"`
	Region       string  `json:"region"`
	City         string  `json:"city"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	ISP          string  `json:"isp"`
	Organization string  `json:"organization"`
}

// OSInfo 操作系统信息
type OSInfo struct {
	OSName     string `json:"os_name"`
	OSVersion  string `json:"os_version"`
	OSFamily   string `json:"os_family"`
	Confidence int    `json:"confidence"`
}

// SecurityInfo 安全信息
type SecurityInfo struct {
	Vulnerabilities []string `json:"vulnerabilities"`
	SSLInfo         *SSLInfo `json:"ssl_info,omitempty"`
	Authentication  string   `json:"authentication"`
}

// SSLInfo SSL信息
type SSLInfo struct {
	Version     string `json:"version"`
	Cipher      string `json:"cipher"`
	Certificate string `json:"certificate"`
}

// NetworkInfo 网络信息
type NetworkInfo struct {
	ResponseTime int     `json:"response_time"`
	TTL          int     `json:"ttl"`
	PacketLoss   float64 `json:"packet_loss"`
}

// NewElasticsearchClient 创建Elasticsearch客户端
func NewElasticsearchClient(config *ESConfig) (*ElasticsearchClient, error) {
	cfg := elasticsearch.Config{
		Addresses: config.URLs,
		Username:  config.Username,
		Password:  config.Password,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建Elasticsearch客户端失败: %v", err)
	}

	// 测试连接
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("Elasticsearch连接测试失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch连接错误: %s", res.Status())
	}

	esClient := &ElasticsearchClient{
		client: client,
		config: config,
	}

	// 创建索引
	if err := esClient.createIndex(); err != nil {
		return nil, fmt.Errorf("创建索引失败: %v", err)
	}

	return esClient, nil
}

// createIndex 创建索引
func (es *ElasticsearchClient) createIndex() error {
	// 检查索引是否存在
	req := esapi.IndicesExistsRequest{
		Index: []string{es.config.Index},
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// 如果索引已存在，直接返回
	if res.StatusCode == 200 {
		return nil
	}

	// 创建索引映射
	mapping := `{
		"mappings": {
			"properties": {
				"ip": {"type": "ip"},
				"port": {"type": "integer"},
				"protocol": {"type": "keyword"},
				"service": {"type": "keyword"},
				"service_version": {"type": "text"},
				"banner": {"type": "text"},
				"state": {"type": "keyword"},
				"scan_time": {"type": "date"},
				"last_update": {"type": "date"},
				"task_id": {"type": "keyword"},
				"node_id": {"type": "keyword"},
				"applications": {
					"type": "nested",
					"properties": {
						"name": {"type": "keyword"},
						"version": {"type": "keyword"},
						"category": {"type": "keyword"},
						"confidence": {"type": "integer"},
						"cpe": {"type": "keyword"},
						"tags": {"type": "keyword"}
					}
				},
				"geo_info": {
					"properties": {
						"country": {"type": "keyword"},
						"country_code": {"type": "keyword"},
						"region": {"type": "keyword"},
						"city": {"type": "keyword"},
						"location": {"type": "geo_point"},
						"isp": {"type": "keyword"},
						"organization": {"type": "keyword"}
					}
				},
				"os_info": {
					"properties": {
						"os_name": {"type": "keyword"},
						"os_version": {"type": "keyword"},
						"os_family": {"type": "keyword"},
						"confidence": {"type": "integer"}
					}
				},
				"security_info": {
					"properties": {
						"vulnerabilities": {"type": "keyword"},
						"authentication": {"type": "keyword"}
					}
				},
				"network_info": {
					"properties": {
						"response_time": {"type": "integer"},
						"ttl": {"type": "integer"},
						"packet_loss": {"type": "float"}
					}
				},
				"metadata": {"type": "object"}
			}
		}
	}`

	createReq := esapi.IndicesCreateRequest{
		Index: es.config.Index,
		Body:  strings.NewReader(mapping),
	}

	createRes, err := createReq.Do(context.Background(), es.client)
	if err != nil {
		return err
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("创建索引失败: %s", createRes.Status())
	}

	return nil
}

// IndexDocument 索引文档
func (es *ElasticsearchClient) IndexDocument(doc *ScanDocument) error {
	// 生成文档ID (ip + port + time)
	docID := fmt.Sprintf("%s_%d_%d", doc.IP, doc.Port, doc.ScanTime.Unix())

	// 序列化文档
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("序列化文档失败: %v", err)
	}

	// 索引文档
	req := esapi.IndexRequest{
		Index:      es.config.Index,
		DocumentID: docID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("索引文档失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("索引文档错误: %s", res.Status())
	}

	return nil
}

// BulkIndexDocuments 批量索引文档
func (es *ElasticsearchClient) BulkIndexDocuments(docs []*ScanDocument) error {
	if len(docs) == 0 {
		return nil
	}

	var buf bytes.Buffer

	for _, doc := range docs {
		// 生成文档ID
		docID := fmt.Sprintf("%s_%d_%d", doc.IP, doc.Port, doc.ScanTime.Unix())

		// 添加索引操作
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": es.config.Index,
				"_id":    docID,
			},
		}

		metaData, _ := json.Marshal(meta)
		buf.Write(metaData)
		buf.WriteByte('\n')

		// 添加文档数据
		docData, _ := json.Marshal(doc)
		buf.Write(docData)
		buf.WriteByte('\n')
	}

	// 执行批量操作
	req := esapi.BulkRequest{
		Body:    &buf,
		Refresh: "true",
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("批量索引失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("批量索引错误: %s", res.Status())
	}

	return nil
}

// SearchResult 搜索结果
type SearchResult struct {
	Total int64          `json:"total"`
	Docs  []ScanDocument `json:"docs"`
}

// SearchDocuments 搜索文档
func (es *ElasticsearchClient) SearchDocuments(query map[string]interface{}) ([]ScanDocument, error) {
	result, err := es.SearchDocumentsWithTotal(query)
	if err != nil {
		return nil, err
	}
	return result.Docs, nil
}

// SearchDocumentsWithTotal 搜索文档并返回总数
func (es *ElasticsearchClient) SearchDocumentsWithTotal(query map[string]interface{}) (*SearchResult, error) {
	// 构建搜索请求
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("编码查询失败: %v", err)
	}

	// 执行搜索
	req := esapi.SearchRequest{
		Index: []string{es.config.Index},
		Body:  &buf,
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("搜索错误: %s", res.Status())
	}

	// 解析响应
	var response struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source ScanDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 提取文档
	var docs []ScanDocument
	for _, hit := range response.Hits.Hits {
		docs = append(docs, hit.Source)
	}

	return &SearchResult{
		Total: response.Hits.Total.Value,
		Docs:  docs,
	}, nil
}

// GetStats 获取索引统计
func (es *ElasticsearchClient) GetStats() (map[string]interface{}, error) {
	req := esapi.IndicesStatsRequest{
		Index: []string{es.config.Index},
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return nil, fmt.Errorf("获取统计失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("获取统计错误: %s", res.Status())
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("解析统计失败: %v", err)
	}

	return stats, nil
}

// Close 关闭客户端
func (es *ElasticsearchClient) Close() error {
	// Elasticsearch客户端不需要显式关闭
	return nil
}