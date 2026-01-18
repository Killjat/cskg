# 台湾省IP段扫描与InfluxDB存储系统架构设计

## 1. 整体架构

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  IP段获取模块   │────▶│  SYN扫描模块    │────▶│ IP位置查询模块  │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                │                         │
                                └─────────────────────────┘
                                              │
                                              ▼
                                     ┌─────────────────┐
                                     │ InfluxDB存储模块│
                                     └─────────────────┘
                                              │
                                              ▼
                                     ┌─────────────────┐
                                     │  数据可视化     │
                                     └─────────────────┘
```

## 2. 数据流程

1. **IP段获取**：
   - 从可靠数据源获取台湾省的IP段列表
   - 支持定期更新IP段数据
   - 输出：台湾省IP段列表（CIDR格式）

2. **SYN扫描**：
   - 对每个IP段进行SYN扫描，检测活跃IP
   - 支持并发扫描，提高效率
   - 输出：活跃IP列表

3. **IP位置查询**：
   - 对活跃IP进行地理位置查询
   - 获取国家、地区、城市、ISP等信息
   - 输出：包含位置信息的IP数据

4. **InfluxDB存储**：
   - 将IP和位置信息存储到InfluxDB
   - 设计合理的数据模型和保留策略
   - 支持批量写入，提高性能

5. **数据可视化**：
   - 通过Grafana等工具可视化InfluxDB中的数据
   - 支持按时间、地区、ISP等维度分析

## 3. 技术选型

| 功能模块 | 技术选型 | 理由 |
|----------|----------|------|
| IP段获取 | Go + HTTP API | 高效、并发支持好，适合网络请求 |
| SYN扫描 | Go + raw socket | 底层控制，高性能，适合大规模扫描 |
| IP位置查询 | MaxMind GeoIP2 | 准确、高效，支持离线查询 |
| InfluxDB存储 | InfluxDB 2.x | 时序数据存储优化，自动生命周期管理 |
| 主程序框架 | Go | 高性能、并发支持好，适合网络和系统编程 |

## 4. 数据模型设计

### 4.1 InfluxDB Measurement设计

```
measurement: taiwan_ip_scan

# 标签（带索引，用于过滤和分组）
tags:
  - ip: 192.168.1.1          # IP地址
  - country: Taiwan           # 国家/地区
  - region: Taipei            # 省/州
  - city: Taipei              # 城市
  - isp: Chunghwa Telecom     # ISP
  - asn: 3462                 # ASN编号
  - ip_segment: 1.2.3.0/24    # 所属IP段

# 字段（不带索引，存储实际值）
fields:
  - is_alive: 1               # 是否活跃（1:活跃，0:不活跃）
  - scan_time: 123.45         # 扫描耗时（毫秒）
  - response_time: 15.5       # 响应时间（毫秒）
  - open_ports: [80, 443]     # 开放端口列表

# 时间戳
timestamp: 2026-01-17T16:00:00Z  # 扫描时间
```

### 4.2 保留策略设计

| 数据类型 | 保留时间 | 采样精度 | 存储桶名称 |
|----------|----------|----------|------------|
| 原始扫描数据 | 30天 | 1秒 | taiwan_ip_scan_raw |
| 每小时聚合数据 | 90天 | 1小时 | taiwan_ip_scan_hourly |
| 每天聚合数据 | 1年 | 1天 | taiwan_ip_scan_daily |

## 5. 模块详细设计

### 5.1 IP段获取模块

**功能**：
- 从可靠数据源获取台湾省IP段
- 支持定期更新
- 支持本地缓存

**实现方式**：
1. 使用第三方IP地址库API（如ip2location、maxmind等）
2. 解析公开的IP段分配信息
3. 定期更新本地IP段数据库

**数据结构**：
```go
type IPSegment struct {
	CIDR      string    `json:"cidr"`      // IP段（CIDR格式）
	Country   string    `json:"country"`   // 国家/地区
	Region    string    `json:"region"`    // 省/州
	City      string    `json:"city"`      // 城市
	ISP       string    `json:"isp"`       // ISP
	ASN       uint      `json:"asn"`       // ASN编号
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
}
```

### 5.2 SYN扫描模块

**功能**：
- 对IP段进行SYN扫描
- 检测活跃IP
- 支持并发扫描
- 支持端口范围扫描

**实现方式**：
1. 使用Go的raw socket实现SYN扫描
2. 采用并发设计，提高扫描效率
3. 实现超时处理和错误恢复

**数据结构**：
```go
type ScanResult struct {
	IP            string    `json:"ip"`            // IP地址
	IsAlive       bool      `json:"is_alive"`       // 是否活跃
	ScanTime      float64   `json:"scan_time"`      // 扫描耗时（毫秒）
	ResponseTime  float64   `json:"response_time"`  // 响应时间（毫秒）
	OpenPorts     []int     `json:"open_ports"`     // 开放端口列表
	ScanTimestamp time.Time `json:"scan_timestamp"` // 扫描时间
}
```

### 5.3 IP位置查询模块

**功能**：
- 查询IP的地理位置信息
- 支持离线查询
- 支持批量查询

**实现方式**：
1. 使用MaxMind GeoIP2数据库
2. 实现内存缓存，提高查询效率
3. 支持定期更新数据库

**数据结构**：
```go
type IPLocation struct {
	IP        string  `json:"ip"`        // IP地址
	Country   string  `json:"country"`   // 国家/地区
	Region    string  `json:"region"`    // 省/州
	City      string  `json:"city"`      // 城市
	ISP       string  `json:"isp"`       // ISP
	ASN       uint    `json:"asn"`       // ASN编号
	Latitude  float64 `json:"latitude"`  // 纬度
	Longitude float64 `json:"longitude"` // 经度
}
```

### 5.4 InfluxDB存储模块

**功能**：
- 将扫描结果和位置信息存储到InfluxDB
- 支持批量写入
- 支持数据聚合

**实现方式**：
1. 使用InfluxDB Go客户端库
2. 实现批量写入，提高性能
3. 配置合理的保留策略

**数据流程**：
```go
// 1. 收集扫描结果和位置信息
scanResult := ScanResult{...}
location := IPLocation{...}

