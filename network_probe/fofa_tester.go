package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// FOFAConfig FOFA API配置
type FOFAConfig struct {
	Email  string `json:"email"`
	Key    string `json:"key"`
	BaseURL string `json:"base_url"`
}

// FOFAResponse FOFA API响应
type FOFAResponse struct {
	Error   bool     `json:"error"`
	ErrMsg  string   `json:"errmsg"`
	Size    int      `json:"size"`
	Page    int      `json:"page"`
	Mode    string   `json:"mode"`
	Query   string   `json:"query"`
	Results [][]string `json:"results"`
}

// FOFATarget FOFA目标
type FOFATarget struct {
	IP       string `json:"ip"`
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
	Title    string `json:"title"`
	Country  string `json:"country"`
	Server   string `json:"server"`
}

// FOFATester FOFA测试器
type FOFATester struct {
	config *FOFAConfig
	client *http.Client
}

// NewFOFATester 创建FOFA测试器
func NewFOFATester(configFile string) (*FOFATester, error) {
	config, err := loadFOFAConfig(configFile)
	if err != nil {
		return nil, err
	}
	
	return &FOFATester{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// loadFOFAConfig 加载FOFA配置
func loadFOFAConfig(configFile string) (*FOFAConfig, error) {
	// 如果配置文件不存在，创建示例配置
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		exampleConfig := &FOFAConfig{
			Email:   "your_email@example.com",
			Key:     "your_fofa_api_key",
			BaseURL: "https://fofa.info/api/v1/search/all",
		}
		
		data, _ := json.MarshalIndent(exampleConfig, "", "  ")
		err := os.WriteFile(configFile, data, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to create config file: %v", err)
		}
		
		return nil, fmt.Errorf("please edit %s with your FOFA credentials", configFile)
	}
	
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	
	var config FOFAConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}
	
	if config.Email == "your_email@example.com" || config.Key == "your_fofa_api_key" {
		return nil, fmt.Errorf("please configure your FOFA credentials in %s", configFile)
	}
	
	return &config, nil
}

// SearchTargets 搜索目标
func (ft *FOFATester) SearchTargets(query string, size int) ([]FOFATarget, error) {
	// Base64编码查询
	encodedQuery := base64.StdEncoding.EncodeToString([]byte(query))
	
	// 构建请求URL
	params := url.Values{}
	params.Add("email", ft.config.Email)
	params.Add("key", ft.config.Key)
	params.Add("qbase64", encodedQuery)
	params.Add("size", strconv.Itoa(size))
	params.Add("fields", "ip,port,protocol,title,country,server")
	
	requestURL := ft.config.BaseURL + "?" + params.Encode()
	
	// 发送请求
	resp, err := ft.client.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	
	// 解析响应
	var fofaResp FOFAResponse
	err = json.Unmarshal(body, &fofaResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}
	
	if fofaResp.Error {
		return nil, fmt.Errorf("FOFA API error: %s", fofaResp.ErrMsg)
	}
	
	// 转换结果
	var targets []FOFATarget
	for _, result := range fofaResp.Results {
		if len(result) >= 6 {
			target := FOFATarget{
				IP:       result[0],
				Port:     result[1],
				Protocol: result[2],
				Title:    result[3],
				Country:  result[4],
				Server:   result[5],
			}
			targets = append(targets, target)
		}
	}
	
	return targets, nil
}

