package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// registerRedisScripts 注册Redis脚本
func (se *ScriptEngine) registerRedisScripts() {
	scripts := []*Script{
		{
			Name:        "redis-info",
			Protocol:    "redis",
			Category:    CategoryDiscovery,
			Description: "收集Redis服务器信息",
			Author:      "Script Engine Team",
			Version:     "1.0",
			Execute:     executeRedisInfo,
		},
		{
			Name:        "redis-config",
			Protocol:    "redis",
			Category:    CategoryDiscovery,
			Description: "获取Redis配置信息",
			Author:      "Script Engine Team",
			Version:     "1.0",
			Execute:     executeRedisConfig,
		},
		{
			Name:        "redis-keys-enum",
			Protocol:    "redis",
			Category:    CategoryDiscovery,
			Description: "枚举Redis键值",
			Author:      "Script Engine Team",
			Version:     "1.0",
			Execute:     executeRedisKeysEnum,
		},
		{
			Name:        "redis-auth-bypass",
			Protocol:    "redis",
			Category:    CategoryVulnerability,
			Description: "检测Redis认证绕过",
			Author:      "Script Engine Team",
			Version:     "1.0",
			Execute:     executeRedisAuthBypass,
		},
		{
			Name:        "redis-rce-check",
			Protocol:    "redis",
			Category:    CategoryVulnerability,
			Description: "检测Redis远程代码执行漏洞",
			Author:      "Script Engine Team",
			Version:     "1.0",
			Execute:     executeRedisRCECheck,
		},
		{
			Name:        "redis-brute-auth",
			Protocol:    "redis",
			Category:    CategoryAuthentication,
			Description: "Redis密码暴力破解",
			Author:      "Script Engine Team",
			Version:     "1.0",
			Execute:     executeRedisBruteAuth,
		},
	}

	for _, script := range scripts {
		se.registry.Register(script)
	}
}

// executeRedisInfo 执行Redis信息收集
func executeRedisInfo(target Target, ctx *ScriptContext) *ScriptResult {
	result := &ScriptResult{
		Success:  false,
		Findings: make(map[string]interface{}),
	}

	ctx.Logger.Debug("开始收集Redis服务器信息")

	// 连接到Redis
	conn, err := net.DialTimeout("tcp", target.String(), ctx.Timeout)
	if err != nil {
		result.Error = fmt.Sprintf("连接失败: %v", err)
		return result
	}
	defer conn.Close()

	// 发送INFO命令
	_, err = conn.Write([]byte("INFO\r\n"))
	if err != nil {
		result.Error = fmt.Sprintf("发送INFO命令失败: %v", err)
		return result
	}

	// 读取响应
	conn.SetReadDeadline(time.Now().Add(ctx.Timeout))
	scanner := bufio.NewScanner(conn)
	
	var response strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		response.WriteString(line + "\n")
		
		// Redis响应以单独的点结束
		if line == "." {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		result.Error = fmt.Sprintf("读取响应失败: %v", err)
		return result
	}

	responseText := response.String()
	
	// 检查是否需要认证
	if strings.Contains(responseText, "NOAUTH") {
		result.Findings["auth_required"] = true
		result.Findings["auth_error"] = "Authentication required"
		result.Success = true
		return result
	}

	// 解析INFO响应
	info := parseRedisInfo(responseText)
	result.Findings = info
	result.Success = true

	ctx.Logger.Debug("成功获取Redis服务器信息")
	return result
}

