package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// WebInfoCollector 网站信息收集器
type WebInfoCollector struct {
	config *Config
	client *http.Client
}

// NewWebInfoCollector 创建新的收集器
func NewWebInfoCollector(config *Config) *WebInfoCollector {
	client := &http.Client{
		Timeout: config.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !config.FollowRedirects {
				return http.ErrUseLastResponse
			}
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return nil
		},
	}

	return &WebInfoCollector{
		config: config,
		client: client,
	}
}

// CollectWebInfo 收集网站信息
func (wic *WebInfoCollector) CollectWebInfo(targetURL string) *WebInfo {
	result := &WebInfo{
		URL:       targetURL,
		Timestamp: time.Now(),
		Status:    "success",
		CrawlStats: &CrawlStats{
			MaxDepth: wic.config.MaxDepth,
		},
	}

	startTime := time.Now()

	// 创建爬取上下文
	ctx := &CrawlContext{
		BaseURL:     targetURL,
		VisitedURLs: make(map[string]bool),
		Queue:       []CrawlItem{{URL: targetURL, Depth: 0}},
		Results:     result,
		Config:      wic.config,
		Stats:       result.CrawlStats,
	}

	// 开始爬取
	err := wic.crawlWebsite(ctx)
	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		return result
	}

	result.CrawlStats.CrawlDuration = time.Since(startTime)
	return result
}

// crawlWebsite 爬取网站
func (wic *WebInfoCollector) crawlWebsite(ctx *CrawlContext) error {
	for len(ctx.Queue) > 0 && ctx.Stats.PagesVisited < wic.config.MaxPages {
		// 取出队列中的第一个URL
		item := ctx.Queue[0]
		ctx.Queue = ctx.Queue[1:]

		// 检查是否已访问
		if ctx.VisitedURLs[item.URL] {
			continue
		}

		// 检查深度限制
		if item.Depth > wic.config.MaxDepth {
			continue
		}

		// 访问页面
		err := wic.processPage(ctx, item)
		if err != nil {
			ctx.Stats.ErrorCount++
			if wic.config.Verbose {
				fmt.Printf("Error processing %s: %v\n", item.URL, err)
			}
			continue
		}

		ctx.VisitedURLs[item.URL] = true
		ctx.Stats.PagesVisited++

		// 添加延迟
		if wic.config.DelayBetweenRequests > 0 {
			time.Sleep(wic.config.DelayBetweenRequests)
		}
	}

	return nil
}

// processPage 处理单个页面
func (wic *WebInfoCollector) processPage(ctx *CrawlContext, item CrawlItem) error {
	// 发送HTTP请求
	req, err := http.NewRequest("GET", item.URL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", wic.config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := wic.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 解析HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return err
	}

	// 如果是第一个页面（主页），提取基础信息
	if item.Depth == 0 {
		wic.extractBasicInfo(ctx.Results, doc, resp)
		if wic.config.ExtractIcons {
			wic.extractIconInfo(ctx.Results, doc, item.URL)
		}
		wic.extractRegistrationInfo(ctx.Results, doc, string(body))
		wic.extractTechnicalInfo(ctx.Results, resp, doc)
		if wic.config.ExtractFooter {
			wic.extractFooterInfo(ctx.Results, doc)
		}
	}

	// 提取文件下载链接
	if wic.config.ExtractFiles {
		wic.extractDownloadLinks(ctx.Results, doc, item.URL)
	}

	// 如果需要深度爬取，添加新的链接到队列
	if item.Depth < wic.config.MaxDepth {
		wic.addLinksToQueue(ctx, doc, item.URL, item.Depth+1)
	}

	return nil
}

// extractBasicInfo 提取基础信息
func (wic *WebInfoCollector) extractBasicInfo(result *WebInfo, doc *goquery.Document, resp *http.Response) {
	result.BasicInfo = &BasicInfo{}

	// 提取标题
	result.BasicInfo.Title = strings.TrimSpace(doc.Find("title").Text())

	// 提取meta信息
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		property, _ := s.Attr("property")
		content, _ := s.Attr("content")

		switch strings.ToLower(name) {
		case "description":
			result.BasicInfo.Description = content
		case "keywords":
			result.BasicInfo.Keywords = content
		case "author":
			result.BasicInfo.Author = content
		}

		switch strings.ToLower(property) {
		case "og:description":
			if result.BasicInfo.Description == "" {
				result.BasicInfo.Description = content
			}
		}

		// 提取字符集
		if charset, exists := s.Attr("charset"); exists {
			result.BasicInfo.Charset = charset
		}
	})

	// 提取语言
	if lang, exists := doc.Find("html").Attr("lang"); exists {
		result.BasicInfo.Language = lang
	}
}

