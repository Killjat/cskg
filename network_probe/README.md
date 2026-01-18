# 网络探测引擎 (Network Probe Engine)

基于Go语言开发的网络服务探测工具，支持TCP/UDP协议探测，内置多种协议解析器，兼容Nmap探测载荷格式。

## 🚀 功能特性

- **多协议支持**: TCP/UDP协议探测
- **内置探测库**: 包含HTTP、SSH、FTP、SMTP、MySQL、Redis等常见服务探测
- **全面探测**: 支持发送所有探测包，发现非标准端口服务
- **协议解析**: 智能解析响应数据，提取服务信息和版本
- **多种探测模式**: port/all/smart三种模式适应不同场景
- **高性能**: 支持并发探测，微秒级响应时间
- **灵活配置**: 可配置超时时间、并发数、重试次数等
- **多种输出**: 支持文本和JSON格式输出
- **Nmap兼容**: 支持Nmap探测载荷格式

## 🎯 非标准端口服务探测

在实际网络环境中，服务经常运行在非标准端口上，例如：
- **22端口运行HTTP服务** - SSH端口上的Web服务
- **80端口运行SSH服务** - HTTP端口上的SSH服务
- **8080端口运行数据库** - Web端口上的数据库服务
- **443端口运行FTP服务** - HTTPS端口上的FTP服务

传统的基于端口的服务识别会错过这些情况。本工具通过发送所有协议的探测包，能够准确识别任意端口上运行的真实服务。

## 📦 项目结构

```
network_probe/
├── main.go        # 主程序和CLI接口
├── types.go       # 类型定义
├── probes.go      # 探测载荷库
├── engine.go      # 探测引擎核心
├── parsers.go     # 协议解析器
├── go.mod         # Go模块文件
└── README.md      # 项目文档
```

## 🛠 安装使用

### 编译运行

```bash
cd network_probe
go mod tidy
go build -o network_probe .
```

### 基本用法

```bash
# 探测单个目标
./network_probe -target baidu.com:80

# 使用主机和端口参数
./network_probe -host 127.0.0.1 -port 22

# 详细输出
./network_probe -target 192.168.1.1:80 -verbose

# JSON格式输出
./network_probe -target 8.8.8.8:53 -output json

# 列出所有可用探测
./network_probe -list-probes

# 显示统计信息
./network_probe -target example.com:443 -stats
```

### 参数说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-target` | 目标地址 (host:port) | - |
| `-host` | 目标主机 | - |
| `-port` | 目标端口 | 0 |
| `-timeout` | 探测超时时间 | 10s |
| `-concurrent` | 并发数 | 10 |
| `-output` | 输出格式 (text/json) | text |
| `-verbose` | 详细输出 | false |
| `-list-probes` | 列出所有探测 | false |
| `-stats` | 显示统计信息 | false |
| `-probe-mode` | 探测模式 (port/all/smart) | all |

### 探测模式说明

| 模式 | 说明 | 适用场景 |
|------|------|----------|
| `port` | 仅使用端口相关的探测 | 快速识别标准端口服务 |
| `all` | 使用所有探测包试探 | 全面探测，发现非标准端口服务 |
| `smart` | 智能模式，优先常见探测 | 平衡速度和覆盖面 |

## 🔍 支持的协议

### TCP协议
- **HTTP**: GET请求、OPTIONS请求
- **SSH**: 版本交换
- **FTP**: 用户认证
- **SMTP**: EHLO命令
- **MySQL**: 握手包
- **Redis**: PING命令
- **PostgreSQL**: 启动消息
- **Telnet**: 选项协商
- **POP3**: 能力查询
- **IMAP**: 能力查询

### UDP协议
- **DNS**: 状态请求
- **SNMP**: GetRequest

## 📊 输出示例

### 文本格式输出

