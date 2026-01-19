package crawler

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"api-hunter/storage"
)

// Parser HTML解析器
type Parser struct {
	detector *Detector
}

// NewParser 创建解析器
func NewParser() *Parser {
	return &Parser{
		detector: NewDetector(),
	}
}

// ParseResult 解析结果
type ParseResult struct {
	Title   string                  `json:"title"`
	Links   []string               `json:"links"`
	APIs    []storage.APIEndpoint  `json:"apis"`
	JSFiles []storage.JSFile       `json:"js_files"`
	Forms   []storage.FormInfo     `json:"forms"`
}

// ParseHTML 解析HTML内容
func (p *Parser) ParseHTML(content, baseURL string) (*ParseResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("解析HTML失败: %v", err)
	}

	result := &ParseResult{
		Links:   []string{},
		APIs:    []storage.APIEndpoint{},
		JSFiles: []storage.JSFile{},
		Forms:   []storage.FormInfo{},
	}

	// 提取标题
	result.Title = doc.Find("title").First().Text()

	// 提取链接
	result.Links = p.extractLinks(doc, baseURL)

	// 提取API端点
	result.APIs = p.extractAPIs(doc, baseURL)

	// 提取JavaScript文件
	result.JSFiles = p.extractJSFiles(doc, baseURL)

	// 提取表单
	result.Forms = p.extractForms(doc, baseURL)

	return result, nil
}

// extractLinks 提取页面链接
func (p *Parser) extractLinks(doc *goquery.Document, baseURL string) []string {
	var links []string
	seen := make(map[string]bool)

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		// 解析URL
		absoluteURL := p.resolveURL(href, baseURL)
		if absoluteURL != "" && !seen[absoluteURL] {
			links = append(links, absoluteURL)
			seen[absoluteURL] = true
		}
	})

	return links
}

// extractAPIs 提取API端点
func (p *Parser) extractAPIs(doc *goquery.Document, baseURL string) []storage.APIEndpoint {
	var apis []storage.APIEndpoint

	// 从JavaScript代码中提取API
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		scriptContent := s.Text()
		if scriptContent != "" {
			jsAPIs := p.detector.DetectAPIsInJS(scriptContent)
			for _, api := range jsAPIs {
				api.CreatedAt = time.Now()
				api.UpdatedAt = time.Now()
				apis = append(apis, api)
			}
		}
	})

	// 从data属性中提取API
	doc.Find("[data-api], [data-url], [data-endpoint]").Each(func(i int, s *goquery.Selection) {
		if apiURL, exists := s.Attr("data-api"); exists {
			apis = append(apis, storage.APIEndpoint{
				URL:       p.resolveURL(apiURL, baseURL),
				Method:    "GET",
				Source:    "data-attribute",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}
		if apiURL, exists := s.Attr("data-url"); exists {
			apis = append(apis, storage.APIEndpoint{
				URL:       p.resolveURL(apiURL, baseURL),
				Method:    "GET",
				Source:    "data-attribute",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}
		if apiURL, exists := s.Attr("data-endpoint"); exists {
			apis = append(apis, storage.APIEndpoint{
				URL:       p.resolveURL(apiURL, baseURL),
				Method:    "GET",
				Source:    "data-attribute",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}
	})

	return apis
}

// extractJSFiles 提取JavaScript文件
func (p *Parser) extractJSFiles(doc *goquery.Document, baseURL string) []storage.JSFile {
	var jsFiles []storage.JSFile
	seen := make(map[string]bool)

	doc.Find("script[src]").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if !exists {
			return
		}

		jsURL := p.resolveURL(src, baseURL)
		if jsURL != "" && !seen[jsURL] && p.isJSFile(jsURL) {
			jsFiles = append(jsFiles, storage.JSFile{
				URL:       jsURL,
				CreatedAt: time.Now(),
			})
			seen[jsURL] = true
		}
	})

	return jsFiles
}

// extractForms 提取表单
func (p *Parser) extractForms(doc *goquery.Document, baseURL string) []storage.FormInfo {
	var forms []storage.FormInfo

	doc.Find("form").Each(func(i int, s *goquery.Selection) {
		action, _ := s.Attr("action")
		method, _ := s.Attr("method")

		if method == "" {
			method = "GET"
		}

		// 提取表单字段
		var fields []map[string]string
		s.Find("input, select, textarea").Each(func(j int, input *goquery.Selection) {
			field := make(map[string]string)
			field["name"], _ = input.Attr("name")
			field["type"], _ = input.Attr("type")
			field["id"], _ = input.Attr("id")
			field["placeholder"], _ = input.Attr("placeholder")
			
			if field["name"] != "" || field["id"] != "" {
				fields = append(fields, field)
			}
		})

		fieldsJSON := ""
		if len(fields) > 0 {
			// 简化的JSON序列化
			fieldsJSON = fmt.Sprintf("%v", fields)
		}

		forms = append(forms, storage.FormInfo{
			Action:    p.resolveURL(action, baseURL),
			Method:    strings.ToUpper(method),
			Fields:    fieldsJSON,
			CreatedAt: time.Now(),
		})
	})

	return forms
}

// resolveURL 解析相对URL为绝对URL
func (p *Parser) resolveURL(href, baseURL string) string {
	if href == "" {
		return ""
	}

	// 跳过特殊协议
	if strings.HasPrefix(href, "javascript:") || 
	   strings.HasPrefix(href, "mailto:") || 
	   strings.HasPrefix(href, "tel:") ||
	   strings.HasPrefix(href, "#") {
		return ""
	}

	// 如果已经是绝对URL
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}

	// 解析base URL
	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	// 解析相对URL
	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}

	// 合并URL
	resolved := base.ResolveReference(ref)
	return resolved.String()
}

// isJSFile 检查是否为JavaScript文件
func (p *Parser) isJSFile(url string) bool {
	jsExtensions := []string{".js", ".jsx", ".ts", ".tsx", ".vue"}
	
	for _, ext := range jsExtensions {
		if strings.HasSuffix(strings.ToLower(url), ext) {
			return true
		}
	}
	
	return false
}

// ExtractMetadata 提取页面元数据
func (p *Parser) ExtractMetadata(doc *goquery.Document) map[string]string {
	metadata := make(map[string]string)

	// 提取meta标签
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		property, _ := s.Attr("property")
		content, _ := s.Attr("content")

		if name != "" && content != "" {
			metadata[name] = content
		}
		if property != "" && content != "" {
			metadata[property] = content
		}
	})

	// 提取title
	if title := doc.Find("title").First().Text(); title != "" {
		metadata["title"] = title
	}

	// 提取description
	if desc := doc.Find("meta[name='description']").AttrOr("content", ""); desc != "" {
		metadata["description"] = desc
	}

	return metadata
}

// ExtractStructuredData 提取结构化数据
func (p *Parser) ExtractStructuredData(doc *goquery.Document) []map[string]interface{} {
	var structuredData []map[string]interface{}

	// 提取JSON-LD
	doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		jsonContent := s.Text()
		if jsonContent != "" {
			// 这里应该解析JSON，简化处理
			data := map[string]interface{}{
				"type":    "json-ld",
				"content": jsonContent,
			}
			structuredData = append(structuredData, data)
		}
	})

	// 提取微数据
	doc.Find("[itemscope]").Each(func(i int, s *goquery.Selection) {
		itemType, _ := s.Attr("itemtype")
		if itemType != "" {
			data := map[string]interface{}{
				"type":     "microdata",
				"itemtype": itemType,
			}
			structuredData = append(structuredData, data)
		}
	})

	return structuredData
}