# ICP备案号查询工具使用指南

## 工具说明

本目录包含三个主要工具，用于查询和验证ICP备案信息：

### 1. `icp_verify.py` - ICP验证工具（推荐）
**功能**: 验证指定域名是否包含特定的ICP备案号

**使用方法**:
```bash
# 验证单个域名
python3 icp_verify.py -icp=京ICP证030173号 -d=baidu.com

# 批量验证（从文件读取域名列表）
python3 icp_verify.py -icp=闽ICP备06031865号 -f=domains.txt -o=result.csv
```

**参数说明**:
- `-icp`: ICP备案号（必需）
- `-d`: 单个域名
- `-f`: 域名列表文件（每行一个域名）
- `-o`: 输出CSV文件名（默认：verified_icp.csv）
- `-t`: 超时时间（秒，默认：10）

---

### 2. `get_domains_from_icp.py` - 域名获取工具
**功能**: 尝试从多个ICP查询网站获取与备案号相关的域名列表

**使用方法**:
```bash
python3 get_domains_from_icp.py 闽ICP备06031865号
```

**输出**: 
- 屏幕显示找到的域名
- 保存到 `domains_found.txt` 文件

---

### 3. `fromicpgeturl.py` - 原始搜索工具（不推荐）
此工具从搜索引擎结果提取URL，但准确度较低，不推荐使用。

---

## 完整工作流程

### 方案A: 已知域名列表（最可靠）

如果你已经有候选域名列表：

1. 创建域名列表文件 `my_domains.txt`:
```
example.com
test.com
demo.cn
```

2. 运行验证：
```bash
python3 icp_verify.py -icp=你的ICP备案号 -f=my_domains.txt -o=verified.csv
```

3. 查看结果文件 `verified.csv`

---

### 方案B: 自动获取域名（可能不完整）

1. 先尝试获取域名列表：
```bash
python3 get_domains_from_icp.py 闽ICP备06031865号
```

2. 使用获取的域名列表进行验证：
```bash
python3 icp_verify.py -icp=闽ICP备06031865号 -f=domains_found.txt -o=verified.csv
```

---

### 方案C: 手动查询（最准确）

1. **访问工信部备案查询系统**（最权威）:
   - URL: https://beian.miit.gov.cn/
   - 直接输入ICP备案号查询

2. **访问第三方查询工具**:
   - 站长工具: https://icp.chinaz.com/
   - 爱站网: https://icp.aizhan.com/
   - 天眼查: https://www.tianyancha.com/

3. 手动整理域名列表，保存到文本文件

4. 使用 `icp_verify.py` 进行验证

---

## 示例

### 示例1: 验证百度的ICP备案
```bash
python3 icp_verify.py -icp=京ICP证030173号 -d=baidu.com
```

**输出**:
```
✓ 找到匹配! https://www.baidu.com/ 包含 京ICP证030173号
```

### 示例2: 批量验证多个域名
创建文件 `test_domains.txt`:
```
baidu.com
qq.com
taobao.com
```

运行:
```bash
python3 icp_verify.py -icp=某ICP备案号 -f=test_domains.txt -o=result.csv
```

---

## 常见问题

### Q1: 为什么查不到域名？
**A**: 可能的原因：
1. ICP备案号已过期或注销
2. 网站已关闭
3. 网站页面底部未显示备案号
4. 网站使用JavaScript动态加载备案号

### Q2: 如何提高查询成功率？
**A**: 建议：
1. 使用官方渠道（工信部网站）查询
2. 结合多个第三方查询工具
3. 手动整理域名列表后再验证

### Q3: 验证工具显示"未找到备案号"但网站确实有？
**A**: 可能原因：
1. 备案号在JavaScript中或iframe中
2. 网站使用了防爬虫机制
3. 备案号格式变体（如"ICP备"vs"icp备"）

---

## 技术说明

### icp_verify.py 验证逻辑
1. 尝试4种URL格式: https/http, www/非www
2. 获取页面HTML内容
3. 搜索多种备案号格式变体：
   - 原始格式（如：京ICP证030173号）
   - 去掉"号"（京ICP证030173）
   - 小写格式
   - ICP备/icp备 变体

### 匹配模式
- 精确匹配
- 大小写不敏感
- 支持格式变体

---

## 输出文件格式

CSV文件包含以下字段：
- 域名
- ICP备案号
- 网站URL（实际访问的URL，包含重定向）
- 状态
- 匹配方式

---

## 建议

1. **优先使用官方查询**：工信部备案系统是最权威的来源
2. **批量验证前先测试**：用一个已知域名测试工具是否正常工作
3. **合理设置超时**：某些网站响应慢，可以增加超时时间
4. **遵守法律法规**：查询ICP信息请遵守相关法律法规

---

## 联系与反馈

如有问题或建议，请提供：
- ICP备案号示例
- 预期结果
- 实际输出
