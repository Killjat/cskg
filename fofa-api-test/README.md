# FOFA API测试工具

这是一个独立的测试程序，使用FOFA API获取100个URL并测试API提取功能。

## 功能特性

- 🔍 **FOFA集成** - 使用FOFA API搜索目标网站
- 🕷️ **API提取** - 从网页中提取各种类型的API端点
- 📊 **详细统计** - 生成完整的测试报告和统计信息
- 💾 **结果保存** - 自动保存JSON格式的详细测试结果

## API提取能力

### JavaScript API调用
- ✅ **Fetch API** - `fetch()` 调用
- ✅ **Axios** - axios HTTP客户端调用
- ✅ **jQuery AJAX** - `$.ajax()`, `$.get()`, `$.post()` 等
- ✅ **XMLHttpRequest** - 原生XHR调用
- ✅ **WebSocket** - WebSocket连接

### API模式识别
- ✅ **REST API路径** - `/api/`, `/v1/`, `/v2/` 等
- ✅ **GraphQL端点** - GraphQL查询端点
- ✅ **JSON端点** - `.json` 文件
- ✅ **相对路径API** - 以 `/` 开头的API路径

## 使用方法

### 1. 构建和运行
```bash
cd fofa-api-test
go run main.go
```

### 2. 配置文件
程序需要FOFA配置文件才能运行：

```bash
# 复制示例配置文件
cp fofa_config.json.example fofa_config.json

# 编辑配置文件，填入你的FOFA凭据
{
  "email": "your_email@example.com",
  "key": "your_fofa_api_key_here", 
  "base_url": "https://fofa.info/api/v1/search/all"
}
```

**注意**: `fofa_config.json` 文件包含敏感信息，已被 `.gitignore` 忽略，不会上传到git仓库。

### 3. 测试流程
1. **目标获取** - 使用多个FOFA查询获取不同类型的网站
2. **API提取** - 对每个网站进行API端点提取
3. **结果统计** - 生成详细的统计报告
4. **数据保存** - 保存JSON格式的完整结果

## 测试查询

程序使用以下FOFA查询来获取测试目标：

```
1. title="API" && country="CN"           # 标题包含API的中国网站
2. body="/api/" && country="CN"          # 页面包含/api/路径的网站  
3. header="application/json" && country="CN"  # 返回JSON的网站
4. body="axios" && country="CN"          # 使用axios的网站
5. body="fetch(" && country="CN"         # 使用fetch的网站
```

每个查询获取20个结果，总共最多100个唯一URL进行测试。

## 输出报告

### 控制台输出
- 实时显示测试进度
- 每个网站的测试结果
- 发现的API端点预览
- 最终统计报告

### JSON报告文件
自动生成 `fofa_api_test_result_YYYYMMDD_HHMMSS.json` 文件，包含：

```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "total_tests": 100,
  "success_count": 85,
  "success_rate": 85.0,
  "total_apis": 342,
  "source_stats": {
    "fetch": 156,
    "axios": 89,
    "jquery": 45,
    "pattern": 52
  },
  "type_stats": {
    "REST": 298,
    "GraphQL": 12,
    "WebSocket": 32
  },
  "method_stats": {
    "GET": 234,
    "POST": 78,
    "PUT": 23,
    "DELETE": 7
  },
  "results": [...]
}
```

## 统计信息

程序会生成以下统计信息：

### 基本统计
- 总测试数量
- 成功数量和成功率
- 总API数量
- 平均每站API数

### 详细分析
- **API来源统计** - 按提取方式分类
- **API类型统计** - REST/GraphQL/WebSocket
- **HTTP方法统计** - GET/POST/PUT/DELETE等
- **最佳结果展示** - API发现最多的网站

## 示例输出

```
🚀 开始FOFA API测试...
============================================================
📡 正在从FOFA获取目标URL...
  查询 1: title="API" && country="CN"
  ✅ 获取到 20 个URL
  查询 2: body="/api/" && country="CN"  
  ✅ 获取到 20 个URL
  ...

📊 准备测试 100 个唯一URL
============================================================

[1/100] 测试: https://api.example.com
  ✅ 成功 | 状态码: 200 | API数: 15 | 响应时间: 1.2s
    - GET /api/v1/users (fetch)
    - POST /api/v1/login (axios)
    - GET /api/v1/data.json (pattern)
    ... 还有 12 个API

============================================================
📋 测试报告
============================================================
总测试数量: 100
成功数量: 85
成功率: 85.00%
总API数量: 342
平均每站API数: 4.02

📊 API来源统计:
  fetch: 156
  axios: 89
  jquery: 45
  pattern: 52

🏆 API发现最多的网站:
  https://api.example.com - 15个API
  https://admin.test.com - 12个API

💾 详细结果已保存到: fofa_api_test_result_20240101_120000.json
✅ 测试完成!
```

## 注意事项

1. **请求频率** - 程序在每次请求间添加500ms延迟，避免过于频繁的请求
2. **网络超时** - HTTP请求超时设置为30秒
3. **错误处理** - 对网络错误和解析错误进行了完善的处理
4. **数据去重** - 自动去除重复的URL和API端点

这个测试工具可以帮助你验证API提取算法的效果，并获得真实网站的API发现统计数据。