package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// HTTPParser HTTP协议解析器
type HTTPParser struct{}

func (p *HTTPParser) Parse(data []byte) (*ParsedInfo, error) {
	content := string(data)
	info := &ParsedInfo{
		Protocol:   "http",
		Fields:     make(map[string]string),
		Confidence: 70,
	}
	
	// 解析HTTP响应
	lines := strings.Split(content, "\r\n")
	if len(lines) == 0 {
		lines = strings.Split(content, "\n")
	}
	
	if len(lines) > 0 {
		// 解析状态行: HTTP/1.1 200 OK
		statusLine := strings.TrimSpace(lines[0])
		info.Fields["status_line"] = statusLine
		
		// 提取HTTP版本和状态码
		statusRe := regexp.MustCompile(`HTTP/(\d+\.\d+)\s+(\d+)\s*(.*)`)
		if match := statusRe.FindStringSubmatch(statusLine); len(match) > 3 {
			info.Fields["http_version"] = match[1]
			info.Fields["status_code"] = match[2]
			info.Fields["status_text"] = strings.TrimSpace(match[3])
		}
		
		// 解析HTTP头部
		headers := make(map[string]string)
		for i := 1; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				break // 头部结束
			}
			
			if colonIdx := strings.Index(line, ":"); colonIdx > 0 {
				key := strings.ToLower(strings.TrimSpace(line[:colonIdx]))
				value := strings.TrimSpace(line[colonIdx+1:])
				headers[key] = value
				info.Fields["header_"+key] = value
			}
		}
		
		// 深度解析服务器信息
		if server, exists := headers["server"]; exists {
			p.parseServerHeader(server, info)
		}
		
		// 解析其他重要头部
		if contentType, exists := headers["content-type"]; exists {
			info.Fields["content_type"] = contentType
		}
		
		if powered, exists := headers["x-powered-by"]; exists {
			info.Fields["powered_by"] = powered
			p.parsePoweredBy(powered, info)
		}
		
		// 检测Web框架和技术栈
		p.detectWebTechnology(headers, info)
	}
	
	info.Service = "http"
	return info, nil
}

// parseServerHeader 深度解析Server头部
func (p *HTTPParser) parseServerHeader(server string, info *ParsedInfo) {
	server = strings.TrimSpace(server)
	info.Fields["server"] = server
	
	// nginx解析
	if nginxRe := regexp.MustCompile(`nginx/(\d+\.\d+(?:\.\d+)?)`); nginxRe.MatchString(server) {
		info.Product = "nginx"
		if match := nginxRe.FindStringSubmatch(server); len(match) > 1 {
			info.Version = match[1]
		}
		info.Confidence = 95
		
		// 检测nginx模块
		if strings.Contains(server, "Ubuntu") {
			info.OS = "Ubuntu"
		}
	} else if apacheRe := regexp.MustCompile(`Apache/(\d+\.\d+(?:\.\d+)?)`); apacheRe.MatchString(server) {
		info.Product = "Apache httpd"
		if match := apacheRe.FindStringSubmatch(server); len(match) > 1 {
			info.Version = match[1]
		}
		info.Confidence = 95
		
		// 检测操作系统
		if strings.Contains(server, "Ubuntu") {
			info.OS = "Ubuntu"
		} else if strings.Contains(server, "CentOS") {
			info.OS = "CentOS"
		} else if strings.Contains(server, "Win32") || strings.Contains(server, "Win64") {
			info.OS = "Windows"
		}
	} else if iisRe := regexp.MustCompile(`Microsoft-IIS/(\d+\.\d+)`); iisRe.MatchString(server) {
		info.Product = "Microsoft IIS"
		if match := iisRe.FindStringSubmatch(server); len(match) > 1 {
			info.Version = match[1]
		}
		info.OS = "Windows"
		info.Confidence = 95
	} else if strings.Contains(strings.ToLower(server), "gunicorn") {
		// gunicorn/19.9.0
		info.Product = "Gunicorn"
		if versionRe := regexp.MustCompile(`gunicorn/(\d+\.\d+\.\d+)`); versionRe.MatchString(server) {
			if match := versionRe.FindStringSubmatch(server); len(match) > 1 {
				info.Version = match[1]
			}
		}
		info.ExtraInfo = "Python WSGI HTTP Server"
		info.Confidence = 90
	} else if strings.Contains(strings.ToLower(server), "cloudflare") {
		info.Product = "Cloudflare"
		info.ExtraInfo = "CDN/Proxy"
		info.Confidence = 85
	}
}

// parsePoweredBy 解析X-Powered-By头部
func (p *HTTPParser) parsePoweredBy(powered string, info *ParsedInfo) {
	powered = strings.ToLower(powered)
	
	if strings.Contains(powered, "php") {
		info.ExtraInfo = "PHP"
		if phpRe := regexp.MustCompile(`php/(\d+\.\d+(?:\.\d+)?)`); phpRe.MatchString(powered) {
			if match := phpRe.FindStringSubmatch(powered); len(match) > 1 {
				info.Fields["php_version"] = match[1]
			}
		}
	} else if strings.Contains(powered, "asp.net") {
		info.ExtraInfo = "ASP.NET"
		info.OS = "Windows"
	} else if strings.Contains(powered, "express") {
		info.ExtraInfo = "Node.js Express"
	}
}

// detectWebTechnology 检测Web技术栈
func (p *HTTPParser) detectWebTechnology(headers map[string]string, info *ParsedInfo) {
	// 检测常见的技术指纹
	technologies := []string{}
	
	// 通过特殊头部检测
	if _, exists := headers["x-aspnet-version"]; exists {
		technologies = append(technologies, "ASP.NET")
		info.OS = "Windows"
	}
	
	if _, exists := headers["x-powered-by-plesk"]; exists {
		technologies = append(technologies, "Plesk")
	}
	
	if _, exists := headers["x-drupal-cache"]; exists {
		technologies = append(technologies, "Drupal")
	}
	
	if _, exists := headers["x-generator"]; exists {
		technologies = append(technologies, headers["x-generator"])
	}
	
	// 通过Set-Cookie检测
	if cookie, exists := headers["set-cookie"]; exists {
		if strings.Contains(strings.ToLower(cookie), "phpsessid") {
			technologies = append(technologies, "PHP")
		} else if strings.Contains(strings.ToLower(cookie), "jsessionid") {
			technologies = append(technologies, "Java/JSP")
		} else if strings.Contains(strings.ToLower(cookie), "asp.net_sessionid") {
			technologies = append(technologies, "ASP.NET")
		}
	}
	
	if len(technologies) > 0 {
		info.Fields["technologies"] = strings.Join(technologies, ", ")
	}
}

func (p *HTTPParser) GetProtocol() string { return "http" }
func (p *HTTPParser) GetConfidence(data []byte) int {
	content := string(data)
	if strings.Contains(content, "HTTP/") {
		return 95
	}
	return 0
}

// SSHParser SSH协议解析器
type SSHParser struct{}

func (p *SSHParser) Parse(data []byte) (*ParsedInfo, error) {
	content := strings.TrimSpace(string(data))
	info := &ParsedInfo{
		Protocol:   "ssh",
		Service:    "ssh",
		Fields:     make(map[string]string),
		Confidence: 95,
	}
	
	// SSH版本字符串格式: SSH-protoversion-softwareversion SP comments CR LF
	sshRe := regexp.MustCompile(`SSH-([.\d]+)-(.+?)(?:\s(.*))?(?:\r|\n|$)`)
	if match := sshRe.FindStringSubmatch(content); len(match) >= 3 {
		protocolVersion := match[1]
		softwareVersion := match[2]
		comments := ""
		if len(match) > 3 {
			comments = strings.TrimSpace(match[3])
		}
		
		info.Fields["protocol_version"] = protocolVersion
		info.Fields["software_version"] = softwareVersion
		if comments != "" {
			info.Fields["comments"] = comments
		}
		
		// 深度解析软件版本
		p.parseSoftwareVersion(softwareVersion, comments, info)
	}
	
	return info, nil
}

// parseSoftwareVersion 深度解析SSH软件版本
func (p *SSHParser) parseSoftwareVersion(software, comments string, info *ParsedInfo) {
	software = strings.TrimSpace(software)
	
	// OpenSSH解析
	if opensshRe := regexp.MustCompile(`OpenSSH[_\s]+(\d+\.\d+(?:p\d+)?)`); opensshRe.MatchString(software) {
		info.Product = "OpenSSH"
		if match := opensshRe.FindStringSubmatch(software); len(match) > 1 {
			info.Version = match[1]
		}
		
		// 检测操作系统
		if strings.Contains(software, "Ubuntu") {
			info.OS = "Ubuntu"
			// 提取Ubuntu版本: OpenSSH_8.2p1 Ubuntu-4ubuntu0.5
			if ubuntuRe := regexp.MustCompile(`Ubuntu-(\d+)ubuntu`); ubuntuRe.MatchString(software) {
				if match := ubuntuRe.FindStringSubmatch(software); len(match) > 1 {
					info.Fields["ubuntu_package"] = match[1]
				}
			}
		} else if strings.Contains(software, "Debian") {
			info.OS = "Debian"
		} else if strings.Contains(software, "CentOS") || strings.Contains(software, "Red Hat") {
			info.OS = "CentOS/RHEL"
		} else if strings.Contains(software, "FreeBSD") {
			info.OS = "FreeBSD"
		}
		
		info.Confidence = 98
		
	} else if strings.Contains(strings.ToLower(software), "dropbear") {
		// Dropbear SSH
		info.Product = "Dropbear SSH"
		if dropbearRe := regexp.MustCompile(`dropbear[_\s]+(\d+\.\d+)`); dropbearRe.MatchString(strings.ToLower(software)) {
			if match := dropbearRe.FindStringSubmatch(strings.ToLower(software)); len(match) > 1 {
				info.Version = match[1]
			}
		}
		info.ExtraInfo = "Lightweight SSH server"
		info.DeviceType = "embedded"
		info.Confidence = 95
		
	} else if strings.Contains(strings.ToLower(software), "libssh") {
		// libssh
		info.Product = "libssh"
		if libsshRe := regexp.MustCompile(`libssh[_\s]+(\d+\.\d+\.\d+)`); libsshRe.MatchString(strings.ToLower(software)) {
			if match := libsshRe.FindStringSubmatch(strings.ToLower(software)); len(match) > 1 {
				info.Version = match[1]
			}
		}
		info.ExtraInfo = "SSH library implementation"
		info.Confidence = 90
		
	} else if strings.Contains(strings.ToLower(software), "cisco") {
		// Cisco SSH
		info.Product = "Cisco SSH"
		info.DeviceType = "network device"
		info.ExtraInfo = "Cisco network equipment"
		info.Confidence = 95
		
	} else if strings.Contains(strings.ToLower(software), "paramiko") {
		// Paramiko (Python SSH)
		info.Product = "Paramiko"
		info.ExtraInfo = "Python SSH implementation"
		info.Confidence = 90
	}
	
	// 解析注释中的额外信息
	if comments != "" {
		p.parseComments(comments, info)
	}
}

