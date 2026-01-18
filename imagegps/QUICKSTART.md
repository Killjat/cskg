# 🚀 快速开始指南

## 一键启动

### macOS/Linux

```bash
cd imagegps
./start.sh
```

### Windows

```bash
cd imagegps
go mod tidy
go run main.go
```

## 访问Web界面

服务启动后，打开浏览器访问：

```
http://localhost:8080
```

## 使用API

### 1. 健康检查

```bash
curl http://localhost:8080/api/health
```

### 2. 上传图片提取GPS

```bash
curl -X POST \
  -F "image=@/path/to/your/photo.jpg" \
  http://localhost:8080/api/upload
```

**响应示例**：

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
    "baidu_map_url": "https://api.map.baidu.com/marker?location=25.033976,121.564472"
  }
}
```

## Python集成示例

```python
import requests

def extract_gps_from_image(image_path):
    url = 'http://localhost:8080/api/upload'
    files = {'image': open(image_path, 'rb')}
    
    response = requests.post(url, files=files)
    result = response.json()
    
    if result['success']:
        gps = result['data']
        print(f"位置: {gps['latitude']}, {gps['longitude']}")
        print(f"Google地图: {gps['google_map_url']}")
    else:
        print(f"提取失败: {result['message']}")

# 使用示例
extract_gps_from_image('/path/to/photo.jpg')
```

## JavaScript集成示例

```javascript
async function uploadImage(file) {
    const formData = new FormData();
    formData.append('image', file);
    
    const response = await fetch('http://localhost:8080/api/upload', {
        method: 'POST',
        body: formData
    });
    
    const result = await response.json();
    
    if (result.success) {
        console.log('GPS信息:', result.data);
        console.log('Google地图:', result.data.google_map_url);
    } else {
        console.log('提取失败:', result.message);
    }
}
```

## 测试图片要求

✅ **可用的图片**：
- 手机拍摄的原始照片（iPhone、Android等）
- 开启了GPS定位的相机拍摄的照片
- 未经处理、保留完整EXIF信息的图片

❌ **不可用的图片**：
- 社交平台下载的图片（已删除EXIF）
- 截图
- 经过图片编辑软件处理的图片
- 扫描的照片

## 端口配置

如需修改默认端口（8080），编辑 `main.go` 文件：

```go
port := ":8080"  // 改为你想要的端口，如 ":9000"
```

## 故障排查

### 问题1：依赖下载失败

```bash
# 设置Go代理
export GOPROXY=https://goproxy.cn,direct
go mod tidy
```

### 问题2：端口被占用

```bash
# 查看端口占用
lsof -i :8080

# 或修改main.go中的端口号
```

### 问题3：图片无GPS信息

确认图片满足以下条件：
1. 拍摄时设备GPS已开启
2. 图片未经处理
3. 格式支持（JPG、PNG、TIFF）

## 集成到CSKG主项目

如需将此模块集成到CSKG主项目，可以：

1. 在主项目的 `main.go` 中导入此模块
2. 或作为独立微服务运行
3. 使用反向代理（Nginx）统一管理所有模块

## 下一步

- 📖 查看完整文档：[README.md](README.md)
- 🧪 运行API测试：`./test_api.sh`
- 🌐 访问Web界面：http://localhost:8080
