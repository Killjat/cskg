package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// NewProbeEngine 创建探测引擎
func NewProbeEngine(config *ProbeConfig) *ProbeEngine {
	if config == nil {
		config = DefaultProbeConfig()
	}
	
	engine := &ProbeEngine{
		config:  config,
		probes:  make(map[string]*Probe),
		parsers: make(map[string]ProtocolParser),
		stats: &ProbeStats{
			ProtocolCounts: make(map[string]int),
		},
	}
	
	// 加载内置探测
	loader := NewProbeLoader()
	engine.probes = loader.LoadBuiltinProbes()
	
	// 加载内置解析器
	engine.loadBuiltinParsers()
	
	return engine
}

// loadBuiltinParsers 加载内置协议解析器
func (pe *ProbeEngine) loadBuiltinParsers() {
	pe.parsers["http"] = &HTTPParser{}
	pe.parsers["https"] = NewHTTPSParser()
	pe.parsers["tls"] = &TLSParser{}
	pe.parsers["ssh"] = &SSHParser{}
	pe.parsers["ftp"] = &FTPParser{}
	pe.parsers["smtp"] = &SMTPParser{}
	pe.parsers["mysql"] = &MySQLParser{}
	pe.parsers["redis"] = &RedisParser{}
	pe.parsers["postgresql"] = &PostgreSQLParser{}
	pe.parsers["dns"] = &DNSParser{}
	pe.parsers["snmp"] = &SNMPParser{}
	pe.parsers["telnet"] = &TelnetParser{}
	pe.parsers["pop3"] = &POP3Parser{}
	pe.parsers["imap"] = &IMAPParser{}
	pe.parsers["mqtt"] = &MQTTParser{}
	pe.parsers["mqtt-ws"] = NewMQTTWebSocketParser()
	pe.parsers["rtsp"] = &RTSPParser{}
	pe.parsers["onvif"] = &ONVIFParser{}
	pe.parsers["onvif-http"] = &ONVIFParser{}
	pe.parsers["hikvision"] = &HikvisionParser{}
	pe.parsers["dahua"] = &DahuaParser{}
	
	// 工控协议解析器
	pe.parsers["modbus"] = &ModbusParser{}
	pe.parsers["dnp3"] = &DNP3Parser{}
	pe.parsers["bacnet"] = &BACnetParser{}
	pe.parsers["opcua"] = &OPCUAParser{}
	pe.parsers["s7"] = &S7Parser{}
}

// ProbeTarget 探测单个目标
func (pe *ProbeEngine) ProbeTarget(target Target) ([]*ProbeResult, error) {
	return pe.ProbeTargetWithMode(target, "all")
}

// ProbeTargetWithMode 使用指定模式探测目标
func (pe *ProbeEngine) ProbeTargetWithMode(target Target, mode string) ([]*ProbeResult, error) {
	loader := NewProbeLoader()
	loader.LoadBuiltinProbes() // 确保加载了探测
	
	var probes []*Probe
	
	switch mode {
	case "port":
		// 仅使用端口相关的探测
		probes = loader.GetProbesByPort(target.Port)
		if len(probes) == 0 {
			if nullProbe, exists := pe.probes["NULL"]; exists {
				probes = []*Probe{nullProbe}
			}
		}
		
	case "all":
		// 使用所有探测，但优先端口相关的
		portProbes := loader.GetProbesByPort(target.Port)
		allProbes := loader.GetAllProbes()
		
		probeMap := make(map[string]bool)
		
		// 先添加端口相关的探测
		for _, probe := range portProbes {
			probes = append(probes, probe)
			probeMap[probe.Name] = true
		}
		
		// 再添加其他探测（避免重复）
		for _, probe := range allProbes {
			if !probeMap[probe.Name] {
				probes = append(probes, probe)
			}
		}
		
	case "smart":
		// 智能模式：先用常见探测，根据结果决定是否继续
		probes = pe.getSmartProbes(target, loader)
		
	default:
		// 默认使用所有探测
		allProbes := loader.GetAllProbes()
		for _, probe := range allProbes {
			probes = append(probes, probe)
		}
	}
	
	// 如果还是没有探测，至少使用NULL探测
	if len(probes) == 0 {
		if nullProbe, exists := pe.probes["NULL"]; exists {
			probes = []*Probe{nullProbe}
		}
	}
	
	var results []*ProbeResult
	var wg sync.WaitGroup
	resultChan := make(chan *ProbeResult, len(probes))
	
	// 限制并发数
	semaphore := make(chan struct{}, pe.config.MaxConcurrency)
	
	for _, probe := range probes {
		wg.Add(1)
		go func(p *Probe) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			result := pe.executeProbe(target, p)
			resultChan <- result
		}(probe)
	}
	
	// 等待所有探测完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// 收集结果
	for result := range resultChan {
		results = append(results, result)
		pe.updateStats(result)
	}
	
	return results, nil
}