// parseRedisInfo 解析Redis INFO响应
func parseRedisInfo(response string) map[string]interface{} {
	info := make(map[string]interface{})
	sections := make(map[string]map[string]string)
	
	lines := strings.Split(response, "\n")
	currentSection := "general"
	sections[currentSection] = make(map[string]string)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			// 检查是否是节标题
			if strings.HasPrefix(line, "# ") {
				sectionName := strings.ToLower(strings.TrimPrefix(line, "# "))
				currentSection = sectionName
				sections[currentSection] = make(map[string]string)
			}
			continue
		}
		
		// 解析键值对
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				sections[currentSection][key] = value
				
				// 提取关键信息到顶层
				switch key {
				case "redis_version":
					info["version"] = value
				case "redis_mode":
					info["mode"] = value
				case "os":
					info["os"] = value
				case "arch_bits":
					info["architecture"] = value + "-bit"
				case "process_id":
					info["pid"] = value
				case "tcp_port":
					info["port"] = value
				case "uptime_in_seconds":
					info["uptime"] = value + " seconds"
				case "connected_clients":
					info["clients"] = value
				case "used_memory_human":
					info["memory_used"] = value
				case "maxmemory_human":
					info["memory_max"] = value
				case "role":
					info["role"] = value
				}
			}
		}
	}
	
	info["sections"] = sections
	return info
}

// executeRedisConfig 执行Redis配置获取
func executeRedisConfig(target Target, ctx *ScriptContext) *ScriptResult {
	result := &ScriptResult{
		Success:  false,
		Findings: make(map[string]interface{}),
	}

	ctx.Logger.Debug("开始获取Redis配置信息")

	conn, err := net.DialTimeout("tcp", target.String(), ctx.Timeout)
	if err != nil {
		result.Error = fmt.Sprintf("连接失败: %v", err)
		return result
	}
	defer conn.Close()

	// 发送CONFIG GET *命令
	_, err = conn.Write([]byte("CONFIG GET *\r\n"))
	if err != nil {
		result.Error = fmt.Sprintf("发送CONFIG命令失败: %v", err)
		return result
	}

	// 读取响应
	conn.SetReadDeadline(time.Now().Add(ctx.Timeout))
	response, err := readRedisResponse(conn)
	if err != nil {
		result.Error = fmt.Sprintf("读取配置失败: %v", err)
		return result
	}

	// 解析配置
	config := parseRedisConfig(response)
	result.Findings["config"] = config
	
	// 检查安全配置
	securityIssues := checkRedisSecurityConfig(config)
	if len(securityIssues) > 0 {
		result.Findings["security_issues"] = securityIssues
	}

	result.Success = true
	ctx.Logger.Debug("成功获取Redis配置信息")
	return result
}

// readRedisResponse 读取Redis响应
func readRedisResponse(conn net.Conn) (string, error) {
	var response strings.Builder
	scanner := bufio.NewScanner(conn)
	
	for scanner.Scan() {
		line := scanner.Text()
		response.WriteString(line + "\n")
		
		// 简单的响应结束检测
		if strings.HasPrefix(line, "+OK") || strings.HasPrefix(line, "-ERR") {
			break
		}
	}
	
	return response.String(), scanner.Err()
}

// parseRedisConfig 解析Redis配置
func parseRedisConfig(response string) map[string]string {
	config := make(map[string]string)
	lines := strings.Split(response, "\n")
	
	for i := 0; i < len(lines)-1; i += 2 {
		if i+1 < len(lines) {
			key := strings.TrimSpace(lines[i])
			value := strings.TrimSpace(lines[i+1])
			
			// 移除Redis协议前缀
			key = strings.TrimPrefix(key, "$")
			value = strings.TrimPrefix(value, "$")
			
			if key != "" && value != "" {
				config[key] = value
			}
		}
	}
	
	return config
}

// checkRedisSecurityConfig 检查Redis安全配置
func checkRedisSecurityConfig(config map[string]string) []string {
	var issues []string
	
	// 检查是否设置了密码
	if requirepass, exists := config["requirepass"]; !exists || requirepass == "" {
		issues = append(issues, "No password authentication configured")
	}
	
	// 检查是否绑定到所有接口
	if bind, exists := config["bind"]; !exists || bind == "0.0.0.0" {
		issues = append(issues, "Redis bound to all interfaces (0.0.0.0)")
	}
	
	// 检查是否启用了保护模式
	if protected, exists := config["protected-mode"]; !exists || protected == "no" {
		issues = append(issues, "Protected mode is disabled")
	}
	
	// 检查危险命令是否被重命名或禁用
	dangerousCommands := []string{"FLUSHDB", "FLUSHALL", "CONFIG", "EVAL", "DEBUG"}
	for _, cmd := range dangerousCommands {
		if renamed, exists := config["rename-command "+cmd]; !exists || renamed != "" {
			issues = append(issues, fmt.Sprintf("Dangerous command %s is not disabled", cmd))
		}
	}
	
	return issues
}

