# IP 地理位置查询工具

这是一个使用 Go 语言编写的 IP 地理位置查询工具，可以从文件中读取一个或多个 IP 地址，获取它们的地理位置信息，包括经纬度、城市、地区、国家等。

## 功能特性

- 从文件中读取多个 IP 地址
- 并发查询，提高效率
- 获取详细的地理位置信息：
  - 城市
  - 地区
  - 国家
  - 经纬度
  - 邮编
  - 时区
- 支持 IPv4 地址

## 安装

```bash
go mod tidy
```

## 编译

```bash
go build -o ipgeolocation main.go
```

## 使用方法

### 准备 IP 地址文件

创建一个文本文件，每行包含一个 IP 地址，例如：

```
8.8.8.8
1.1.1.1
202.108.22.5
```

### 运行程序

```bash
# 使用 go run 运行
./ipgeolocation -file ips.txt

# 或直接编译后运行
./ipgeolocation -file ips.txt
```

### 参数说明

- `-file`：指定包含 IP 地址的文件路径（必填）

## 输出示例

```
IP: 8.8.8.8
  城市: Mountain View
  地区: California
  国家: US
  经纬度: 38.008800, -122.117500
  邮编: 94043
  时区: America/Los_Angeles

IP: 1.1.1.1
  城市: Brisbane
  地区: Queensland
  国家: AU
  经纬度: -27.467900, 153.028100
  邮编: 9010
  时区: Australia/Brisbane
```

## 技术说明

- 使用 `ipinfo.io` 提供的免费 IP 地理位置 API
- 并发查询使用 Go 协程和 WaitGroup 实现
- 支持跨平台运行

## 注意事项

- 该程序使用免费的 IP 地理位置 API，有一定的请求限制
- 对于私有 IP 地址（如 192.168.x.x、10.x.x.x 等），可能无法获取详细的地理位置信息
- 确保程序有网络连接，能够访问 ipinfo.io
