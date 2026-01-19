# SSL证书分析工具

## 功能特性

### 基础功能
- 探测目标网站的SSL/TLS证书
- 提取证书关键信息（颁发者、有效期、主题、SAN等）
- 分析证书链完整性
- 检测证书安全性问题
- 支持批量URL检测
- 结果以JSON/CSV格式输出

### 高级安全分析 ⭐
- **威胁情报分析**: 基于IOC匹配的恶意证书检测
- **钓鱼检测**: 品牌拼写错误、同形异义字符攻击检测
- **DGA检测**: 基于熵值和模式的域名生成算法检测
- **时间线分析**: 证书签发时间异常检测
- **异常检测**: 多维度证书异常行为分析
- **风险评分**: 0-100分的综合安全评分
- **智能建议**: 基于分析结果的安全建议

### 相关网站搜索
- **搜索使用相同证书的其他网站**
- 支持多种搜索引擎（FOFA、Shodan、Censys、crt.sh）
- 基础设施关联分析
- 威胁情报扩展

## 使用方法

### 单个URL检测
```bash
./cert-analyzer -url https://example.com
```

### 批量检测
```bash
./cert-analyzer -file urls.txt
```

### 输出到文件
```bash
./cert-analyzer -url https://example.com -output results.json
```

### 启用高级安全分析
```bash
./cert-analyzer -u https://example.com --enable-advanced
```

### 威胁情报和钓鱼检测
```bash
./cert-analyzer -u https://example.com --enable-threat-intel --enable-phishing
```

### DGA检测
```bash
./cert-analyzer -u https://example.com --enable-dga
```

### 综合分析（搜索 + 高级分析）
```bash
./cert-analyzer -u https://example.com --enable-search --enable-advanced --config config.json
```

## 输出格式

```json
{
  "url": "https://example.com",
  "timestamp": "2026-01-19T10:30:00Z",
  "status": "success",
  "certificate": {
    "subject": {
      "common_name": "example.com",
      "organization": "Example Corp",
      "country": "US"
    },
    "issuer": {
      "common_name": "DigiCert SHA2 Secure Server CA",
      "organization": "DigiCert Inc",
      "country": "US"
    },
    "validity": {
      "not_before": "2023-01-01T00:00:00Z",
      "not_after": "2024-01-01T00:00:00Z",
      "days_remaining": 180
    },
    "san_domains": ["example.com", "www.example.com"],
    "signature_algorithm": "SHA256-RSA",
    "public_key": {
      "algorithm": "RSA",
      "size": 2048
    },
    "serial_number": "0123456789ABCDEF",
    "fingerprint_sha1": "AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD",
    "fingerprint_sha256": "11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00",
    "related_sites": {
      "search_method": "crtsh,fofa",
      "total_found": 15,
      "search_time_ms": 2500,
      "sites": [
        {
          "url": "https://www.example.org",
          "domain": "www.example.org",
          "port": 443,
          "title": "Example Organization",
          "server": "nginx/1.18.0",
          "country": "US",
          "confidence": 0.9,
          "source": "FOFA",
          "last_seen": "2026-01-19T10:00:00Z"
        },
        {
          "url": "https://api.example.com",
          "domain": "api.example.com", 
          "port": 443,
          "confidence": 0.7,
          "source": "crt.sh"
        }
      ],
      "last_updated": "2026-01-19T10:30:00Z"
    }
  },
  "security_analysis": {
    "is_expired": false,
    "expires_soon": false,
    "is_self_signed": false,
    "weak_signature": false,
    "certificate_chain_valid": true,
    "security_score": 85
  }
}
```

## 编译

```bash
go build -o cert-analyzer
```

## 配置文件

要启用相关网站搜索功能，需要创建配置文件 `config.json`：

```json
{
  "fofa": {
    "email": "your-email@example.com",
    "key": "your-fofa-api-key-here",
    "enabled": true
  },
  "shodan": {
    "api_key": "your-shodan-api-key-here",
    "enabled": false
  },
  "censys": {
    "app_id": "your-censys-app-id",
    "secret": "your-censys-secret",
    "enabled": false
  }
}
```

### 支持的搜索引擎

- **crt.sh**: 免费的证书透明度日志搜索（无需API密钥）
- **FOFA**: 网络空间搜索引擎（需要API密钥）
- **Shodan**: 物联网设备搜索引擎（需要API密钥）
- **Censys**: 互联网扫描平台（需要API密钥）

### 搜索参数

- `--enable-search`: 启用相关网站搜索
- `--search-methods`: 指定搜索方法（逗号分隔）
- `--max-search-results`: 最大搜索结果数量
- `--search-timeout`: 搜索超时时间
- `--config`: 配置文件路径