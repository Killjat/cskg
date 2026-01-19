package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"ip-discovery/apnic"
	"ip-discovery/scanner"
	"ip-discovery/storage"
)

// Config 配置结构
type Config struct {
	InfluxDB struct {
		URL            string `yaml:"url"`
		Token          string `yaml:"token"`
		Organization   string `yaml:"organization"`
		SegmentsBucket string `yaml:"segments_bucket"`
		AliveBucket    string `yaml:"alive_bucket"`
	} `yaml:"influxdb"`

	APNIC struct {
		DelegatedURL string `yaml:"delegated_url"`
		CacheFile    string `yaml:"cache_file"`
		CacheHours   int    `yaml:"cache_hours"`
	} `yaml:"apnic"`

	Scanner struct {
		Workers       int `yaml:"workers"`
		PingTimeout   int `yaml:"ping_timeout"`
		IpsPerSegment int `yaml:"ips_per_segment"`
		ScanInterval  int `yaml:"scan_interval"`
	} `yaml:"scanner"`

	Logging struct {
		Level string `yaml:"level"`
		File  string `yaml:"file"`
	} `yaml:"logging"`
}

var (
	configFile string
	config     Config
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "ip-discovery",
		Short: "IP发现系统",
		Long:  `通过APNIC数据获取台湾省IP段，进行探活扫描并存储到InfluxDB`,
	}

	// 全局标志
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "配置文件路径")

	// 子命令
	rootCmd.AddCommand(fetchCmd())
	rootCmd.AddCommand(scanCmd())
	rootCmd.AddCommand(statsCmd())
	rootCmd.AddCommand(testCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// loadConfig 加载配置文件
func loadConfig() error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	return nil
}

// fetchCmd 获取APNIC数据命令
func fetchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fetch",
		Short: "获取并解析APNIC数据",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadConfig(); err != nil {
				return err
			}

			// 创建APNIC获取器
			fetcher := apnic.NewFetcher(
				config.APNIC.DelegatedURL,
				config.APNIC.CacheFile,
				config.APNIC.CacheHours,
			)

			// 获取数据
			fmt.Println("正在获取APNIC数据...")
			dataFile, err := fetcher.FetchData()
			if err != nil {
				return fmt.Errorf("获取APNIC数据失败: %v", err)
			}

			// 解析数据
			parser := apnic.NewParser("TW") // 台湾省代码
			fmt.Println("正在解析台湾省IP段...")
			segments, err := parser.ParseFile(dataFile)
			if err != nil {
				return fmt.Errorf("解析APNIC数据失败: %v", err)
			}

			// 拆分为C段
			fmt.Println("正在拆分为C段...")
			cSegments := parser.SplitToCSegments(segments)

			// 连接InfluxDB
			influxClient := storage.NewInfluxDBClient(
				config.InfluxDB.URL,
				config.InfluxDB.Token,
				config.InfluxDB.Organization,
				config.InfluxDB.SegmentsBucket,
				config.InfluxDB.AliveBucket,
			)
			defer influxClient.Close()

			// 测试连接
			if err := influxClient.TestConnection(); err != nil {
				log.Printf("InfluxDB连接测试失败: %v", err)
			}

			// 写入IP段信息
			fmt.Println("正在写入IP段信息到InfluxDB...")
			if err := influxClient.WriteIPSegments(cSegments); err != nil {
				return fmt.Errorf("写入IP段信息失败: %v", err)
			}

			fmt.Printf("成功处理 %d 个IP段，拆分为 %d 个C段并写入InfluxDB\n", len(segments), len(cSegments))
			return nil
		},
	}
}

// scanCmd 扫描命令
func scanCmd() *cobra.Command {
	var (
		testCIDR string
		maxSegments int
	)

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "扫描IP段进行探活",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadConfig(); err != nil {
				return err
			}

			// 连接InfluxDB
			influxClient := storage.NewInfluxDBClient(
				config.InfluxDB.URL,
				config.InfluxDB.Token,
				config.InfluxDB.Organization,
				config.InfluxDB.SegmentsBucket,
				config.InfluxDB.AliveBucket,
			)
			defer influxClient.Close()

			// 创建扫描器
			segmentScanner := scanner.NewSegmentScanner(
				influxClient,
				config.Scanner.Workers,
				time.Duration(config.Scanner.PingTimeout)*time.Millisecond,
				config.Scanner.IpsPerSegment,
				time.Duration(config.Scanner.ScanInterval)*time.Millisecond,
			)

			// 如果指定了测试CIDR，只扫描该段
			if testCIDR != "" {
				fmt.Printf("测试扫描CIDR: %s\n", testCIDR)
				results, err := segmentScanner.ScanSingleSegment(testCIDR)
				if err != nil {
					return fmt.Errorf("扫描失败: %v", err)
				}

				// 显示结果
				aliveCount := 0
				for _, result := range results {
					if result.IsAlive {
						aliveCount++
						fmt.Printf("存活IP: %s (响应时间: %v)\n", result.IP, result.ResponseTime)
					}
				}
				fmt.Printf("扫描完成，共发现 %d 个存活IP\n", aliveCount)
				return nil
			}

			// 这里应该从InfluxDB读取IP段列表进行扫描
			// 为了演示，我们创建一些测试段
			fmt.Println("开始扫描所有IP段...")
			fmt.Println("注意：实际使用时应从InfluxDB读取IP段列表")

			// 示例：扫描一些公共DNS服务器段
			testSegments := []*storage.IPSegment{
				{CIDR: "8.8.8.0/24", Country: "TW", Type: "ipv4", Status: "allocated", CreatedAt: time.Now()},
				{CIDR: "1.1.1.0/24", Country: "TW", Type: "ipv4", Status: "allocated", CreatedAt: time.Now()},
			}

			if maxSegments > 0 && len(testSegments) > maxSegments {
				testSegments = testSegments[:maxSegments]
			}

			return segmentScanner.ScanSegments(testSegments)
		},
	}

	cmd.Flags().StringVar(&testCIDR, "cidr", "", "测试扫描指定的CIDR段")
	cmd.Flags().IntVar(&maxSegments, "max", 0, "最大扫描段数（0表示无限制）")

	return cmd
}

