package main

import (
	"fmt"
	"sync"
	"time"
)

// ScriptEngine 脚本引擎
type ScriptEngine struct {
	config   *ScriptConfig
	registry *ScriptRegistry
	logger   Logger
	mutex    sync.RWMutex
}

// NewScriptEngine 创建脚本引擎
func NewScriptEngine(config *ScriptConfig) *ScriptEngine {
	if config == nil {
		config = &ScriptConfig{
			Timeout:      30 * time.Second,
			Concurrent:   10,
			Verbose:      false,
			OutputFormat: "text",
		}
	}

	engine := &ScriptEngine{
		config:   config,
		registry: NewScriptRegistry(),
		logger:   &SimpleLogger{Verbose: config.Verbose},
	}

	// 加载内置脚本
	engine.loadBuiltinScripts()

	return engine
}

// loadBuiltinScripts 加载内置脚本
func (se *ScriptEngine) loadBuiltinScripts() {
	se.logger.Info("正在加载内置脚本...")

	// 注册Modbus脚本
	se.registerModbusScripts()
	
	// 注册Redis脚本
	se.registerRedisScripts()
	
	// 注册MQTT脚本
	se.registerMQTTScripts()
	
	// 注册MySQL脚本
	se.registerMySQLScripts()
	
	// 注册Kerberos脚本
	se.registerKerberosScripts()

	se.logger.Info("已加载 %d 个脚本", se.registry.Count())
}

// ExecuteScripts 执行脚本
func (se *ScriptEngine) ExecuteScripts(target Target, protocol, scriptNames, category string) (*TargetResult, error) {
	start := time.Now()
	
	se.logger.Info("开始对目标 %s 执行脚本", target.String())
	
	// 如果未指定协议，尝试自动检测
	if protocol == "" {
		detectedProtocol, err := se.detectProtocol(target)
		if err != nil {
			return nil, fmt.Errorf("协议检测失败: %v", err)
		}
		protocol = detectedProtocol
		se.logger.Info("检测到协议: %s", protocol)
	}

	// 选择要执行的脚本
	scripts, err := se.selectScripts(protocol, scriptNames, category)
	if err != nil {
		return nil, err
	}

	if len(scripts) == 0 {
		return nil, fmt.Errorf("未找到匹配的脚本")
	}

	se.logger.Info("选择了 %d 个脚本执行", len(scripts))

	// 创建结果对象
	result := &TargetResult{
		Target:          target.String(),
		Protocol:        protocol,
		ScriptResults:   make([]*ScriptResult, 0),
		Findings:        make(map[string]interface{}),
		Vulnerabilities: make([]Vulnerability, 0),
		Timestamp:       time.Now(),
	}

	// 执行脚本
	scriptResults := se.executeScriptsConcurrent(target, scripts)
	result.ScriptResults = scriptResults

	// 合并结果
	se.mergeResults(result)

	result.Duration = time.Since(start)
	se.logger.Info("脚本执行完成，耗时: %v", result.Duration)

	return result, nil
}

// detectProtocol 检测协议
func (se *ScriptEngine) detectProtocol(target Target) (string, error) {
	// TODO: 集成network_probe的协议检测功能
	// 这里先返回一个简单的端口映射
	portProtocolMap := map[int]string{
		502:   "modbus",
		6379:  "redis",
		1883:  "mqtt",
		3306:  "mysql",
		88:    "kerberos",
		80:    "http",
		443:   "https",
		22:    "ssh",
		21:    "ftp",
		25:    "smtp",
		53:    "dns",
		161:   "snmp",
		389:   "ldap",
		1812:  "radius",
		123:   "ntp",
	}

	if protocol, exists := portProtocolMap[target.Port]; exists {
		return protocol, nil
	}

	return "", fmt.Errorf("无法检测端口 %d 的协议", target.Port)
}

// selectScripts 选择脚本
func (se *ScriptEngine) selectScripts(protocol, scriptNames, category string) ([]*Script, error) {
	var scripts []*Script

	// 解析脚本名称
	names := ParseScriptNames(scriptNames)

	if len(names) == 1 && names[0] == "all" {
		// 选择所有匹配的脚本
		if category != "" {
			scripts = se.registry.GetByProtocolAndCategory(protocol, category)
		} else {
			scripts = se.registry.GetByProtocol(protocol)
		}
	} else {
		// 选择指定的脚本
		for _, name := range names {
			script, exists := se.registry.Get(name)
			if !exists {
				return nil, fmt.Errorf("脚本 %s 不存在", name)
			}
			
			// 检查协议匹配
			if script.Protocol != protocol {
				se.logger.Warn("脚本 %s 的协议 %s 与目标协议 %s 不匹配", 
					name, script.Protocol, protocol)
				continue
			}
			
			// 检查类别匹配
			if category != "" && script.Category != category {
				continue
			}
			
			scripts = append(scripts, script)
		}
	}

	return scripts, nil
}

// executeScriptsConcurrent 并发执行脚本
func (se *ScriptEngine) executeScriptsConcurrent(target Target, scripts []*Script) []*ScriptResult {
	results := make([]*ScriptResult, len(scripts))
	
	// 创建工作池
	semaphore := make(chan struct{}, se.config.Concurrent)
	var wg sync.WaitGroup

	for i, script := range scripts {
		wg.Add(1)
		go func(index int, s *Script) {
			defer wg.Done()
			
			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			// 执行脚本
			result := se.executeScript(target, s)
			results[index] = result
		}(i, script)
	}

	wg.Wait()
	return results
}

