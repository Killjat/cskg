# 工业协议蜜罐系统开发文档

## 1. 开发环境搭建

### 1.1 硬件要求
- CPU：至少4核
- 内存：至少8GB
- 存储空间：至少100GB
- 网络：稳定的网络连接

### 1.2 软件要求
- 操作系统：macOS 12+ 或 CentOS 7/8 或 Ubuntu 20.04/22.04
- Go语言：1.21+（推荐使用最新稳定版）
- Git：用于版本控制
- 编辑器：VS Code、GoLand 或其他支持Go语言的编辑器
- SQLite：用于本地开发和测试

### 1.3 Go环境配置

#### 1.3.1 安装Go语言

**macOS**：
```bash
# 使用Homebrew安装
brew install go

# 验证安装
go version
```

**CentOS**：
```bash
# 下载Go安装包
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz

# 解压到/usr/local
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# 添加环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export GOROOT=/usr/local/go' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

#### 1.3.2 配置Go代理

为了加速依赖下载，配置Go代理：
```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

### 1.4 项目克隆

```bash
git clone <repository-url>
cd honeypot-system
```

### 1.5 依赖安装

```bash
go mod download
```

## 2. 项目结构

### 2.1 目录结构

```
honeypot-system/
├── cmd/                     # 命令行入口
│   └── api/                 # API服务入口
│       └── main.go          # 主程序入口
├── config/                  # 配置文件
│   └── config.yaml.example  # 配置文件模板
├── internal/                # 内部包
│   ├── device/              # 设备指纹识别模块
│   │   └── fingerprint.go   # 设备指纹识别实现
│   ├── logger/              # 日志管理模块
│   │   └── logger.go        # 日志管理实现
│   ├── packet/              # 数据包捕获模块
│   │   └── capture.go       # 数据包捕获实现
│   └── web/                  # Web服务模块
│       └── server.go         # Web服务实现
├── web/                     # Web界面资源
│   ├── static/              # 静态文件
│   └── templates/           # HTML模板
│       └── index.html       # 主页模板
├── go.mod                   # Go模块依赖
├── go.sum                   # 依赖校验和
├── start.sh                 # 启动脚本（macOS）
├── start_centos.sh          # 启动脚本（CentOS）
└── README.md                # 项目说明文档
```

### 2.2 核心文件说明

| 文件路径                     | 说明                                 |
|--------------------------|------------------------------------|
| `cmd/api/main.go`        | 程序主入口，负责初始化和启动各个模块                |
| `config/config.yaml`      | 系统配置文件，包含蜜罐、数据包捕获、设备指纹等配置        |
| `internal/device/fingerprint.go` | 设备指纹识别模块，负责识别设备信息             |
| `internal/packet/capture.go` | 数据包捕获模块，负责监听和解析网络流量             |
| `internal/web/server.go`  | Web服务模块，提供API接口和Web界面              |
| `internal/logger/logger.go` | 日志管理模块，负责生成和管理系统日志             |
| `web/templates/index.html` | Web界面主页模板                         |

## 3. 模块开发指南

### 3.1 配置管理模块

#### 3.1.1 配置文件结构

配置文件采用YAML格式，包含以下主要部分：

```yaml
# 蜜罐系统基本配置
honeypot:
  name: "Industrial Protocol Honeypot"
  version: "1.0.0"
  log_path: "./logs"
  log_level: "info"

# 数据包捕获配置
packet_capture:
  enabled: true
  interfaces: ["any"]
  ports: [502, 3306, 6379, 9092, 80, 9999]
  full_capture: true
  save_path: "./data/pcap"

# 设备指纹配置
device_fingerprint:
  enabled: true
  db_path: "./data/fingerprints.db"
  rules:
    user_agent_analysis: true
    ja3_fingerprinting: true
    tcp_window_scaling: true
    tls_extensions: true

# Web服务配置
web:
  enabled: true
  host: "0.0.0.0"
  port: 8080
  https: false
  session_timeout: 3600
```

#### 3.1.2 添加新配置项

1. 在`config/config.go`中添加对应的结构体字段
2. 在`setDefaults()`函数中设置默认值
3. 更新配置文件模板`config/config.yaml.example`

### 3.2 数据包捕获模块

#### 3.2.1 添加新协议支持

1. 在`internal/packet/capture.go`的`parseProtocol`函数中添加新的端口处理
2. 实现对应的协议解析函数，如`parseNewProtocol`
3. 更新配置文件中的端口列表

#### 3.2.2 示例：添加新协议支持

