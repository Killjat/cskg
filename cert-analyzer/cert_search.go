package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// CertificateSearcher 证书搜索器
type CertificateSearcher struct {
	config *CertificateSearchConfig
	client *http.Client
}

// NewCertificateSearcher 创建证书搜索器
func NewCertificateSearcher(config *CertificateSearchConfig) *CertificateSearcher {
	return &CertificateSearcher{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// SearchRelatedSites 搜索使用相同证书的网站
func (cs *CertificateSearcher) SearchRelatedSites(cert *x509.Certificate) *RelatedSites {
	if !cs.config.EnableSearch {
		return nil
	}

	startTime := time.Now()
	result := &RelatedSites{
		SearchMethod: strings.Join(cs.config.SearchMethods, ","),
		LastUpdated:  time.Now(),
		Sites:        make([]RelatedSite, 0),
	}

	// 计算证书指纹
	sha1Fingerprint := cs.calculateSHA1Fingerprint(cert)
	sha256Fingerprint := cs.calculateSHA256Fingerprint(cert)
	serialNumber := fmt.Sprintf("%X", cert.SerialNumber)

	var allSites []RelatedSite

	// 尝试不同的搜索方法
	for _, method := range cs.config.SearchMethods {
		var sites []RelatedSite
		var err error

		switch strings.ToLower(method) {
		case "fofa":
			sites, err = cs.searchWithFOFA(sha1Fingerprint, sha256Fingerprint, serialNumber)
		case "shodan":
			sites, err = cs.searchWithShodan(sha1Fingerprint, sha256Fingerprint)
		case "censys":
			sites, err = cs.searchWithCensys(sha1Fingerprint, sha256Fingerprint)
		case "crtsh":
			sites, err = cs.searchWithCrtSh(serialNumber, cert.Subject.CommonName)
		default:
			continue
		}

		if err != nil {
			if result.SearchError == "" {
				result.SearchError = fmt.Sprintf("%s: %v", method, err)
			} else {
				result.SearchError += fmt.Sprintf("; %s: %v", method, err)
			}
			continue
		}

		// 合并结果
		allSites = append(allSites, sites...)
	}

	// 去重和排序
	uniqueSites := cs.deduplicateSites(allSites)
	
	// 限制结果数量
	if len(uniqueSites) > cs.config.MaxResults {
		uniqueSites = uniqueSites[:cs.config.MaxResults]
	}

	result.Sites = uniqueSites
	result.TotalFound = len(uniqueSites)
	result.SearchTime = time.Since(startTime).Milliseconds()

	return result
}

// searchWithFOFA 使用FOFA搜索
func (cs *CertificateSearcher) searchWithFOFA(sha1, sha256, serial string) ([]RelatedSite, error) {
	if cs.config.FOFAConfig == nil || !cs.config.FOFAConfig.Enabled {
		return nil, fmt.Errorf("FOFA not configured")
	}

	var sites []RelatedSite

	// 使用SHA1指纹搜索
	query := fmt.Sprintf("cert=\"%s\"", sha1)
	fofaSites, err := cs.queryFOFA(query)
	if err == nil {
		sites = append(sites, fofaSites...)
	}

	// 使用序列号搜索
	if len(sites) < cs.config.MaxResults/2 {
		query = fmt.Sprintf("cert.serial=\"%s\"", serial)
		serialSites, err := cs.queryFOFA(query)
		if err == nil {
			sites = append(sites, serialSites...)
		}
	}

	return sites, nil
}

// queryFOFA 查询FOFA API
func (cs *CertificateSearcher) queryFOFA(query string) ([]RelatedSite, error) {
	// 构建FOFA API请求
	baseURL := "https://fofa.info/api/v1/search/all"
	params := url.Values{}
	params.Set("email", cs.config.FOFAConfig.Email)
	params.Set("key", cs.config.FOFAConfig.Key)
	params.Set("qbase64", cs.base64Encode(query))
	params.Set("fields", "host,title,server,country,org,lastupdatetime")
	params.Set("size", strconv.Itoa(cs.config.MaxResults))

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := cs.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("FOFA request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("FOFA API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read FOFA response: %v", err)
	}

	var fofaResp struct {
		Error   bool        `json:"error"`
		Message string      `json:"errmsg"`
		Size    int         `json:"size"`
		Results [][]string  `json:"results"`
	}

	if err := json.Unmarshal(body, &fofaResp); err != nil {
		return nil, fmt.Errorf("failed to parse FOFA response: %v", err)
	}

	if fofaResp.Error {
		return nil, fmt.Errorf("FOFA API error: %s", fofaResp.Message)
	}

	var sites []RelatedSite
	for _, result := range fofaResp.Results {
		if len(result) >= 6 {
			site := RelatedSite{
				URL:          cs.normalizeURL(result[0]),
				Domain:       cs.extractDomain(result[0]),
				Port:         cs.extractPort(result[0]),
				Title:        result[1],
				Server:       result[2],
				Country:      result[3],
				Organization: result[4],
				Confidence:   0.9, // FOFA结果置信度较高
				Source:       "FOFA",
			}

			// 解析最后更新时间
			if lastSeen, err := time.Parse("2006-01-02 15:04:05", result[5]); err == nil {
				site.LastSeen = lastSeen
			}

			sites = append(sites, site)
		}
	}

	return sites, nil
}

// searchWithShodan 使用Shodan搜索
func (cs *CertificateSearcher) searchWithShodan(sha1, sha256 string) ([]RelatedSite, error) {
	if cs.config.ShodanConfig == nil || !cs.config.ShodanConfig.Enabled {
		return nil, fmt.Errorf("Shodan not configured")
	}

	// Shodan搜索查询
	query := fmt.Sprintf("ssl.cert.fingerprint:%s", sha1)
	
	baseURL := "https://api.shodan.io/shodan/host/search"
	params := url.Values{}
	params.Set("key", cs.config.ShodanConfig.APIKey)
	params.Set("query", query)
	params.Set("limit", strconv.Itoa(cs.config.MaxResults))

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := cs.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("Shodan request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Shodan API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Shodan response: %v", err)
	}

	var shodanResp struct {
		Matches []struct {
			IPStr    string `json:"ip_str"`
			Port     int    `json:"port"`
			Hostnames []string `json:"hostnames"`
			Location struct {
				Country string `json:"country_name"`
			} `json:"location"`
			Timestamp string `json:"timestamp"`
			HTTP      struct {
				Title  string `json:"title"`
				Server string `json:"server"`
			} `json:"http"`
		} `json:"matches"`
	}

	if err := json.Unmarshal(body, &shodanResp); err != nil {
		return nil, fmt.Errorf("failed to parse Shodan response: %v", err)
	}

	var sites []RelatedSite
	for _, match := range shodanResp.Matches {
		domain := match.IPStr
		if len(match.Hostnames) > 0 {
			domain = match.Hostnames[0]
		}

		site := RelatedSite{
			URL:         fmt.Sprintf("https://%s:%d", domain, match.Port),
			Domain:      domain,
			Port:        match.Port,
			Title:       match.HTTP.Title,
			Server:      match.HTTP.Server,
			Country:     match.Location.Country,
			Confidence:  0.85, // Shodan结果置信度
			Source:      "Shodan",
		}

		// 解析时间戳
		if lastSeen, err := time.Parse(time.RFC3339, match.Timestamp); err == nil {
			site.LastSeen = lastSeen
		}

		sites = append(sites, site)
	}

	return sites, nil
}

// searchWithCensys 使用Censys搜索
func (cs *CertificateSearcher) searchWithCensys(sha1, sha256 string) ([]RelatedSite, error) {
	if cs.config.CensysConfig == nil || !cs.config.CensysConfig.Enabled {
		return nil, fmt.Errorf("Censys not configured")
	}

	// 这里可以实现Censys API调用
	// 由于Censys API较复杂，这里提供基本框架
	return []RelatedSite{}, nil
}

// searchWithCrtSh 使用crt.sh搜索
func (cs *CertificateSearcher) searchWithCrtSh(serial, commonName string) ([]RelatedSite, error) {
	// crt.sh是免费的证书透明度日志搜索
	query := fmt.Sprintf("https://crt.sh/?q=%s&output=json", url.QueryEscape(commonName))
	
	resp, err := cs.client.Get(query)
	if err != nil {
		return nil, fmt.Errorf("crt.sh request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("crt.sh error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read crt.sh response: %v", err)
	}

	var crtshResp []struct {
		CommonName string `json:"common_name"`
		NameValue  string `json:"name_value"`
		IssuerName string `json:"issuer_name"`
		NotBefore  string `json:"not_before"`
		NotAfter   string `json:"not_after"`
	}

	if err := json.Unmarshal(body, &crtshResp); err != nil {
		return nil, fmt.Errorf("failed to parse crt.sh response: %v", err)
	}

	var sites []RelatedSite
	domainSet := make(map[string]bool)

	for _, cert := range crtshResp {
		// 提取域名
		domains := strings.Split(cert.NameValue, "\n")
		for _, domain := range domains {
			domain = strings.TrimSpace(domain)
			if domain != "" && !domainSet[domain] {
				domainSet[domain] = true
				
				site := RelatedSite{
					URL:        fmt.Sprintf("https://%s", domain),
					Domain:     domain,
					Port:       443,
					Confidence: 0.7, // crt.sh结果置信度中等
					Source:     "crt.sh",
				}

				sites = append(sites, site)
				
				if len(sites) >= cs.config.MaxResults {
					break
				}
			}
		}
		
		if len(sites) >= cs.config.MaxResults {
			break
		}
	}

	return sites, nil
}

// 辅助函数
func (cs *CertificateSearcher) calculateSHA1Fingerprint(cert *x509.Certificate) string {
	hash := sha1.Sum(cert.Raw)
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

func (cs *CertificateSearcher) calculateSHA256Fingerprint(cert *x509.Certificate) string {
	hash := sha256.Sum256(cert.Raw)
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

func (cs *CertificateSearcher) base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func (cs *CertificateSearcher) normalizeURL(host string) string {
	if !strings.HasPrefix(host, "http") {
		return "https://" + host
	}
	return host
}

func (cs *CertificateSearcher) extractDomain(host string) string {
	// 提取域名部分
	if strings.Contains(host, "://") {
		parts := strings.Split(host, "://")
		if len(parts) > 1 {
			host = parts[1]
		}
	}
	
	if strings.Contains(host, ":") {
		parts := strings.Split(host, ":")
		return parts[0]
	}
	
	return host
}

func (cs *CertificateSearcher) extractPort(host string) int {
	if strings.Contains(host, ":") {
		parts := strings.Split(host, ":")
		if len(parts) > 1 {
			if port, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
				return port
			}
		}
	}
	return 443 // 默认HTTPS端口
}

func (cs *CertificateSearcher) deduplicateSites(sites []RelatedSite) []RelatedSite {
	seen := make(map[string]bool)
	var unique []RelatedSite

	for _, site := range sites {
		key := fmt.Sprintf("%s:%d", site.Domain, site.Port)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, site)
		}
	}

	return unique
}