// executeProbe 执行单个探测
func (pe *ProbeEngine) executeProbe(target Target, probe *Probe) *ProbeResult {
	result := &ProbeResult{
		Target:    fmt.Sprintf("%s:%d", target.Host, target.Port),
		Port:      target.Port,
		ProbeName: probe.Name,
		Protocol:  probe.Protocol,
		Timestamp: time.Now(),
	}
	
	start := time.Now()
	defer func() {
		result.Duration = time.Since(start)
	}()
	
	// 根据探测类型执行
	var response []byte
	var err error
	
	if probe.Type == ProbeTypeTCP {
		response, err = pe.executeTCPProbe(target, probe)
	} else if probe.Type == ProbeTypeUDP {
		response, err = pe.executeUDPProbe(target, probe)
	} else {
		err = fmt.Errorf("unsupported probe type: %s", probe.Type)
	}
	
	if err != nil {
		result.Error = err.Error()
		result.Success = false
		return result
	}
	
	result.Success = true
	result.Response = response
	result.ResponseHex = hex.EncodeToString(response)
	
	// 协议解析
	if parser, exists := pe.parsers[probe.Protocol]; exists {
		if parsedInfo, parseErr := parser.Parse(response); parseErr == nil {
			result.ParsedInfo = parsedInfo
			result.Banner = pe.generateStructuredBanner(response, probe.Protocol, parsedInfo)
		}
	} else {
		// 默认banner提取
		result.Banner = pe.extractBanner(response, probe.Protocol)
	}
	
	return result
}

// executeTCPProbe 执行TCP探测
func (pe *ProbeEngine) executeTCPProbe(target Target, probe *Probe) ([]byte, error) {
	// 创建连接
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", target.Host, target.Port), pe.config.ConnectTimeout)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %v", err)
	}
	defer conn.Close()
	
	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(pe.config.ReadTimeout))
	
	// 发送探测载荷
	if len(probe.Payload) > 0 {
		_, err = conn.Write(probe.Payload)
		if err != nil {
			return nil, fmt.Errorf("write failed: %v", err)
		}
	}
	
	// 读取响应
	buffer := make([]byte, pe.config.MaxResponseSize)
	n, err := conn.Read(buffer)
	if err != nil {
		// 对于某些服务（如SSH），连接后立即发送banner，不需要发送数据
		if n > 0 {
			return buffer[:n], nil
		}
		return nil, fmt.Errorf("read failed: %v", err)
	}
	
	return buffer[:n], nil
}

// executeUDPProbe 执行UDP探测
func (pe *ProbeEngine) executeUDPProbe(target Target, probe *Probe) ([]byte, error) {
	// 创建UDP连接
	conn, err := net.DialTimeout("udp", fmt.Sprintf("%s:%d", target.Host, target.Port), pe.config.ConnectTimeout)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %v", err)
	}
	defer conn.Close()
	
	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(pe.config.ReadTimeout))
	
	// 发送探测载荷
	if len(probe.Payload) > 0 {
		_, err = conn.Write(probe.Payload)
		if err != nil {
			return nil, fmt.Errorf("write failed: %v", err)
		}
	}
	
	// 读取响应
	buffer := make([]byte, pe.config.MaxResponseSize)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("read failed: %v", err)
	}
	
	return buffer[:n], nil
}

// extractBanner 提取banner信息
func (pe *ProbeEngine) extractBanner(data []byte, protocol string) string {
	if len(data) == 0 {
		return ""
	}
	
	// 转换为字符串，处理不可打印字符
	banner := ""
	for _, b := range data {
		if b >= 32 && b <= 126 {
			banner += string(b)
		} else if b == '\r' || b == '\n' || b == '\t' {
			banner += string(b)
		} else {
			banner += fmt.Sprintf("\\x%02x", b)
		}
	}
	
	// 限制banner长度
	if len(banner) > 512 {
		banner = banner[:512] + "..."
	}
	
	return banner
}

// updateStats 更新统计信息
func (pe *ProbeEngine) updateStats(result *ProbeResult) {
	pe.stats.TotalProbes++
	if result.Success {
		pe.stats.SuccessProbes++
		pe.stats.ProtocolCounts[result.Protocol]++
	} else {
		pe.stats.FailedProbes++
	}
	
	pe.stats.TotalDuration += result.Duration
	if pe.stats.TotalProbes > 0 {
		pe.stats.AvgDuration = pe.stats.TotalDuration / time.Duration(pe.stats.TotalProbes)
	}
}

// GetStats 获取统计信息
func (pe *ProbeEngine) GetStats() *ProbeStats {
	return pe.stats
}

