package fingerprint

import (
	"encoding/base64"
	"regexp"
	"strings"
)

// Fingerprint 指纹识别结果
type Fingerprint struct {
	Product     string   // 产品名称
	Version     string   // 版本号
	Category    string   // 类别（Web服务器、应用服务器、数据库等）
	OS          string   // 操作系统
	DeviceType  string   // 设备类型
	CPE         string   // CPE标识
	Vendor      string   // 厂商
	Tags        []string // 标签
	Confidence  int      // 置信度 (0-100)
	RawBanner   string   // 原始Banner
	Description string   // 描述
}

// Rule 指纹识别规则
type Rule struct {
	Name       string
	Category   string
	Vendor     string
	Pattern    *regexp.Regexp
	Version    *regexp.Regexp
	Confidence int
	Tags       []string
}

var fingerprintRules = []*Rule{
	// Web服务器
	{
		Name:       "Nginx",
		Category:   "Web服务器",
		Vendor:     "Nginx Inc.",
		Pattern:    regexp.MustCompile(`(?i)nginx`),
		Version:    regexp.MustCompile(`nginx[/\s]+(\d+\.\d+\.\d+)`),
		Confidence: 95,
		Tags:       []string{"web", "http", "proxy"},
	},
	{
		Name:       "Apache",
		Category:   "Web服务器",
		Vendor:     "Apache Software Foundation",
		Pattern:    regexp.MustCompile(`(?i)apache`),
		Version:    regexp.MustCompile(`Apache[/\s]+(\d+\.\d+\.\d+)`),
		Confidence: 95,
		Tags:       []string{"web", "http"},
	},
	{
		Name:       "IIS",
		Category:   "Web服务器",
		Vendor:     "Microsoft",
		Pattern:    regexp.MustCompile(`(?i)Microsoft-IIS`),
		Version:    regexp.MustCompile(`Microsoft-IIS[/\s]+(\d+\.\d+)`),
		Confidence: 95,
		Tags:       []string{"web", "http", "windows"},
	},
	{
		Name:       "GHost",
		Category:   "Web服务器",
		Vendor:     "Unknown",
		Pattern:    regexp.MustCompile(`(?i)GHost`),
		Confidence: 90,
		Tags:       []string{"web", "http"},
	},
	{
		Name:       "Tomcat",
		Category:   "应用服务器",
		Vendor:     "Apache Software Foundation",
		Pattern:    regexp.MustCompile(`(?i)Apache-Coyote|Tomcat`),
		Version:    regexp.MustCompile(`Tomcat[/\s]+(\d+\.\d+\.\d+)`),
		Confidence: 90,
		Tags:       []string{"java", "application-server"},
	},
	{
		Name:       "Jetty",
		Category:   "应用服务器",
		Vendor:     "Eclipse Foundation",
		Pattern:    regexp.MustCompile(`(?i)Jetty`),
		Version:    regexp.MustCompile(`Jetty[/\s]+(\d+\.\d+)`),
		Confidence: 90,
		Tags:       []string{"java", "application-server"},
	},

	// CDN与负载均衡
	{
		Name:       "Cloudflare",
		Category:   "CDN",
		Vendor:     "Cloudflare Inc.",
		Pattern:    regexp.MustCompile(`(?i)cloudflare|cf-ray`),
		Confidence: 95,
		Tags:       []string{"cdn", "security", "proxy"},
	},
	{
		Name:       "Akamai",
		Category:   "CDN",
		Vendor:     "Akamai Technologies",
		Pattern:    regexp.MustCompile(`(?i)AkamaiGHost`),
		Confidence: 95,
		Tags:       []string{"cdn", "proxy"},
	},
	{
		Name:       "F5 BIG-IP",
		Category:   "负载均衡",
		Vendor:     "F5 Networks",
		Pattern:    regexp.MustCompile(`(?i)BIG-IP|F5`),
		Confidence: 90,
		Tags:       []string{"load-balancer", "proxy"},
	},

	// 编程语言与框架
	{
		Name:       "PHP",
		Category:   "编程语言",
		Vendor:     "PHP Group",
		Pattern:    regexp.MustCompile(`(?i)PHP[/\s]+|X-Powered-By.*PHP`),
		Version:    regexp.MustCompile(`PHP[/\s]+(\d+\.\d+\.\d+)`),
		Confidence: 90,
		Tags:       []string{"php", "language"},
	},
	{
		Name:       "ASP.NET",
		Category:   "Web框架",
		Vendor:     "Microsoft",
		Pattern:    regexp.MustCompile(`(?i)ASP\.NET|X-AspNet-Version`),
		Version:    regexp.MustCompile(`X-AspNet-Version:\s*(\d+\.\d+)`),
		Confidence: 90,
		Tags:       []string{"dotnet", "framework", "windows"},
	},
	{
		Name:       "Express",
		Category:   "Web框架",
		Vendor:     "OpenJS Foundation",
		Pattern:    regexp.MustCompile(`(?i)X-Powered-By.*Express`),
		Confidence: 85,
		Tags:       []string{"nodejs", "javascript", "framework"},
	},
	{
		Name:       "Django",
		Category:   "Web框架",
		Vendor:     "Django Software Foundation",
		Pattern:    regexp.MustCompile(`(?i)django|csrftoken`),
		Confidence: 80,
		Tags:       []string{"python", "framework"},
	},
	{
		Name:       "Flask",
		Category:   "Web框架",
		Vendor:     "Pallets",
		Pattern:    regexp.MustCompile(`(?i)Werkzeug`),
		Confidence: 80,
		Tags:       []string{"python", "framework"},
	},
	{
		Name:       "Spring Boot",
		Category:   "Web框架",
		Vendor:     "Pivotal/VMware",
		Pattern:    regexp.MustCompile(`(?i)Spring|X-Application-Context`),
		Confidence: 80,
		Tags:       []string{"java", "framework"},
	},

	// CMS内容管理系统
	{
		Name:       "WordPress",
		Category:   "CMS",
		Vendor:     "Automattic",
		Pattern:    regexp.MustCompile(`(?i)wp-content|wordpress|wp-includes`),
		Confidence: 85,
		Tags:       []string{"cms", "php", "blog"},
	},
	{
		Name:       "Joomla",
		Category:   "CMS",
		Vendor:     "Open Source Matters",
		Pattern:    regexp.MustCompile(`(?i)joomla|/components/com_`),
		Confidence: 85,
		Tags:       []string{"cms", "php"},
	},
	{
		Name:       "Drupal",
		Category:   "CMS",
		Vendor:     "Drupal Association",
		Pattern:    regexp.MustCompile(`(?i)drupal|X-Generator.*Drupal`),
		Confidence: 85,
		Tags:       []string{"cms", "php"},
	},

	// 数据库
	{
		Name:       "MySQL",
		Category:   "数据库",
		Vendor:     "Oracle",
		Pattern:    regexp.MustCompile(`(?i)mysql`),
		Version:    regexp.MustCompile(`(\d+\.\d+\.\d+).*mysql`),
		Confidence: 95,
		Tags:       []string{"database", "sql"},
	},
	{
		Name:       "PostgreSQL",
		Category:   "数据库",
		Vendor:     "PostgreSQL Global Development Group",
		Pattern:    regexp.MustCompile(`(?i)postgres`),
		Confidence: 95,
		Tags:       []string{"database", "sql"},
	},
	{
		Name:       "Redis",
		Category:   "缓存数据库",
		Vendor:     "Redis Ltd.",
		Pattern:    regexp.MustCompile(`(?i)\$\d+\r\n|redis_version|REDIS`),
		Version:    regexp.MustCompile(`redis_version:(\d+\.\d+\.\d+)`),
		Confidence: 95,
		Tags:       []string{"database", "cache", "nosql"},
	},
	{
		Name:       "MongoDB",
		Category:   "数据库",
		Vendor:     "MongoDB Inc.",
		Pattern:    regexp.MustCompile(`(?i)mongodb|mongo`),
		Confidence: 90,
		Tags:       []string{"database", "nosql"},
	},
	{
		Name:       "Elasticsearch",
		Category:   "搜索引擎",
		Vendor:     "Elastic",
		Pattern:    regexp.MustCompile(`(?i)elasticsearch`),
		Version:    regexp.MustCompile(`"number"\s*:\s*"(\d+\.\d+\.\d+)"`),
		Confidence: 95,
		Tags:       []string{"search", "database", "analytics"},
	},

	// SSH服务
	{
		Name:       "OpenSSH",
		Category:   "SSH服务",
		Vendor:     "OpenBSD",
		Pattern:    regexp.MustCompile(`(?i)SSH-.*OpenSSH`),
		Version:    regexp.MustCompile(`OpenSSH[_\s]+(\d+\.\d+)`),
		Confidence: 95,
		Tags:       []string{"ssh", "remote-access"},
	},

	// FTP服务
	{
		Name:       "vsftpd",
		Category:   "FTP服务",
		Vendor:     "Chris Evans",
		Pattern:    regexp.MustCompile(`(?i)vsftpd`),
		Version:    regexp.MustCompile(`vsftpd\s+(\d+\.\d+\.\d+)`),
		Confidence: 95,
		Tags:       []string{"ftp", "file-transfer"},
	},
	{
		Name:       "ProFTPD",
		Category:   "FTP服务",
		Vendor:     "ProFTPD Project",
		Pattern:    regexp.MustCompile(`(?i)ProFTPD`),
		Version:    regexp.MustCompile(`ProFTPD\s+(\d+\.\d+\.\d+)`),
		Confidence: 95,
		Tags:       []string{"ftp", "file-transfer"},
	},

	// 邮件服务
	{
		Name:       "Postfix",
		Category:   "邮件服务",
		Vendor:     "Wietse Venema",
		Pattern:    regexp.MustCompile(`(?i)220.*Postfix`),
		Confidence: 95,
		Tags:       []string{"smtp", "email"},
	},
	{
		Name:       "Exim",
		Category:   "邮件服务",
		Vendor:     "Exim Development",
		Pattern:    regexp.MustCompile(`(?i)220.*Exim`),
		Confidence: 95,
		Tags:       []string{"smtp", "email"},
	},

	// 操作系统
	{
		Name:       "Ubuntu",
		Category:   "操作系统",
		Vendor:     "Canonical",
		Pattern:    regexp.MustCompile(`(?i)Ubuntu`),
		Confidence: 70,
		Tags:       []string{"linux", "os"},
	},
	{
		Name:       "CentOS",
		Category:   "操作系统",
		Vendor:     "Red Hat",
		Pattern:    regexp.MustCompile(`(?i)CentOS`),
		Confidence: 70,
		Tags:       []string{"linux", "os"},
	},
	{
		Name:       "Debian",
		Category:   "操作系统",
		Vendor:     "Debian Project",
		Pattern:    regexp.MustCompile(`(?i)Debian`),
		Confidence: 70,
		Tags:       []string{"linux", "os"},
	},
	{
		Name:       "Windows Server",
		Category:   "操作系统",
		Vendor:     "Microsoft",
		Pattern:    regexp.MustCompile(`(?i)Win32|Windows|Microsoft-IIS`),
		Confidence: 70,
		Tags:       []string{"windows", "os"},
	},

	// 安全设备
	{
		Name:       "WAF",
		Category:   "安全设备",
		Vendor:     "Various",
		Pattern:    regexp.MustCompile(`(?i)X-WAF|WebKnight|NAXSI|ModSecurity`),
		Confidence: 80,
		Tags:       []string{"security", "waf"},
	},
}

