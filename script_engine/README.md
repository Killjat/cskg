# 🚀 Script Engine - 深度协议探测脚本系统

## 📋 项目概述

Script Engine是一个类似Nmap NSE的深度协议探测脚本系统，基于我们的network_probe项目构建。它能够对已识别的协议进行深度安全检测、漏洞发现和信息收集。

## 🎯 设计目标

### 核心理念
- **协议专精** - 基于49种协议的深度理解
- **模块化设计** - 每个协议独立的脚本包
- **智能执行** - 根据探测结果自动选择合适的脚本
- **安全导向** - 专注于安全漏洞和风险发现

### 技术优势
- **原生Go性能** - 高并发、低延迟
- **现代化架构** - 微服务、容器化部署
- **智能调度** - AI辅助的脚本选择和执行
- **结果关联** - 多维度的安全分析

## 🏗️ 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   协议探测      │───▶│   脚本引擎      │───▶│   结果分析      │
│ (network_probe) │    │ (script_engine) │    │ (vulnerability) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  协议识别结果   │    │  深度探测脚本   │    │  安全评估报告   │
│  - 协议类型     │    │  - 信息收集     │    │  - 漏洞列表     │
│  - 置信度       │    │  - 漏洞检测     │    │  - 风险评级     │
│  - 服务信息     │    │  - 认证测试     │    │  - 修复建议     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📦 脚本分类

### 🏭 工控协议脚本
- **Modbus** - 工业控制系统 (100%识别成功率)
- **DNP3** - 电力系统SCADA
- **BACnet** - 楼宇自动化
- **OPC UA** - 工业4.0标准
- **S7** - 西门子PLC

### 🗄️ 数据库协议脚本
- **MySQL** - 关系数据库 (80%识别成功率)
- **Redis** - 内存数据库 (100%识别成功率)
- **MongoDB** - NoSQL数据库 (70%识别成功率)
- **PostgreSQL** - 高级关系数据库
- **Oracle** - 企业级数据库

### 🌐 IoT协议脚本
- **MQTT** - 消息队列 (100%识别成功率)
- **CoAP** - 受限设备协议
- **LoRaWAN** - 低功耗广域网
- **AMQP** - 高级消息队列

### 🏢 企业协议脚本
- **Kerberos** - 网络认证 (100%识别成功率)
- **LDAP** - 目录服务
- **RADIUS** - 远程认证 (70%识别成功率)
- **NTP** - 网络时间协议

## 🎨 脚本类型

### 1. 信息收集脚本 (Discovery)
- 服务版本识别
- 配置信息获取
- 系统信息收集
- 网络拓扑发现

### 2. 漏洞检测脚本 (Vulnerability)
- 已知CVE检测
- 配置错误发现
- 弱密码检测
- 权限提升漏洞

### 3. 认证测试脚本 (Authentication)
- 默认凭据测试
- 暴力破解攻击
- 认证绕过检测
- 会话劫持测试

### 4. 利用验证脚本 (Exploitation)
- 漏洞利用验证
- 代码执行测试
- 数据泄露检测
- 拒绝服务测试

## 🚀 快速开始

### 环境要求
- Go 1.21+
- 依赖network_probe项目
- Linux/macOS/Windows

### 安装部署
```bash
# 克隆项目
git clone <repository>
cd script_engine

# 编译项目
go build -o script_engine .

# 运行示例
./script_engine -target 192.168.1.100:502 -protocol modbus -scripts info,vuln
```

### 基本用法
```bash
# 对Modbus设备进行深度探测
./script_engine -target 192.168.1.100:502 -protocol modbus

# 对Redis服务器进行漏洞扫描
./script_engine -target 192.168.1.100:6379 -protocol redis -category vuln

# 批量扫描多个目标
./script_engine -targets targets.txt -auto-detect
```

## 📊 输出格式

### JSON格式
```json
{
  "target": "192.168.1.100:502",
  "protocol": "modbus",
  "timestamp": "2026-01-19T10:30:00Z",
  "scripts_executed": [
    {
      "name": "modbus-device-info",
      "category": "discovery",
      "success": true,
      "findings": {
        "device_id": "1",
        "vendor": "Schneider Electric",
        "model": "M340",
        "firmware": "v2.70"
      }
    }
  ],
  "vulnerabilities": [
    {
      "cve": "CVE-2020-7491",
      "severity": "high",
      "description": "Authentication bypass in Modbus TCP",
      "exploit_available": true
    }
  ]
}
```

### 文本格式
```
🎯 目标: 192.168.1.100:502 (Modbus TCP)
📊 执行脚本: 5个
✅ 成功: 4个
❌ 失败: 1个

📋 发现信息:
  🏷️  设备ID: 1
  🏭 厂商: Schneider Electric
  📦 型号: M340
  🔧 固件: v2.70

🚨 安全漏洞:
  ⚠️  CVE-2020-7491 (高危)
      认证绕过漏洞
      影响: 未授权访问设备
      修复: 升级固件到v2.80+
```

## 🔧 开发指南

### 脚本开发
```go
// 示例: Modbus设备信息收集脚本
type ModbusInfoScript struct {
    BaseScript
}

func (s *ModbusInfoScript) Execute(target Target, ctx *ScriptContext) *ScriptResult {
    // 实现具体的探测逻辑
    result := &ScriptResult{
        ScriptName: "modbus-device-info",
        Success:    true,
        Findings:   make(map[string]interface{}),
    }
    
    // 发送Modbus查询请求
    response, err := s.sendModbusQuery(target, 0x11) // Read Device Identification
    if err != nil {
        result.Success = false
        result.Error = err.Error()
        return result
    }
    
    // 解析响应
    deviceInfo := s.parseDeviceInfo(response)
    result.Findings = deviceInfo
    
    return result
}
```

### 脚本注册
```go
func init() {
    RegisterScript(&ModbusInfoScript{
        BaseScript: BaseScript{
            Name:        "modbus-device-info",
            Protocol:    "modbus",
            Category:    "discovery",
            Description: "Collect Modbus device information",
            Author:      "Script Engine Team",
            Version:     "1.0",
        },
    })
}
```

## 🎯 路线图

### Phase 1: 核心框架 (2-3周)
- [x] 项目架构设计
- [ ] 脚本引擎核心
- [ ] 脚本加载器
- [ ] 结果处理器
- [ ] 基础CLI界面

### Phase 2: 高价值脚本 (4-6周)
- [ ] Modbus脚本包 (优先级最高)
- [ ] Redis脚本包
- [ ] MQTT脚本包
- [ ] MySQL脚本包
- [ ] Kerberos脚本包

### Phase 3: 扩展功能 (持续)
- [ ] Web管理界面
- [ ] 漏洞数据库集成
- [ ] AI辅助脚本推荐
- [ ] 分布式扫描支持
- [ ] 报告生成系统

## 🤝 贡献指南

### 脚本贡献
1. Fork项目
2. 创建脚本分支
3. 实现脚本逻辑
4. 添加测试用例
5. 提交Pull Request

### 代码规范
- 遵循Go代码规范
- 添加详细注释
- 包含错误处理
- 编写单元测试

## 📄 许可证

MIT License - 详见LICENSE文件

## 🙏 致谢

- 基于network_probe项目的协议识别能力
- 参考Nmap NSE的设计理念
- 感谢开源社区的贡献

---

**让我们一起构建下一代网络安全扫描系统！** 🚀