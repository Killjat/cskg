package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// RuleManager 规则管理器
type RuleManager struct {
	engine   *BannerEngine
	rulesDir string
}

// NewRuleManager 创建规则管理器
func NewRuleManager(engine *BannerEngine, rulesDir string) *RuleManager {
	return &RuleManager{
		engine:   engine,
		rulesDir: rulesDir,
	}
}

// SaveRule 保存规则到文件
func (rm *RuleManager) SaveRule(rule *Rule) error {
	// 确保规则目录存在
	if err := os.MkdirAll(rm.rulesDir, 0755); err != nil {
		return fmt.Errorf("创建规则目录失败: %v", err)
	}
	
	// 生成文件名
	filename := fmt.Sprintf("%s.json", rule.ID)
	filepath := filepath.Join(rm.rulesDir, filename)
	
	// 序列化规则
	data, err := json.MarshalIndent(rule, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化规则失败: %v", err)
	}
	
	// 写入文件
	if err := ioutil.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("写入规则文件失败: %v", err)
	}
	
	return nil
}

// LoadRulesFromDir 从目录加载所有规则
func (rm *RuleManager) LoadRulesFromDir() error {
	if _, err := os.Stat(rm.rulesDir); os.IsNotExist(err) {
		return nil // 目录不存在，跳过
	}
	
	files, err := ioutil.ReadDir(rm.rulesDir)
	if err != nil {
		return fmt.Errorf("读取规则目录失败: %v", err)
	}
	
	var rules []*Rule
	
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		ext := filepath.Ext(file.Name())
		if ext != ".json" && ext != ".yaml" && ext != ".yml" {
			continue
		}
		
		filePath := filepath.Join(rm.rulesDir, file.Name())
		rule, err := rm.loadRuleFromFile(filePath)
		if err != nil {
			fmt.Printf("警告: 加载规则文件 %s 失败: %v\n", file.Name(), err)
			continue
		}
		
		if rule != nil {
			rules = append(rules, rule)
		}
	}
	
	if len(rules) > 0 {
		return rm.engine.LoadRules(rules)
	}
	
	return nil
}

// loadRuleFromFile 从文件加载单个规则
func (rm *RuleManager) loadRuleFromFile(filepath string) (*Rule, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	
	var rule Rule
	ext := filepath.Ext(filepath)
	
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &rule)
	default:
		return nil, fmt.Errorf("不支持的文件格式: %s (目前只支持JSON)", ext)
	}
	
	if err != nil {
		return nil, err
	}
	
	return &rule, nil
}

// SaveSimpleRule 保存简化规则
func (rm *RuleManager) SaveSimpleRule(simpleRule *SimpleRule) error {
	// 转换为完整规则
	rule := &Rule{
		ID:          fmt.Sprintf("user_%d", time.Now().UnixNano()),
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
		rule.Confidence = 80
	}
	
	// 添加到引擎
	if err := rm.engine.AddRule(rule); err != nil {
		return err
	}
	
	// 保存到文件
	return rm.SaveRule(rule)
}

// DeleteRule 删除规则
func (rm *RuleManager) DeleteRule(ruleID string) error {
	// 从引擎中删除
	if !rm.engine.RemoveRule(ruleID) {
		return fmt.Errorf("规则 %s 不存在", ruleID)
	}
	
	// 删除文件
	filename := fmt.Sprintf("%s.json", ruleID)
	filepath := filepath.Join(rm.rulesDir, filename)
	
	if _, err := os.Stat(filepath); err == nil {
		return os.Remove(filepath)
	}
	
	return nil
}

// ExportRules 导出所有规则
func (rm *RuleManager) ExportRules(filename string) error {
	rules := rm.engine.GetRules()
	
	ruleSet := &RuleSet{
		Version:     "1.0",
		Description: "Exported rules",
		Author:      "rule_manager",
		Rules:       rules,
	}
	
	var data []byte
	var err error
	
	ext := filepath.Ext(filename)
	switch ext {
	case ".json":
		data, err = json.MarshalIndent(ruleSet, "", "  ")
	default:
		return fmt.Errorf("不支持的导出格式: %s (目前只支持JSON)", ext)
	}
	
	if err != nil {
		return fmt.Errorf("序列化规则失败: %v", err)
	}
	
	return ioutil.WriteFile(filename, data, 0644)
}

// ImportRules 导入规则
func (rm *RuleManager) ImportRules(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}
	
	var ruleSet RuleSet
	ext := filepath.Ext(filename)
	
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &ruleSet)
	default:
		return fmt.Errorf("不支持的文件格式: %s (目前只支持JSON)", ext)
	}
	
	if err != nil {
		return fmt.Errorf("解析文件失败: %v", err)
	}
	
	// 添加规则到引擎
	for _, rule := range ruleSet.Rules {
		if err := rm.engine.AddRule(rule); err != nil {
			fmt.Printf("警告: 添加规则 %s 失败: %v\n", rule.ID, err)
			continue
		}
		
		// 保存到文件
		if err := rm.SaveRule(rule); err != nil {
			fmt.Printf("警告: 保存规则 %s 失败: %v\n", rule.ID, err)
		}
	}
	
	return nil
}

// CreateSimpleRuleTemplate 创建简化规则模板
func (rm *RuleManager) CreateSimpleRuleTemplate(filename string) error {
	template := &SimpleRule{
		Service:     "example_service",
		Pattern:     `(?i)example[/\s]+(\d+\.\d+\.\d+)`,
		Product:     "Example Product",
		Version:     "$1",
		Description: "Example service detection rule",
		Confidence:  80,
	}
	
	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化模板失败: %v", err)
	}
	
	return ioutil.WriteFile(filename, data, 0644)
}

// ValidateRule 验证规则
func (rm *RuleManager) ValidateRule(rule *Rule) error {
	if rule.Service == "" {
		return fmt.Errorf("服务名称不能为空")
	}
	
	if rule.Pattern == "" {
		return fmt.Errorf("匹配模式不能为空")
	}
	
	// 验证正则表达式
	if _, err := regexp.Compile(rule.Pattern); err != nil {
		return fmt.Errorf("正则表达式无效: %v", err)
	}
	
	if rule.Confidence < 0 || rule.Confidence > 100 {
		return fmt.Errorf("置信度必须在0-100之间")
	}
	
	return nil
}

// GetRuleStats 获取规则统计
func (rm *RuleManager) GetRuleStats() map[string]interface{} {
	rules := rm.engine.GetRules()
	
	stats := map[string]interface{}{
		"total_rules": len(rules),
		"by_service":  make(map[string]int),
		"by_author":   make(map[string]int),
	}
	
	serviceCount := make(map[string]int)
	authorCount := make(map[string]int)
	
	for _, rule := range rules {
		serviceCount[rule.Service]++
		authorCount[rule.Author]++
	}
	
	stats["by_service"] = serviceCount
	stats["by_author"] = authorCount
	
	return stats
}