# 网络空间扫描工具

一个使用 Go 语言实现的高性能网络空间扫描工具，支持 IP 探活、TCP/UDP 端口扫描、应用识别和响应包保存。

## 功能特性

✅ **IP 探活**
- 支持单个 IP
- 支持 IP 范围 (如: 192.168.1.1-192.168.1.10)
- 支持 CIDR 格式 (如: 192.168.1.0/24)

✅ **端口扫描**
- TCP 端口扫描
- UDP 端口扫描
- 从配置文件读取端口列表
- 高并发扫描

✅ **应用识别**
- 自动识别常见服务 (HTTP, SSH, MySQL, Redis 等)
- 获取服务 Banner 信息
- 应用指纹识别

✅ **响应包保存**
- 保存目标返回的响应包
- 按 IP、端口、协议分类存储
- 支持二进制数据保存

✅ **多种输出格式**
- JSON 格式
- CSV 格式
- TXT 文本格式

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置目标

编辑 `targets.txt` 文件，添加扫描目标：

```
# 单个IP
192.168.1.1

# IP范围
192.168.1.1-192.168.1.10

# CIDR格式
192.168.1.0/24
```

### 3. 配置端口

编辑 `config.yaml` 文件，设置要扫描的端口：

```yaml
ports:
  tcp:
    - 21
    - 22
    - 80
    - 443
    - 3306
  udp:
    - 53
    - 161
```

### 4. 运行扫描

```bash
# 使用默认配置
go run .

# 指定配置文件
go run . -c config.yaml -t targets.txt

# 自定义参数
go run . -t targets.txt -o ./results -w 200 -timeout 3000
```

### 5. 编译可执行文件

```bash
# Linux
go build -o scanner

# Windows
go build -o scanner.exe

# macOS
go build -o scanner
```

## 命令行参数

```
-c          配置文件路径 (默认: config.yaml)
-t          目标文件路径 (默认: targets.txt)
-o          输出目录 (默认: ./results)
-w          并发协程数 (默认: 100)
-timeout    超时时间，毫秒 (默认: 2000)
```

## 配置文件说明

### config.yaml

```yaml
# 扫描设置
scan:
  workers: 100        # 并发协程数
  timeout: 2000       # 超时时间(毫秒)
  retry: 1           # 重试次数

# 端口配置
ports:
  tcp:
    - 21    # FTP
    - 22    # SSH
    - 80    # HTTP
    - 443   # HTTPS
    - 3306  # MySQL
  udp:
    - 53    # DNS
    - 161   # SNMP

# 输出配置
output:
  directory: "./results"    # 输出目录
  format: "json"           # 格式: json/csv/txt
  save_response: true      # 是否保存响应包

# 目标文件
targets:
  file: "./targets.txt"
```

## 输出结果

### JSON 格式示例

```json
{
  "scan_time": "20240107_153045",
  "total": 10,
  "alive": 5,
  "results": [
    {
      "IP": "192.168.1.1",
      "IsAlive": true,
      "TCPPorts": [
        {
          "Port": 80,
          "Protocol": "tcp",
          "State": "open",
          "Service": "HTTP",
          "Banner": "HTTP/1.1 200 OK...",
          "Response": "base64encoded..."
        }
      ],
      "UDPPorts": []
    }
  ]
}
```

### CSV 格式

```
IP地址,端口,协议,状态,服务,Banner
192.168.1.1,80,tcp,open,HTTP,HTTP/1.1 200 OK...
192.168.1.1,443,tcp,open,HTTPS,
```

### 响应包存储结构

```
results/
├── responses/
│   ├── 192.168.1.1/
│   │   ├── 192.168.1.1_80_tcp.bin
│   │   └── 192.168.1.1_443_tcp.bin
│   └── 192.168.1.2/
├── scan_result_20240107_153045.json
└── scan_result_20240107_153045.csv
```

## 使用示例

### 示例 1: 扫描本地网络

```bash
# targets.txt
192.168.1.0/24

# 运行
go run . -t targets.txt -w 200
```

### 示例 2: 扫描特定IP范围

```bash
# targets.txt
10.0.0.1-10.0.0.50

# 运行
go run . -t targets.txt -o ./scan_results
```

### 示例 3: 快速扫描常见端口

```yaml
# config.yaml
ports:
  tcp:
    - 80
    - 443
    - 22
    - 3389
```

```bash
go run . -c config.yaml -timeout 1000 -w 500
```

## 性能优化

1. **调整并发数**: 根据网络带宽和目标数量调整 `-w` 参数
2. **设置合适的超时**: 内网扫描可以设置较小的超时值 (500-1000ms)
3. **分批扫描**: 大量目标可以分批扫描，避免资源耗尽

## 支持的服务

| 端口  | 服务          | 协议    |
|-------|---------------|---------|
| 21    | FTP           | TCP     |
| 22    | SSH           | TCP     |
| 23    | Telnet        | TCP     |
| 25    | SMTP          | TCP     |
| 53    | DNS           | TCP/UDP |
| 80    | HTTP          | TCP     |
| 443   | HTTPS         | TCP     |
| 3306  | MySQL         | TCP     |
| 3389  | RDP           | TCP     |
| 6379  | Redis         | TCP     |
| 27017 | MongoDB       | TCP     |
| ...   | 更多服务      | ...     |

## 注意事项

⚠️ **法律声明**
- 仅在授权范围内使用此工具
- 未经授权的扫描可能违反法律法规
- 使用者需对使用后果负责

⚠️ **使用建议**
- 建议先在内网或自己的服务器上测试
- 注意控制扫描速度，避免对目标造成影响
- 大规模扫描前请确保有足够的权限

## 开发计划

- [ ] ICMP 真实探活 (需要 root 权限)
- [ ] 更多应用指纹识别
- [ ] Web 界面
- [ ] 分布式扫描
- [ ] 漏洞检测模块

## License

MIT License
