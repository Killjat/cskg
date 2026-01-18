#!/bin/bash

# 打包脚本，用于将整个项目打包成一个压缩文件

# 项目名称
PROJECT_NAME="self_evolving_wappalyzer"

# 打包目录
PACK_DIR="$PROJECT_NAME-$(date +%Y%m%d)"

# 创建打包目录
mkdir -p $PACK_DIR

# 复制必要文件
cp -r *.py $PACK_DIR/
cp -r *.json $PACK_DIR/
cp README.md $PACK_DIR/
cp TECHNICAL_DOCUMENTATION.md $PACK_DIR/

# 复制模板目录（用于Web仪表盘）
mkdir -p $PACK_DIR/templates
cp -r templates/* $PACK_DIR/templates/

# 复制模型文件（如果存在）
if [ -f "tech_predictor_model.pkl" ]; then
    cp tech_predictor_model.pkl $PACK_DIR/
fi
if [ -f "vectorizer.pkl" ]; then
    cp vectorizer.pkl $PACK_DIR/
fi

# 复制日志文件（如果存在）
cp -r *.log $PACK_DIR/ 2>/dev/null || true

# 创建依赖安装脚本
cat > $PACK_DIR/install_deps.sh << 'EOF'
#!/bin/bash

# 安装依赖

echo "开始安装依赖..."

# 检查Python版本
python3 --version

# 安装pip依赖
echo "安装Python依赖..."
pip3 install requests scikit-learn numpy flask

echo "依赖安装完成！"
EOF

# 创建运行脚本
cat > $PACK_DIR/run.sh << 'EOF'
#!/bin/bash

# 运行脚本

echo "=== 自进化Wappalyzer系统 ==="
echo "1. 运行集成系统（单次）"
echo "2. 运行集成系统（连续学习模式）"
echo "3. CMS指纹学习（更新CMS指纹）"
echo "4. 启动Web仪表盘"
echo "5. 运行自动化执行（推荐）"
echo "6. 退出"

echo -n "请选择操作："
read choice

case $choice in
    1)
        echo "运行集成系统（单次）..."
        python3 -c "from integrated_system import IntegratedWappalyzerSystem; integrated_system = IntegratedWappalyzerSystem(); integrated_system.smart_collect_and_learn(10)"
        ;;
    2)
        echo "运行集成系统（连续学习模式）..."
        python3 integrated_system.py
        ;;
    3)
        echo "运行CMS指纹学习（更新CMS指纹）..."
        python3 -c "from integrated_system import IntegratedWappalyzerSystem; integrated_system = IntegratedWappalyzerSystem(); integrated_system.cms_fingerprint_learning(10, 2)"
        ;;
    4)
        echo "启动Web仪表盘..."
        echo "Web仪表盘将在 http://localhost:5001 上运行"
        echo "按 Ctrl+C 停止服务器"
        python3 web_dashboard.py
        ;;
    5)
        echo "启动自动化执行..."
        echo "自动化执行将持续运行，无需用户输入"
        echo "Web仪表盘将自动启动，访问地址：http://localhost:5001"
        echo "按 Ctrl+C 停止服务器"
        python3 auto_run.py
        ;;
    6)
        echo "退出"
        exit 0
        ;;
    *)
        echo "无效选择"
        exit 1
        ;;
esac
EOF

# 给脚本添加执行权限
chmod +x $PACK_DIR/install_deps.sh
chmod +x $PACK_DIR/run.sh

# 创建CentOS运行说明
cat > $PACK_DIR/CENTOS_RUN.md << 'EOF'
# 在CentOS上运行自进化Wappalyzer系统

## 系统要求
- CentOS 7+ 或 Rocky Linux 8+
- Python 3.6+
- 至少1GB内存

## 安装步骤

1. **安装Python**
   ```bash
   yum install -y python3 python3-pip
   ```

2. **安装依赖**
   ```bash
   chmod +x install_deps.sh
   ./install_deps.sh
   ```

3. **运行系统**
   ```bash
   chmod +x run.sh
   ./run.sh
   ```

## 运行选项

### 选项1：单次运行集成系统
执行一次完整的学习流程：
- 智能获取网站
- 扫描识别技术
- 训练机器学习模型
- 更新规则

### 选项2：连续学习模式
持续运行学习流程（默认3轮，每轮5个网站）

### 选项3：CMS指纹学习
专门针对CMS技术进行指纹学习和更新

### 选项4：启动Web仪表盘
启动Web仪表盘，用于展示新获得的指纹信息：
- 访问地址：http://localhost:5001
- 展示统计信息卡片
- 以表格形式展示指纹信息
- 支持数据刷新

## 项目文件说明

| 文件 | 功能 |
|------|------|
| main.py | 主Wappalyzer系统 |
| integrated_system.py | 集成系统控制器 |
| smart_targets.py | 智能目标生成器 |
| ml_predictor.py | 机器学习预测器 |
| web_dashboard.py | Web仪表盘应用 |
| templates/index.html | Web仪表盘HTML模板 |
| technologies.json | 技术检测规则 |
| config.json | 系统配置 |
| education_sites.json | 教育网站列表 |
| scan_results.json | 扫描结果 |
| merge_rules.py | 规则合并工具 |
| update_rules.py | 规则更新工具 |
| inspect_baidu.py | 百度网站检查工具 |
| generate_education_sites.py | 教育网站生成器 |
| demo_url_recognition.py | URL识别演示 |

## 自定义配置

### 修改连续学习参数
编辑 `integrated_system.py` 文件，修改以下行：
```python
# 运行连续学习模式
# 参数1：学习轮数
# 参数2：每轮获取的网站数量
integrated_system.run_continuous_learning(3, 5)
```

### 修改智能目标生成器配置
编辑 `smart_targets.py` 文件，修改 `__init__` 方法中的配置

### 修改检测规则
编辑 `technologies.json` 文件，添加或修改检测规则

## 日志文件

系统生成以下日志文件：
- `integrated_system.log`：集成系统日志
- `smart_targets.log`：智能目标生成器日志
- `ml_predictor.log`：机器学习预测器日志

## 故障排除

### 问题：无法获取网站
解决方案：
- 检查网络连接
- 检查DNS配置
- 调整 `smart_targets.py` 中的超时设置

### 问题：检测结果不准确
解决方案：
- 运行多次学习流程，让模型更好地学习
- 添加更多训练数据
- 手动调整 `technologies.json` 中的规则

### 问题：内存不足
解决方案：
- 减少每轮扫描的网站数量
- 减少机器学习模型的复杂度
- 增加系统内存
EOF

# 打包成tar.gz文件
tar -czf $PACK_DIR.tar.gz $PACK_DIR

# 清理临时目录
rm -rf $PACK_DIR

echo "打包完成！"
echo "打包文件：$PACK_DIR.tar.gz"
echo "使用方法："
echo "1. 将压缩包上传到CentOS服务器"
echo "2. 解压：tar -xzf $PACK_DIR.tar.gz"
echo "3. 进入目录：cd $PACK_DIR"
echo "4. 安装依赖：./install_deps.sh"
echo "5. 运行系统：./run.sh"
