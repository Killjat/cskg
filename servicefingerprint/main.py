#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
服务指纹探测工具 v2.0
增强功能：
1. 当没有给目标时，自动进行全网爬虫扫描URL
2. 当无法识别技术时，学习指纹规则并更新指纹数据库
"""

import os
import sys
import time
import requests
import yaml
import csv
import json
import re
from concurrent.futures import ThreadPoolExecutor
from tqdm import tqdm
from urllib.parse import urlparse, urljoin
from collections import defaultdict


class ServiceFingerprintScanner:
    """服务指纹扫描器"""
    
    def __init__(self, config_path="config.yaml"):
        """初始化扫描器"""
        self.config = self.load_config(config_path)
        self.fingerprints = self.load_fingerprints(self.config['fingerprint']['file'])
        self.targets = []
        self.discovered_urls = set()  # 用于去重的已发现URL集合
        self.session = requests.Session()
        self.session.headers.update({'User-Agent': self.config['http']['user_agent']})
        self.session.timeout = self.config['targets']['timeout']
        self.session.max_redirects = self.config['http']['max_redirects']
        
    def load_config(self, config_path):
        """加载配置文件"""
        default_config = {
            'targets': {
                'file': 'targets.txt',
                'timeout': 5,
                'concurrency': 10
            },
            'http': {
                'user_agent': 'ServiceFingerprint/1.0',
                'follow_redirects': True,
                'max_redirects': 5
            },
            'fingerprint': {
                'file': 'fingerprints.yaml'
            },
            'output': {
                'file': 'results.csv',
                'format': 'csv',
                'verbose': True
            },
            'crawler': {
                'enabled': True,
                'max_depth': 2,
                'max_urls': 100,
                'seed_urls': ['http://example.com', 'https://github.com']
            },
            'learning': {
                'enabled': True,
                'min_confidence': 0.8
            }
        }
        
        # 如果配置文件存在，加载并合并
        if os.path.exists(config_path):
            with open(config_path, 'r', encoding='utf-8') as f:
                user_config = yaml.safe_load(f)
                if user_config:
                    # 合并配置
                    self.merge_dict(default_config, user_config)
        
        return default_config
    
    def merge_dict(self, default, user):
        """合并字典"""
        for key, value in user.items():
            if isinstance(value, dict) and key in default:
                self.merge_dict(default[key], value)
            else:
                default[key] = value
    
    def load_fingerprints(self, fingerprint_path):
        """加载指纹数据库"""
        if not os.path.exists(fingerprint_path):
            print(f"警告: 指纹文件 {fingerprint_path} 不存在，使用空指纹数据库")
            return []
        
        with open(fingerprint_path, 'r', encoding='utf-8') as f:
            data = yaml.safe_load(f)
            return data.get('fingerprints', [])
    
    def load_targets(self, targets_path):
        """加载目标列表"""
        self.targets = []
        
        if os.path.exists(targets_path):
            with open(targets_path, 'r', encoding='utf-8') as f:
                for line in f:
                    line = line.strip()
                    if line and not line.startswith('#'):
                        # 确保URL格式正确
                        if not line.startswith(('http://', 'https://')):
                            line = 'http://' + line
                        self.targets.append(line)
                        self.discovered_urls.add(line)
        
        # 如果没有目标，启动爬虫模式
        if not self.targets and self.config['crawler']['enabled']:
            print("未找到目标，启动全网爬虫模式")
            self.crawl_urls()
    
    def crawl_urls(self):
        """全网爬虫，发现URL"""
        print(f"启动爬虫，从种子URL开始爬取")
        
        # 使用种子URL作为初始目标
        queue = []
        for seed_url in self.config['crawler']['seed_urls']:
            queue.append((seed_url, 0))
            self.discovered_urls.add(seed_url)
        
        crawled_count = 0
        
        while queue and crawled_count < self.config['crawler']['max_urls']:
            url, depth = queue.pop(0)
            
            if depth >= self.config['crawler']['max_depth']:
                continue
            
            print(f"爬行: {url} (深度: {depth})")
            crawled_count += 1
            
            try:
                # 发送HTTP请求
                resp = self.session.get(url, allow_redirects=True)
                
                # 提取页面中的所有链接
                urls = self.extract_links(url, resp.text)
                
                # 添加到队列和已发现URL集合
                for new_url in urls:
                    if new_url not in self.discovered_urls:
                        self.discovered_urls.add(new_url)
                        queue.append((new_url, depth + 1))
            except Exception as e:
                print(f"爬行失败: {url} - {e}")
                continue
        
        # 将发现的URL添加到目标列表
        self.targets = list(self.discovered_urls)
        print(f"爬虫完成，发现 {len(self.targets)} 个URL")
    
    def extract_links(self, base_url, html):
        """从HTML中提取链接"""
        links = []
        
        # 提取所有href属性
        href_pattern = re.compile(r'href=["\'](.*?)["\']', re.IGNORECASE)
        matches = href_pattern.findall(html)
        
        for match in matches:
            # 处理相对URL
            url = urljoin(base_url, match)
            
            # 解析URL，只保留http/https协议
            parsed = urlparse(url)
            if parsed.scheme in ['http', 'https']:
                # 移除片段标识符
                url = parsed.scheme + '://' + parsed.netloc + parsed.path
                links.append(url)
        
        return links
    
    def scan(self):
        """执行扫描"""
        print(f"正在加载目标列表: {self.config['targets']['file']}")
        self.load_targets(self.config['targets']['file'])
        print(f"成功加载 {len(self.targets)} 个目标")
        
        print(f"\n正在加载指纹数据库: {self.config['fingerprint']['file']}")
        self.fingerprints = self.load_fingerprints(self.config['fingerprint']['file'])
        print(f"成功加载 {len(self.fingerprints)} 个指纹规则")
        
        print("\n开始扫描...")
        
        results = []
        unknown_responses = []  # 存储无法识别的响应，用于学习
        
        # 使用线程池并发扫描
        with ThreadPoolExecutor(max_workers=self.config['targets']['concurrency']) as executor:
            future_to_url = {executor.submit(self.scan_target, url): url for url in self.targets}
            
            for future in tqdm(future_to_url, desc="扫描进度"):
                url = future_to_url[future]
                try:
                    result = future.result()
                    results.append(result)
                    
                    # 如果无法识别技术，保存响应用于学习
                    if self.config['learning']['enabled'] and not result['technologies']:
                        unknown_responses.append(result)
                except Exception as e:
                    print(f"扫描 {url} 时出错: {e}")
        
        # 如果启用了学习功能，处理无法识别的响应
        if self.config['learning']['enabled'] and unknown_responses:
            self.learn_fingerprints(unknown_responses)
        
        return results
    
    def scan_target(self, url):
        """扫描单个目标"""
        result = {
            'url': url,
            'status_code': 0,
            'server': '',
            'technologies': [],
            'headers': {},
            'error': '',
            'timestamp': time.strftime('%Y-%m-%d %H:%M:%S')
        }
        
        try:
            # 发送HTTP请求
            resp = self.session.get(url, allow_redirects=self.config['http']['follow_redirects'])
            result['status_code'] = resp.status_code
            
            # 获取Server头
            if 'Server' in resp.headers:
                result['server'] = resp.headers['Server']
            
            # 保存所有响应头
            result['headers'] = dict(resp.headers)
            
            # 匹配指纹
            result['technologies'] = self.match_fingerprints(resp.headers, resp.text)
            
        except Exception as e:
            result['error'] = str(e)
        
        return result
    
    def match_fingerprints(self, headers, body):
        """匹配指纹"""
        matched_techs = []
        
        for fp in self.fingerprints:
            matched = False
            for match_rule in fp['matches']:
                if self.check_match(match_rule, headers, body):
                    matched = True
                    break
            
            if matched:
                matched_techs.append({
                    'name': fp['name'],
                    'category': fp['category']
                })
        
        return matched_techs
    
    def check_match(self, match_rule, headers, body):
        """检查匹配规则"""
        match_type = match_rule.get('type')
        
        if match_type == 'header':
            key = match_rule.get('key')
            value = match_rule.get('value')
            if key in headers:
                return re.search(value, headers[key], re.IGNORECASE) is not None
        elif match_type == 'html':
            pattern = match_rule.get('pattern', '')
            return re.search(pattern, body, re.IGNORECASE) is not None
        
        return False
    
    def learn_fingerprints(self, unknown_responses):
        """从无法识别的响应中学习新的指纹规则"""
        print(f"\n开始学习新的指纹规则，共有 {len(unknown_responses)} 个未知响应")
        
        # 分析响应头，找出常见的Server头值
        server_headers = defaultdict(int)
        for resp in unknown_responses:
            if resp['status_code'] == 200 and 'headers' in resp:
                if 'Server' in resp['headers']:
                    server = resp['headers']['Server']
                    server_headers[server] += 1
        
        # 分析X-Powered-By头
        powered_by_headers = defaultdict(int)
        for resp in unknown_responses:
            if resp['status_code'] == 200 and 'headers' in resp:
                if 'X-Powered-By' in resp['headers']:
                    powered_by = resp['headers']['X-Powered-By']
                    powered_by_headers[powered_by] += 1
        
        # 生成新的指纹规则
        new_fingerprints = []
        total_responses = len(unknown_responses)
        
        # 基于Server头生成规则
        for server, count in server_headers.items():
            confidence = count / total_responses
            if confidence >= self.config['learning']['min_confidence']:
                # 提取服务器名称
                server_name = server.split('/')[0] if '/' in server else server
                server_name = re.sub(r'\s+', '_', server_name)
                
                new_fp = {
                    'name': server_name,
                    'category': 'web_server',
                    'matches': [
                        {
                            'type': 'header',
                            'key': 'Server',
                            'value': server_name
                        }
                    ]
                }
                new_fingerprints.append(new_fp)
        
        # 基于X-Powered-By头生成规则
        for powered_by, count in powered_by_headers.items():
            confidence = count / total_responses
            if confidence >= self.config['learning']['min_confidence']:
                # 提取技术名称
                tech_name = powered_by.split('/')[0] if '/' in powered_by else powered_by
                
                new_fp = {
                    'name': tech_name,
                    'category': 'programming_language',
                    'matches': [
                        {
                            'type': 'header',
                            'key': 'X-Powered-By',
                            'value': tech_name
                        }
                    ]
                }
                new_fingerprints.append(new_fp)
        
        # 添加新的指纹规则到数据库
        if new_fingerprints:
            # 加载现有的指纹数据库
            fingerprints = self.load_fingerprints(self.config['fingerprint']['file'])
            
            # 添加新的指纹规则
            for new_fp in new_fingerprints:
                # 检查是否已存在
                exists = False
                for fp in fingerprints:
                    if fp['name'] == new_fp['name'] and fp['category'] == new_fp['category']:
                        exists = True
                        break
                
                if not exists:
                    fingerprints.append(new_fp)
                    print(f"学习到新规则: {new_fp['name']} ({new_fp['category']})")
            
            # 保存更新后的指纹数据库
            self.save_fingerprints(fingerprints)
    
    def save_fingerprints(self, fingerprints):
        """保存指纹数据库"""
        data = {'fingerprints': fingerprints}
        with open(self.config['fingerprint']['file'], 'w', encoding='utf-8') as f:
            yaml.dump(data, f, allow_unicode=True, sort_keys=False)
        
        print(f"指纹数据库已更新，保存到: {self.config['fingerprint']['file']}")
    
    def save_results(self, results, output_file, output_format):
        """保存结果到文件"""
        if output_format == 'csv':
            self.save_results_csv(results, output_file)
        elif output_format == 'json':
            self.save_results_json(results, output_file)
        elif output_format == 'txt':
            self.save_results_txt(results, output_file)
    
    def save_results_csv(self, results, output_file):
        """保存为CSV格式"""
        with open(output_file, 'w', newline='', encoding='utf-8') as f:
            fieldnames = ['URL', '状态码', '服务器', '技术', '错误', '时间戳']
            writer = csv.DictWriter(f, fieldnames=fieldnames)
            writer.writeheader()
            
            for result in results:
                techs = '; '.join([f"{t['name']} ({t['category']})" for t in result['technologies']])
                writer.writerow({
                    'URL': result['url'],
                    '状态码': result['status_code'],
                    '服务器': result['server'],
                    '技术': techs,
                    '错误': result['error'],
                    '时间戳': result['timestamp']
                })
    
    def save_results_json(self, results, output_file):
        """保存为JSON格式"""
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(results, f, ensure_ascii=False, indent=2)
    
    def save_results_txt(self, results, output_file):
        """保存为TXT格式"""
        with open(output_file, 'w', encoding='utf-8') as f:
            for result in results:
                f.write(f"URL: {result['url']}\n")
                f.write(f"状态码: {result['status_code']}\n")
                if result['server']:
                    f.write(f"服务器: {result['server']}\n")
                
                if result['technologies']:
                    f.write("技术: ")
                    techs = ', '.join([f"{t['name']} ({t['category']})" for t in result['technologies']])
                    f.write(f"{techs}\n")
                else:
                    f.write("技术: 无\n")
                
                if result['error']:
                    f.write(f"错误: {result['error']}\n")
                
                f.write(f"时间戳: {result['timestamp']}\n")
                f.write("=" * 50 + "\n")


def main():
    """主函数"""
    print("服务指纹探测工具 v2.0")
    print("增强功能: 全网爬虫 + 自动学习指纹规则")
    print("=" * 60)
    
    # 解析命令行参数
    import argparse
    parser = argparse.ArgumentParser(description='服务指纹探测工具')
    parser.add_argument('--config', default='config.yaml', help='配置文件路径')
    parser.add_argument('--targets', help='目标文件路径（覆盖配置）')
    parser.add_argument('--output', help='输出文件路径（覆盖配置）')
    args = parser.parse_args()
    
    # 创建扫描器
    scanner = ServiceFingerprintScanner(args.config)
    
    # 覆盖配置
    if args.targets:
        scanner.config['targets']['file'] = args.targets
    if args.output:
        scanner.config['output']['file'] = args.output
    
    # 加载目标
    scanner.load_targets(scanner.config['targets']['file'])
    
    # 如果没有目标，退出
    if not scanner.targets:
        print("未找到目标，退出")
        return
    
    # 执行扫描
    results = scanner.scan()
    
    # 打印结果
    print("\n扫描结果")
    print("=" * 50)
    
    for result in results:
        print(f"\nURL: {result['url']}")
        print(f"状态码: {result['status_code']}")
        if result['server']:
            print(f"服务器: {result['server']}")
        
        if result['technologies']:
            print("技术:")
            for tech in result['technologies']:
                print(f"  - {tech['name']} ({tech['category']})")
        else:
            print("技术: 无")
        
        if result['error']:
            print(f"错误: {result['error']}")
    
    # 保存结果
    scanner.save_results(results, scanner.config['output']['file'], scanner.config['output']['format'])
    print(f"\n结果已保存到: {scanner.config['output']['file']}")
    print("\n扫描任务完成！")


if __name__ == '__main__':
    main()
