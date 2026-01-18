#!/usr/bin/env python3
import json
import re
import requests
from urllib.parse import urlparse
import os

# 导入智能目标生成器
from smart_targets import SmartTargetGenerator
# 导入机器学习预测器
from ml_predictor import MLTechnologyPredictor

class SelfEvolvingWappalyzer:
    def __init__(self, rules_file='technologies.json'):
        self.rules_file = rules_file
        self.technologies = {}
        self.categories = {}
        self.evolution_log = []
        self.load_rules()
        # 初始化机器学习预测器
        self.ml_predictor = MLTechnologyPredictor()
        print("已初始化机器学习技术预测器")
    
    def load_rules(self):
        """Load rules from JSON file"""
        if os.path.exists(self.rules_file):
            with open(self.rules_file, 'r', encoding='utf-8') as f:
                data = json.load(f)
                self.technologies = data.get('technologies', {})
                self.categories = data.get('categories', {})
            print(f"Loaded {len(self.technologies)} technologies from {self.rules_file}")
        else:
            print(f"Rules file {self.rules_file} not found, starting with empty ruleset")
    
    def save_rules(self):
        """Save rules to JSON file"""
        data = {
            'technologies': self.technologies,
            'categories': self.categories
        }
        with open(self.rules_file, 'w', encoding='utf-8') as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
        print(f"Saved {len(self.technologies)} technologies to {self.rules_file}")
    
    def scan(self, url):
        """Scan a website for technologies using both rule-based detection and machine learning"""
        print(f"Scanning {url}...")
        detected = {}
        
        try:
            response = requests.get(url, timeout=10)
            html = response.text
            
            # 1. 基于规则的检测
            print("  - 使用基于规则的检测...")
            
            # Check headers
            for tech_name, tech in self.technologies.items():
                if 'headers' in tech:
                    for header, patterns in tech['headers'].items():
                        if header in response.headers:
                            header_value = response.headers[header]
                            for pattern in patterns:
                                if re.search(pattern, header_value, re.IGNORECASE):
                                    detected[tech_name] = {
                                        'name': tech['name'],
                                        'category': tech['category'],
                                        'detection_method': f'rule:header:{header}',
                                        'confidence': 0.9
                                    }
            
            # Check HTML content
            for tech_name, tech in self.technologies.items():
                if tech_name not in detected:
                    # Check HTML patterns
                    if 'html' in tech:
                        for pattern in tech['html']:
                            if re.search(pattern, html, re.IGNORECASE):
                                detected[tech_name] = {
                                    'name': tech['name'],
                                    'category': tech['category'],
                                    'detection_method': 'rule:html',
                                    'confidence': 0.8
                                }
                                break
                    
                    # Check script patterns if not already detected
                    if tech_name not in detected and 'scripts' in tech:
                        for pattern in tech['scripts']:
                            if re.search(pattern, html, re.IGNORECASE):
                                detected[tech_name] = {
                                    'name': tech['name'],
                                    'category': tech['category'],
                                    'detection_method': 'rule:script',
                                    'confidence': 0.7
                                }
                                break
            
            # 2. 机器学习预测
            print("  - 使用机器学习预测...")
            website_data = {
                'url': url,
                'headers': dict(response.headers),
                'html': html[:10000]  # 只使用前10000个字符，避免过长
            }
            
            ml_result = self.ml_predictor.predict_technology(website_data)
            if ml_result['technology'] and ml_result['confidence'] > 0.5:
                tech_name = ml_result['technology']
                # 检查该技术是否已在规则检测中被发现
                if tech_name not in detected:
                    # 如果技术在我们的规则库中，使用规则库中的信息
                    if tech_name in self.technologies:
                        tech_info = self.technologies[tech_name]
                        detected[tech_name] = {
                            'name': tech_info['name'],
                            'category': tech_info['category'],
                            'detection_method': f'ml:prediction',
                            'confidence': ml_result['confidence']
                        }
                    else:
                        # 否则使用机器学习预测的结果
                        detected[tech_name] = {
                            'name': tech_name,
                            'category': 'Unknown',
                            'detection_method': f'ml:prediction',
                            'confidence': ml_result['confidence']
                        }
                else:
                    # 如果已存在，更新置信度
                    existing_confidence = detected[tech_name]['confidence']
                    if ml_result['confidence'] > existing_confidence:
                        detected[tech_name]['confidence'] = ml_result['confidence']
                        detected[tech_name]['detection_method'] = f"combined:{detected[tech_name]['detection_method']},ml:prediction"
            
            # 3. 打印检测结果
            print(f"Detected {len(detected)} technologies on {url}:")
            for tech_name, info in detected.items():
                print(f"- {info['name']} ({info['category']}) via {info['detection_method']} (置信度: {info['confidence']:.2f})")
            
            return detected
            
        except Exception as e:
            print(f"Error scanning {url}: {e}")
            return {}
    
    def extract_features_from_response(self, response):
        """从响应中提取特征"""
        features = {
            'headers': {},
            'html': [],
            'scripts': []
        }
        
        # 提取响应头特征
        for header, value in response.headers.items():
            # 只提取有意义的响应头
            if header.lower() in ['server', 'x-powered-by', 'x-generator', 'generator', 'set-cookie']:
                features['headers'][header] = value
        
        # 提取HTML特征
        html = response.text.lower()
        
        # 提取meta标签
        meta_tags = re.findall(r'<meta[^>]+>', html)
        for meta in meta_tags:
            if 'generator' in meta or 'powered-by' in meta or 'description' in meta:
                features['html'].append(meta)
        
        # 提取script标签
        scripts = re.findall(r'<script[^>]+src=["\']([^"\']+)["\']', html)
        features['scripts'] = scripts[:10]  # 只保留前10个
        
        # 提取关键词
        keywords = ['react', 'vue', 'angular', 'jquery', 'bootstrap', 'wordpress', 'joomla', 'drupal',
                   'magento', 'shopify', 'wix', 'squarespace', 'ghost', 'django', 'flask', 'laravel',
                   'node.js', 'express', 'php', 'python', 'ruby', 'java', 'asp.net']
        
        found_keywords = []
        for keyword in keywords:
            if keyword.lower() in html:
                found_keywords.append(keyword)
        
        features['keywords'] = found_keywords
        
        return features
    
    def evolve(self, url, expected_techs):
        """Evolve rules based on expected technologies"""
        print(f"\nEvolving rules using {url}...")
        try:
            response = requests.get(url, timeout=10)
            
            # Get current detection results
            current_detections = self.scan(url)
            current_tech_names = set(current_detections.keys())
            expected_tech_names = set(expected_techs.keys())
            
            # Identify missing detections
            missing = expected_tech_names - current_tech_names
            
            if missing:
                print(f"Missing detections: {missing}")
                
                # Extract potential patterns from response
                headers = response.headers
                html = response.text
                
                # 提取响应特征
                features = self.extract_features_from_response(response)
                
                for tech_name in missing:
                    tech_info = expected_techs[tech_name]
                    
                    # 检查是否是机器学习检测到的技术
                    is_ml_detected = tech_info.get('ml_detected', False)
                    
                    # 为机器学习检测到的技术生成特征关键字名称
                    if is_ml_detected and not tech_name:
                        # 从特征中提取关键字
                        if features['keywords']:
                            # 使用第一个关键字作为技术名称
                            tech_name = features['keywords'][0]
                            print(f"从特征关键字生成技术名称: {tech_name}")
                        else:
                            # 如果没有关键字，使用默认名称
                            tech_name = f"ML_Detected_{int(datetime.now().timestamp())}"
                            print(f"生成默认技术名称: {tech_name}")
                    
                    # 确保技术存在于规则库中
                    if tech_name not in self.technologies:
                        description = f"Auto-detected technology: {tech_info['name']}"
                        if is_ml_detected:
                            description = f"{tech_info['name']} (机器学习所得)"
                        
                        self.technologies[tech_name] = {
                            'name': tech_info['name'],
                            'category': tech_info['category'],
                            'description': description,
                            'website': '',
                            'headers': {},
                            'html': [],
                            'scripts': []
                        }
                    
                    # 分析响应头
                    print(f"Analyzing headers for {tech_name}...")
                    for header, value in features['headers'].items():
                        # 检查是否有技术相关的模式
                        patterns = []
                        
                        # 提取响应头中的特征值
                        if tech_name.lower() in value.lower() or tech_info['name'].lower() in value.lower():
                            patterns.append(value[:50])  # 取前50个字符作为模式
                        
                        # 为机器学习检测到的技术添加所有相关响应头
                        if is_ml_detected:
                            patterns.append(value[:50])
                        
                        # 添加到规则中
                        for pattern in patterns:
                            escaped_pattern = re.escape(pattern)
                            if header not in self.technologies[tech_name]['headers']:
                                self.technologies[tech_name]['headers'][header] = []
                            
                            if escaped_pattern not in self.technologies[tech_name]['headers'][header]:
                                self.technologies[tech_name]['headers'][header].append(escaped_pattern)
                                self.evolution_log.append({
                                    'tech': tech_name,
                                    'method': 'header',
                                    'header': header,
                                    'pattern': escaped_pattern,
                                    'source': url
                                })
                    
                    # 分析HTML特征
                    print(f"Analyzing HTML for {tech_name}...")
                    
                    # 添加meta标签特征
                    for meta in features['html']:
                        escaped_meta = re.escape(meta)[:100]  # 限制长度
                        if escaped_meta not in self.technologies[tech_name]['html']:
                            self.technologies[tech_name]['html'].append(escaped_meta)
                            self.evolution_log.append({
                                'tech': tech_name,
                                'method': 'html',
                                'pattern': escaped_meta,
                                'source': url
                            })
                    
                    # 添加script标签特征
                    for script in features['scripts']:
                        escaped_script = re.escape(script)
                        if escaped_script not in self.technologies[tech_name]['scripts']:
                            self.technologies[tech_name]['scripts'].append(escaped_script)
                            self.evolution_log.append({
                                'tech': tech_name,
                                'method': 'script',
                                'pattern': escaped_script,
                                'source': url
                            })
                    
                    # 为机器学习检测到的技术添加关键字特征
                    if is_ml_detected and features['keywords']:
                        for keyword in features['keywords']:
                            # 添加关键字作为HTML模式
                            keyword_pattern = f"\\b{keyword}\\b"  # 使用单词边界
                            if keyword_pattern not in self.technologies[tech_name]['html']:
                                self.technologies[tech_name]['html'].append(keyword_pattern)
                                self.evolution_log.append({
                                    'tech': tech_name,
                                    'method': 'html',
                                    'pattern': keyword_pattern,
                                    'source': url
                                })
                
                # Save evolved rules
                self.save_rules()
            else:
                print("All expected technologies were detected. No evolution needed.")
            
        except Exception as e:
            print(f"Error during evolution: {e}")
    
    def add_technology(self, tech_name, name, category, description="", website="", patterns=None):
        """Add a new technology to the rules"""
        if tech_name not in self.technologies:
            self.technologies[tech_name] = {
                'name': name,
                'category': category,
                'description': description,
                'website': website,
                'headers': {},
                'html': [],
                'scripts': []
            }
            
            if patterns:
                if 'headers' in patterns:
                    self.technologies[tech_name]['headers'] = patterns['headers']
                if 'html' in patterns:
                    self.technologies[tech_name]['html'] = patterns['html']
                if 'scripts' in patterns:
                    self.technologies[tech_name]['scripts'] = patterns['scripts']
            
            print(f"Added technology: {name}")
            self.save_rules()
        else:
            print(f"Technology {tech_name} already exists")
    
    def remove_technology(self, tech_name):
        """Remove a technology from the rules"""
        if tech_name in self.technologies:
            del self.technologies[tech_name]
            print(f"Removed technology: {tech_name}")
            self.save_rules()
        else:
            print(f"Technology {tech_name} not found")
    
    def show_stats(self):
        """Show statistics about the rules"""
        print("\n=== Self-Evolving Wappalyzer Statistics ===")
        print(f"Total technologies: {len(self.technologies)}")
        print(f"Total categories: {len(self.categories)}")
        print(f"Evolution events: {len(self.evolution_log)}")
        
        # Category breakdown
        category_counts = {}
        for tech in self.technologies.values():
            category = tech['category']
            category_counts[category] = category_counts.get(category, 0) + 1
        
        print("\nTechnologies by category:")
        for category, count in category_counts.items():
            print(f"- {category}: {count}")
        
        if self.evolution_log:
            print("\nRecent evolution events:")
            for event in self.evolution_log[-5:]:
                print(f"- {event['tech']} via {event['method']} from {event['source']}")
    
    def get_top_domains(self, count=100):
        """获取热门网站域名列表"""
        # 这里使用一些常见的网站域名作为示例
        # 在实际应用中，可以从DNS查询、公共域名列表或其他数据源获取
        common_domains = [
            "https://www.baidu.com",
            "https://www.google.com",
            "https://www.youtube.com",
            "https://www.facebook.com",
            "https://www.wikipedia.org",
            "https://www.qq.com",
            "https://www.twitter.com",
            "https://www.instagram.com",
            "https://www.taobao.com",
            "https://www.amazon.com",
            "https://www.microsoft.com",
            "https://www.netflix.com",
            "https://www.reddit.com",
            "https://www.360.cn",
            "https://www.sohu.com",
            "https://www.sina.com.cn",
            "https://www.jd.com",
            "https://www.tmall.com",
            "https://www.yahoo.com",
            "https://www.bing.com",
            "https://www.github.com",
            "https://www.tencent.com",
            "https://www.alibaba.com",
            "https://www.zoom.us",
            "https://www.linkedin.com",
            "https://www.pinterest.com",
            "https://www.tumblr.com",
            "https://www.microsoftonline.com",
            "https://www.office.com",
            "https://www.huawei.com",
            "https://www.apple.com",
            "https://www.samsung.com",
            "https://www.mi.com",
            "https://www.xiaomi.com",
            "https://www.lenovo.com",
            "https://www.asus.com",
            "https://www.dell.com",
            "https://www.hp.com",
            "https://www.ibm.com",
            "https://www.oracle.com",
            "https://www.adobe.com",
            "https://www.autodesk.com",
            "https://www.elastic.co",
            "https://www.elastic.io",
            "https://www.elasticsearch.org",
            "https://www.elastic.co/elasticsearch",
            "https://www.elastic.co/kibana",
            "https://www.elastic.co/logstash",
            "https://www.elastic.co/beats",
            "https://www.elastic.co/cloud",
            "https://www.elastic.co/security",
            "https://www.elastic.co/observability",
            "https://www.elastic.co/enterprise-search",
            "https://www.elastic.co/elasticsearch-service",
            "https://www.elastic.co/kibana-service",
            "https://www.elastic.co/logstash-service",
            "https://www.elastic.co/beats-service",
            "https://www.elastic.co/cloud-on-kubernetes",
            "https://www.elastic.co/elasticsearch-operator",
            "https://www.elastic.co/elasticsearch-sql",
            "https://www.elastic.co/elasticsearch-hadoop",
            "https://www.elastic.co/elasticsearch-py",
            "https://www.elastic.co/elasticsearch-js",
            "https://www.elastic.co/elasticsearch-java",
            "https://www.elastic.co/elasticsearch-net",
            "https://www.elastic.co/elasticsearch-php",
            "https://www.elastic.co/elasticsearch-ruby",
            "https://www.elastic.co/elasticsearch-go",
            "https://www.elastic.co/elasticsearch-dotnet",
            "https://www.elastic.co/elasticsearch-perl",
            "https://www.elastic.co/elasticsearch-swift",
            "https://www.elastic.co/elasticsearch-kotlin",
            "https://www.elastic.co/elasticsearch-scala",
            "https://www.elastic.co/elasticsearch-groovy",
            "https://www.elastic.co/elasticsearch-csharp",
            "https://www.elastic.co/elasticsearch-cpp",
            "https://www.elastic.co/elasticsearch-rust",
            "https://www.elastic.co/elasticsearch-dart",
            "https://www.elastic.co/elasticsearch-elixir",
            "https://www.elastic.co/elasticsearch-clojure",
            "https://www.elastic.co/elasticsearch-haskell",
            "https://www.elastic.co/elasticsearch-erlang",
            "https://www.elastic.co/elasticsearch-lua",
            "https://www.elastic.co/elasticsearch-nim",
            "https://www.elastic.co/elasticsearch-ocaml",
            "https://www.elastic.co/elasticsearch-pascal",
            "https://www.elastic.co/elasticsearch-r",
            "https://www.elastic.co/elasticsearch-scheme",
            "https://www.elastic.co/elasticsearch-smalltalk",
            "https://www.elastic.co/elasticsearch-typescript",
            "https://www.elastic.co/elasticsearch-v",
            "https://www.elastic.co/elasticsearch-zig"
        ]
        
        # 如果请求的数量小于列表长度，返回前count个
        # 否则返回整个列表
        return common_domains[:count] if count <= len(common_domains) else common_domains
    
    def batch_scan(self, domains):
        """批量扫描网站"""
        print(f"\n开始批量扫描 {len(domains)} 个网站...")
        results = {}
        
        for domain in domains:
            detected = self.scan(domain)
            results[domain] = detected
        
        # 打印扫描结果摘要
        print(f"\n=== 批量扫描结果摘要 ===")
        total_detected = 0
        for domain, detected in results.items():
            count = len(detected)
            total_detected += count
            print(f"{domain}: 检测到 {count} 种技术")
        
        print(f"\n总共扫描 {len(domains)} 个网站，检测到 {total_detected} 种技术")
        
        return results
    
    def smart_scan(self, count=50, config=None):
        """使用智能目标生成器扫描网站"""
        # 创建智能目标生成器实例
        generator = SmartTargetGenerator(config)
        
        # 生成扫描目标
        targets = generator.generate_targets(count)
        
        # 批量扫描这些目标
        return self.batch_scan(targets)
    
    def cms_fingerprint_update(self, target_count=20):
        """更新CMS指纹"""
        from integrated_system import IntegratedWappalyzerSystem
        integrated_system = IntegratedWappalyzerSystem()
        return integrated_system.cms_fingerprint_learning(target_count)

if __name__ == "__main__":
    # Create instance
    wappalyzer = SelfEvolvingWappalyzer()
    
    # Example usage
    wappalyzer.show_stats()
    
    # 选项1：使用传统方法批量扫描
    # domains = wappalyzer.get_top_domains(100)
    # results = wappalyzer.batch_scan(domains)
    
    # 选项2：使用智能目标生成器扫描（推荐）
    # 配置智能目标生成器
    smart_config = {
        'max_targets': 10,
        'validation_timeout': 5,
        'sources': {
            'alexa': True
        },
        'filters': {
            'tld': [],  # 允许所有顶级域名
            'min_estimated_visits': 0
        }
    }
    # 执行智能扫描
    results = wappalyzer.smart_scan(5, smart_config)
    
    # 显示最终统计信息
    wappalyzer.show_stats()