// Identify 识别回包指纹
func Identify(banner string, response []byte) []*Fingerprint {
	var results []*Fingerprint
	
	// 解码Base64响应（如果需要）
	var decodedResponse string
	if response != nil && len(response) > 0 {
		decoded, err := base64.StdEncoding.DecodeString(string(response))
		if err == nil {
			decodedResponse = string(decoded)
		}
	}
	
	// 合并Banner和响应内容
	fullContent := banner
	if decodedResponse != "" {
		fullContent = banner + "\n" + decodedResponse
	}
	
	if fullContent == "" {
		return results
	}
	
	// 遍历所有规则进行匹配
	for _, rule := range fingerprintRules {
		if rule.Pattern.MatchString(fullContent) {
			fp := &Fingerprint{
				Product:     rule.Name,
				Category:    rule.Category,
				Vendor:      rule.Vendor,
				Tags:        rule.Tags,
				Confidence:  rule.Confidence,
				RawBanner:   banner,
				Description: rule.Category + " - " + rule.Name,
			}
			
			// 提取版本号
			if rule.Version != nil {
				if matches := rule.Version.FindStringSubmatch(fullContent); len(matches) > 1 {
					fp.Version = matches[1]
				}
			}
			
			// 推断操作系统
			fp.OS = inferOS(fullContent)
			
			// 生成CPE
			fp.CPE = generateCPE(fp)
			
			results = append(results, fp)
		}
	}
	
	return results
}

