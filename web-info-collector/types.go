package main

import (
	"time"
)

// WebInfo 网站信息结构
type WebInfo struct {
	URL            string          `json:"url"`
	Timestamp      time.Time       `json:"timestamp"`
	Status         string          `json:"status"`
	Error          string          `json:"error,omitempty"`
	BasicInfo      *BasicInfo      `json:"basic_info,omitempty"`
	Icons          *IconInfo       `json:"icons,omitempty"`
	RegistrationInfo *RegistrationInfo `json:"registration_info,omitempty"`
	DownloadLinks  []DownloadLink  `json:"download_links,omitempty"`
	FooterInfo     *FooterInfo     `json:"footer_info,omitempty"`
	TechnicalInfo  *TechnicalInfo  `json:"technical_info,omitempty"`
	CrawlStats     *CrawlStats     `json:"crawl_stats,omitempty"`
}

// BasicInfo 基础信息
type BasicInfo struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Keywords    string `json:"keywords,omitempty"`
	Language    string `json:"language,omitempty"`
	Charset     string `json:"charset,omitempty"`
	Author      string `json:"author,omitempty"`
}

// IconInfo 图标信息
type IconInfo struct {
	Favicon         string `json:"favicon,omitempty"`
	AppleTouchIcon  string `json:"apple_touch_icon,omitempty"`
	Icons           []Icon `json:"icons,omitempty"`
}

// Icon 图标详细信息
type Icon struct {
	URL  string `json:"url"`
	Size string `json:"size,omitempty"`
	Type string `json:"type,omitempty"`
	Rel  string `json:"rel,omitempty"`
}

// RegistrationInfo 备案信息
type RegistrationInfo struct {
	ICPLicense   string `json:"icp_license,omitempty"`
	PoliceRecord string `json:"police_record,omitempty"`
	Organization string `json:"organization,omitempty"`
	ICPType      string `json:"icp_type,omitempty"`      // 企业/个人
	Province     string `json:"province,omitempty"`      // 备案省份
	RecordDate   string `json:"record_date,omitempty"`   // 备案日期
}

// DownloadLink 下载链接信息
type DownloadLink struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Type     string `json:"type,omitempty"`
	Size     string `json:"size,omitempty"`
	Title    string `json:"title,omitempty"`
	Context  string `json:"context,omitempty"` // 链接上下文
}

// FooterInfo 页脚信息
type FooterInfo struct {
	Copyright   string        `json:"copyright,omitempty"`
	ContactInfo *ContactInfo  `json:"contact_info,omitempty"`
	Links       []FooterLink  `json:"links,omitempty"`
	SocialMedia []SocialLink  `json:"social_media,omitempty"`
	RawText     string        `json:"raw_text,omitempty"`
}

// ContactInfo 联系信息
type ContactInfo struct {
	Email   []string `json:"email,omitempty"`
	Phone   []string `json:"phone,omitempty"`
	Address string   `json:"address,omitempty"`
	QQ      []string `json:"qq,omitempty"`
	WeChat  []string `json:"wechat,omitempty"`
}

// FooterLink 页脚链接
type FooterLink struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

// SocialLink 社交媒体链接
type SocialLink struct {
	Platform string `json:"platform"`
	URL      string `json:"url"`
	Username string `json:"username,omitempty"`
}

// TechnicalInfo 技术信息
type TechnicalInfo struct {
	Server      string   `json:"server,omitempty"`
	PoweredBy   string   `json:"powered_by,omitempty"`
	CMS         string   `json:"cms,omitempty"`
	Frameworks  []string `json:"frameworks,omitempty"`
	Analytics   []string `json:"analytics,omitempty"`
	CDN         string   `json:"cdn,omitempty"`
	StatusCode  int      `json:"status_code"`
	ContentType string   `json:"content_type,omitempty"`
}