// parseComments 解析SSH注释信息
func (p *SSHParser) parseComments(comments string, info *ParsedInfo) {
	comments = strings.ToLower(comments)
	
	// 检测蜜罐
	honeypotKeywords := []string{"honeypot", "cowrie", "kippo", "dionaea"}
	for _, keyword := range honeypotKeywords {
		if strings.Contains(comments, keyword) {
			info.ExtraInfo = "Possible honeypot"
			info.Fields["honeypot_indicator"] = keyword
			break
		}
	}
	
	// 检测云服务提供商
	if strings.Contains(comments, "aws") || strings.Contains(comments, "amazon") {
		info.Fields["cloud_provider"] = "AWS"
	} else if strings.Contains(comments, "azure") {
		info.Fields["cloud_provider"] = "Azure"
	} else if strings.Contains(comments, "gcp") || strings.Contains(comments, "google") {
		info.Fields["cloud_provider"] = "Google Cloud"
	}
}

func (p *SSHParser) GetProtocol() string { return "ssh" }
func (p *SSHParser) GetConfidence(data []byte) int {
	if strings.HasPrefix(string(data), "SSH-") {
		return 95
	}
	return 0
}

// FTPParser FTP协议解析器
type FTPParser struct{}

func (p *FTPParser) Parse(data []byte) (*ParsedInfo, error) {
	content := string(data)
	info := &ParsedInfo{
		Protocol:   "ftp",
		Service:    "ftp",
		Fields:     make(map[string]string),
		Confidence: 80,
	}
	
	// FTP响应格式: 3位数字码 + 空格 + 消息
	ftpRe := regexp.MustCompile(`^(\d{3})\s+(.+)`)
	if match := ftpRe.FindStringSubmatch(content); len(match) > 2 {
		info.Fields["response_code"] = match[1]
		info.Fields["message"] = match[2]
		
		// 解析vsftpd
		if strings.Contains(strings.ToLower(match[2]), "vsftpd") {
			info.Product = "vsftpd"
			if versionRe := regexp.MustCompile(`vsftpd\s+(\S+)`); versionRe.MatchString(match[2]) {
				if vMatch := versionRe.FindStringSubmatch(match[2]); len(vMatch) > 1 {
					info.Version = vMatch[1]
				}
			}
		}
	}
	
	return info, nil
}

func (p *FTPParser) GetProtocol() string { return "ftp" }
func (p *FTPParser) GetConfidence(data []byte) int {
	content := string(data)
	if regexp.MustCompile(`^2\d{2}\s`).MatchString(content) {
		return 80
	}
	return 0
}

// SMTPParser SMTP协议解析器
type SMTPParser struct{}

func (p *SMTPParser) Parse(data []byte) (*ParsedInfo, error) {
	content := string(data)
	info := &ParsedInfo{
		Protocol:   "smtp",
		Service:    "smtp",
		Fields:     make(map[string]string),
		Confidence: 80,
	}
	
	// SMTP响应格式类似FTP
	smtpRe := regexp.MustCompile(`^(\d{3})\s+(.+)`)
	if match := smtpRe.FindStringSubmatch(content); len(match) > 2 {
		info.Fields["response_code"] = match[1]
		info.Fields["message"] = match[2]
		
		// 解析Postfix
		if strings.Contains(strings.ToLower(match[2]), "postfix") {
			info.Product = "Postfix"
			info.Confidence = 90
		}
	}
	
	return info, nil
}

func (p *SMTPParser) GetProtocol() string { return "smtp" }
func (p *SMTPParser) GetConfidence(data []byte) int {
	content := string(data)
	if regexp.MustCompile(`^2\d{2}\s`).MatchString(content) && strings.Contains(content, "SMTP") {
		return 85
	}
	return 0
}

// MySQLParser MySQL协议解析器
type MySQLParser struct{}

func (p *MySQLParser) Parse(data []byte) (*ParsedInfo, error) {
	info := &ParsedInfo{
		Protocol:   "mysql",
		Service:    "mysql",
		Product:    "MySQL",
		Fields:     make(map[string]string),
		Confidence: 85,
	}
	
	if len(data) < 5 {
		return info, nil
	}
	
	// MySQL握手包解析
	// 格式: [packet_length:3][packet_number:1][protocol_version:1][server_version:string\0][thread_id:4]...
	
	// 跳过包长度(3字节)和序列号(1字节)
	if len(data) > 4 {
		protocolVersion := data[4]
		info.Fields["protocol_version"] = strconv.Itoa(int(protocolVersion))
		
		// MySQL协议版本通常是10
		if protocolVersion == 10 {
			info.Confidence = 95
		}
		
		// 解析服务器版本字符串
		if len(data) > 5 {
			versionStart := 5
			versionEnd := versionStart
			
			// 查找版本字符串结束符(\0)
			for i := versionStart; i < len(data) && i < versionStart+50; i++ {
				if data[i] == 0 {
					versionEnd = i
					break
				}
			}
			
			if versionEnd > versionStart {
				versionStr := string(data[versionStart:versionEnd])
				info.Fields["server_version"] = versionStr
				
				// 深度解析版本信息
				p.parseVersionString(versionStr, info)
				
				// 继续解析握手包的其他字段
				if versionEnd+1 < len(data) {
					p.parseHandshakePacket(data[versionEnd+1:], info)
				}
			}
		}
	}
	
	return info, nil
}

// parseVersionString 深度解析MySQL版本字符串
func (p *MySQLParser) parseVersionString(version string, info *ParsedInfo) {
	// MySQL版本格式示例:
	// 5.7.35-log
	// 8.0.27-0ubuntu0.20.04.1
	// 5.6.51-cll-lve
	// 10.3.32-MariaDB-0ubuntu0.20.04.1
	
	// 提取主版本号
	if versionRe := regexp.MustCompile(`^(\d+\.\d+\.\d+)`); versionRe.MatchString(version) {
		if match := versionRe.FindStringSubmatch(version); len(match) > 1 {
			info.Version = match[1]
		}
	}
	
	// 检测MySQL变种
	if strings.Contains(strings.ToLower(version), "mariadb") {
		info.Product = "MariaDB"
		info.Confidence = 98
		
		// MariaDB版本解析
		if mariaRe := regexp.MustCompile(`(\d+\.\d+\.\d+)-MariaDB`); mariaRe.MatchString(version) {
			if match := mariaRe.FindStringSubmatch(version); len(match) > 1 {
				info.Version = match[1]
			}
		}
		
	} else if strings.Contains(strings.ToLower(version), "percona") {
		info.Product = "Percona Server"
		info.Confidence = 98
		
	} else {
		info.Product = "MySQL"
	}
	
	// 检测操作系统
	if strings.Contains(version, "ubuntu") {
		info.OS = "Ubuntu"
		// 提取Ubuntu版本: 8.0.27-0ubuntu0.20.04.1
		if ubuntuRe := regexp.MustCompile(`ubuntu0\.(\d+\.\d+)`); ubuntuRe.MatchString(version) {
			if match := ubuntuRe.FindStringSubmatch(version); len(match) > 1 {
				info.Fields["ubuntu_version"] = match[1]
			}
		}
	} else if strings.Contains(version, "debian") {
		info.OS = "Debian"
	} else if strings.Contains(version, "el7") || strings.Contains(version, "el8") {
		info.OS = "CentOS/RHEL"
		if strings.Contains(version, "el7") {
			info.Fields["rhel_version"] = "7"
		} else if strings.Contains(version, "el8") {
			info.Fields["rhel_version"] = "8"
		}
	}
	
	// 检测特殊构建
	if strings.Contains(version, "-log") {
		info.Fields["logging_enabled"] = "true"
	}
	
	if strings.Contains(version, "cll-lve") {
		info.ExtraInfo = "CloudLinux LVE"
	}
	
	// 检测云服务提供商特征
	if strings.Contains(version, "rds") {
		info.Fields["cloud_provider"] = "AWS RDS"
	}
}

// parseHandshakePacket 解析MySQL握手包的其他字段
func (p *MySQLParser) parseHandshakePacket(data []byte, info *ParsedInfo) {
	if len(data) < 4 {
		return
	}
	
	// 解析线程ID (4字节)
	threadID := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	info.Fields["thread_id"] = strconv.FormatUint(uint64(threadID), 10)
	
	// 跳过salt的第一部分(8字节)和填充符(1字节)
	if len(data) >= 13 {
		// 解析服务器能力标志(2字节)
		capabilities := uint16(data[13]) | uint16(data[14])<<8
		info.Fields["capabilities"] = fmt.Sprintf("0x%04x", capabilities)
		
		// 解析能力标志
		capabilityFlags := []string{}
		if capabilities&0x0001 != 0 {
			capabilityFlags = append(capabilityFlags, "LONG_PASSWORD")
		}
		if capabilities&0x0002 != 0 {
			capabilityFlags = append(capabilityFlags, "FOUND_ROWS")
		}
		if capabilities&0x0004 != 0 {
			capabilityFlags = append(capabilityFlags, "LONG_FLAG")
		}
		if capabilities&0x0008 != 0 {
			capabilityFlags = append(capabilityFlags, "CONNECT_WITH_DB")
		}
		if capabilities&0x0800 != 0 {
			capabilityFlags = append(capabilityFlags, "PROTOCOL_41")
		}
		if capabilities&0x8000 != 0 {
			capabilityFlags = append(capabilityFlags, "SSL")
		}
		
		if len(capabilityFlags) > 0 {
			info.Fields["capability_flags"] = strings.Join(capabilityFlags, ", ")
		}
		
		// SSL支持检测
		if capabilities&0x8000 != 0 {
			info.Fields["ssl_support"] = "true"
		}
	}
}

