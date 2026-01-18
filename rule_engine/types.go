package main

import (
	"regexp"
	"time"
)

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name        string            `json:"name"`         // 服务名称
	Product     string            `json:"product"`      // 产品名称
	Version     string            `json:"version"`      // 版本号
	Info        string            `json:"info"`         // 附加信息
	Hostname    string            `json:"hostname"`     // 主机名
	OS          string            `json:"os"`           // 操作系统
	DeviceType  string            `json:"device_type"` // 设备类型
	CPE         string            `json:"cpe"`          // CPE标识
	Confidence  int               `json:"confidence"`   // 置信度 (0-100)
	RuleID      string            `json:"rule_id"`      // 匹配的规则ID
	MatchedText string            `json:"matched_text"` // 匹配的文本
	Metadata    map[string]string `json:"metadata"`     // 额外元数据
}

// Rule 匹配规则
type Rule struct {
	ID          string `json:"id"`          // 规则ID
	Service     string `json:"service"`     // 服务名称
	Pattern     string `json:"pattern"`     // 匹配模式（正则表达式）
	Product     string `json:"product"`     // 产品名称模板
	Version     string `json:"version"`     // 版本提取模板
	Info        string `json:"info"`        // 信息模板
	Hostname    string `json:"hostname"`    // 主机名模板
	OS          string `json:"os"`          // 操作系统模板
	DeviceType  string `json:"device_type"` // 设备类型模板
	CPE         string `json:"cpe"`         // CPE模板
	Confidence  int    `json:"confidence"`  // 置信度
	Description string `json:"description"` // 规则描述
	Author      string `json:"author"`      // 规则作者
	CreateTime  string `json:"create_time"` // 创建时间
	
	// 内部字段
	compiledRegex *regexp.Regexp `json:"-"`
}

// SimpleRule 简化的规则格式（用户友好）
type SimpleRule struct {
	Service     string `yaml:"service" json:"service"`         // 服务名称
	Pattern     string `yaml:"pattern" json:"pattern"`         // 匹配模式
	Product     string `yaml:"product" json:"product"`         // 产品名称
	Version     string `yaml:"version" json:"version"`         // 版本（可选）
	Description string `yaml:"description" json:"description"` // 描述（可选）
	Confidence  int    `yaml:"confidence" json:"confidence"`   // 置信度（可选，默认80）
}

// RuleSet 规则集
type RuleSet struct {
	Version     string  `yaml:"version" json:"version"`
	Description string  `yaml:"description" json:"description"`
	Author      string  `yaml:"author" json:"author"`
	Rules       []*Rule `yaml:"rules" json:"rules"`
}

// EngineConfig 引擎配置
type EngineConfig struct {
	RulesDir        string        `yaml:"rules_dir" json:"rules_dir"`               // 规则目录
	CacheEnabled    bool          `yaml:"cache_enabled" json:"cache_enabled"`       // 是否启用缓存
	CacheSize       int           `yaml:"cache_size" json:"cache_size"`             // 缓存大小
	CacheTTL        time.Duration `yaml:"cache_ttl" json:"cache_ttl"`               // 缓存TTL
	MaxConcurrency  int           `yaml:"max_concurrency" json:"max_concurrency"`   // 最大并发数
	DefaultConfidence int         `yaml:"default_confidence" json:"default_confidence"` // 默认置信度
}

// DefaultConfig 默认配置
func DefaultConfig() *EngineConfig {
	return &EngineConfig{
		RulesDir:          "./rules",
		CacheEnabled:      true,
		CacheSize:         1000,
		CacheTTL:          time.Hour,
		MaxConcurrency:    50,
		DefaultConfidence: 80,
	}
}

// EngineStats 引擎统计
type EngineStats struct {
	TotalRules     int           `json:"total_rules"`
	TotalMatches   int64         `json:"total_matches"`
	CacheHits      int64         `json:"cache_hits"`
	CacheMisses    int64         `json:"cache_misses"`
	AvgMatchTime   time.Duration `json:"avg_match_time"`
	LastReloadTime time.Time     `json:"last_reload_time"`
}