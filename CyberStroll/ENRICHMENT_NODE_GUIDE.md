# 网站数据富化节点使用指南

## 📋 概述

网站数据富化节点是CyberStroll系统的第5个核心节点，专门负责对Web资产进行深度数据富化。它会循环读取Elasticsearch中的HTTP/HTTPS协议资产，并对其进行多维度的数据分析和富化。

## 🎯 主要功能

### 1. 证书信息富化
- **功能**: 提取HTTPS网站的SSL/TLS证书信息
- **参考**: cert-analyzer项目
- **数据包括**:
  - 证书主题和颁发者
  - 证书有效期
  - 签名算法和公钥算法
  - DNS名称和IP地址
  - 是否为CA证书和自签名证书

### 2. API信息富化
- **功能**: 发现和分析网站的API接口
- **参考**: api-hunter项目
- **数据包括**:
  - API端点列表
  - 支持的HTTP方法
  - API文档地址
  - API版本和框架信息

### 3. 网站信息富化
- **功能**: 提取网站的基本信息和元数据
- **参考**: web-info-collector项目
- **数据包括**:
  - 网站标题、描述、关键词
  - 网站语言和字符集
  - 生成器和作者信息
  - 链接、图片、脚本等资源

### 4. 网站指纹识别
- **功能**: 识别网站使用的技术栈和框架
- **数据包括**:
  - 技术名称和版本
  - 技术分类
  - 识别置信度
  - 识别证据

### 5. 网站内容信息
- **功能**: 分析HTTP响应的详细信息
- **数据包括**:
  - HTTP状态码和响应头
  - 内容类型和长度
  - 响应体预览和哈希
  - 响应时间

## 🏗️ 系统架构

```
┌─────────────────────┐
│   网站富化节点       │
│                    │
│  ┌───────────────┐  │
│  │  工作协程1    │  │
│  └───────────────┘  │
│  ┌───────────────┐  │
│  │  工作协程2    │  │
│  └───────────────┘  │
│  ┌───────────────┐  │
│  │  工作协程N    │  │
│  └───────────────┘  │
└─────────────────────┘
         │
         ▼
┌─────────────────────┐
│   Elasticsearch     │
│                    │
│ 1. 读取Web资产      │
│ 2. 更新富化数据     │
└─────────────────────┘
```

## ⚙️ 配置说明

### 配置文件: `configs/enrichment_node.yaml`

```yaml
# 节点配置
node:
  id: "enrichment-node-001"
  name: "网站数据富化节点1"
  region: "default"

# Elasticsearch配置
elasticsearch:
  urls:
    - "http://localhost:9200"
  index: "cyberstroll_ip_scan"
  username: ""
  password: ""
  timeout: 30

# 富化配置
enrichment:
  batch_size: 50              # 每批处理的资产数量
  worker_count: 5             # 工作协程数量
  scan_interval: "5m"         # 扫描间隔
  request_timeout: "30s"      # HTTP请求超时
  max_retries: 3              # 最大重试次数
  
  # 功能开关
  enable_cert: true           # 启用证书信息富化
  enable_api: true            # 启用API信息富化
  enable_web_info: true       # 启用网站信息富化
  enable_fingerprint: true    # 启用指纹识别
  enable_content: true        # 启用内容信息富化
```

### 配置参数说明

| 参数 | 说明 | 默认值 | 建议值 |
|------|------|--------|--------|
| batch_size | 每批处理的资产数量 | 50 | 10-100 |
| worker_count | 工作协程数量 | 5 | 3-10 |
| scan_interval | 扫描间隔 | 5m | 1m-30m |
| request_timeout | HTTP请求超时 | 30s | 10s-60s |
| max_retries | 最大重试次数 | 3 | 1-5 |

## 🚀 启动和运行

### 1. 启动节点

```bash
# 使用默认配置启动
./bin/enrichment_node

# 使用指定配置启动
./bin/enrichment_node --config configs/enrichment_node.yaml
```

### 2. 查看运行状态

```bash
# 查看日志
tail -f logs/enrichment_node.log

# 查看进程状态
ps aux | grep enrichment_node
```