func (p *MySQLParser) GetProtocol() string { return "mysql" }
func (p *MySQLParser) GetConfidence(data []byte) int {
	if len(data) > 4 && data[4] == 10 { // MySQL protocol version 10
		return 90
	}
	return 0
}

// RedisParser Redis协议解析器
type RedisParser struct{}

func (p *RedisParser) Parse(data []byte) (*ParsedInfo, error) {
	content := string(data)
	info := &ParsedInfo{
		Protocol:   "redis",
		Service:    "redis",
		Product:    "Redis",
		Fields:     make(map[string]string),
		Confidence: 95,
	}
	
	// Redis RESP协议
	if strings.HasPrefix(content, "+PONG") {
		info.Fields["response"] = "PONG"
	} else if strings.HasPrefix(content, "-ERR") {
		info.Fields["error"] = strings.TrimPrefix(content, "-ERR ")
	}
	
	return info, nil
}

func (p *RedisParser) GetProtocol() string { return "redis" }
func (p *RedisParser) GetConfidence(data []byte) int {
	content := string(data)
	if strings.HasPrefix(content, "+PONG") || strings.HasPrefix(content, "-ERR") {
		return 95
	}
	return 0
}

// PostgreSQLParser PostgreSQL协议解析器
type PostgreSQLParser struct{}

func (p *PostgreSQLParser) Parse(data []byte) (*ParsedInfo, error) {
	info := &ParsedInfo{
		Protocol:   "postgresql",
		Service:    "postgresql",
		Product:    "PostgreSQL",
		Fields:     make(map[string]string),
		Confidence: 80,
	}
	
	// PostgreSQL错误响应通常以'E'开头
	if len(data) > 0 && data[0] == 'E' {
		info.Fields["response_type"] = "error"
	}
	
	return info, nil
}

func (p *PostgreSQLParser) GetProtocol() string { return "postgresql" }
func (p *PostgreSQLParser) GetConfidence(data []byte) int {
	if len(data) > 0 && (data[0] == 'E' || data[0] == 'R') {
		return 75
	}
	return 0
}

// DNSParser DNS协议解析器
type DNSParser struct{}

func (p *DNSParser) Parse(data []byte) (*ParsedInfo, error) {
	info := &ParsedInfo{
		Protocol:   "dns",
		Service:    "dns",
		Fields:     make(map[string]string),
		Confidence: 80,
	}
	
	if len(data) >= 12 {
		// DNS头部解析
		flags := (uint16(data[2]) << 8) | uint16(data[3])
		info.Fields["flags"] = fmt.Sprintf("0x%04x", flags)
		
		if flags&0x8000 != 0 {
			info.Fields["type"] = "response"
		} else {
			info.Fields["type"] = "query"
		}
	}
	
	return info, nil
}

func (p *DNSParser) GetProtocol() string { return "dns" }
func (p *DNSParser) GetConfidence(data []byte) int {
	if len(data) >= 12 {
		return 80
	}
	return 0
}

// SNMPParser SNMP协议解析器
type SNMPParser struct{}

func (p *SNMPParser) Parse(data []byte) (*ParsedInfo, error) {
	info := &ParsedInfo{
		Protocol:   "snmp",
		Service:    "snmp",
		Fields:     make(map[string]string),
		Confidence: 75,
	}
	
	// SNMP ASN.1 BER编码
	if len(data) > 0 && data[0] == 0x30 {
		info.Fields["asn1_type"] = "sequence"
	}
	
	return info, nil
}

func (p *SNMPParser) GetProtocol() string { return "snmp" }
func (p *SNMPParser) GetConfidence(data []byte) int {
	if len(data) > 0 && data[0] == 0x30 {
		return 70
	}
	return 0
}

// TelnetParser Telnet协议解析器
type TelnetParser struct{}

func (p *TelnetParser) Parse(data []byte) (*ParsedInfo, error) {
	info := &ParsedInfo{
		Protocol:   "telnet",
		Service:    "telnet",
		Fields:     make(map[string]string),
		Confidence: 70,
	}
	
	// Telnet选项协商
	if len(data) > 0 && data[0] == 0xFF {
		info.Fields["telnet_command"] = "IAC"
		info.Confidence = 85
	}
	
	return info, nil
}

func (p *TelnetParser) GetProtocol() string { return "telnet" }
func (p *TelnetParser) GetConfidence(data []byte) int {
	if len(data) > 0 && data[0] == 0xFF {
		return 80
	}
	return 0
}

// POP3Parser POP3协议解析器
type POP3Parser struct{}

func (p *POP3Parser) Parse(data []byte) (*ParsedInfo, error) {
	content := string(data)
	info := &ParsedInfo{
		Protocol:   "pop3",
		Service:    "pop3",
		Fields:     make(map[string]string),
		Confidence: 80,
	}
	
	if strings.HasPrefix(content, "+OK") {
		info.Fields["response"] = "OK"
		info.Confidence = 90
	} else if strings.HasPrefix(content, "-ERR") {
		info.Fields["response"] = "ERR"
	}
	
	return info, nil
}

func (p *POP3Parser) GetProtocol() string { return "pop3" }
func (p *POP3Parser) GetConfidence(data []byte) int {
	content := string(data)
	if strings.HasPrefix(content, "+OK") || strings.HasPrefix(content, "-ERR") {
		return 85
	}
	return 0
}

// IMAPParser IMAP协议解析器
type IMAPParser struct{}

func (p *IMAPParser) Parse(data []byte) (*ParsedInfo, error) {
	content := string(data)
	info := &ParsedInfo{
		Protocol:   "imap",
		Service:    "imap",
		Fields:     make(map[string]string),
		Confidence: 80,
	}
	
	// IMAP响应格式: * OK [CAPABILITY ...] 或 A001 OK ...
	imapRe := regexp.MustCompile(`^(\*|A\d+)\s+(OK|NO|BAD)\s+(.*)`)
	if match := imapRe.FindStringSubmatch(content); len(match) > 3 {
		info.Fields["tag"] = match[1]
		info.Fields["status"] = match[2]
		info.Fields["message"] = match[3]
		info.Confidence = 90
	}
	
	return info, nil
}

func (p *IMAPParser) GetProtocol() string { return "imap" }
func (p *IMAPParser) GetConfidence(data []byte) int {
	content := string(data)
	if regexp.MustCompile(`^(\*|A\d+)\s+(OK|NO|BAD)`).MatchString(content) {
		return 85
	}
	return 0
}
// TLSParser TLS/SSL协议解析器
type TLSParser struct{}

func (p *TLSParser) Parse(data []byte) (*ParsedInfo, error) {
	info := &ParsedInfo{
		Protocol:   "tls",
		Service:    "tls",
		Fields:     make(map[string]string),
		Confidence: 80,
	}
	
	if len(data) < 5 {
		return info, nil
	}
	
	// TLS Record Header解析
	// [Content Type:1][Version:2][Length:2][Data...]
	contentType := data[0]
	version := (uint16(data[1]) << 8) | uint16(data[2])
	length := (uint16(data[3]) << 8) | uint16(data[4])
	
	info.Fields["content_type"] = fmt.Sprintf("%d", contentType)
	info.Fields["record_length"] = fmt.Sprintf("%d", length)
	
	// 解析TLS版本
	switch version {
	case 0x0300:
		info.Version = "SSL 3.0"
		info.Product = "SSL"
	case 0x0301:
		info.Version = "TLS 1.0"
		info.Product = "TLS"
	case 0x0302:
		info.Version = "TLS 1.1"
		info.Product = "TLS"
	case 0x0303:
		info.Version = "TLS 1.2"
		info.Product = "TLS"
	case 0x0304:
		info.Version = "TLS 1.3"
		info.Product = "TLS"
	default:
		info.Version = fmt.Sprintf("Unknown (0x%04x)", version)
		info.Product = "TLS/SSL"
	}
	
	info.Fields["tls_version"] = info.Version
	
	// 解析Content Type
	switch contentType {
	case 20:
		info.Fields["message_type"] = "ChangeCipherSpec"
	case 21:
		info.Fields["message_type"] = "Alert"
		if len(data) >= 7 {
			alertLevel := data[5]
			alertDesc := data[6]
			info.Fields["alert_level"] = fmt.Sprintf("%d", alertLevel)
			info.Fields["alert_description"] = fmt.Sprintf("%d", alertDesc)
			
			// 解析常见的Alert描述
			p.parseAlertDescription(alertDesc, info)
		}
	case 22:
		info.Fields["message_type"] = "Handshake"
		if len(data) >= 6 {
			handshakeType := data[5]
			info.Fields["handshake_type"] = fmt.Sprintf("%d", handshakeType)
			
			// 解析握手消息类型
			p.parseHandshakeType(handshakeType, data[5:], info)
		}
	case 23:
		info.Fields["message_type"] = "ApplicationData"
	default:
		info.Fields["message_type"] = fmt.Sprintf("Unknown (%d)", contentType)
	}
	
	info.Confidence = 95
	return info, nil
}