// GetProtocolQueries 获取各协议的FOFA查询语句
func GetProtocolQueries() map[string]string {
	return map[string]string{
		// 工控协议
		"modbus":   `port="502" && protocol="modbus"`,
		"dnp3":     `port="20000" || port="19999"`,
		"bacnet":   `port="47808" && protocol="bacnet"`,
		"opcua":    `port="4840" || (port="4843" && protocol="opcua")`,
		"s7":       `port="102" && (protocol="s7" || title="S7")`,
		
		// 数据库协议
		"mysql":        `port="3306" && protocol="mysql"`,
		"postgresql":   `port="5432" && protocol="postgresql"`,
		"redis":        `port="6379" && protocol="redis"`,
		"sqlserver":    `port="1433" && protocol="mssql"`,
		"oracle":       `port="1521" && protocol="oracle"`,
		"mongodb":      `port="27017" && protocol="mongodb"`,
		"elasticsearch": `port="9200" && protocol="elasticsearch"`,
		"influxdb":     `port="8086" && title="InfluxDB"`,
		"cassandra":    `port="9042" && protocol="cassandra"`,
		"neo4j":        `port="7687" && protocol="neo4j"`,
		
		// IoT协议
		"mqtt":    `port="1883" && protocol="mqtt"`,
		"coap":    `port="5683" && protocol="coap"`,
		"lorawan": `port="1700" && protocol="lorawan"`,
		"amqp":    `port="5672" && protocol="amqp"`,
		
		// 企业基础设施协议
		"ldap":     `port="389" && protocol="ldap"`,
		"kerberos": `port="88" && protocol="kerberos"`,
		"radius":   `port="1812" && protocol="radius"`,
		"ntp":      `port="123" && protocol="ntp"`,
		"syslog":   `port="514" && protocol="syslog"`,
		
		// 安全协议
		"openvpn":   `port="1194" && protocol="openvpn"`,
		"wireguard": `port="51820" && protocol="wireguard"`,
		
		// 电信协议
		"sip": `port="5060" && protocol="sip"`,
		
		// 云服务协议
		"docker":     `port="2375" && title="Docker"`,
		"kubernetes": `port="6443" && title="Kubernetes"`,
		
		// 摄像头协议
		"rtsp":       `port="554" && protocol="rtsp"`,
		"onvif":      `port="80" && title="ONVIF"`,
		"hikvision":  `title="Hikvision" || server="Hikvision"`,
		"dahua":      `title="Dahua" || server="Dahua"`,
		
		// 网络基础协议
		"http":  `port="80" && protocol="http"`,
		"https": `port="443" && protocol="https"`,
		"ssh":   `port="22" && protocol="ssh"`,
		"ftp":   `port="21" && protocol="ftp"`,
		"smtp":  `port="25" && protocol="smtp"`,
		"dns":   `port="53" && protocol="dns"`,
		"snmp":  `port="161" && protocol="snmp"`,
		"telnet": `port="23" && protocol="telnet"`,
		"pop3":  `port="110" && protocol="pop3"`,
		"imap":  `port="143" && protocol="imap"`,
	}
}

// TestProtocol 测试单个协议
func (ft *FOFATester) TestProtocol(protocolName string, query string, probeEngine *ProbeEngine) (*ProtocolTestResult, error) {
	fmt.Printf("🔍 正在搜索 %s 协议资产...\n", protocolName)
	
	// 搜索目标
	targets, err := ft.SearchTargets(query, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to search targets: %v", err)
	}
	
	if len(targets) == 0 {
		fmt.Printf("⚠️  未找到 %s 协议的资产\n", protocolName)
		return &ProtocolTestResult{
			Protocol:     protocolName,
			Query:        query,
			TargetsFound: 0,
			Results:      []TestResult{},
		}, nil
	}
	
	fmt.Printf("✅ 找到 %d 个 %s 协议资产，开始测试...\n", len(targets), protocolName)
	
	// 测试每个目标
	var results []TestResult
	for i, target := range targets {
		fmt.Printf("  [%d/%d] 测试 %s:%s...", i+1, len(targets), target.IP, target.Port)
		
		port, _ := strconv.Atoi(target.Port)
		probeTarget := Target{
			Host: target.IP,
			Port: port,
		}
		
		// 执行探测
		probeResults, err := probeEngine.ProbeTargetWithMode(probeTarget, "port")
		
		testResult := TestResult{
			Target:      fmt.Sprintf("%s:%s", target.IP, target.Port),
			FOFAInfo:    target,
			Success:     false,
			ProbeResults: probeResults,
			Error:       "",
		}
		
		if err != nil {
			testResult.Error = err.Error()
			fmt.Printf(" ❌ 错误: %v\n", err)
		} else {
			// 检查是否有成功的探测结果
			for _, result := range probeResults {
				if result.Success && strings.Contains(strings.ToLower(result.Protocol), strings.ToLower(protocolName)) {
					testResult.Success = true
					testResult.DetectedProtocol = result.Protocol
					testResult.Banner = result.Banner
					testResult.Confidence = result.ParsedInfo.Confidence
					break
				}
			}
			
			if testResult.Success {
				fmt.Printf(" ✅ 成功检测到 %s (置信度: %d%%)\n", testResult.DetectedProtocol, testResult.Confidence)
			} else {
				fmt.Printf(" ❌ 未检测到目标协议\n")
			}
		}
		
		results = append(results, testResult)
		
		// 添加延迟避免过于频繁的请求
		time.Sleep(100 * time.Millisecond)
	}
	
	return &ProtocolTestResult{
		Protocol:     protocolName,
		Query:        query,
		TargetsFound: len(targets),
		Results:      results,
	}, nil
}