// ProbeMultipleTargets 探测多个目标
func (pe *ProbeEngine) ProbeMultipleTargets(targets []Target) (map[string][]*ProbeResult, error) {
	results := make(map[string][]*ProbeResult)
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	// 限制并发数
	semaphore := make(chan struct{}, pe.config.MaxConcurrency)
	
	for _, target := range targets {
		wg.Add(1)
		go func(t Target) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			targetResults, err := pe.ProbeTarget(t)
			if err == nil {
				mu.Lock()
				key := fmt.Sprintf("%s:%d", t.Host, t.Port)
				results[key] = targetResults
				mu.Unlock()
			}
		}(target)
	}
	
	wg.Wait()
	return results, nil
}

// ProbeWithContext 带上下文的探测
func (pe *ProbeEngine) ProbeWithContext(ctx context.Context, target Target) ([]*ProbeResult, error) {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, pe.config.Timeout)
	defer cancel()
	
	// 在goroutine中执行探测
	resultChan := make(chan []*ProbeResult, 1)
	errorChan := make(chan error, 1)
	
	go func() {
		results, err := pe.ProbeTarget(target)
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- results
		}
	}()
	
	// 等待结果或超时
	select {
	case results := <-resultChan:
		return results, nil
	case err := <-errorChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// getSmartProbes 获取智能探测列表
func (pe *ProbeEngine) getSmartProbes(target Target, loader *ProbeLoader) []*Probe {
	var probes []*Probe
	
	// 第一阶段：使用高优先级探测（稀有度低的）
	allProbes := loader.GetAllProbes()
	
	// 按稀有度排序，优先使用常见探测
	var sortedProbes []*Probe
	for _, probe := range allProbes {
		sortedProbes = append(sortedProbes, probe)
	}
	
	// 简单排序：稀有度低的优先
	for i := 0; i < len(sortedProbes)-1; i++ {
		for j := i + 1; j < len(sortedProbes); j++ {
			if sortedProbes[i].Rarity > sortedProbes[j].Rarity {
				sortedProbes[i], sortedProbes[j] = sortedProbes[j], sortedProbes[i]
			}
		}
	}
	
	// 选择前10个最常见的探测
	maxProbes := 10
	if len(sortedProbes) < maxProbes {
		maxProbes = len(sortedProbes)
	}
	
	for i := 0; i < maxProbes; i++ {
		probes = append(probes, sortedProbes[i])
	}
	
	return probes
}
// ProbeMultipleTargetsWithMode 使用指定模式探测多个目标
func (pe *ProbeEngine) ProbeMultipleTargetsWithMode(targets []Target, mode string) (map[string][]*ProbeResult, error) {
	results := make(map[string][]*ProbeResult)
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	// 限制并发数
	semaphore := make(chan struct{}, pe.config.MaxConcurrency)
	
	for _, target := range targets {
		wg.Add(1)
		go func(t Target) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			targetResults, err := pe.ProbeTargetWithMode(t, mode)
			if err == nil {
				mu.Lock()
				key := fmt.Sprintf("%s:%d", t.Host, t.Port)
				results[key] = targetResults
				mu.Unlock()
			}
		}(target)
	}
	
	wg.Wait()
	return results, nil
}
// generateStructuredBanner 生成结构化的banner信息
func (pe *ProbeEngine) generateStructuredBanner(data []byte, protocol string, parsedInfo *ParsedInfo) string {
	if parsedInfo == nil {
		return pe.extractBanner(data, protocol)
	}
	
	var banner strings.Builder
	
	switch protocol {
	case "http":
		return pe.generateHTTPBanner(parsedInfo)
	case "https":
		return pe.generateHTTPSBanner(parsedInfo)
	case "tls":
		return pe.generateTLSBanner(parsedInfo)
	case "ssh":
		return pe.generateSSHBanner(parsedInfo)
	case "mysql":
		return pe.generateMySQLBanner(parsedInfo)
	case "mqtt":
		return pe.generateMQTTBanner(parsedInfo)
	case "mqtt-ws":
		return pe.generateMQTTWebSocketBanner(parsedInfo)
	case "rtsp":
		return pe.generateRTSPBanner(parsedInfo)
	case "onvif":
	case "onvif-http":
		return pe.generateONVIFBanner(parsedInfo)
	case "hikvision":
		return pe.generateHikvisionBanner(parsedInfo)
	case "dahua":
		return pe.generateDahuaBanner(parsedInfo)
	case "modbus":
		return pe.generateModbusBanner(parsedInfo)
	case "dnp3":
		return pe.generateDNP3Banner(parsedInfo)
	case "bacnet":
		return pe.generateBACnetBanner(parsedInfo)
	case "opcua":
		return pe.generateOPCUABanner(parsedInfo)
	case "s7":
		return pe.generateS7Banner(parsedInfo)
	case "ftp":
		return pe.generateFTPBanner(parsedInfo)
	case "smtp":
		return pe.generateSMTPBanner(parsedInfo)
	default:
		// 通用结构化banner
		if parsedInfo.Product != "" {
			banner.WriteString(parsedInfo.Product)
			if parsedInfo.Version != "" {
				banner.WriteString(" v")
				banner.WriteString(parsedInfo.Version)
			}
		}
		
		if parsedInfo.OS != "" {
			if banner.Len() > 0 {
				banner.WriteString(" on ")
			}
			banner.WriteString(parsedInfo.OS)
		}
		
		if parsedInfo.ExtraInfo != "" {
			if banner.Len() > 0 {
				banner.WriteString(" (")
			}
			banner.WriteString(parsedInfo.ExtraInfo)
			if banner.Len() > len(parsedInfo.ExtraInfo) {
				banner.WriteString(")")
			}
		}
	}
	
	if banner.Len() == 0 {
		return pe.extractBanner(data, protocol)
	}
	
	return banner.String()
}

