#!/bin/bash

# 交互式运行脚本

echo "=== 自进化Wappalyzer系统 ==="
echo "1. 常规运行（单次学习）"
echo "2. 连续学习模式"
echo "3. CMS指纹学习"
echo "4. 特定行业网站学习"
echo "5. 特定行业CMS指纹学习"
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
        echo "运行CMS指纹学习..."
        python3 -c "from integrated_system import IntegratedWappalyzerSystem; integrated_system = IntegratedWappalyzerSystem(); integrated_system.cms_fingerprint_learning(10, 2)"
        ;;
    4)
        echo "=== 特定行业网站学习 ==="
        echo "可用行业："
        echo "1. 电子商务 (ecommerce)"
        echo "2. 教育 (education)"
        echo "3. 金融 (finance)"
        echo "4. 医疗健康 (healthcare)"
        echo "5. 科技 (technology)"
        echo -n "请选择行业："
        read industry_choice
        
        case $industry_choice in
            1) industry="ecommerce" ;;
            2) industry="education" ;;
            3) industry="finance" ;;
            4) industry="healthcare" ;;
            5) industry="technology" ;;
            *) echo "无效选择"; exit 1 ;;
        esac
        
        echo "正在对${industry}行业网站进行学习..."
        python3 -c "from integrated_system import IntegratedWappalyzerSystem; integrated_system = IntegratedWappalyzerSystem(); integrated_system.smart_collect_and_learn(10, industry='${industry}')"
        ;;
    5)
        echo "=== 特定行业CMS指纹学习 ==="
        echo "可用行业："
        echo "1. 电子商务 (ecommerce)"
        echo "2. 教育 (education)"
        echo "3. 金融 (finance)"
        echo "4. 医疗健康 (healthcare)"
        echo "5. 科技 (technology)"
        echo -n "请选择行业："
        read industry_choice
        
        case $industry_choice in
            1) industry="ecommerce" ;;
            2) industry="education" ;;
            3) industry="finance" ;;
            4) industry="healthcare" ;;
            5) industry="technology" ;;
            *) echo "无效选择"; exit 1 ;;
        esac
        
        echo "正在对${industry}行业的CMS网站进行指纹学习..."
        python3 -c "from integrated_system import IntegratedWappalyzerSystem; integrated_system = IntegratedWappalyzerSystem(); integrated_system.cms_fingerprint_learning(10, 2, industry='${industry}')"
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
