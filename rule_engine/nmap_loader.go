package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// NmapLoader Nmap指纹库加载器
type NmapLoader struct {
	engine *BannerEngine
}

// NewNmapLoader 创建Nmap加载器
func NewNmapLoader(engine *BannerEngine) *NmapLoader {
	return &NmapLoader{
		engine: engine,
	}
}

// LoadFromFile 从Nmap指纹文件加载
func (nl *NmapLoader) LoadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	return nl.LoadFromReader(file)
}

// LoadFromReader 从Reader加载
func (nl *NmapLoader) LoadFromReader(reader io.Reader) error {
	rules, err := nl.parseNmapFile(reader)
	if err != nil {
		return err
	}

	return nl.engine.LoadRules(rules)
}

// parseNmapFile 解析Nmap指纹文件
func (nl *NmapLoader) parseNmapFile(reader io.Reader) ([]*Rule, error) {
	var rules []*Rule
	var currentProbe string
	
	scanner := bufio.NewScanner(reader)
	lineNum := 0
	
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// 解析Probe指令
		if strings.HasPrefix(line, "Probe ") {
			probe := nl.parseProbe(line)
			if probe != "" {
				currentProbe = probe
			}
			continue
		}
		
		// 解析match指令
		if strings.HasPrefix(line, "match ") {
			rule := nl.parseMatch(line, currentProbe, false)
			if rule != nil {
				rule.ID = fmt.Sprintf("nmap_%d", len(rules)+1)
				rules = append(rules, rule)
			}
			continue
		}
		
		// 解析softmatch指令
		if strings.HasPrefix(line, "softmatch ") {
			rule := nl.parseMatch(line, currentProbe, true)
			if rule != nil {
				rule.ID = fmt.Sprintf("nmap_soft_%d", len(rules)+1)
				rule.Confidence = rule.Confidence - 20 // 软匹配置信度降低
				if rule.Confidence < 50 {
					rule.Confidence = 50
				}
				rules = append(rules, rule)
			}
			continue
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}
	
	return rules, nil
}

// parseProbe 解析Probe指令
func (nl *NmapLoader) parseProbe(line string) string {
	// 格式: Probe TCP GetRequest q|GET / HTTP/1.0\r\n\r\n|
	parts := strings.Fields(line)
	if len(parts) >= 3 {
		return parts[2] // 返回探测名称
	}
	return ""
}

// parseMatch 解析match指令
func (nl *NmapLoader) parseMatch(line string, probe string, isSoft bool) *Rule {
	// 移除match或softmatch前缀
	content := line
	if strings.HasPrefix(line, "match ") {
		content = strings.TrimPrefix(line, "match ")
	} else if strings.HasPrefix(line, "softmatch ") {
		content = strings.TrimPrefix(line, "softmatch ")
	}
	
	// 解析服务名和模式
	parts := strings.Fields(content)
	if len(parts) < 2 {
		return nil
	}
	
	service := parts[0]
	patternStr := parts[1]
	
	// 解析正则表达式模式
	pattern := nl.parsePattern(patternStr)
	if pattern == "" {
		return nil
	}
	
	// 创建规则
	rule := &Rule{
		Service:     service,
		Pattern:     pattern,
		Confidence:  85,
		Description: fmt.Sprintf("Nmap %s detection", service),
		Author:      "nmap",
	}
	
	// 解析版本信息
	if len(parts) > 2 {
		versionInfo := strings.Join(parts[2:], " ")
		nl.parseVersionInfo(rule, versionInfo)
	}
	
	return rule
}

// parsePattern 解析正则表达式模式
func (nl *NmapLoader) parsePattern(patternStr string) string {
	// 处理 m|pattern|flags 或 m/pattern/flags 格式
	if len(patternStr) < 3 {
		return ""
	}
	
	if patternStr[0] == 'm' && len(patternStr) > 2 {
		delimiter := patternStr[1]
		content := patternStr[2:]
		
		// 找到结束分隔符
		endIndex := strings.LastIndex(content, string(delimiter))
		if endIndex > 0 {
			pattern := content[:endIndex]
			return nl.unescapePattern(pattern)
		}
	}
	
	return ""
}

