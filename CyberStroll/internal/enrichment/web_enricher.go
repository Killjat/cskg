package enrichment

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/cskg/CyberStroll/internal/storage"
)

// WebEnricher 网站数据富化器
type WebEnricher struct {
	esClient    *storage.ElasticsearchClient
	httpClient  *http.Client
	config      *EnrichmentConfig
	logger      *log.Logger
	stats       *EnrichmentStats
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// EnrichmentConfig 富化配置
type EnrichmentConfig struct {
	BatchSize       int           `yaml:"batch_size"`
	WorkerCount     int           `yaml:"worker_count"`
	ScanInterval    time.Duration `yaml:"scan_interval"`
	RequestTimeout  time.Duration `yaml:"request_timeout"`
	MaxRetries      int           `yaml:"max_retries"`
	EnableCert      bool          `yaml:"enable_cert"`
	EnableAPI       bool          `yaml:"enable_api"`
	EnableWebInfo   bool          `yaml:"enable_web_info"`
	EnableFingerprint bool        `yaml:"enable_fingerprint"`
	EnableContent   bool          `yaml:"enable_content"`
}

// EnrichmentStats 富化统计
type EnrichmentStats struct {
	TotalProcessed   int64 `json:"total_processed"`
	SuccessEnriched  int64 `json:"success_enriched"`
	FailedEnriched   int64 `json:"failed_enriched"`
	LastProcessTime  int64 `json:"last_process_time"`
	ActiveWorkers    int   `json:"active_workers"`
	mutex            sync.RWMutex
}

// WebAsset Web资产信息
type WebAsset struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
	URL      string `json:"url"`
	LastScan time.Time `json:"last_scan"`
}

// EnrichmentData 富化数据
type EnrichmentData struct {
	CertInfo      *CertificateInfo `json:"cert_info,omitempty"`
	APIInfo       *APIInfo         `json:"api_info,omitempty"`
	WebInfo       *WebsiteInfo     `json:"web_info,omitempty"`
	Fingerprints  []Fingerprint    `json:"fingerprints,omitempty"`
	ContentInfo   *ContentInfo     `json:"content_info,omitempty"`
	EnrichTime    time.Time        `json:"enrich_time"`
}

// CertificateInfo 证书信息
type CertificateInfo struct {
	Subject         string    `json:"subject"`
	Issuer          string    `json:"issuer"`
	SerialNumber    string    `json:"serial_number"`
	NotBefore       time.Time `json:"not_before"`
	NotAfter        time.Time `json:"not_after"`
	SignatureAlg    string    `json:"signature_algorithm"`
	PublicKeyAlg    string    `json:"public_key_algorithm"`
	DNSNames        []string  `json:"dns_names"`
	IPAddresses     []string  `json:"ip_addresses"`
	IsCA            bool      `json:"is_ca"`
	IsSelfSigned    bool      `json:"is_self_signed"`
}

// APIInfo API信息
type APIInfo struct {
	Endpoints     []string          `json:"endpoints"`
	Methods       []string          `json:"methods"`
	Parameters    []string          `json:"parameters"`
	Documentation string            `json:"documentation"`
	Version       string            `json:"version"`
	Framework     string            `json:"framework"`
	Metadata      map[string]string `json:"metadata"`
}

// WebsiteInfo 网站信息
type WebsiteInfo struct {
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Keywords      string            `json:"keywords"`
	Language      string            `json:"language"`
	Charset       string            `json:"charset"`
	Generator     string            `json:"generator"`
	Author        string            `json:"author"`
	Copyright     string            `json:"copyright"`
	Links         []string          `json:"links"`
	Images        []string          `json:"images"`
	Scripts       []string          `json:"scripts"`
	Stylesheets   []string          `json:"stylesheets"`
	MetaTags      map[string]string `json:"meta_tags"`
}

// Fingerprint 指纹信息
type Fingerprint struct {
	Technology string            `json:"technology"`
	Version    string            `json:"version"`
	Category   string            `json:"category"`
	Confidence int               `json:"confidence"`
	Evidence   map[string]string `json:"evidence"`
}