```
🔍 网络探测引擎
==============================
🎯 开始探测 1 个目标...

🎯 目标: baidu.com:80
------------------------------------------------------------
✅ 成功探测: 2/3

1. ✅ NULL (tcp) - 耗时: 45.2ms

2. ✅ GetRequest (http) - 耗时: 52.1ms
   📄 Banner: "HTTP/1.1 200 OK\r\nServer: nginx/1.20.1\r\n..."
   🏷️  产品: nginx v1.20.1 (置信度: 90%)

📊 探测统计:
----------------------------------------
总探测数: 3
成功探测: 2
失败探测: 1
成功率: 66.7%
平均耗时: 48.6ms
总耗时: 145.9ms

协议分布:
  tcp: 1
  http: 1
```

### JSON格式输出

```json
{
  "results": {
    "baidu.com:80": [
      {
        "target": "baidu.com:80",
        "port": 80,
        "probe_name": "GetRequest",
        "protocol": "http",
        "success": true,
        "response": "...",
        "response_hex": "485454502f312e3120323030204f4b...",
        "banner": "HTTP/1.1 200 OK\\r\\nServer: nginx/1.20.1...",
        "parsed_info": {
          "protocol": "http",
          "service": "http",
          "product": "nginx",
          "version": "1.20.1",
          "confidence": 90,
          "fields": {
            "server": "nginx/1.20.1",
            "status_line": "HTTP/1.1 200 OK"
          }
        },
        "duration": 52100000,
        "timestamp": "2026-01-19T10:30:45Z"
      }
    ]
  },
  "timestamp": "2026-01-19T10:30:45Z",
  "summary": {
    "total_targets": 1,
    "total_probes": 3,
    "success_probes": 2,
    "success_rate": 66.7
  }
}
```

## 🔧 技术架构

### 核心组件

1. **ProbeEngine**: 探测引擎核心，负责管理探测流程
2. **ProbeLoader**: 探测载荷加载器，支持内置和Nmap格式
3. **ProtocolParser**: 协议解析器接口，支持多种协议解析
4. **ProbeResult**: 探测结果结构，包含完整的响应信息

### 工作流程

1. **目标解析**: 解析用户输入的目标地址
2. **探测选择**: 根据端口选择合适的探测载荷
3. **并发执行**: 使用goroutine并发执行探测
4. **协议解析**: 使用对应的解析器解析响应数据
5. **结果输出**: 格式化输出探测结果

### 性能特性

- **并发探测**: 支持可配置的并发数
- **超时控制**: 连接超时和读取超时分别控制
- **资源管理**: 自动管理连接资源，防止泄漏
- **错误处理**: 完善的错误处理和重试机制

## 🎯 使用场景

- **网络资产发现**: 快速识别网络中的活跃服务
- **服务指纹识别**: 识别服务类型和版本信息
- **安全评估**: 网络安全评估和渗透测试
- **运维监控**: 服务可用性监控和健康检查
- **协议分析**: 网络协议研究和分析

## 🔮 扩展开发

### 添加新的探测

```go
// 在probes.go中添加新探测
{
    Name:        "CustomProbe",
    Type:        ProbeTypeTCP,
    Payload:     []byte("CUSTOM COMMAND\r\n"),
    Ports:       []int{9999},
    Protocol:    "custom",
    Description: "Custom protocol probe",
    Timeout:     5,
    Rarity:      5,
}
```

### 添加新的解析器

```go
// 实现ProtocolParser接口
type CustomParser struct{}

func (p *CustomParser) Parse(data []byte) (*ParsedInfo, error) {
    // 解析逻辑
    return &ParsedInfo{
        Protocol: "custom",
        Service:  "custom-service",
        // ...
    }, nil
}

func (p *CustomParser) GetProtocol() string { return "custom" }
func (p *CustomParser) GetConfidence(data []byte) int { return 80 }
```

## 📝 开发计划

- [ ] 支持IPv6地址
- [ ] 支持CIDR网段扫描
- [ ] 支持从文件读取目标列表
- [ ] 支持自定义探测载荷文件
- [ ] 支持结果导出到数据库
- [ ] 支持Web界面管理
- [ ] 支持分布式探测

## 🤝 贡献指南

欢迎提交Issue和Pull Request来改进项目！

## 📄 许可证

MIT License