# 自进化Wappalyzer系统技术文档

## 1. 系统概述

自进化Wappalyzer系统是一个基于机器学习的网站技术识别系统，能够自动学习和更新技术指纹，无需人工干预。系统通过智能收集网站数据、分析网站技术、训练机器学习模型、更新检测规则，实现持续进化和改进。

### 1.1 核心功能

- **智能目标获取**：从多种来源获取网站目标，包括本地文件、行业网站列表、Alexa排名等
- **网站技术识别**：使用基于规则的检测和机器学习预测相结合的方式识别网站技术
- **机器学习训练**：自动收集训练数据，训练和更新机器学习模型
- **规则自动更新**：基于检测结果和机器学习预测，自动更新检测规则
- **Web仪表盘**：提供直观的Web界面，展示统计信息和指纹信息
- **自动化执行**：无需用户输入，持续运行学习周期

### 1.2 系统特点

- **自进化能力**：自动学习和更新技术指纹，无需人工干预
- **多源目标获取**：支持从多种来源获取网站目标
- **混合检测方法**：结合基于规则的检测和机器学习预测
- **持续运行**：支持自动化执行，持续学习和更新
- **直观的Web界面**：提供Web仪表盘，便于查看结果
- **易于部署**：提供自动化部署脚本，便于在Linux服务器上部署

## 2. 系统架构

### 2.1 整体架构

自进化Wappalyzer系统采用模块化设计，主要包含以下核心组件：

```
┌─────────────────────────────────────────────────────────────────┐
│                     自进化Wappalyzer系统                      │
├─────────────┬─────────────┬──────────────┬─────────────────────┤
│  目标生成器 │  扫描引擎   │  机器学习模块│   Web仪表盘        │
│SmartTargets │ Scan Engine │ ML Predictor │  Web Dashboard      │
├─────────────┼─────────────┼──────────────┼─────────────────────┤
│  目标获取   │  规则检测   │  模型训练    │   数据展示        │
│  目标筛选   │  特征提取   │  模型预测    │   统计分析        │
│  目标验证   │  结果整合   │  规则更新    │   数据刷新        │
└─────────────┴─────────────┴──────────────┴─────────────────────┘
         │               │               │               │
         └───────────────┴───────────────┴───────────────┘
                          │
                 ┌────────┴─────────┐
                 │     数据存储     │
                 ├─────────────────┤
                 │  技术规则库     │
                 │  训练数据集     │
                 │  扫描结果       │
                 │  日志文件       │
                 └─────────────────┘
```

### 2.2 核心组件

#### 2.2.1 智能目标生成器 (SmartTargets)

负责从多种来源获取网站目标，包括：
- 本地文件 (education_sites.json)
- 行业网站列表
- Alexa排名
- 随机生成

主要功能：
- 目标获取：从不同来源获取网站目标
- 目标验证：验证网站是否可访问
- 目标筛选：根据配置筛选符合条件的网站
- 目标扩展：当目标不足时，从其他来源补充

#### 2.2.2 扫描引擎 (Scan Engine)

负责扫描网站，识别网站技术，主要包含：
- 基于规则的检测：使用technologies.json中的规则检测网站技术
- 特征提取：从网站响应中提取特征
- 结果整合：整合检测结果，生成最终的技术识别结果

主要功能：
- 网站扫描：发送HTTP请求，获取网站响应
- 规则匹配：根据规则检测网站技术
- 特征提取：提取响应头、HTML、脚本等特征
- 结果整合：整合基于规则和机器学习的检测结果

#### 2.2.3 机器学习模块 (ML Predictor)

负责机器学习模型的训练和预测，主要包含：
- 特征提取：从网站响应中提取特征
- 模型训练：使用训练数据训练机器学习模型
- 模型预测：使用训练好的模型预测网站技术
- 规则更新：基于预测结果更新检测规则

主要功能：
- 特征提取：将网站响应转换为机器学习模型可处理的特征
- 模型训练：使用支持向量机(SVM)训练分类模型
- 模型预测：预测网站使用的技术
- 规则生成：基于预测结果生成新的检测规则

#### 2.2.4 Web仪表盘 (Web Dashboard)