// ContentInfo 内容信息
type ContentInfo struct {
	StatusCode    int               `json:"status_code"`
	ContentType   string            `json:"content_type"`
	ContentLength int64             `json:"content_length"`
	Headers       map[string]string `json:"headers"`
	BodyHash      string            `json:"body_hash"`
	BodyPreview   string            `json:"body_preview"`
	ResponseTime  int64             `json:"response_time"`
}

// NewWebEnricher 创建网站数据富化器
func NewWebEnricher(esClient *storage.ElasticsearchClient, config *EnrichmentConfig, logger *log.Logger) *WebEnricher {
	if config == nil {
		config = &EnrichmentConfig{
			BatchSize:         50,
			WorkerCount:       5,
			ScanInterval:      time.Minute * 5,
			RequestTimeout:    time.Second * 30,
			MaxRetries:        3,
			EnableCert:        true,
			EnableAPI:         true,
			EnableWebInfo:     true,
			EnableFingerprint: true,
			EnableContent:     true,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	// 创建HTTP客户端
	httpClient := &http.Client{
		Timeout: config.RequestTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 跳过证书验证以获取证书信息
			},
		},
	}

	return &WebEnricher{
		esClient:   esClient,
		httpClient: httpClient,
		config:     config,
		logger:     logger,
		stats:      &EnrichmentStats{},
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start 启动富化器
func (we *WebEnricher) Start() error {
	we.logger.Println("启动网站数据富化器...")

	// 启动工作协程
	for i := 0; i < we.config.WorkerCount; i++ {
		we.wg.Add(1)
		go we.worker(i, &we.wg)
	}

	// 启动统计打印协程
	we.wg.Add(1)
	go we.printStats(&we.wg)

	return nil
}

// Stop 停止富化器
func (we *WebEnricher) Stop() {
	we.logger.Println("正在停止网站数据富化器...")
	we.cancel()
	we.wg.Wait()
	we.logger.Println("网站数据富化器已停止")
}

// worker 工作协程
func (we *WebEnricher) worker(workerID int, wg *sync.WaitGroup) {
	defer wg.Done()

	we.logger.Printf("启动富化工作协程 %d", workerID)

	ticker := time.NewTicker(we.config.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-we.ctx.Done():
			we.logger.Printf("富化工作协程 %d 停止", workerID)
			return
		case <-ticker.C:
			we.processWebAssets(workerID)
		}
	}
}

// processWebAssets 处理Web资产
func (we *WebEnricher) processWebAssets(workerID int) {
	we.logger.Printf("工作协程 %d 开始处理Web资产", workerID)

	// 查询需要富化的Web资产
	assets, err := we.getWebAssetsForEnrichment()
	if err != nil {
		we.logger.Printf("获取Web资产失败: %v", err)
		return
	}

	if len(assets) == 0 {
		we.logger.Printf("工作协程 %d 没有找到需要富化的Web资产", workerID)
		return
	}

	we.logger.Printf("工作协程 %d 找到 %d 个Web资产需要富化", workerID, len(assets))

	// 处理每个资产
	for _, asset := range assets {
		select {
		case <-we.ctx.Done():
			return
		default:
			we.enrichWebAsset(asset)
		}
	}
}

// getWebAssetsForEnrichment 获取需要富化的Web资产
func (we *WebEnricher) getWebAssetsForEnrichment() ([]*WebAsset, error) {
	// 构建查询：查找http/https协议且未富化或富化时间较久的资产
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"terms": map[string]interface{}{
							"service": []string{"http", "https"},
						},
					},
					{
						"term": map[string]interface{}{
							"state": "open",
						},
					},
				},
				"should": []map[string]interface{}{
					{
						"bool": map[string]interface{}{
							"must_not": map[string]interface{}{
								"exists": map[string]interface{}{
									"field": "enrichment_data.enrich_time",
								},
							},
						},
					},
					{
						"range": map[string]interface{}{
							"enrichment_data.enrich_time": map[string]interface{}{
								"lt": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
							},
						},
					},
				},
				"minimum_should_match": 1,
			},
		},
		"size": we.config.BatchSize,
		"sort": []map[string]interface{}{
			{
				"scan_time": map[string]interface{}{
					"order": "desc",
				},
			},
		},
	}

	docs, err := we.esClient.SearchDocuments(query)
	if err != nil {
		return nil, fmt.Errorf("搜索Web资产失败: %v", err)
	}

	var assets []*WebAsset
	for _, doc := range docs {
		if doc.Service == "http" || doc.Service == "https" {
			asset := &WebAsset{
				IP:       doc.IP,
				Port:     doc.Port,
				Protocol: doc.Protocol,
				Service:  doc.Service,
				URL:      fmt.Sprintf("%s://%s:%d", doc.Service, doc.IP, doc.Port),
				LastScan: doc.ScanTime,
			}
			assets = append(assets, asset)
		}
	}

	return assets, nil
}

