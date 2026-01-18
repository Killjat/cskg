package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/cskg/assetdiscovery/common"
)

// WebScanner Web站点扫描器结构体
type WebScanner struct {
	client *Client
	httpClient *http.Client
}

// NewWebScanner 创建新的Web扫描器
func NewWebScanner(client *Client) *WebScanner {
	// 创建HTTP客户端
	httpClient := &http.Client{
		Timeout: time.Duration(client.config.Scan.HTTPTimeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 跳过证书验证
			MaxIdleConns:    100,
			IdleConnTimeout: 90 * time.Second,
		},
	}

	return &WebScanner{
		client:     client,
		httpClient: httpClient,
	}
}

// ScanWebService 扫描Web服务，提取详细信息
func (ws *WebScanner) ScanWebService(result *common.Result) {
	// 构建URL
	protocol := "http"
	if result.Service == "https" || result.Port == 443 {
		protocol = "https"
	}
	url := protocol + "://" + fmt.Sprintf("%s:%d", result.Target, result.Port)

	log.Printf("Scanning web service at %s...", url)

	// 发送HTTP请求
	resp, body, err := ws.sendHTTPRequest(url)
	if err != nil {
		log.Printf("Error scanning web service %s: %v", url, err)
		return
	}

	// 提取Web信息
	webInfo := &common.WebInfo{
		URL:        url,
		StatusCode: resp.StatusCode,
		Headers:    make(map[string]string),
		Fingerprint: []string{},
	}

	// 提取响应头
	for key, values := range resp.Header {
		webInfo.Headers[key] = strings.Join(values, ", ")
	}

	// 提取标题
	webInfo.Title = ws.extractTitle(body)

	// 检测登录框
	webInfo.HasLogin = ws.detectLoginForm(body)

	// 识别网站指纹
	webInfo.Fingerprint = ws.identifyFingerprint(resp, body)

	// 提取ICP备案信息
	webInfo.ICPInfo = ws.getICPInfo(result.Target)

	// 设置WebInfo到结果中
	result.WebInfo = webInfo

	log.Printf("Web service scanned successfully: %s - %s", url, webInfo.Title)
}

// sendHTTPRequest 发送HTTP请求并返回响应和body
func (ws *WebScanner) sendHTTPRequest(urlStr string) (*http.Response, string, error) {
	// 创建HTTP请求
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, "", err
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8,en-US;q=0.5,en;q=0.3")

	// 发送请求
	resp, err := ws.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return resp, string(body), nil
}

// extractTitle 从HTML中提取标题
func (ws *WebScanner) extractTitle(body string) string {
	// 使用正则表达式提取标题
	re := regexp.MustCompile(`<title[^>]*>(.*?)</title>`)
	matches := re.FindStringSubmatch(body)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// 如果没有找到标题，返回空字符串
	return ""
}

// detectLoginForm 检测HTML中是否包含登录框
func (ws *WebScanner) detectLoginForm(body string) bool {
	// 简化实现，检测常见的登录表单特征
	loginIndicators := []string{
		"<input.*?type=[\"']password[\"']",
		"<form.*?login",
		"<form.*?auth",
		"<form.*?signin",
		"<input.*?name=[\"']username[\"']",
		"<input.*?name=[\"']user[\"']",
		"<input.*?name=[\"']email[\"']",
	}

	for _, indicator := range loginIndicators {
		if matched, _ := regexp.MatchString(indicator, body); matched {
			return true
		}
	}

	return false
}

// identifyFingerprint 识别网站指纹
func (ws *WebScanner) identifyFingerprint(resp *http.Response, body string) []string {
	// 简化实现，检测常见的Web服务器和框架
	var fingerprints []string

	// 检测Server头
	if server, exists := resp.Header["Server"]; exists && len(server) > 0 {
		fingerprints = append(fingerprints, "Server: "+server[0])
	}

	// 检测X-Powered-By头
	if poweredBy, exists := resp.Header["X-Powered-By"]; exists && len(poweredBy) > 0 {
		fingerprints = append(fingerprints, "X-Powered-By: "+poweredBy[0])
	}

	// 检测X-Frame-Options头
	if frameOptions, exists := resp.Header["X-Frame-Options"]; exists && len(frameOptions) > 0 {
		fingerprints = append(fingerprints, "X-Frame-Options: "+frameOptions[0])
	}

	// 检测Content-Type头
	if contentType, exists := resp.Header["Content-Type"]; exists && len(contentType) > 0 {
		fingerprints = append(fingerprints, "Content-Type: "+contentType[0])
	}

	// 检测常见的框架特征
	frameworks := map[string]string{
		"WordPress":         "wp-content",
		"Drupal":            "sites/all",
		"Joomla":            "joomla",
		"Laravel":           "laravel",
		"Vue.js":            "vue",
		"React":             "react",
		"Angular":           "angular",
		"jQuery":            "jquery",
		"Bootstrap":         "bootstrap",
	}

	for name, indicator := range frameworks {
		if strings.Contains(strings.ToLower(body), strings.ToLower(indicator)) {
			fingerprints = append(fingerprints, "Framework: "+name)
		}
	}

	return fingerprints
}

// getICPInfo 获取ICP备案信息
func (ws *WebScanner) getICPInfo(domain string) *common.ICPInfo {
	// 简化实现，实际应该调用ICP查询API
	// 这里返回模拟数据
	return &common.ICPInfo{
		Domain:      domain,
		ICP:         "", // 实际应从备案信息中提取
		CompanyName: "",
		Valid:       false,
	}
}