// 2. 构建InfluxDB数据点
point := influxdb2.NewPointWithMeasurement("taiwan_ip_scan").
	AddTag("ip", scanResult.IP).
	AddTag("country", location.Country).
	AddTag("region", location.Region).
	AddTag("city", location.City).
	AddTag("isp", location.ISP).
	AddTag("asn", fmt.Sprintf("%d", location.ASN)).
	AddTag("ip_segment", "1.2.3.0/24").
	AddField("is_alive", scanResult.IsAlive).
	AddField("scan_time", scanResult.ScanTime).
	AddField("response_time", scanResult.ResponseTime).
	AddField("open_ports", scanResult.OpenPorts).
	SetTime(scanResult.ScanTimestamp)

// 3. 写入InfluxDB
writeAPI.WritePoint(point)
```

## 6. 性能优化设计

### 6.1 并发设计

- **IP段扫描并发**：每个IP段分配一个或多个goroutine
- **IP扫描并发**：每个IP段内部使用并发扫描，控制并发数避免网络拥塞
- **位置查询并发**：批量查询IP位置信息，减少网络请求

### 6.2 缓存设计

- **IP段缓存**：本地缓存IP段数据，定期更新
- **位置信息缓存**：缓存已查询的IP位置信息，减少重复查询
- **扫描结果缓存**：缓存近期扫描结果，避免重复扫描

### 6.3 资源控制

- **速率限制**：控制SYN扫描速率，避免触发防火墙限制
- **内存控制**：限制并发数和缓存大小，避免内存溢出
- **超时处理**：设置合理的超时时间，避免资源长时间占用

## 7. 可靠性设计

- **错误恢复**：实现自动重试机制，处理网络异常和超时
- **日志记录**：详细记录扫描过程和结果，便于调试和分析
- **监控告警**：监控系统运行状态，异常情况及时告警
- **数据校验**：对IP段和扫描结果进行校验，确保数据准确性

## 8. 部署架构

### 8.1 单机部署

适合小规模扫描（如单个IP段或少量IP段）：

```
┌─────────────────────────────────────────────────┐
│                    单机系统                     │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────┐ │
│  │ IP段获取 │──│ SYN扫描 │──│ 位置查询 │──│Influx│ │
│  └─────────┘  └─────────┘  └─────────┘  └─────┘ │
└─────────────────────────────────────────────────┘
```

### 8.2 分布式部署

适合大规模扫描（如多个IP段或全台湾省IP段）：

```
┌─────────────────┐     ┌─────────────────┐
│   控制节点      │────▶│   扫描节点集群   │
└─────────────────┘     └─────────────────┘
        │                         │
        ▼                         ▼
┌─────────────────┐     ┌─────────────────┐
│  InfluxDB集群   │◀────│ IP位置查询服务  │
└─────────────────┘     └─────────────────┘
        │
        ▼
┌─────────────────┐
│   Grafana       │
└─────────────────┘
```

## 9. 扩展性设计

- **模块化设计**：各个功能模块独立，便于扩展和替换
- **配置驱动**：通过配置文件控制扫描参数、并发数等
- **插件机制**：支持添加新的IP段源、扫描方式和存储后端
- **API接口**：提供RESTful API，支持外部系统集成

## 10. 安全设计

- **权限控制**：限制系统访问权限，避免滥用
- **数据加密**：敏感数据加密存储和传输
- **审计日志**：记录所有操作，便于追溯
- **合规性**：遵守相关法律法规，合理使用扫描技术

## 11. 后续扩展功能

- **端口服务识别**：识别开放端口上的服务类型
- **漏洞扫描**：对活跃IP进行漏洞扫描
- **历史趋势分析**：分析IP活跃度的历史变化
- **异常检测**：检测异常IP活动
- **自动报告生成**：定期生成扫描报告

## 12. 实施计划

1. **第一阶段**：实现基础功能
   - IP段获取模块
   - SYN扫描模块
   - IP位置查询模块
   - InfluxDB存储模块

2. **第二阶段**：性能优化和可靠性提升
   - 并发优化
   - 缓存机制
   - 错误恢复和监控

3. **第三阶段**：扩展功能和分布式部署
   - 端口服务识别
   - 漏洞扫描
   - 分布式部署支持

4. **第四阶段**：可视化和分析
   - Grafana仪表盘
   - 历史趋势分析
   - 异常检测

## 13. 总结

本架构设计提供了一个完整的台湾省IP段扫描与InfluxDB存储系统解决方案，涵盖了从IP段获取、SYN扫描、IP位置查询到InfluxDB存储的全流程。系统采用模块化设计，具有良好的性能、可靠性和扩展性，适合不同规模的扫描需求。