### 3. 停止节点

```bash
# 发送停止信号
kill -TERM <PID>

# 或使用Ctrl+C停止
```

## 📊 监控和统计

### 日志信息

富化节点会输出以下类型的日志：

```
[ENRICHMENT] 启动网站数据富化器...
[ENRICHMENT] 启动富化工作协程 0
[ENRICHMENT] 工作协程 0 开始处理Web资产
[ENRICHMENT] 工作协程 0 找到 25 个Web资产需要富化
[ENRICHMENT] 开始富化Web资产: https://example.com:443
[ENRICHMENT] Web资产富化完成: https://example.com:443
[ENRICHMENT] 富化统计: 总处理=100, 成功=95, 失败=5, 活跃工作协程=5
```

### 统计指标

- **总处理数**: 已处理的Web资产总数
- **成功富化数**: 成功富化的资产数量
- **失败富化数**: 富化失败的资产数量
- **活跃工作协程数**: 当前活跃的工作协程数量
- **最后处理时间**: 最后一次处理的时间戳

## 🔧 故障排除

### 常见问题

1. **Elasticsearch连接失败**
   ```
   错误: 创建Elasticsearch客户端失败: Elasticsearch连接测试失败
   解决: 检查ES服务状态和网络连接
   ```

2. **HTTP请求超时**
   ```
   错误: 请求失败 https://example.com: context deadline exceeded
   解决: 增加request_timeout配置或检查网络状况
   ```

3. **证书验证失败**
   ```
   错误: TLS连接失败: x509: certificate signed by unknown authority
   解决: 这是正常现象，系统会跳过证书验证以获取证书信息
   ```

4. **内存使用过高**
   ```
   解决: 减少batch_size和worker_count配置
   ```

### 调试模式

启用调试模式获取更详细的日志：

```yaml
# 在配置文件中设置
debug: true
logging:
  level: "debug"
```

## 🔄 协同工作机制

### 多节点协同

富化节点支持多个实例同时运行，通过以下机制避免重复处理：

1. **时间戳检查**: 只处理未富化或富化时间超过24小时的资产
2. **批量处理**: 每个节点处理不同的资产批次
3. **错误重试**: 失败的资产会在下次扫描中重新尝试

### 部署建议

- **单机部署**: 1-2个富化节点实例
- **集群部署**: 3-5个富化节点实例
- **大规模部署**: 根据Web资产数量动态调整

## 📈 性能优化

### 1. 配置优化

```yaml
# 高性能配置
enrichment:
  batch_size: 100
  worker_count: 10
  scan_interval: "1m"
  request_timeout: "15s"
```

### 2. 系统优化

```bash
# 增加文件描述符限制
ulimit -n 65536

# 优化网络参数
echo 'net.core.somaxconn = 65535' >> /etc/sysctl.conf
sysctl -p
```

### 3. 监控指标

- **处理速度**: 每分钟处理的资产数量
- **成功率**: 成功富化的比例
- **响应时间**: HTTP请求的平均响应时间
- **资源使用**: CPU和内存使用情况

## 🔐 安全注意事项

1. **网络安全**: 富化节点会主动访问目标网站，注意网络安全策略
2. **数据隐私**: 富化的数据可能包含敏感信息，注意数据保护
3. **访问频率**: 避免对目标网站造成过大压力，合理设置扫描间隔
4. **证书处理**: 系统会跳过证书验证，注意安全风险

## 📚 扩展开发

### 添加新的富化功能

1. 在`EnrichmentData`结构中添加新字段
2. 实现对应的富化方法
3. 在`enrichWebAsset`方法中调用新功能
4. 更新配置文件添加功能开关

### 集成外部工具

富化节点设计时考虑了与现有工具的集成：

- **cert-analyzer**: 证书分析工具
- **api-hunter**: API发现工具
- **web-info-collector**: 网站信息收集工具

可以通过调用这些工具的API或命令行接口来增强富化能力。

---

## 📞 技术支持

如有问题，请：
1. 查看日志文件获取详细错误信息
2. 检查配置文件和网络连接
3. 参考故障排除章节
4. 联系开发团队或提交GitHub Issue