提供直观的Web界面，展示系统运行情况和检测结果，主要包含：
- 统计信息卡片：展示总指纹数、机器学习所得指纹数等
- 指纹列表：以表格形式展示指纹信息
- 数据刷新：支持手动刷新数据

主要功能：
- 统计信息展示：显示系统统计数据
- 指纹列表展示：以表格形式展示指纹信息
- 数据刷新：支持手动刷新数据
- 响应式设计：适配不同屏幕尺寸

## 3. 技术栈

| 类别 | 技术/框架 | 用途 |
|------|-----------|------|
| 开发语言 | Python 3.6+ | 系统核心开发语言 |
| Web框架 | Flask | Web仪表盘开发 |
| 机器学习 | scikit-learn | 机器学习模型训练和预测 |
| 数据处理 | numpy | 数据处理和数值计算 |
| HTTP请求 | requests | 发送HTTP请求，获取网站响应 |
| 正则表达式 | re | 模式匹配和特征提取 |
| 数据存储 | JSON | 存储规则、配置、扫描结果等 |
| 日志记录 | logging | 系统日志记录 |
| 部署工具 | Docker (可选) | 容器化部署 |

## 4. 安装和部署

### 4.1 系统要求

- **操作系统**：CentOS 7+ / RHEL 7+ / Debian 9+ / Ubuntu 18.04+
- **Python版本**：3.6+
- **内存**：至少1GB
- **CPU**：至少1核
- **磁盘空间**：至少500MB

### 4.2 部署方式

#### 4.2.1 使用自动化部署脚本

```bash
# 上传部署包到服务器
# 解压部署包
tar -xzf self_evolving_wappalyzer-20260111.tar.gz

# 进入目录
cd self_evolving_wappalyzer-20260111

# 安装依赖
chmod +x install_deps.sh
./install_deps.sh

# 运行系统
chmod +x run.sh
./run.sh
```

#### 4.2.2 手动部署

```bash
# 安装系统依赖
yum install -y python3 python3-pip python3-devel gcc

# 克隆代码仓库
git clone <repository-url>
cd tranwebfinger

# 安装Python依赖
pip3 install requests scikit-learn numpy flask

# 运行系统
python3 auto_run.py
```

## 5. 使用说明

### 5.1 交互式运行

```bash
# 运行交互式脚本
./run.sh

# 选择运行模式
1. 运行集成系统（单次）
2. 运行集成系统（连续学习模式）
3. CMS指纹学习（更新CMS指纹）
4. 启动Web仪表盘
5. 运行自动化执行（推荐）
6. 退出
```

### 5.2 自动化执行

```bash
# 运行自动化执行脚本
python3 auto_run.py
```

### 5.3 访问Web仪表盘

- **访问地址**：http://localhost:5001
- **主要功能**：
  - 查看统计信息卡片
  - 查看指纹列表
  - 刷新数据

## 6. 配置选项

### 6.1 主要配置文件

#### 6.1.1 config.json

系统主配置文件，包含系统配置和扫描目标配置：

```json
{
  "self_evolving_wappalyzer": {
    "rules_file": "technologies.json",
    "timeout": 10,
    "scan_headers": true,
    "scan_html": true,
    "scan_scripts": true,
    "evolution_enabled": true,
    "confidence_threshold": 0.7,
    "log_evolution": true,
    "max_evolution_events": 100
  },
  "scan_targets": {
    "enabled": true,
    "sources": {
      "file": {
        "enabled": true,
        "file_path": "education_sites.json",
        "limit": 1000
      },
      "crawler": {
        "enabled": true,
        "max_depth": 2,
        "max_urls": 500
      },
      "alexa": {
        "enabled": false,
        "limit": 100
      },
      "tranco": {
        "enabled": false,
        "limit": 100
      }
    },
    "filters": {
      "tld": ["edu", "ac", "edu.cn", "ac.uk"],
      "min_estimated_visits": 0,
      "tech_category": [],
      "industry": ["education"],
      "country": []
    },
    "crawl_strategy": {
      "enabled": true,
      "max_concurrent": 10,
      "retry_count": 3,
      "delay_between_requests": 1,
      "respect_robots_txt": true,
      "follow_redirects": true,
      "timeout": 15
    }
  }
}
```

