
import requests
import csv
import time
import re
from urllib.parse import urljoin, urlparse
import argparse
from typing import List, Set
import logging

# 配置日志
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

class ICPWebsiteFinder:
    def __init__(self, timeout: int = 10):
        self.timeout = timeout
        self.session = requests.Session()
        self.session.headers.update({
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36'
        })
        
    def find_websites_by_icp(self, icp_number: str) -> List[dict]:
        """通过ICP备案号查找使用该备案号的网站"""
        websites = []
        
        # 使用多个搜索引擎查询ICP备案号
        search_engines = [
            "https://www.baidu.com/s?wd=",
            "https://www.so.com/s?q=",
            "https://www.sogou.com/web?query="
        ]
        
        # 构造查询关键词
        queries = [
            f"\"{icp_number}\"",  # 精确匹配
            f"备案号 {icp_number}",
            f"ICP {icp_number}",
            f"网站备案 {icp_number}"
        ]
        
        found_urls: Set[str] = set()
        
        for engine in search_engines:
            for query in queries:
                try:
                    search_url = engine + requests.utils.quote(query)
                    logger.info(f"正在搜索: {search_url}")
                    
                    response = self.session.get(search_url, timeout=self.timeout)
                    urls = self.extract_urls(response.text, icp_number)
                    
                    for url in urls:
                        if url not in found_urls:
                            found_urls.add(url)
                            domain = urlparse(url).netloc
                            websites.append({
                                'domain': domain,
                                'icp_number': icp_number,
                                'website_url': url,
                                'status': '待验证'
                            })
                            
                    time.sleep(1)  # 避免请求过于频繁
                    
                except Exception as e:
                    logger.error(f"搜索失败 {engine + query}: {e}")
                    continue
        
        # 验证网站有效性
        logger.info(f"发现 {len(websites)} 个网站，正在验证有效性...")
        for website in websites:
            website['status'] = self.verify_website(website['website_url'])
            
        return websites
    
    def extract_urls(self, html_content: str, icp_number: str) -> List[str]:
        """从HTML内容中提取与ICP备案号相关的网址"""
        urls = []
        
        # 匹配http/https链接
        url_pattern = r'https?://[^\s"\'<>]+'
        all_urls = re.findall(url_pattern, html_content)
        
        # 过滤和清理URL
        clean_urls = []
        for url in all_urls:
            # 移除末尾的标点符号
            url = re.sub(r'[.,;!?]+$', '', url)
            if self.is_valid_url(url):
                clean_urls.append(url)
        
        # 进一步筛选可能与ICP相关的URL
        for url in clean_urls:
            # 检查URL所在的上下文是否包含ICP信息
            context_start = max(0, html_content.find(url) - 200)
            context_end = min(len(html_content), html_content.find(url) + len(url) + 200)
            context = html_content[context_start:context_end]
            
            # 如果上下文中包含ICP相关信息，则认为是相关网站
            if icp_number.replace("号", "") in context or "备案" in context:
                urls.append(url)
        
        # 如果没有找到相关URL，则返回所有有效URL（备选方案）
        if not urls:
            urls = clean_urls[:20]  # 限制数量避免过多
            
        return urls
    
    def is_valid_url(self, url: str) -> bool:
        """验证URL有效性"""
        try:
            result = urlparse(url)
            return all([result.scheme, result.netloc]) and result.scheme in ['http', 'https']
        except:
            return False
    
    def verify_website(self, website_url: str) -> str:
        """验证网站是否可访问"""
        try:
            response = self.session.head(website_url, timeout=5, allow_redirects=True)
            if response.status_code == 200:
                return "正常访问"
            else:
                return f"状态码:{response.status_code}"
        except:
            try:
                # 如果HEAD请求失败，尝试GET请求
                response = self.session.get(website_url, timeout=5)
                if response.status_code == 200:
                    return "正常访问"
                else:
                    return f"状态码:{response.status_code}"
            except:
                return "无法访问"
    
    def deduplicate_websites(self, websites: List[dict]) -> List[dict]:
        """去重网站列表"""
        seen_urls = set()
        unique_websites = []
        
        for website in websites:
            if website['website_url'] not in seen_urls:
                seen_urls.add(website['website_url'])
                unique_websites.append(website)
                
        return unique_websites
    
    def save_to_csv(self, websites: List[dict], filename: str):
        """保存结果到CSV文件"""
        with open(filename, 'w', newline='', encoding='utf-8-sig') as csvfile:
            fieldnames = ['域名', 'ICP备案号', '网站URL', '状态']
            writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
            
            writer.writeheader()
            for website in websites:
                writer.writerow({
                    '域名': website['domain'],
                    'ICP备案号': website['icp_number'],
                    '网站URL': website['website_url'],
                    '状态': website['status']
                })

def load_icp_numbers_from_file(filename: str) -> List[str]:
    """从文件加载ICP备案号"""
    icp_numbers = []
    try:
        with open(filename, 'r', encoding='utf-8') as file:
            for line in file:
                line = line.strip()
                if line and not line.startswith('#'):
                    icp_numbers.append(line)
    except FileNotFoundError:
        print(f"文件 {filename} 未找到")
    except Exception as e:
        print(f"读取文件失败: {e}")
    
    return icp_numbers

def main():
    parser = argparse.ArgumentParser(description='通过ICP备案信息搜索相关网站')
    parser.add_argument('-icp', '--icp_number', help='ICP备案号')
    parser.add_argument('-i', '--input', help='包含多个ICP备案号的文件')
    parser.add_argument('-o', '--output', default='icp_websites.csv', help='输出CSV文件名')
    parser.add_argument('-t', '--timeout', type=int, default=10, help='请求超时时间(秒)')
    
    args = parser.parse_args()
    
    # 检查参数
    if not args.icp_number and not args.input:
        print("请提供ICP备案号或包含ICP备案号的文件")
        print("用法:")
        print("  单个ICP查询: python icp_website_finder.py -icp=京ICP备12345678号")
        print("  批量查询: python icp_website_finder.py -i=icp_list.txt")
        return
    
    # 获取ICP备案号列表
    icp_numbers = []
    if args.icp_number:
        icp_numbers = [args.icp_number]
    else:
        icp_numbers = load_icp_numbers_from_file(args.input)
    
    if not icp_numbers:
        print("未提供有效的ICP备案号")
        return
    
    # 创建查找器
    finder = ICPWebsiteFinder(timeout=args.timeout)
    
    # 处理每个ICP备案号
    all_websites = []
    
    for icp in icp_numbers:
        print(f"正在查询ICP备案号: {icp}")
        
        # 查找网站
        print("正在查找相关网站...")
        websites = finder.find_websites_by_icp(icp)
        
        # 去重
        websites = finder.deduplicate_websites(websites)
        
        # 添加到总结果
        all_websites.extend(websites)
        
        print(f"ICP {icp} 查询完成，找到 {len(websites)} 个网站")
        time.sleep(2)  # 避免请求过于频繁
    
    # 保存结果
    if all_websites:
        print(f"总共找到 {len(all_websites)} 个网站，正在保存到 {args.output}...")
        finder.save_to_csv(all_websites, args.output)
        print("查询完成! 结果已保存到CSV文件")
        
        # 显示统计信息
        valid_count = sum(1 for website in all_websites if website['status'] == '正常访问')
        print(f"统计信息: 总计 {len(all_websites)} 个网站，有效 {valid_count} 个")
    else:
        print("未找到任何相关网站")

if __name__ == "__main__":
    main()
