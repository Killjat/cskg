package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// registerModbusScripts 注册Modbus脚本
func (se *ScriptEngine) registerModbusScripts() {
	scripts := []*Script{
		{
			Name:        "modbus-device-info",
			Protocol:    "modbus",
			Category:    CategoryDiscovery,
			Description: "收集Modbus设备信息",
			Author:      "Script Engine Team",
			Version:     "1.0",
			Execute:     executeModbusDeviceInfo,
		},
		{
			Name:        "modbus-function-scan",
			Protocol:    "modbus",
			Category:    CategoryDiscovery,
			Description: "扫描Modbus功能码",
			Author:      "Script Engine Team",
			Version:     "1.0",
			Execute:     executeModbusFunctionScan,
		},
		{
			Name:        "modbus-coil-enum",
			Protocol:    "modbus",
			Category:    CategoryDiscovery,
			Description: "枚举Modbus线圈",
			Author:      "Script Engine Team",
			Version:     "1.0",
			Execute:     executeModbusCoilEnum,
		},
		{
			Name:        "modbus-register-read",
			Protocol:    "modbus",
			Category:    CategoryDiscovery,
			Description: "读取Modbus寄存器",
			Author:      "Script Engine Team",
			Version:     "1.0",
			Execute:     executeModbusRegisterRead,
		},
		{
			Name:        "modbus-auth-bypass",
			Protocol:    "modbus",
			Category:    CategoryVulnerability,
			Description: "检测Modbus认证绕过漏洞",
			Author:      "Script Engine Team",
			Version:     "1.0",
			Execute:     executeModbusAuthBypass,
		},
		{
			Name:        "modbus-dos-test",
			Protocol:    "modbus",
			Category:    CategoryVulnerability,
			Description: "检测Modbus拒绝服务漏洞",
			Author:      "Script Engine Team",
			Version:     "1.0",
			Execute:     executeModbusDosTest,
		},
	}

	for _, script := range scripts {
		se.registry.Register(script)
	}
}

// executeModbusDeviceInfo 执行Modbus设备信息收集
func executeModbusDeviceInfo(target Target, ctx *ScriptContext) *ScriptResult {
	result := &ScriptResult{
		Success:  false,
		Findings: make(map[string]interface{}),
	}

	ctx.Logger.Debug("开始收集Modbus设备信息")

	// 连接到目标
	conn, err := net.DialTimeout("tcp", target.String(), ctx.Timeout)
	if err != nil {
		result.Error = fmt.Sprintf("连接失败: %v", err)
		return result
	}
	defer conn.Close()

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(ctx.Timeout))

	// 发送设备识别请求 (功能码 0x11)
	deviceIdRequest := []byte{
		0x00, 0x01, // 事务ID
		0x00, 0x00, // 协议ID
		0x00, 0x06, // 长度
		0x01,       // 单元ID
		0x11,       // 功能码: Read Device Identification
		0x01,       // MEI类型
		0x00,       // 读取设备ID代码
		0x00,       // 对象ID
	}

	_, err = conn.Write(deviceIdRequest)
	if err != nil {
		result.Error = fmt.Sprintf("发送请求失败: %v", err)
		return result
	}

	// 读取响应
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		result.Error = fmt.Sprintf("读取响应失败: %v", err)
		return result
	}

	response = response[:n]
	ctx.Logger.Debug("收到响应: %d 字节", n)

	// 解析响应
	if len(response) < 8 {
		result.Error = "响应长度不足"
		return result
	}

	// 检查功能码
	if response[7] == 0x11 {
		// 成功响应
		deviceInfo := parseModbusDeviceInfo(response[8:])
		result.Findings = deviceInfo
		result.Success = true
		
		ctx.Logger.Debug("成功获取设备信息")
	} else if response[7] == 0x91 {
		// 异常响应
		if len(response) > 8 {
			exceptionCode := response[8]
			result.Findings["exception_code"] = fmt.Sprintf("0x%02X", exceptionCode)
			result.Findings["exception_desc"] = getModbusExceptionDescription(exceptionCode)
		}
		result.Success = true // 即使是异常响应，也说明设备存在
	} else {
		result.Error = fmt.Sprintf("未知响应功能码: 0x%02X", response[7])
		return result
	}

	// 尝试获取更多信息
	additionalInfo := getModbusAdditionalInfo(conn, ctx)
	for key, value := range additionalInfo {
		result.Findings[key] = value
	}

	return result
}

