# Taiwan IP Fetcher - 台湾网站IP获取工具

## 功能说明

自动获取100个台湾热门网站的IP地址，并保存到`targets.txt`文件中，用于网络空间扫描。

## 网站类别

包含以下类别的台湾网站：
- 政府机构（gov.tw等）
- 新闻媒体（自由时报、联合报、中时等）
- 电商平台（PChome、Momo购物等）
- 银行金融（国泰、中信、玉山等）
- 电信运营商（中华电信、台湾大哥大等）
- 教育机构（台大、成大、清华等）
- 交通运输（高铁、捷运等）
- 科技公司（台积电、联发科、华硕等）
- 社交论坛（PTT、Mobile01、Dcard等）
- 其他服务

## 使用方法

### 方式一：直接运行（推荐）

```bash
cd /Users/jatsmith/CodeBuddy/cskg/cyberspacescan/taiwan/cmd
go run main.go
```

### 方式二：编译后运行

```bash
cd /Users/jatsmith/CodeBuddy/cskg/cyberspacescan/taiwan/cmd
go build -o taiwan_fetcher
./taiwan_fetcher
```

### 方式三：自定义参数

```bash
# 指定输出文件
go run main.go -output=/path/to/targets.txt

# 指定获取数量
go run main.go -count=50

# 组合使用
go run main.go -output=custom.txt -count=80
```

## 参数说明

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-output` | `../targets.txt` | 输出文件路径 |
| `-count` | `100` | 获取IP数量 |

## 输出格式

生成的`targets.txt`文件格式：

```
# 台湾网站IP地址列表
# 生成时间: 2026-01-07 10:30:00
# 总数量: 100

140.112.8.116  # www.ntu.edu.tw
211.75.132.177  # www.gov.tw
61.219.11.28  # www.pchome.com.tw
...
```

## 项目结构

```
taiwan/
├── taiwan.go          # 核心功能包
├── cmd/
│   └── main.go        # 命令行工具
└── README.md          # 说明文档
```

## 功能特点

- ✅ 包含100+个台湾热门网站
- ✅ 自动DNS解析获取IP地址
- ✅ 超时控制（5秒）
- ✅ 失败自动跳过，继续下一个
- ✅ 实时显示解析进度
- ✅ 输出文件包含域名注释
- ✅ 可自定义输出路径和数量
- ✅ 防止请求过快（100ms间隔）

## 与扫描器集成

获取IP后，可直接用于网络空间扫描：

```bash
# 1. 获取台湾网站IP
cd taiwan/cmd
go run main.go

# 2. 执行扫描
cd ../..
go run main.go -targets=targets.txt
```

## 注意事项

1. 需要网络连接来解析DNS
2. 部分网站可能解析失败（防火墙、DNS限制等）
3. IP地址可能随时间变化，建议定期更新
4. 仅用于合法的网络安全研究和测试

## 依赖

- Go 1.16+
- 标准库（无需额外依赖）

## License

MIT
