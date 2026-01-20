package main

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// BannerEngine Banner匹配引擎
type BannerEngine struct {
	config *EngineConfig
	rules  []*Rule
	cache  *sync.Map
	stats  *EngineStats
	mutex  sync.RWMutex
}

// NewBannerEngine 创建新的Banner引擎
func NewBannerEngine(config *EngineConfig) *BannerEngine {
	if config == nil {
		config = DefaultConfig()
	}
	
	return &BannerEngine{
		config: config,
		rules:  make([]*Rule, 0),
		cache:  &sync.Map{},
		stats: &EngineStats{
			LastReloadTime: time.Now(),
		},
	}
}

// LoadRules 加载规则
func (e *BannerEngine) LoadRules(rules []*Rule) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	compiledRules := make([]*Rule, 0, len(rules))
	
	for _, rule := range rules {
		// 编译正则表达式
		if rule.Pattern != "" {
			regex, err := regexp.Compile(rule.Pattern)
			if err != nil {
				fmt.Printf("警告: 规则 %s 的正则表达式编译失败: %v\n", rule.ID, err)
				continue
			}
			rule.compiledRegex = regex
		}
		
		// 设置默认置信度
		if rule.Confidence == 0 {
			rule.Confidence = e.config.DefaultConfidence
		}
		
		compiledRules = append(compiledRules, rule)
	}
	
	e.rules = compiledRules
	e.stats.TotalRules = len(e.rules)
	e.stats.LastReloadTime = time.Now()
	
	// 清空缓存
	e.cache = &sync.Map{}
	
	return nil
}

// AddRule 添加单个规则
func (e *BannerEngine) AddRule(rule *Rule) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	// 编译正则表达式
	if rule.Pattern != "" {
		regex, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return fmt.Errorf("正则表达式编译失败: %v", err)
		}
		rule.compiledRegex = regex
	}
	
	// 设置默认值
	if rule.Confidence == 0 {
		rule.Confidence = e.config.DefaultConfidence
	}
	if rule.ID == "" {
		rule.ID = fmt.Sprintf("rule_%d", time.Now().Unix())
	}
	if rule.CreateTime == "" {
		rule.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	}
	
	e.rules = append(e.rules, rule)
	e.stats.TotalRules = len(e.rules)
	
	return nil
}

// AddSimpleRule 添加简化规则
func (e *BannerEngine) AddSimpleRule(simpleRule *SimpleRule) error {
	rule := &Rule{
		ID:          fmt.Sprintf("simple_%d", time.Now().UnixNano()),
		Service:     simpleRule.Service,
		Pattern:     simpleRule.Pattern,
		Product:     simpleRule.Product,
		Version:     simpleRule.Version,
		Description: simpleRule.Description,
		Confidence:  simpleRule.Confidence,
		Author:      "user",
		CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
	}
	
	if rule.Confidence == 0 {
		rule.Confidence = e.config.DefaultConfidence
	}
	
	return e.AddRule(rule)
}

// Match 匹配Banner
func (e *BannerEngine) Match(banner string) []*ServiceInfo {
	if banner == "" {
		return nil
	}
	
	start := time.Now()
	defer func() {
		atomic.AddInt64(&e.stats.TotalMatches, 1)
		e.stats.AvgMatchTime = time.Since(start)
	}()
	
	// 检查缓存
	if e.config.CacheEnabled {
		if cached := e.getFromCache(banner); cached != nil {
			atomic.AddInt64(&e.stats.CacheHits, 1)
			return cached
		}
		atomic.AddInt64(&e.stats.CacheMisses, 1)
	}
	
	// 执行匹配
	results := e.matchBanner(banner)
	
	// 缓存结果
	if e.config.CacheEnabled && len(results) > 0 {
		e.putToCache(banner, results)
	}
	
	return results
}

