package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"honeypot-system/internal/device"
	"honeypot-system/internal/logger"
	"honeypot-system/internal/session"
)

// Server Web服务器结构体
type Server struct {
	// 监听地址
	host string
	// 监听端口
	port int
	// 启用HTTPS
	https bool
	// 会话超时时间
	sessionTimeout int
	// 设备指纹管理器
	fingerprintManager *device.FingerprintManager
	// 会话管理器
	sessionManager interface {
		GetAllSessions() []*session.SessionInfo
	}
	// 日志记录器
	logger logger.Logger
	// Gin引擎
	engine *gin.Engine
	// HTTP服务器
	server *http.Server
}

// NewServer 创建新的Web服务器
func NewServer(host string, port int, https bool, sessionTimeout int, fingerprintManager *device.FingerprintManager, sessionManager interface {
	GetAllSessions() []*session.SessionInfo
}, logger logger.Logger) *Server {
	// 创建Gin引擎
	engine := gin.Default()

	// 设置静态文件服务
	engine.Static("/static", "./web/static")
	// 设置模板目录
	engine.LoadHTMLGlob("./web/templates/*")

	s := &Server{
		host:              host,
		port:              port,
		https:             https,
		sessionTimeout:    sessionTimeout,
		fingerprintManager: fingerprintManager,
		sessionManager:     sessionManager,
		logger:            logger,
		engine:            engine,
	}

	// 注册路由
	s.registerRoutes()

	return s
}

// registerRoutes 注册路由
func (s *Server) registerRoutes() {
	// 主页
	s.engine.GET("/", s.handleIndex)
	// 设备指纹列表
	s.engine.GET("/fingerprints", s.handleFingerprints)
	// 设备指纹详情
	s.engine.GET("/fingerprints/:id", s.handleFingerprintDetail)
	// 会话列表
	s.engine.GET("/sessions", s.handleSessions)
	// API - 获取设备指纹列表
	s.engine.GET("/api/fingerprints", s.handleAPIFingerprints)
	// API - 获取设备指纹详情
	s.engine.GET("/api/fingerprints/:id", s.handleAPIFingerprintDetail)
	// API - 获取会话列表
	s.engine.GET("/api/sessions", s.handleAPISessions)
	// API - 获取统计信息
	s.engine.GET("/api/stats", s.handleAPIStats)
}

// handleIndex 处理主页请求
func (s *Server) handleIndex(c *gin.Context) {
	// 获取所有设备指纹
	fingerprints := s.fingerprintManager.GetAllFingerprints()
	// 获取所有会话
	sessions := s.sessionManager.GetAllSessions()

	// 渲染模板
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":        "Industrial Protocol Honeypot",
		"fingerprints": fingerprints,
		"sessions":     sessions,
		"timestamp":    time.Now().Format(time.RFC3339),
	})
}

// handleFingerprints 处理设备指纹列表请求
func (s *Server) handleFingerprints(c *gin.Context) {
	// 获取所有设备指纹
	fingerprints := s.fingerprintManager.GetAllFingerprints()

	// 渲染模板
	c.HTML(http.StatusOK, "fingerprints.html", gin.H{
		"title":        "Device Fingerprints",
		"fingerprints": fingerprints,
		"timestamp":    time.Now().Format(time.RFC3339),
	})
}

// handleFingerprintDetail 处理设备指纹详情请求
func (s *Server) handleFingerprintDetail(c *gin.Context) {
	// 获取设备指纹ID
	id := c.Param("id")

	// 获取设备指纹
	fingerprint, exists := s.fingerprintManager.GetFingerprint(id)
	if !exists {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "Error",
			"message": "Device fingerprint not found",
		})
		return
	}

	// 渲染模板
	c.HTML(http.StatusOK, "fingerprint_detail.html", gin.H{
		"title":       "Device Fingerprint Detail",
		"fingerprint": fingerprint,
		"timestamp":   time.Now().Format(time.RFC3339),
	})
}

