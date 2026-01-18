# 自进化Wappalyzer系统部署指南

## 系统概述

自进化Wappalyzer系统是一个基于机器学习的网站技术识别系统，能够自动学习和更新技术指纹。本部署指南将帮助您在Linux服务器上部署和运行该系统。

## 系统要求

- **操作系统**：CentOS 7+ / RHEL 7+ / Debian 9+ / Ubuntu 18.04+
- **Python版本**：3.6+
- **内存**：至少1GB
- **CPU**：至少1核
- **磁盘空间**：至少500MB

## 部署方式

### 方式一：使用自动化部署脚本

1. **获取部署包**
   - 从开发环境获取打包好的部署包：`self_evolving_wappalyzer-YYYYMMDD.tar.gz`
   - 获取自动化部署脚本：`auto_deploy.sh`

2. **上传文件到服务器**
   - 将部署包和自动化部署脚本上传到服务器的任意目录，例如 `/tmp`

3. **执行自动化部署脚本**
   ```bash
   # 切换到部署包所在目录
   cd /tmp
   
   # 给脚本添加执行权限
   chmod +x auto_deploy.sh
   
   # 执行部署脚本
   ./auto_deploy.sh
   ```

4. **使用URL进行部署**（可选）
   ```bash
   # 从URL下载部署包并进行部署
   ./auto_deploy.sh "http://example.com/self_evolving_wappalyzer-YYYYMMDD.tar.gz"
   ```

### 方式二：手动部署

1. **安装系统依赖**
   ```bash
   # CentOS/RHEL
   yum install -y python3 python3-pip python3-devel gcc
   
   # Debian/Ubuntu
   apt-get install -y python3 python3-pip python3-dev gcc
   ```

2. **创建安装目录**
   ```bash
   mkdir -p /opt/self_evolving_wappalyzer
   cd /opt/self_evolving_wappalyzer
   ```

3. **上传并解压部署包**
   ```bash
   # 上传部署包到/opt/self_evolving_wappalyzer目录
   tar -xzf self_evolving_wappalyzer-YYYYMMDD.tar.gz
   mv self_evolving_wappalyzer-YYYYMMDD/* .
   rm -rf self_evolving_wappalyzer-YYYYMMDD self_evolving_wappalyzer-YYYYMMDD.tar.gz
   ```

4. **安装Python依赖**
   ```bash
   # 给脚本添加执行权限
   chmod +x install_deps.sh
   
   # 执行依赖安装脚本
   ./install_deps.sh
   ```

5. **启动系统服务**（可选）
   ```bash
   # 手动启动Web仪表盘服务
   systemctl start wappalyzer-dashboard
   
   # 设置Web仪表盘服务开机自启
   systemctl enable wappalyzer-dashboard
   ```

## 运行系统

### 交互式运行

```bash
# 切换到安装目录
cd /opt/self_evolving_wappalyzer

# 给脚本添加执行权限
chmod +x run.sh

# 执行运行脚本
./run.sh
```

### 命令行参数说明

```
=== 自进化Wappalyzer系统 ===
1. 运行集成系统（单次）
2. 运行集成系统（连续学习模式）
3. CMS指纹学习（更新CMS指纹）
4. 启动Web仪表盘
5. 退出
```

## Web仪表盘

Web仪表盘用于展示新获得的指纹信息：

- **访问地址**：http://localhost:5001
- **主要功能**：
  - 展示统计信息卡片
  - 以表格形式展示指纹信息
  - 支持数据刷新

## 服务管理

### Web仪表盘服务

```bash
# 启动服务
systemctl start wappalyzer-dashboard

# 停止服务
systemctl stop wappalyzer-dashboard

# 重启服务
systemctl restart wappalyzer-dashboard

# 查看服务状态
systemctl status wappalyzer-dashboard

# 设置开机自启
systemctl enable wappalyzer-dashboard

# 取消开机自启
systemctl disable wappalyzer-dashboard
```

## 配置说明

### 主要配置文件

- **config.json**：系统配置文件
- **technologies.json**：技术检测规则
- **education_sites.json**：教育网站列表

### 自定义配置

1. **修改连续学习参数**
   编辑 `integrated_system.py` 文件，修改以下行：
   ```python
   # 运行连续学习模式
   # 参数1：学习轮数
   # 参数2：每轮获取的网站数量
   integrated_system.run_continuous_learning(3, 5)
   ```

2. **修改Web仪表盘端口**
   编辑 `web_dashboard.py` 文件，修改以下行：
   ```python
   app.run(debug=True, host='0.0.0.0', port=5001)
   ```

## 日志管理

### 日志文件

- **integrated_system.log**：集成系统日志
- **smart_targets.log**：智能目标生成器日志
- **ml_predictor.log**：机器学习预测器日志

### 查看日志

```bash
# 查看集成系统日志
tail -f integrated_system.log

# 查看智能目标生成器日志
tail -f smart_targets.log

# 查看机器学习预测器日志
tail -f ml_predictor.log
```

## 故障排除

### 问题：无法获取网站
**解决方案**：
- 检查网络连接
- 检查DNS配置
- 调整 `smart_targets.py` 中的超时设置

### 问题：检测结果不准确
**解决方案**：
- 运行多次学习流程，让模型更好地学习
- 添加更多训练数据
- 手动调整 `technologies.json` 中的规则

### 问题：内存不足
**解决方案**：
- 减少每轮扫描的网站数量
- 减少机器学习模型的复杂度
- 增加系统内存

### 问题：Web仪表盘无法访问
**解决方案**：
- 检查Web仪表盘服务是否正在运行：`systemctl status wappalyzer-dashboard`
- 检查防火墙设置，确保端口5001已开放
- 检查日志文件，查看是否有错误信息：`tail -f integrated_system.log`

## 更新系统

1. **获取最新的部署包**
2. **上传部署包到服务器**
3. **停止当前服务**（如果正在运行）
   ```bash
   systemctl stop wappalyzer-dashboard
   ```
4. **执行自动化部署脚本**
   ```bash
   ./auto_deploy.sh
   ```
5. **启动服务**
   ```bash
   systemctl start wappalyzer-dashboard
   ```

## 卸载系统

1. **停止服务**
   ```bash
   systemctl stop wappalyzer-dashboard
   systemctl disable wappalyzer-dashboard
   ```

2. **删除服务配置文件**
   ```bash
   rm -f /etc/systemd/system/wappalyzer-dashboard.service
   systemctl daemon-reload
   ```

3. **删除安装目录**
   ```bash
   rm -rf /opt/self_evolving_wappalyzer
   ```

## 技术支持

- **项目文档**：`TECHNICAL_DOCUMENTATION.md`
- **README**：`README.md`
- **日志文件**：查看相关日志文件获取更多信息

## 版本信息

- **系统名称**：自进化Wappalyzer系统
- **版本号**：YYYYMMDD
- **部署日期**：YYYY-MM-DD
- **Python版本**：3.6+

---

**自进化Wappalyzer系统部署指南**
**版本：1.0**
**日期：2026-01-11**
