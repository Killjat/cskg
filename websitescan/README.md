# 网站扫描工具

一个功能强大的网站扫描工具，可以批量扫描网站，获取网站的基本信息、技术栈、ICP备案号等，并识别登录框。支持FOFA风格的WEB展示界面。

## 功能特性

### 核心扫描功能
- ✅ 获取网站标题和名称
- ✅ 识别网站使用的框架、服务、应用
- ✅ 检测网站使用的编程语言
- ✅ 提取网站ICP备案号
- ✅ 识别网站登录框
- ✅ 支持批量扫描多个网站
- ✅ 结果保存为CSV格式

### WEB展示功能（FOFA风格）
- 📊 **统计面板**：显示总网站数、框架数量、语言数量、含登录框网站数量
- 🔍 **搜索功能**：支持按URL、标题、技术栈进行搜索
- 🎯 **筛选功能**：支持按框架、语言、登录框进行筛选
- 📋 **详情查看**：点击"详情"按钮查看网站完整信息
- 📥 **结果导出**：支持将筛选后的结果导出为CSV
- 📱 **响应式设计**：适配不同屏幕尺寸

## 安装依赖

```bash
pip3 install -r requirements.txt
```

## 使用方法

### 1. 准备目标文件

创建一个包含网站URL的文本文件，每行一个URL，支持http://和https://前缀，也可以省略前缀（自动添加https://）。

示例 `targets.txt`：
```
https://www.baidu.com
https://www.example.com
bing.com
```

### 2. 运行扫描

```bash
python3 main.py -t targets.txt -o scan_results.csv
```

### 3. 启动WEB展示服务

```bash
python3 web/server.py
```

### 4. 访问WEB界面

打开浏览器访问：http://localhost:8080

### 参数说明

- `-t, --targets`: 包含网站URL的文件路径（必填）
- `-o, --output`: 输出CSV文件路径（默认：scan_results.csv）

## 输出结果

### CSV输出

扫描结果将保存为CSV文件，包含以下字段：

| 字段名 | 描述 |
|--------|------|
| url | 网站URL |
| title | 网站标题 |
| site_name | 网站名称 |
| frameworks | 使用的框架 |
| services | 使用的服务 |
| applications | 使用的应用 |
| programming_languages | 使用的编程语言 |
| icp | ICP备案号 |
| has_login_form | 是否有登录框 |
| error | 扫描错误信息（如果有） |

### WEB界面

WEB界面采用FOFA风格设计，包含：
- 搜索栏：支持关键词搜索
- 筛选栏：支持框架、语言、登录框筛选
- 统计卡片：显示扫描结果的统计信息
- 结果表格：以表格形式展示扫描结果
- 详情模态框：点击"详情"查看完整信息

## 技术栈

### 后端
- Python 3
- Requests - HTTP请求
- BeautifulSoup4 - HTML解析
- python-wappalyzer - 技术栈识别
- LXML - 高效HTML解析

### WEB服务
- Flask - Web框架
- HTML5 + CSS3 + JavaScript - 前端界面

## 项目结构

```
websitescan/
├── web/
│   ├── templates/
│   │   └── index.html    # WEB展示模板
│   └── server.py         # Flask WEB服务
├── main.py              # 主程序入口
├── scanner.py           # 扫描核心功能
├── utils.py             # 工具函数
├── requirements.txt     # 依赖列表
├── README.md            # 项目说明
├── targets.txt          # 示例目标文件
└── scan_results.csv     # 扫描结果（自动生成）
```

## 注意事项

1. 请遵守相关法律法规，仅扫描您有权限扫描的网站
2. 扫描过程中会发送HTTP请求，请合理控制扫描频率
3. 部分网站可能有反爬机制，可能导致扫描失败
4. ICP备案号提取基于正则表达式，可能存在识别不准确的情况
5. WEB服务默认运行在8080端口，如被占用请修改server.py文件中的端口设置

## 示例

### 扫描示例目标

```bash
# 扫描示例目标文件
python3 main.py -t targets.txt

# 指定输出文件
python3 main.py -t targets.txt -o my_results.csv
```

### 启动WEB服务

```bash
# 启动WEB服务
python3 web/server.py

# 访问WEB界面
# http://localhost:8080
```

## WEB功能使用技巧

1. **搜索功能**：在搜索框中输入关键词，可以搜索网站的URL、标题、技术栈等信息
2. **筛选功能**：
   - 框架筛选：选择特定框架，只显示使用该框架的网站
   - 语言筛选：选择特定语言，只显示使用该语言的网站
   - 登录框筛选：选择"有"或"无"，筛选包含或不包含登录框的网站
3. **详情查看**：点击每条结果后的"详情"按钮，可以查看网站的完整信息
4. **结果导出**：点击"导出CSV"按钮，可以将当前筛选结果导出为CSV文件
5. **分页浏览**：结果较多时，使用分页功能浏览更多结果

## 浏览器兼容性

- Chrome 80+
- Firefox 75+
- Safari 13+
- Edge 80+

## 常见问题

### Q: 扫描速度太慢？
A: 扫描过程中会自动添加1秒的延迟，避免请求过快被封IP。可以修改main.py文件中的time.sleep(1)来调整延迟时间。

### Q: WEB服务无法访问？
A: 请检查端口是否被占用，默认端口为8080。可以修改web/server.py文件中的port参数来更换端口。

### Q: 扫描结果中ICP备案号为空？
A: ICP备案号提取基于正则表达式，部分网站可能使用了不同的格式或位置，导致提取失败。可以查看网站源码，手动验证ICP备案号。

### Q: 登录框识别不准确？
A: 登录框识别基于HTML结构和关键词匹配，部分网站的登录框可能使用了特殊的设计，导致识别失败。可以查看网站源码，手动验证是否有登录框。