// parseAlertDescription 解析TLS Alert描述
func (p *TLSParser) parseAlertDescription(desc byte, info *ParsedInfo) {
	alertDescriptions := map[byte]string{
		0:   "close_notify",
		10:  "unexpected_message",
		20:  "bad_record_mac",
		21:  "decryption_failed",
		22:  "record_overflow",
		30:  "decompression_failure",
		40:  "handshake_failure",
		41:  "no_certificate",
		42:  "bad_certificate",
		43:  "unsupported_certificate",
		44:  "certificate_revoked",
		45:  "certificate_expired",
		46:  "certificate_unknown",
		47:  "illegal_parameter",
		48:  "unknown_ca",
		49:  "access_denied",
		50:  "decode_error",
		51:  "decrypt_error",
		60:  "export_restriction",
		70:  "protocol_version",
		71:  "insufficient_security",
		80:  "internal_error",
		90:  "user_canceled",
		100: "no_renegotiation",
		110: "unsupported_extension",
	}
	
	if desc_str, exists := alertDescriptions[desc]; exists {
		info.Fields["alert_description_name"] = desc_str
	}
}

// parseHandshakeType 解析TLS握手消息类型
func (p *TLSParser) parseHandshakeType(hsType byte, data []byte, info *ParsedInfo) {
	handshakeTypes := map[byte]string{
		0:  "HelloRequest",
		1:  "ClientHello",
		2:  "ServerHello",
		11: "Certificate",
		12: "ServerKeyExchange",
		13: "CertificateRequest",
		14: "ServerHelloDone",
		15: "CertificateVerify",
		16: "ClientKeyExchange",
		20: "Finished",
	}
	
	if hsType_str, exists := handshakeTypes[hsType]; exists {
		info.Fields["handshake_type_name"] = hsType_str
	}
	
	// 如果是ServerHello，尝试解析更多信息
	if hsType == 2 && len(data) >= 38 {
		p.parseServerHello(data, info)
	}
}

// parseServerHello 解析ServerHello消息
func (p *TLSParser) parseServerHello(data []byte, info *ParsedInfo) {
	if len(data) < 38 {
		return
	}
	
	// ServerHello结构:
	// [HandshakeType:1][Length:3][Version:2][Random:32][SessionIDLength:1][SessionID:var][CipherSuite:2][CompressionMethod:1][Extensions...]
	
	// 跳过握手头部 (4字节)
	offset := 4
	
	// 解析协议版本
	if offset+2 <= len(data) {
		version := (uint16(data[offset]) << 8) | uint16(data[offset+1])
		switch version {
		case 0x0303:
			info.Fields["negotiated_version"] = "TLS 1.2"
		case 0x0304:
			info.Fields["negotiated_version"] = "TLS 1.3"
		default:
			info.Fields["negotiated_version"] = fmt.Sprintf("0x%04x", version)
		}
		offset += 2
	}
	
	// 跳过Random (32字节)
	offset += 32
	
	// 解析Session ID
	if offset < len(data) {
		sessionIDLen := data[offset]
		info.Fields["session_id_length"] = fmt.Sprintf("%d", sessionIDLen)
		offset += 1 + int(sessionIDLen)
	}
	
	// 解析Cipher Suite
	if offset+2 <= len(data) {
		cipherSuite := (uint16(data[offset]) << 8) | uint16(data[offset+1])
		info.Fields["cipher_suite"] = fmt.Sprintf("0x%04x", cipherSuite)
		
		// 解析常见的Cipher Suite
		p.parseCipherSuite(cipherSuite, info)
		offset += 2
	}
	
	// 解析压缩方法
	if offset < len(data) {
		compression := data[offset]
		info.Fields["compression_method"] = fmt.Sprintf("%d", compression)
	}
}

// parseCipherSuite 解析Cipher Suite
func (p *TLSParser) parseCipherSuite(suite uint16, info *ParsedInfo) {
	cipherSuites := map[uint16]string{
		0x002f: "TLS_RSA_WITH_AES_128_CBC_SHA",
		0x0035: "TLS_RSA_WITH_AES_256_CBC_SHA",
		0x003c: "TLS_RSA_WITH_AES_128_CBC_SHA256",
		0x003d: "TLS_RSA_WITH_AES_256_CBC_SHA256",
		0x009c: "TLS_RSA_WITH_AES_128_GCM_SHA256",
		0x009d: "TLS_RSA_WITH_AES_256_GCM_SHA384",
		0xc007: "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA",
		0xc009: "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
		0xc00a: "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
		0xc011: "TLS_ECDHE_RSA_WITH_RC4_128_SHA",
		0xc013: "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
		0xc014: "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
		0xc023: "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256",
		0xc024: "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384",
		0xc027: "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256",
		0xc028: "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384",
		0xc02b: "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		0xc02c: "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		0xc02f: "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		0xc030: "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		0xcca8: "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
		0xcca9: "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
	}
	
	if suite_str, exists := cipherSuites[suite]; exists {
		info.Fields["cipher_suite_name"] = suite_str
		
		// 分析加密强度
		if strings.Contains(suite_str, "AES_256") || strings.Contains(suite_str, "CHACHA20") {
			info.Fields["encryption_strength"] = "Strong"
		} else if strings.Contains(suite_str, "AES_128") {
			info.Fields["encryption_strength"] = "Medium"
		} else if strings.Contains(suite_str, "RC4") {
			info.Fields["encryption_strength"] = "Weak"
		}
		
		// 检测前向保密
		if strings.Contains(suite_str, "ECDHE") || strings.Contains(suite_str, "DHE") {
			info.Fields["forward_secrecy"] = "Yes"
		} else {
			info.Fields["forward_secrecy"] = "No"
		}
	}
}

func (p *TLSParser) GetProtocol() string { return "tls" }
func (p *TLSParser) GetConfidence(data []byte) int {
	if len(data) >= 3 {
		// 检查TLS Record Header
		contentType := data[0]
		version := (uint16(data[1]) << 8) | uint16(data[2])
		
		// 检查Content Type是否有效
		if contentType >= 20 && contentType <= 23 {
			// 检查版本是否有效
			if version >= 0x0300 && version <= 0x0304 {
				return 95
			}
		}
	}
	return 0
}

// HTTPSParser HTTPS协议解析器 (HTTP over TLS)
type HTTPSParser struct {
	httpParser *HTTPParser
	tlsParser  *TLSParser
}

func NewHTTPSParser() *HTTPSParser {
	return &HTTPSParser{
		httpParser: &HTTPParser{},
		tlsParser:  &TLSParser{},
	}
}

func (p *HTTPSParser) Parse(data []byte) (*ParsedInfo, error) {
	// 首先尝试解析为TLS
	if len(data) >= 5 && data[0] >= 20 && data[0] <= 23 {
		return p.tlsParser.Parse(data)
	}
	
	// 如果不是TLS记录，尝试解析为HTTP (可能是明文HTTP响应)
	return p.httpParser.Parse(data)
}

func (p *HTTPSParser) GetProtocol() string { return "https" }
func (p *HTTPSParser) GetConfidence(data []byte) int {
	// 优先检查TLS
	if tlsConf := p.tlsParser.GetConfidence(data); tlsConf > 0 {
		return tlsConf
	}
	
	// 然后检查HTTP
	return p.httpParser.GetConfidence(data)
}
// MQTTParser MQTT协议解析器
type MQTTParser struct{}

func (p *MQTTParser) Parse(data []byte) (*ParsedInfo, error) {
	info := &ParsedInfo{
		Protocol:   "mqtt",
		Service:    "mqtt",
		Product:    "MQTT Broker",
		Fields:     make(map[string]string),
		Confidence: 80,
	}
	
	if len(data) < 2 {
		return info, nil
	}
	
	// MQTT Fixed Header解析
	// [Message Type + Flags:1][Remaining Length:1-4]
	messageType := (data[0] >> 4) & 0x0F
	dup := (data[0] >> 3) & 0x01
	qos := (data[0] >> 1) & 0x03
	retain := data[0] & 0x01
	
	info.Fields["message_type"] = fmt.Sprintf("%d", messageType)
	info.Fields["dup"] = fmt.Sprintf("%d", dup)
	info.Fields["qos"] = fmt.Sprintf("%d", qos)
	info.Fields["retain"] = fmt.Sprintf("%d", retain)
	
	// 解析消息类型
	messageTypeName := p.getMQTTMessageTypeName(messageType)
	info.Fields["message_type_name"] = messageTypeName
	
	// 解析剩余长度
	remainingLength, lengthBytes := p.decodeMQTTLength(data[1:])
	if remainingLength >= 0 {
		info.Fields["remaining_length"] = fmt.Sprintf("%d", remainingLength)
		info.Fields["length_bytes"] = fmt.Sprintf("%d", lengthBytes)
	}
	
	// 根据消息类型进行详细解析
	switch messageType {
	case 1: // CONNECT
		p.parseMQTTConnect(data, info)
	case 2: // CONNACK
		p.parseMQTTConnack(data, info)
		info.Confidence = 95 // CONNACK响应说明这确实是MQTT服务器
	case 3: // PUBLISH
		p.parseMQTTPublish(data, info)
	case 4: // PUBACK
		info.Fields["message_description"] = "Publish Acknowledgment"
	case 8: // SUBSCRIBE
		info.Fields["message_description"] = "Subscribe Request"
	case 9: // SUBACK
		info.Fields["message_description"] = "Subscribe Acknowledgment"
	case 12: // PINGREQ
		info.Fields["message_description"] = "Ping Request"
	case 13: // PINGRESP
		info.Fields["message_description"] = "Ping Response"
		info.Confidence = 90
	case 14: // DISCONNECT
		info.Fields["message_description"] = "Disconnect"
	}
	
	return info, nil
}

// getMQTTMessageTypeName 获取MQTT消息类型名称
func (p *MQTTParser) getMQTTMessageTypeName(messageType byte) string {
	messageTypes := map[byte]string{
		0:  "Reserved",
		1:  "CONNECT",
		2:  "CONNACK",
		3:  "PUBLISH",
		4:  "PUBACK",
		5:  "PUBREC",
		6:  "PUBREL",
		7:  "PUBCOMP",
		8:  "SUBSCRIBE",
		9:  "SUBACK",
		10: "UNSUBSCRIBE",
		11: "UNSUBACK",
		12: "PINGREQ",
		13: "PINGRESP",
		14: "DISCONNECT",
		15: "Reserved",
	}
	
	if name, exists := messageTypes[messageType]; exists {
		return name
	}
	return fmt.Sprintf("Unknown (%d)", messageType)
}