// generateHTTPBanner 生成HTTP结构化banner
func (pe *ProbeEngine) generateHTTPBanner(info *ParsedInfo) string {
	var banner strings.Builder
	
	// HTTP状态信息
	if statusLine, exists := info.Fields["status_line"]; exists {
		banner.WriteString(statusLine)
	}
	
	// 服务器信息
	if info.Product != "" {
		banner.WriteString(" | Server: ")
		banner.WriteString(info.Product)
		if info.Version != "" {
			banner.WriteString("/")
			banner.WriteString(info.Version)
		}
	}
	
	// 操作系统
	if info.OS != "" {
		banner.WriteString(" (")
		banner.WriteString(info.OS)
		banner.WriteString(")")
	}
	
	// 技术栈
	if tech, exists := info.Fields["technologies"]; exists {
		banner.WriteString(" | Tech: ")
		banner.WriteString(tech)
	}
	
	// Powered by
	if powered, exists := info.Fields["powered_by"]; exists {
		banner.WriteString(" | Powered by: ")
		banner.WriteString(powered)
	}
	
	return banner.String()
}

// generateSSHBanner 生成SSH结构化banner
func (pe *ProbeEngine) generateSSHBanner(info *ParsedInfo) string {
	var banner strings.Builder
	
	// SSH协议版本
	if protocolVer, exists := info.Fields["protocol_version"]; exists {
		banner.WriteString("SSH-")
		banner.WriteString(protocolVer)
	}
	
	// 产品信息
	if info.Product != "" {
		if banner.Len() > 0 {
			banner.WriteString(" | ")
		}
		banner.WriteString(info.Product)
		if info.Version != "" {
			banner.WriteString(" ")
			banner.WriteString(info.Version)
		}
	}
	
	// 操作系统
	if info.OS != "" {
		banner.WriteString(" on ")
		banner.WriteString(info.OS)
		
		// Ubuntu包版本
		if ubuntuPkg, exists := info.Fields["ubuntu_package"]; exists {
			banner.WriteString(" (package: ")
			banner.WriteString(ubuntuPkg)
			banner.WriteString(")")
		}
	}
	
	// 设备类型
	if info.DeviceType != "" {
		banner.WriteString(" [")
		banner.WriteString(info.DeviceType)
		banner.WriteString("]")
	}
	
	// 额外信息
	if info.ExtraInfo != "" {
		banner.WriteString(" - ")
		banner.WriteString(info.ExtraInfo)
	}
	
	// 云服务提供商
	if cloud, exists := info.Fields["cloud_provider"]; exists {
		banner.WriteString(" (")
		banner.WriteString(cloud)
		banner.WriteString(")")
	}
	
	return banner.String()
}

// generateMySQLBanner 生成MySQL结构化banner
func (pe *ProbeEngine) generateMySQLBanner(info *ParsedInfo) string {
	var banner strings.Builder
	
	// 产品和版本
	if info.Product != "" {
		banner.WriteString(info.Product)
		if info.Version != "" {
			banner.WriteString(" ")
			banner.WriteString(info.Version)
		}
	}
	
	// 协议版本
	if protocolVer, exists := info.Fields["protocol_version"]; exists {
		banner.WriteString(" (Protocol ")
		banner.WriteString(protocolVer)
		banner.WriteString(")")
	}
	
	// 操作系统
	if info.OS != "" {
		banner.WriteString(" on ")
		banner.WriteString(info.OS)
		
		// 系统版本
		if osVer, exists := info.Fields["ubuntu_version"]; exists {
			banner.WriteString(" ")
			banner.WriteString(osVer)
		} else if rhelVer, exists := info.Fields["rhel_version"]; exists {
			banner.WriteString(" ")
			banner.WriteString(rhelVer)
		}
	}
	
	// SSL支持
	if ssl, exists := info.Fields["ssl_support"]; exists && ssl == "true" {
		banner.WriteString(" | SSL: Enabled")
	}
	
	// 日志功能
	if logging, exists := info.Fields["logging_enabled"]; exists && logging == "true" {
		banner.WriteString(" | Logging: Enabled")
	}
	
	// 云服务
	if cloud, exists := info.Fields["cloud_provider"]; exists {
		banner.WriteString(" | Cloud: ")
		banner.WriteString(cloud)
	}
	
	// 额外信息
	if info.ExtraInfo != "" {
		banner.WriteString(" | ")
		banner.WriteString(info.ExtraInfo)
	}
	
	return banner.String()
}

