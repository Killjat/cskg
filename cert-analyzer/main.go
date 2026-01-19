package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	// 命令行参数
	targetURL       string
	inputFile       string
	outputFile      string
	outputFormat    string
	timeout         time.Duration
	concurrency     int
	skipVerify      bool
	followRedirects bool
	maxRedirects    int
	userAgent       string
	verbose         bool
	enableSearch    bool
	searchMethods   string
	maxSearchResults int
	searchTimeout   time.Duration
	configFile      string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "cert-analyzer",
		Short: "SSL/TLS Certificate Analyzer",
		Long: `A comprehensive SSL/TLS certificate analysis tool that extracts and analyzes 
certificate information from websites and outputs results in JSON format.`,
		Run: runAnalysis,
	}

	// 添加版本命令
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("cert-analyzer version 1.0.0")
			fmt.Println("SSL/TLS Certificate Analysis Tool")
		},
	}
	rootCmd.AddCommand(versionCmd)

	// 添加命令行参数
	rootCmd.Flags().StringVarP(&targetURL, "url", "u", "", "Target URL to analyze (required if no input file)")
	rootCmd.Flags().StringVarP(&inputFile, "file", "f", "", "Input file containing URLs (one per line)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	rootCmd.Flags().StringVar(&outputFormat, "format", "json", "Output format (json, csv)")
	rootCmd.Flags().DurationVarP(&timeout, "timeout", "t", 10*time.Second, "Connection timeout")
	rootCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 5, "Number of concurrent connections for batch analysis")
	rootCmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "Skip certificate verification")
	rootCmd.Flags().BoolVar(&followRedirects, "follow-redirects", true, "Follow HTTP redirects")
	rootCmd.Flags().IntVar(&maxRedirects, "max-redirects", 5, "Maximum number of redirects to follow")
	rootCmd.Flags().StringVar(&userAgent, "user-agent", "cert-analyzer/1.0", "User agent string")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	
	// 搜索相关参数
	rootCmd.Flags().BoolVar(&enableSearch, "enable-search", false, "Enable searching for related sites using the same certificate")
	rootCmd.Flags().StringVar(&searchMethods, "search-methods", "crtsh", "Search methods (comma-separated): fofa,shodan,censys,crtsh")
	rootCmd.Flags().IntVar(&maxSearchResults, "max-search-results", 20, "Maximum number of related sites to find")
	rootCmd.Flags().DurationVar(&searchTimeout, "search-timeout", 30*time.Second, "Search timeout")
	rootCmd.Flags().StringVar(&configFile, "config", "", "Configuration file for API keys (JSON format)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runAnalysis(cmd *cobra.Command, args []string) {
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
		Timeout:         timeout,
		SkipVerify:      skipVerify,
		FollowRedirects: followRedirects,
		MaxRedirects:    maxRedirects,
		UserAgent:       userAgent,
		Verbose:         verbose,
	}

	// 配置搜索功能
	if enableSearch {
		searchConfig := &CertificateSearchConfig{
			EnableSearch:  true,
			SearchMethods: strings.Split(searchMethods, ","),
			MaxResults:    maxSearchResults,
			Timeout:       searchTimeout,
		}

		// 从配置文件加载API密钥
		if configFile != "" {
			if err := loadSearchConfig(configFile, searchConfig); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to load search config: %v\n", err)
			}
		}

		config.SearchConfig = searchConfig
	}

	// 创建分析器
	analyzer := NewCertificateAnalyzer(config)

	if verbose {
		fmt.Fprintf(os.Stderr, "Starting certificate analysis...\n")
		fmt.Fprintf(os.Stderr, "Timeout: %v\n", timeout)
		fmt.Fprintf(os.Stderr, "Skip Verify: %v\n", skipVerify)
	}

	var output interface{}

	if targetURL != "" {
		// 单个URL分析
		if verbose {
			fmt.Fprintf(os.Stderr, "Analyzing URL: %s\n", targetURL)
		}
		result := analyzer.AnalyzeURL(targetURL)
		output = result
	} else {
		// 批量分析
		if verbose {
			fmt.Fprintf(os.Stderr, "Batch analyzing URLs from file: %s\n", inputFile)
			fmt.Fprintf(os.Stderr, "Concurrency: %d\n", concurrency)
		}
		
		batchAnalyzer := NewBatchAnalyzer(analyzer, concurrency)
		batchResult, err := batchAnalyzer.AnalyzeFromFile(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		
		if verbose {
			fmt.Fprintf(os.Stderr, "Batch analysis completed: %d/%d successful\n", 
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
		fmt.Fprintf(os.Stderr, "Analysis completed successfully\n")
	}
}

func outputResults(data interface{}, filename, format string) error {
	var output []byte
	var err error

	switch format {
	case "json":
		output, err = json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
	case "csv":
		if filename == "" {
			return fmt.Errorf("CSV output requires an output file")
		}
		
		// CSV导出通过BatchAnalyzer处理
		if batchResult, ok := data.(*BatchResult); ok {
			analyzer := NewCertificateAnalyzer(&Config{})
			batchAnalyzer := NewBatchAnalyzer(analyzer, 1)
			return batchAnalyzer.ExportResults(batchResult, filename, "csv")
		} else {
			return fmt.Errorf("CSV export only supported for batch results")
		}
	default:
		return fmt.Errorf("unsupported output format: %s", format)
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

// loadSearchConfig 从配置文件加载搜索配置
func loadSearchConfig(filename string, config *CertificateSearchConfig) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var fileConfig struct {
		FOFA struct {
			Email   string `json:"email"`
			Key     string `json:"key"`
			Enabled bool   `json:"enabled"`
		} `json:"fofa"`
		Shodan struct {
			APIKey  string `json:"api_key"`
			Enabled bool   `json:"enabled"`
		} `json:"shodan"`
		Censys struct {
			AppID   string `json:"app_id"`
			Secret  string `json:"secret"`
			Enabled bool   `json:"enabled"`
		} `json:"censys"`
	}

	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return err
	}

	// 配置FOFA
	if fileConfig.FOFA.Email != "" && fileConfig.FOFA.Key != "" {
		config.FOFAConfig = &FOFAConfig{
			Email:   fileConfig.FOFA.Email,
			Key:     fileConfig.FOFA.Key,
			Enabled: fileConfig.FOFA.Enabled,
		}
	}

	// 配置Shodan
	if fileConfig.Shodan.APIKey != "" {
		config.ShodanConfig = &ShodanConfig{
			APIKey:  fileConfig.Shodan.APIKey,
			Enabled: fileConfig.Shodan.Enabled,
		}
	}

	// 配置Censys
	if fileConfig.Censys.AppID != "" && fileConfig.Censys.Secret != "" {
		config.CensysConfig = &CensysConfig{
			AppID:   fileConfig.Censys.AppID,
			Secret:  fileConfig.Censys.Secret,
			Enabled: fileConfig.Censys.Enabled,
		}
	}

	return nil
}