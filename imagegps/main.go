package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"imagegps/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("=== 图片GPS地理位置提取系统 ===")
	fmt.Println("系统启动中...")

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建路由
	r := gin.Default()

	// 增加更详细的日志记录
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	// 设置文件上传大小限制 (20MB)
	r.MaxMultipartMemory = 20 << 20

	// 静态文件服务
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")

	// 路由配置
	r.GET("/", handler.IndexHandler)
	r.POST("/api/upload", handler.UploadImageHandler)
	r.GET("/api/health", handler.HealthHandler)

	// 启动服务，增加超时设置
	port := ":8080"
	fmt.Printf("服务已启动，访问地址: http://localhost%s\n", port)
	fmt.Println("服务已启动，手机访问地址: http://192.168.0.103:8080")
	fmt.Println("按 Ctrl+C 停止服务")

	// 创建HTTP服务器，增加超时设置
	server := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,  // 增加读取超时到30秒
		WriteTimeout: 30 * time.Second,  // 增加写入超时到30秒
		IdleTimeout:  60 * time.Second,  // 增加空闲超时到60秒
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