// decodeMQTTLength 解码MQTT变长字段
func (p *MQTTParser) decodeMQTTLength(data []byte) (int, int) {
	multiplier := 1
	length := 0
	index := 0
	
	for index < len(data) {
		if index >= 4 { // MQTT规范：最多4字节
			return -1, -1
		}
		
		encodedByte := data[index]
		length += int(encodedByte&127) * multiplier
		
		if (encodedByte & 128) == 0 {
			break
		}
		
		multiplier *= 128
		index++
	}
	
	return length, index + 1
}

// parseMQTTConnect 解析MQTT CONNECT消息
func (p *MQTTParser) parseMQTTConnect(data []byte, info *ParsedInfo) {
	if len(data) < 10 {
		return
	}
	
	// 跳过Fixed Header
	_, lengthBytes := p.decodeMQTTLength(data[1:])
	offset := 1 + lengthBytes
	
	if offset+6 > len(data) {
		return
	}
	
	// 解析协议名长度
	protocolNameLen := int(data[offset])<<8 | int(data[offset+1])
	offset += 2
	
	if offset+protocolNameLen > len(data) {
		return
	}
	
	// 解析协议名
	protocolName := string(data[offset : offset+protocolNameLen])
	info.Fields["protocol_name"] = protocolName
	offset += protocolNameLen
	
	if offset+4 > len(data) {
		return
	}
	
	// 解析协议级别
	protocolLevel := data[offset]
	info.Fields["protocol_level"] = fmt.Sprintf("%d", protocolLevel)
	
	// 根据协议级别确定MQTT版本
	switch protocolLevel {
	case 3:
		info.Version = "3.1"
	case 4:
		info.Version = "3.1.1"
	case 5:
		info.Version = "5.0"
	default:
		info.Version = fmt.Sprintf("Unknown (%d)", protocolLevel)
	}
	
	offset++
	
	// 解析连接标志
	connectFlags := data[offset]
	info.Fields["clean_session"] = fmt.Sprintf("%d", (connectFlags>>1)&0x01)
	info.Fields["will_flag"] = fmt.Sprintf("%d", (connectFlags>>2)&0x01)
	info.Fields["will_qos"] = fmt.Sprintf("%d", (connectFlags>>3)&0x03)
	info.Fields["will_retain"] = fmt.Sprintf("%d", (connectFlags>>5)&0x01)
	info.Fields["password_flag"] = fmt.Sprintf("%d", (connectFlags>>6)&0x01)
	info.Fields["username_flag"] = fmt.Sprintf("%d", (connectFlags>>7)&0x01)
	offset++
	
	// 解析Keep Alive
	keepAlive := int(data[offset])<<8 | int(data[offset+1])
	info.Fields["keep_alive"] = fmt.Sprintf("%d", keepAlive)
}

// parseMQTTConnack 解析MQTT CONNACK消息
func (p *MQTTParser) parseMQTTConnack(data []byte, info *ParsedInfo) {
	// 跳过Fixed Header
	_, lengthBytes := p.decodeMQTTLength(data[1:])
	offset := 1 + lengthBytes
	
	if offset+2 > len(data) {
		return
	}
	
	// 解析连接确认标志
	connectAckFlags := data[offset]
	sessionPresent := connectAckFlags & 0x01
	info.Fields["session_present"] = fmt.Sprintf("%d", sessionPresent)
	offset++
	
	// 解析返回码
	returnCode := data[offset]
	info.Fields["return_code"] = fmt.Sprintf("%d", returnCode)
	
	// 解析返回码含义
	returnCodeName := p.getMQTTReturnCodeName(returnCode)
	info.Fields["return_code_name"] = returnCodeName
	
	// 根据返回码判断连接状态
	if returnCode == 0 {
		info.ExtraInfo = "Connection Accepted"
	} else {
		info.ExtraInfo = fmt.Sprintf("Connection Refused: %s", returnCodeName)
	}
}

// getMQTTReturnCodeName 获取MQTT返回码名称
func (p *MQTTParser) getMQTTReturnCodeName(code byte) string {
	returnCodes := map[byte]string{
		0: "Connection Accepted",
		1: "Connection Refused: Unacceptable Protocol Version",
		2: "Connection Refused: Identifier Rejected",
		3: "Connection Refused: Server Unavailable",
		4: "Connection Refused: Bad User Name or Password",
		5: "Connection Refused: Not Authorized",
	}
	
	if name, exists := returnCodes[code]; exists {
		return name
	}
	return fmt.Sprintf("Unknown (%d)", code)
}

// parseMQTTPublish 解析MQTT PUBLISH消息
func (p *MQTTParser) parseMQTTPublish(data []byte, info *ParsedInfo) {
	// 跳过Fixed Header
	_, lengthBytes := p.decodeMQTTLength(data[1:])
	offset := 1 + lengthBytes
	
	if offset+2 > len(data) {
		return
	}
	
	// 解析主题长度
	topicLen := int(data[offset])<<8 | int(data[offset+1])
	offset += 2
	
	if offset+topicLen > len(data) {
		return
	}
	
	// 解析主题
	topic := string(data[offset : offset+topicLen])
	info.Fields["topic"] = topic
	offset += topicLen
	
	// 如果QoS > 0，还有Packet Identifier
	qos := (data[0] >> 1) & 0x03
	if qos > 0 && offset+2 <= len(data) {
		packetID := int(data[offset])<<8 | int(data[offset+1])
		info.Fields["packet_id"] = fmt.Sprintf("%d", packetID)
		offset += 2
	}
	
	// 剩余的是消息内容
	if offset < len(data) {
		payloadLen := len(data) - offset
		info.Fields["payload_length"] = fmt.Sprintf("%d", payloadLen)
		
		// 如果payload不太大，显示内容
		if payloadLen <= 100 {
			payload := string(data[offset:])
			info.Fields["payload"] = payload
		}
	}
}

func (p *MQTTParser) GetProtocol() string { return "mqtt" }
func (p *MQTTParser) GetConfidence(data []byte) int {
	if len(data) < 2 {
		return 0
	}
	
	// 检查MQTT Fixed Header
	messageType := (data[0] >> 4) & 0x0F
	
	// 检查消息类型是否有效
	if messageType >= 1 && messageType <= 14 && messageType != 0 && messageType != 15 {
		// 检查剩余长度编码是否有效
		if len(data) > 1 {
			_, lengthBytes := p.decodeMQTTLength(data[1:])
			if lengthBytes > 0 && lengthBytes <= 4 {
				// CONNACK响应置信度最高
				if messageType == 2 {
					return 95
				}
				// 其他有效消息类型
				return 80
			}
		}
	}
	
	return 0
}

// MQTTWebSocketParser MQTT over WebSocket解析器
type MQTTWebSocketParser struct {
	mqttParser *MQTTParser
	httpParser *HTTPParser
}

func NewMQTTWebSocketParser() *MQTTWebSocketParser {
	return &MQTTWebSocketParser{
		mqttParser: &MQTTParser{},
		httpParser: &HTTPParser{},
	}
}

func (p *MQTTWebSocketParser) Parse(data []byte) (*ParsedInfo, error) {
	// 首先尝试解析为HTTP WebSocket升级响应
	if strings.Contains(string(data), "HTTP/") && strings.Contains(strings.ToLower(string(data)), "websocket") {
		info, err := p.httpParser.Parse(data)
		if err == nil {
			info.Protocol = "mqtt-ws"
			info.Service = "mqtt-websocket"
			
			// 检查是否包含MQTT相关头部
			if strings.Contains(strings.ToLower(string(data)), "mqtt") {
				info.Product = "MQTT over WebSocket"
				info.ExtraInfo = "WebSocket MQTT Broker"
				info.Confidence = 90
			}
		}
		return info, err
	}
	
	// 如果不是HTTP响应，尝试解析为MQTT
	return p.mqttParser.Parse(data)
}

func (p *MQTTWebSocketParser) GetProtocol() string { return "mqtt-ws" }
func (p *MQTTWebSocketParser) GetConfidence(data []byte) int {
	content := strings.ToLower(string(data))
	
	// WebSocket升级响应
	if strings.Contains(content, "http/") && strings.Contains(content, "websocket") {
		if strings.Contains(content, "mqtt") {
			return 90
		}
		return 70
	}
	
	// 直接MQTT数据
	return p.mqttParser.GetConfidence(data)
}
// RTSPParser RTSP协议解析器
type RTSPParser struct{}

func (p *RTSPParser) Parse(data []byte) (*ParsedInfo, error) {
	content := string(data)
	info := &ParsedInfo{
		Protocol:   "rtsp",
		Service:    "rtsp",
		Product:    "RTSP Server",
		Fields:     make(map[string]string),
		Confidence: 80,
	}
	
	lines := strings.Split(content, "\r\n")
	if len(lines) == 0 {
		lines = strings.Split(content, "\n")
	}
	
	if len(lines) > 0 {
		// 解析RTSP状态行: RTSP/1.0 200 OK
		statusLine := strings.TrimSpace(lines[0])
		info.Fields["status_line"] = statusLine
		
		rtspRe := regexp.MustCompile(`RTSP/(\d+\.\d+)\s+(\d+)\s*(.*)`)
		if match := rtspRe.FindStringSubmatch(statusLine); len(match) > 3 {
			info.Version = match[1]
			info.Fields["status_code"] = match[2]
			info.Fields["status_text"] = strings.TrimSpace(match[3])
			info.Confidence = 95
		}
		
		// 解析RTSP头部
		for i := 1; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				break
			}
			
			if colonIdx := strings.Index(line, ":"); colonIdx > 0 {
				key := strings.ToLower(strings.TrimSpace(line[:colonIdx]))
				value := strings.TrimSpace(line[colonIdx+1:])
				info.Fields["header_"+key] = value
				
				// 解析特定头部
				switch key {
				case "server":
					p.parseRTSPServer(value, info)
				case "public":
					info.Fields["supported_methods"] = value
				case "cseq":
					info.Fields["sequence"] = value
				case "session":
					info.Fields["session_id"] = value
				}
			}
		}
	}
	
	return info, nil
}

