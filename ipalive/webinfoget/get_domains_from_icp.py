#!/usr/bin/env python3
"""
通过ICP备案号获取相关域名
使用多个公开的ICP查询接口
"""

import requests
import re
import json
from urllib.parse import quote
import logging

logging.basicConfig(level=logging.INFO, format='%(message)s')
logger = logging.getLogger(__name__)


def query_icp_chinaz(icp_number: str):
    """站长工具ICP查询"""
    logger.info(f"\n[方法1] 站长工具查询...")
    
    try:
        # 注意：这个网站可能需要在浏览器中手动查询
        url = f"https://icp.chinaz.com/{quote(icp_number)}"
        logger.info(f"URL: {url}")
        logger.info("提示：此网站可能需要手动访问查询")
        
        headers = {
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
        }
        
        response = requests.get(url, headers=headers, timeout=10)
        
        if response.status_code == 200:
            # 尝试从页面提取域名
            domains = re.findall(r'(?:https?://)?([a-zA-Z0-9][-a-zA-Z0-9]{0,62}(?:\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+)', response.text)
            
            # 过滤明显不相关的域名
            filtered = []
            exclude_domains = ['chinaz.com', 'bdstatic.com', 'baidu.com', 'google', 'javascript']
            
            for domain in set(domains):
                if not any(ex in domain for ex in exclude_domains):
                    filtered.append(domain)
            
            if filtered:
                logger.info(f"找到 {len(filtered)} 个候选域名")
                return filtered[:10]
            else:
                logger.info("未找到域名")
        
    except Exception as e:
        logger.error(f"查询失败: {e}")
    
    return []


def query_icp_aizhan(icp_number: str):
    """爱站ICP查询"""
    logger.info(f"\n[方法2] 爱站工具查询...")
    
    try:
        url = f"https://icp.aizhan.com/{quote(icp_number)}/"
        logger.info(f"URL: {url}")
        
        headers = {
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
        }
        
        response = requests.get(url, headers=headers, timeout=10)
        
        if response.status_code == 200:
            # 提取域名
            domains = re.findall(r'<td[^>]*>([a-zA-Z0-9][-a-zA-Z0-9]{0,62}\.[a-zA-Z]{2,})</td>', response.text)
            
            if domains:
                logger.info(f"找到 {len(domains)} 个域名")
                return list(set(domains))[:10]
            else:
                logger.info("未找到域名")
        
    except Exception as e:
        logger.error(f"查询失败: {e}")
    
    return []


def query_beian_query(icp_number: str):
    """备案查询网站"""
    logger.info(f"\n[方法3] 备案查询网...")
    
    try:
        # 多个备案查询网站
        urls = [
            f"http://www.beianbeian.com/search/{quote(icp_number)}",
            f"https://www.tianyancha.com/search?key={quote(icp_number)}"
        ]
        
        for url in urls:
            logger.info(f"尝试: {url}")
            
            headers = {
                'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
            }
            
            response = requests.get(url, headers=headers, timeout=10)
            
            if response.status_code == 200:
                # 提取域名
                domains = re.findall(r'(?:https?://)?([a-zA-Z0-9][-a-zA-Z0-9]{0,62}\.[a-zA-Z]{2,})', response.text)
                
                # 过滤
                filtered = []
                exclude = ['beianbeian.com', 'tianyancha.com', 'baidu.com', 'google', 'cnzz.com']
                
                for domain in set(domains):
                    if not any(ex in domain for ex in exclude) and '.' in domain:
                        filtered.append(domain)
                
                if filtered:
                    logger.info(f"找到 {len(filtered)} 个候选域名")
                    return filtered[:10]
        
    except Exception as e:
        logger.error(f"查询失败: {e}")
    
    return []


def search_engine_extract(icp_number: str):
    """从搜索引擎结果提取"""
    logger.info(f"\n[方法4] 搜索引擎提取...")
    
    try:
        # 使用Google搜索（在中国可能需要代理）
        search_url = f"https://www.google.com/search?q={quote(icp_number)}"
        
        headers = {
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
        }
        
        response = requests.get(search_url, headers=headers, timeout=10)
        
        if response.status_code == 200:
            # 提取域名
            domains = re.findall(r'https?://([a-zA-Z0-9][-a-zA-Z0-9]{0,62}\.[a-zA-Z]{2,})', response.text)
            
            # 过滤
            filtered = []
            exclude = ['google.com', 'gstatic.com', 'youtube.com']
            
            for domain in set(domains):
                if not any(ex in domain for ex in exclude):
                    filtered.append(domain)
            
            if filtered:
                logger.info(f"找到 {len(filtered)} 个候选域名")
                return filtered[:10]
        
    except Exception as e:
        logger.error(f"查询失败: {e}")
    
    return []


def main():
    import sys
    
    if len(sys.argv) < 2:
        print("用法: python3 get_domains_from_icp.py <ICP备案号>")
        print("示例: python3 get_domains_from_icp.py 闽ICP备06031865号")
        sys.exit(1)
    
    icp_number = sys.argv[1]
    
    print("=" * 70)
    print(f"查询ICP备案号: {icp_number}")
    print("=" * 70)
    
    all_domains = []
    
    # 尝试多种方法
    methods = [
        query_icp_chinaz,
        query_icp_aizhan,
        query_beian_query,
        search_engine_extract
    ]
    
    for method in methods:
        try:
            domains = method(icp_number)
            if domains:
                all_domains.extend(domains)
        except Exception as e:
            logger.error(f"方法 {method.__name__} 执行失败: {e}")
    
    # 去重
    unique_domains = list(set(all_domains))
    
    print("\n" + "=" * 70)
    print("查询结果汇总")
    print("=" * 70)
    
    if unique_domains:
        print(f"\n找到 {len(unique_domains)} 个候选域名:\n")
        for i, domain in enumerate(unique_domains, 1):
            print(f"{i}. {domain}")
        
        # 保存到文件
        output_file = "domains_found.txt"
        with open(output_file, 'w', encoding='utf-8') as f:
            f.write(f"# ICP备案号: {icp_number}\n")
            f.write(f"# 查询时间: {__import__('datetime').datetime.now()}\n\n")
            for domain in unique_domains:
                f.write(f"{domain}\n")
        
        print(f"\n域名列表已保存到: {output_file}")
        print(f"\n下一步: 使用以下命令验证这些域名")
        print(f"python3 icp_verify.py -icp={icp_number} -f={output_file} -o=verified_result.csv")
        
    else:
        print("\n未找到任何域名")
        print("\n建议:")
        print("1. 手动访问工信部备案系统查询: https://beian.miit.gov.cn/")
        print("2. 访问站长工具查询: https://icp.chinaz.com/")
        print("3. 访问爱站查询: https://icp.aizhan.com/")
        print(f"4. 手动创建 domains.txt 文件，每行一个域名，然后运行:")
        print(f"   python3 icp_verify.py -icp={icp_number} -f=domains.txt")


if __name__ == "__main__":
    main()
