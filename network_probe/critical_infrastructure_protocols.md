# 关键基础设施协议完整清单

## 🏭 工控协议 (Industrial Control Protocols)

### 1. **Modbus** - 工业通信标准
- **端口**: 502 (TCP), 502 (UDP)
- **用途**: PLC、HMI、SCADA通信
- **变种**: Modbus RTU, Modbus ASCII, Modbus TCP
- **厂商**: 施耐德、西门子、ABB等

### 2. **DNP3 (Distributed Network Protocol)**
- **端口**: 20000, 19999
- **用途**: 电力系统SCADA通信
- **应用**: 变电站、配电网自动化

### 3. **IEC 61850** - 电力系统通信标准
- **端口**: 102, 8102
- **用途**: 变电站自动化系统
- **协议**: MMS, GOOSE, SMV

### 4. **BACnet (Building Automation and Control Networks)**
- **端口**: 47808 (UDP), 47808 (TCP)
- **用途**: 楼宇自动化系统
- **应用**: 暖通空调、照明、安防

### 5. **EtherNet/IP** - 工业以太网
- **端口**: 44818, 2222
- **用途**: 罗克韦尔自动化设备通信
- **基于**: CIP (Common Industrial Protocol)

### 6. **PROFINET** - 西门子工业以太网
- **端口**: 34962, 34963, 34964
- **用途**: 西门子PLC和设备通信
- **实时**: 实时工业通信

### 7. **EtherCAT** - 实时工业以太网
- **端口**: 34980
- **用途**: 高性能运动控制
- **特点**: 微秒级实时性

### 8. **OPC UA (OPC Unified Architecture)**
- **端口**: 4840, 4843 (安全)
- **用途**: 工业4.0标准通信
- **特点**: 跨平台、安全、语义互操作

### 9. **S7 Protocol** - 西门子专有协议
- **端口**: 102
- **用途**: 西门子S7系列PLC通信
- **协议**: ISO-TSAP

### 10. **Omron FINS** - 欧姆龙协议
- **端口**: 9600
- **用途**: 欧姆龙PLC通信

## 🌐 IoT协议 (Internet of Things Protocols)

### 1. **MQTT** - 消息队列遥测传输 ✅已支持
- **端口**: 1883, 8883 (TLS)
- **用途**: 物联网消息传输
- **特点**: 轻量级、发布/订阅

### 2. **CoAP (Constrained Application Protocol)**
- **端口**: 5683 (UDP), 5684 (DTLS)
- **用途**: 受限设备通信
- **基于**: UDP, RESTful

### 3. **LoRaWAN** - 低功耗广域网
- **端口**: 1700 (UDP)
- **用途**: 长距离低功耗通信
- **应用**: 智慧城市、农业物联网

### 4. **Zigbee** - 无线个域网
- **频段**: 2.4GHz, 915MHz, 868MHz
- **用途**: 智能家居、工业自动化
- **特点**: 网状网络、低功耗

### 5. **Thread** - 低功耗网状网络
- **基于**: 6LoWPAN, IPv6
- **用途**: 智能家居设备互联
- **支持**: Google, Apple, Amazon

### 6. **Matter (Thread/WiFi)**
- **端口**: 5540
- **用途**: 智能家居统一标准
- **支持**: 主要科技公司联盟

### 7. **LwM2M (Lightweight M2M)**
- **端口**: 5683, 5684
- **用途**: 物联网设备管理
- **基于**: CoAP

### 8. **DDS (Data Distribution Service)**
- **端口**: 7400-7500
- **用途**: 实时系统数据分发
- **应用**: 工业4.0, 自动驾驶

### 9. **AMQP (Advanced Message Queuing Protocol)**
- **端口**: 5672, 5671 (TLS)
- **用途**: 企业消息队列
- **实现**: RabbitMQ, Apache Qpid

### 10. **XMPP (Extensible Messaging and Presence Protocol)**
- **端口**: 5222, 5223 (TLS)
- **用途**: 即时消息、物联网通信

## 🗄️ 数据库协议 (Database Protocols)

### 1. **MySQL** ✅已支持
- **端口**: 3306
- **协议**: MySQL Protocol
- **用途**: 关系型数据库

