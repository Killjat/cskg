# Banner指纹识别引擎

一个基于Go语言的高性能Banner指纹识别引擎，支持Nmap规则库和用户自定义规则。

## 🚀 特性

- **高性能**: 预编译正则表达式，智能缓存机制
- **Nmap兼容**: 支持加载标准的nmap-service-probes文件
- **简单易用**: 提供简化的规则添加方式
- **交互模式**: 支持命令行交互操作
- **规则管理**: 支持规则的增删改查和导入导出
- **多格式支持**: 支持JSON和YAML格式的规则文件

## 📦 安装和使用

### 编译
```bash
go build -o banner_engine .
```

### 基本使用
```bash
# 匹配单个Banner
./banner_engine -banner "SSH-2.0-OpenSSH_8.2p1"

# 使用Nmap规则文件
./banner_engine -rules /usr/share/nmap/nmap-service-probes -banner "nginx/1.18.0"

# 交互模式
./banner_engine -interactive

# JSON输出
./banner_engine -banner "Apache/2.4.41" -output json
```

## 🎯 交互模式

进入交互模式后，可以使用以下命令：

```
banner> help                    # 显示帮助
banner> SSH-2.0-OpenSSH_8.2p1   # 直接输入Banner进行匹配
banner> match nginx/1.18.0      # 使用match命令匹配
banner> stats                   # 显示统计信息
banner> rules                   # 显示已加载的规则
banner> template                # 创建规则模板
banner> quit                    # 退出
```

## 📝 添加自定义规则

### 方法1: 交互模式添加
```bash
banner> add {"service":"myapp","pattern":"MyApp[/\\s]+(\\d+\\.\\d+)","product":"My Application","version":"$1","confidence":85}
```

### 方法2: 创建规则文件
```bash
# 创建模板
banner> template

# 编辑 rule_template.yaml
service: myapp
pattern: 'MyApp[/\s]+(\d+\.\d+)'
product: My Application
version: $1
description: My custom application
confidence: 85
```

### 方法3: 直接创建JSON文件
在 `./rules/` 目录下创建 `.json` 文件：

```json
{
  "id": "myapp_1",
  "service": "myapp",
  "pattern": "MyApp[/\\s]+(\\d+\\.\\d+)",
  "product": "My Application",
  "version": "$1",
  "confidence": 85,
  "description": "My custom application",
  "author": "user"
}
```

## 🔧 规则格式说明

### 完整规则格式
```json
{
  "id": "规则唯一标识",
  "service": "服务名称",
  "pattern": "正则表达式模式",
  "product": "产品名称模板",
  "version": "版本提取模板",
  "info": "附加信息模板",
  "hostname": "主机名模板",
  "os": "操作系统模板",
  "device_type": "设备类型模板",
  "cpe": "CPE标识模板",
  "confidence": 85,
  "description": "规则描述",
  "author": "规则作者"
}
```

### 简化规则格式
```yaml
service: http
pattern: 'nginx[/\s]+(\d+\.\d+\.\d+)'
product: nginx
version: $1
description: Nginx web server
confidence: 90
```

### 模板变量
在产品名称、版本等字段中可以使用以下变量：
- `$1`, `$2`, `$3` ... - 正则表达式捕获组
- 直接文本 - 固定字符串

## 📊 示例

### 1. Web服务器识别
```bash
# Nginx
banner> nginx/1.18.0 (Ubuntu)
✅ 匹配到 1 个服务:
1. http (nginx) v1.18.0 - 置信度: 90%

# Apache
banner> Apache/2.4.41 (Ubuntu)
✅ 匹配到 1 个服务:
1. http (Apache httpd) v2.4.41 - 置信度: 90%
```

### 2. SSH服务识别
```bash
banner> SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5
✅ 匹配到 1 个服务:
1. ssh (OpenSSH) v8.2p1 - 置信度: 95%
   信息: protocol 2.0
```

### 3. 数据库服务识别
```bash
banner> 5.7.34-0ubuntu0.18.04.1-log mysql_native_password
✅ 匹配到 1 个服务:
1. mysql (MySQL) v5.7.34 - 置信度: 90%
```

## 🛠️ 高级功能

### 批量导入规则
```bash
# 从文件导入规则
banner> import my_rules.json

# 导出当前规则
banner> export backup_rules.json
```

### 性能调优
```bash
# 设置最小置信度过滤
./banner_engine -banner "test" -min-confidence 80

# 自定义规则目录
./banner_engine -rules-dir /path/to/rules -interactive
```

## 📁 目录结构

```
banner_engine/
├── main.go              # 主程序
├── types.go             # 类型定义
├── engine.go            # 核心引擎
├── nmap_loader.go       # Nmap加载器
├── rule_manager.go      # 规则管理器
├── rules/               # 用户规则目录
│   ├── custom1.json
│   └── custom2.yaml
└── README.md
```

## 🔍 内置规则

引擎内置了常见服务的识别规则：

- **Web服务器**: nginx, Apache, IIS
- **SSH服务**: OpenSSH
- **数据库**: MySQL, Redis
- **FTP服务**: vsftpd
- **邮件服务**: Postfix

## 🚀 性能特点

- **快速匹配**: 预编译正则表达式，毫秒级响应
- **智能缓存**: 自动缓存匹配结果，提升重复查询性能
- **内存优化**: 高效的内存使用，支持大规模规则集
- **并发安全**: 支持多协程并发访问

## 📈 扩展性

- **插件化设计**: 易于添加新的规则加载器
- **标准接口**: 兼容Nmap等主流工具的规则格式
- **灵活配置**: 支持多种配置选项和自定义参数

## 🤝 贡献

欢迎提交Issue和Pull Request来改进项目！

## 📄 许可证

MIT License