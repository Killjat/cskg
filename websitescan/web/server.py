#!/usr/bin/env python3
from flask import Flask, render_template, jsonify, request
import os
import csv
import json

app = Flask(__name__, template_folder='templates')

# 扫描结果缓存
scan_results = []
results_file = 'scan_results.csv'  # 扫描结果文件路径

# 调试：打印当前工作目录和文件路径
import os
print(f"当前工作目录: {os.getcwd()}")
print(f"绝对文件路径: {os.path.abspath(results_file)}")
print(f"文件是否存在: {os.path.exists(results_file)}")

# 加载扫描结果
@app.before_request
def load_scan_results():
    global scan_results
    # 只在首次请求时加载
    if not scan_results:
        scan_results = []
        
        if os.path.exists(results_file):
            with open(results_file, 'r', encoding='utf-8') as f:
                reader = csv.DictReader(f)
                for row in reader:
                    # 处理列表类型字段
                    for field in ['frameworks', 'services', 'applications', 'programming_languages']:
                        if row[field]:
                            row[field] = [item.strip() for item in row[field].split(',')]
                        else:
                            row[field] = []
                    
                    # 处理布尔类型字段
                    row['has_login_form'] = row['has_login_form'].lower() == 'true'
                    
                    scan_results.append(row)
        
        print(f"加载了 {len(scan_results)} 条扫描结果")

# 首页路由
@app.route('/')
def index():
    return render_template('index.html')

# 添加父目录到Python路径，确保能导入scanner模块
import sys
import os
sys.path.append(os.path.dirname(os.path.abspath(__file__)) + '/..')

from scanner import WebsiteScanner
from utils import load_targets

# 扫描任务页面路由
@app.route('/scan')
def scan():
    return render_template('scan.html')

# API: 获取所有扫描结果
@app.route('/api/results')
def get_results():
    return jsonify(scan_results)

# API: 刷新扫描结果
@app.route('/api/refresh')
def refresh_results():
    load_scan_results()
    return jsonify({'status': 'success', 'message': f'刷新成功，共 {len(scan_results)} 条结果'})

# API: 获取统计信息
@app.route('/api/stats')
def get_stats():
    total = len(scan_results)
    
    # 计算框架数量
    frameworks = set()
    for result in scan_results:
        frameworks.update(result['frameworks'])
    
    # 计算语言数量
    languages = set()
    for result in scan_results:
        languages.update(result['programming_languages'])
    
    # 计算含登录框的网站数量
    login_count = sum(1 for result in scan_results if result['has_login_form'])
    
    stats = {
        'total': total,
        'framework_count': len(frameworks),
        'language_count': len(languages),
        'login_count': login_count
    }
    
    return jsonify(stats)

# API: 单个URL扫描
@app.route('/api/scan/single', methods=['POST'])
def scan_single():
    try:
        
        data = request.get_json()
        url = data.get('url')
        
        if not url:
            return jsonify({'status': 'error', 'message': '缺少URL参数'}), 400
        
        # 执行扫描
        scanner = WebsiteScanner()
        result = scanner.scan(url)
        
        # 保存结果到CSV
        import csv
        
        # 读取现有结果
        existing_results = []
        if os.path.exists(results_file):
            with open(results_file, 'r', encoding='utf-8') as f:
                reader = csv.DictReader(f)
                existing_results = list(reader)
        
        # 处理当前结果
        result_copy = result.copy()
        for key in ['frameworks', 'services', 'applications', 'programming_languages']:
            if key in result_copy:
                result_copy[key] = ', '.join(result_copy[key]) if isinstance(result_copy[key], list) else str(result_copy[key])
        
        # 检查是否已存在相同URL的结果
        existing_index = -1
        for i, existing in enumerate(existing_results):
            if existing['url'] == url:
                existing_index = i
                break
        
        if existing_index >= 0:
            # 更新现有结果
            existing_results[existing_index] = result_copy
        else:
            # 添加新结果
            existing_results.append(result_copy)
        
        # 保存到文件
        fields = ['url', 'title', 'site_name', 'frameworks', 'services', 'applications', 'programming_languages', 'icp', 'has_login_form', 'error']
        with open(results_file, 'w', newline='', encoding='utf-8') as f:
            writer = csv.DictWriter(f, fieldnames=fields)
            writer.writeheader()
            writer.writerows(existing_results)
        
        # 重新加载结果缓存
        load_scan_results()
        
        return jsonify({'status': 'success', 'message': f'扫描完成：{result.get("title", "")}', 'result': result})
        
    except Exception as e:
        return jsonify({'status': 'error', 'message': str(e)}), 500

