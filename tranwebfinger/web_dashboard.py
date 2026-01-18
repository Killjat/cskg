#!/usr/bin/env python3
"""
Web仪表盘，用于展示新获得的指纹信息
"""

from flask import Flask, render_template, jsonify
import json
import os
import datetime

app = Flask(__name__)

# 配置文件路径
SCAN_RESULTS_FILE = 'scan_results.json'
TECHNOLOGIES_FILE = 'technologies.json'

@app.route('/')
def index():
    """主页面，展示指纹信息"""
    return render_template('index.html')

@app.route('/api/scan_results')
def api_scan_results():
    """获取扫描结果"""
    if not os.path.exists(SCAN_RESULTS_FILE):
        return jsonify([])
    
    with open(SCAN_RESULTS_FILE, 'r', encoding='utf-8') as f:
        data = json.load(f)
    
    return jsonify(data)

@app.route('/api/technologies')
def api_technologies():
    """获取技术指纹库"""
    if not os.path.exists(TECHNOLOGIES_FILE):
        return jsonify({})
    
    with open(TECHNOLOGIES_FILE, 'r', encoding='utf-8') as f:
        data = json.load(f)
    
    return jsonify(data)

@app.route('/api/new_fingerprints')
def api_new_fingerprints():
    """获取新获得的指纹"""
    if not os.path.exists(TECHNOLOGIES_FILE):
        return jsonify([])
    
    with open(TECHNOLOGIES_FILE, 'r', encoding='utf-8') as f:
        data = json.load(f)
    
    technologies = data.get('technologies', {})
    
    # 筛选机器学习所得的指纹
    new_fingerprints = []
    for tech_name, tech_info in technologies.items():
        if '机器学习所得' in tech_info.get('description', ''):
            # 提取关键字
            keywords = []
            
            # 从headers中提取关键字
            if 'headers' in tech_info:
                for header, patterns in tech_info['headers'].items():
                    for pattern in patterns:
                        keywords.extend(pattern.split())
            
            # 从html中提取关键字
            if 'html' in tech_info:
                for pattern in tech_info['html']:
                    keywords.extend(pattern.split())
            
            # 从scripts中提取关键字
            if 'scripts' in tech_info:
                for pattern in tech_info['scripts']:
                    keywords.extend(pattern.split())
            
            # 去重并筛选有效关键字
            unique_keywords = list(set(keywords))
            valid_keywords = [kw for kw in unique_keywords if len(kw) > 3 and not kw.isdigit()][:10]
            
            # 从education_sites.json获取示例URL
            url = 'N/A'
            try:
                with open('education_sites.json', 'r', encoding='utf-8') as f:
                    education_data = json.load(f)
                    if education_data.get('sites'):
                        # 为每个指纹随机分配一个教育网站URL
                        import random
                        url = random.choice(education_data['sites'])
            except Exception as e:
                pass
            
            new_fingerprints.append({
                'url': url,
                'name': tech_name,
                'display_name': tech_info.get('name', tech_name),
                'category': tech_info.get('category', 'Unknown'),
                'description': tech_info.get('description', ''),
                'keywords': valid_keywords,
                'headers': len(tech_info.get('headers', {})),
                'html_patterns': len(tech_info.get('html', [])),
                'script_patterns': len(tech_info.get('scripts', []))
            })
    
    return jsonify(new_fingerprints)

# 创建templates目录和index.html文件
if not os.path.exists('templates'):
    os.makedirs('templates')

# 注意：HTML模板已直接创建在templates/index.html文件中，不需要在代码中生成

if __name__ == '__main__':
    print("=== 指纹仪表盘启动 ===")
    print("访问地址: http://localhost:5001")
    print("按 Ctrl+C 停止服务器")
    print("======================")
    app.run(debug=True, host='0.0.0.0', port=5001)