// generateFTPBanner 生成FTP结构化banner
func (pe *ProbeEngine) generateFTPBanner(info *ParsedInfo) string {
	var banner strings.Builder
	
	// 响应码和消息
	if code, exists := info.Fields["response_code"]; exists {
		banner.WriteString("FTP ")
		banner.WriteString(code)
		
		if message, exists := info.Fields["message"]; exists {
			banner.WriteString(" ")
			banner.WriteString(message)
		}
	}
	
	// 产品信息
	if info.Product != "" {
		banner.WriteString(" | ")
		banner.WriteString(info.Product)
		if info.Version != "" {
			banner.WriteString(" ")
			banner.WriteString(info.Version)
		}
	}
	
	return banner.String()
}

// generateSMTPBanner 生成SMTP结构化banner
func (pe *ProbeEngine) generateSMTPBanner(info *ParsedInfo) string {
	var banner strings.Builder
	
	// 响应码和消息
	if code, exists := info.Fields["response_code"]; exists {
		banner.WriteString("SMTP ")
		banner.WriteString(code)
		
		if message, exists := info.Fields["message"]; exists {
			banner.WriteString(" ")
			banner.WriteString(message)
		}
	}
	
	// 产品信息
	if info.Product != "" {
		banner.WriteString(" | ")
		banner.WriteString(info.Product)
	}
	
	return banner.String()
}
// generateTLSBanner 生成TLS结构化banner
func (pe *ProbeEngine) generateTLSBanner(info *ParsedInfo) string {
	var banner strings.Builder
	
	// TLS版本和产品
	if info.Product != "" {
		banner.WriteString(info.Product)
		if info.Version != "" {
			banner.WriteString(" ")
			banner.WriteString(info.Version)
		}
	}
	
	// 消息类型
	if msgType, exists := info.Fields["message_type"]; exists {
		banner.WriteString(" | ")
		banner.WriteString(msgType)
		
		// 握手类型
		if hsType, exists := info.Fields["handshake_type_name"]; exists {
			banner.WriteString(" (")
			banner.WriteString(hsType)
			banner.WriteString(")")
		}
		
		// Alert信息
		if alertDesc, exists := info.Fields["alert_description_name"]; exists {
			banner.WriteString(" (")
			banner.WriteString(alertDesc)
			banner.WriteString(")")
		}
	}
	
	// Cipher Suite
	if cipherName, exists := info.Fields["cipher_suite_name"]; exists {
		banner.WriteString(" | Cipher: ")
		banner.WriteString(cipherName)
		
		// 加密强度
		if strength, exists := info.Fields["encryption_strength"]; exists {
			banner.WriteString(" [")
			banner.WriteString(strength)
			banner.WriteString("]")
		}
		
		// 前向保密
		if fs, exists := info.Fields["forward_secrecy"]; exists && fs == "Yes" {
			banner.WriteString(" [PFS]")
		}
	}
	
	return banner.String()
}

