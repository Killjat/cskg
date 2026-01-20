# 🔍 FOFA协议检测能力测试指南

## 📋 概述

本工具集成了FOFA API，用于获取真实的网络资产进行协议检测能力测试。通过FOFA搜索各种协议的真实服务器，然后使用我们的网络探测引擎进行检测，验证系统的实际检测能力。

## 🚀 快速开始

### 1. 准备FOFA凭据

首先需要FOFA会员账号和API Key：

1. 登录 [FOFA官网](https://fofa.info)
2. 进入个人中心 → API管理
3. 获取您的API Key

### 2. 配置文件设置

创建 `fofa_config.json` 配置文件：

```json
{
  "email": "your_email@example.com",
  "key": "your_fofa_api_key", 
  "base_url": "https://fofa.info/api/v1/search/all"
}
```

### 3. 运行测试

```bash
# 方式1: 使用便捷脚本
./run_fofa_test.sh

# 方式2: 直接运行Go程序
go run fofa_test_main.go fofa_tester.go

# 方式3: 测试单个协议
go run fofa_test_main.go fofa_tester.go -protocol mysql -verbose
```

## 🎯 测试模式

### 1. 全协议测试
测试所有49种支持的协议，每种协议搜索10个真实资产进行检测。

```bash
go run fofa_test_main.go fofa_tester.go -verbose
```

### 2. 单协议测试
针对特定协议进行深度测试。

```bash
go run fofa_test_main.go fofa_tester.go -protocol modbus -verbose
```

### 3. 自定义输出
指定报告输出文件。

```bash
go run fofa_test_main.go fofa_tester.go -output my_test_report.json
```

## 📊 支持的协议

### 🏭 工控协议 (5种)
- **modbus** - Modbus TCP (端口502)
- **dnp3** - DNP3电力协议 (端口20000,19999)  
- **bacnet** - BACnet楼宇自动化 (端口47808)
- **opcua** - OPC UA工业4.0 (端口4840,4843)
- **s7** - 西门子S7 PLC (端口102)

### 🗄️ 数据库协议 (10种)
- **mysql** - MySQL数据库 (端口3306)
- **postgresql** - PostgreSQL (端口5432)
- **redis** - Redis缓存 (端口6379)
- **sqlserver** - SQL Server (端口1433)
- **oracle** - Oracle数据库 (端口1521)
- **mongodb** - MongoDB (端口27017)
- **elasticsearch** - Elasticsearch (端口9200)
- **influxdb** - InfluxDB时序数据库 (端口8086)
- **cassandra** - Cassandra (端口9042)
- **neo4j** - Neo4j图数据库 (端口7687)

### 🌐 IoT协议 (4种)
- **mqtt** - MQTT消息队列 (端口1883)
- **coap** - CoAP受限设备协议 (端口5683)
- **lorawan** - LoRaWAN (端口1700)
- **amqp** - AMQP消息队列 (端口5672)

### 🏢 企业基础设施协议 (5种)
- **ldap** - LDAP目录服务 (端口389)
- **kerberos** - Kerberos认证 (端口88)
- **radius** - RADIUS认证 (端口1812)
- **ntp** - NTP时间同步 (端口123)
- **syslog** - Syslog日志 (端口514)

### 🔒 安全协议 (2种)
- **openvpn** - OpenVPN (端口1194)
- **wireguard** - WireGuard (端口51820)

### 📡 电信协议 (1种)
- **sip** - SIP语音协议 (端口5060)

### ☁️ 云服务协议 (2种)
- **docker** - Docker API (端口2375)
- **kubernetes** - Kubernetes API (端口6443)

### 📷 摄像头协议 (4种)
- **rtsp** - RTSP流媒体 (端口554)
- **onvif** - ONVIF设备管理 (端口80)
- **hikvision** - 海康威视
- **dahua** - 大华摄像头

### 🌐 网络基础协议 (10种)
- **http** - HTTP协议 (端口80)
- **https** - HTTPS协议 (端口443)
- **ssh** - SSH协议 (端口22)
- **ftp** - FTP协议 (端口21)
- **smtp** - SMTP邮件 (端口25)
- **dns** - DNS协议 (端口53)
- **snmp** - SNMP管理 (端口161)
- **telnet** - Telnet协议 (端口23)
- **pop3** - POP3邮件 (端口110)
- **imap** - IMAP邮件 (端口143)

**总计: 49种协议**

## 📈 测试报告

### 报告内容
测试完成后会生成详细的JSON报告，包含：

1. **总体统计**
   - 测试协议数量
   - 总目标数量
   - 成功/失败统计
   - 总体成功率

2. **协议详情**
   - 每个协议的目标数量
   - 检测成功率
   - 平均置信度
   - 详细的检测结果

3. **目标信息**
   - IP地址和端口
   - FOFA提供的元信息
   - 检测到的协议和Banner
   - 置信度评分

### 报告示例
```json
{
  "timestamp": "2024-01-20T10:30:00Z",
  "total_protocols": 49,
  "tested_protocols": 45,
  "statistics": {
    "total_targets": 450,
    "successful_tests": 387,
    "failed_tests": 63,
    "overall_success_rate": 86.0,
    "protocol_stats": {
      "mysql": {
        "targets_found": 10,
        "successful_tests": 9,
        "success_rate": 90.0,
        "avg_confidence": 95.2
      }
    }
  }
}
```

## 🎯 测试策略

### FOFA查询优化
每个协议都有专门优化的FOFA查询语句：

```bash
# 工控协议示例
modbus: port="502" && protocol="modbus"
opcua: port="4840" || (port="4843" && protocol="opcua")

# 数据库协议示例  
mysql: port="3306" && protocol="mysql"
mongodb: port="27017" && protocol="mongodb"

# IoT协议示例
mqtt: port="1883" && protocol="mqtt"
coap: port="5683" && protocol="coap"
```

### 检测逻辑
1. **FOFA搜索**: 使用优化的查询语句搜索真实资产
2. **目标筛选**: 每个协议最多测试10个目标
3. **协议探测**: 使用port模式进行快速探测
4. **结果验证**: 检查探测结果是否匹配预期协议
5. **置信度评估**: 记录检测置信度和Banner信息

## 🔧 高级用法

### 自定义协议测试
```bash
# 测试特定的工控协议
for protocol in modbus dnp3 bacnet opcua s7; do
    go run fofa_test_main.go fofa_tester.go -protocol $protocol -verbose
done

# 测试数据库协议
for protocol in mysql postgresql redis mongodb; do
    go run fofa_test_main.go fofa_tester.go -protocol $protocol
done
```

### 批量测试脚本
```bash
#!/bin/bash
protocols=("mysql" "redis" "mongodb" "elasticsearch" "ssh" "http")
for protocol in "${protocols[@]}"; do
    echo "Testing $protocol..."
    go run fofa_test_main.go fofa_tester.go -protocol "$protocol" -output "report_${protocol}.json"
done
```

## 📊 性能优化

### 请求频率控制
- 协议间延迟: 500ms
- 目标间延迟: 100ms
- 避免触发FOFA API限制

### 并发控制
- 探测引擎并发数: 10
- 单个协议顺序测试
- 避免网络拥塞

## ⚠️ 注意事项

### 1. API配额
- FOFA会员有API调用次数限制
- 建议分批测试，避免一次性消耗过多配额
- 可以先测试少量协议验证效果

### 2. 网络礼仪
- 测试使用的都是公开的网络资产
- 仅进行协议识别，不进行任何攻击行为
- 遵守网络安全法律法规

### 3. 结果解读
- 成功率受网络环境影响
- 某些协议可能因防火墙等原因检测失败
- 重点关注协议识别的准确性而非连通性

## 🎉 预期效果

基于我们的49种协议支持，预期测试结果：

- **总体成功率**: 80-90%
- **工控协议**: 85-95% (Modbus, OPC UA等)
- **数据库协议**: 90-95% (MySQL, Redis等)
- **网络基础协议**: 95%+ (HTTP, SSH等)
- **IoT协议**: 70-85% (MQTT, CoAP等)

这个测试将全面验证我们系统的实际检测能力，为进一步优化提供数据支持！

## 🚀 开始测试

```bash
# 1. 配置FOFA凭据
cp fofa_config.json.example fofa_config.json
# 编辑fofa_config.json填入您的凭据

# 2. 运行测试
./run_fofa_test.sh

# 3. 查看结果
ls -la fofa_test_report_*.json
```

准备好验证我们49种协议的检测能力了吗？🎯