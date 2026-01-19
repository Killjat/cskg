package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// BatchAnalyzer 批量分析器
type BatchAnalyzer struct {
	analyzer    *CertificateAnalyzer
	concurrency int
}

// NewBatchAnalyzer 创建批量分析器
func NewBatchAnalyzer(analyzer *CertificateAnalyzer, concurrency int) *BatchAnalyzer {
	return &BatchAnalyzer{
		analyzer:    analyzer,
		concurrency: concurrency,
	}
}

// AnalyzeFromFile 从文件读取URL列表并批量分析
func (ba *BatchAnalyzer) AnalyzeFromFile(filename string) (*BatchResult, error) {
	urls, err := ba.readURLsFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read URLs from file: %v", err)
	}

	return ba.AnalyzeURLs(urls), nil
}

// AnalyzeURLs 批量分析URL列表
func (ba *BatchAnalyzer) AnalyzeURLs(urls []string) *BatchResult {
	result := &BatchResult{
		TotalURLs: len(urls),
		Results:   make([]CertificateResult, 0, len(urls)),
	}

	// 创建工作通道
	urlChan := make(chan string, len(urls))
	resultChan := make(chan CertificateResult, len(urls))

	// 发送URL到通道
	for _, url := range urls {
		urlChan <- url
	}
	close(urlChan)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < ba.concurrency; i++ {
		wg.Add(1)
		go ba.worker(urlChan, resultChan, &wg)
	}

	// 等待所有工作完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	for certResult := range resultChan {
		result.Results = append(result.Results, certResult)
		if certResult.Status == "success" {
			result.SuccessCount++
		} else {
			result.FailureCount++
		}
	}

	// 生成摘要
	result.Summary = ba.generateSummary(result.Results)

	return result
}

// worker 工作协程
func (ba *BatchAnalyzer) worker(urlChan <-chan string, resultChan chan<- CertificateResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for url := range urlChan {
		result := ba.analyzer.AnalyzeURL(url)
		resultChan <- *result
	}
}

// readURLsFromFile 从文件读取URL列表
func (ba *BatchAnalyzer) readURLsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			// 确保URL有协议前缀
			if !strings.HasPrefix(line, "http://") && !strings.HasPrefix(line, "https://") {
				line = "https://" + line
			}
			urls = append(urls, line)
		}
	}

	return urls, scanner.Err()
}

// generateSummary 生成批量分析摘要
func (ba *BatchAnalyzer) generateSummary(results []CertificateResult) *BatchSummary {
	summary := &BatchSummary{
		CommonIssuers: make([]string, 0),
	}

	issuerCount := make(map[string]int)
	totalScore := 0
	validResults := 0

	for _, result := range results {
		if result.Status != "success" || result.Certificate == nil {
			continue
		}

		validResults++

		// 统计过期证书
		if result.SecurityAnalysis.IsExpired {
			summary.ExpiredCerts++
		}

		// 统计即将过期的证书
		if result.SecurityAnalysis.ExpiresSoon {
			summary.ExpiringSoon++
		}

		// 统计自签名证书
		if result.SecurityAnalysis.IsSelfSigned {
			summary.SelfSignedCerts++
		}

		// 统计弱签名
		if result.SecurityAnalysis.WeakSignature {
			summary.WeakSignatures++
		}

		// 统计颁发者
		issuer := result.Certificate.Issuer.CommonName
		if issuer != "" {
			issuerCount[issuer]++
		}

		// 累计安全评分
		totalScore += result.SecurityAnalysis.SecurityScore
	}

	// 计算平均安全评分
	if validResults > 0 {
		summary.AverageScore = float64(totalScore) / float64(validResults)
	}

	// 找出最常见的颁发者
	for issuer, count := range issuerCount {
		if count >= 2 { // 至少出现2次才算常见
			summary.CommonIssuers = append(summary.CommonIssuers, fmt.Sprintf("%s (%d)", issuer, count))
		}
	}

	return summary
}

// ExportResults 导出结果到文件
func (ba *BatchAnalyzer) ExportResults(result *BatchResult, filename string, format string) error {
	switch strings.ToLower(format) {
	case "json":
		return ba.exportJSON(result, filename)
	case "csv":
		return ba.exportCSV(result, filename)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// exportJSON 导出为JSON格式
func (ba *BatchAnalyzer) exportJSON(result *BatchResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// exportCSV 导出为CSV格式
func (ba *BatchAnalyzer) exportCSV(result *BatchResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入CSV头部
	headers := []string{
		"URL", "Status", "Common Name", "Issuer", "Not Before", "Not After", 
		"Days Remaining", "Is Expired", "Expires Soon", "Self Signed", 
		"Weak Signature", "Security Score", "TLS Version", "Cipher Suite",
	}
	
	if err := writer.Write(headers); err != nil {
		return err
	}

	// 写入数据行
	for _, res := range result.Results {
		row := []string{
			res.URL,
			res.Status,
		}

		if res.Certificate != nil {
			row = append(row,
				res.Certificate.Subject.CommonName,
				res.Certificate.Issuer.CommonName,
				res.Certificate.Validity.NotBefore.Format(time.RFC3339),
				res.Certificate.Validity.NotAfter.Format(time.RFC3339),
				fmt.Sprintf("%d", res.Certificate.Validity.DaysRemaining),
				fmt.Sprintf("%t", res.SecurityAnalysis.IsExpired),
				fmt.Sprintf("%t", res.SecurityAnalysis.ExpiresSoon),
				fmt.Sprintf("%t", res.SecurityAnalysis.IsSelfSigned),
				fmt.Sprintf("%t", res.SecurityAnalysis.WeakSignature),
				fmt.Sprintf("%d", res.SecurityAnalysis.SecurityScore),
			)

			if res.ConnectionInfo != nil {
				row = append(row,
					res.ConnectionInfo.TLSVersion,
					res.ConnectionInfo.CipherSuite,
				)
			} else {
				row = append(row, "", "")
			}
		} else {
			// 填充空值
			for i := 0; i < len(headers)-2; i++ {
				row = append(row, "")
			}
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}