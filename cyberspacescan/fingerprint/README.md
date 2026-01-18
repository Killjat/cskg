# 🔍 指纹识别包 (Fingerprint)

专门用于网络扫描回包的指纹识别，支持识别各种服务、框架、中间件等。

## 📦 功能特性

- ✅ **多类别识别**：Web服务器、应用服务器、数据库、CMS、框架等
- ✅ **版本提取**：自动提取软件版本号
- ✅ **操作系统推断**：基于Banner信息推断OS
- ✅ **CPE生成**：生成标准CPE标识符
- ✅ **置信度评分**：每个识别结果都有置信度评分
- ✅ **标签系统**：支持基于标签的筛选和分类
- ✅ **Base64解码**：自动处理Base64编码的响应包

## 🎯 支持的指纹类型

### Web服务器
- Nginx
- Apache
- IIS (Microsoft)
- GHost
- Tomcat
- Jetty

### CDN与负载均衡
- Cloudflare
- Akamai
- F5 BIG-IP

### 编程语言与框架
- PHP
- ASP.NET
- Express (Node.js)
- Django (Python)
- Flask (Python)
- Spring Boot (Java)

### CMS内容管理系统
- WordPress
- Joomla
- Drupal

### 数据库
- MySQL
- PostgreSQL
- Redis
- MongoDB
- Elasticsearch

### 其他服务
- OpenSSH
- vsftpd / ProFTPD
- Postfix / Exim
- 操作系统识别

## 🚀 使用方法

### 基础用法

```go
import "cskg/cyberspacescan/fingerprint"

// 使用Banner识别
banner := "HTTP/1.1 200 OK\r\nServer: nginx/1.18.0\r\n"
fingerprints := fingerprint.Identify(banner, nil)

for _, fp := range fingerprints {
    fmt.Printf("产品: %s, 版本: %s, 类别: %s\n", 
        fp.Product, fp.Version, fp.Category)
}
```

### 使用Banner + 响应包

```go
banner := "HTTP/1.1 200 OK\r\nServer: nginx\r\n"
response := []byte("base64EncodedResponse...")

fingerprints := fingerprint.Identify(banner, response)
```

### 快速识别（仅Banner）

```go
fingerprints := fingerprint.IdentifyQuick(banner)
```

### 获取最高置信度结果

```go
top := fingerprint.GetTopFingerprint(banner, response)
if top != nil {
    fmt.Printf("最可能是: %s (置信度: %d%%)\n", 
        top.Product, top.Confidence)
}
```

### 检查是否包含特定标签

```go
if fingerprint.HasTag(banner, nil, "web") {
    fmt.Println("这是一个Web服务")
}
```

### 获取所有识别类别

```go
categories := fingerprint.GetCategories(banner, response)
fmt.Println("识别到的类别:", categories)
```

## 📊 数据结构

### Fingerprint 结构

```go
type Fingerprint struct {
    Product     string   // 产品名称，如 "Nginx"
    Version     string   // 版本号，如 "1.18.0"
    Category    string   // 类别，如 "Web服务器"
    OS          string   // 操作系统，如 "Linux/Ubuntu"
    DeviceType  string   // 设备类型
    CPE         string   // CPE标识，如 "cpe:/a:nginx:nginx:1.18.0"
    Vendor      string   // 厂商，如 "Nginx Inc."
    Tags        []string // 标签，如 ["web", "http", "proxy"]
    Confidence  int      // 置信度 (0-100)
    RawBanner   string   // 原始Banner
    Description string   // 描述
}
```

## 🧪 运行测试

```bash
cd fingerprint
go test -v
```

## 🎮 运行示例

```bash
cd fingerprint/examples
go run demo.go
```

## 📝 示例输出

```
📌 示例1: Nginx服务器
  [1] 产品: Nginx
      类别: Web服务器
      厂商: Nginx Inc.
      置信度: 95%
      标签: [web http proxy]
      CPE: cpe:/a:nginx_inc.:nginx:*

📌 示例2: Apache + PHP
  [1] 产品: Apache
      版本: 2.4.41
      类别: Web服务器
      厂商: Apache Software Foundation
      系统: Linux/Ubuntu
      置信度: 95%
      标签: [web http]
      CPE: cpe:/a:apache_software_foundation:apache:2.4.41
  [2] 产品: PHP
      版本: 7.4.3
      类别: 编程语言
      厂商: PHP Group
      置信度: 90%
      标签: [php language]
      CPE: cpe:/a:php_group:php:7.4.3
```

## 🔧 扩展指纹规则

在 `fingerprint.go` 的 `fingerprintRules` 数组中添加新规则：

```go
{
    Name:       "自定义服务",
    Category:   "服务类别",
    Vendor:     "厂商名称",
    Pattern:    regexp.MustCompile(`(?i)匹配模式`),
    Version:    regexp.MustCompile(`版本提取模式`),
    Confidence: 90,
    Tags:       []string{"标签1", "标签2"},
}
```

## 🎯 实际应用场景

### 1. 集成到扫描器

```go
// 在扫描结果中添加指纹识别
type ScanResult struct {
    IP          string
    Port        int
    Banner      string
    Response    []byte
    Fingerprint *fingerprint.Fingerprint // 添加指纹字段
}

// 扫描时识别指纹
result.Fingerprint = fingerprint.GetTopFingerprint(
    result.Banner, 
    result.Response,
)
```

### 2. 统计分析

```go
// 统计某个网段使用的Web服务器
webServers := make(map[string]int)
for _, result := range scanResults {
    fps := fingerprint.Identify(result.Banner, result.Response)
    for _, fp := range fps {
        if fp.Category == "Web服务器" {
            webServers[fp.Product]++
        }
    }
}
```

### 3. 安全审计

```go
// 查找过时版本的服务
for _, result := range scanResults {
    fps := fingerprint.Identify(result.Banner, result.Response)
    for _, fp := range fps {
        if fp.Product == "Apache" && fp.Version < "2.4.0" {
            fmt.Printf("发现过时版本: %s %s at %s\n", 
                fp.Product, fp.Version, result.IP)
        }
    }
}
```

## 📈 性能特点

- ⚡ 快速匹配：使用正则表达式预编译
- 💾 内存高效：规则在全局共享
- 🔄 并发安全：无状态设计，支持并发调用

## 🤝 贡献

欢迎添加更多指纹规则！请确保：
1. 规则准确性
2. 合理的置信度评分
3. 完善的版本提取模式
4. 添加对应的测试用例

## 📄 许可证

MIT License