// parseRTSPServer 解析RTSP服务器信息
func (p *RTSPParser) parseRTSPServer(server string, info *ParsedInfo) {
	info.Fields["server"] = server
	
	// 检测常见的RTSP服务器
	serverLower := strings.ToLower(server)
	
	if strings.Contains(serverLower, "hikvision") {
		info.Product = "Hikvision IP Camera"
		info.ExtraInfo = "Hikvision RTSP Server"
		info.Confidence = 98
		
		// 提取版本信息
		if versionRe := regexp.MustCompile(`hikvision.*?(\d+\.\d+\.\d+)`); versionRe.MatchString(serverLower) {
			if match := versionRe.FindStringSubmatch(serverLower); len(match) > 1 {
				info.Version = match[1]
			}
		}
	} else if strings.Contains(serverLower, "dahua") {
		info.Product = "Dahua IP Camera"
		info.ExtraInfo = "Dahua RTSP Server"
		info.Confidence = 98
	} else if strings.Contains(serverLower, "axis") {
		info.Product = "AXIS IP Camera"
		info.ExtraInfo = "AXIS RTSP Server"
		info.Confidence = 98
	} else if strings.Contains(serverLower, "uniview") || strings.Contains(serverLower, "unv") {
		info.Product = "Uniview IP Camera"
		info.ExtraInfo = "Uniview RTSP Server"
		info.Confidence = 98
	} else if strings.Contains(serverLower, "gstreamer") {
		info.Product = "GStreamer RTSP Server"
		info.ExtraInfo = "Open Source RTSP Server"
	} else if strings.Contains(serverLower, "live555") {
		info.Product = "Live555 RTSP Server"
		info.ExtraInfo = "Live555 Media Server"
	}
}

func (p *RTSPParser) GetProtocol() string { return "rtsp" }
func (p *RTSPParser) GetConfidence(data []byte) int {
	content := string(data)
	if strings.HasPrefix(content, "RTSP/") {
		return 95
	}
	if strings.Contains(content, "RTSP/1.0") {
		return 90
	}
	return 0
}

// ONVIFParser ONVIF协议解析器
type ONVIFParser struct{}

func (p *ONVIFParser) Parse(data []byte) (*ParsedInfo, error) {
	content := string(data)
	info := &ParsedInfo{
		Protocol:   "onvif",
		Service:    "onvif",
		Product:    "ONVIF Device",
		Fields:     make(map[string]string),
		Confidence: 80,
	}
	
	// 检查是否为SOAP响应
	if strings.Contains(content, "soap:Envelope") || strings.Contains(content, "SOAP-ENV:Envelope") {
		info.Fields["message_type"] = "SOAP Response"
		info.Confidence = 90
		
		// 解析设备信息
		if strings.Contains(content, "GetDeviceInformationResponse") {
			info.Fields["response_type"] = "DeviceInformation"
			
			// 提取制造商信息
			if manufacturerRe := regexp.MustCompile(`<tds:Manufacturer>(.*?)</tds:Manufacturer>`); manufacturerRe.MatchString(content) {
				if match := manufacturerRe.FindStringSubmatch(content); len(match) > 1 {
					manufacturer := strings.TrimSpace(match[1])
					info.Fields["manufacturer"] = manufacturer
					info.Product = manufacturer + " ONVIF Device"
				}
			}
			
			// 提取型号信息
			if modelRe := regexp.MustCompile(`<tds:Model>(.*?)</tds:Model>`); modelRe.MatchString(content) {
				if match := modelRe.FindStringSubmatch(content); len(match) > 1 {
					info.Fields["model"] = strings.TrimSpace(match[1])
				}
			}
			
			// 提取固件版本
			if firmwareRe := regexp.MustCompile(`<tds:FirmwareVersion>(.*?)</tds:FirmwareVersion>`); firmwareRe.MatchString(content) {
				if match := firmwareRe.FindStringSubmatch(content); len(match) > 1 {
					info.Version = strings.TrimSpace(match[1])
				}
			}
			
			// 提取序列号
			if serialRe := regexp.MustCompile(`<tds:SerialNumber>(.*?)</tds:SerialNumber>`); serialRe.MatchString(content) {
				if match := serialRe.FindStringSubmatch(content); len(match) > 1 {
					info.Fields["serial_number"] = strings.TrimSpace(match[1])
				}
			}
			
			info.Confidence = 98
		}
		
		// 检查是否为WS-Discovery响应
		if strings.Contains(content, "ProbeMatches") {
			info.Fields["response_type"] = "WS-Discovery ProbeMatches"
			info.Service = "onvif-discovery"
			
			// 提取设备类型
			if typeRe := regexp.MustCompile(`<d:Types>(.*?)</d:Types>`); typeRe.MatchString(content) {
				if match := typeRe.FindStringSubmatch(content); len(match) > 1 {
					info.Fields["device_types"] = strings.TrimSpace(match[1])
				}
			}
			
			// 提取XAddrs (设备地址)
			if xaddrsRe := regexp.MustCompile(`<d:XAddrs>(.*?)</d:XAddrs>`); xaddrsRe.MatchString(content) {
				if match := xaddrsRe.FindStringSubmatch(content); len(match) > 1 {
					info.Fields["device_addresses"] = strings.TrimSpace(match[1])
				}
			}
		}
	}
	
	return info, nil
}

func (p *ONVIFParser) GetProtocol() string { return "onvif" }
func (p *ONVIFParser) GetConfidence(data []byte) int {
	content := strings.ToLower(string(data))
	if strings.Contains(content, "onvif") && strings.Contains(content, "soap") {
		return 95
	}
	if strings.Contains(content, "soap:envelope") && strings.Contains(content, "device") {
		return 80
	}
	return 0
}

// HikvisionParser 海康威视协议解析器
type HikvisionParser struct{}

func (p *HikvisionParser) Parse(data []byte) (*ParsedInfo, error) {
	content := string(data)
	info := &ParsedInfo{
		Protocol:   "hikvision",
		Service:    "hikvision",
		Product:    "Hikvision IP Camera",
		Fields:     make(map[string]string),
		Confidence: 85,
	}
	
	// 检查HTTP响应
	if strings.Contains(content, "HTTP/") {
		// 解析HTTP状态
		lines := strings.Split(content, "\r\n")
		if len(lines) > 0 {
			statusLine := strings.TrimSpace(lines[0])
			info.Fields["status_line"] = statusLine
			
			if strings.Contains(statusLine, "200 OK") {
				info.Confidence = 95
			} else if strings.Contains(statusLine, "401") {
				info.Fields["auth_required"] = "true"
				info.ExtraInfo = "Authentication Required"
			}
		}
		
		// 检查服务器头部
		if serverMatch := regexp.MustCompile(`(?i)server:\s*(.+)`).FindStringSubmatch(content); len(serverMatch) > 1 {
			server := strings.TrimSpace(serverMatch[1])
			info.Fields["server"] = server
			
			if strings.Contains(strings.ToLower(server), "hikvision") {
				info.Confidence = 98
			}
		}
		
		// 检查ISAPI响应
		if strings.Contains(content, "<DeviceInfo") {
			info.Fields["response_type"] = "ISAPI DeviceInfo"
			
			// 提取设备型号
			if modelRe := regexp.MustCompile(`<model>(.*?)</model>`); modelRe.MatchString(content) {
				if match := modelRe.FindStringSubmatch(content); len(match) > 1 {
					info.Fields["model"] = strings.TrimSpace(match[1])
				}
			}
			
			// 提取固件版本
			if firmwareRe := regexp.MustCompile(`<firmwareVersion>(.*?)</firmwareVersion>`); firmwareRe.MatchString(content) {
				if match := firmwareRe.FindStringSubmatch(content); len(match) > 1 {
					info.Version = strings.TrimSpace(match[1])
				}
			}
			
			// 提取序列号
			if serialRe := regexp.MustCompile(`<serialNumber>(.*?)</serialNumber>`); serialRe.MatchString(content) {
				if match := serialRe.FindStringSubmatch(content); len(match) > 1 {
					info.Fields["serial_number"] = strings.TrimSpace(match[1])
				}
			}
			
			info.Confidence = 98
		}
	}
	
	return info, nil
}

func (p *HikvisionParser) GetProtocol() string { return "hikvision" }
func (p *HikvisionParser) GetConfidence(data []byte) int {
	content := strings.ToLower(string(data))
	if strings.Contains(content, "hikvision") {
		return 95
	}
	if strings.Contains(content, "isapi") {
		return 85
	}
	return 0
}

// DahuaParser 大华协议解析器
type DahuaParser struct{}

func (p *DahuaParser) Parse(data []byte) (*ParsedInfo, error) {
	info := &ParsedInfo{
		Protocol:   "dahua",
		Service:    "dahua",
		Product:    "Dahua IP Camera",
		Fields:     make(map[string]string),
		Confidence: 85,
	}
	
	if len(data) < 4 {
		return info, nil
	}
	
	// 检查大华协议头部
	if data[0] == 0xa0 {
		info.Fields["protocol_header"] = "0xa0"
		info.Confidence = 90
		
		// 解析包长度
		if len(data) >= 4 {
			length := uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
			info.Fields["packet_length"] = fmt.Sprintf("%d", length)
		}
		
		// 解析命令类型
		if len(data) >= 8 {
			cmdType := uint32(data[4])<<24 | uint32(data[5])<<16 | uint32(data[6])<<8 | uint32(data[7])
			info.Fields["command_type"] = fmt.Sprintf("0x%08x", cmdType)
			
			// 解析常见命令类型
			switch cmdType {
			case 0x01:
				info.Fields["command_name"] = "Login Request"
			case 0x02:
				info.Fields["command_name"] = "Login Response"
				info.Confidence = 95
			case 0x03:
				info.Fields["command_name"] = "Logout"
			case 0x1000:
				info.Fields["command_name"] = "Keep Alive"
			}
		}
		
		// 解析会话ID
		if len(data) >= 16 {
			sessionID := uint32(data[12])<<24 | uint32(data[13])<<16 | uint32(data[14])<<8 | uint32(data[15])
			info.Fields["session_id"] = fmt.Sprintf("0x%08x", sessionID)
		}
	}
	
	return info, nil
}