// extractIconInfo 提取图标信息
func (wic *WebInfoCollector) extractIconInfo(result *WebInfo, doc *goquery.Document, baseURL string) {
	result.Icons = &IconInfo{
		Icons: make([]Icon, 0),
	}

	// 解析基础URL
	base, _ := url.Parse(baseURL)

	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		rel, _ := s.Attr("rel")
		href, _ := s.Attr("href")
		iconType, _ := s.Attr("type")
		sizes, _ := s.Attr("sizes")

		if href == "" {
			return
		}

		// 转换为绝对URL
		iconURL := wic.resolveURL(base, href)

		switch strings.ToLower(rel) {
		case "icon", "shortcut icon":
			if result.Icons.Favicon == "" {
				result.Icons.Favicon = iconURL
			}
		case "apple-touch-icon", "apple-touch-icon-precomposed":
			if result.Icons.AppleTouchIcon == "" {
				result.Icons.AppleTouchIcon = iconURL
			}
		}

		// 收集所有图标
		if strings.Contains(strings.ToLower(rel), "icon") {
			icon := Icon{
				URL:  iconURL,
				Size: sizes,
				Type: iconType,
				Rel:  rel,
			}
			result.Icons.Icons = append(result.Icons.Icons, icon)
		}
	})

	// 如果没有找到favicon，尝试默认路径
	if result.Icons.Favicon == "" {
		defaultFavicon := base.Scheme + "://" + base.Host + "/favicon.ico"
		result.Icons.Favicon = defaultFavicon
	}
}

// extractRegistrationInfo 提取备案信息
func (wic *WebInfoCollector) extractRegistrationInfo(result *WebInfo, doc *goquery.Document, htmlContent string) {
	result.RegistrationInfo = &RegistrationInfo{}

	// ICP备案号正则表达式
	icpPatterns := []string{
		`([京津沪渝冀豫云辽黑湘皖鲁新苏浙赣鄂桂甘晋蒙陕吉闽贵粤青藏川宁琼使领]ICP备\d+号(?:-\d+)??)`,
		`(ICP备\d+号(?:-\d+)?)`,
		`(备案号[：:]\s*([京津沪渝冀豫云辽黑湘皖鲁新苏浙赣鄂桂甘晋蒙陕吉闽贵粤青藏川宁琼使领]?ICP备\d+号(?:-\d+)?))`,
	}

	// 网安备案号正则表达式
	policePatterns := []string{
		`([京津沪渝冀豫云辽黑湘皖鲁新苏浙赣鄂桂甘晋蒙陕吉闽贵粤青藏川宁琼]公网安备\d+号)`,
		`(公网安备\d+号)`,
		`(\d{11}号)`, // 简化的公安备案号
	}

	// 在HTML内容中搜索ICP备案信息
	for _, pattern := range icpPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(htmlContent)
		if len(matches) > 1 {
			result.RegistrationInfo.ICPLicense = strings.TrimSpace(matches[1])
			break
		}
	}

	// 在HTML内容中搜索网安备案信息
	for _, pattern := range policePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(htmlContent)
		if len(matches) > 1 {
			result.RegistrationInfo.PoliceRecord = strings.TrimSpace(matches[1])
			break
		}
	}

	// 尝试从页脚链接中提取备案信息
	doc.Find("footer, .footer, #footer").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		
		// 搜索ICP备案
		for _, pattern := range icpPatterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(text)
			if len(matches) > 1 && result.RegistrationInfo.ICPLicense == "" {
				result.RegistrationInfo.ICPLicense = strings.TrimSpace(matches[1])
			}
		}

		// 搜索网安备案
		for _, pattern := range policePatterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(text)
			if len(matches) > 1 && result.RegistrationInfo.PoliceRecord == "" {
				result.RegistrationInfo.PoliceRecord = strings.TrimSpace(matches[1])
			}
		}
	})

	// 提取组织信息
	orgPatterns := []string{
		`版权所有[：:]\s*([^©\n\r]+)`,
		`©\s*\d{4}\s*([^©\n\r\.]+)`,
		`Copyright.*?(\d{4}).*?([^©\n\r\.]+)`,
	}

	for _, pattern := range orgPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(htmlContent)
		if len(matches) > 1 {
			org := strings.TrimSpace(matches[len(matches)-1])
			if len(org) > 0 && len(org) < 100 {
				result.RegistrationInfo.Organization = org
				break
			}
		}
	}
}