// generateHTTPSBanner 生成HTTPS结构化banner
func (pe *ProbeEngine) generateHTTPSBanner(info *ParsedInfo) string {
	// HTTPS可能返回TLS握手信息或HTTP响应
	if info.Protocol == "tls" {
		return pe.generateTLSBanner(info)
	} else {
		// 如果是HTTP响应，添加HTTPS前缀
		httpBanner := pe.generateHTTPBanner(info)
		if httpBanner != "" {
			return "HTTPS | " + httpBanner
		}
	}
	
	return "HTTPS Service"
}
// generateMQTTBanner 生成MQTT结构化banner
func (pe *ProbeEngine) generateMQTTBanner(info *ParsedInfo) string {
	var banner strings.Builder
	
	// MQTT产品和版本
	if info.Product != "" {
		banner.WriteString(info.Product)
		if info.Version != "" {
			banner.WriteString(" v")
			banner.WriteString(info.Version)
		}
	}
	
	// 消息类型
	if msgType, exists := info.Fields["message_type_name"]; exists {
		if banner.Len() > 0 {
			banner.WriteString(" | ")
		}
		banner.WriteString(msgType)
		
		// 返回码信息 (CONNACK)
		if returnCode, exists := info.Fields["return_code_name"]; exists {
			banner.WriteString(" (")
			banner.WriteString(returnCode)
			banner.WriteString(")")
		}
		
		// 主题信息 (PUBLISH)
		if topic, exists := info.Fields["topic"]; exists {
			banner.WriteString(" Topic: ")
			banner.WriteString(topic)
		}
	}
	
	// 协议信息
	if protocolName, exists := info.Fields["protocol_name"]; exists {
		banner.WriteString(" | Protocol: ")
		banner.WriteString(protocolName)
		
		if protocolLevel, exists := info.Fields["protocol_level"]; exists {
			banner.WriteString(" Level ")
			banner.WriteString(protocolLevel)
		}
	}
	
	// Keep Alive
	if keepAlive, exists := info.Fields["keep_alive"]; exists {
		banner.WriteString(" | Keep-Alive: ")
		banner.WriteString(keepAlive)
		banner.WriteString("s")
	}
	
	// 额外信息
	if info.ExtraInfo != "" {
		banner.WriteString(" | ")
		banner.WriteString(info.ExtraInfo)
	}
	
	return banner.String()
}

// generateMQTTWebSocketBanner 生成MQTT WebSocket结构化banner
func (pe *ProbeEngine) generateMQTTWebSocketBanner(info *ParsedInfo) string {
	var banner strings.Builder
	
	banner.WriteString("MQTT over WebSocket")
	
	// 如果有HTTP信息，添加服务器信息
	if server, exists := info.Fields["server"]; exists {
		banner.WriteString(" | Server: ")
		banner.WriteString(server)
	}
	
	// 如果有MQTT特定信息，添加
	if msgType, exists := info.Fields["message_type_name"]; exists {
		banner.WriteString(" | MQTT: ")
		banner.WriteString(msgType)
	}
	
	// 额外信息
	if info.ExtraInfo != "" {
		banner.WriteString(" | ")
		banner.WriteString(info.ExtraInfo)
	}
	
	return banner.String()
}
// generateRTSPBanner 生成RTSP结构化banner
func (pe *ProbeEngine) generateRTSPBanner(info *ParsedInfo) string {
	var banner strings.Builder
	
	// RTSP版本和状态
	if statusLine, exists := info.Fields["status_line"]; exists {
		banner.WriteString(statusLine)
	}
	
	// 服务器信息
	if server, exists := info.Fields["header_server"]; exists {
		banner.WriteString(" | Server: ")
		banner.WriteString(server)
	}
	
	// 产品信息
	if info.Product != "" && info.Product != "RTSP Server" {
		banner.WriteString(" | ")
		banner.WriteString(info.Product)
		if info.Version != "" {
			banner.WriteString(" v")
			banner.WriteString(info.Version)
		}
	}
	
	// 支持的方法
	if methods, exists := info.Fields["supported_methods"]; exists {
		banner.WriteString(" | Methods: ")
		banner.WriteString(methods)
	}
	
	// 额外信息
	if info.ExtraInfo != "" {
		banner.WriteString(" | ")
		banner.WriteString(info.ExtraInfo)
	}
	
	return banner.String()
}

// generateONVIFBanner 生成ONVIF结构化banner
func (pe *ProbeEngine) generateONVIFBanner(info *ParsedInfo) string {
	var banner strings.Builder
	
	// ONVIF设备信息
	if info.Product != "" {
		banner.WriteString(info.Product)
		if info.Version != "" {
			banner.WriteString(" (Firmware: ")
			banner.WriteString(info.Version)
			banner.WriteString(")")
		}
	}
	
	// 制造商信息
	if manufacturer, exists := info.Fields["manufacturer"]; exists {
		if banner.Len() > 0 {
			banner.WriteString(" | ")
		}
		banner.WriteString("Manufacturer: ")
		banner.WriteString(manufacturer)
	}
	
	// 型号信息
	if model, exists := info.Fields["model"]; exists {
		banner.WriteString(" | Model: ")
		banner.WriteString(model)
	}
	
	// 序列号
	if serial, exists := info.Fields["serial_number"]; exists {
		banner.WriteString(" | S/N: ")
		banner.WriteString(serial)
	}
	
	// 响应类型
	if responseType, exists := info.Fields["response_type"]; exists {
		banner.WriteString(" | ")
		banner.WriteString(responseType)
	}
	
	// 设备地址 (WS-Discovery)
	if addresses, exists := info.Fields["device_addresses"]; exists {
		banner.WriteString(" | Addresses: ")
		banner.WriteString(addresses)
	}
	
	return banner.String()
}