// executeRedisKeysEnum 执行Redis键值枚举
func executeRedisKeysEnum(target Target, ctx *ScriptContext) *ScriptResult {
	result := &ScriptResult{
		Success:  false,
		Findings: make(map[string]interface{}),
	}

	ctx.Logger.Debug("开始枚举Redis键值")

	conn, err := net.DialTimeout("tcp", target.String(), ctx.Timeout)
	if err != nil {
		result.Error = fmt.Sprintf("连接失败: %v", err)
		return result
	}
	defer conn.Close()

	// 发送KEYS *命令 (限制数量避免性能问题)
	_, err = conn.Write([]byte("KEYS *\r\n"))
	if err != nil {
		result.Error = fmt.Sprintf("发送KEYS命令失败: %v", err)
		return result
	}

	// 读取响应
	conn.SetReadDeadline(time.Now().Add(ctx.Timeout))
	response, err := readRedisResponse(conn)
	if err != nil {
		result.Error = fmt.Sprintf("读取键列表失败: %v", err)
		return result
	}

	// 解析键列表
	keys := parseRedisKeys(response)
	result.Findings["keys"] = keys
	result.Findings["key_count"] = len(keys)
	
	// 分析键模式
	patterns := analyzeKeyPatterns(keys)
	result.Findings["key_patterns"] = patterns
	
	// 获取一些键的值 (限制数量)
	if len(keys) > 0 {
		sampleKeys := keys
		if len(keys) > 10 {
			sampleKeys = keys[:10]
		}
		
		keyValues := make(map[string]interface{})
		for _, key := range sampleKeys {
			value, err := getRedisKeyValue(conn, key)
			if err == nil {
				keyValues[key] = value
			}
		}
		result.Findings["sample_values"] = keyValues
	}

	result.Success = true
	ctx.Logger.Debug("成功枚举 %d 个Redis键", len(keys))
	return result
}

// parseRedisKeys 解析Redis键列表
func parseRedisKeys(response string) []string {
	var keys []string
	lines := strings.Split(response, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "*") && !strings.HasPrefix(line, "$") {
			keys = append(keys, line)
		}
	}
	
	return keys
}

// analyzeKeyPatterns 分析键模式
func analyzeKeyPatterns(keys []string) map[string]int {
	patterns := make(map[string]int)
	
	for _, key := range keys {
		// 简单的模式分析
		if strings.Contains(key, ":") {
			parts := strings.Split(key, ":")
			if len(parts) > 0 {
				prefix := parts[0]
				patterns[prefix+":*"]++
			}
		} else {
			patterns["simple"]++
		}
	}
	
	return patterns
}

// getRedisKeyValue 获取Redis键值
func getRedisKeyValue(conn net.Conn, key string) (string, error) {
	// 发送GET命令
	cmd := fmt.Sprintf("GET %s\r\n", key)
	_, err := conn.Write([]byte(cmd))
	if err != nil {
		return "", err
	}
	
	// 读取响应
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	response, err := readRedisResponse(conn)
	if err != nil {
		return "", err
	}
	
	// 简单解析值
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "$") && !strings.HasPrefix(line, "+") {
			return line, nil
		}
	}
	
	return "", fmt.Errorf("no value found")
}