// enrichWebAsset 富化单个Web资产
func (we *WebEnricher) enrichWebAsset(asset *WebAsset) {
	we.logger.Printf("开始富化Web资产: %s", asset.URL)

	enrichData := &EnrichmentData{
		EnrichTime: time.Now(),
	}

	// 执行各种富化操作
	if we.config.EnableContent {
		enrichData.ContentInfo = we.enrichContent(asset)
	}

	if we.config.EnableCert && asset.Service == "https" {
		enrichData.CertInfo = we.enrichCertificate(asset)
	}

	if we.config.EnableWebInfo {
		enrichData.WebInfo = we.enrichWebsiteInfo(asset)
	}

	if we.config.EnableFingerprint {
		enrichData.Fingerprints = we.enrichFingerprints(asset)
	}

	if we.config.EnableAPI {
		enrichData.APIInfo = we.enrichAPIInfo(asset)
	}

	// 更新到ES
	if err := we.updateEnrichmentData(asset, enrichData); err != nil {
		we.logger.Printf("更新富化数据失败: %v", err)
		we.updateStats(false)
		return
	}

	we.logger.Printf("Web资产富化完成: %s", asset.URL)
	we.updateStats(true)
}
// enrichContent 富化内容信息
func (we *WebEnricher) enrichContent(asset *WebAsset) *ContentInfo {
	startTime := time.Now()
	
	resp, err := we.httpClient.Get(asset.URL)
	if err != nil {
		we.logger.Printf("请求失败 %s: %v", asset.URL, err)
		return nil
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		we.logger.Printf("读取响应体失败 %s: %v", asset.URL, err)
		return nil
	}

	// 提取headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// 计算响应时间
	responseTime := time.Since(startTime).Milliseconds()

	// 生成body预览（前1000字符）
	bodyPreview := string(body)
	if len(bodyPreview) > 1000 {
		bodyPreview = bodyPreview[:1000] + "..."
	}

	return &ContentInfo{
		StatusCode:    resp.StatusCode,
		ContentType:   resp.Header.Get("Content-Type"),
		ContentLength: int64(len(body)),
		Headers:       headers,
		BodyHash:      fmt.Sprintf("%x", body), // 简化的hash
		BodyPreview:   bodyPreview,
		ResponseTime:  responseTime,
	}
}

// enrichCertificate 富化证书信息
func (we *WebEnricher) enrichCertificate(asset *WebAsset) *CertificateInfo {
	if asset.Service != "https" {
		return nil
	}

	// 解析URL获取host
	u, err := url.Parse(asset.URL)
	if err != nil {
		return nil
	}

	// 建立TLS连接获取证书
	conn, err := tls.Dial("tcp", u.Host, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		we.logger.Printf("TLS连接失败 %s: %v", asset.URL, err)
		return nil
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil
	}

	cert := certs[0]

	// 提取DNS名称和IP地址
	var dnsNames []string
	var ipAddresses []string
	
	for _, name := range cert.DNSNames {
		dnsNames = append(dnsNames, name)
	}
	
	for _, ip := range cert.IPAddresses {
		ipAddresses = append(ipAddresses, ip.String())
	}

	return &CertificateInfo{
		Subject:         cert.Subject.String(),
		Issuer:          cert.Issuer.String(),
		SerialNumber:    cert.SerialNumber.String(),
		NotBefore:       cert.NotBefore,
		NotAfter:        cert.NotAfter,
		SignatureAlg:    cert.SignatureAlgorithm.String(),
		PublicKeyAlg:    cert.PublicKeyAlgorithm.String(),
		DNSNames:        dnsNames,
		IPAddresses:     ipAddresses,
		IsCA:            cert.IsCA,
		IsSelfSigned:    cert.Issuer.String() == cert.Subject.String(),
	}
}

