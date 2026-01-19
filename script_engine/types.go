package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// Target 探测目标
type Target struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// String 返回目标的字符串表示
func (t Target) String() string {
	return fmt.Sprintf("%s:%d", t.Host, t.Port)
}

// ParseTarget 解析目标字符串
func ParseTarget(target string) (Target, error) {
	host, portStr, err := net.SplitHostPort(target)
	if err != nil {
		return Target{}, fmt.Errorf("invalid target format: %s", target)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return Target{}, fmt.Errorf("invalid port: %s", portStr)
	}

	return Target{Host: host, Port: port}, nil
}

// ScriptConfig 脚本引擎配置
type ScriptConfig struct {
	Timeout      time.Duration `json:"timeout"`
	Concurrent   int           `json:"concurrent"`
	Verbose      bool          `json:"verbose"`
	OutputFormat string        `json:"output_format"`
}

// Script 脚本定义
type Script struct {
	Name         string                                                    `json:"name"`
	Protocol     string                                                    `json:"protocol"`
	Category     string                                                    `json:"category"`
	Description  string                                                    `json:"description"`
	Author       string                                                    `json:"author"`
	Version      string                                                    `json:"version"`
	Dependencies []string                                                  `json:"dependencies"`
	Execute      func(target Target, ctx *ScriptContext) *ScriptResult    `json:"-"`
}

// ScriptCategory 脚本类别常量
const (
	CategoryDiscovery      = "discovery"
	CategoryVulnerability  = "vulnerability"
	CategoryAuthentication = "authentication"
	CategoryExploitation   = "exploitation"
)

// ScriptContext 脚本执行上下文
type ScriptContext struct {
	Config    *ScriptConfig          `json:"config"`
	Target    Target                 `json:"target"`
	Protocol  string                 `json:"protocol"`
	Timeout   time.Duration          `json:"timeout"`
	Variables map[string]interface{} `json:"variables"`
	Logger    Logger                 `json:"-"`
}