// executeRedisAuthBypass 执行Redis认证绕过检测
func executeRedisAuthBypass(target Target, ctx *ScriptContext) *ScriptResult {
	result := &ScriptResult{
		Success:         false,
		Findings:        make(map[string]interface{}),
		Vulnerabilities: make([]Vulnerability, 0),
	}

	ctx.Logger.Debug("开始检测Redis认证绕过漏洞")

	conn, err := net.DialTimeout("tcp", target.String(), ctx.Timeout)
	if err != nil {
		result.Error = fmt.Sprintf("连接失败: %v", err)
		return result
	}
	defer conn.Close()

	// 测试1: 无认证访问
	ctx.Logger.Debug("测试无认证访问")
	
	_, err = conn.Write([]byte("INFO\r\n"))
	if err == nil {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		response, err := readRedisResponse(conn)
		if err == nil && !strings.Contains(response, "NOAUTH") {
			// 无需认证即可访问
			result.Findings["no_auth_required"] = true
			
			vuln := Vulnerability{
				CVE:         "CWE-306",
				Title:       "Missing Authentication for Critical Function",
				Description: "Redis服务器未启用认证，允许未授权访问",
				Severity:    SeverityHigh,
				CVSS:        7.5,
				ExploitAvailable: true,
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		} else {
			result.Findings["no_auth_required"] = false
		}
	}

	// 测试2: 默认密码
	ctx.Logger.Debug("测试默认密码")
	
	defaultPasswords := []string{"", "redis", "password", "123456", "admin"}
	for _, password := range defaultPasswords {
		authCmd := fmt.Sprintf("AUTH %s\r\n", password)
		_, err = conn.Write([]byte(authCmd))
		if err == nil {
			response, err := readRedisResponse(conn)
			if err == nil && strings.Contains(response, "+OK") {
				result.Findings["default_password"] = password
				
				vuln := Vulnerability{
					CVE:         "CWE-521",
					Title:       "Weak Password Requirements",
					Description: fmt.Sprintf("Redis使用弱密码: %s", password),
					Severity:    SeverityHigh,
					CVSS:        8.1,
					ExploitAvailable: true,
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
				break
			}
		}
	}

	// 测试3: 配置文件泄露
	ctx.Logger.Debug("测试配置访问")
	
	_, err = conn.Write([]byte("CONFIG GET *\r\n"))
	if err == nil {
		response, err := readRedisResponse(conn)
		if err == nil && !strings.Contains(response, "ERR") {
			result.Findings["config_accessible"] = true
			
			// 检查敏感配置
			if strings.Contains(response, "requirepass") {
				result.Findings["password_in_config"] = true
				
				vuln := Vulnerability{
					CVE:         "CWE-200",
					Title:       "Information Exposure",
					Description: "Redis配置信息可被未授权访问，可能泄露敏感信息",
					Severity:    SeverityMedium,
					CVSS:        5.3,
					ExploitAvailable: true,
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
			}
		}
	}

	result.Success = true
	ctx.Logger.Debug("Redis认证绕过检测完成，发现 %d 个漏洞", len(result.Vulnerabilities))
	return result
}

// executeRedisRCECheck 执行Redis远程代码执行检测
func executeRedisRCECheck(target Target, ctx *ScriptContext) *ScriptResult {
	result := &ScriptResult{
		Success:         false,
		Findings:        make(map[string]interface{}),
		Vulnerabilities: make([]Vulnerability, 0),
	}

	ctx.Logger.Debug("开始检测Redis远程代码执行漏洞")

	conn, err := net.DialTimeout("tcp", target.String(), ctx.Timeout)
	if err != nil {
		result.Error = fmt.Sprintf("连接失败: %v", err)
		return result
	}
	defer conn.Close()

	// 测试1: EVAL命令可用性
	ctx.Logger.Debug("测试EVAL命令")
	
	evalCmd := "EVAL \"return 'test'\" 0\r\n"
	_, err = conn.Write([]byte(evalCmd))
	if err == nil {
		response, err := readRedisResponse(conn)
		if err == nil && strings.Contains(response, "test") {
			result.Findings["eval_enabled"] = true
			
			vuln := Vulnerability{
				CVE:         "CWE-94",
				Title:       "Code Injection",
				Description: "Redis EVAL命令可用，可能允许Lua脚本注入",
				Severity:    SeverityHigh,
				CVSS:        8.1,
				ExploitAvailable: true,
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}

	// 测试2: 文件写入能力
	ctx.Logger.Debug("测试文件写入")
	
	// 尝试设置dir和dbfilename
	commands := []string{
		"CONFIG GET dir\r\n",
		"CONFIG GET dbfilename\r\n",
	}
	
	canWriteFile := true
	for _, cmd := range commands {
		_, err = conn.Write([]byte(cmd))
		if err != nil {
			canWriteFile = false
			break
		}
		
		response, err := readRedisResponse(conn)
		if err != nil || strings.Contains(response, "ERR") {
			canWriteFile = false
			break
		}
	}
	
	if canWriteFile {
		result.Findings["file_write_possible"] = true
		
		vuln := Vulnerability{
			CVE:         "CWE-22",
			Title:       "Path Traversal",
			Description: "Redis可配置文件路径，可能允许任意文件写入",
			Severity:    SeverityCritical,
			CVSS:        9.8,
			ExploitAvailable: true,
		}
		result.Vulnerabilities = append(result.Vulnerabilities, vuln)
	}

	// 测试3: 模块加载
	ctx.Logger.Debug("测试模块加载")
	
	moduleCmd := "MODULE LIST\r\n"
	_, err = conn.Write([]byte(moduleCmd))
	if err == nil {
		response, err := readRedisResponse(conn)
		if err == nil && !strings.Contains(response, "ERR") {
			result.Findings["modules_supported"] = true
			
			// 如果支持模块，这可能是一个RCE向量
			vuln := Vulnerability{
				CVE:         "CWE-829",
				Title:       "Inclusion of Functionality from Untrusted Control Sphere",
				Description: "Redis支持动态模块加载，可能被利用执行恶意代码",
				Severity:    SeverityHigh,
				CVSS:        7.2,
				ExploitAvailable: true,
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}

	result.Success = true
	ctx.Logger.Debug("Redis RCE检测完成，发现 %d 个漏洞", len(result.Vulnerabilities))
	return result
}

// executeRedisBruteAuth 执行Redis密码暴力破解
func executeRedisBruteAuth(target Target, ctx *ScriptContext) *ScriptResult {
	result := &ScriptResult{
		Success:  false,
		Findings: make(map[string]interface{}),
	}

	ctx.Logger.Debug("开始Redis密码暴力破解")

	conn, err := net.DialTimeout("tcp", target.String(), ctx.Timeout)
	if err != nil {
		result.Error = fmt.Sprintf("连接失败: %v", err)
		return result
	}
	defer conn.Close()

	// 常见密码字典
	passwords := []string{
		"", "redis", "password", "123456", "admin", "root", "test",
		"qwerty", "123123", "abc123", "password123", "admin123",
		"redis123", "letmein", "welcome", "guest", "user",
	}

	attemptCount := 0
	successPassword := ""

	for _, password := range passwords {
		attemptCount++
		ctx.Logger.Debug("尝试密码: %s", password)

		var authCmd string
		if password == "" {
			// 测试无密码
			authCmd = "PING\r\n"
		} else {
			authCmd = fmt.Sprintf("AUTH %s\r\n", password)
		}

		_, err = conn.Write([]byte(authCmd))
		if err != nil {
			continue
		}

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		response, err := readRedisResponse(conn)
		if err != nil {
			continue
		}

		if strings.Contains(response, "+OK") || strings.Contains(response, "+PONG") {
			successPassword = password
			break
		}

		// 添加延迟避免被检测
		time.Sleep(100 * time.Millisecond)
	}

	result.Findings["attempts"] = attemptCount
	result.Findings["passwords_tested"] = passwords[:attemptCount]

	if successPassword != "" {
		result.Findings["success"] = true
		result.Findings["password"] = successPassword
		
		if successPassword == "" {
			result.Findings["auth_method"] = "no_password"
		} else {
			result.Findings["auth_method"] = "weak_password"
		}
	} else {
		result.Findings["success"] = false
	}

	result.Success = true
	ctx.Logger.Debug("Redis暴力破解完成，尝试了 %d 个密码", attemptCount)
	return result
}