// handleAPIFingerprints 处理API设备指纹列表请求
func (s *Server) handleAPIFingerprints(c *gin.Context) {
	// 获取所有设备指纹
	fingerprints := s.fingerprintManager.GetAllFingerprints()

	// 返回JSON响应
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"fingerprints": fingerprints,
		"count":       len(fingerprints),
	})
}

// handleAPIFingerprintDetail 处理API设备指纹详情请求
func (s *Server) handleAPIFingerprintDetail(c *gin.Context) {
	// 获取设备指纹ID
	id := c.Param("id")

	// 获取设备指纹
	fingerprint, exists := s.fingerprintManager.GetFingerprint(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Device fingerprint not found",
		})
		return
	}

	// 返回JSON响应
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"fingerprint": fingerprint,
	})
}

// handleSessions 处理会话列表请求
func (s *Server) handleSessions(c *gin.Context) {
	// 获取所有会话
	sessions := s.sessionManager.GetAllSessions()

	// 渲染模板
	c.HTML(http.StatusOK, "sessions.html", gin.H{
		"title":     "Sessions",
		"sessions":  sessions,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// handleAPISessions 处理API会话列表请求
func (s *Server) handleAPISessions(c *gin.Context) {
	// 获取所有会话
	sessions := s.sessionManager.GetAllSessions()

	// 返回JSON响应
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// handleAPIStats 处理API统计信息请求
func (s *Server) handleAPIStats(c *gin.Context) {
	// 获取所有设备指纹
	fingerprints := s.fingerprintManager.GetAllFingerprints()
	// 获取所有会话
	sessions := s.sessionManager.GetAllSessions()

	// 计算统计信息
	stats := map[string]interface{}{
		"total_fingerprints": len(fingerprints),
		"total_sessions":     len(sessions),
		"total_connections":  0,
		"unique_ips":         make(map[string]bool),
		"os_distribution":    make(map[string]int),
		"device_types":       make(map[string]int),
		"protocol_distribution": make(map[string]int),
	}

	// 统计详细信息
	for _, fp := range fingerprints {
		// 统计连接数
		stats["total_connections"] = stats["total_connections"].(int) + fp.ConnectionCount
		// 统计唯一IP
		stats["unique_ips"].(map[string]bool)[fp.ClientIP] = true
		// 统计操作系统分布
		os := fp.DeviceInfo.OS
		if os == "" {
			os = "Unknown"
		}
		stats["os_distribution"].(map[string]int)[os]++
		// 统计设备类型
		deviceType := fp.DeviceInfo.DeviceType
		if deviceType == "" {
			deviceType = "Unknown"
		}
		stats["device_types"].(map[string]int)[deviceType]++
		// 统计协议分布
		protocol := fp.Protocol
		stats["protocol_distribution"].(map[string]int)[protocol]++
	}

	// 计算唯一IP数量
	stats["unique_ips_count"] = len(stats["unique_ips"].(map[string]bool))
	// 移除临时映射
	delete(stats, "unique_ips")

	// 返回JSON响应
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"stats":     stats,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Start 启动Web服务器
func (s *Server) Start() error {
	// 构建服务器地址
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	// 创建HTTP服务器
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.engine,
	}

	// 启动服务器
	go func() {
		if s.https {
			s.logger.Info(fmt.Sprintf("Web server started on https://%s", addr))
			if err := s.server.ListenAndServeTLS("./web/cert.pem", "./web/key.pem"); err != nil && err != http.ErrServerClosed {
				s.logger.Error(fmt.Sprintf("Failed to start HTTPS server: %v", err))
			}
		} else {
			s.logger.Info(fmt.Sprintf("Web server started on http://%s", addr))
			if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				s.logger.Error(fmt.Sprintf("Failed to start HTTP server: %v", err))
			}
		}
	}()

	return nil
}

// Stop 停止Web服务器
func (s *Server) Stop() {
	if s.server != nil {
		// 优雅关闭服务器
		if err := s.server.Close(); err != nil {
			s.logger.Error(fmt.Sprintf("Failed to stop web server: %v", err))
		} else {
			s.logger.Info("Web server stopped successfully")
		}
	}
}