### 2. **PostgreSQL** ✅已支持
- **端口**: 5432
- **协议**: PostgreSQL Protocol
- **用途**: 高级关系型数据库

### 3. **Microsoft SQL Server**
- **端口**: 1433, 1434 (Browser)
- **协议**: TDS (Tabular Data Stream)
- **用途**: 企业级数据库

### 4. **Oracle Database**
- **端口**: 1521, 1522
- **协议**: TNS (Transparent Network Substrate)
- **用途**: 企业级数据库

### 5. **Redis** ✅已支持
- **端口**: 6379
- **协议**: RESP (Redis Serialization Protocol)
- **用途**: 内存数据库、缓存

### 6. **MongoDB**
- **端口**: 27017, 27018, 27019
- **协议**: MongoDB Wire Protocol
- **用途**: NoSQL文档数据库

### 7. **Cassandra**
- **端口**: 9042, 9160 (Thrift)
- **协议**: CQL (Cassandra Query Language)
- **用途**: 分布式NoSQL数据库

### 8. **Elasticsearch**
- **端口**: 9200, 9300
- **协议**: HTTP REST API
- **用途**: 搜索引擎、日志分析

### 9. **InfluxDB**
- **端口**: 8086, 8088
- **协议**: HTTP API, Line Protocol
- **用途**: 时序数据库

### 10. **CouchDB**
- **端口**: 5984
- **协议**: HTTP REST API
- **用途**: 文档型NoSQL数据库

### 11. **Neo4j**
- **端口**: 7474, 7687 (Bolt)
- **协议**: Bolt Protocol
- **用途**: 图数据库

### 12. **ClickHouse**
- **端口**: 8123, 9000, 9009
- **协议**: HTTP, Native TCP
- **用途**: 列式分析数据库

## 🏢 企业基础设施协议

### 1. **LDAP (Lightweight Directory Access Protocol)**
- **端口**: 389, 636 (LDAPS)
- **用途**: 目录服务、身份认证
- **实现**: Active Directory, OpenLDAP

### 2. **Kerberos**
- **端口**: 88, 464 (密码更改)
- **用途**: 网络认证协议
- **应用**: Windows域认证

### 3. **RADIUS**
- **端口**: 1812, 1813
- **用途**: 远程认证拨入用户服务
- **应用**: 网络设备认证

### 4. **TACACS+**
- **端口**: 49
- **用途**: 终端访问控制器访问控制系统
- **应用**: 网络设备管理

### 5. **DHCP**
- **端口**: 67, 68 (UDP)
- **用途**: 动态主机配置协议
- **功能**: IP地址分配

### 6. **NTP (Network Time Protocol)**
- **端口**: 123 (UDP)
- **用途**: 网络时间同步
- **重要性**: 日志关联、安全审计

### 7. **Syslog**
- **端口**: 514 (UDP), 6514 (TCP/TLS)
- **用途**: 系统日志传输
- **标准**: RFC 3164, RFC 5424

## 🌐 网络基础设施协议

### 1. **BGP (Border Gateway Protocol)**
- **端口**: 179
- **用途**: 互联网路由协议
- **重要性**: 全球互联网骨干

### 2. **OSPF (Open Shortest Path First)**
- **协议号**: 89
- **用途**: 内部网关协议
- **应用**: 企业网络路由

### 3. **MPLS (Multiprotocol Label Switching)**
- **用途**: 高性能网络转发
- **应用**: 运营商网络

### 4. **VXLAN (Virtual Extensible LAN)**
- **端口**: 4789 (UDP)
- **用途**: 网络虚拟化
- **应用**: 数据中心网络

### 5. **GRE (Generic Routing Encapsulation)**
- **协议号**: 47
- **用途**: 隧道协议
- **应用**: VPN、网络互联

## 🔒 安全基础设施协议

### 1. **IPSec**
- **协议**: ESP (50), AH (51), IKE (500/4500)
- **用途**: IP层安全
- **应用**: VPN、网络加密

### 2. **SSL/TLS** ✅已支持
- **端口**: 443, 993, 995等
- **用途**: 传输层安全
- **应用**: HTTPS、邮件加密

### 3. **OpenVPN**
- **端口**: 1194 (UDP/TCP)
- **用途**: VPN连接
- **特点**: 开源、跨平台