```go
// parseProtocol 解析具体协议
func (pc *PacketCapture) parseProtocol(packet gopacket.Packet, srcIP string, srcPort int, dstIP string, dstPort int, protocol string, rawData []byte) {
    switch dstPort {
    case 502:
        pc.parseModbus(rawData)
    case 3306:
        pc.parseMySQL(rawData)
    case 6379:
        pc.parseRedis(rawData)
    case 9092:
        pc.parseKafka(rawData)
    case 8080: // 新增协议端口
        pc.parseNewProtocol(rawData) // 新增协议解析函数
    default:
        pc.logger.Debug(fmt.Sprintf("Unknown protocol on port %d", dstPort))
    }
}

// parseNewProtocol 解析新协议
func (pc *PacketCapture) parseNewProtocol(rawData []byte) {
    // 实现新协议的解析逻辑
    pc.logger.Debug(fmt.Sprintf("New protocol data: %x", rawData))
}
```

### 3.3 设备指纹识别模块

#### 3.3.1 设备识别算法扩展

1. 在`identifyDevice`函数中添加新的识别算法
2. 可以结合JA3指纹、TCP窗口缩放、TLS扩展等特征
3. 可以使用第三方指纹库或自定义规则

#### 3.3.2 示例：添加JA3指纹识别

```go
// identifyDevice 识别设备信息
func (fm *FingerprintManager) identifyDevice(rawData []byte) DeviceInfo {
    deviceInfo := DeviceInfo{
        OS:           "Unknown",
        DeviceType:   "Industrial Device",
        Manufacturer: "Unknown",
        DeviceModel:  "Unknown",
    }
    
    // JA3指纹识别
    if fm.ja3Fingerprinting {
        ja3Hash := fm.calculateJA3Hash(rawData)
        deviceInfo.JA3Hash = ja3Hash
        // 根据JA3哈希查询设备信息
        if device, exists := fm.ja3DeviceMap[ja3Hash]; exists {
            deviceInfo.OS = device.OS
            deviceInfo.DeviceType = device.DeviceType
            deviceInfo.Manufacturer = device.Manufacturer
            deviceInfo.DeviceModel = device.DeviceModel
        }
    }
    
    return deviceInfo
}
```

### 3.4 Web服务模块

#### 3.4.1 添加新API接口

1. 在`registerRoutes`函数中注册新路由
2. 实现对应的处理函数
3. 更新HTML模板（如果需要）

#### 3.4.2 示例：添加新API接口

```go
// registerRoutes 注册路由
func (s *Server) registerRoutes() {
    // 现有路由...
    
    // 新增API接口
    s.engine.GET("/api/new-endpoint", s.handleNewEndpoint)
}

// handleNewEndpoint 处理新API请求
func (s *Server) handleNewEndpoint(c *gin.Context) {
    // 实现API逻辑
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "New endpoint response",
    })
}
```

## 4. 编码规范

### 4.1 Go语言编码规范

1. **命名规范**：
   - 包名：使用小写字母，简短且有意义
   - 函数名：使用驼峰命名，首字母大写表示导出函数
   - 变量名：使用驼峰命名，首字母小写表示私有变量
   - 常量名：使用大写字母和下划线

2. **代码格式**：
   - 使用`go fmt`或`goimports`自动格式化代码
   - 每行代码长度不超过120个字符
   - 使用4个空格缩进（不使用制表符）

3. **注释规范**：
   - 为导出函数、结构体、接口添加文档注释
   - 为复杂的代码逻辑添加注释
   - 注释使用英文，清晰明了

4. **错误处理**：
   - 不要忽略错误，始终处理或返回错误
   - 使用`errors.New`或`fmt.Errorf`创建错误
   - 错误信息应包含足够的上下文信息

5. **并发安全**：
   - 共享变量使用互斥锁或其他同步机制保护
   - 避免在循环中创建goroutine
   - 使用`context`管理goroutine的生命周期

### 4.2 提交规范

1. **提交信息格式**：
   ```
   <type>: <description>
   
   <body>
   ```

2. **提交类型**：
   - `feat`：新功能
   - `fix`：修复bug
   - `docs`：文档更新
   - `style`：代码风格调整
   - `refactor`：代码重构
   - `test`：测试代码
   - `chore`：构建过程或辅助工具的变动

3. **提交示例**：
   ```
   feat: 添加Modbus协议支持
   
   - 实现Modbus TCP协议解析
   - 添加Modbus设备指纹识别
   - 更新配置文件模板
   ```

## 5. 构建与部署流程

### 5.1 本地开发构建

```bash
# 编译项目
go build -o honeypot_server ./cmd/api

# 运行项目
./honeypot_server
```