// IdentifyQuick 快速识别（仅使用Banner）
func IdentifyQuick(banner string) []*Fingerprint {
	return Identify(banner, nil)
}

// inferOS 推断操作系统
func inferOS(content string) string {
	lowerContent := strings.ToLower(content)
	
	if strings.Contains(lowerContent, "ubuntu") {
		return "Linux/Ubuntu"
	}
	if strings.Contains(lowerContent, "centos") {
		return "Linux/CentOS"
	}
	if strings.Contains(lowerContent, "debian") {
		return "Linux/Debian"
	}
	if strings.Contains(lowerContent, "redhat") || strings.Contains(lowerContent, "rhel") {
		return "Linux/Red Hat"
	}
	if strings.Contains(lowerContent, "fedora") {
		return "Linux/Fedora"
	}
	if strings.Contains(lowerContent, "windows") || strings.Contains(lowerContent, "win32") || 
	   strings.Contains(lowerContent, "microsoft-iis") {
		return "Windows"
	}
	if strings.Contains(lowerContent, "freebsd") {
		return "FreeBSD"
	}
	if strings.Contains(lowerContent, "openbsd") {
		return "OpenBSD"
	}
	if strings.Contains(lowerContent, "netbsd") {
		return "NetBSD"
	}
	if strings.Contains(lowerContent, "unix") {
		return "Unix"
	}
	
	return "Unknown"
}