// generateHikvisionBanner 生成海康威视结构化banner
func (pe *ProbeEngine) generateHikvisionBanner(info *ParsedInfo) string {
	var banner strings.Builder
	
	// 产品信息
	banner.WriteString("Hikvision IP Camera")
	
	// 固件版本
	if info.Version != "" {
		banner.WriteString(" (Firmware: ")
		banner.WriteString(info.Version)
		banner.WriteString(")")
	}
	
	// HTTP状态
	if statusLine, exists := info.Fields["status_line"]; exists {
		banner.WriteString(" | ")
		banner.WriteString(statusLine)
	}
	
	// 型号信息
	if model, exists := info.Fields["model"]; exists {
		banner.WriteString(" | Model: ")
		banner.WriteString(model)
	}
	
	// 序列号
	if serial, exists := info.Fields["serial_number"]; exists {
		banner.WriteString(" | S/N: ")
		banner.WriteString(serial)
	}
	
	// 服务器信息
	if server, exists := info.Fields["server"]; exists {
		banner.WriteString(" | Server: ")
		banner.WriteString(server)
	}
	
	// 认证状态
	if authRequired, exists := info.Fields["auth_required"]; exists && authRequired == "true" {
		banner.WriteString(" | Auth Required")
	}
	
	// 响应类型
	if responseType, exists := info.Fields["response_type"]; exists {
		banner.WriteString(" | ")
		banner.WriteString(responseType)
	}
	
	return banner.String()
}

// generateDahuaBanner 生成大华结构化banner
func (pe *ProbeEngine) generateDahuaBanner(info *ParsedInfo) string {
	var banner strings.Builder
	
	// 产品信息
	banner.WriteString("Dahua IP Camera")
	
	// 协议头部
	if header, exists := info.Fields["protocol_header"]; exists {
		banner.WriteString(" | Protocol Header: ")
		banner.WriteString(header)
	}
	
	// 命令信息
	if cmdName, exists := info.Fields["command_name"]; exists {
		banner.WriteString(" | Command: ")
		banner.WriteString(cmdName)
	} else if cmdType, exists := info.Fields["command_type"]; exists {
		banner.WriteString(" | Command Type: ")
		banner.WriteString(cmdType)
	}
	
	// 会话ID
	if sessionID, exists := info.Fields["session_id"]; exists {
		banner.WriteString(" | Session: ")
		banner.WriteString(sessionID)
	}
	
	// 包长度
	if length, exists := info.Fields["packet_length"]; exists {
		banner.WriteString(" | Length: ")
		banner.WriteString(length)
	}
	
	return banner.String()
}
// generateModbusBanner 生成Modbus结构化banner
func (pe *ProbeEngine) generateModbusBanner(info *ParsedInfo) string {
	var banner strings.Builder
	
	// Modbus产品信息
	if info.Product != "" {
		banner.WriteString(info.Product)
	}
	
	// 功能码信息
	if functionName, exists := info.Fields["function_name"]; exists {
		if banner.Len() > 0 {
			banner.WriteString(" | ")
		}
		banner.WriteString("Function: ")
		banner.WriteString(functionName)
		
		if functionCode, exists := info.Fields["function_code"]; exists {
			banner.WriteString(" (")
			banner.WriteString(functionCode)
			banner.WriteString(")")
		}
	}
	
	// 单元ID
	if unitID, exists := info.Fields["unit_id"]; exists {
		banner.WriteString(" | Unit ID: ")
		banner.WriteString(unitID)
	}
	
	// 异常响应
	if exception, exists := info.Fields["exception_response"]; exists && exception == "true" {
		banner.WriteString(" | EXCEPTION")
		if exceptionCode, exists := info.Fields["exception_code"]; exists {
			banner.WriteString(" (Code: ")
			banner.WriteString(exceptionCode)
			banner.WriteString(")")
		}
	}
	
	// 额外信息
	if info.ExtraInfo != "" {
		banner.WriteString(" | ")
		banner.WriteString(info.ExtraInfo)
	}
	
	return banner.String()
}

// generateDNP3Banner 生成DNP3结构化banner
func (pe *ProbeEngine) generateDNP3Banner(info *ParsedInfo) string {
	var banner strings.Builder
	
	// DNP3产品信息
	if info.Product != "" {
		banner.WriteString(info.Product)
	}
	
	// 功能信息
	if functionName, exists := info.Fields["function_name"]; exists {
		if banner.Len() > 0 {
			banner.WriteString(" | ")
		}
		banner.WriteString("Function: ")
		banner.WriteString(functionName)
	}
	
	// 源和目标地址
	if src, exists := info.Fields["source"]; exists {
		banner.WriteString(" | Src: ")
		banner.WriteString(src)
	}
	
	if dest, exists := info.Fields["destination"]; exists {
		banner.WriteString(" | Dest: ")
		banner.WriteString(dest)
	}
	
	// 控制信息
	if direction, exists := info.Fields["direction"]; exists {
		if direction == "1" {
			banner.WriteString(" | Direction: Master->Outstation")
		} else {
			banner.WriteString(" | Direction: Outstation->Master")
		}
	}
	
	// 额外信息
	if info.ExtraInfo != "" {
		banner.WriteString(" | ")
		banner.WriteString(info.ExtraInfo)
	}
	
	return banner.String()
}