func (p *DahuaParser) GetProtocol() string { return "dahua" }
func (p *DahuaParser) GetConfidence(data []byte) int {
	if len(data) >= 4 && data[0] == 0xa0 {
		return 90
	}
	return 0
}

// ModbusParser Modbus TCP协议解析器
type ModbusParser struct{}

func (p *ModbusParser) Parse(data []byte) (*ParsedInfo, error) {
	info := &ParsedInfo{
		Protocol:   "modbus",
		Service:    "modbus",
		Product:    "Modbus TCP Server",
		Fields:     make(map[string]string),
		Confidence: 80,
	}
	
	if len(data) < 8 {
		return info, nil
	}
	
	// Modbus TCP ADU解析
	// [Transaction ID:2][Protocol ID:2][Length:2][Unit ID:1][Function Code:1][Data:N]
	
	transactionID := uint16(data[0])<<8 | uint16(data[1])
	protocolID := uint16(data[2])<<8 | uint16(data[3])
	length := uint16(data[4])<<8 | uint16(data[5])
	unitID := data[6]
	functionCode := data[7]
	
	info.Fields["transaction_id"] = fmt.Sprintf("%d", transactionID)
	info.Fields["protocol_id"] = fmt.Sprintf("%d", protocolID)
	info.Fields["length"] = fmt.Sprintf("%d", length)
	info.Fields["unit_id"] = fmt.Sprintf("%d", unitID)
	info.Fields["function_code"] = fmt.Sprintf("%d", functionCode)
	
	// 检查协议ID (Modbus TCP应该是0)
	if protocolID == 0 {
		info.Confidence = 95
	}
	
	// 解析功能码
	functionName := p.getModbusFunctionName(functionCode)
	info.Fields["function_name"] = functionName
	
	// 检查是否为异常响应
	if functionCode >= 0x80 {
		info.Fields["exception_response"] = "true"
		exceptionCode := ""
		if len(data) > 8 {
			exceptionCode = fmt.Sprintf("%d", data[8])
		}
		info.Fields["exception_code"] = exceptionCode
		info.ExtraInfo = "Modbus Exception Response"
	} else {
		info.ExtraInfo = fmt.Sprintf("Modbus Function: %s", functionName)
	}
	
	return info, nil
}

// getModbusFunctionName 获取Modbus功能码名称
func (p *ModbusParser) getModbusFunctionName(code byte) string {
	functions := map[byte]string{
		0x01: "Read Coils",
		0x02: "Read Discrete Inputs",
		0x03: "Read Holding Registers",
		0x04: "Read Input Registers",
		0x05: "Write Single Coil",
		0x06: "Write Single Register",
		0x0F: "Write Multiple Coils",
		0x10: "Write Multiple Registers",
		0x16: "Mask Write Register",
		0x17: "Read/Write Multiple Registers",
	}
	
	if code >= 0x80 {
		baseCode := code - 0x80
		if name, exists := functions[baseCode]; exists {
			return name + " (Exception)"
		}
		return "Exception Response"
	}
	
	if name, exists := functions[code]; exists {
		return name
	}
	return fmt.Sprintf("Unknown (0x%02x)", code)
}

func (p *ModbusParser) GetProtocol() string { return "modbus" }
func (p *ModbusParser) GetConfidence(data []byte) int {
	if len(data) >= 8 {
		// 检查协议ID
		protocolID := uint16(data[2])<<8 | uint16(data[3])
		if protocolID == 0 {
			// 检查功能码是否有效
			functionCode := data[7]
			if functionCode <= 0x18 || (functionCode >= 0x80 && functionCode <= 0x98) {
				return 90
			}
		}
	}
	return 0
}

// DNP3Parser DNP3协议解析器
type DNP3Parser struct{}

func (p *DNP3Parser) Parse(data []byte) (*ParsedInfo, error) {
	info := &ParsedInfo{
		Protocol:   "dnp3",
		Service:    "dnp3",
		Product:    "DNP3 Outstation",
		Fields:     make(map[string]string),
		Confidence: 80,
	}
	
	if len(data) < 10 {
		return info, nil
	}
	
	// DNP3 Link Layer Frame解析
	// [Start:2][Length:1][Control:1][Dest:2][Src:2][CRC:2]
	
	// 检查起始字节
	if data[0] == 0x05 && data[1] == 0x64 {
		info.Confidence = 90
		info.Fields["start_bytes"] = "0x0564"
		
		length := data[2]
		control := data[3]
		dest := uint16(data[4])<<8 | uint16(data[5])
		src := uint16(data[6])<<8 | uint16(data[7])
		
		info.Fields["length"] = fmt.Sprintf("%d", length)
		info.Fields["control"] = fmt.Sprintf("0x%02x", control)
		info.Fields["destination"] = fmt.Sprintf("%d", dest)
		info.Fields["source"] = fmt.Sprintf("%d", src)
		
		// 解析控制字段
		dir := (control >> 7) & 0x01
		prm := (control >> 6) & 0x01
		fcb := (control >> 5) & 0x01
		fcv := (control >> 4) & 0x01
		function := control & 0x0F
		
		info.Fields["direction"] = fmt.Sprintf("%d", dir)
		info.Fields["primary"] = fmt.Sprintf("%d", prm)
		info.Fields["frame_count_bit"] = fmt.Sprintf("%d", fcb)
		info.Fields["frame_count_valid"] = fmt.Sprintf("%d", fcv)
		info.Fields["function_code"] = fmt.Sprintf("%d", function)
		
		// 解析功能码
		functionName := p.getDNP3FunctionName(function, prm == 1)
		info.Fields["function_name"] = functionName
		info.ExtraInfo = fmt.Sprintf("DNP3 %s", functionName)
		
		info.Confidence = 95
	}
	
	return info, nil
}

// getDNP3FunctionName 获取DNP3功能码名称
func (p *DNP3Parser) getDNP3FunctionName(code byte, isPrimary bool) string {
	if isPrimary {
		// Primary station functions
		primaryFunctions := map[byte]string{
			0:  "Reset Link",
			1:  "Reset User Process",
			2:  "Test Link",
			3:  "Confirmed User Data",
			4:  "Unconfirmed User Data",
			9:  "Request Link Status",
		}
		if name, exists := primaryFunctions[code]; exists {
			return name
		}
	} else {
		// Secondary station functions
		secondaryFunctions := map[byte]string{
			0:  "ACK",
			1:  "NACK",
			11: "Link Status",
			14: "Link Not Functioning",
			15: "Link Not Used",
		}
		if name, exists := secondaryFunctions[code]; exists {
			return name
		}
	}
	
	return fmt.Sprintf("Unknown (%d)", code)
}

func (p *DNP3Parser) GetProtocol() string { return "dnp3" }
func (p *DNP3Parser) GetConfidence(data []byte) int {
	if len(data) >= 2 && data[0] == 0x05 && data[1] == 0x64 {
		return 95
	}
	return 0
}

// BACnetParser BACnet协议解析器
type BACnetParser struct{}

func (p *BACnetParser) Parse(data []byte) (*ParsedInfo, error) {
	info := &ParsedInfo{
		Protocol:   "bacnet",
		Service:    "bacnet",
		Product:    "BACnet Device",
		Fields:     make(map[string]string),
		Confidence: 80,
	}
	
	if len(data) < 4 {
		return info, nil
	}
	
	// BACnet/IP BVLC Header解析
	// [Type:1][Function:1][Length:2]
	
	bvlcType := data[0]
	bvlcFunction := data[1]
	bvlcLength := uint16(data[2])<<8 | uint16(data[3])
	
	info.Fields["bvlc_type"] = fmt.Sprintf("0x%02x", bvlcType)
	info.Fields["bvlc_function"] = fmt.Sprintf("0x%02x", bvlcFunction)
	info.Fields["bvlc_length"] = fmt.Sprintf("%d", bvlcLength)
	
	// 检查BACnet/IP类型
	if bvlcType == 0x81 {
		info.Confidence = 90
		info.Fields["network_type"] = "BACnet/IP"
		
		// 解析BVLC功能
		functionName := p.getBACnetBVLCFunction(bvlcFunction)
		info.Fields["bvlc_function_name"] = functionName
		
		// 如果有NPDU，继续解析
		if len(data) > 4 {
			p.parseBACnetNPDU(data[4:], info)
		}
		
		info.ExtraInfo = fmt.Sprintf("BACnet/IP %s", functionName)
	}
	
	return info, nil
}

// getBACnetBVLCFunction 获取BACnet BVLC功能名称
func (p *BACnetParser) getBACnetBVLCFunction(function byte) string {
	functions := map[byte]string{
		0x00: "BVLC-Result",
		0x01: "Write-Broadcast-Distribution-Table",
		0x02: "Read-Broadcast-Distribution-Table",
		0x03: "Read-Broadcast-Distribution-Table-Ack",
		0x04: "Forwarded-NPDU",
		0x05: "Register-Foreign-Device",
		0x06: "Read-Foreign-Device-Table",
		0x07: "Read-Foreign-Device-Table-Ack",
		0x08: "Delete-Foreign-Device-Table-Entry",
		0x09: "Distribute-Broadcast-To-Network",
		0x0A: "Original-Unicast-NPDU",
		0x0B: "Original-Broadcast-NPDU",
	}
	
	if name, exists := functions[function]; exists {
		return name
	}
	return fmt.Sprintf("Unknown (0x%02x)", function)
}