// ProtocolTestResult 协议测试结果
type ProtocolTestResult struct {
	Protocol     string       `json:"protocol"`
	Query        string       `json:"query"`
	TargetsFound int          `json:"targets_found"`
	Results      []TestResult `json:"results"`
}

// TestResult 单个测试结果
type TestResult struct {
	Target           string        `json:"target"`
	FOFAInfo         FOFATarget    `json:"fofa_info"`
	Success          bool          `json:"success"`
	DetectedProtocol string        `json:"detected_protocol"`
	Banner           string        `json:"banner"`
	Confidence       int           `json:"confidence"`
	ProbeResults     []*ProbeResult `json:"probe_results"`
	Error            string        `json:"error"`
}

// RunFullTest 运行完整测试
func (ft *FOFATester) RunFullTest(probeEngine *ProbeEngine) (*FullTestReport, error) {
	fmt.Println("🚀 开始FOFA协议检测能力测试")
	fmt.Println(strings.Repeat("=", 50))
	
	queries := GetProtocolQueries()
	var allResults []*ProtocolTestResult
	
	totalProtocols := len(queries)
	currentProtocol := 0
	
	for protocol, query := range queries {
		currentProtocol++
		fmt.Printf("\n[%d/%d] 测试协议: %s\n", currentProtocol, totalProtocols, protocol)
		fmt.Printf("查询语句: %s\n", query)
		
		result, err := ft.TestProtocol(protocol, query, probeEngine)
		if err != nil {
			fmt.Printf("❌ 协议 %s 测试失败: %v\n", protocol, err)
			continue
		}
		
		allResults = append(allResults, result)
		
		// 计算成功率
		successCount := 0
		for _, r := range result.Results {
			if r.Success {
				successCount++
			}
		}
		
		if result.TargetsFound > 0 {
			successRate := float64(successCount) / float64(result.TargetsFound) * 100
			fmt.Printf("📊 %s 协议测试完成: %d/%d 成功 (%.1f%%)\n", 
				protocol, successCount, result.TargetsFound, successRate)
		}
		
		// 添加协议间延迟
		time.Sleep(500 * time.Millisecond)
	}
	
	// 生成测试报告
	report := &FullTestReport{
		Timestamp:       time.Now(),
		TotalProtocols:  len(queries),
		TestedProtocols: len(allResults),
		Results:         allResults,
	}
	
	report.GenerateStatistics()
	
	return report, nil
}

// FullTestReport 完整测试报告
type FullTestReport struct {
	Timestamp       time.Time              `json:"timestamp"`
	TotalProtocols  int                    `json:"total_protocols"`
	TestedProtocols int                    `json:"tested_protocols"`
	Results         []*ProtocolTestResult  `json:"results"`
	Statistics      TestStatistics         `json:"statistics"`
}

// TestStatistics 测试统计
type TestStatistics struct {
	TotalTargets    int     `json:"total_targets"`
	SuccessfulTests int     `json:"successful_tests"`
	FailedTests     int     `json:"failed_tests"`
	OverallSuccessRate float64 `json:"overall_success_rate"`
	ProtocolStats   map[string]ProtocolStats `json:"protocol_stats"`
}