// parseModbusDeviceInfo 解析Modbus设备信息
func parseModbusDeviceInfo(data []byte) map[string]interface{} {
	info := make(map[string]interface{})
	
	if len(data) < 4 {
		return info
	}

	// 解析MEI响应
	meiType := data[0]
	readDeviceCode := data[1]
	conformityLevel := data[2]
	moreFollows := data[3]

	info["mei_type"] = fmt.Sprintf("0x%02X", meiType)
	info["read_device_code"] = fmt.Sprintf("0x%02X", readDeviceCode)
	info["conformity_level"] = fmt.Sprintf("0x%02X", conformityLevel)
	info["more_follows"] = moreFollows > 0

	// 解析对象列表
	if len(data) > 5 {
		numObjects := data[4]
		info["num_objects"] = numObjects

		offset := 5
		objects := make(map[string]string)

		for i := 0; i < int(numObjects) && offset < len(data); i++ {
			if offset+2 >= len(data) {
				break
			}

			objectId := data[offset]
			objectLen := data[offset+1]
			offset += 2

			if offset+int(objectLen) > len(data) {
				break
			}

			objectValue := string(data[offset : offset+int(objectLen)])
			offset += int(objectLen)

			switch objectId {
			case 0x00:
				objects["vendor_name"] = objectValue
				info["vendor"] = objectValue
			case 0x01:
				objects["product_code"] = objectValue
				info["product"] = objectValue
			case 0x02:
				objects["major_minor_revision"] = objectValue
				info["version"] = objectValue
			case 0x03:
				objects["vendor_url"] = objectValue
			case 0x04:
				objects["product_name"] = objectValue
			case 0x05:
				objects["model_name"] = objectValue
				info["model"] = objectValue
			case 0x06:
				objects["user_application_name"] = objectValue
			default:
				objects[fmt.Sprintf("object_%02x", objectId)] = objectValue
			}
		}

		info["objects"] = objects
	}

	return info
}

// getModbusExceptionDescription 获取Modbus异常描述
func getModbusExceptionDescription(code byte) string {
	descriptions := map[byte]string{
		0x01: "Illegal Function",
		0x02: "Illegal Data Address",
		0x03: "Illegal Data Value",
		0x04: "Slave Device Failure",
		0x05: "Acknowledge",
		0x06: "Slave Device Busy",
		0x08: "Memory Parity Error",
		0x0A: "Gateway Path Unavailable",
		0x0B: "Gateway Target Device Failed to Respond",
	}

	if desc, exists := descriptions[code]; exists {
		return desc
	}
	return "Unknown Exception"
}

