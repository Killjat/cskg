# IP发现系统

## 项目概述

IP发现系统通过APNIC数据获取台湾省的IP地址段，将其拆分成C段进行探活扫描，并将结果存储到InfluxDB的不同bucket中。

## 功能特性

1. **APNIC数据获取**: 从APNIC获取台湾省的IP地址分配数据
2. **IP段拆分**: 将获取的IP段拆分成C段（/24）
3. **IP探活**: 对C段中的IP进行ping探活检测
4. **InfluxDB存储**: 
   - IP段信息存储到 `taiwan_ip_segments` bucket
   - 探活结果存储到 `taiwan_ip_alive` bucket

## 项目结构

```
ip-discovery/
├── main.go              # 主程序入口
├── config.yaml          # 配置文件
├── go.mod               # Go模块文件
├── apnic/
│   ├── fetcher.go       # APNIC数据获取
│   └── parser.go        # APNIC数据解析
├── scanner/
│   ├── ping.go          # IP探活功能
│   └── segment.go       # IP段处理
├── storage/
│   ├── influxdb.go      # InfluxDB存储
│   └── models.go        # 数据模型
└── utils/
    └── ip.go            # IP工具函数
```

## 使用方法

```bash
# 1. 获取APNIC数据并解析IP段
./ip-discovery fetch

# 2. 扫描IP段并探活
./ip-discovery scan

# 3. 查看统计信息
./ip-discovery stats
```

## 配置说明

详见 `config.yaml` 文件中的注释。