# 网站信息收集工具 (Web Info Collector)

## 功能特性

### 基础信息收集
- **网站标题 (Title)**: 提取网页标题信息
- **网站图标 (Icon)**: 获取favicon和各种尺寸的图标
- **ICP备案信息**: 提取工信部ICP备案号
- **网安备案信息**: 提取公安部网安备案号
- **文件下载链接**: 发现并收集页面中的文件下载链接
- **页脚信息**: 提取网站页脚的版权、联系方式等信息

### 高级功能
- **批量扫描**: 支持批量URL处理
- **深度爬取**: 可配置爬取深度
- **智能识别**: 基于正则表达式和DOM解析的智能信息提取
- **多格式输出**: JSON、CSV、HTML报告
- **并发处理**: 高效的并发爬取机制

## 使用方法

### 单个网站分析
```bash
./web-info-collector -u https://example.com
```

### 批量分析
```bash
./web-info-collector -f urls.txt
```

### 深度爬取
```bash
./web-info-collector -u https://example.com -d 2 --max-pages 50
```

### 输出到文件
```bash
./web-info-collector -u https://example.com -o results.json --format json
```

## 输出格式

```json
{
  "url": "https://example.com",
  "timestamp": "2026-01-19T10:30:00Z",
  "status": "success",
  "basic_info": {
    "title": "Example Website - 示例网站",
    "description": "This is an example website",
    "keywords": "example, demo, website"
  },
  "icons": {
    "favicon": "https://example.com/favicon.ico",
    "apple_touch_icon": "https://example.com/apple-touch-icon.png",
    "icons": [
      {
        "url": "https://example.com/icon-192.png",
        "size": "192x192",
        "type": "image/png"
      }
    ]
  },
  "registration_info": {
    "icp_license": "京ICP备12345678号-1",
    "police_record": "京公网安备11010802012345号",
    "organization": "北京示例科技有限公司"
  },
  "download_links": [
    {
      "url": "https://example.com/files/document.pdf",
      "filename": "document.pdf",
      "type": "application/pdf",
      "size": "2.5MB"
    }
  ],
  "footer_info": {
    "copyright": "© 2024 Example Corp. All rights reserved.",
    "contact_info": {
      "email": "contact@example.com",
      "phone": "+86-10-12345678",
      "address": "北京市朝阳区示例大厦"
    },
    "links": [
      {"text": "关于我们", "url": "/about"},
      {"text": "联系我们", "url": "/contact"}
    ]
  },
  "technical_info": {
    "server": "nginx/1.18.0",
    "powered_by": "PHP/7.4.0",
    "cms": "WordPress 5.8",
    "frameworks": ["jQuery", "Bootstrap"]
  }
}
```

## 编译和安装

```bash
go build -o web-info-collector
```

## 配置选项

- `--depth, -d`: 爬取深度 (默认: 1)
- `--max-pages`: 最大页面数 (默认: 10)
- `--timeout`: 请求超时时间 (默认: 30s)
- `--concurrent, -c`: 并发数 (默认: 5)
- `--user-agent`: 用户代理字符串
- `--follow-redirects`: 跟随重定向 (默认: true)
- `--extract-files`: 提取文件链接 (默认: true)
- `--extract-footer`: 提取页脚信息 (默认: true)