// ProtocolStats 协议统计
type ProtocolStats struct {
	TargetsFound   int     `json:"targets_found"`
	SuccessfulTests int     `json:"successful_tests"`
	SuccessRate    float64 `json:"success_rate"`
	AvgConfidence  float64 `json:"avg_confidence"`
}

// GenerateStatistics 生成统计信息
func (report *FullTestReport) GenerateStatistics() {
	stats := TestStatistics{
		ProtocolStats: make(map[string]ProtocolStats),
	}
	
	for _, result := range report.Results {
		protocolStat := ProtocolStats{
			TargetsFound: result.TargetsFound,
		}
		
		var totalConfidence int
		for _, testResult := range result.Results {
			stats.TotalTargets++
			if testResult.Success {
				stats.SuccessfulTests++
				protocolStat.SuccessfulTests++
				totalConfidence += testResult.Confidence
			} else {
				stats.FailedTests++
			}
		}
		
		if protocolStat.TargetsFound > 0 {
			protocolStat.SuccessRate = float64(protocolStat.SuccessfulTests) / float64(protocolStat.TargetsFound) * 100
		}
		
		if protocolStat.SuccessfulTests > 0 {
			protocolStat.AvgConfidence = float64(totalConfidence) / float64(protocolStat.SuccessfulTests)
		}
		
		stats.ProtocolStats[result.Protocol] = protocolStat
	}
	
	if stats.TotalTargets > 0 {
		stats.OverallSuccessRate = float64(stats.SuccessfulTests) / float64(stats.TotalTargets) * 100
	}
	
	report.Statistics = stats
}

// PrintReport 打印测试报告
func (report *FullTestReport) PrintReport() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📊 FOFA协议检测能力测试报告")
	fmt.Println(strings.Repeat("=", 80))
	
	fmt.Printf("🕒 测试时间: %s\n", report.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("🔍 测试协议数: %d/%d\n", report.TestedProtocols, report.TotalProtocols)
	fmt.Printf("🎯 总测试目标: %d\n", report.Statistics.TotalTargets)
	fmt.Printf("✅ 成功检测: %d\n", report.Statistics.SuccessfulTests)
	fmt.Printf("❌ 检测失败: %d\n", report.Statistics.FailedTests)
	fmt.Printf("📈 总体成功率: %.1f%%\n", report.Statistics.OverallSuccessRate)
	
	fmt.Println("\n📋 各协议检测详情:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-15s %-8s %-8s %-10s %-12s\n", "协议", "目标数", "成功数", "成功率", "平均置信度")
	fmt.Println(strings.Repeat("-", 80))
	
	for protocol, stats := range report.Statistics.ProtocolStats {
		fmt.Printf("%-15s %-8d %-8d %-9.1f%% %-11.1f%%\n", 
			protocol, stats.TargetsFound, stats.SuccessfulTests, 
			stats.SuccessRate, stats.AvgConfidence)
	}
	
	fmt.Println(strings.Repeat("-", 80))
	
	// 显示最佳和最差协议
	var bestProtocol, worstProtocol string
	var bestRate, worstRate float64 = -1, 101
	
	for protocol, stats := range report.Statistics.ProtocolStats {
		if stats.TargetsFound > 0 {
			if stats.SuccessRate > bestRate {
				bestRate = stats.SuccessRate
				bestProtocol = protocol
			}
			if stats.SuccessRate < worstRate {
				worstRate = stats.SuccessRate
				worstProtocol = protocol
			}
		}
	}
	
	if bestProtocol != "" {
		fmt.Printf("🏆 最佳检测协议: %s (%.1f%%)\n", bestProtocol, bestRate)
	}
	if worstProtocol != "" {
		fmt.Printf("⚠️  待优化协议: %s (%.1f%%)\n", worstProtocol, worstRate)
	}
}

// SaveReport 保存测试报告
func (report *FullTestReport) SaveReport(filename string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filename, data, 0644)
}