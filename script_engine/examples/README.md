# 🌰 Script Engine 使用示例

## 基本使用

### 1. 列出所有可用脚本
```bash
./script_engine -list-scripts
```

### 2. 对Modbus设备进行深度探测
```bash
# 执行所有Modbus脚本
./script_engine -target 192.168.1.100:502 -protocol modbus

# 只执行信息收集脚本
./script_engine -target 192.168.1.100:502 -protocol modbus -category discovery

# 执行特定脚本
./script_engine -target 192.168.1.100:502 -protocol modbus -scripts modbus-device-info,modbus-function-scan
```

### 3. 对Redis服务器进行安全检测
```bash
# 完整安全检测
./script_engine -target 192.168.1.100:6379 -protocol redis

# 只检测漏洞
./script_engine -target 192.168.1.100:6379 -protocol redis -category vulnerability

# 暴力破解测试
./script_engine -target 192.168.1.100:6379 -protocol redis -scripts redis-brute-auth
```

### 4. 批量扫描
```bash
# 从文件读取目标列表
./script_engine -targets targets.txt -auto-detect

# 高并发扫描
./script_engine -targets targets.txt -concurrent 20 -timeout 10s
```

### 5. 输出格式控制
```bash
# JSON格式输出
./script_engine -target 192.168.1.100:502 -protocol modbus -output json

# 保存结果到文件
./script_engine -target 192.168.1.100:502 -protocol modbus -output json -output-file result.json

# 详细输出
./script_engine -target 192.168.1.100:502 -protocol modbus -verbose
```

## 高级用法

### 1. 工控设备安全评估
```bash
# Modbus设备完整评估
./script_engine -target 192.168.1.100:502 -protocol modbus -verbose -output json -output-file modbus_assessment.json

# 多个工控协议测试
for protocol in modbus dnp3 bacnet opcua s7; do
    ./script_engine -target 192.168.1.100:502 -protocol $protocol -output json -output-file ${protocol}_result.json
done
```

### 2. 数据库安全扫描
```bash
# Redis安全扫描
./script_engine -target 192.168.1.100:6379 -protocol redis -category vulnerability -verbose

# MySQL安全检测
./script_engine -target 192.168.1.100:3306 -protocol mysql -scripts mysql-info,mysql-auth-bypass

# 多数据库扫描
databases=("mysql:3306" "redis:6379" "mongodb:27017")
for db in "${databases[@]}"; do
    IFS=':' read -r protocol port <<< "$db"
    ./script_engine -target "192.168.1.100:$port" -protocol "$protocol" -verbose
done
```

### 3. IoT设备发现
```bash
# MQTT代理检测
./script_engine -target 192.168.1.100:1883 -protocol mqtt -verbose

# 批量IoT协议检测
iot_ports=("1883:mqtt" "5683:coap" "5672:amqp")
for item in "${iot_ports[@]}"; do
    IFS=':' read -r port protocol <<< "$item"
    ./script_engine -target "192.168.1.100:$port" -protocol "$protocol"
done
```

### 4. 企业网络评估
```bash
# Kerberos域控检测
./script_engine -target 192.168.1.10:88 -protocol kerberos -verbose

# LDAP目录服务检测
./script_engine -target 192.168.1.10:389 -protocol ldap

# 完整企业协议扫描
enterprise_services=("88:kerberos" "389:ldap" "1812:radius" "123:ntp")
for service in "${enterprise_services[@]}"; do
    IFS=':' read -r port protocol <<< "$service"
    ./script_engine -target "192.168.1.10:$port" -protocol "$protocol" -verbose
done
```

## 目标文件格式

### targets.txt 示例
```
192.168.1.100:502
192.168.1.101:6379
192.168.1.102:1883
192.168.1.103:3306
10.0.0.50:88
```

## 输出示例

### 文本格式输出
```
🎯 目标: 192.168.1.100:502 (modbus)
📊 执行脚本: 6个
✅ 成功: 5个
❌ 失败: 1个

📋 发现信息:
  🏷️  设备ID: 1
  🏭 厂商: Schneider Electric
  📦 型号: M340
  🔧 固件: v2.70

🚨 安全漏洞:
  ⚠️  CWE-306 (高危)
      认证绕过漏洞
      影响: 未授权访问设备
      修复: 启用认证机制
```

### JSON格式输出
```json
{
  "target": "192.168.1.100:502",
  "protocol": "modbus",
  "timestamp": "2026-01-19T10:30:00Z",
  "findings": {
    "device_id": "1",
    "vendor": "Schneider Electric",
    "model": "M340",
    "firmware": "v2.70"
  },
  "vulnerabilities": [
    {
      "cve": "CWE-306",
      "severity": "high",
      "description": "Missing Authentication for Critical Function",
      "exploit_available": true
    }
  ],
  "script_results": [
    {
      "script_name": "modbus-device-info",
      "category": "discovery",
      "success": true,
      "duration": "150ms"
    }
  ]
}
```

## 脚本开发示例

### 自定义脚本模板
```go
// 自定义Modbus脚本示例
func executeCustomModbusScript(target Target, ctx *ScriptContext) *ScriptResult {
    result := &ScriptResult{
        Success:  false,
        Findings: make(map[string]interface{}),
    }

    // 连接到目标
    conn, err := net.DialTimeout("tcp", target.String(), ctx.Timeout)
    if err != nil {
        result.Error = fmt.Sprintf("连接失败: %v", err)
        return result
    }
    defer conn.Close()

    // 实现具体的探测逻辑
    // ...

    result.Success = true
    return result
}
```

## 性能优化建议

### 1. 并发控制
```bash
# 根据网络环境调整并发数
./script_engine -targets large_list.txt -concurrent 50  # 高速网络
./script_engine -targets large_list.txt -concurrent 10  # 普通网络
./script_engine -targets large_list.txt -concurrent 5   # 慢速网络
```

### 2. 超时设置
```bash
# 根据目标响应速度调整超时
./script_engine -target slow_device:502 -timeout 30s    # 慢速设备
./script_engine -target fast_device:502 -timeout 5s     # 快速设备
```

### 3. 脚本选择
```bash
# 快速扫描 - 只执行信息收集
./script_engine -target 192.168.1.100:502 -category discovery

# 深度扫描 - 执行所有脚本
./script_engine -target 192.168.1.100:502 -scripts all

# 安全扫描 - 只执行漏洞检测
./script_engine -target 192.168.1.100:502 -category vulnerability
```

## 故障排除

### 常见问题

1. **连接超时**
   ```bash
   # 增加超时时间
   ./script_engine -target 192.168.1.100:502 -timeout 30s
   ```

2. **权限不足**
   ```bash
   # 某些脚本可能需要特殊权限
   sudo ./script_engine -target 192.168.1.100:502 -protocol modbus
   ```

3. **防火墙阻断**
   ```bash
   # 使用详细输出查看具体错误
   ./script_engine -target 192.168.1.100:502 -verbose
   ```

4. **协议检测失败**
   ```bash
   # 手动指定协议
   ./script_engine -target 192.168.1.100:502 -protocol modbus
   ```

### 调试模式
```bash
# 启用详细日志
./script_engine -target 192.168.1.100:502 -verbose

# 保存详细日志到文件
./script_engine -target 192.168.1.100:502 -verbose 2>&1 | tee debug.log
```