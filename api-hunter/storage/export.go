package storage

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ExportFormat 导出格式
type ExportFormat string

const (
	FormatJSON     ExportFormat = "json"
	FormatCSV      ExportFormat = "csv"
	FormatMarkdown ExportFormat = "markdown"
	FormatHTML     ExportFormat = "html"
)

// ExportOptions 导出选项
type ExportOptions struct {
	Format         ExportFormat `json:"format"`
	OutputPath     string       `json:"output_path"`
	SessionID      string       `json:"session_id"`
	IncludeDetails bool         `json:"include_details"`
}

// ExportResult 导出结果
type ExportResult struct {
	FilePath    string `json:"file_path"`
	RecordCount int    `json:"record_count"`
	FileSize    int64  `json:"file_size"`
}

// Exporter 导出器
type Exporter struct {
	db *Database
}

// NewExporter 创建导出器
func NewExporter(db *Database) *Exporter {
	return &Exporter{db: db}
}

// Export 导出数据
func (e *Exporter) Export(options ExportOptions) (*ExportResult, error) {
	// 获取数据
	apis, err := e.db.GetAPIEndpoints(options.SessionID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("获取API数据失败: %v", err)
	}

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(options.OutputPath), 0755); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %v", err)
	}

	var filePath string
	switch options.Format {
	case FormatJSON:
		filePath, err = e.exportJSON(apis, options)
	case FormatCSV:
		filePath, err = e.exportCSV(apis, options)
	case FormatMarkdown:
		filePath, err = e.exportMarkdown(apis, options)
	case FormatHTML:
		filePath, err = e.exportHTML(apis, options)
	default:
		return nil, fmt.Errorf("不支持的导出格式: %s", options.Format)
	}

	if err != nil {
		return nil, err
	}

	// 获取文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &ExportResult{
		FilePath:    filePath,
		RecordCount: len(apis),
		FileSize:    fileInfo.Size(),
	}, nil
}