// generateBACnetBanner 生成BACnet结构化banner
func (pe *ProbeEngine) generateBACnetBanner(info *ParsedInfo) string {
	var banner strings.Builder
	
	// BACnet产品信息
	if info.Product != "" {
		banner.WriteString(info.Product)
	}
	
	// 网络类型
	if networkType, exists := info.Fields["network_type"]; exists {
		if banner.Len() > 0 {
			banner.WriteString(" | ")
		}
		banner.WriteString(networkType)
	}
	
	// BVLC功能
	if bvlcFunction, exists := info.Fields["bvlc_function_name"]; exists {
		banner.WriteString(" | ")
		banner.WriteString(bvlcFunction)
	}
	
	// NPDU版本
	if npduVersion, exists := info.Fields["npdu_version"]; exists {
		banner.WriteString(" | NPDU v")
		banner.WriteString(npduVersion)
	}
	
	// 网络优先级
	if priority, exists := info.Fields["network_priority"]; exists && priority != "0" {
		banner.WriteString(" | Priority: ")
		banner.WriteString(priority)
	}
	
	// 额外信息
	if info.ExtraInfo != "" {
		banner.WriteString(" | ")
		banner.WriteString(info.ExtraInfo)
	}
	
	return banner.String()
}

// generateOPCUABanner 生成OPC UA结构化banner
func (pe *ProbeEngine) generateOPCUABanner(info *ParsedInfo) string {
	var banner strings.Builder
	
	// OPC UA产品信息
	if info.Product != "" {
		banner.WriteString(info.Product)
	}
	
	// 消息类型
	if messageName, exists := info.Fields["message_name"]; exists {
		if banner.Len() > 0 {
			banner.WriteString(" | ")
		}
		banner.WriteString("Message: ")
		banner.WriteString(messageName)
	}
	
	// 协议版本
	if version, exists := info.Fields["protocol_version"]; exists {
		banner.WriteString(" | Protocol v")
		banner.WriteString(version)
	} else if serverVersion, exists := info.Fields["server_protocol_version"]; exists {
		banner.WriteString(" | Server Protocol v")
		banner.WriteString(serverVersion)
	}
	
	// 缓冲区大小
	if receiveBuffer, exists := info.Fields["receive_buffer_size"]; exists {
		banner.WriteString(" | RxBuffer: ")
		banner.WriteString(receiveBuffer)
	}
	
	if sendBuffer, exists := info.Fields["send_buffer_size"]; exists {
		banner.WriteString(" | TxBuffer: ")
		banner.WriteString(sendBuffer)
	}
	
	// Endpoint URL
	if endpointURL, exists := info.Fields["endpoint_url"]; exists {
		banner.WriteString(" | Endpoint: ")
		banner.WriteString(endpointURL)
	}
	
	// 额外信息
	if info.ExtraInfo != "" {
		banner.WriteString(" | ")
		banner.WriteString(info.ExtraInfo)
	}
	
	return banner.String()
}

// generateS7Banner 生成S7结构化banner
func (pe *ProbeEngine) generateS7Banner(info *ParsedInfo) string {
	var banner strings.Builder
	
	// S7产品信息
	if info.Product != "" {
		banner.WriteString(info.Product)
	}
	
	// TPKT信息
	if tpktVersion, exists := info.Fields["tpkt_version"]; exists {
		if banner.Len() > 0 {
			banner.WriteString(" | ")
		}
		banner.WriteString("TPKT v")
		banner.WriteString(tpktVersion)
	}
	
	// COTP PDU类型
	if cotpPduName, exists := info.Fields["cotp_pdu_name"]; exists {
		banner.WriteString(" | COTP: ")
		banner.WriteString(cotpPduName)
	}
	
	// 传输类别
	if transportClass, exists := info.Fields["transport_class"]; exists {
		banner.WriteString(" | Class ")
		banner.WriteString(transportClass)
	}
	
	// 连接引用
	if srcRef, exists := info.Fields["source_reference"]; exists {
		banner.WriteString(" | SrcRef: ")
		banner.WriteString(srcRef)
	}
	
	if destRef, exists := info.Fields["destination_reference"]; exists {
		banner.WriteString(" | DestRef: ")
		banner.WriteString(destRef)
	}
	
	// 额外信息
	if info.ExtraInfo != "" {
		banner.WriteString(" | ")
		banner.WriteString(info.ExtraInfo)
	}
	
	return banner.String()
}