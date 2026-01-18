package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/cskg/CyberStroll/internal/config"
	"github.com/cskg/CyberStroll/internal/elasticsearch"
	"github.com/cskg/CyberStroll/internal/fingerprint"
	"github.com/cskg/CyberStroll/internal/search"
)

func main() {
	// 解析命令行参数
	query := flag.String("query", "", "查询条件，用于筛选要分析的banner数据")
	field := flag.String("field", "", "要提取的字段名，如：server、http、ssh、ftp、smtp、mqtt等")
	flag.Parse()

	// 加载配置
	defaultConfigPath := config.GetDefaultConfigPath()
	cfg, err := config.LoadConfig(defaultConfigPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化Elasticsearch客户端
	esClient, err := elasticsearch.NewClient(&cfg.Elasticsearch)
	if err != nil {
		log.Fatalf("Failed to create elasticsearch client: %v", err)
	}

	// 初始化搜索服务
	searchService := search.NewService(esClient)

	// 初始化指纹分析服务
	fingerprintService := fingerprint.NewService(searchService)

	// 创建上下文
	ctx := context.Background()

	// 执行指纹分析
	var analysisResults map[string][]string
	if *query != "" {
		if *field != "" {
			fmt.Printf("正在分析符合条件 '%s' 的banner数据，提取字段 '%s'...\n\n", *query, *field)
			analysisResults, err = fingerprintService.AnalyzeBannersByQuery(ctx, *query, *field)
		} else {
			fmt.Printf("正在分析符合条件 '%s' 的banner数据...\n\n", *query)
			analysisResults, err = fingerprintService.AnalyzeBannersByQuery(ctx, *query, "")
		}
	} else {
		if *field != "" {
			fmt.Printf("正在分析所有banner数据，提取字段 '%s'...\n\n", *field)
			analysisResults, err = fingerprintService.AnalyzeBanners(ctx, *field)
		} else {
			fmt.Println("正在分析所有banner数据...\n")
			analysisResults, err = fingerprintService.AnalyzeBanners(ctx, "")
		}
	}

	if err != nil {
		log.Fatalf("Failed to analyze banners: %v", err)
	}

	// 格式化并输出结果
	output := fingerprintService.FormatAnalysisResults(analysisResults)
	fmt.Println(output)
}