// exportJSON 导出为JSON格式
func (e *Exporter) exportJSON(apis []APIEndpoint, options ExportOptions) (string, error) {
	filePath := options.OutputPath
	if !strings.HasSuffix(filePath, ".json") {
		filePath += ".json"
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 创建导出数据结构
	exportData := map[string]interface{}{
		"session_id":   options.SessionID,
		"export_time":  time.Now().Format("2006-01-02 15:04:05"),
		"total_count":  len(apis),
		"apis":         apis,
	}

	if options.IncludeDetails {
		// 添加统计信息
		stats, _ := e.db.GetStatistics(options.SessionID)
		exportData["statistics"] = stats
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return filePath, encoder.Encode(exportData)
}

// exportCSV 导出为CSV格式
func (e *Exporter) exportCSV(apis []APIEndpoint, options ExportOptions) (string, error) {
	filePath := options.OutputPath
	if !strings.HasSuffix(filePath, ".csv") {
		filePath += ".csv"
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	headers := []string{"ID", "URL", "Method", "Path", "Domain", "Type", "Status", "Content-Type", "Source", "Created At"}
	if err := writer.Write(headers); err != nil {
		return "", err
	}

	// 写入数据
	for _, api := range apis {
		record := []string{
			fmt.Sprintf("%d", api.ID),
			api.URL,
			api.Method,
			api.Path,
			api.Domain,
			api.Type,
			fmt.Sprintf("%d", api.Status),
			api.ContentType,
			api.Source,
			api.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}

	return filePath, nil
}

// exportMarkdown 导出为Markdown格式
func (e *Exporter) exportMarkdown(apis []APIEndpoint, options ExportOptions) (string, error) {
	filePath := options.OutputPath
	if !strings.HasSuffix(filePath, ".md") {
		filePath += ".md"
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 写入标题和基本信息
	fmt.Fprintf(file, "# API Discovery Report\n\n")
	fmt.Fprintf(file, "**Session ID:** %s\n", options.SessionID)
	fmt.Fprintf(file, "**Export Time:** %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "**Total APIs:** %d\n\n", len(apis))

	if options.IncludeDetails {
		// 添加统计信息
		stats, err := e.db.GetStatistics(options.SessionID)
		if err == nil {
			fmt.Fprintf(file, "## Statistics\n\n")
			fmt.Fprintf(file, "- **Total Pages:** %d\n", stats.TotalPages)
			fmt.Fprintf(file, "- **REST APIs:** %d\n", stats.RESTAPIs)
			fmt.Fprintf(file, "- **GraphQL APIs:** %d\n", stats.GraphQLAPIs)
			fmt.Fprintf(file, "- **WebSocket APIs:** %d\n", stats.WebSocketAPIs)
			fmt.Fprintf(file, "- **JavaScript Files:** %d\n", stats.JSFiles)
			fmt.Fprintf(file, "- **Forms:** %d\n", stats.Forms)
			fmt.Fprintf(file, "- **Domains:** %s\n\n", strings.Join(stats.Domains, ", "))
		}
	}

	// 按域名分组
	domainAPIs := make(map[string][]APIEndpoint)
	for _, api := range apis {
		domainAPIs[api.Domain] = append(domainAPIs[api.Domain], api)
	}

	// 写入API列表
	fmt.Fprintf(file, "## API Endpoints\n\n")
	for domain, domainAPIList := range domainAPIs {
		fmt.Fprintf(file, "### %s\n\n", domain)
		fmt.Fprintf(file, "| Method | Path | Type | Status | Source |\n")
		fmt.Fprintf(file, "|--------|------|------|--------|--------|\n")
		
		for _, api := range domainAPIList {
			fmt.Fprintf(file, "| %s | %s | %s | %d | %s |\n",
				api.Method, api.Path, api.Type, api.Status, api.Source)
		}
		fmt.Fprintf(file, "\n")
	}

	return filePath, nil
}

// exportHTML 导出为HTML格式
func (e *Exporter) exportHTML(apis []APIEndpoint, options ExportOptions) (string, error) {
	filePath := options.OutputPath
	if !strings.HasSuffix(filePath, ".html") {
		filePath += ".html"
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// HTML模板
	htmlTemplate := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API Discovery Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .stats { display: flex; gap: 20px; margin-bottom: 20px; }
        .stat-card { background: #e3f2fd; padding: 15px; border-radius: 5px; text-align: center; }
        .stat-number { font-size: 24px; font-weight: bold; color: #1976d2; }
        .stat-label { color: #666; }
        table { width: 100%; border-collapse: collapse; margin-bottom: 20px; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f5f5f5; font-weight: bold; }
        .method { padding: 4px 8px; border-radius: 3px; color: white; font-size: 12px; }
        .method-GET { background-color: #4caf50; }
        .method-POST { background-color: #ff9800; }
        .method-PUT { background-color: #2196f3; }
        .method-DELETE { background-color: #f44336; }
        .method-PATCH { background-color: #9c27b0; }
        .type-REST { color: #4caf50; }
        .type-GraphQL { color: #e91e63; }
        .type-WebSocket { color: #ff9800; }
        .domain-section { margin-bottom: 30px; }
        .domain-title { color: #1976d2; border-bottom: 2px solid #1976d2; padding-bottom: 5px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>API Discovery Report</h1>
        <p><strong>Session ID:</strong> {{.SessionID}}</p>
        <p><strong>Export Time:</strong> {{.ExportTime}}</p>
        <p><strong>Total APIs:</strong> {{.TotalAPIs}}</p>
    </div>

    {{if .Statistics}}
    <div class="stats">
        <div class="stat-card">
            <div class="stat-number">{{.Statistics.TotalPages}}</div>
            <div class="stat-label">Pages</div>
        </div>
        <div class="stat-card">
            <div class="stat-number">{{.Statistics.RESTAPIs}}</div>
            <div class="stat-label">REST APIs</div>
        </div>
        <div class="stat-card">
            <div class="stat-number">{{.Statistics.GraphQLAPIs}}</div>
            <div class="stat-label">GraphQL APIs</div>
        </div>
        <div class="stat-card">
            <div class="stat-number">{{.Statistics.WebSocketAPIs}}</div>
            <div class="stat-label">WebSocket APIs</div>
        </div>
        <div class="stat-card">
            <div class="stat-number">{{.Statistics.JSFiles}}</div>
            <div class="stat-label">JS Files</div>
        </div>
    </div>
    {{end}}

    {{range $domain, $apis := .DomainAPIs}}
    <div class="domain-section">
        <h2 class="domain-title">{{$domain}}</h2>
        <table>
            <thead>
                <tr>
                    <th>Method</th>
                    <th>Path</th>
                    <th>Type</th>
                    <th>Status</th>
                    <th>Source</th>
                    <th>Created At</th>
                </tr>
            </thead>
            <tbody>
                {{range $apis}}
                <tr>
                    <td><span class="method method-{{.Method}}">{{.Method}}</span></td>
                    <td>{{.Path}}</td>
                    <td><span class="type-{{.Type}}">{{.Type}}</span></td>
                    <td>{{.Status}}</td>
                    <td>{{.Source}}</td>
                    <td>{{.CreatedAt.Format "2006-01-02 15:04:05"}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>
    {{end}}
</body>
</html>`

	// 准备模板数据
	domainAPIs := make(map[string][]APIEndpoint)
	for _, api := range apis {
		domainAPIs[api.Domain] = append(domainAPIs[api.Domain], api)
	}

	data := map[string]interface{}{
		"SessionID":   options.SessionID,
		"ExportTime":  time.Now().Format("2006-01-02 15:04:05"),
		"TotalAPIs":   len(apis),
		"DomainAPIs":  domainAPIs,
	}

	if options.IncludeDetails {
		stats, err := e.db.GetStatistics(options.SessionID)
		if err == nil {
			data["Statistics"] = stats
		}
	}

	// 解析并执行模板
	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	return filePath, tmpl.Execute(file, data)
}