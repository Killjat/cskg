package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	// 命令行参数
	targetURL       string
	inputFile       string
	outputFile      string
	outputFormat    string
	maxDepth        int
	maxPages        int
	timeout         time.Duration
	concurrent      int
	userAgent       string
	followRedirects bool
	extractFiles    bool
	extractFooter   bool
	extractIcons    bool
	verbose         bool
	delayBetween    time.Duration
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "web-info-collector",
		Short: "Website Information Collector",
		Long: `A comprehensive website information collection tool that extracts:
- Basic info (title, description, keywords)
- Icons (favicon, apple-touch-icon)
- Registration info (ICP license, police record)
- Download links and files
- Footer information and contact details
- Technical information (server, CMS, frameworks)`,
		Run: runCollection,
	}

	// 添加版本命令
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("web-info-collector version 1.0.0")
			fmt.Println("Website Information Collection Tool")
		},
	}
	rootCmd.AddCommand(versionCmd)

	// 添加命令行参数
	rootCmd.Flags().StringVarP(&targetURL, "url", "u", "", "Target URL to analyze (required if no input file)")
	rootCmd.Flags().StringVarP(&inputFile, "file", "f", "", "Input file containing URLs (one per line)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	rootCmd.Flags().StringVar(&outputFormat, "format", "json", "Output format (json, csv, html)")
	rootCmd.Flags().IntVarP(&maxDepth, "depth", "d", 1, "Maximum crawl depth")
	rootCmd.Flags().IntVar(&maxPages, "max-pages", 10, "Maximum pages to crawl per site")
	rootCmd.Flags().DurationVarP(&timeout, "timeout", "t", 30*time.Second, "Request timeout")
	rootCmd.Flags().IntVarP(&concurrent, "concurrent", "c", 5, "Number of concurrent workers for batch processing")
	rootCmd.Flags().StringVar(&userAgent, "user-agent", "Mozilla/5.0 (compatible; WebInfoCollector/1.0)", "User agent string")
	rootCmd.Flags().BoolVar(&followRedirects, "follow-redirects", true, "Follow HTTP redirects")
	rootCmd.Flags().BoolVar(&extractFiles, "extract-files", true, "Extract download links")
	rootCmd.Flags().BoolVar(&extractFooter, "extract-footer", true, "Extract footer information")
	rootCmd.Flags().BoolVar(&extractIcons, "extract-icons", true, "Extract icon information")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.Flags().DurationVar(&delayBetween, "delay", 1*time.Second, "Delay between requests")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runCollection(cmd *cobra.Command, args []string) {
	// 验证参数
	if targetURL == "" && inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: Either --url or --file must be specified\n")
		os.Exit(1)
	}

	if targetURL != "" && inputFile != "" {
		fmt.Fprintf(os.Stderr, "Error: Cannot specify both --url and --file\n")
		os.Exit(1)
	}

	// 创建配置
	config := &Config{
		MaxDepth:             maxDepth,
		MaxPages:             maxPages,
		Timeout:              timeout,
		Concurrent:           concurrent,
		UserAgent:            userAgent,
		FollowRedirects:      followRedirects,
		ExtractFiles:         extractFiles,
		ExtractFooter:        extractFooter,
		ExtractIcons:         extractIcons,
		Verbose:              verbose,
		DelayBetweenRequests: delayBetween,
	}

	// 创建收集器
	collector := NewWebInfoCollector(config)

	if verbose {
		fmt.Fprintf(os.Stderr, "Starting website information collection...\n")
		fmt.Fprintf(os.Stderr, "Max Depth: %d, Max Pages: %d\n", maxDepth, maxPages)
		fmt.Fprintf(os.Stderr, "Timeout: %v, Concurrent: %d\n", timeout, concurrent)
	}

	var output interface{}

	if targetURL != "" {
		// 单个URL收集
		if verbose {
			fmt.Fprintf(os.Stderr, "Collecting info from URL: %s\n", targetURL)
		}
		result := collector.CollectWebInfo(targetURL)
		output = result
	} else {
		// 批量收集
		if verbose {
			fmt.Fprintf(os.Stderr, "Batch collecting from file: %s\n", inputFile)
		}
		
		batchCollector := NewBatchCollector(collector, concurrent)
		batchResult, err := batchCollector.CollectFromFile(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		
		if verbose {
			fmt.Fprintf(os.Stderr, "Batch collection completed: %d/%d successful\n", 
				batchResult.SuccessCount, batchResult.TotalURLs)
		}
		
		output = batchResult
	}

	// 输出结果
	if err := outputResults(output, outputFile, outputFormat); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Collection completed successfully\n")
	}
}

func outputResults(data interface{}, filename, format string) error {
	switch format {
	case "json":
		return outputJSON(data, filename)
	case "csv", "html":
		if filename == "" {
			return fmt.Errorf("%s output requires an output file", format)
		}
		
		// CSV和HTML导出通过BatchCollector处理
		if batchResult, ok := data.(*BatchResult); ok {
			collector := NewWebInfoCollector(&Config{})
			batchCollector := NewBatchCollector(collector, 1)
			return batchCollector.ExportResults(batchResult, filename, format)
		} else {
			return fmt.Errorf("%s export only supported for batch results", format)
		}
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func outputJSON(data interface{}, filename string) error {
	var output []byte
	var err error

	output, err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// 写入文件或标准输出
	if filename != "" {
		err = os.WriteFile(filename, output, 0644)
		if err != nil {
			return fmt.Errorf("failed to write file: %v", err)
		}
	} else {
		fmt.Print(string(output))
	}

	return nil
}