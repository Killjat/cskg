package crawler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"api-hunter/storage"
	"api-hunter/utils"
)

// Spider 爬虫引擎
type Spider struct {
	config    *Config
	db        *storage.Database
	parser    *Parser
	detector  *Detector
	client    *http.Client
	visited   sync.Map
	queue     chan *CrawlTask
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	sessionID string
}

// Config 爬虫配置
type Config struct {
	MaxWorkers    int           `yaml:"max_workers"`
	Delay         time.Duration `yaml:"delay"`
	Timeout       time.Duration `yaml:"timeout"`
	MaxDepth      int           `yaml:"max_depth"`
	MaxPages      int           `yaml:"max_pages"`
	UserAgents    []string      `yaml:"user_agents"`
	UseHeadless   bool          `yaml:"use_headless"`
	AllowedDomains []string     `yaml:"allowed_domains"`
	BlockedDomains []string     `yaml:"blocked_domains"`
	BlockedPaths   []string     `yaml:"blocked_paths"`
}

// CrawlTask 爬取任务
type CrawlTask struct {
	URL       string
	Depth     int
	Referer   string
	Method    string
	Headers   map[string]string
	Body      string
}

// NewSpider 创建爬虫实例
func NewSpider(config *Config, db *storage.Database, sessionID string) *Spider {
	ctx, cancel := context.WithCancel(context.Background())
	
	client := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	spider := &Spider{
		config:    config,
		db:        db,
		client:    client,
		queue:     make(chan *CrawlTask, 1000),
		ctx:       ctx,
		cancel:    cancel,
		sessionID: sessionID,
	}

	spider.parser = NewParser()
	spider.detector = NewDetector()

	return spider
}

// Start 启动爬虫
func (s *Spider) Start(targetURL string) error {
	log.Printf("开始爬取: %s", targetURL)

	// 创建会话记录
	session := &storage.CrawlSession{
		SessionID: s.sessionID,
		TargetURL: targetURL,
		Status:    "running",
		StartTime: time.Now(),
		Depth:     s.config.MaxDepth,
	}
	
	if err := s.db.CreateSession(session); err != nil {
		return fmt.Errorf("创建会话失败: %v", err)
	}

	// 启动工作协程
	for i := 0; i < s.config.MaxWorkers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	// 添加初始任务
	s.addTask(&CrawlTask{
		URL:    targetURL,
		Depth:  0,
		Method: "GET",
	})

	// 等待完成
	s.wg.Wait()
	close(s.queue)

	// 更新会话状态
	endTime := time.Now()
	return s.db.UpdateSession(s.sessionID, "completed", &endTime)
}

// Stop 停止爬虫
func (s *Spider) Stop() {
	log.Println("停止爬虫...")
	s.cancel()
}

// worker 工作协程
func (s *Spider) worker(id int) {
	defer s.wg.Done()
	
	log.Printf("Worker %d 启动", id)
	
	for {
		select {
		case task := <-s.queue:
			if task == nil {
				return
			}
			s.processTask(task)
			
			// 延迟控制
			if s.config.Delay > 0 {
				time.Sleep(s.config.Delay)
			}
			
		case <-s.ctx.Done():
			return
		}
	}
}

