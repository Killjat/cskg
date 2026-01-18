#!/usr/bin/env python3
"""
基于配置的扫描器，支持多源目标获取和爬虫功能
"""

import json
import os
import requests
import re
import time
from urllib.parse import urlparse, urljoin
from concurrent.futures import ThreadPoolExecutor, as_completed
from main import SelfEvolvingWappalyzer

class CrawlScanner:
    def __init__(self, config_path="config.json"):
        """初始化基于配置的扫描器"""
        self.config_path = config_path
        self.config = self.load_config()
        self.wappalyzer = SelfEvolvingWappalyzer()
        self.visited_urls = set()
        self.discovered_urls = set()
        self.results = {}
        
    def load_config(self):
        """加载配置文件"""
        if not os.path.exists(self.config_path):
            raise FileNotFoundError(f"配置文件 {self.config_path} 不存在")
        
        with open(self.config_path, 'r', encoding='utf-8') as f:
            config = json.load(f)
        
        return config
    
    def get_targets_from_file(self):
        """从文件中获取目标"""
        file_config = self.config.get("scan_targets", {}).get("sources", {}).get("file", {})
        if not file_config.get("enabled", False):
            return []
        
        file_path = file_config.get("file_path")
        limit = file_config.get("limit", 1000)
        
        if not os.path.exists(file_path):
            print(f"警告: 文件 {file_path} 不存在")
            return []
        
        with open(file_path, 'r', encoding='utf-8') as f:
            data = json.load(f)
        
        sites = data.get("sites", [])
        return sites[:limit]
    
    def crawl_url(self, url, depth=0):
        """爬取单个URL"""
        if depth > self.config.get("scan_targets", {}).get("sources", {}).get("crawler", {}).get("max_depth", 2):
            return []
        
        if url in self.visited_urls:
            return []
        
        self.visited_urls.add(url)
        discovered = []
        
        try:
            crawler_config = self.config.get("scan_targets", {}).get("sources", {}).get("crawler", {})
            headers = {
                "User-Agent": crawler_config.get("user_agent", "Mozilla/5.0 (compatible; SelfEvolvingWappalyzer/1.0)")
            }
            
            response = requests.get(url, timeout=self.config.get("scan_targets", {}).get("crawl_strategy", {}).get("timeout", 15), 
                                  headers=headers, allow_redirects=True)
            
            if response.status_code != 200:
                return []
            
            # 检查是否符合过滤条件
            parsed_url = urlparse(url)
            
            # 检查TLD过滤
            tld_filters = self.config.get("scan_targets", {}).get("filters", {}).get("tld", [])
            if tld_filters:
                domain_parts = parsed_url.netloc.split('.')
                tld = '.'.join(domain_parts[-2:]) if len(domain_parts) > 1 else domain_parts[0]
                if tld not in tld_filters:
                    return []
            
            # 检查路径过滤
            allowed_paths = self.config.get("scan_targets", {}).get("sources", {}).get("crawler", {}).get("allowed_paths", [])
            if allowed_paths:
                path = parsed_url.path
                if not any(allowed in path for allowed in allowed_paths):
                    return []
            
            disallowed_paths = self.config.get("scan_targets", {}).get("sources", {}).get("crawler", {}).get("disallowed_paths", [])
            if disallowed_paths:
                path = parsed_url.path
                if any(disallowed in path for disallowed in disallowed_paths):
                    return []
            
            discovered.append(url)
            
            # 提取链接继续爬取
            if depth < self.config.get("scan_targets", {}).get("sources", {}).get("crawler", {}).get("max_depth", 2):
                links = re.findall(r'<a[^>]+href=["\']([^"\']+)["\']', response.text)
                for link in links:
                    # 转换为绝对URL
                    abs_url = urljoin(url, link)
                    
                    # 只处理HTTP/HTTPS链接
                    if abs_url.startswith("http://") or abs_url.startswith("https://"):
                        # 检查域名限制
                        parsed_abs = urlparse(abs_url)
                        allowed_domains = self.config.get("scan_targets", {}).get("sources", {}).get("crawler", {}).get("allowed_domains", [])
                        if allowed_domains:
                            if parsed_abs.netloc not in allowed_domains:
                                continue
                        
                        disallowed_domains = self.config.get("scan_targets", {}).get("sources", {}).get("crawler", {}).get("disallowed_domains", [])
                        if disallowed_domains:
                            if parsed_abs.netloc in disallowed_domains:
                                continue
                        
                        # 继续爬取
                        child_discovered = self.crawl_url(abs_url, depth + 1)
                        discovered.extend(child_discovered)
            
        except Exception as e:
            print(f"爬取 {url} 时出错: {e}")
        
        return discovered
    
    def crawl_targets(self):
        """执行爬虫获取目标"""
        crawler_config = self.config.get("scan_targets", {}).get("sources", {}).get("crawler", {})
        if not crawler_config.get("enabled", False):
            return []
        
        start_urls = crawler_config.get("start_urls", [])
        max_urls = crawler_config.get("max_urls", 500)
        
        discovered_urls = []
        
        for start_url in start_urls:
            if len(discovered_urls) >= max_urls:
                break
            
            print(f"从 {start_url} 开始爬取...")
            urls = self.crawl_url(start_url)
            discovered_urls.extend(urls)
            
            if len(discovered_urls) >= max_urls:
                break
        
        return discovered_urls[:max_urls]
    
    def get_alexa_targets(self):
        """从Alexa获取目标"""
        alexa_config = self.config.get("scan_targets", {}).get("sources", {}).get("alexa", {})
        if not alexa_config.get("enabled", False):
            return []
        
        limit = alexa_config.get("limit", 100)
        # 这里可以实现从Alexa API获取目标的逻辑
        # 目前返回空列表
        return []
    
    def get_tranco_targets(self):
        """从Tranco获取目标"""
        tranco_config = self.config.get("scan_targets", {}).get("sources", {}).get("tranco", {})
        if not tranco_config.get("enabled", False):
            return []
        
        limit = tranco_config.get("limit", 100)
        # 这里可以实现从Tranco获取目标的逻辑
        # 目前返回空列表
        return []
    
    def get_all_targets(self):
        """获取所有目标"""
        print("\n=== 获取扫描目标 ===")
        
        all_targets = []
        
        # 从文件获取
        file_targets = self.get_targets_from_file()
        all_targets.extend(file_targets)
        print(f"从文件获取 {len(file_targets)} 个目标")
        
        # 从爬虫获取
        crawl_targets = self.crawl_targets()
        all_targets.extend(crawl_targets)
        print(f"从爬虫获取 {len(crawl_targets)} 个目标")
        
        # 从Alexa获取
        alexa_targets = self.get_alexa_targets()
        all_targets.extend(alexa_targets)
        print(f"从Alexa获取 {len(alexa_targets)} 个目标")
        
        # 从Tranco获取
        tranco_targets = self.get_tranco_targets()
        all_targets.extend(tranco_targets)
        print(f"从Tranco获取 {len(tranco_targets)} 个目标")
        
        # 去重
        unique_targets = list(set(all_targets))
        print(f"去重后总共 {len(unique_targets)} 个目标")
        
        return unique_targets
    
    def scan_targets(self, targets):
        """扫描目标"""
        print(f"\n=== 开始扫描 {len(targets)} 个目标 ===")
        
        crawl_strategy = self.config.get("scan_targets", {}).get("crawl_strategy", {})
        max_concurrent = crawl_strategy.get("max_concurrent", 10)
        
        results = {}
        
        with ThreadPoolExecutor(max_workers=max_concurrent) as executor:
            future_to_url = {executor.submit(self.wappalyzer.scan, url): url for url in targets}
            
            for future in as_completed(future_to_url):
                url = future_to_url[future]
                try:
                    detected = future.result()
                    results[url] = detected
                    
                    # 打印进度
                    print(f"已扫描 {len(results)}/{len(targets)}: {url} - 检测到 {len(detected)} 种技术")
                    
                    # 延迟
                    time.sleep(crawl_strategy.get("delay_between_requests", 1))
                except Exception as e:
                    print(f"扫描 {url} 时出错: {e}")
                    results[url] = {}
        
        return results
    
    def save_results(self, results, filename="scan_results.json"):
        """保存扫描结果"""
        print(f"\n=== 保存扫描结果 ===")
        
        data = {
            "total": len(results),
            "timestamp": time.strftime("%Y-%m-%d %H:%M:%S"),
            "results": results
        }
        
        with open(filename, 'w', encoding='utf-8') as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
        
        print(f"成功保存 {len(results)} 个结果到 {filename}")
    
    def run(self):
        """执行完整的扫描流程"""
        print("=== 基于配置的扫描器开始运行 ===")
        
        # 1. 获取所有目标
        targets = self.get_all_targets()
        
        if not targets:
            print("没有获取到任何目标，扫描结束")
            return
        
        # 2. 执行扫描
        results = self.scan_targets(targets)
        
        # 3. 保存结果
        self.save_results(results)
        
        # 4. 打印摘要
        self.print_summary(results)
        
        print("\n=== 扫描完成 ===")
    
    def print_summary(self, results):
        """打印扫描摘要"""
        print("\n=== 扫描结果摘要 ===")
        total_sites = len(results)
        total_tech = 0
        unique_tech = set()
        
        for url, detected in results.items():
            total_tech += len(detected)
            for tech_name in detected.keys():
                unique_tech.add(tech_name)
        
        print(f"总共扫描 {total_sites} 个网站")
        print(f"总共检测到 {total_tech} 种技术")
        print(f"检测到 {len(unique_tech)} 种不同的技术")
        
        # 统计检测到技术的网站数量
        tech_count = {}
        for url, detected in results.items():
            for tech_name in detected.keys():
                tech_count[tech_name] = tech_count.get(tech_name, 0) + 1
        
        # 打印排名前10的技术
        print("\n排名前10的技术:")
        sorted_tech = sorted(tech_count.items(), key=lambda x: x[1], reverse=True)[:10]
        for tech, count in sorted_tech:
            print(f"- {tech}: {count} 个网站")

if __name__ == "__main__":
    scanner = CrawlScanner()
    scanner.run()