#### 6.1.2 technologies.json

技术检测规则文件，包含所有技术的检测规则：

```json
{
  "technologies": {
    "WordPress": {
      "name": "WordPress",
      "category": "CMS",
      "description": "Free and open-source content management system based on PHP and MySQL.",
      "website": "https://wordpress.org",
      "headers": {
        "X-Powered-By": ["WordPress"],
        "Set-Cookie": ["wordpress_", "wp-"],
        "Link": ["wp-content/", "wp-includes/"]
      },
      "html": ["<meta name=\"generator\" content=\"WordPress"]",
      "scripts": ["wp-content/", "wp-includes/"]
    }
  },
  "categories": {
    "CMS": {
      "name": "Content Management Systems",
      "priority": 1
    }
  }
}
```

## 7. API文档

### 7.1 主要API接口

#### 7.1.1 GET /api/scan_results

获取扫描结果

**请求参数**：无

**响应示例**：
```json
[
  {
    "url": "https://example.com",
    "detected_technologies": [
      {
        "name": "WordPress",
        "category": "CMS",
        "confidence": 0.95
      }
    ],
    "scan_time": "2026-01-11 20:52:39"
  }
]
```

#### 7.1.2 GET /api/technologies

获取技术指纹库

**请求参数**：无

**响应示例**：
```json
{
  "technologies": {
    "WordPress": {
      "name": "WordPress",
      "category": "CMS",
      "description": "Free and open-source content management system based on PHP and MySQL.",
      "website": "https://wordpress.org",
      "headers": {
        "X-Powered-By": ["WordPress"]
      },
      "html": ["<meta name=\"generator\" content=\"WordPress"]",
      "scripts": ["wp-content/", "wp-includes/"]
    }
  },
  "categories": {
    "CMS": {
      "name": "Content Management Systems",
      "priority": 1
    }
  }
}
```

#### 7.1.3 GET /api/new_fingerprints

获取新获得的指纹

**请求参数**：无

**响应示例**：
```json
[
  {
    "name": "React",
    "display_name": "React",
    "category": "JavaScript Frameworks",
    "description": "React (机器学习所得)",
    "keywords": ["react", "react-dom", "create-react-app"],
    "headers": 1,
    "html_patterns": 2,
    "script_patterns": 3
  }
]
```

## 8. 核心组件详细说明

### 8.1 SelfEvolvingWappalyzer类

系统核心类，负责协调各个组件，实现系统的核心功能：

#### 8.1.1 主要方法

- **__init__**：初始化系统，加载规则，初始化机器学习预测器
- **load_rules**：加载检测规则
- **save_rules**：保存检测规则
- **scan**：扫描网站，识别技术
- **extract_features_from_response**：从响应中提取特征
- **evolve**：基于预期技术更新检测规则
- **batch_scan**：批量扫描网站
- **smart_scan**：使用智能目标生成器扫描网站

### 8.2 IntegratedWappalyzerSystem类

集成系统类，负责协调智能目标生成、网站扫描、机器学习训练和规则更新：

#### 8.2.1 主要方法

- **__init__**：初始化集成系统，创建各个组件实例
- **smart_collect_and_learn**：智能收集网站数据并进行学习
- **run_continuous_learning**：连续运行学习流程
- **cms_fingerprint_learning**：专门针对CMS技术进行指纹学习

### 8.3 SmartTargetGenerator类

智能目标生成器类，负责从多种来源获取网站目标：

#### 8.3.1 主要方法

- **__init__**：初始化智能目标生成器，加载配置
- **validate_domain**：验证域名是否有效且可访问
- **get_alexa_top_sites**：从Alexa获取顶级网站
- **get_industry_websites**：根据行业获取网站
- **get_cms_websites**：获取CMS网站
- **get_targets_from_file**：从文件中读取目标网站
- **generate_targets**：智能生成扫描目标

### 8.4 MLTechnologyPredictor类

机器学习预测器类，负责机器学习模型的训练和预测：

#### 8.4.1 主要方法