// processTask 处理爬取任务
func (s *Spider) processTask(task *CrawlTask) {
	// 检查是否已访问
	if _, loaded := s.visited.LoadOrStore(task.URL, true); loaded {
		return
	}

	// 检查深度限制
	if task.Depth > s.config.MaxDepth {
		return
	}

	// 检查URL过滤
	if !s.isAllowedURL(task.URL) {
		return
	}

	log.Printf("爬取页面: %s (深度: %d)", task.URL, task.Depth)

	// 检查是否已爬取
	if s.db.IsPageCrawled(s.sessionID, task.URL) {
		return
	}

	var content string
	var err error

	// 根据配置选择爬取方式
	if s.config.UseHeadless {
		content, err = s.crawlWithHeadless(task)
	} else {
		content, err = s.crawlWithHTTP(task)
	}

	// 创建页面记录
	page := &storage.CrawledPage{
		SessionID: s.sessionID,
		URL:       task.URL,
		Depth:     task.Depth,
		Size:      int64(len(content)),
	}

	if err != nil {
		log.Printf("爬取失败 %s: %v", task.URL, err)
		page.Error = err.Error()
		s.db.SaveCrawledPage(page)
		return
	}

	// 解析页面
	parseResult, err := s.parser.ParseHTML(content, task.URL)
	if err != nil {
		log.Printf("解析页面失败 %s: %v", task.URL, err)
		page.Error = err.Error()
	} else {
		page.Title = parseResult.Title
		page.Links = len(parseResult.Links)
		page.APIs = len(parseResult.APIs)
		page.JSFiles = len(parseResult.JSFiles)
		page.Forms = len(parseResult.Forms)

		// 保存发现的API
		for _, api := range parseResult.APIs {
			api.Source = "crawler"
			s.db.SaveAPIEndpoint(&api)
		}

		// 保存JS文件
		for _, jsFile := range parseResult.JSFiles {
			jsFile.SessionID = s.sessionID
			s.db.SaveJSFile(&jsFile)
		}

		// 保存表单信息
		for _, form := range parseResult.Forms {
			form.SessionID = s.sessionID
			form.PageURL = task.URL
			s.db.SaveFormInfo(&form)
		}

		// 添加新的爬取任务
		for _, link := range parseResult.Links {
			s.addTask(&CrawlTask{
				URL:     link,
				Depth:   task.Depth + 1,
				Referer: task.URL,
				Method:  "GET",
			})
		}
	}

	// 保存页面记录
	s.db.SaveCrawledPage(page)
}

// crawlWithHTTP 使用HTTP客户端爬取
func (s *Spider) crawlWithHTTP(task *CrawlTask) (string, error) {
	req, err := http.NewRequestWithContext(s.ctx, task.Method, task.URL, strings.NewReader(task.Body))
	if err != nil {
		return "", err
	}

	// 设置请求头
	req.Header.Set("User-Agent", s.getRandomUserAgent())
	if task.Referer != "" {
		req.Header.Set("Referer", task.Referer)
	}
	
	for key, value := range task.Headers {
		req.Header.Set(key, value)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 检查内容类型
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/json") {
		return "", fmt.Errorf("不支持的内容类型: %s", contentType)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	html, err := doc.Html()
	if err != nil {
		return "", err
	}

	return html, nil
}

// crawlWithHeadless 使用无头浏览器爬取
func (s *Spider) crawlWithHeadless(task *CrawlTask) (string, error) {
	ctx, cancel := chromedp.NewContext(s.ctx)
	defer cancel()

	var content string
	
	err := chromedp.Run(ctx,
		chromedp.Navigate(task.URL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.OuterHTML("html", &content),
	)

	return content, err
}

// addTask 添加爬取任务
func (s *Spider) addTask(task *CrawlTask) {
	select {
	case s.queue <- task:
	case <-s.ctx.Done():
	default:
		// 队列满了，丢弃任务
	}
}

// isAllowedURL 检查URL是否允许爬取
func (s *Spider) isAllowedURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// 检查协议
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	// 检查域名白名单
	if len(s.config.AllowedDomains) > 0 {
		allowed := false
		for _, domain := range s.config.AllowedDomains {
			if strings.Contains(u.Host, domain) {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	// 检查域名黑名单
	for _, domain := range s.config.BlockedDomains {
		if strings.Contains(u.Host, domain) {
			return false
		}
	}

	// 检查路径黑名单
	for _, path := range s.config.BlockedPaths {
		if strings.Contains(u.Path, path) {
			return false
		}
	}

	return true
}

// getRandomUserAgent 获取随机User-Agent
func (s *Spider) getRandomUserAgent() string {
	if len(s.config.UserAgents) == 0 {
		return "Mozilla/5.0 (compatible; API-Hunter/1.0)"
	}
	
	return utils.RandomChoice(s.config.UserAgents)
}

// GetStatistics 获取爬取统计
func (s *Spider) GetStatistics() (*storage.ScanStatistics, error) {
	return s.db.GetStatistics(s.sessionID)
}