// extractDownloadLinks 提取下载链接
func (wic *WebInfoCollector) extractDownloadLinks(result *WebInfo, doc *goquery.Document, baseURL string) {
	base, _ := url.Parse(baseURL)
	
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href == "" {
			return
		}

		// 转换为绝对URL
		linkURL := wic.resolveURL(base, href)
		
		// 检查是否是文件下载链接
		if wic.isDownloadLink(linkURL) {
			filename := wic.extractFilename(linkURL)
			fileType := wic.getFileType(filename)
			
			downloadLink := DownloadLink{
				URL:      linkURL,
				Filename: filename,
				Type:     fileType,
				Title:    strings.TrimSpace(s.Text()),
				Context:  wic.getContext(s),
			}

			result.DownloadLinks = append(result.DownloadLinks, downloadLink)
		}
	})

	result.CrawlStats.DownloadLinks = len(result.DownloadLinks)
}

// extractFooterInfo 提取页脚信息
func (wic *WebInfoCollector) extractFooterInfo(result *WebInfo, doc *goquery.Document) {
	result.FooterInfo = &FooterInfo{
		Links:       make([]FooterLink, 0),
		SocialMedia: make([]SocialLink, 0),
		ContactInfo: &ContactInfo{
			Email: make([]string, 0),
			Phone: make([]string, 0),
			QQ:    make([]string, 0),
			WeChat: make([]string, 0),
		},
	}

	// 查找页脚元素
	footerSelectors := []string{"footer", ".footer", "#footer", ".site-footer", ".page-footer"}
	var footerElement *goquery.Selection

	for _, selector := range footerSelectors {
		element := doc.Find(selector)
		if element.Length() > 0 {
			footerElement = element.First()
			break
		}
	}

	if footerElement == nil {
		return
	}

	// 提取原始文本
	result.FooterInfo.RawText = strings.TrimSpace(footerElement.Text())

	// 提取版权信息
	copyrightPatterns := []string{
		`©.*?\d{4}.*?[^\n\r]*`,
		`Copyright.*?\d{4}.*?[^\n\r]*`,
		`版权所有.*?[^\n\r]*`,
	}

	for _, pattern := range copyrightPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindString(result.FooterInfo.RawText)
		if matches != "" {
			result.FooterInfo.Copyright = strings.TrimSpace(matches)
			break
		}
	}

	// 提取联系信息
	wic.extractContactInfo(result.FooterInfo.ContactInfo, result.FooterInfo.RawText)

	// 提取页脚链接
	footerElement.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		text := strings.TrimSpace(s.Text())
		
		if href != "" && text != "" {
			// 检查是否是社交媒体链接
			if platform := wic.getSocialPlatform(href); platform != "" {
				socialLink := SocialLink{
					Platform: platform,
					URL:      href,
					Username: wic.extractUsername(href),
				}
				result.FooterInfo.SocialMedia = append(result.FooterInfo.SocialMedia, socialLink)
			} else {
				footerLink := FooterLink{
					Text: text,
					URL:  href,
				}
				result.FooterInfo.Links = append(result.FooterInfo.Links, footerLink)
			}
		}
	})
}

// extractTechnicalInfo 提取技术信息
func (wic *WebInfoCollector) extractTechnicalInfo(result *WebInfo, resp *http.Response, doc *goquery.Document) {
	result.TechnicalInfo = &TechnicalInfo{
		StatusCode:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		Server:      resp.Header.Get("Server"),
		PoweredBy:   resp.Header.Get("X-Powered-By"),
		Frameworks:  make([]string, 0),
		Analytics:   make([]string, 0),
	}

	// 检测CDN
	if cdnHeader := resp.Header.Get("X-Cache"); cdnHeader != "" {
		result.TechnicalInfo.CDN = cdnHeader
	}

	// 检测框架和库
	wic.detectFrameworks(result.TechnicalInfo, doc)
	
	// 检测分析工具
	wic.detectAnalytics(result.TechnicalInfo, doc)
	
	// 检测CMS
	wic.detectCMS(result.TechnicalInfo, doc, resp)
}

// 辅助函数
func (wic *WebInfoCollector) resolveURL(base *url.URL, href string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	
	resolved, err := base.Parse(href)
	if err != nil {
		return href
	}
	
	return resolved.String()
}

func (wic *WebInfoCollector) isDownloadLink(urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	
	ext := strings.ToLower(path.Ext(parsedURL.Path))
	_, exists := FileExtensions[ext]
	return exists
}

func (wic *WebInfoCollector) extractFilename(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	
	return path.Base(parsedURL.Path)
}

func (wic *WebInfoCollector) getFileType(filename string) string {
	ext := strings.ToLower(path.Ext(filename))
	if fileType, exists := FileExtensions[ext]; exists {
		return fileType
	}
	return "application/octet-stream"
}

func (wic *WebInfoCollector) getContext(s *goquery.Selection) string {
	parent := s.Parent()
	if parent.Length() > 0 {
		return strings.TrimSpace(parent.Text())
	}
	return ""
}