// enrichWebsiteInfo 富化网站信息
func (we *WebEnricher) enrichWebsiteInfo(asset *WebAsset) *WebsiteInfo {
	resp, err := we.httpClient.Get(asset.URL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	content := string(body)
	
	webInfo := &WebsiteInfo{
		MetaTags: make(map[string]string),
	}

	// 简单的HTML解析（实际应该使用专业的HTML解析器）
	webInfo.Title = we.extractBetween(content, "<title>", "</title>")
	webInfo.Description = we.extractMetaContent(content, "description")
	webInfo.Keywords = we.extractMetaContent(content, "keywords")
	webInfo.Generator = we.extractMetaContent(content, "generator")
	webInfo.Author = we.extractMetaContent(content, "author")

	// 提取链接、图片、脚本等
	webInfo.Links = we.extractLinks(content)
	webInfo.Images = we.extractImages(content)
	webInfo.Scripts = we.extractScripts(content)

	return webInfo
}

// enrichFingerprints 富化指纹信息
func (we *WebEnricher) enrichFingerprints(asset *WebAsset) []Fingerprint {
	var fingerprints []Fingerprint

	resp, err := we.httpClient.Get(asset.URL)
	if err != nil {
		return fingerprints
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fingerprints
	}

	content := string(body)
	headers := resp.Header

	// 基于响应头的指纹识别
	if server := headers.Get("Server"); server != "" {
		if strings.Contains(strings.ToLower(server), "nginx") {
			fingerprints = append(fingerprints, Fingerprint{
				Technology: "Nginx",
				Category:   "Web Server",
				Confidence: 90,
				Evidence:   map[string]string{"header": "Server: " + server},
			})
		}
		if strings.Contains(strings.ToLower(server), "apache") {
			fingerprints = append(fingerprints, Fingerprint{
				Technology: "Apache",
				Category:   "Web Server", 
				Confidence: 90,
				Evidence:   map[string]string{"header": "Server: " + server},
			})
		}
	}

	// 基于内容的指纹识别
	if strings.Contains(content, "wp-content") || strings.Contains(content, "wordpress") {
		fingerprints = append(fingerprints, Fingerprint{
			Technology: "WordPress",
			Category:   "CMS",
			Confidence: 85,
			Evidence:   map[string]string{"content": "wp-content detected"},
		})
	}

	if strings.Contains(content, "jquery") {
		fingerprints = append(fingerprints, Fingerprint{
			Technology: "jQuery",
			Category:   "JavaScript Library",
			Confidence: 80,
			Evidence:   map[string]string{"content": "jquery detected"},
		})
	}

	return fingerprints
}

// enrichAPIInfo 富化API信息
func (we *WebEnricher) enrichAPIInfo(asset *WebAsset) *APIInfo {
	apiInfo := &APIInfo{
		Metadata: make(map[string]string),
	}

	// 尝试常见的API文档路径
	apiPaths := []string{
		"/api/docs",
		"/swagger",
		"/swagger-ui",
		"/api-docs",
		"/docs",
		"/openapi.json",
		"/swagger.json",
	}

	for _, path := range apiPaths {
		apiURL := asset.URL + path
		resp, err := we.httpClient.Get(apiURL)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			apiInfo.Documentation = apiURL
			apiInfo.Endpoints = append(apiInfo.Endpoints, path)
			
			// 尝试解析API文档内容
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				content := string(body)
				if strings.Contains(content, "swagger") {
					apiInfo.Framework = "Swagger"
				}
				if strings.Contains(content, "openapi") {
					apiInfo.Framework = "OpenAPI"
				}
			}
			break
		}
	}

	return apiInfo
}

