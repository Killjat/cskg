#!/bin/bash

# 网页克隆工具打包脚本
# 用于将应用打包为压缩包，方便在CentOS上部署

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}====================================="
echo -e "    网页克隆工具打包脚本"
echo -e "=====================================${NC}"

# 定义打包目录和文件名
PACKAGE_NAME="web-clone-tool-$(date +%Y%m%d_%H%M%S).tar.gz"
TMP_DIR="web-clone-tool"

# 创建临时目录
mkdir -p ${TMP_DIR}

# 复制必要的文件
echo -e "${YELLOW}复制必要的文件...${NC}"

# 核心应用文件
cp app.py ${TMP_DIR}/
cp send_clone_request.py ${TMP_DIR}/
cp requirements.txt ${TMP_DIR}/

# 部署脚本
cp deploy_centos.sh ${TMP_DIR}/

# 设置执行权限
chmod +x ${TMP_DIR}/deploy_centos.sh
chmod +x ${TMP_DIR}/app.py
chmod +x ${TMP_DIR}/send_clone_request.py

# 创建README文件
echo -e "${YELLOW}创建README文件...${NC}"
cat > ${TMP_DIR}/README.md << EOF
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
   tar -zxvf ${PACKAGE_NAME}
   cd ${TMP_DIR}
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
EOF

# 打包文件
echo -e "${YELLOW}打包文件...${NC}"
tar -zcvf ${PACKAGE_NAME} ${TMP_DIR}/

# 清理临时目录
rm -rf ${TMP_DIR}

echo -e "${GREEN}====================================="
echo -e "    打包完成！"
echo -e "=====================================${NC}"
echo -e "${GREEN}打包文件:${NC} ${PACKAGE_NAME}"
echo -e "${GREEN}文件大小:${NC} $(du -h ${PACKAGE_NAME} | cut -f1)"
echo -e "${GREEN}包含文件:${NC}"
echo -e "  - app.py (主应用)",
echo -e "  - send_clone_request.py (客户端脚本)",
echo -e "  - requirements.txt (依赖列表)",
echo -e "  - deploy_centos.sh (CentOS部署脚本)",
echo -e "  - README.md (使用说明)",
echo -e "${GREEN}=====================================${NC}"
echo -e "${YELLOW}使用方法:${NC}"
echo -e "1. 将 ${PACKAGE_NAME} 上传到 CentOS 服务器"
echo -e "2. 解压: tar -zxvf ${PACKAGE_NAME}"
echo -e "3. 进入目录: cd $(basename ${PACKAGE_NAME} .tar.gz)"
echo -e "4. 运行部署脚本: sudo ./deploy_centos.sh"
echo -e "${GREEN}=====================================${NC}"