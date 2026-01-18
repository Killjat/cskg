# CSKG - 网络空间安全工具集

Cyber Space Knowledge Graph - 集成多种网络安全扫描和分析工具的综合平台

---

## 📦 项目模块

### 1. 🌐 cyberspacescan - 网络空间扫描
网络空间资产发现和漏洞扫描工具

**位置**: `cyberspacescan/`  
**语言**: Go, Python  
**功能**: 端口扫描、服务识别、指纹识别

---

### 2. 🔍 assetdiscovery - 资产发现
自动化资产发现和管理系统

**位置**: `assetdiscovery/`  
**语言**: Go  
**功能**: 域名发现、子域名枚举、资产管理

---

### 3. 💻 websitescan - 网站扫描
Web应用安全扫描工具

**位置**: `websitescan/`  
**语言**: Python  
**功能**: 网站存活检测、HTTP信息收集、漏洞扫描

---

### 4. 📡 ipalive - IP存活检测
快速批量IP存活性检测工具

**位置**: `ipalive/`  
**语言**: Go, Python  
**功能**: ICMP检测、TCP检测、批量扫描

---

### 5. 🏠 lan_scanner - 局域网扫描器
局域网设备发现和端口扫描

**位置**: `lan_scanner/`  
**语言**: Python  
**功能**: 局域网设备发现、端口扫描、Web管理界面

---

### 6. 📍 imagegps - 图片GPS定位提取 ⭐ NEW
从图片EXIF中提取GPS地理位置信息

**位置**: `imagegps/`  
**语言**: Go  
**启动**: `cd imagegps && ./start.sh`  
**访问**: http://localhost:8080

#### 核心功能
- ✅ **GPS提取**: 从图片EXIF提取经纬度、海拔
- 📍 **精确定位**: 经纬度精确到6位小数
- 🗺️ **地图集成**: 自动生成Google/百度地图链接
- 📸 **设备信息**: 显示拍摄时间、相机型号
- 🌐 **Web界面**: 美观的拖拽上传界面
- 🔌 **API接口**: RESTful API，易于集成

#### 快速开始
```bash
cd imagegps
./start.sh
# 访问 http://localhost:8080
```

#### API示例
```bash
curl -X POST \
  -F "image=@photo.jpg" \
  http://localhost:8080/api/upload
```

**详细文档**: [imagegps/README.md](imagegps/README.md)

---

## 🚀 快速开始

### 环境要求
- Go 1.21+
- Python 3.8+
- 根据各模块需求安装依赖

### 通用启动流程

#### Go模块
```bash
cd [模块目录]
go mod download
go run main.go
```

#### Python模块
```bash
cd [模块目录]
pip install -r requirements.txt
python main.py
```

---

## 📊 项目结构

```
cskg/
├── assetdiscovery/      # 资产发现模块
├── cyberspacescan/      # 网络空间扫描
├── ipalive/             # IP存活检测
├── lan_scanner/         # 局域网扫描器
├── websitescan/         # 网站扫描
├── imagegps/            # 图片GPS提取 (NEW)
├── main.go              # 主程序（台湾IP段工具）
└── README.md            # 本文件
```

---

## 🎯 应用场景

### 1. 网络安全审计
- 资产发现和管理
- 漏洞扫描和评估
- 端口和服务识别

### 2. 渗透测试
- 信息收集
- 目标探测
- 漏洞利用

### 3. 数字取证
- **图片GPS提取** (NEW): 照片拍摄位置分析
- 网络流量分析
- 资产溯源

### 4. 安全监控
- 资产持续监控
- 异常检测
- 威胁情报

---

## 🆕 最新更新

### ImageGPS 模块 (2026-01-08)
新增图片GPS地理位置提取功能，支持：
- 从手机照片提取拍摄位置
- Web界面拖拽上传
- RESTful API接口
- 自动生成地图链接

**使用方法**:
```bash
cd imagegps
./start.sh
```

访问 http://localhost:8080 即可使用Web界面

---

## 📖 文档

每个模块都包含独立的README文档：
- `assetdiscovery/README.md`
- `cyberspacescan/README.md`
- `websitescan/README.md`
- `lan_scanner/README.md`
- `imagegps/README.md` ⭐

---

## 🔧 开发指南

### 添加新模块

1. 在项目根目录创建模块文件夹
2. 遵循现有模块的目录结构
3. 添加README文档
4. 更新本文件的模块列表

### 代码规范
- Go代码遵循Go官方规范
- Python代码遵循PEP 8
- 所有模块需包含README文档

---

## 🤝 贡献

欢迎提交Issue和Pull Request！

---

## ⚠️ 免责声明

本工具仅供安全研究和合法授权测试使用。使用者应遵守当地法律法规，不得用于非法用途。

---

## 📞 联系方式

- 项目地址: `/Users/jatsmith/CodeBuddy/cskg`
- 项目团队: CSKG Project Team

---

## 📝 许可证

MIT License

---

**最后更新**: 2026-01-08  
**版本**: v1.1.0 (新增ImageGPS模块)