// updateEnrichmentData 更新富化数据到ES
func (we *WebEnricher) updateEnrichmentData(asset *WebAsset, enrichData *EnrichmentData) error {
	// 先获取现有文档
	existingDoc, err := we.getExistingDocument(asset)
	if err != nil {
		we.logger.Printf("获取现有文档失败: %v", err)
		return err
	}

	// 如果现有文档的Metadata为空，初始化它
	if existingDoc.Metadata == nil {
		existingDoc.Metadata = make(map[string]interface{})
	}

	// 添加富化数据到现有文档
	existingDoc.Metadata["enrichment_data"] = enrichData
	existingDoc.LastUpdate = time.Now()

	// 更新到ES
	return we.esClient.IndexDocument(existingDoc)
}

// getExistingDocument 获取现有文档
func (we *WebEnricher) getExistingDocument(asset *WebAsset) (*storage.ScanDocument, error) {
	// 构建查询：根据IP和端口查找文档
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"ip": asset.IP,
						},
					},
					{
						"term": map[string]interface{}{
							"port": asset.Port,
						},
					},
				},
			},
		},
		"size": 1,
		"sort": []map[string]interface{}{
			{
				"scan_time": map[string]interface{}{
					"order": "desc",
				},
			},
		},
	}

	docs, err := we.esClient.SearchDocuments(query)
	if err != nil {
		return nil, fmt.Errorf("搜索现有文档失败: %v", err)
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("未找到IP %s 端口 %d 的现有文档", asset.IP, asset.Port)
	}

	return &docs[0], nil
}

// updateStats 更新统计信息
func (we *WebEnricher) updateStats(success bool) {
	we.stats.mutex.Lock()
	defer we.stats.mutex.Unlock()

	we.stats.TotalProcessed++
	if success {
		we.stats.SuccessEnriched++
	} else {
		we.stats.FailedEnriched++
	}
	we.stats.LastProcessTime = time.Now().Unix()
}

// GetStats 获取统计信息
func (we *WebEnricher) GetStats() *EnrichmentStats {
	we.stats.mutex.RLock()
	defer we.stats.mutex.RUnlock()

	return &EnrichmentStats{
		TotalProcessed:  we.stats.TotalProcessed,
		SuccessEnriched: we.stats.SuccessEnriched,
		FailedEnriched:  we.stats.FailedEnriched,
		LastProcessTime: we.stats.LastProcessTime,
		ActiveWorkers:   we.config.WorkerCount,
	}
}

// printStats 打印统计信息
func (we *WebEnricher) printStats(wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-we.ctx.Done():
			return
		case <-ticker.C:
			stats := we.GetStats()
			we.logger.Printf("富化统计: 总处理=%d, 成功=%d, 失败=%d, 活跃工作协程=%d",
				stats.TotalProcessed, stats.SuccessEnriched, stats.FailedEnriched, stats.ActiveWorkers)
		}
	}
}

// 辅助函数
func (we *WebEnricher) extractBetween(content, start, end string) string {
	startIdx := strings.Index(strings.ToLower(content), strings.ToLower(start))
	if startIdx == -1 {
		return ""
	}
	startIdx += len(start)
	
	endIdx := strings.Index(strings.ToLower(content[startIdx:]), strings.ToLower(end))
	if endIdx == -1 {
		return ""
	}
	
	return strings.TrimSpace(content[startIdx : startIdx+endIdx])
}

func (we *WebEnricher) extractMetaContent(content, name string) string {
	// 简化实现，实际应该使用正则表达式
	return ""
}

func (we *WebEnricher) extractLinks(content string) []string {
	var links []string
	// 简化实现，实际应该使用HTML解析器
	return links
}

func (we *WebEnricher) extractImages(content string) []string {
	var images []string
	// 简化实现，实际应该使用HTML解析器
	return images
}

func (we *WebEnricher) extractScripts(content string) []string {
	var scripts []string
	// 简化实现，实际应该使用HTML解析器
	return scripts
}