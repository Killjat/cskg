package storage

import (
	"time"
)

// APIEndpoint API端点信息
type APIEndpoint struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	URL         string    `gorm:"size:500;not null;index" json:"url"`
	Method      string    `gorm:"size:10;not null" json:"method"`
	Path        string    `gorm:"size:500;not null;index" json:"path"`
	Domain      string    `gorm:"size:255;not null;index" json:"domain"`
	Type        string    `gorm:"size:50" json:"type"` // REST, GraphQL, WebSocket
	Status      int       `json:"status"`
	ContentType string    `gorm:"size:100" json:"content_type"`
	Parameters  string    `gorm:"type:text" json:"parameters"` // JSON格式存储
	Headers     string    `gorm:"type:text" json:"headers"`    // JSON格式存储
	Response    string    `gorm:"type:text" json:"response"`   // 响应示例
	Source      string    `gorm:"size:100" json:"source"`     // 发现来源：crawler, js, form
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CrawlSession 爬取会话
type CrawlSession struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	SessionID   string    `gorm:"size:100;not null;unique;index" json:"session_id"`
	TargetURL   string    `gorm:"size:500;not null" json:"target_url"`
	Status      string    `gorm:"size:20;not null" json:"status"` // running, completed, failed, paused
	StartTime   time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	PagesFound  int       `json:"pages_found"`
	APIsFound   int       `json:"apis_found"`
	Depth       int       `json:"depth"`
	Config      string    `gorm:"type:text" json:"config"` // JSON格式存储配置
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CrawledPage 已爬取页面
type CrawledPage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SessionID string    `gorm:"size:100;not null;index" json:"session_id"`
	URL       string    `gorm:"size:500;not null;index" json:"url"`
	Title     string    `gorm:"size:255" json:"title"`
	Status    int       `json:"status"`
	Depth     int       `json:"depth"`
	Size      int64     `json:"size"`
	Links     int       `json:"links"`     // 发现的链接数量
	APIs      int       `json:"apis"`      // 发现的API数量
	JSFiles   int       `json:"js_files"`  // JS文件数量
	Forms     int       `json:"forms"`     // 表单数量
	Error     string    `gorm:"type:text" json:"error"`
	CreatedAt time.Time `json:"created_at"`
}

// JSFile JavaScript文件信息
type JSFile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SessionID string    `gorm:"size:100;not null;index" json:"session_id"`
	URL       string    `gorm:"size:500;not null;index" json:"url"`
	Size      int64     `json:"size"`
	APIs      int       `json:"apis"`      // 发现的API数量
	Content   string    `gorm:"type:longtext" json:"content"` // JS文件内容
	Analyzed  bool      `gorm:"default:false" json:"analyzed"`
	CreatedAt time.Time `json:"created_at"`
}

// FormInfo 表单信息
type FormInfo struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SessionID string    `gorm:"size:100;not null;index" json:"session_id"`
	PageURL   string    `gorm:"size:500;not null" json:"page_url"`
	Action    string    `gorm:"size:500" json:"action"`
	Method    string    `gorm:"size:10" json:"method"`
	Fields    string    `gorm:"type:text" json:"fields"` // JSON格式存储字段信息
	CreatedAt time.Time `json:"created_at"`
}

// APIPattern API模式匹配规则
type APIPattern struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Pattern     string    `gorm:"size:500;not null" json:"pattern"`
	Type        string    `gorm:"size:50" json:"type"`
	Description string    `gorm:"type:text" json:"description"`
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ScanStatistics 扫描统计信息
type ScanStatistics struct {
	SessionID     string    `json:"session_id"`
	TotalPages    int       `json:"total_pages"`
	TotalAPIs     int       `json:"total_apis"`
	RESTAPIs      int       `json:"rest_apis"`
	GraphQLAPIs   int       `json:"graphql_apis"`
	WebSocketAPIs int       `json:"websocket_apis"`
	JSFiles       int       `json:"js_files"`
	Forms         int       `json:"forms"`
	Domains       []string  `json:"domains"`
	StartTime     time.Time `json:"start_time"`
	Duration      string    `json:"duration"`
}

// APIGroup API分组信息
type APIGroup struct {
	Path   string `json:"path"`
	Count  int    `json:"count"`
	APIs   []APIEndpoint `json:"apis"`
}

// DomainInfo 域名信息
type DomainInfo struct {
	Domain    string `json:"domain"`
	PageCount int    `json:"page_count"`
	APICount  int    `json:"api_count"`
}