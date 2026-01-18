# 局域网扫描与流量监控系统

## 项目简介

本项目是一个基于Python的局域网扫描与流量监控系统，能够自动发现局域网内的活跃设备，扫描设备开放端口，识别服务类型和应用程序，并实时监听网络流量。所有扫描结果和流量数据将保存到MySQL数据库中，通过Web界面进行展示和管理，支持CSV格式数据导出。

## 功能特点

### 1. 设备扫描
- 自动发现局域网内所有活跃设备
- 获取设备IP地址、MAC地址和主机名
- 支持指定网络地址和扫描速度

### 2. 端口扫描
- 扫描设备开放端口（支持1-65536端口范围）
- 识别TCP和UDP端口
- 自动识别常用服务类型
- 支持不同扫描速度设置

### 3. 流量监控
- 实时监听局域网内的网络流量
- 记录数据包的源IP、目标IP、端口、协议和长度
- 支持流量数据可视化展示

### 4. 数据管理
- 将所有数据保存到MySQL数据库
- 提供Web界面展示设备信息、端口信息和流量数据
- 支持CSV格式数据导出
- 响应式设计，支持不同设备访问

## 技术栈

- **后端**：Python 3.10+
- **Web框架**：Flask
- **网络扫描**：Scapy、Socket
- **数据库**：MySQL
- **前端**：HTML5、CSS3、JavaScript、Bootstrap 5

## 项目结构

```
lan_scanner/
├── database/          # 数据库相关模块
│   └── db.py         # 数据库连接和操作
├── scanner/           # 扫描相关模块
│   └── scanner.py     # 局域网扫描和流量监听
├── web/               # Web界面相关文件
│   ├── templates/     # HTML模板
│   │   ├── index.html        # 首页
│   │   ├── devices.html      # 设备列表
│   │   ├── device_detail.html # 设备详情
│   │   └── traffic.html      # 流量监控
│   └── server.py      # Web服务器
├── main.py           # 主程序入口
├── requirements.txt  # 项目依赖
├── test_db.py        # 数据库测试脚本
└── README.md         # 项目说明文档
```

## 安装和使用

### 1. 安装依赖

```bash
pip install -r requirements.txt
```

### 2. 配置MySQL数据库

- 安装MySQL数据库
- 创建数据库：`CREATE DATABASE lan_scan;
- 创建用户并授权：
  ```sql
  CREATE USER 'lan_scan_user'@'localhost' IDENTIFIED BY 'password';
  GRANT ALL PRIVILEGES ON lan_scan.* TO 'lan_scan_user'@'localhost';
  FLUSH PRIVILEGES;
  ```
- 修改 `database/db.py` 中的数据库连接信息

### 3. 运行项目

#### 3.1 扫描局域网设备

```bash
# 扫描本地网络
python main.py scan

# 指定网络地址
python main.py scan -n 192.168.1.0/24

# 设置扫描速度（T1最慢最准确，T5最快）
python main.py scan -s T2
```

#### 3.2 启动流量监听

```bash
python main.py traffic
```

#### 3.3 启动Web服务器

```bash
cd web
python server.py
```

Web界面访问地址：http://localhost:5000

## Web界面功能

### 1. 首页
- 系统介绍和功能说明
- 快速导航到设备列表和流量监控

### 2. 设备列表
- 显示所有扫描到的设备
- 展示设备IP、MAC地址、主机名和状态
- 支持设备详情查看
- 支持设备数据CSV导出

### 3. 设备详情
- 显示设备的详细信息
- 展示设备开放端口列表
- 显示端口的协议、状态、服务和应用
- 支持端口数据CSV导出

### 4. 流量监控
- 显示实时和历史流量数据
- 展示数据包的源IP、目标IP、端口和协议
- 支持流量数据CSV导出
- 自动刷新流量数据

## 注意事项

1. 运行扫描和流量监听功能需要管理员权限
2. 扫描速度设置会影响扫描结果的准确性和扫描时间
3. 流量监听会产生大量数据，建议定期清理数据库
4. Web服务器默认监听5000端口，可以在 `web/server.py` 中修改

## 许可证

MIT License
