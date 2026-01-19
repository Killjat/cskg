package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"api-hunter/analyzer"
	"api-hunter/crawler"
	"api-hunter/storage"
)

var (
	configFile string
	config     Config
)

// Config 应用配置
type Config struct {
	Crawler  crawler.Config         `mapstructure:"crawler"`
	Database storage.DatabaseConfig `mapstructure:"database"`
	Web      WebConfig              `mapstructure:"web"`
	Export   ExportConfig           `mapstructure:"export"`
	Logging  LoggingConfig          `mapstructure:"logging"`
}

// WebConfig Web配置
type WebConfig struct {
	Port       int    `mapstructure:"port"`
	Host       string `mapstructure:"host"`
	StaticDir  string `mapstructure:"static_dir"`
	TemplateDir string `mapstructure:"template_dir"`
}

// ExportConfig 导出配置
type ExportConfig struct {
	DefaultFormat   string `mapstructure:"default_format"`
	OutputDir       string `mapstructure:"output_dir"`
	IncludeDetails  bool   `mapstructure:"include_details"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "api-hunter",
		Short: "API Hunter - 网页API接口发现工具",
		Long:  `API Hunter 是一个专业的网页API接口发现工具，通过深度爬虫技术自动发现和分析网站中的API接口。`,
	}

	// 全局标志
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "配置文件路径")

	// 添加子命令
	rootCmd.AddCommand(scanCmd())
	rootCmd.AddCommand(webCmd())
	rootCmd.AddCommand(exportCmd())
	rootCmd.AddCommand(analyzeCmd())
	rootCmd.AddCommand(statsCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// loadConfig 加载配置文件
func loadConfig() error {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	// 设置默认值
	viper.SetDefault("crawler.max_workers", 10)
	viper.SetDefault("crawler.delay", "1s")
	viper.SetDefault("crawler.timeout", "30s")
	viper.SetDefault("crawler.max_depth", 5)
	viper.SetDefault("crawler.max_pages", 1000)

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	return nil
}

// scanCmd 扫描命令
func scanCmd() *cobra.Command {
	var (
		targetURL string
		depth     int
		workers   int
		sessionID string
	)

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "扫描网站API接口",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadConfig(); err != nil {
				return err
			}

			if targetURL == "" {
				return fmt.Errorf("请指定目标URL")
			}

			// 生成会话ID
			if sessionID == "" {
				sessionID = fmt.Sprintf("scan_%d", time.Now().Unix())
			}

			// 连接数据库
			db, err := storage.NewDatabase(config.Database)
			if err != nil {
				return fmt.Errorf("连接数据库失败: %v", err)
			}
			defer db.Close()

			// 创建爬虫配置
			crawlerConfig := config.Crawler
			if depth > 0 {
				crawlerConfig.MaxDepth = depth
			}
			if workers > 0 {
				crawlerConfig.MaxWorkers = workers
			}

			// 创建爬虫
			spider := crawler.NewSpider(&crawlerConfig, db, sessionID)

			// 开始扫描
			log.Printf("开始扫描: %s (会话ID: %s)", targetURL, sessionID)
			return spider.Start(targetURL)
		},
	}

	cmd.Flags().StringVarP(&targetURL, "url", "u", "", "目标URL (必需)")
	cmd.Flags().IntVarP(&depth, "depth", "d", 0, "爬取深度")
	cmd.Flags().IntVarP(&workers, "workers", "w", 0, "并发数")
	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "会话ID")

	cmd.MarkFlagRequired("url")

	return cmd
}

// webCmd Web界面命令
func webCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "web",
		Short: "启动Web管理界面",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadConfig(); err != nil {
				return err
			}

			// 连接数据库
			db, err := storage.NewDatabase(config.Database)
			if err != nil {
				return fmt.Errorf("连接数据库失败: %v", err)
			}
			defer db.Close()

			// Web配置
			webConfig := config.Web
			if port > 0 {
				webConfig.Port = port
			}

			// 启动简单的Web服务器
			log.Printf("Web界面启动: http://localhost:%d", webConfig.Port)
			return startWebServer(webConfig.Port, db)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 0, "端口号")

	return cmd
}

// startWebServer 启动Web服务器
func startWebServer(port int, db *storage.Database) error {
	if port == 0 {
		port = 8080
	}
	
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Web服务器启动在: http://localhost%s", addr)
	
	// 简单的HTTP服务器实现
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>API Hunter</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 2rem; border-radius: 10px; text-align: center; }
        .content { margin-top: 2rem; }
        .card { background: white; border-radius: 10px; padding: 1.5rem; margin: 1rem 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
    </style>
</head>
<body>
    <div class="header">
        <h1>🔍 API Hunter</h1>
        <p>专业的网页API接口发现工具</p>
    </div>
    <div class="content">
        <div class="card">
            <h2>欢迎使用 API Hunter</h2>
            <p>API Hunter 是一个专业的网页API接口发现工具，通过深度爬虫技术自动发现和分析网站中的API接口。</p>
            <h3>主要功能：</h3>
            <ul>
                <li>🕷️ 深度网页爬虫 - 智能爬取网站页面</li>
                <li>🔍 API自动发现 - 从HTML、JavaScript、表单中提取API端点</li>
                <li>📊 多格式导出 - 支持JSON、CSV、Markdown、HTML格式导出</li>
                <li>📈 统计分析 - 详细的扫描统计和API分类分析</li>
            </ul>
            <h3>使用方法：</h3>
            <p>1. 使用命令行扫描网站：<code>./api-hunter scan -u https://example.com</code></p>
            <p>2. 查看扫描结果：<code>./api-hunter stats</code></p>
            <p>3. 导出结果：<code>./api-hunter export -s session_id -f json -o results.json</code></p>
        </div>
    </div>
</body>
</html>
		`)
	})
	
	return http.ListenAndServe(addr, nil)
}