// generateCPE 生成CPE标识
func generateCPE(fp *Fingerprint) string {
	if fp.Product == "" {
		return ""
	}
	
	vendor := strings.ToLower(strings.ReplaceAll(fp.Vendor, " ", "_"))
	product := strings.ToLower(strings.ReplaceAll(fp.Product, " ", "_"))
	version := fp.Version
	
	if vendor == "" {
		vendor = "*"
	}
	if version == "" {
		version = "*"
	}
	
	return "cpe:/a:" + vendor + ":" + product + ":" + version
}

// GetFingerprints 批量识别
func GetFingerprints(banner string, response []byte) map[string]*Fingerprint {
	fps := Identify(banner, response)
	result := make(map[string]*Fingerprint)
	
	for _, fp := range fps {
		// 使用产品名作为键，避免重复
		key := fp.Product
		if existing, ok := result[key]; !ok || fp.Confidence > existing.Confidence {
			result[key] = fp
		}
	}
	
	return result
}

// GetTopFingerprint 获取最高置信度的指纹
func GetTopFingerprint(banner string, response []byte) *Fingerprint {
	fps := Identify(banner, response)
	
	if len(fps) == 0 {
		return nil
	}
	
	// 返回置信度最高的
	top := fps[0]
	for _, fp := range fps[1:] {
		if fp.Confidence > top.Confidence {
			top = fp
		}
	}
	
	return top
}

// GetCategories 获取所有识别到的类别
func GetCategories(banner string, response []byte) []string {
	fps := Identify(banner, response)
	categoryMap := make(map[string]bool)
	
	for _, fp := range fps {
		categoryMap[fp.Category] = true
	}
	
	var categories []string
	for category := range categoryMap {
		categories = append(categories, category)
	}
	
	return categories
}

// HasTag 检查是否包含特定标签
func HasTag(banner string, response []byte, tag string) bool {
	fps := Identify(banner, response)
	
	for _, fp := range fps {
		for _, t := range fp.Tags {
			if strings.EqualFold(t, tag) {
				return true
			}
		}
	}
	
	return false
}