// getModbusAdditionalInfo 获取Modbus附加信息
func getModbusAdditionalInfo(conn net.Conn, ctx *ScriptContext) map[string]interface{} {
	info := make(map[string]interface{})

	// 尝试读取保持寄存器 (功能码 0x03)
	readHoldingRequest := []byte{
		0x00, 0x02, // 事务ID
		0x00, 0x00, // 协议ID
		0x00, 0x06, // 长度
		0x01,       // 单元ID
		0x03,       // 功能码: Read Holding Registers
		0x00, 0x00, // 起始地址
		0x00, 0x01, // 寄存器数量
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write(readHoldingRequest)
	if err == nil {
		response := make([]byte, 256)
		n, err := conn.Read(response)
		if err == nil && n >= 9 {
			if response[7] == 0x03 {
				info["holding_registers_accessible"] = true
				if n >= 11 {
					value := binary.BigEndian.Uint16(response[9:11])
					info["first_holding_register"] = value
				}
			} else if response[7] == 0x83 {
				info["holding_registers_accessible"] = false
				if n > 8 {
					info["holding_register_exception"] = fmt.Sprintf("0x%02X", response[8])
				}
			}
		}
	}

	return info
}

// executeModbusFunctionScan 执行Modbus功能码扫描
func executeModbusFunctionScan(target Target, ctx *ScriptContext) *ScriptResult {
	result := &ScriptResult{
		Success:  false,
		Findings: make(map[string]interface{}),
	}

	ctx.Logger.Debug("开始扫描Modbus功能码")

	// 连接到目标
	conn, err := net.DialTimeout("tcp", target.String(), ctx.Timeout)
	if err != nil {
		result.Error = fmt.Sprintf("连接失败: %v", err)
		return result
	}
	defer conn.Close()

	// 要测试的功能码
	functionCodes := []byte{
		0x01, // Read Coils
		0x02, // Read Discrete Inputs
		0x03, // Read Holding Registers
		0x04, // Read Input Registers
		0x05, // Write Single Coil
		0x06, // Write Single Register
		0x0F, // Write Multiple Coils
		0x10, // Write Multiple Registers
		0x11, // Read Device Identification
		0x14, // Read File Record
		0x15, // Write File Record
		0x16, // Mask Write Register
		0x17, // Read/Write Multiple Registers
	}

	supportedFunctions := make([]string, 0)
	unsupportedFunctions := make([]string, 0)
	functionDetails := make(map[string]interface{})

	for i, funcCode := range functionCodes {
		ctx.Logger.Debug("测试功能码: 0x%02X", funcCode)

		// 构造测试请求
		request := buildModbusTestRequest(funcCode, i)
		
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, err := conn.Write(request)
		if err != nil {
			continue
		}

		response := make([]byte, 256)
		n, err := conn.Read(response)
		if err != nil {
			continue
		}

		if n >= 8 {
			responseFunc := response[7]
			funcName := getModbusFunctionName(funcCode)

			if responseFunc == funcCode {
				// 正常响应
				supportedFunctions = append(supportedFunctions, funcName)
				functionDetails[funcName] = "supported"
			} else if responseFunc == (funcCode | 0x80) {
				// 异常响应
				if n > 8 {
					exceptionCode := response[8]
					if exceptionCode == 0x01 {
						// Illegal Function - 不支持
						unsupportedFunctions = append(unsupportedFunctions, funcName)
						functionDetails[funcName] = "unsupported"
					} else {
						// 其他异常 - 支持但参数错误
						supportedFunctions = append(supportedFunctions, funcName)
						functionDetails[funcName] = fmt.Sprintf("supported (exception: 0x%02X)", exceptionCode)
					}
				}
			}
		}

		// 添加延迟避免过于频繁的请求
		time.Sleep(100 * time.Millisecond)
	}

	result.Findings["supported_functions"] = supportedFunctions
	result.Findings["unsupported_functions"] = unsupportedFunctions
	result.Findings["function_details"] = functionDetails
	result.Findings["total_tested"] = len(functionCodes)
	result.Findings["total_supported"] = len(supportedFunctions)

	result.Success = true
	ctx.Logger.Debug("功能码扫描完成，支持 %d/%d 个功能", len(supportedFunctions), len(functionCodes))

	return result
}

// buildModbusTestRequest 构建Modbus测试请求
func buildModbusTestRequest(funcCode byte, transactionId int) []byte {
	switch funcCode {
	case 0x01, 0x02: // Read Coils/Discrete Inputs
		return []byte{
			byte(transactionId >> 8), byte(transactionId), // 事务ID
			0x00, 0x00, // 协议ID
			0x00, 0x06, // 长度
			0x01,       // 单元ID
			funcCode,   // 功能码
			0x00, 0x00, // 起始地址
			0x00, 0x01, // 数量
		}
	case 0x03, 0x04: // Read Holding/Input Registers
		return []byte{
			byte(transactionId >> 8), byte(transactionId), // 事务ID
			0x00, 0x00, // 协议ID
			0x00, 0x06, // 长度
			0x01,       // 单元ID
			funcCode,   // 功能码
			0x00, 0x00, // 起始地址
			0x00, 0x01, // 数量
		}
	case 0x05: // Write Single Coil
		return []byte{
			byte(transactionId >> 8), byte(transactionId), // 事务ID
			0x00, 0x00, // 协议ID
			0x00, 0x06, // 长度
			0x01,       // 单元ID
			funcCode,   // 功能码
			0x00, 0x00, // 地址
			0x00, 0x00, // 值 (OFF)
		}
	case 0x06: // Write Single Register
		return []byte{
			byte(transactionId >> 8), byte(transactionId), // 事务ID
			0x00, 0x00, // 协议ID
			0x00, 0x06, // 长度
			0x01,       // 单元ID
			funcCode,   // 功能码
			0x00, 0x00, // 地址
			0x00, 0x00, // 值
		}
	case 0x11: // Read Device Identification
		return []byte{
			byte(transactionId >> 8), byte(transactionId), // 事务ID
			0x00, 0x00, // 协议ID
			0x00, 0x06, // 长度
			0x01,       // 单元ID
			funcCode,   // 功能码
			0x01,       // MEI类型
			0x00,       // 读取设备ID代码
			0x00,       // 对象ID
		}
	default:
		// 通用请求
		return []byte{
			byte(transactionId >> 8), byte(transactionId), // 事务ID
			0x00, 0x00, // 协议ID
			0x00, 0x06, // 长度
			0x01,       // 单元ID
			funcCode,   // 功能码
			0x00, 0x00, // 参数1
			0x00, 0x01, // 参数2
		}
	}
}

// getModbusFunctionName 获取Modbus功能码名称
func getModbusFunctionName(code byte) string {
	names := map[byte]string{
		0x01: "Read Coils",
		0x02: "Read Discrete Inputs",
		0x03: "Read Holding Registers",
		0x04: "Read Input Registers",
		0x05: "Write Single Coil",
		0x06: "Write Single Register",
		0x0F: "Write Multiple Coils",
		0x10: "Write Multiple Registers",
		0x11: "Read Device Identification",
		0x14: "Read File Record",
		0x15: "Write File Record",
		0x16: "Mask Write Register",
		0x17: "Read/Write Multiple Registers",
	}

	if name, exists := names[code]; exists {
		return name
	}
	return fmt.Sprintf("Function 0x%02X", code)
}

// executeModbusCoilEnum 执行Modbus线圈枚举
func executeModbusCoilEnum(target Target, ctx *ScriptContext) *ScriptResult {
	result := &ScriptResult{
		Success:  false,
		Findings: make(map[string]interface{}),
	}

	ctx.Logger.Debug("开始枚举Modbus线圈")

	// 连接到目标
	conn, err := net.DialTimeout("tcp", target.String(), ctx.Timeout)
	if err != nil {
		result.Error = fmt.Sprintf("连接失败: %v", err)
		return result
	}
	defer conn.Close()

	// 扫描线圈地址范围
	coilRanges := []struct {
		start uint16
		count uint16
		name  string
	}{
		{0, 16, "0-15"},
		{100, 16, "100-115"},
		{1000, 16, "1000-1015"},
		{10000, 16, "10000-10015"},
	}

	accessibleRanges := make([]string, 0)
	coilValues := make(map[string]interface{})

	for _, coilRange := range coilRanges {
		ctx.Logger.Debug("扫描线圈范围: %s", coilRange.name)

		// 构造读取线圈请求
		request := []byte{
			0x00, 0x01, // 事务ID
			0x00, 0x00, // 协议ID
			0x00, 0x06, // 长度
			0x01,                              // 单元ID
			0x01,                              // 功能码: Read Coils
			byte(coilRange.start >> 8),        // 起始地址高字节
			byte(coilRange.start),             // 起始地址低字节
			byte(coilRange.count >> 8),        // 数量高字节
			byte(coilRange.count),             // 数量低字节
		}

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, err := conn.Write(request)
		if err != nil {
			continue
		}

		response := make([]byte, 256)
		n, err := conn.Read(response)
		if err != nil {
			continue
		}

		if n >= 9 && response[7] == 0x01 {
			// 成功读取
			accessibleRanges = append(accessibleRanges, coilRange.name)
			
			byteCount := response[8]
			if n >= 9+int(byteCount) {
				coilData := response[9 : 9+byteCount]
				coilValues[coilRange.name] = fmt.Sprintf("%X", coilData)
			}
		} else if n >= 9 && response[7] == 0x81 {
			// 异常响应
			exceptionCode := response[8]
			coilValues[coilRange.name] = fmt.Sprintf("Exception: 0x%02X", exceptionCode)
		}

		time.Sleep(100 * time.Millisecond)
	}

	result.Findings["accessible_ranges"] = accessibleRanges
	result.Findings["coil_values"] = coilValues
	result.Findings["total_ranges_tested"] = len(coilRanges)
	result.Findings["accessible_ranges_count"] = len(accessibleRanges)

	result.Success = true
	ctx.Logger.Debug("线圈枚举完成，发现 %d 个可访问范围", len(accessibleRanges))

	return result
}

// executeModbusRegisterRead 执行Modbus寄存器读取
func executeModbusRegisterRead(target Target, ctx *ScriptContext) *ScriptResult {
	result := &ScriptResult{
		Success:  false,
		Findings: make(map[string]interface{}),
	}

	ctx.Logger.Debug("开始读取Modbus寄存器")

	// 连接到目标
	conn, err := net.DialTimeout("tcp", target.String(), ctx.Timeout)
	if err != nil {
		result.Error = fmt.Sprintf("连接失败: %v", err)
		return result
	}
	defer conn.Close()

	// 扫描寄存器地址范围
	registerRanges := []struct {
		start uint16
		count uint16
		name  string
		funcCode byte
		desc  string
	}{
		{0, 10, "holding_0-9", 0x03, "Holding Registers 0-9"},
		{100, 10, "holding_100-109", 0x03, "Holding Registers 100-109"},
		{1000, 10, "holding_1000-1009", 0x03, "Holding Registers 1000-1009"},
		{0, 10, "input_0-9", 0x04, "Input Registers 0-9"},
		{100, 10, "input_100-109", 0x04, "Input Registers 100-109"},
	}

	accessibleRegisters := make(map[string]interface{})
	registerValues := make(map[string]interface{})

	for _, regRange := range registerRanges {
		ctx.Logger.Debug("扫描寄存器范围: %s", regRange.desc)

		// 构造读取寄存器请求
		request := []byte{
			0x00, 0x01, // 事务ID
			0x00, 0x00, // 协议ID
			0x00, 0x06, // 长度
			0x01,                         // 单元ID
			regRange.funcCode,            // 功能码
			byte(regRange.start >> 8),    // 起始地址高字节
			byte(regRange.start),         // 起始地址低字节
			byte(regRange.count >> 8),    // 数量高字节
			byte(regRange.count),         // 数量低字节
		}

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, err := conn.Write(request)
		if err != nil {
			continue
		}

		response := make([]byte, 256)
		n, err := conn.Read(response)
		if err != nil {
			continue
		}

		if n >= 9 && response[7] == regRange.funcCode {
			// 成功读取
			accessibleRegisters[regRange.name] = true
			
			byteCount := response[8]
			if n >= 9+int(byteCount) {
				registerData := response[9 : 9+byteCount]
				
				// 解析寄存器值
				values := make([]uint16, 0)
				for i := 0; i < len(registerData); i += 2 {
					if i+1 < len(registerData) {
						value := binary.BigEndian.Uint16(registerData[i : i+2])
						values = append(values, value)
					}
				}
				registerValues[regRange.name] = values
			}
		} else if n >= 9 && response[7] == (regRange.funcCode|0x80) {
			// 异常响应
			accessibleRegisters[regRange.name] = false
			if n > 8 {
				exceptionCode := response[8]
				registerValues[regRange.name] = fmt.Sprintf("Exception: 0x%02X", exceptionCode)
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	result.Findings["accessible_registers"] = accessibleRegisters
	result.Findings["register_values"] = registerValues
	result.Findings["total_ranges_tested"] = len(registerRanges)

	// 统计可访问的寄存器数量
	accessibleCount := 0
	for _, accessible := range accessibleRegisters {
		if accessible.(bool) {
			accessibleCount++
		}
	}
	result.Findings["accessible_ranges_count"] = accessibleCount

	result.Success = true
	ctx.Logger.Debug("寄存器读取完成，发现 %d 个可访问范围", accessibleCount)

	return result
}

// executeModbusAuthBypass 执行Modbus认证绕过检测
func executeModbusAuthBypass(target Target, ctx *ScriptContext) *ScriptResult {
	result := &ScriptResult{
		Success:         false,
		Findings:        make(map[string]interface{}),
		Vulnerabilities: make([]Vulnerability, 0),
	}

	ctx.Logger.Debug("开始检测Modbus认证绕过漏洞")

	// 连接到目标
	conn, err := net.DialTimeout("tcp", target.String(), ctx.Timeout)
	if err != nil {
		result.Error = fmt.Sprintf("连接失败: %v", err)
		return result
	}
	defer conn.Close()

	// 测试1: 无认证直接访问
	ctx.Logger.Debug("测试无认证访问")
	
	// 尝试读取设备信息
	deviceIdRequest := []byte{
		0x00, 0x01, // 事务ID
		0x00, 0x00, // 协议ID
		0x00, 0x06, // 长度
		0x01,       // 单元ID
		0x11,       // 功能码: Read Device Identification
		0x01,       // MEI类型
		0x00,       // 读取设备ID代码
		0x00,       // 对象ID
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Write(deviceIdRequest)
	if err == nil {
		response := make([]byte, 256)
		n, err := conn.Read(response)
		if err == nil && n >= 8 {
			if response[7] == 0x11 {
				// 成功获取设备信息，说明无认证保护
				result.Findings["no_authentication"] = true
				
				vuln := Vulnerability{
					CVE:         "CWE-306",
					Title:       "Missing Authentication for Critical Function",
					Description: "Modbus设备未实施认证机制，允许未授权访问",
					Severity:    SeverityHigh,
					CVSS:        7.5,
					ExploitAvailable: true,
					References: []string{
						"https://cwe.mitre.org/data/definitions/306.html",
					},
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
			}
		}
	}

	// 测试2: 尝试写操作
	ctx.Logger.Debug("测试未授权写操作")
	
	// 尝试写入单个线圈 (地址0，值OFF)
	writeCoilRequest := []byte{
		0x00, 0x02, // 事务ID
		0x00, 0x00, // 协议ID
		0x00, 0x06, // 长度
		0x01,       // 单元ID
		0x05,       // 功能码: Write Single Coil
		0x00, 0x00, // 地址
		0x00, 0x00, // 值 (OFF)
	}

	_, err = conn.Write(writeCoilRequest)
	if err == nil {
		response := make([]byte, 256)
		n, err := conn.Read(response)
		if err == nil && n >= 8 {
			if response[7] == 0x05 {
				// 成功写入，说明允许未授权写操作
				result.Findings["unauthorized_write"] = true
				
				vuln := Vulnerability{
					CVE:         "CWE-862",
					Title:       "Missing Authorization",
					Description: "Modbus设备允许未授权的写操作，可能导致设备状态被恶意修改",
					Severity:    SeverityCritical,
					CVSS:        9.1,
					ExploitAvailable: true,
					References: []string{
						"https://cwe.mitre.org/data/definitions/862.html",
					},
				}
				result.Vulnerabilities = append(result.Vulnerabilities, vuln)
			} else if response[7] == 0x85 {
				// 写操作被拒绝
				result.Findings["unauthorized_write"] = false
				if n > 8 {
					exceptionCode := response[8]
					result.Findings["write_exception"] = fmt.Sprintf("0x%02X", exceptionCode)
				}
			}
		}
	}

	// 测试3: 单元ID枚举
	ctx.Logger.Debug("测试单元ID枚举")
	
	validUnitIds := make([]int, 0)
	for unitId := 1; unitId <= 10; unitId++ {
		testRequest := []byte{
			0x00, byte(unitId + 10), // 事务ID
			0x00, 0x00, // 协议ID
			0x00, 0x06, // 长度
			byte(unitId), // 单元ID
			0x03,         // 功能码: Read Holding Registers
			0x00, 0x00,   // 起始地址
			0x00, 0x01,   // 数量
		}

		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, err = conn.Write(testRequest)
		if err == nil {
			response := make([]byte, 256)
			n, err := conn.Read(response)
			if err == nil && n >= 8 {
				if response[7] == 0x03 || response[7] == 0x83 {
					// 有响应，说明单元ID有效
					validUnitIds = append(validUnitIds, unitId)
				}
			}
		}
		
		time.Sleep(100 * time.Millisecond)
	}

	result.Findings["valid_unit_ids"] = validUnitIds
	result.Findings["unit_id_count"] = len(validUnitIds)

	if len(validUnitIds) > 1 {
		result.Findings["multiple_units"] = true
	}

	result.Success = true
	ctx.Logger.Debug("认证绕过检测完成，发现 %d 个漏洞", len(result.Vulnerabilities))

	return result
}

// executeModbusDosTest 执行Modbus拒绝服务测试
func executeModbusDosTest(target Target, ctx *ScriptContext) *ScriptResult {
	result := &ScriptResult{
		Success:         false,
		Findings:        make(map[string]interface{}),
		Vulnerabilities: make([]Vulnerability, 0),
	}

	ctx.Logger.Debug("开始检测Modbus拒绝服务漏洞")

	// 测试1: 大量并发连接
	ctx.Logger.Debug("测试大量并发连接")
	
	maxConnections := 0
	for i := 0; i < 20; i++ {
		conn, err := net.DialTimeout("tcp", target.String(), 2*time.Second)
		if err != nil {
			break
		}
		maxConnections++
		
		// 保持连接短暂时间
		go func(c net.Conn) {
			time.Sleep(1 * time.Second)
			c.Close()
		}(conn)
	}

	result.Findings["max_concurrent_connections"] = maxConnections
	
	if maxConnections >= 15 {
		result.Findings["connection_limit_vulnerable"] = true
		
		vuln := Vulnerability{
			CVE:         "CWE-770",
			Title:       "Allocation of Resources Without Limits or Throttling",
			Description: "Modbus设备允许过多并发连接，可能导致资源耗尽",
			Severity:    SeverityMedium,
			CVSS:        5.3,
			ExploitAvailable: true,
		}
		result.Vulnerabilities = append(result.Vulnerabilities, vuln)
	}

	// 等待连接关闭
	time.Sleep(2 * time.Second)

	// 测试2: 恶意请求
	ctx.Logger.Debug("测试恶意请求")
	
	conn, err := net.DialTimeout("tcp", target.String(), ctx.Timeout)
	if err != nil {
		result.Error = fmt.Sprintf("连接失败: %v", err)
		return result
	}
	defer conn.Close()

	// 发送超大请求
	maliciousRequest := make([]byte, 1024)
	maliciousRequest[0] = 0x00 // 事务ID高字节
	maliciousRequest[1] = 0x01 // 事务ID低字节
	maliciousRequest[2] = 0x00 // 协议ID高字节
	maliciousRequest[3] = 0x00 // 协议ID低字节
	maliciousRequest[4] = 0x03 // 长度高字节 (错误的长度)
	maliciousRequest[5] = 0xFF // 长度低字节 (错误的长度)
	maliciousRequest[6] = 0x01 // 单元ID
	maliciousRequest[7] = 0x03 // 功能码

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Write(maliciousRequest)
	if err == nil {
		response := make([]byte, 256)
		n, err := conn.Read(response)
		if err != nil {
			// 连接被关闭，可能触发了保护机制
			result.Findings["malicious_request_protection"] = true
		} else if n > 0 {
			// 仍有响应，可能存在解析漏洞
			result.Findings["malicious_request_response"] = true
			
			vuln := Vulnerability{
				CVE:         "CWE-20",
				Title:       "Improper Input Validation",
				Description: "Modbus设备对恶意格式的请求处理不当",
				Severity:    SeverityMedium,
				CVSS:        4.3,
				ExploitAvailable: false,
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
		}
	}

	// 测试3: 功能码洪水攻击
	ctx.Logger.Debug("测试功能码洪水攻击")
	
	floodCount := 0
	startTime := time.Now()
	
	for time.Since(startTime) < 3*time.Second {
		floodRequest := []byte{
			0x00, byte(floodCount), // 事务ID
			0x00, 0x00, // 协议ID
			0x00, 0x06, // 长度
			0x01,       // 单元ID
			0x03,       // 功能码
			0x00, 0x00, // 地址
			0x00, 0x01, // 数量
		}

		conn.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
		_, err = conn.Write(floodRequest)
		if err != nil {
			break
		}
		
		floodCount++
	}

	result.Findings["flood_requests_sent"] = floodCount
	result.Findings["flood_rate_per_second"] = float64(floodCount) / 3.0

	if floodCount > 100 {
		result.Findings["flood_vulnerable"] = true
		
		vuln := Vulnerability{
			CVE:         "CWE-400",
			Title:       "Uncontrolled Resource Consumption",
			Description: "Modbus设备未限制请求频率，可能导致拒绝服务",
			Severity:    SeverityMedium,
			CVSS:        5.3,
			ExploitAvailable: true,
		}
		result.Vulnerabilities = append(result.Vulnerabilities, vuln)
	}

	result.Success = true
	ctx.Logger.Debug("拒绝服务测试完成，发现 %d 个漏洞", len(result.Vulnerabilities))

	return result
}