- **__init__**：初始化机器学习预测器，加载模型和向量器
- **extract_features**：从网站数据中提取特征
- **train_model**：训练机器学习模型
- **predict_technology**：预测网站技术
- **update_model**：更新机器学习模型

## 9. 开发指南

### 9.1 系统结构

```
tranwebfinger/
├── main.py                # 主系统类
├── integrated_system.py   # 集成系统类
├── smart_targets.py       # 智能目标生成器
├── ml_predictor.py        # 机器学习预测器
├── web_dashboard.py       # Web仪表盘应用
├── templates/             # Web仪表盘模板
│   └── index.html        # 主页面模板
├── auto_run.py            # 自动化执行脚本
├── package.sh             # 打包脚本
├── install_deps.sh        # 依赖安装脚本
├── run.sh                 # 运行脚本
├── config.json            # 配置文件
├── technologies.json      # 技术检测规则
├── education_sites.json   # 教育网站列表
├── tech_predictor_model.pkl  # 机器学习模型
├── vectorizer.pkl         # 特征向量器
└── TECHNICAL_DOCUMENTATION.md  # 技术文档
```

### 9.2 开发流程

1. **克隆代码仓库**
2. **安装开发依赖**
3. **修改代码**
4. **测试修改**
5. **提交代码**
6. **生成部署包**

### 9.3 测试方法

- **单元测试**：对核心组件进行单元测试
- **集成测试**：测试系统的集成功能
- **功能测试**：测试系统的核心功能
- **性能测试**：测试系统的性能

## 10. 故障排除

### 10.1 常见问题

#### 10.1.1 无法获取网站

**解决方案**：
- 检查网络连接
- 检查DNS配置
- 调整smart_targets.py中的超时设置
- 确保education_sites.json文件存在且格式正确

#### 10.1.2 检测结果不准确

**解决方案**：
- 运行多次学习流程，让模型更好地学习
- 添加更多训练数据
- 手动调整technologies.json中的规则
- 调整机器学习模型的参数

#### 10.1.3 内存不足

**解决方案**：
- 减少每轮扫描的网站数量
- 减少机器学习模型的复杂度
- 增加系统内存

#### 10.1.4 Web仪表盘无法访问

**解决方案**：
- 检查Web仪表盘服务是否正在运行
- 检查防火墙设置，确保端口5001已开放
- 检查日志文件，查看是否有错误信息

## 11. 日志管理

### 11.1 主要日志文件

- **integrated_system.log**：集成系统日志，记录系统运行情况
- **smart_targets.log**：智能目标生成器日志，记录目标获取情况
- **ml_predictor.log**：机器学习预测器日志，记录模型训练和预测情况
- **auto_run.log**：自动化执行日志，记录自动化执行情况

### 11.2 日志查看方法

```bash
# 查看最新的日志
tail -f integrated_system.log

# 查看指定数量的日志
tail -n 100 integrated_system.log

# 搜索关键字
grep "error" integrated_system.log
```

## 12. 未来 roadmap

### 12.1 短期计划

- [ ] 支持更多的目标来源
- [ ] 优化机器学习模型，提高预测准确性
- [ ] 增加更多的技术类别和行业支持
- [ ] 实现更详细的报告生成功能
- [ ] 支持多语言

### 12.2 长期计划

- [ ] 支持分布式部署
- [ ] 实现实时监控和告警功能
- [ ] 支持更复杂的技术关系分析
- [ ] 实现自动漏洞检测功能
- [ ] 支持更多的机器学习算法
- [ ] 实现自动规则优化功能

## 13. 联系方式

- **项目负责人**：[负责人姓名]
- **技术支持**：[技术支持邮箱]
- **GitHub仓库**：[GitHub仓库地址]
- **文档地址**：[文档地址]

## 14. 版本信息

| 版本 | 发布日期 | 主要更新 |
|------|----------|----------|
| 1.0.0 | 2026-01-11 | 初始版本，包含核心功能 |

## 15. 许可证

[许可证信息]

---

**自进化Wappalyzer系统技术文档**
**版本：1.0.0**
**日期：2026-01-11**
**作者：AI Assistant**
