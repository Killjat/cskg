# CyberStroll 搜索界面使用指南

## 概述

CyberStroll 搜索节点提供了类似 FOFA 风格的网络空间搜索界面，允许用户通过多种条件搜索和查询扫描结果。

## 功能特性

### 🔍 搜索功能
- **多条件搜索**: 支持 IP、端口、Banner、服务、协议、国家等多维度搜索
- **模糊匹配**: 支持关键词模糊搜索和精确匹配
- **范围查询**: 支持 IP 段、端口范围查询
- **分页浏览**: 支持大量结果的分页显示
- **实时统计**: 显示搜索结果的统计信息

### 📊 数据展示
- **详细信息**: 显示 IP、端口、服务、Banner、地理位置等完整信息
- **可视化统计**: 展示服务分布、协议统计、地理分布等
- **响应式设计**: 支持桌面和移动设备访问

### 📤 数据导出
- **JSON 格式**: 导出结构化数据
- **CSV 格式**: 导出表格数据，便于分析

## 部署和启动

### 本地部署

1. **构建搜索节点**
```bash
./build_search_node.sh
```

2. **启动搜索节点**
```bash
./start_search_node.sh
```

3. **访问界面**
- Web 界面: http://localhost:8082
- API 接口: http://localhost:8082/api/

### Docker 部署

1. **使用 Docker Compose 启动**
```bash
docker-compose up -d search-node
```

2. **访问界面**
- Web 界面: http://localhost:8082

## API 接口

### 搜索接口

**GET /api/search**

查询参数:
- `query`: 通用搜索关键词
- `ip`: IP 地址或 IP 段 (如: 192.168.1.1 或 192.168.1.0/24)
- `port`: 端口号或端口范围 (如: 80 或 80-8080)
- `banner`: Banner 内容搜索
- `service`: 服务类型 (如: http, ssh, ftp)
- `protocol`: 协议类型 (tcp, udp)
- `country`: 国家名称
- `page`: 页码 (默认: 1)
- `size`: 每页结果数 (默认: 20, 最大: 100)

**示例请求:**
```bash
# 搜索 Apache 服务器
curl "http://localhost:8082/api/search?query=apache"

# 搜索指定 IP 段的 HTTP 服务
curl "http://localhost:8082/api/search?ip=192.168.1.0/24&service=http"

# 搜索端口范围
curl "http://localhost:8082/api/search?port=8000-9000"

# 组合搜索
curl "http://localhost:8082/api/search?service=ssh&country=China&page=1&size=50"
```

**响应格式:**
```json
{
  "total": 1234,
  "page": 1,
  "size": 20,
  "results": [
    {
      "ip": "192.168.1.100",
      "port": 80,
      "protocol": "tcp",
      "service": "http",
      "service_version": "Apache/2.4.41",
      "banner": "HTTP/1.1 200 OK\r\nServer: Apache/2.4.41...",
      "state": "open",
      "scan_time": "2026-01-20T12:00:00Z",
      "geo_info": {
        "country": "China",
        "city": "Beijing",
        "latitude": 39.9042,
        "longitude": 116.4074
      }
    }
  ],
  "stats": {
    "total_results": 20,
    "services": {"http": 15, "https": 5},
    "protocols": {"tcp": 20},
    "countries": {"China": 12, "United States": 8}
  }
}
```

### 统计接口

**GET /api/stats**

获取 Elasticsearch 索引统计信息。

### 导出接口

**GET /api/export**

导出搜索结果，支持与搜索接口相同的查询参数。

查询参数:
- `format`: 导出格式 (json, csv)
- 其他搜索参数同搜索接口

**示例:**
```bash
# 导出 JSON 格式
curl "http://localhost:8082/api/export?service=http&format=json" -o results.json

# 导出 CSV 格式
curl "http://localhost:8082/api/export?service=http&format=csv" -o results.csv
```

## 搜索语法

### IP 地址搜索
- **单个 IP**: `192.168.1.1`
- **IP 段**: `192.168.1.0/24`
- **IP 范围**: `192.168.1.1-192.168.1.100`

### 端口搜索
- **单个端口**: `80`
- **端口范围**: `80-8080`
- **多个端口**: `80,443,8080`

