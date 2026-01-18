package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/server-monitor/internal/collector"
	"github.com/example/server-monitor/internal/config"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置文件
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("加载配置文件失败，使用默认配置: %v", err)
		// 创建默认配置
		cfg = &config.Config{}
		config.SetDefaults(cfg) // 设置默认值
		config.SaveConfig(cfg, "config.yaml")
	}

	// 初始化采集器
	loginCollector := collector.NewLoginCollector(cfg.Collector.LoginInterval)
	processCollector := collector.NewProcessCollector(cfg.Collector.ProcessInterval)
	commandCollector := collector.NewCommandCollector(cfg.Collector.CommandInterval)
	systemStatsCollector := collector.NewSystemStatsCollector(10) // 每10秒采集一次系统统计信息
	
	// 初始化文件采集器
	fileCollector, err := collector.NewFileCollector(cfg.Collector.RecursiveWatch)
	if err != nil {
		log.Printf("初始化文件采集器失败: %v", err)
	}

	// 启动采集器
	loginCollector.Start()
	processCollector.Start()
	commandCollector.Start()
	systemStatsCollector.Start()
	
	if fileCollector != nil {
		fileCollector.Start()
		// 添加监控路径
		for _, path := range cfg.Collector.FileWatchPaths {
			if err := fileCollector.AddWatch(path, cfg.Collector.RecursiveWatch); err != nil {
				log.Printf("添加文件监控路径失败: %v", err)
			}
		}
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	r := gin.Default()

	// 静态文件服务
	r.Static("/static", "./static")

	// 主页面
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// API路由
	setupRoutes(r, loginCollector, processCollector, commandCollector, fileCollector, systemStatsCollector)

	// 启动HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: r,
	}

	// 启动服务器（非阻塞）
	go func() {
		log.Printf("服务器启动，监听地址: %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")

	// 停止采集器
	loginCollector.Stop()
	processCollector.Stop()
	commandCollector.Stop()
	systemStatsCollector.Stop()
	
	if fileCollector != nil {
		fileCollector.Stop()
	}

	log.Println("服务器已关闭")
}

// setupRoutes 设置API路由
func setupRoutes(r *gin.Engine, 
	loginCollector collector.LoginCollector,
	processCollector collector.ProcessCollector,
	commandCollector collector.CommandCollector,
	fileCollector collector.FileCollector,
	systemStatsCollector collector.SystemStatsCollector) {

	// API版本前缀
	v1 := r.Group("/api/v1")
	{
		// 登录信息
		v1.GET("/logins", func(c *gin.Context) {
			logins, err := loginCollector.CollectCurrentLogins()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": logins})
		})

		// 进程信息
		v1.GET("/processes", func(c *gin.Context) {
			processes, err := processCollector.CollectProcesses()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": processes})
		})

		v1.GET("/processes/:pid", func(c *gin.Context) {
			pid := c.Param("pid")
			pidInt := 0
			fmt.Sscanf(pid, "%d", &pidInt)
			process, err := processCollector.GetProcessByPID(pidInt)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": process})
		})

		// 进程数量统计
		v1.GET("/processes/count", func(c *gin.Context) {
			total, err := processCollector.GetProcessCount()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			running, err := processCollector.GetRunningProcessCount()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": gin.H{
				"total":   total,
				"running": running,
			}})
		})

		// 命令信息
		v1.GET("/commands", func(c *gin.Context) {
			commands, err := commandCollector.CollectCurrentCommands()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": commands})
		})

		// 文件操作信息
		v1.GET("/file-operations", func(c *gin.Context) {
			if fileCollector == nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "文件采集器未初始化"})
				return
			}
			fileOps, err := fileCollector.CollectFileOperations()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": fileOps})
		})

		// 系统统计信息
		v1.GET("/system-stats", func(c *gin.Context) {
			stats, err := systemStatsCollector.CollectSystemStats()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": stats})
		})

		// 健康检查
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}
}
