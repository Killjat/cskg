#!/usr/bin/env python3
"""
验证网站是否包含指定的ICP备案号
可以输入候选域名列表进行批量验证
"""

import requests
import csv
import argparse
from typing import List, Dict
import logging

logging.basicConfig(level=logging.INFO, format='%(message)s')
logger = logging.getLogger(__name__)


class ICPVerifier:
    def __init__(self, timeout: int = 10):
        self.timeout = timeout
        self.session = requests.Session()
        self.session.headers.update({
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
        })
    
    def verify_domain(self, domain: str, icp_number: str) -> Dict:
        """验证单个域名是否包含指定的ICP备案号"""
        # 尝试多种URL格式
        urls_to_try = [
            f"https://{domain}",
            f"http://{domain}",
            f"https://www.{domain}",
            f"http://www.{domain}"
        ]
        
        for url in urls_to_try:
            try:
                logger.info(f"正在检查: {url}")
                response = self.session.get(url, timeout=self.timeout, allow_redirects=True)
                
                if response.status_code == 200:
                    # 检查页面内容
                    content = response.text
                    
                    # 打印部分内容用于调试
                    if 'icp' in content.lower() or '备案' in content:
                        logger.debug(f"页面包含备案相关信息")
                    
                    # 多种匹配方式
                    icp_variations = [
                        icp_number,
                        icp_number.replace('号', ''),
                        icp_number.lower(),
                        icp_number.replace('ICP备', 'ICP'),
                        icp_number.replace('ICP备', 'icp备'),
                        icp_number.replace('ICP证', 'ICP'),
                    ]
                    
                    for variant in icp_variations:
                        if variant in content or variant.lower() in content.lower():
                            logger.info(f"✓ 找到匹配! {url} 包含 {variant}")
                            return {
                                'domain': domain,
                                'icp_number': icp_number,
                                'website_url': response.url,  # 使用最终URL（处理重定向）
                                'status': '已验证',
                                'match_variant': variant
                            }
                
                logger.info(f"  状态码: {response.status_code}, 未找到备案号")
                
            except requests.exceptions.Timeout:
                logger.warning(f"  超时: {url}")
            except requests.exceptions.ConnectionError:
                logger.warning(f"  连接失败: {url}")
            except Exception as e:
                logger.warning(f"  错误: {url} - {str(e)[:50]}")
        
        logger.info(f"✗ {domain} 未找到备案号")
        return None
    
    def verify_domains_from_list(self, domains: List[str], icp_number: str) -> List[Dict]:
        """批量验证域名列表"""
        results = []
        
        logger.info(f"\n开始验证 {len(domains)} 个域名是否包含 ICP备案号: {icp_number}\n")
        logger.info("=" * 70)
        
        for i, domain in enumerate(domains, 1):
            logger.info(f"\n[{i}/{len(domains)}] 验证域名: {domain}")
            logger.info("-" * 70)
            
            result = self.verify_domain(domain.strip(), icp_number)
            if result:
                results.append(result)
                logger.info(f"✓ 成功: {result['website_url']}")
            
            logger.info("")
        
        return results
    
    def save_to_csv(self, results: List[Dict], filename: str):
        """保存结果到CSV"""
        if not results:
            logger.warning("没有结果可保存")
            return
        
        with open(filename, 'w', newline='', encoding='utf-8-sig') as f:
            fieldnames = ['域名', 'ICP备案号', '网站URL', '状态', '匹配方式']
            writer = csv.DictWriter(f, fieldnames=fieldnames)
            writer.writeheader()
            
            for result in results:
                writer.writerow({
                    '域名': result['domain'],
                    'ICP备案号': result['icp_number'],
                    '网站URL': result['website_url'],
                    '状态': result['status'],
                    '匹配方式': result.get('match_variant', '')
                })
        
        logger.info(f"结果已保存到: {filename}")


def load_domains_from_file(filename: str) -> List[str]:
    """从文件加载域名列表"""
    domains = []
    try:
        with open(filename, 'r', encoding='utf-8') as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith('#'):
                    domains.append(line)
    except FileNotFoundError:
        logger.error(f"文件未找到: {filename}")
    except Exception as e:
        logger.error(f"读取文件失败: {e}")
    
    return domains


def main():
    parser = argparse.ArgumentParser(description='验证域名是否包含指定的ICP备案号')
    parser.add_argument('-icp', '--icp_number', required=True, help='ICP备案号')
    parser.add_argument('-d', '--domain', help='单个域名')
    parser.add_argument('-f', '--file', help='域名列表文件（每行一个域名）')
    parser.add_argument('-o', '--output', default='verified_icp.csv', help='输出CSV文件名')
    parser.add_argument('-t', '--timeout', type=int, default=10, help='超时时间(秒)')
    
    args = parser.parse_args()
    
    # 获取域名列表
    domains = []
    if args.domain:
        domains = [args.domain]
    elif args.file:
        domains = load_domains_from_file(args.file)
    else:
        print("错误: 请提供域名 (-d) 或域名列表文件 (-f)")
        print("\n用法示例:")
        print("  单个域名: python3 icp_verify.py -icp=闽ICP备06031865号 -d=example.com")
        print("  批量验证: python3 icp_verify.py -icp=闽ICP备06031865号 -f=domains.txt")
        return
    
    if not domains:
        logger.error("没有要验证的域名")
        return
    
    # 创建验证器
    verifier = ICPVerifier(timeout=args.timeout)
    
    # 执行验证
    results = verifier.verify_domains_from_list(domains, args.icp_number)
    
    # 输出结果
    print("\n" + "=" * 70)
    print("验证完成!")
    print("=" * 70)
    
    if results:
        print(f"\n找到 {len(results)} 个包含该ICP备案号的网站:\n")
        for i, result in enumerate(results, 1):
            print(f"{i}. {result['domain']}")
            print(f"   URL: {result['website_url']}")
            print(f"   匹配: {result.get('match_variant', '')}")
            print()
        
        # 保存结果
        verifier.save_to_csv(results, args.output)
    else:
        print(f"\n未找到包含ICP备案号 '{args.icp_number}' 的网站")
        print("\n可能原因:")
        print("1. 提供的域名列表中没有使用该备案号的网站")
        print("2. 网站无法访问或已关闭")
        print("3. 网站页面中没有显示备案号")


if __name__ == "__main__":
    main()
