# 网络空间扫描Web界面

## 功能特性

### 🎯 核心功能
- **实时扫描进度**：WebSocket实时推送扫描进度和结果
- **可视化界面**：现代化UI设计，响应式布局
- **历史结果管理**：查看、下载历史扫描结果
- **配置灵活**：自定义扫描参数（目标、端口、并发、超时）

### 📊 界面模块

1. **扫描配置区**
   - 目标输入（支持IP/CIDR/范围）
   - TCP/UDP端口配置
   - 并发数和超时设置

2. **实时进度显示**
   - 进度条动画
   - 统计数据（总目标、存活主机、开放端口）
   - 实时日志输出

3. **历史结果**
   - 结果列表展示
   - 在线查看JSON
   - 一键下载

## 使用方法

### 启动服务

```bash
cd /Users/jatsmith/CodeBuddy/cskg/cyberspacescan/web
./webserver
```

### 访问界面

打开浏览器访问：
```
http://localhost:8888
```

### 配置扫描

1. 在"扫描配置"区域输入目标IP
2. 配置TCP/UDP端口
3. 设置并发数和超时时间
4. 点击"🚀 开始扫描"

### 查看结果

1. 扫描完成后，结果自动显示在"历史扫描结果"区域
2. 点击"👁️ 查看"按钮在线查看JSON结果
3. 点击"⬇️ 下载"按钮下载结果文件

## 技术栈

- **后端**：Go + Gorilla WebSocket
- **前端**：原生HTML + CSS + JavaScript
- **通信**：WebSocket实时双向通信
- **存储**：本地文件系统（results目录）

## API接口

### WebSocket
- `ws://localhost:8888/ws` - WebSocket连接端点

### HTTP API
- `GET /` - 主页
- `POST /api/scan` - 启动扫描
- `GET /api/results` - 获取结果列表
- `GET /api/result/{filename}` - 获取单个结果详情

## 消息格式

### WebSocket消息

**进度消息**
```json
{
  "type": "progress",
  "message": "正在扫描目标 50/95",
  "data": {
    "current": 50,
    "total": 95,
    "percentage": 52.6,
    "alive_hosts": 25,
    "open_ports": 50
  }
}
```

**完成消息**
```json
{
  "type": "complete",
  "message": "扫描完成！",
  "data": {
    "total": 95,
    "alive": 51,
    "open_ports": 56,
    "duration": "5.2s",
    "result_file": "scan_result_20260107_185446.json"
  }
}
```

## 目录结构

```
web/
├── server.go           # Web服务器主程序
├── go.mod             # Go模块依赖
├── go.sum             # 依赖校验
├── templates/
│   └── index.html     # 主页HTML模板
├── webserver          # 编译后的可执行文件
└── README.md          # 说明文档
```

## 特色功能

### 🎨 UI/UX特性
- 渐变背景和卡片设计
- 平滑动画和过渡效果
- 响应式布局（支持移动端）
- 实时进度条和统计数字
- 彩色日志分类（信息/成功/警告/错误）

### 🔧 技术特性
- WebSocket自动重连
- 并发安全的客户端管理
- 嵌入式文件系统（embed）
- 跨域支持（CORS）
- 优雅的错误处理

## 停止服务

```bash
# 查找进程
ps aux | grep webserver

# 停止服务
kill <PID>
```

## 端口配置

默认端口：`8888`

如需修改，编辑`server.go`中的端口配置：
```go
port := ":8888"  // 修改为其他端口
```

## 安全注意事项

1. 仅在本地或受信任网络使用
2. 生产环境需添加认证机制
3. 建议配置防火墙规则
4. 定期清理历史扫描结果

## 未来计划

- [ ] 用户认证和权限管理
- [ ] 扫描任务队列
- [ ] 结果可视化图表
- [ ] 导出多种格式（PDF/Excel）
- [ ] 扫描策略模板
- [ ] API限流和防护

## License

MIT