// CrawlStats 爬取统计
type CrawlStats struct {
	PagesVisited   int           `json:"pages_visited"`
	TotalLinks     int           `json:"total_links"`
	DownloadLinks  int           `json:"download_links"`
	CrawlDuration  time.Duration `json:"crawl_duration"`
	MaxDepth       int           `json:"max_depth"`
	ErrorCount     int           `json:"error_count"`
}

// BatchResult 批量处理结果
type BatchResult struct {
	TotalURLs    int        `json:"total_urls"`
	SuccessCount int        `json:"success_count"`
	FailureCount int        `json:"failure_count"`
	Results      []WebInfo  `json:"results"`
	Summary      *Summary   `json:"summary,omitempty"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      time.Time  `json:"end_time"`
	Duration     time.Duration `json:"duration"`
}

// Summary 批量处理摘要
type Summary struct {
	CommonTechnologies []TechCount    `json:"common_technologies,omitempty"`
	ICPStatistics      *ICPStats      `json:"icp_statistics,omitempty"`
	FileTypeStats      []FileTypeCount `json:"file_type_stats,omitempty"`
	DomainStats        []DomainCount   `json:"domain_stats,omitempty"`
}

// TechCount 技术统计
type TechCount struct {
	Technology string `json:"technology"`
	Count      int    `json:"count"`
	Percentage float64 `json:"percentage"`
}

// ICPStats ICP统计
type ICPStats struct {
	TotalWithICP    int `json:"total_with_icp"`
	TotalWithPolice int `json:"total_with_police"`
	ICPPercentage   float64 `json:"icp_percentage"`
	PolicePercentage float64 `json:"police_percentage"`
}

// FileTypeCount 文件类型统计
type FileTypeCount struct {
	FileType string `json:"file_type"`
	Count    int    `json:"count"`
}

// DomainCount 域名统计
type DomainCount struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

// Config 配置信息
type Config struct {
	MaxDepth        int           `json:"max_depth"`
	MaxPages        int           `json:"max_pages"`
	Timeout         time.Duration `json:"timeout"`
	Concurrent      int           `json:"concurrent"`
	UserAgent       string        `json:"user_agent"`
	FollowRedirects bool          `json:"follow_redirects"`
	ExtractFiles    bool          `json:"extract_files"`
	ExtractFooter   bool          `json:"extract_footer"`
	ExtractIcons    bool          `json:"extract_icons"`
	Verbose         bool          `json:"verbose"`
	DelayBetweenRequests time.Duration `json:"delay_between_requests"`
}

// CrawlContext 爬取上下文
type CrawlContext struct {
	BaseURL     string
	VisitedURLs map[string]bool
	Queue       []CrawlItem
	Results     *WebInfo
	Config      *Config
	Stats       *CrawlStats
}

// CrawlItem 爬取项目
type CrawlItem struct {
	URL   string
	Depth int
}

// FileExtensions 常见文件扩展名
var FileExtensions = map[string]string{
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".zip":  "application/zip",
	".rar":  "application/x-rar-compressed",
	".7z":   "application/x-7z-compressed",
	".tar":  "application/x-tar",
	".gz":   "application/gzip",
	".mp3":  "audio/mpeg",
	".mp4":  "video/mp4",
	".avi":  "video/x-msvideo",
	".mov":  "video/quicktime",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".svg":  "image/svg+xml",
	".txt":  "text/plain",
	".csv":  "text/csv",
	".xml":  "application/xml",
	".json": "application/json",
	".exe":  "application/x-msdownload",
	".dmg":  "application/x-apple-diskimage",
	".apk":  "application/vnd.android.package-archive",
}

// SocialPlatforms 社交媒体平台识别
var SocialPlatforms = map[string]string{
	"weibo.com":     "微博",
	"qq.com":        "QQ",
	"wechat.com":    "微信",
	"facebook.com":  "Facebook",
	"twitter.com":   "Twitter",
	"linkedin.com":  "LinkedIn",
	"instagram.com": "Instagram",
	"youtube.com":   "YouTube",
	"github.com":    "GitHub",
	"zhihu.com":     "知乎",
	"douyin.com":    "抖音",
	"bilibili.com":  "哔哩哔哩",
}