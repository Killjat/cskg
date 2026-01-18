# 网页克隆工具 - 高可用版本

## 功能介绍
- 网页克隆：支持克隆网页的标题、头部、正文等内容
- 多种克隆格式：简化版、完整版、信息JSON
- 高可用：基于FastAPI + Uvicorn，支持epoll/kqueue高效I/O多路复用
- 自动IP检测：返回可外部访问的URL
- 随机端口：避免端口冲突

## 部署方式

### 在CentOS上部署

1. 将压缩包上传到CentOS服务器
2. 解压压缩包：
   ```bash
   tar -zxvf web-clone-tool-*.tar.gz
   cd web-clone-tool
   ```
3. 运行部署脚本：
   ```bash
   sudo ./deploy_centos.sh
   ```

### 手动启动（开发环境）

```bash
# 安装依赖
pip3 install -r requirements.txt

# 启动服务
python3 app.py
```

## 使用说明

### 通过API使用

1. 克隆网页：
   ```bash
   curl -X POST -H "Content-Type: application/json" -d '{"url": "https://example.com"}' http://<server-ip>:<port>/api/clone
   ```

2. 查看已克隆页面：
   ```bash
   curl http://<server-ip>:<port>/api/cloned-pages
   ```

3. 查看服务器信息：
   ```bash
   curl http://<server-ip>:<port>/api/info
   ```

### 通过客户端脚本使用

```bash
python3 send_clone_request.py http://<server-ip>:<port> https://example.com
```

## 服务管理

- 查看状态：`systemctl status web-clone-tool.service`
- 启动服务：`systemctl start web-clone-tool.service`
- 停止服务：`systemctl stop web-clone-tool.service`
- 重启服务：`systemctl restart web-clone-tool.service`
- 查看日志：`journalctl -u web-clone-tool.service -f`

## 技术栈
- Python 3.10+
- FastAPI
- Uvicorn
- BeautifulSoup4
- Requests
