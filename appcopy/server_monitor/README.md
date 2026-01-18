# 服务器监控系统

## 项目简介

服务器监控系统是一个用于监控Linux服务器的关键活动的工具，包括：
- 当前登录账号信息
- 文件系统操作（创建、读取、写入、删除、复制、移动）
- 进程状态和命令执行情况
- 当前正在执行的命令
- 系统资源使用情况

## 功能特性

- **实时监控**：实时采集和展示服务器各项指标
- **Web界面**：提供现代化的Web大屏展示
- **API接口**：提供RESTful API接口，便于与其他系统集成
- **数据持久化**：支持多种存储方式（SQLite、MySQL等）
- **灵活配置**：支持通过配置文件自定义系统行为
- **自动部署**：提供一键部署脚本
- **系统服务**：支持Systemd服务管理

## 技术栈

- **后端框架**：Gin (Go)
- **前端技术**：HTML5 + CSS3 + JavaScript
- **数据可视化**：Chart.js
- **配置管理**：YAML
- **数据存储**：SQLite（默认）、MySQL
- **系统服务**：Systemd

## 系统要求

- **操作系统**：Linux（Ubuntu/Debian、CentOS/RHEL、Rocky Linux、AlmaLinux）
- **架构**：x86_64、ARM64
- **Go版本**：1.21.0+（部署脚本会自动安装）
- **权限**：需要root权限进行部署

## 部署方式

### 一键部署

```bash
# 下载项目到本地
git clone <项目仓库地址>
cd server-monitor

# 给部署脚本添加执行权限
chmod +x deploy.sh

# 执行部署脚本
sudo ./deploy.sh deploy
```

### 手动部署

1. **安装Go环境**
   ```bash
   # 下载并安装Go 1.21.0
   wget https://golang.org/dl/go1.21.0.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
   echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc
   source ~/.bashrc
   ```

2. **编译项目**
   ```bash
   cd server-monitor
   go mod tidy
   go build -o server-monitor main.go
   ```

3. **创建配置文件**
   ```bash
   cp config.yaml.example config.yaml
   # 根据需要修改配置文件
   ```

4. **启动服务**
   ```bash
   # 直接启动
   ./server-monitor
   
   # 或者使用nohup在后台运行
   nohup ./server-monitor > server-monitor.log 2>&1 &
   ```

5. **配置Systemd服务**
   ```bash
   sudo cp systemd/server-monitor.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl start server-monitor
   sudo systemctl enable server-monitor
   ```

## 访问方式

- **Web界面**：http://<服务器IP>:8081/
- **API接口**：http://<服务器IP>:8081/api/v1/
- **健康检查**：http://<服务器IP>:8081/api/v1/health

## API接口文档

### 系统信息

- `GET /api/v1/health` - 健康检查
- `GET /api/v1/system-stats` - 获取系统统计信息

### 登录信息

- `GET /api/v1/logins` - 获取当前登录信息

### 进程信息

- `GET /api/v1/processes` - 获取进程列表
- `GET /api/v1/processes/:pid` - 获取指定进程信息
- `GET /api/v1/processes/count` - 获取进程数量统计

### 命令信息

- `GET /api/v1/commands` - 获取当前执行的命令

### 文件操作

- `GET /api/v1/file-operations` - 获取文件操作记录

## 配置文件说明

配置文件采用YAML格式，主要包含以下部分：

### 服务器配置
```yaml
server:
  port: 8081          # 服务端口
  host: 0.0.0.0        # 监听地址（0.0.0.0表示监听所有网络接口）
  read_timeout: 30     # 读取超时时间（秒）
  write_timeout: 30    # 写入超时时间（秒）
```

### 采集器配置
```yaml
collector:
  login_interval: 5    # 登录信息采集间隔（秒）
  process_interval: 2  # 进程信息采集间隔（秒）
  command_interval: 2  # 命令信息采集间隔（秒）
  file_watch_paths:    # 文件监控路径
  - /var/log
  - /etc
  recursive_watch: false  # 是否递归监控目录
```

### 存储配置
```yaml
storage:
  type: sqlite         # 存储类型（sqlite、mysql）
  file_path: ./server_monitor.db  # SQLite数据库文件路径
  host: localhost      # 数据库主机
  port: 3306           # 数据库端口
  username: root       # 数据库用户名
  password: ""         # 数据库密码
  database: server_monitor  # 数据库名称
```

### 告警配置
```yaml
alert:
  enabled: false       # 是否启用告警
  levels:              # 告警级别
  - warning
  - error
  - critical
  notification:        # 告警通知方式
  - email
  smtp:                # SMTP配置
    host: localhost
    port: 25
    username: ""
    password: ""
    from: ""
    to: ""
```

## 服务管理

### Systemd服务命令

```bash
# 启动服务
sudo systemctl start server-monitor

# 停止服务
sudo systemctl stop server-monitor

# 重启服务
sudo systemctl restart server-monitor

# 查看服务状态
sudo systemctl status server-monitor

# 设置开机自启
sudo systemctl enable server-monitor

# 查看日志
sudo journalctl -u server-monitor -f
```

## 日志管理

- **系统日志**：通过Systemd管理，使用journalctl查看
- **应用日志**：默认输出到标准输出，可通过重定向保存到文件

## 安全建议

1. **配置访问控制**：
   - 考虑限制允许访问的IP地址
   - 考虑添加用户名密码认证
   - 考虑使用HTTPS加密传输

2. **定期更新**：
   - 定期更新系统和依赖
   - 定期检查安全漏洞

3. **防火墙设置**：
   - 配置防火墙规则，只允许必要的端口访问
   - 考虑使用Nginx或Apache作为反向代理

## 卸载方式

```bash
# 执行卸载脚本
sudo ./deploy.sh uninstall
```

## 开发指南

### 项目结构

```
server-monitor/
├── internal/           # 内部包
│   ├── collector/      # 数据采集模块
│   ├── config/         # 配置管理
│   ├── model/          # 数据模型
│   └── storage/        # 数据存储
├── static/             # 静态资源（Web界面）
├── deploy.sh           # 部署脚本
├── go.mod              # Go模块文件
├── go.sum              # Go依赖校验文件
├── main.go             # 主程序入口
└── config.yaml         # 配置文件
```

### 编译命令

```bash
# 编译项目
go build -o server-monitor main.go

# 交叉编译（例如编译ARM64版本）
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o server-monitor-arm64 main.go
```

## 许可证

[MIT License](LICENSE)

## 贡献

欢迎提交Issue和Pull Request！

## 联系方式

- 项目地址：<项目仓库地址>
- 问题反馈：<Issues地址>
- 邮件：<维护者邮箱>