// exportCmd 导出命令
func exportCmd() *cobra.Command {
	var (
		sessionID string
		format    string
		output    string
	)

	cmd := &cobra.Command{
		Use:   "export",
		Short: "导出扫描结果",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadConfig(); err != nil {
				return err
			}

			// 连接数据库
			db, err := storage.NewDatabase(config.Database)
			if err != nil {
				return fmt.Errorf("连接数据库失败: %v", err)
			}
			defer db.Close()

			// 创建导出器
			exporter := storage.NewExporter(db)

			// 导出选项
			options := storage.ExportOptions{
				Format:         storage.ExportFormat(format),
				OutputPath:     output,
				SessionID:      sessionID,
				IncludeDetails: config.Export.IncludeDetails,
			}

			// 执行导出
			result, err := exporter.Export(options)
			if err != nil {
				return fmt.Errorf("导出失败: %v", err)
			}

			log.Printf("导出完成: %s (%d 条记录, %s)", 
				result.FilePath, result.RecordCount, formatFileSize(result.FileSize))

			return nil
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "会话ID")
	cmd.Flags().StringVarP(&format, "format", "f", "json", "导出格式 (json, csv, markdown, html)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "输出文件路径")

	cmd.MarkFlagRequired("session")
	cmd.MarkFlagRequired("output")

	return cmd
}

// analyzeCmd 分析命令
func analyzeCmd() *cobra.Command {
	var sessionID string

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "分析JavaScript文件中的API",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadConfig(); err != nil {
				return err
			}

			// 连接数据库
			db, err := storage.NewDatabase(config.Database)
			if err != nil {
				return fmt.Errorf("连接数据库失败: %v", err)
			}
			defer db.Close()

			// 创建JavaScript分析器
			jsAnalyzer := analyzer.NewJSAnalyzer(db)

			// 分析JavaScript文件
			log.Printf("开始分析JavaScript文件 (会话: %s)", sessionID)
			return jsAnalyzer.AnalyzeJSFiles(sessionID)
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "会话ID")
	cmd.MarkFlagRequired("session")

	return cmd
}

// statsCmd 统计命令
func statsCmd() *cobra.Command {
	var sessionID string

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "显示扫描统计信息",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadConfig(); err != nil {
				return err
			}

			// 连接数据库
			db, err := storage.NewDatabase(config.Database)
			if err != nil {
				return fmt.Errorf("连接数据库失败: %v", err)
			}
			defer db.Close()

			if sessionID != "" {
				// 显示特定会话的统计信息
				stats, err := db.GetStatistics(sessionID)
				if err != nil {
					return fmt.Errorf("获取统计信息失败: %v", err)
				}

				fmt.Printf("=== 会话统计信息: %s ===\n", sessionID)
				fmt.Printf("总页面数: %d\n", stats.TotalPages)
				fmt.Printf("总API数: %d\n", stats.TotalAPIs)
				fmt.Printf("REST APIs: %d\n", stats.RESTAPIs)
				fmt.Printf("GraphQL APIs: %d\n", stats.GraphQLAPIs)
				fmt.Printf("WebSocket APIs: %d\n", stats.WebSocketAPIs)
				fmt.Printf("JavaScript文件: %d\n", stats.JSFiles)
				fmt.Printf("表单: %d\n", stats.Forms)
				fmt.Printf("开始时间: %s\n", stats.StartTime.Format("2006-01-02 15:04:05"))
				if stats.Duration != "" {
					fmt.Printf("持续时间: %s\n", stats.Duration)
				}
				fmt.Printf("涉及域名: %v\n", stats.Domains)
			} else {
				// 显示所有会话
				sessions, err := db.GetSessions(10, 0)
				if err != nil {
					return fmt.Errorf("获取会话列表失败: %v", err)
				}

				fmt.Printf("=== 最近的扫描会话 ===\n")
				for _, session := range sessions {
					fmt.Printf("会话ID: %s\n", session.SessionID)
					fmt.Printf("  目标URL: %s\n", session.TargetURL)
					fmt.Printf("  状态: %s\n", session.Status)
					fmt.Printf("  开始时间: %s\n", session.StartTime.Format("2006-01-02 15:04:05"))
					if session.EndTime != nil {
						fmt.Printf("  结束时间: %s\n", session.EndTime.Format("2006-01-02 15:04:05"))
					}
					fmt.Printf("  页面数: %d\n", session.PagesFound)
					fmt.Printf("  API数: %d\n", session.APIsFound)
					fmt.Println()
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "会话ID (可选)")

	return cmd
}

// formatFileSize 格式化文件大小
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}