// ScriptResult 脚本执行结果
type ScriptResult struct {
	ScriptName      string                 `json:"script_name"`
	Category        string                 `json:"category"`
	Success         bool                   `json:"success"`
	Error           string                 `json:"error,omitempty"`
	Duration        time.Duration          `json:"duration"`
	Findings        map[string]interface{} `json:"findings"`
	Vulnerabilities []Vulnerability        `json:"vulnerabilities"`
	NextScripts     []string               `json:"next_scripts,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
}

// TargetResult 目标探测结果
type TargetResult struct {
	Target          string            `json:"target"`
	Protocol        string            `json:"protocol"`
	ScriptResults   []*ScriptResult   `json:"script_results"`
	Findings        map[string]interface{} `json:"findings"`
	Vulnerabilities []Vulnerability   `json:"vulnerabilities"`
	Duration        time.Duration     `json:"duration"`
	Timestamp       time.Time         `json:"timestamp"`
}

// Vulnerability 漏洞信息
type Vulnerability struct {
	CVE               string            `json:"cve"`
	Title             string            `json:"title"`
	Description       string            `json:"description"`
	Severity          string            `json:"severity"`
	CVSS              float64           `json:"cvss"`
	ExploitAvailable  bool              `json:"exploit_available"`
	References        []string          `json:"references"`
	AffectedVersions  []string          `json:"affected_versions"`
	FixedVersions     []string          `json:"fixed_versions"`
	Metadata          map[string]string `json:"metadata"`
}

// SeverityLevel 漏洞严重程度
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
	SeverityInfo     = "info"
)

// Logger 日志接口
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// SimpleLogger 简单日志实现
type SimpleLogger struct {
	Verbose bool
}

// Debug 调试日志
func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	if l.Verbose {
		fmt.Printf("[DEBUG] "+msg+"\n", args...)
	}
}

// Info 信息日志
func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	fmt.Printf("[INFO] "+msg+"\n", args...)
}

// Warn 警告日志
func (l *SimpleLogger) Warn(msg string, args ...interface{}) {
	fmt.Printf("[WARN] "+msg+"\n", args...)
}

// Error 错误日志
func (l *SimpleLogger) Error(msg string, args ...interface{}) {
	fmt.Printf("[ERROR] "+msg+"\n", args...)
}

// ScriptRegistry 脚本注册表
type ScriptRegistry struct {
	scripts map[string]*Script
}

// NewScriptRegistry 创建脚本注册表
func NewScriptRegistry() *ScriptRegistry {
	return &ScriptRegistry{
		scripts: make(map[string]*Script),
	}
}

// Register 注册脚本
func (sr *ScriptRegistry) Register(script *Script) error {
	if script.Name == "" {
		return fmt.Errorf("script name cannot be empty")
	}
	
	if script.Protocol == "" {
		return fmt.Errorf("script protocol cannot be empty")
	}
	
	if script.Execute == nil {
		return fmt.Errorf("script execute function cannot be nil")
	}
	
	sr.scripts[script.Name] = script
	return nil
}

// Get 获取脚本
func (sr *ScriptRegistry) Get(name string) (*Script, bool) {
	script, exists := sr.scripts[name]
	return script, exists
}

// GetByProtocol 根据协议获取脚本
func (sr *ScriptRegistry) GetByProtocol(protocol string) []*Script {
	var scripts []*Script
	for _, script := range sr.scripts {
		if script.Protocol == protocol {
			scripts = append(scripts, script)
		}
	}
	return scripts
}

// GetByCategory 根据类别获取脚本
func (sr *ScriptRegistry) GetByCategory(category string) []*Script {
	var scripts []*Script
	for _, script := range sr.scripts {
		if script.Category == category {
			scripts = append(scripts, script)
		}
	}
	return scripts
}

// GetByProtocolAndCategory 根据协议和类别获取脚本
func (sr *ScriptRegistry) GetByProtocolAndCategory(protocol, category string) []*Script {
	var scripts []*Script
	for _, script := range sr.scripts {
		if script.Protocol == protocol && script.Category == category {
			scripts = append(scripts, script)
		}
	}
	return scripts
}

// GetAll 获取所有脚本
func (sr *ScriptRegistry) GetAll() []*Script {
	var scripts []*Script
	for _, script := range sr.scripts {
		scripts = append(scripts, script)
	}
	return scripts
}

// List 列出脚本名称
func (sr *ScriptRegistry) List() []string {
	var names []string
	for name := range sr.scripts {
		names = append(names, name)
	}
	return names
}

// Count 获取脚本数量
func (sr *ScriptRegistry) Count() int {
	return len(sr.scripts)
}

// FilterScripts 过滤脚本
func FilterScripts(scripts []*Script, filter func(*Script) bool) []*Script {
	var filtered []*Script
	for _, script := range scripts {
		if filter(script) {
			filtered = append(filtered, script)
		}
	}
	return filtered
}

// ParseScriptNames 解析脚本名称列表
func ParseScriptNames(scriptStr string) []string {
	if scriptStr == "" || scriptStr == "all" {
		return []string{"all"}
	}
	
	names := strings.Split(scriptStr, ",")
	var result []string
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name != "" {
			result = append(result, name)
		}
	}
	return result
}

// LoadTargetsFromFile 从文件加载目标列表
func LoadTargetsFromFile(filename string) ([]Target, error) {
	// TODO: 实现从文件加载目标
	return nil, fmt.Errorf("LoadTargetsFromFile not implemented yet")
}

// BaseScript 基础脚本结构
type BaseScript struct {
	Name        string
	Protocol    string
	Category    string
	Description string
	Author      string
	Version     string
}

// GetName 获取脚本名称
func (bs *BaseScript) GetName() string {
	return bs.Name
}

// GetProtocol 获取协议
func (bs *BaseScript) GetProtocol() string {
	return bs.Protocol
}

// GetCategory 获取类别
func (bs *BaseScript) GetCategory() string {
	return bs.Category
}

// GetDescription 获取描述
func (bs *BaseScript) GetDescription() string {
	return bs.Description
}

// ScriptInterface 脚本接口
type ScriptInterface interface {
	GetName() string
	GetProtocol() string
	GetCategory() string
	GetDescription() string
	Execute(target Target, ctx *ScriptContext) *ScriptResult
}

// ExecutionStats 执行统计
type ExecutionStats struct {
	TotalScripts     int           `json:"total_scripts"`
	SuccessfulScripts int           `json:"successful_scripts"`
	FailedScripts    int           `json:"failed_scripts"`
	TotalDuration    time.Duration `json:"total_duration"`
	AverageDuration  time.Duration `json:"average_duration"`
	ProtocolStats    map[string]int `json:"protocol_stats"`
	CategoryStats    map[string]int `json:"category_stats"`
}

// CalculateStats 计算执行统计
func CalculateStats(results []*ScriptResult) *ExecutionStats {
	stats := &ExecutionStats{
		ProtocolStats: make(map[string]int),
		CategoryStats: make(map[string]int),
	}
	
	stats.TotalScripts = len(results)
	
	for _, result := range results {
		if result.Success {
			stats.SuccessfulScripts++
		} else {
			stats.FailedScripts++
		}
		
		stats.TotalDuration += result.Duration
		stats.CategoryStats[result.Category]++
	}
	
	if stats.TotalScripts > 0 {
		stats.AverageDuration = stats.TotalDuration / time.Duration(stats.TotalScripts)
	}
	
	return stats
}