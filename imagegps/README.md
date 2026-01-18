# 📍 ImageGPS - 图片GPS地理位置提取系统

## 功能简介

ImageGPS 是一个用于从图片中提取GPS地理位置信息的工具。它可以读取图片EXIF数据中的GPS信息，并显示拍摄位置的经纬度、海拔、拍摄时间等详细信息。

## 核心功能

- ✅ **GPS信息提取**：自动从图片EXIF数据中提取GPS坐标
- 📍 **精确定位**：显示纬度、经度、海拔信息（精确到6位小数）
- 🗺️ **地图集成**：生成Google地图和百度地图链接，一键查看位置
- 📸 **设备信息**：显示拍摄设备、相机型号、拍摄时间
- 🌐 **Web界面**：美观的Web界面，支持拖拽上传
- 📤 **API接口**：提供RESTful API，方便集成到其他系统

## 技术栈

- **后端**：Go 1.21+ 
- **Web框架**：Gin
- **EXIF解析**：goexif
- **前端**：原生HTML/CSS/JavaScript

## 快速开始

### 1. 安装依赖

```bash
cd imagegps
go mod download
```

### 2. 启动服务

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动

### 3. 使用Web界面

打开浏览器访问 `http://localhost:8080`，上传包含GPS信息的图片即可。

## API接口

### 上传图片提取GPS信息

**接口**: `POST /api/upload`

**请求参数**:
- `image`: 图片文件 (multipart/form-data)

**支持格式**: JPG, JPEG, PNG, TIFF

**文件大小限制**: 10MB

**响应示例**:

```json
{
  "success": true,
  "message": "成功提取GPS位置信息",
  "data": {
    "latitude": 25.033976,
    "longitude": 121.564472,
    "altitude": 15.5,
    "latitude_ref": "N",
    "longitude_ref": "E",
    "has_gps": true,
    "datetime": "2024-01-08 14:30:25",
    "make": "Apple",
    "model": "iPhone 14 Pro",
    "google_map_url": "https://www.google.com/maps?q=25.033976,121.564472",
    "baidu_map_url": "https://api.map.baidu.com/marker?location=25.033976,121.564472&title=拍摄位置&content=从图片提取的位置&output=html"
  }
}
```

### 健康检查

**接口**: `GET /api/health`

**响应示例**:

```json
{
  "status": "ok",
  "service": "imagegps"
}
```

## 项目结构

```
imagegps/
├── main.go              # 主程序入口
├── go.mod               # Go依赖管理
├── handler/             # HTTP请求处理器
│   └── upload.go        # 图片上传处理
├── utils/               # 工具函数
│   └── exif.go          # EXIF/GPS信息提取
├── web/                 # Web资源
│   └── templates/       # HTML模板
│       └── index.html   # 前端页面
└── README.md            # 说明文档
```

## 使用场景

1. **数字取证**：分析图片拍摄位置
2. **旅游记录**：整理旅行照片的地理位置
3. **安全审计**：检查图片是否泄露位置信息
4. **数据分析**：批量提取照片的地理信息进行分析

## 注意事项

⚠️ **隐私提醒**：
- 图片中的GPS信息可能暴露您的位置隐私
- 在社交媒体分享照片前，建议先删除EXIF信息
- 本工具仅在本地处理，不会上传或保存您的图片

📝 **技术说明**：
- 只能提取包含EXIF/GPS信息的图片
- 经过社交平台处理的图片通常已删除EXIF信息
- 截图和编辑过的图片通常不包含GPS信息

## 编译部署

### 编译为可执行文件

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o imagegps-linux

# Windows
GOOS=windows GOARCH=amd64 go build -o imagegps.exe

# macOS
GOOS=darwin GOARCH=amd64 go build -o imagegps-mac
```

### Docker部署（可选）

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o imagegps

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/imagegps .
COPY --from=builder /app/web ./web
EXPOSE 8080
CMD ["./imagegps"]
```

## 后续扩展

- [ ] 批量处理多张图片
- [ ] 导出GPS信息为CSV/Excel
- [ ] 在地图上显示多个照片的位置
- [ ] 支持视频文件的GPS提取
- [ ] 添加图片压缩和EXIF清除功能

## 开源协议

MIT License

## 作者

CSKG Project Team