// parseVersionInfo 解析版本信息
func (nl *NmapLoader) parseVersionInfo(rule *Rule, versionStr string) {
	// 解析 p/product/ v/version/ i/info/ h/hostname/ o/os/ d/device/ 格式
	re := regexp.MustCompile(`([pvihod])/([^/]*)/`)
	matches := re.FindAllStringSubmatch(versionStr, -1)
	
	for _, match := range matches {
		if len(match) == 3 {
			key := match[1]
			value := match[2]
			
			switch key {
			case "p":
				rule.Product = value
			case "v":
				rule.Version = value
			case "i":
				rule.Info = value
			case "h":
				rule.Hostname = value
			case "o":
				rule.OS = value
			case "d":
				rule.DeviceType = value
			}
		}
	}
}

// unescapePattern 处理转义字符
func (nl *NmapLoader) unescapePattern(pattern string) string {
	// 处理常见的转义字符
	pattern = strings.ReplaceAll(pattern, "\\r", "\r")
	pattern = strings.ReplaceAll(pattern, "\\n", "\n")
	pattern = strings.ReplaceAll(pattern, "\\t", "\t")
	pattern = strings.ReplaceAll(pattern, "\\0", "\x00")
	pattern = strings.ReplaceAll(pattern, "\\\\", "\\")
	
	// 处理十六进制转义 \x##
	hexRe := regexp.MustCompile(`\\x([0-9a-fA-F]{2})`)
	pattern = hexRe.ReplaceAllStringFunc(pattern, func(match string) string {
		hex := match[2:]
		if val, err := strconv.ParseInt(hex, 16, 8); err == nil {
			return string(byte(val))
		}
		return match
	})
	
	return pattern
}

// LoadBuiltinRules 加载内置规则
func (nl *NmapLoader) LoadBuiltinRules() error {
	rules := []*Rule{
		{
			ID:          "nginx",
			Service:     "http",
			Pattern:     `(?i)nginx[/\s]+(\d+\.\d+\.\d+)`,
			Product:     "nginx",
			Version:     "$1",
			Confidence:  90,
			Description: "Nginx Web Server",
			Author:      "builtin",
		},
		{
			ID:          "apache",
			Service:     "http",
			Pattern:     `(?i)Apache[/\s]+(\d+\.\d+\.\d+)`,
			Product:     "Apache httpd",
			Version:     "$1",
			Confidence:  90,
			Description: "Apache HTTP Server",
			Author:      "builtin",
		},
		{
			ID:          "openssh",
			Service:     "ssh",
			Pattern:     `SSH-([.\d]+)-OpenSSH[_\s]+(\S+)`,
			Product:     "OpenSSH",
			Version:     "$2",
			Info:        "protocol $1",
			Confidence:  95,
			Description: "OpenSSH Server",
			Author:      "builtin",
		},
		{
			ID:          "mysql",
			Service:     "mysql",
			Pattern:     `(\d+\.\d+\.\d+).*mysql`,
			Product:     "MySQL",
			Version:     "$1",
			Confidence:  90,
			Description: "MySQL Database Server",
			Author:      "builtin",
		},
		{
			ID:          "ftp_vsftpd",
			Service:     "ftp",
			Pattern:     `220.*vsftpd\s+(\S+)`,
			Product:     "vsftpd",
			Version:     "$1",
			Confidence:  95,
			Description: "vsftpd FTP Server",
			Author:      "builtin",
		},
		{
			ID:          "smtp_postfix",
			Service:     "smtp",
			Pattern:     `220.*Postfix`,
			Product:     "Postfix",
			Confidence:  85,
			Description: "Postfix SMTP Server",
			Author:      "builtin",
		},
		{
			ID:          "redis",
			Service:     "redis",
			Pattern:     `\+PONG\r?\n`,
			Product:     "Redis",
			Confidence:  95,
			Description: "Redis Key-Value Store",
			Author:      "builtin",
		},
		{
			ID:          "iis",
			Service:     "http",
			Pattern:     `Microsoft-IIS[/\s]+(\d+\.\d+)`,
			Product:     "Microsoft IIS httpd",
			Version:     "$1",
			Confidence:  90,
			Description: "Microsoft IIS Web Server",
			Author:      "builtin",
		},
	}
	
	return nl.engine.LoadRules(rules)
}