### 服务搜索
- **HTTP 服务**: `service:http`
- **SSH 服务**: `service:ssh`
- **FTP 服务**: `service:ftp`

### Banner 搜索
- **包含关键词**: `banner:Apache`
- **版本信息**: `banner:"Apache/2.4"`
- **错误信息**: `banner:"404 Not Found"`

### 地理位置搜索
- **国家**: `country:China`
- **城市**: `city:Beijing`

### 组合搜索
```
ip:192.168.1.0/24 AND port:80
service:http AND country:China
banner:Apache AND port:80-8080
```

## Web 界面使用

### 搜索页面功能

1. **搜索框**: 输入关键词进行全文搜索
2. **过滤器**: 使用各种条件过滤结果
   - IP 地址过滤
   - 端口过滤
   - 服务类型过滤
   - 协议过滤
   - 国家过滤
   - Banner 内容过滤

3. **结果展示**:
   - IP 地址和端口信息
   - 服务类型和版本
   - Banner 详细信息
   - 地理位置信息
   - 扫描时间

4. **统计面板**: 显示搜索结果的统计信息
5. **分页导航**: 浏览大量搜索结果
6. **导出功能**: 导出搜索结果

### 搜索技巧

1. **精确搜索**: 使用引号包围关键词
2. **范围搜索**: 使用连字符表示范围
3. **组合搜索**: 同时使用多个过滤条件
4. **通配符**: 使用 * 进行模糊匹配

## 配置说明

### 搜索节点配置 (configs/search_node.yaml)

```yaml
# 节点基本信息
node:
  id: "search-node-001"
  name: "搜索节点1"
  region: "default"

# Elasticsearch配置
elasticsearch:
  urls:
    - "http://localhost:9200"
  index: "cyberstroll_ip_scan"
  username: ""
  password: ""
  timeout: 30

# Web服务配置
web:
  host: "0.0.0.0"
  port: 8082
  tls:
    enabled: false
    cert_file: ""
    key_file: ""

# 日志配置
logging:
  level: "info"
  file: "logs/search_node.log"
  max_size: "100MB"
  max_backups: 10
  max_age: 30
  compress: true
```

## 性能优化

### Elasticsearch 优化
1. **索引优化**: 合理设置分片和副本数量
2. **查询优化**: 使用合适的查询类型
3. **缓存优化**: 启用查询缓存
4. **硬件优化**: 增加内存和 SSD 存储

### 搜索节点优化
1. **并发控制**: 限制同时搜索请求数量
2. **结果缓存**: 缓存热门搜索结果
3. **分页优化**: 合理设置分页大小
4. **超时控制**: 设置合理的查询超时时间

## 故障排除

### 常见问题

1. **无法连接 Elasticsearch**
   - 检查 Elasticsearch 是否运行
   - 验证连接地址和端口
   - 检查网络连通性

2. **搜索结果为空**
   - 确认索引中有数据
   - 检查搜索条件是否正确
   - 验证索引名称配置

3. **搜索速度慢**
   - 优化 Elasticsearch 配置
   - 减少搜索结果数量
   - 使用更精确的搜索条件

4. **Web 界面无法访问**
   - 检查端口是否被占用
   - 验证防火墙设置
   - 查看搜索节点日志

### 日志查看

```bash
# 查看搜索节点日志
tail -f logs/search_node.log

# 查看 Docker 容器日志
docker logs cyberstroll-search-node
```

## 安全考虑

1. **访问控制**: 配置适当的网络访问控制
2. **数据保护**: 保护敏感的扫描数据
3. **查询限制**: 限制大量数据的导出
4. **日志审计**: 记录搜索和访问日志

## 扩展功能

### 未来计划
- [ ] 高级搜索语法支持
- [ ] 搜索结果可视化图表
- [ ] 用户认证和权限管理
- [ ] 搜索历史和收藏功能
- [ ] API 限流和配额管理
- [ ] 多语言界面支持

## 技术支持

如有问题或建议，请查看:
- 项目文档: `DEVELOPMENT_GUIDE.md`
- 部署指南: `DEPLOYMENT_GUIDE.md`
- 系统日志: `logs/search_node.log`