func (wic *WebInfoCollector) extractContactInfo(contactInfo *ContactInfo, text string) {
	// 提取邮箱
	emailPattern := `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`
	emailRe := regexp.MustCompile(emailPattern)
	emails := emailRe.FindAllString(text, -1)
	contactInfo.Email = append(contactInfo.Email, emails...)

	// 提取电话
	phonePatterns := []string{
		`\d{3,4}-\d{7,8}`,           // 固定电话
		`1[3-9]\d{9}`,               // 手机号
		`\+86\s*1[3-9]\d{9}`,        // 带国际区号的手机
		`\(\d{3,4}\)\s*\d{7,8}`,     // 带括号的区号
	}
	
	for _, pattern := range phonePatterns {
		phoneRe := regexp.MustCompile(pattern)
		phones := phoneRe.FindAllString(text, -1)
		contactInfo.Phone = append(contactInfo.Phone, phones...)
	}

	// 提取QQ号
	qqPattern := `QQ[：:]\s*(\d{5,12})`
	qqRe := regexp.MustCompile(qqPattern)
	qqMatches := qqRe.FindAllStringSubmatch(text, -1)
	for _, match := range qqMatches {
		if len(match) > 1 {
			contactInfo.QQ = append(contactInfo.QQ, match[1])
		}
	}
}

func (wic *WebInfoCollector) getSocialPlatform(urlStr string) string {
	for domain, platform := range SocialPlatforms {
		if strings.Contains(urlStr, domain) {
			return platform
		}
	}
	return ""
}

func (wic *WebInfoCollector) extractUsername(urlStr string) string {
	// 简单的用户名提取逻辑
	parts := strings.Split(urlStr, "/")
	if len(parts) > 3 {
		return parts[len(parts)-1]
	}
	return ""
}

func (wic *WebInfoCollector) detectFrameworks(techInfo *TechnicalInfo, doc *goquery.Document) {
	// 检测常见的JavaScript框架
	frameworks := map[string]string{
		"jquery":     "jQuery",
		"bootstrap":  "Bootstrap",
		"vue":        "Vue.js",
		"react":      "React",
		"angular":    "Angular",
		"layui":      "Layui",
		"element-ui": "Element UI",
	}

	doc.Find("script[src]").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		srcLower := strings.ToLower(src)
		
		for key, name := range frameworks {
			if strings.Contains(srcLower, key) {
				techInfo.Frameworks = append(techInfo.Frameworks, name)
			}
		}
	})
}

func (wic *WebInfoCollector) detectAnalytics(techInfo *TechnicalInfo, doc *goquery.Document) {
	// 检测分析工具
	analytics := map[string]string{
		"google-analytics": "Google Analytics",
		"gtag":            "Google Analytics",
		"baidu":           "百度统计",
		"cnzz":            "CNZZ",
		"51la":            "51LA",
	}

	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		content := s.Text()
		
		for key, name := range analytics {
			if strings.Contains(strings.ToLower(src), key) || 
			   strings.Contains(strings.ToLower(content), key) {
				techInfo.Analytics = append(techInfo.Analytics, name)
			}
		}
	})
}

func (wic *WebInfoCollector) detectCMS(techInfo *TechnicalInfo, doc *goquery.Document, resp *http.Response) {
	// 检测CMS系统
	cmsIndicators := map[string]string{
		"wp-content":     "WordPress",
		"wordpress":      "WordPress",
		"drupal":         "Drupal",
		"joomla":         "Joomla",
		"discuz":         "Discuz",
		"dedecms":        "DedeCMS",
		"phpcms":         "PHPCMS",
		"empire":         "帝国CMS",
	}

	// 检查HTML内容
	html, _ := doc.Html()
	htmlLower := strings.ToLower(html)
	
	for indicator, cms := range cmsIndicators {
		if strings.Contains(htmlLower, indicator) {
			techInfo.CMS = cms
			break
		}
	}

	// 检查HTTP头
	for header, value := range resp.Header {
		headerLower := strings.ToLower(header + ": " + strings.Join(value, " "))
		for indicator, cms := range cmsIndicators {
			if strings.Contains(headerLower, indicator) {
				techInfo.CMS = cms
				return
			}
		}
	}
}

func (wic *WebInfoCollector) addLinksToQueue(ctx *CrawlContext, doc *goquery.Document, baseURL string, depth int) {
	base, _ := url.Parse(baseURL)
	
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href == "" {
			return
		}

		// 转换为绝对URL
		linkURL := wic.resolveURL(base, href)
		
		// 只处理同域名的链接
		linkParsed, err := url.Parse(linkURL)
		if err != nil || linkParsed.Host != base.Host {
			return
		}

		// 避免重复访问
		if !ctx.VisitedURLs[linkURL] {
			ctx.Queue = append(ctx.Queue, CrawlItem{
				URL:   linkURL,
				Depth: depth,
			})
			ctx.Stats.TotalLinks++
		}
	})
}