// matchBanner 执行Banner匹配
func (e *BannerEngine) matchBanner(banner string) []*ServiceInfo {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	
	var results []*ServiceInfo
	
	for _, rule := range e.rules {
		if service := e.matchRule(rule, banner); service != nil {
			results = append(results, service)
		}
	}
	
	// 按置信度排序
	if len(results) > 1 {
		for i := 0; i < len(results)-1; i++ {
			for j := i + 1; j < len(results); j++ {
				if results[i].Confidence < results[j].Confidence {
					results[i], results[j] = results[j], results[i]
				}
			}
		}
	}
	
	return results
}

// matchRule 匹配单个规则
func (e *BannerEngine) matchRule(rule *Rule, banner string) *ServiceInfo {
	if rule.compiledRegex == nil {
		return nil
	}
	
	matches := rule.compiledRegex.FindStringSubmatch(banner)
	if matches == nil {
		return nil
	}
	
	// 创建服务信息
	service := &ServiceInfo{
		Name:        rule.Service,
		Confidence:  rule.Confidence,
		RuleID:      rule.ID,
		MatchedText: matches[0],
		Metadata:    make(map[string]string),
	}
	
	// 提取信息
	service.Product = e.extractInfo(rule.Product, matches)
	service.Version = e.extractInfo(rule.Version, matches)
	service.Info = e.extractInfo(rule.Info, matches)
	service.Hostname = e.extractInfo(rule.Hostname, matches)
	service.OS = e.extractInfo(rule.OS, matches)
	service.DeviceType = e.extractInfo(rule.DeviceType, matches)
	service.CPE = e.extractInfo(rule.CPE, matches)
	
	// 如果没有产品名，使用服务名
	if service.Product == "" {
		service.Product = rule.Service
	}
	
	return service
}

// extractInfo 提取信息（支持$1, $2等占位符）
func (e *BannerEngine) extractInfo(template string, matches []string) string {
	if template == "" {
		return ""
	}
	
	result := template
	
	// 替换 $1, $2, ... 占位符
	for i := 1; i < len(matches); i++ {
		placeholder := fmt.Sprintf("$%d", i)
		result = strings.ReplaceAll(result, placeholder, matches[i])
	}
	
	return result
}

// 缓存相关方法
func (e *BannerEngine) getCacheKey(banner string) string {
	h := md5.New()
	h.Write([]byte(banner))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (e *BannerEngine) getFromCache(banner string) []*ServiceInfo {
	key := e.getCacheKey(banner)
	if value, ok := e.cache.Load(key); ok {
		if cached, ok := value.([]*ServiceInfo); ok {
			return cached
		}
	}
	return nil
}

func (e *BannerEngine) putToCache(banner string, results []*ServiceInfo) {
	key := e.getCacheKey(banner)
	e.cache.Store(key, results)
	
	// 简单的TTL实现
	go func() {
		time.Sleep(e.config.CacheTTL)
		e.cache.Delete(key)
	}()
}

// GetStats 获取统计信息
func (e *BannerEngine) GetStats() *EngineStats {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	
	stats := *e.stats
	return &stats
}

// GetRules 获取所有规则
func (e *BannerEngine) GetRules() []*Rule {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	
	rules := make([]*Rule, len(e.rules))
	copy(rules, e.rules)
	return rules
}

// GetRuleByID 根据ID获取规则
func (e *BannerEngine) GetRuleByID(id string) *Rule {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	
	for _, rule := range e.rules {
		if rule.ID == id {
			return rule
		}
	}
	return nil
}

// RemoveRule 删除规则
func (e *BannerEngine) RemoveRule(id string) bool {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	for i, rule := range e.rules {
		if rule.ID == id {
			e.rules = append(e.rules[:i], e.rules[i+1:]...)
			e.stats.TotalRules = len(e.rules)
			return true
		}
	}
	return false
}

// ClearCache 清空缓存
func (e *BannerEngine) ClearCache() {
	e.cache = &sync.Map{}
}

// GetBestMatch 获取最佳匹配
func (e *BannerEngine) GetBestMatch(banner string) *ServiceInfo {
	results := e.Match(banner)
	if len(results) > 0 {
		return results[0] // 已按置信度排序
	}
	return nil
}