// executeScript 执行单个脚本
func (se *ScriptEngine) executeScript(target Target, script *Script) *ScriptResult {
	start := time.Now()
	
	se.logger.Debug("执行脚本: %s", script.Name)
	
	// 创建脚本上下文
	ctx := &ScriptContext{
		Config:    se.config,
		Target:    target,
		Protocol:  script.Protocol,
		Timeout:   se.config.Timeout,
		Variables: make(map[string]interface{}),
		Logger:    se.logger,
	}

	// 执行脚本
	result := script.Execute(target, ctx)
	if result == nil {
		result = &ScriptResult{
			ScriptName: script.Name,
			Category:   script.Category,
			Success:    false,
			Error:      "脚本返回空结果",
		}
	}

	// 设置基本信息
	result.ScriptName = script.Name
	result.Category = script.Category
	result.Duration = time.Since(start)
	result.Timestamp = time.Now()

	if result.Success {
		se.logger.Debug("脚本 %s 执行成功，耗时: %v", script.Name, result.Duration)
	} else {
		se.logger.Debug("脚本 %s 执行失败: %s，耗时: %v", 
			script.Name, result.Error, result.Duration)
	}

	return result
}

// mergeResults 合并脚本结果
func (se *ScriptEngine) mergeResults(result *TargetResult) {
	for _, scriptResult := range result.ScriptResults {
		if !scriptResult.Success {
			continue
		}

		// 合并发现信息
		for key, value := range scriptResult.Findings {
			result.Findings[key] = value
		}

		// 合并漏洞信息
		result.Vulnerabilities = append(result.Vulnerabilities, scriptResult.Vulnerabilities...)
	}
}

// RegisterScript 注册脚本
func (se *ScriptEngine) RegisterScript(script *Script) error {
	se.mutex.Lock()
	defer se.mutex.Unlock()
	
	return se.registry.Register(script)
}

// GetScript 获取脚本
func (se *ScriptEngine) GetScript(name string) (*Script, bool) {
	se.mutex.RLock()
	defer se.mutex.RUnlock()
	
	return se.registry.Get(name)
}

// GetAllScripts 获取所有脚本
func (se *ScriptEngine) GetAllScripts() []*Script {
	se.mutex.RLock()
	defer se.mutex.RUnlock()
	
	return se.registry.GetAll()
}

// GetScriptsByProtocol 根据协议获取脚本
func (se *ScriptEngine) GetScriptsByProtocol(protocol string) []*Script {
	se.mutex.RLock()
	defer se.mutex.RUnlock()
	
	return se.registry.GetByProtocol(protocol)
}

// GetScriptsByCategory 根据类别获取脚本
func (se *ScriptEngine) GetScriptsByCategory(category string) []*Script {
	se.mutex.RLock()
	defer se.mutex.RUnlock()
	
	return se.registry.GetByCategory(category)
}

// GetScriptCount 获取脚本数量
func (se *ScriptEngine) GetScriptCount() int {
	se.mutex.RLock()
	defer se.mutex.RUnlock()
	
	return se.registry.Count()
}

// ValidateScript 验证脚本
func (se *ScriptEngine) ValidateScript(script *Script) error {
	if script.Name == "" {
		return fmt.Errorf("脚本名称不能为空")
	}
	
	if script.Protocol == "" {
		return fmt.Errorf("脚本协议不能为空")
	}
	
	if script.Category == "" {
		return fmt.Errorf("脚本类别不能为空")
	}
	
	if script.Execute == nil {
		return fmt.Errorf("脚本执行函数不能为空")
	}
	
	// 验证类别
	validCategories := []string{
		CategoryDiscovery,
		CategoryVulnerability,
		CategoryAuthentication,
		CategoryExploitation,
	}
	
	valid := false
	for _, cat := range validCategories {
		if script.Category == cat {
			valid = true
			break
		}
	}
	
	if !valid {
		return fmt.Errorf("无效的脚本类别: %s", script.Category)
	}
	
	return nil
}

// ReloadScripts 重新加载脚本
func (se *ScriptEngine) ReloadScripts() {
	se.mutex.Lock()
	defer se.mutex.Unlock()
	
	se.logger.Info("重新加载脚本...")
	
	// 清空注册表
	se.registry = NewScriptRegistry()
	
	// 重新加载内置脚本
	se.loadBuiltinScripts()
	
	se.logger.Info("脚本重新加载完成，共 %d 个脚本", se.registry.Count())
}

// GetEngineStats 获取引擎统计信息
func (se *ScriptEngine) GetEngineStats() map[string]interface{} {
	se.mutex.RLock()
	defer se.mutex.RUnlock()
	
	stats := make(map[string]interface{})
	
	// 基本统计
	stats["total_scripts"] = se.registry.Count()
	stats["config"] = se.config
	
	// 按协议统计
	protocolStats := make(map[string]int)
	categoryStats := make(map[string]int)
	
	for _, script := range se.registry.GetAll() {
		protocolStats[script.Protocol]++
		categoryStats[script.Category]++
	}
	
	stats["protocol_stats"] = protocolStats
	stats["category_stats"] = categoryStats
	
	return stats
}