// parseBACnetNPDU 解析BACnet NPDU
func (p *BACnetParser) parseBACnetNPDU(data []byte, info *ParsedInfo) {
	if len(data) < 2 {
		return
	}
	
	version := data[0]
	control := data[1]
	
	info.Fields["npdu_version"] = fmt.Sprintf("%d", version)
	info.Fields["npdu_control"] = fmt.Sprintf("0x%02x", control)
	
	// 解析控制字段
	networkLayerMessage := (control >> 7) & 0x01
	destinationSpecifier := (control >> 5) & 0x01
	sourceSpecifier := (control >> 3) & 0x01
	expectingReply := (control >> 2) & 0x01
	networkPriority := control & 0x03
	
	info.Fields["network_layer_message"] = fmt.Sprintf("%d", networkLayerMessage)
	info.Fields["destination_specifier"] = fmt.Sprintf("%d", destinationSpecifier)
	info.Fields["source_specifier"] = fmt.Sprintf("%d", sourceSpecifier)
	info.Fields["expecting_reply"] = fmt.Sprintf("%d", expectingReply)
	info.Fields["network_priority"] = fmt.Sprintf("%d", networkPriority)
}

func (p *BACnetParser) GetProtocol() string { return "bacnet" }
func (p *BACnetParser) GetConfidence(data []byte) int {
	if len(data) >= 4 && data[0] == 0x81 {
		// 检查BVLC功能码是否有效
		function := data[1]
		if function <= 0x0B {
			return 90
		}
	}
	return 0
}

// OPCUAParser OPC UA协议解析器
type OPCUAParser struct{}

func (p *OPCUAParser) Parse(data []byte) (*ParsedInfo, error) {
	info := &ParsedInfo{
		Protocol:   "opcua",
		Service:    "opcua",
		Product:    "OPC UA Server",
		Fields:     make(map[string]string),
		Confidence: 80,
	}
	
	if len(data) < 8 {
		return info, nil
	}
	
	// OPC UA Message Header解析
	// [MessageType:3][ChunkType:1][MessageSize:4]
	
	messageType := string(data[0:3])
	chunkType := data[3]
	messageSize := uint32(data[4]) | uint32(data[5])<<8 | uint32(data[6])<<16 | uint32(data[7])<<24
	
	info.Fields["message_type"] = messageType
	info.Fields["chunk_type"] = string([]byte{chunkType})
	info.Fields["message_size"] = fmt.Sprintf("%d", messageSize)
	
	// 检查消息类型
	switch messageType {
	case "HEL":
		info.Fields["message_name"] = "Hello"
		info.Confidence = 95
		p.parseOPCUAHello(data[8:], info)
	case "ACK":
		info.Fields["message_name"] = "Acknowledge"
		info.Confidence = 95
		p.parseOPCUAAcknowledge(data[8:], info)
	case "ERR":
		info.Fields["message_name"] = "Error"
		info.Confidence = 95
	case "MSG":
		info.Fields["message_name"] = "Message"
		info.Confidence = 90
	case "OPN":
		info.Fields["message_name"] = "OpenSecureChannel"
		info.Confidence = 90
	case "CLO":
		info.Fields["message_name"] = "CloseSecureChannel"
		info.Confidence = 90
	default:
		info.Fields["message_name"] = "Unknown"
	}
	
	info.ExtraInfo = fmt.Sprintf("OPC UA %s", info.Fields["message_name"])
	
	return info, nil
}

// parseOPCUAHello 解析OPC UA Hello消息
func (p *OPCUAParser) parseOPCUAHello(data []byte, info *ParsedInfo) {
	if len(data) < 20 {
		return
	}
	
	// Hello Message Body
	// [Version:4][ReceiveBufferSize:4][SendBufferSize:4][MaxMessageSize:4][MaxChunkCount:4][EndpointUrl:String]
	
	version := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	receiveBufferSize := uint32(data[4]) | uint32(data[5])<<8 | uint32(data[6])<<16 | uint32(data[7])<<24
	sendBufferSize := uint32(data[8]) | uint32(data[9])<<8 | uint32(data[10])<<16 | uint32(data[11])<<24
	maxMessageSize := uint32(data[12]) | uint32(data[13])<<8 | uint32(data[14])<<16 | uint32(data[15])<<24
	maxChunkCount := uint32(data[16]) | uint32(data[17])<<8 | uint32(data[18])<<16 | uint32(data[19])<<24
	
	info.Fields["protocol_version"] = fmt.Sprintf("%d", version)
	info.Fields["receive_buffer_size"] = fmt.Sprintf("%d", receiveBufferSize)
	info.Fields["send_buffer_size"] = fmt.Sprintf("%d", sendBufferSize)
	info.Fields["max_message_size"] = fmt.Sprintf("%d", maxMessageSize)
	info.Fields["max_chunk_count"] = fmt.Sprintf("%d", maxChunkCount)
	
	// 解析Endpoint URL
	if len(data) > 24 {
		urlLength := uint32(data[20]) | uint32(data[21])<<8 | uint32(data[22])<<16 | uint32(data[23])<<24
		if len(data) >= int(24+urlLength) {
			endpointURL := string(data[24 : 24+urlLength])
			info.Fields["endpoint_url"] = endpointURL
		}
	}
}

// parseOPCUAAcknowledge 解析OPC UA Acknowledge消息
func (p *OPCUAParser) parseOPCUAAcknowledge(data []byte, info *ParsedInfo) {
	if len(data) < 20 {
		return
	}
	
	// Acknowledge Message Body (类似Hello)
	version := uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24
	receiveBufferSize := uint32(data[4]) | uint32(data[5])<<8 | uint32(data[6])<<16 | uint32(data[7])<<24
	sendBufferSize := uint32(data[8]) | uint32(data[9])<<8 | uint32(data[10])<<16 | uint32(data[11])<<24
	maxMessageSize := uint32(data[12]) | uint32(data[13])<<8 | uint32(data[14])<<16 | uint32(data[15])<<24
	maxChunkCount := uint32(data[16]) | uint32(data[17])<<8 | uint32(data[18])<<16 | uint32(data[19])<<24
	
	info.Fields["server_protocol_version"] = fmt.Sprintf("%d", version)
	info.Fields["server_receive_buffer_size"] = fmt.Sprintf("%d", receiveBufferSize)
	info.Fields["server_send_buffer_size"] = fmt.Sprintf("%d", sendBufferSize)
	info.Fields["server_max_message_size"] = fmt.Sprintf("%d", maxMessageSize)
	info.Fields["server_max_chunk_count"] = fmt.Sprintf("%d", maxChunkCount)
}

func (p *OPCUAParser) GetProtocol() string { return "opcua" }
func (p *OPCUAParser) GetConfidence(data []byte) int {
	if len(data) >= 4 {
		messageType := string(data[0:3])
		validTypes := []string{"HEL", "ACK", "ERR", "MSG", "OPN", "CLO"}
		
		for _, validType := range validTypes {
			if messageType == validType {
				return 95
			}
		}
	}
	return 0
}

// S7Parser 西门子S7协议解析器
type S7Parser struct{}

func (p *S7Parser) Parse(data []byte) (*ParsedInfo, error) {
	info := &ParsedInfo{
		Protocol:   "s7",
		Service:    "s7",
		Product:    "Siemens S7 PLC",
		Fields:     make(map[string]string),
		Confidence: 80,
	}
	
	if len(data) < 4 {
		return info, nil
	}
	
	// 检查TPKT Header (RFC 1006)
	if data[0] == 0x03 && data[1] == 0x00 {
		info.Confidence = 90
		info.Fields["tpkt_version"] = "3"
		
		tpktLength := uint16(data[2])<<8 | uint16(data[3])
		info.Fields["tpkt_length"] = fmt.Sprintf("%d", tpktLength)
		
		// 解析COTP Header
		if len(data) > 4 {
			p.parseCOTPHeader(data[4:], info)
		}
		
		info.ExtraInfo = "Siemens S7 Communication"
	}
	
	return info, nil
}

// parseCOTPHeader 解析COTP头部
func (p *S7Parser) parseCOTPHeader(data []byte, info *ParsedInfo) {
	if len(data) < 1 {
		return
	}
	
	cotpLength := data[0]
	info.Fields["cotp_length"] = fmt.Sprintf("%d", cotpLength)
	
	if len(data) > 1 {
		pduType := data[1]
		info.Fields["cotp_pdu_type"] = fmt.Sprintf("0x%02x", pduType)
		
		// 解析PDU类型
		switch pduType {
		case 0xE0:
			info.Fields["cotp_pdu_name"] = "Connection Request (CR)"
			info.Confidence = 95
		case 0xD0:
			info.Fields["cotp_pdu_name"] = "Connection Confirm (CC)"
			info.Confidence = 95
		case 0x80:
			info.Fields["cotp_pdu_name"] = "Disconnect Request (DR)"
		case 0xC0:
			info.Fields["cotp_pdu_name"] = "Disconnect Confirm (DC)"
		case 0xF0:
			info.Fields["cotp_pdu_name"] = "Data (DT)"
		default:
			info.Fields["cotp_pdu_name"] = "Unknown"
		}
		
		// 解析连接参数
		if pduType == 0xE0 || pduType == 0xD0 {
			p.parseS7ConnectionParams(data[2:], info)
		}
	}
}

// parseS7ConnectionParams 解析S7连接参数
func (p *S7Parser) parseS7ConnectionParams(data []byte, info *ParsedInfo) {
	if len(data) < 4 {
		return
	}
	
	// 解析目标引用和源引用
	destRef := uint16(data[0])<<8 | uint16(data[1])
	srcRef := uint16(data[2])<<8 | uint16(data[3])
	
	info.Fields["destination_reference"] = fmt.Sprintf("%d", destRef)
	info.Fields["source_reference"] = fmt.Sprintf("%d", srcRef)
	
	if len(data) > 4 {
		classOption := data[4]
		info.Fields["class_option"] = fmt.Sprintf("0x%02x", classOption)
		
		// 解析类别
		class := (classOption >> 4) & 0x0F
		info.Fields["transport_class"] = fmt.Sprintf("%d", class)
	}
}

func (p *S7Parser) GetProtocol() string { return "s7" }
func (p *S7Parser) GetConfidence(data []byte) int {
	if len(data) >= 4 && data[0] == 0x03 && data[1] == 0x00 {
		// 检查COTP PDU类型
		if len(data) > 5 {
			pduType := data[5]
			if pduType == 0xE0 || pduType == 0xD0 || pduType == 0xF0 {
				return 95
			}
		}
		return 80
	}
	return 0
}