// statsCmd 统计命令
func statsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "显示扫描统计信息",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadConfig(); err != nil {
				return err
			}

			// 连接InfluxDB
			influxClient := storage.NewInfluxDBClient(
				config.InfluxDB.URL,
				config.InfluxDB.Token,
				config.InfluxDB.Organization,
				config.InfluxDB.SegmentsBucket,
				config.InfluxDB.AliveBucket,
			)
			defer influxClient.Close()

			// 测试连接
			if err := influxClient.TestConnection(); err != nil {
				return fmt.Errorf("InfluxDB连接失败: %v", err)
			}

			// 获取统计信息
			segmentCount, err := influxClient.GetSegmentCount()
			if err != nil {
				log.Printf("获取IP段数量失败: %v", err)
				segmentCount = 0
			}

			aliveCount, err := influxClient.GetAliveIPCount()
			if err != nil {
				log.Printf("获取存活IP数量失败: %v", err)
				aliveCount = 0
			}

			// 获取最近的存活IP
			recentIPs, err := influxClient.GetRecentAliveIPs(10)
			if err != nil {
				log.Printf("获取最近存活IP失败: %v", err)
			}

			// 显示统计信息
			fmt.Println("=== IP发现系统统计信息 ===")
			fmt.Printf("IP段总数: %d\n", segmentCount)
			fmt.Printf("存活IP数量: %d\n", aliveCount)
			fmt.Println()

			if len(recentIPs) > 0 {
				fmt.Println("最近发现的存活IP:")
				for _, ip := range recentIPs {
					fmt.Printf("  %s (%s) - %s\n", ip.IP, ip.CIDR, ip.ScanTime.Format("2006-01-02 15:04:05"))
				}
			}

			return nil
		},
	}
}

// testCmd 测试命令
func testCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "测试系统组件",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadConfig(); err != nil {
				return err
			}

			fmt.Println("=== 测试InfluxDB连接 ===")
			influxClient := storage.NewInfluxDBClient(
				config.InfluxDB.URL,
				config.InfluxDB.Token,
				config.InfluxDB.Organization,
				config.InfluxDB.SegmentsBucket,
				config.InfluxDB.AliveBucket,
			)
			defer influxClient.Close()

			if err := influxClient.TestConnection(); err != nil {
				fmt.Printf("❌ InfluxDB连接失败: %v\n", err)
			} else {
				fmt.Printf("✅ InfluxDB连接成功\n")
			}

			fmt.Println("\n=== 测试APNIC数据获取 ===")
			fetcher := apnic.NewFetcher(
				config.APNIC.DelegatedURL,
				config.APNIC.CacheFile,
				config.APNIC.CacheHours,
			)

			// 检查缓存状态
			valid, modTime, err := fetcher.GetCacheInfo()
			if err != nil {
				fmt.Printf("📁 缓存文件不存在\n")
			} else {
				status := "过期"
				if valid {
					status = "有效"
				}
				fmt.Printf("📁 缓存文件: %s (%s, 修改时间: %s)\n", 
					config.APNIC.CacheFile, status, modTime.Format("2006-01-02 15:04:05"))
			}

			fmt.Println("\n=== 测试ping功能 ===")
			pingScanner := scanner.NewPingScanner(
				time.Duration(config.Scanner.PingTimeout)*time.Millisecond,
				1,
			)

			testIPs := []string{"8.8.8.8", "1.1.1.1", "114.114.114.114"}
			for _, ip := range testIPs {
				result := pingScanner.ScanIP(ip)
				status := "❌"
				if result.IsAlive {
					status = "✅"
				}
				fmt.Printf("%s %s (响应时间: %v)\n", status, ip, result.ResponseTime)
			}

			return nil
		},
	}
}