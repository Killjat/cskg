package web

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"api-hunter/storage"
)

// Config Web服务器配置
type Config struct {
	Port       int    `yaml:"port"`
	Host       string `yaml:"host"`
	StaticDir  string `yaml:"static_dir"`
	TemplateDir string `yaml:"template_dir"`
	Auth       AuthConfig `yaml:"auth"`
	CORS       CORSConfig `yaml:"cors"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	Enabled bool     `yaml:"enabled"`
	Origins []string `yaml:"origins"`
	Methods []string `yaml:"methods"`
	Headers []string `yaml:"headers"`
}

// Server Web服务器
type Server struct {
	config   *Config
	db       *Database
	router   *gin.Engine
	handlers *Handlers
}

// NewServer 创建Web服务器
func NewServer(config *Config, db *storage.Database) *Server {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 配置CORS
	if config.CORS.Enabled {
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowOrigins = config.CORS.Origins
		corsConfig.AllowMethods = config.CORS.Methods
		corsConfig.AllowHeaders = config.CORS.Headers
		router.Use(cors.New(corsConfig))
	}

	server := &Server{
		config: config,
		db:     &Database{db: db},
		router: router,
	}

	server.handlers = NewHandlers(server.db)
	server.setupRoutes()

	return server
}

// Database Web数据库包装器
type Database struct {
	db *storage.Database
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 静态文件
	s.router.Static("/static", s.config.StaticDir)
	s.router.LoadHTMLGlob(s.config.TemplateDir + "/*")

	// 首页
	s.router.GET("/", s.handlers.Index)

	// API路由组
	api := s.router.Group("/api/v1")
	{
		// 会话管理
		api.GET("/sessions", s.handlers.GetSessions)
		api.GET("/sessions/:id", s.handlers.GetSession)
		api.DELETE("/sessions/:id", s.handlers.DeleteSession)
		api.GET("/sessions/:id/stats", s.handlers.GetSessionStats)

		// API端点
		api.GET("/apis", s.handlers.GetAPIs)
		api.GET("/apis/search", s.handlers.SearchAPIs)
		api.GET("/apis/domains", s.handlers.GetDomainStats)
		api.GET("/apis/types", s.handlers.GetAPIsByType)

		// 页面管理
		api.GET("/pages", s.handlers.GetPages)
		api.GET("/pages/:id", s.handlers.GetPage)

		// JavaScript文件
		api.GET("/jsfiles", s.handlers.GetJSFiles)
		api.POST("/jsfiles/analyze", s.handlers.AnalyzeJSFiles)

		// 导出功能
		api.POST("/export", s.handlers.ExportData)

		// 系统信息
		api.GET("/system/info", s.handlers.GetSystemInfo)
	}

	// 认证中间件
	if s.config.Auth.Enabled {
		api.Use(s.basicAuth())
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	log.Printf("Web服务器启动在: http://%s", addr)
	return s.router.Run(addr)
}

// basicAuth 基本认证中间件
func (s *Server) basicAuth() gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		s.config.Auth.Username: s.config.Auth.Password,
	})
}

// Handlers 处理器
type Handlers struct {
	db *Database
}

// NewHandlers 创建处理器
func NewHandlers(db *Database) *Handlers {
	return &Handlers{db: db}
}

// Index 首页
func (h *Handlers) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "API Hunter - 网页API发现工具",
	})
}

// GetSessions 获取会话列表
func (h *Handlers) GetSessions(c *gin.Context) {
	limit := 20
	offset := 0
	
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := c.Query("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	sessions, err := h.db.db.GetSessions(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// GetSession 获取单个会话
func (h *Handlers) GetSession(c *gin.Context) {
	sessionID := c.Param("id")
	
	session, err := h.db.db.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// DeleteSession 删除会话
func (h *Handlers) DeleteSession(c *gin.Context) {
	sessionID := c.Param("id")
	
	if err := h.db.db.DeleteSession(sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "会话删除成功"})
}

// GetSessionStats 获取会话统计
func (h *Handlers) GetSessionStats(c *gin.Context) {
	sessionID := c.Param("id")
	
	stats, err := h.db.db.GetStatistics(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetAPIs 获取API列表
func (h *Handlers) GetAPIs(c *gin.Context) {
	sessionID := c.Query("session_id")
	limit := 50
	offset := 0
	
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := c.Query("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	apis, err := h.db.db.GetAPIEndpoints(sessionID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"apis":  apis,
		"total": len(apis),
	})
}

// SearchAPIs 搜索API
func (h *Handlers) SearchAPIs(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	limit := 50
	offset := 0
	
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := c.Query("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	apis, err := h.db.db.SearchAPIs(keyword, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"apis":    apis,
		"total":   len(apis),
		"keyword": keyword,
	})
}

// GetDomainStats 获取域名统计
func (h *Handlers) GetDomainStats(c *gin.Context) {
	domains, err := h.db.db.GetDomainStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domains": domains,
		"total":   len(domains),
	})
}

// GetAPIsByType 按类型获取API
func (h *Handlers) GetAPIsByType(c *gin.Context) {
	apiType := c.Query("type")
	if apiType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API类型不能为空"})
		return
	}

	limit := 50
	offset := 0
	
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := c.Query("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	apis, err := h.db.db.GetAPIsByType(apiType, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"apis":  apis,
		"total": len(apis),
		"type":  apiType,
	})
}

// GetPages 获取页面列表
func (h *Handlers) GetPages(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID不能为空"})
		return
	}

	limit := 50
	offset := 0
	
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := c.Query("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}

	pages, err := h.db.db.GetCrawledPages(sessionID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pages": pages,
		"total": len(pages),
	})
}

// GetPage 获取单个页面
func (h *Handlers) GetPage(c *gin.Context) {
	// 这里需要实现获取单个页面的逻辑
	c.JSON(http.StatusNotImplemented, gin.H{"error": "功能未实现"})
}

// GetJSFiles 获取JavaScript文件列表
func (h *Handlers) GetJSFiles(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID不能为空"})
		return
	}

	jsFiles, err := h.db.db.GetUnanalyzedJSFiles(sessionID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"js_files": jsFiles,
		"total":    len(jsFiles),
	})
}

// AnalyzeJSFiles 分析JavaScript文件
func (h *Handlers) AnalyzeJSFiles(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 这里需要调用JavaScript分析器
	// analyzer := analyzer.NewJSAnalyzer(h.db.db)
	// err := analyzer.AnalyzeJSFiles(req.SessionID)
	
	c.JSON(http.StatusOK, gin.H{
		"message":    "JavaScript文件分析已启动",
		"session_id": req.SessionID,
	})
}

// ExportData 导出数据
func (h *Handlers) ExportData(c *gin.Context) {
	var req struct {
		SessionID      string `json:"session_id" binding:"required"`
		Format         string `json:"format" binding:"required"`
		IncludeDetails bool   `json:"include_details"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成文件名
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("api_export_%s_%s.%s", req.SessionID, timestamp, req.Format)
	outputPath := fmt.Sprintf("./exports/%s", filename)

	// 创建导出器并导出
	exporter := storage.NewExporter(h.db.db)
	options := storage.ExportOptions{
		Format:         storage.ExportFormat(req.Format),
		OutputPath:     outputPath,
		SessionID:      req.SessionID,
		IncludeDetails: req.IncludeDetails,
	}

	result, err := exporter.Export(options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "导出成功",
		"result":  result,
	})
}

// GetSystemInfo 获取系统信息
func (h *Handlers) GetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":    "1.0.0",
		"build_time": "2024-01-01",
		"go_version": "1.21",
		"uptime":     time.Since(time.Now()).String(),
	})
}