### 5.2 交叉编译

```bash
# 编译Linux版本
export GOOS=linux GOARCH=amd64
go build -o honeypot_server_linux ./cmd/api

# 编译Windows版本
export GOOS=windows GOARCH=amd64
go build -o honeypot_server_windows.exe ./cmd/api
```

### 5.3 容器化构建

```bash
# 构建Docker镜像
docker build -t honeypot-system .

# 运行Docker容器
docker run -d --name honeypot -p 8080:8080 -v ./data:/app/data -v ./logs:/app/logs honeypot-system
```

### 5.4 CI/CD集成

推荐使用GitHub Actions或GitLab CI进行持续集成和部署：

1. **CI流程**：
   - 代码提交后自动运行单元测试
   - 运行代码质量检查（如`golint`、`gofmt`）
   - 构建项目
   - 运行集成测试

2. **CD流程**：
   - 测试通过后自动构建Docker镜像
   - 推送Docker镜像到镜像仓库
   - 部署到测试环境或生产环境

## 6. 开发工具推荐

### 6.1 编辑器插件

**VS Code**：
- Go：官方Go语言插件，提供代码补全、调试、测试等功能
- GitLens：增强Git功能，显示代码作者和提交信息
- YAML：YAML文件语法高亮和验证
- Docker：Docker容器管理

**GoLand**：
- 内置Go语言支持，无需额外插件
- 强大的代码分析和重构功能
- 内置Docker支持
- 内置数据库工具

### 6.2 命令行工具

- `gofmt`：代码格式化
- `goimports`：自动添加和移除导入
- `golint`：代码质量检查
- `go vet`：静态代码分析
- `gosec`：安全代码检查
- `dlv`：Go调试器
- `godoc`：查看Go文档

## 7. 常见问题与解决方案

### 7.1 依赖下载失败

**问题**：`go mod download` 下载依赖失败

**解决方案**：
```bash
# 配置Go代理
go env -w GOPROXY=https://goproxy.cn,direct

# 清理模块缓存
go clean -modcache

# 重新下载依赖
go mod download
```

### 7.2 编译错误

**问题**：编译时出现 `undefined: xxx` 错误

**解决方案**：
- 检查是否缺少导入
- 检查函数名或变量名是否拼写错误
- 检查依赖版本是否正确
- 运行 `go mod tidy` 清理依赖

### 7.3 运行时错误

**问题**：运行时出现 `panic: xxx` 错误

**解决方案**：
- 查看错误堆栈信息，定位问题所在
- 检查配置文件是否正确
- 检查文件路径和权限
- 添加适当的错误处理

### 7.4 数据包捕获失败

**问题**：无法捕获网络数据包

**解决方案**：
- 确保程序以管理员或root权限运行
- 检查网络接口名称是否正确
- 检查防火墙设置，确保允许数据包捕获
- 确保系统已安装所需的依赖库

## 8. 开发流程

1. **需求分析**：理解需求文档，明确开发任务
2. **设计**：根据设计文档，设计模块结构和接口
3. **编码**：按照编码规范编写代码
4. **测试**：编写单元测试和集成测试
5. **代码审查**：进行代码审查，确保代码质量
6. **提交**：提交代码，编写清晰的提交信息
7. **构建**：构建项目，确保编译通过
8. **部署**：部署到测试环境进行测试
9. **调试**：根据测试结果调试和修复问题
10. **发布**：部署到生产环境

## 9. 版本管理

### 9.1 版本号格式

使用语义化版本号：`MAJOR.MINOR.PATCH`

- **MAJOR**：不兼容的API变更
- **MINOR**：向后兼容的功能添加
- **PATCH**：向后兼容的bug修复

### 9.2 分支管理

- `main`：主分支，用于生产环境部署
- `develop`：开发分支，用于集成新功能
- `feature/*`：功能分支，用于开发新功能
- `bugfix/*`：bug修复分支，用于修复bug
- `release/*`：发布分支，用于准备发布版本

## 10. 文档更新

- 代码变更后及时更新相关文档
- 文档应与代码保持同步
- 文档应清晰、准确、易于理解
- 可以使用Markdown格式编写文档

# 总结

本开发文档详细介绍了工业协议蜜罐系统的开发环境搭建、项目结构、模块开发指南、编码规范、构建与部署流程等内容。开发人员应遵循本文档的指导，确保代码质量和系统稳定性。

在开发过程中，应注重代码的可维护性、可扩展性和安全性，确保系统能够满足需求文档中的所有要求。同时，应定期进行代码审查和测试，确保系统的可靠性和安全性。