# API: 多个URL扫描
@app.route('/api/scan/multiple', methods=['POST'])
def scan_multiple():
    try:
        import time
        
        data = request.get_json()
        urls = data.get('urls', [])
        
        if not urls:
            return jsonify({'status': 'error', 'message': '缺少URL列表'}), 400
        
        scanner = WebsiteScanner()
        results = []
        
        # 读取现有结果
        import csv
        existing_results = []
        if os.path.exists(results_file):
            with open(results_file, 'r', encoding='utf-8') as f:
                reader = csv.DictReader(f)
                existing_results = list(reader)
        
        # 执行扫描
        for url in urls:
            try:
                result = scanner.scan(url)
                results.append(result)
                time.sleep(1)  # 避免请求过快
            except Exception as e:
                results.append({
                    'url': url,
                    'title': '',
                    'site_name': '',
                    'frameworks': [],
                    'services': [],
                    'applications': [],
                    'programming_languages': [],
                    'icp': '',
                    'has_login_form': False,
                    'error': str(e)
                })
        
        # 更新结果
        existing_urls = {r['url']: i for i, r in enumerate(existing_results)}
        
        for result in results:
            result_copy = result.copy()
            for key in ['frameworks', 'services', 'applications', 'programming_languages']:
                if key in result_copy:
                    result_copy[key] = ', '.join(result_copy[key]) if isinstance(result_copy[key], list) else str(result_copy[key])
            
            if result['url'] in existing_urls:
                # 更新现有结果
                existing_results[existing_urls[result['url']]] = result_copy
            else:
                # 添加新结果
                existing_results.append(result_copy)
        
        # 保存到文件
        fields = ['url', 'title', 'site_name', 'frameworks', 'services', 'applications', 'programming_languages', 'icp', 'has_login_form', 'error']
        with open(results_file, 'w', newline='', encoding='utf-8') as f:
            writer = csv.DictWriter(f, fieldnames=fields)
            writer.writeheader()
            writer.writerows(existing_results)
        
        # 重新加载结果缓存
        load_scan_results()
        
        return jsonify({'status': 'success', 'message': f'扫描完成，共扫描 {len(urls)} 个URL，成功 {len([r for r in results if not r.get("error")])} 个', 'results': results})
        
    except Exception as e:
        return jsonify({'status': 'error', 'message': str(e)}), 500

# API: 文件导入扫描
@app.route('/api/scan/file', methods=['POST'])
def scan_file():
    try:
        import tempfile
        import time
        
        if 'file' not in request.files:
            return jsonify({'status': 'error', 'message': '缺少文件参数'}), 400
        
        file = request.files['file']
        if file.filename == '':
            return jsonify({'status': 'error', 'message': '未选择文件'}), 400
        
        # 保存临时文件
        with tempfile.NamedTemporaryFile(mode='w', suffix='.txt', delete=False) as temp:
            temp.write(file.read().decode('utf-8'))
            temp_path = temp.name
        
        try:
            # 读取URL列表
            urls = load_targets(temp_path)
            
            if not urls:
                return jsonify({'status': 'error', 'message': '文件中没有有效URL'}), 400
            
            # 执行扫描
            scanner = WebsiteScanner()
            results = []
            
            # 读取现有结果
            import csv
            existing_results = []
            if os.path.exists(results_file):
                with open(results_file, 'r', encoding='utf-8') as f:
                    reader = csv.DictReader(f)
                    existing_results = list(reader)
            
            # 执行扫描
            for url in urls:
                try:
                    result = scanner.scan(url)
                    results.append(result)
                    time.sleep(1)  # 避免请求过快
                except Exception as e:
                    results.append({
                        'url': url,
                        'title': '',
                        'site_name': '',
                        'frameworks': [],
                        'services': [],
                        'applications': [],
                        'programming_languages': [],
                        'icp': '',
                        'has_login_form': False,
                        'error': str(e)
                    })
            
            # 更新结果
            existing_urls = {r['url']: i for i, r in enumerate(existing_results)}
            
            for result in results:
                result_copy = result.copy()
                for key in ['frameworks', 'services', 'applications', 'programming_languages']:
                    if key in result_copy:
                        result_copy[key] = ', '.join(result_copy[key]) if isinstance(result_copy[key], list) else str(result_copy[key])
                
                if result['url'] in existing_urls:
                    # 更新现有结果
                    existing_results[existing_urls[result['url']]] = result_copy
                else:
                    # 添加新结果
                    existing_results.append(result_copy)
            
            # 保存到文件
            fields = ['url', 'title', 'site_name', 'frameworks', 'services', 'applications', 'programming_languages', 'icp', 'has_login_form', 'error']
            with open(results_file, 'w', newline='', encoding='utf-8') as f:
                writer = csv.DictWriter(f, fieldnames=fields)
                writer.writeheader()
                writer.writerows(existing_results)
            
            # 重新加载结果缓存
            load_scan_results()
            
            return jsonify({'status': 'success', 'message': f'文件扫描完成，共扫描 {len(urls)} 个URL，成功 {len([r for r in results if not r.get("error")])} 个', 'results': results})
            
        finally:
            # 删除临时文件
            os.unlink(temp_path)
            
    except Exception as e:
        return jsonify({'status': 'error', 'message': str(e)}), 500

if __name__ == '__main__':
    print("🚀 网站扫描WEB服务启动")
    print("📡 访问地址: http://localhost:8080")
    print(f"📊 扫描结果文件: {results_file}")
    app.run(debug=True, host='0.0.0.0', port=8080)
