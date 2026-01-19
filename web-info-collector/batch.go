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

// BatchCollector 批量收集器
type BatchCollector struct {
	collector   *WebInfoCollector
	concurrency int
}

// NewBatchCollector 创建批量收集器
func NewBatchCollector(collector *WebInfoCollector, concurrency int) *BatchCollector {
	return &BatchCollector{
		collector:   collector,
		concurrency: concurrency,
	}
}

// CollectFromFile 从文件读取URL列表并批量收集
func (bc *BatchCollector) CollectFromFile(filename string) (*BatchResult, error) {
	urls, err := bc.readURLsFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read URLs from file: %v", err)
	}

	return bc.CollectFromURLs(urls), nil
}

// CollectFromURLs 批量收集URL列表
func (bc *BatchCollector) CollectFromURLs(urls []string) *BatchResult {
	startTime := time.Now()
	
	result := &BatchResult{
		TotalURLs: len(urls),
		Results:   make([]WebInfo, 0, len(urls)),
		StartTime: startTime,
	}

	// 创建工作通道
	urlChan := make(chan string, len(urls))
	resultChan := make(chan WebInfo, len(urls))

	// 发送URL到通道
	for _, url := range urls {
		urlChan <- url
	}
	close(urlChan)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < bc.concurrency; i++ {
		wg.Add(1)
		go bc.worker(urlChan, resultChan, &wg)
	}

	// 等待所有工作完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	for webInfo := range resultChan {
		result.Results = append(result.Results, webInfo)
		if webInfo.Status == "success" {
			result.SuccessCount++
		} else {
			result.FailureCount++
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// 生成摘要
	result.Summary = bc.generateSummary(result.Results)

	return result
}

// worker 工作协程
func (bc *BatchCollector) worker(urlChan <-chan string, resultChan chan<- WebInfo, wg *sync.WaitGroup) {
	defer wg.Done()

	for url := range urlChan {
		result := bc.collector.CollectWebInfo(url)
		resultChan <- *result
	}
}

// readURLsFromFile 从文件读取URL列表
func (bc *BatchCollector) readURLsFromFile(filename string) ([]string, error) {
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

// generateSummary 生成批量收集摘要
func (bc *BatchCollector) generateSummary(results []WebInfo) *Summary {
	summary := &Summary{
		CommonTechnologies: make([]TechCount, 0),
		ICPStatistics:      &ICPStats{},
		FileTypeStats:      make([]FileTypeCount, 0),
		DomainStats:        make([]DomainCount, 0),
	}

	techCount := make(map[string]int)
	fileTypeCount := make(map[string]int)
	domainCount := make(map[string]int)
	
	totalWithICP := 0
	totalWithPolice := 0
	validResults := 0

	for _, result := range results {
		if result.Status != "success" {
			continue
		}

		validResults++

		// 统计技术栈
		if result.TechnicalInfo != nil {
			if result.TechnicalInfo.CMS != "" {
				techCount[result.TechnicalInfo.CMS]++
			}
			if result.TechnicalInfo.Server != "" {
				techCount[result.TechnicalInfo.Server]++
			}
			for _, framework := range result.TechnicalInfo.Frameworks {
				techCount[framework]++
			}
		}

		// 统计备案信息
		if result.RegistrationInfo != nil {
			if result.RegistrationInfo.ICPLicense != "" {
				totalWithICP++
			}
			if result.RegistrationInfo.PoliceRecord != "" {
				totalWithPolice++
			}
		}

		// 统计文件类型
		for _, download := range result.DownloadLinks {
			fileType := strings.Split(download.Type, "/")[0]
			fileTypeCount[fileType]++
		}

		// 统计域名
		if result.BasicInfo != nil && result.URL != "" {
			// 简单提取域名
			parts := strings.Split(result.URL, "/")
			if len(parts) > 2 {
				domain := parts[2]
				domainCount[domain]++
			}
		}
	}

	// 计算ICP统计
	if validResults > 0 {
		summary.ICPStatistics.TotalWithICP = totalWithICP
		summary.ICPStatistics.TotalWithPolice = totalWithPolice
		summary.ICPStatistics.ICPPercentage = float64(totalWithICP) / float64(validResults) * 100
		summary.ICPStatistics.PolicePercentage = float64(totalWithPolice) / float64(validResults) * 100
	}

	// 转换技术统计
	for tech, count := range techCount {
		if count >= 2 { // 至少出现2次
			percentage := float64(count) / float64(validResults) * 100
			summary.CommonTechnologies = append(summary.CommonTechnologies, TechCount{
				Technology: tech,
				Count:      count,
				Percentage: percentage,
			})
		}
	}

	// 转换文件类型统计
	for fileType, count := range fileTypeCount {
		summary.FileTypeStats = append(summary.FileTypeStats, FileTypeCount{
			FileType: fileType,
			Count:    count,
		})
	}

	// 转换域名统计
	for domain, count := range domainCount {
		summary.DomainStats = append(summary.DomainStats, DomainCount{
			Domain: domain,
			Count:  count,
		})
	}

	return summary
}

// ExportResults 导出结果到文件
func (bc *BatchCollector) ExportResults(result *BatchResult, filename string, format string) error {
	switch strings.ToLower(format) {
	case "json":
		return bc.exportJSON(result, filename)
	case "csv":
		return bc.exportCSV(result, filename)
	case "html":
		return bc.exportHTML(result, filename)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// exportJSON 导出为JSON格式
func (bc *BatchCollector) exportJSON(result *BatchResult, filename string) error {
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
func (bc *BatchCollector) exportCSV(result *BatchResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入CSV头部
	headers := []string{
		"URL", "Status", "Title", "Description", "ICP License", "Police Record",
		"Organization", "Server", "CMS", "Frameworks", "Download Links Count",
		"Footer Copyright", "Contact Email", "Contact Phone", "Crawl Duration",
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

		if res.BasicInfo != nil {
			row = append(row, res.BasicInfo.Title, res.BasicInfo.Description)
		} else {
			row = append(row, "", "")
		}

		if res.RegistrationInfo != nil {
			row = append(row,
				res.RegistrationInfo.ICPLicense,
				res.RegistrationInfo.PoliceRecord,
				res.RegistrationInfo.Organization,
			)
		} else {
			row = append(row, "", "", "")
		}

		if res.TechnicalInfo != nil {
			frameworks := strings.Join(res.TechnicalInfo.Frameworks, "; ")
			row = append(row,
				res.TechnicalInfo.Server,
				res.TechnicalInfo.CMS,
				frameworks,
			)
		} else {
			row = append(row, "", "", "")
		}

		row = append(row, fmt.Sprintf("%d", len(res.DownloadLinks)))

		if res.FooterInfo != nil {
			emails := ""
			phones := ""
			if res.FooterInfo.ContactInfo != nil {
				emails = strings.Join(res.FooterInfo.ContactInfo.Email, "; ")
				phones = strings.Join(res.FooterInfo.ContactInfo.Phone, "; ")
			}
			row = append(row, res.FooterInfo.Copyright, emails, phones)
		} else {
			row = append(row, "", "", "")
		}

		if res.CrawlStats != nil {
			row = append(row, res.CrawlStats.CrawlDuration.String())
		} else {
			row = append(row, "")
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// exportHTML 导出为HTML格式
func (bc *BatchCollector) exportHTML(result *BatchResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>网站信息收集报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin-bottom: 30px; }
        .stat-card { background: white; border: 1px solid #ddd; padding: 15px; border-radius: 5px; text-align: center; }
        .stat-number { font-size: 2em; font-weight: bold; color: #007bff; }
        .stat-label { color: #666; margin-top: 5px; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .success { color: green; }
        .error { color: red; }
        .download-links { max-width: 200px; overflow: hidden; text-overflow: ellipsis; }
    </style>
</head>
<body>
    <div class="header">
        <h1>网站信息收集报告</h1>
        <p>生成时间: ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
        <p>扫描耗时: ` + result.Duration.String() + `</p>
    </div>

    <div class="summary">
        <div class="stat-card">
            <div class="stat-number">` + fmt.Sprintf("%d", result.TotalURLs) + `</div>
            <div class="stat-label">总URL数</div>
        </div>
        <div class="stat-card">
            <div class="stat-number">` + fmt.Sprintf("%d", result.SuccessCount) + `</div>
            <div class="stat-label">成功</div>
        </div>
        <div class="stat-card">
            <div class="stat-number">` + fmt.Sprintf("%d", result.FailureCount) + `</div>
            <div class="stat-label">失败</div>
        </div>`

	if result.Summary != nil && result.Summary.ICPStatistics != nil {
		html += `
        <div class="stat-card">
            <div class="stat-number">` + fmt.Sprintf("%.1f%%", result.Summary.ICPStatistics.ICPPercentage) + `</div>
            <div class="stat-label">ICP备案率</div>
        </div>
        <div class="stat-card">
            <div class="stat-number">` + fmt.Sprintf("%.1f%%", result.Summary.ICPStatistics.PolicePercentage) + `</div>
            <div class="stat-label">网安备案率</div>
        </div>`
	}

	html += `
    </div>

    <h2>详细结果</h2>
    <table>
        <thead>
            <tr>
                <th>URL</th>
                <th>状态</th>
                <th>标题</th>
                <th>ICP备案</th>
                <th>网安备案</th>
                <th>组织</th>
                <th>服务器</th>
                <th>下载链接</th>
            </tr>
        </thead>
        <tbody>`

	for _, res := range result.Results {
		statusClass := "error"
		if res.Status == "success" {
			statusClass = "success"
		}

		title := ""
		if res.BasicInfo != nil {
			title = res.BasicInfo.Title
		}

		icp := ""
		police := ""
		org := ""
		if res.RegistrationInfo != nil {
			icp = res.RegistrationInfo.ICPLicense
			police = res.RegistrationInfo.PoliceRecord
			org = res.RegistrationInfo.Organization
		}

		server := ""
		if res.TechnicalInfo != nil {
			server = res.TechnicalInfo.Server
		}

		downloadCount := len(res.DownloadLinks)

		html += fmt.Sprintf(`
            <tr>
                <td><a href="%s" target="_blank">%s</a></td>
                <td class="%s">%s</td>
                <td>%s</td>
                <td>%s</td>
                <td>%s</td>
                <td>%s</td>
                <td>%s</td>
                <td>%d个文件</td>
            </tr>`,
			res.URL, res.URL, statusClass, res.Status, title, icp, police, org, server, downloadCount)
	}

	html += `
        </tbody>
    </table>
</body>
</html>`

	_, err = file.WriteString(html)
	return err
}