### 4. **WireGuard**
- **端口**: 51820 (UDP)
- **用途**: 现代VPN协议
- **特点**: 简单、高性能

### 5. **PPTP**
- **端口**: 1723
- **用途**: 点对点隧道协议
- **状态**: 已过时，不安全

## ☁️ 云服务协议

### 1. **AWS API**
- **端口**: 443 (HTTPS)
- **协议**: REST API, GraphQL
- **服务**: EC2, S3, RDS等

### 2. **Azure API**
- **端口**: 443 (HTTPS)
- **协议**: REST API
- **服务**: 虚拟机、存储等

### 3. **Google Cloud API**
- **端口**: 443 (HTTPS)
- **协议**: REST API, gRPC
- **服务**: Compute Engine, BigQuery等

### 4. **Kubernetes API**
- **端口**: 6443, 8080
- **协议**: REST API
- **用途**: 容器编排管理

### 5. **Docker API**
- **端口**: 2375, 2376 (TLS)
- **协议**: REST API
- **用途**: 容器管理

## 📡 电信基础设施协议

### 1. **SIP (Session Initiation Protocol)**
- **端口**: 5060, 5061 (TLS)
- **用途**: VoIP通话建立
- **应用**: IP电话系统

### 2. **RTP/RTCP**
- **端口**: 动态分配
- **用途**: 实时传输协议
- **应用**: 语音、视频传输

### 3. **H.323**
- **端口**: 1720, 1503
- **用途**: 多媒体通信
- **应用**: 视频会议系统

### 4. **MGCP (Media Gateway Control Protocol)**
- **端口**: 2427, 2727
- **用途**: 媒体网关控制
- **应用**: 电信网络

### 5. **Diameter**
- **端口**: 3868
- **用途**: AAA协议 (认证、授权、计费)
- **应用**: 移动网络、IMS

## 🚀 新兴协议

### 1. **gRPC**
- **端口**: 通常使用HTTP/2 (443, 80)
- **用途**: 高性能RPC框架
- **特点**: 基于HTTP/2, Protocol Buffers

### 2. **GraphQL**
- **端口**: 通常使用HTTP (80, 443)
- **用途**: API查询语言
- **特点**: 灵活的数据获取

### 3. **WebRTC**
- **端口**: 动态分配
- **用途**: 实时通信
- **应用**: 浏览器音视频通话

### 4. **QUIC**
- **端口**: 443 (UDP)
- **用途**: 快速UDP互联网连接
- **特点**: HTTP/3基础协议

## 📊 协议优先级建议

### 🔴 高优先级 (关键基础设施)
1. **Modbus** - 工控系统核心
2. **DNP3** - 电力系统关键
3. **OPC UA** - 工业4.0标准
4. **CoAP** - IoT核心协议
5. **SQL Server** - 企业数据库
6. **Oracle** - 企业数据库
7. **MongoDB** - 现代应用数据库

### 🟡 中优先级 (重要补充)
1. **BACnet** - 楼宇自动化
2. **EtherNet/IP** - 工业网络
3. **PROFINET** - 西门子生态
4. **LoRaWAN** - 物联网通信
5. **Elasticsearch** - 大数据分析
6. **InfluxDB** - 时序数据

### 🟢 低优先级 (特定场景)
1. **EtherCAT** - 高端运动控制
2. **Thread** - 智能家居
3. **Neo4j** - 图数据库
4. **Cassandra** - 大规模分布式

## 🎯 实施建议

基于当前已有的24个探测，建议按以下顺序添加：

### 第一阶段 (工控协议)
1. **Modbus TCP** - 最重要的工控协议
2. **DNP3** - 电力系统核心
3. **OPC UA** - 工业4.0标准
4. **BACnet** - 楼宇自动化

### 第二阶段 (数据库协议)
1. **SQL Server** - 企业级数据库
2. **Oracle** - 企业级数据库
3. **MongoDB** - NoSQL数据库
4. **Elasticsearch** - 搜索引擎

### 第三阶段 (IoT协议)
1. **CoAP** - 受限设备协议
2. **LoRaWAN** - 长距离IoT
3. **AMQP** - 企业消息队列
4. **DDS** - 实时数据分发

这样可以将我们的协议支持从24个扩展到40+个，覆盖所有关键基础设施领域！