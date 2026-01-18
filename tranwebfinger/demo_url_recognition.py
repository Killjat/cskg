#!/usr/bin/env python3
"""
演示脚本：识别教育网站URL的技术栈

流程：
1. 读取教育网站列表
2. 初始化SelfEvolvingWappalyzer
3. 批量扫描网站
4. 输出识别结果
"""

import json
from main import SelfEvolvingWappalyzer

def load_education_sites(filename="education_sites.json"):
    """加载教育网站列表"""
    print(f"正在加载 {filename} 文件...")
    with open(filename, 'r', encoding='utf-8') as f:
        data = json.load(f)
    print(f"成功加载 {data['total']} 个教育网站")
    return data['sites']

def main():
    # 1. 加载教育网站列表
    education_sites = load_education_sites()
    
    # 2. 初始化SelfEvolvingWappalyzer
    print("\n初始化Self-Evolving Wappalyzer...")
    wappalyzer = SelfEvolvingWappalyzer()
    
    # 3. 批量扫描网站（这里只扫描前10个作为演示）
    sample_sites = education_sites[:10]
    print(f"\n开始扫描 {len(sample_sites)} 个教育网站...")
    
    results = wappalyzer.batch_scan(sample_sites)
    
    # 4. 输出详细识别结果
    print("\n=== 详细识别结果 ===")
    for site, detected in results.items():
        if detected:
            print(f"\n{site} 识别到的技术：")
            for tech_name, info in detected.items():
                print(f"  - {info['name']} ({info['category']}) - 置信度: {info['confidence']:.2f} - 检测方式: {info['detection_method']}")
        else:
            print(f"\n{site} 未识别到任何技术")
    
    print("\n=== 识别流程总结 ===")
    print("1. 读取并加载URL列表")
    print("2. 初始化SelfEvolvingWappalyzer实例")
    print("3. 对每个URL执行以下操作：")
    print("   a. 发送HTTP请求获取网站响应")
    print("   b. 基于规则检测（检查响应头、HTML内容、脚本）")
    print("   c. 机器学习预测技术栈")
    print("   d. 合并检测结果，更新置信度")
    print("4. 输出扫描结果")

if __name__ == "__main__":
    main()
