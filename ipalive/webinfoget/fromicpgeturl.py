
import requests
import csv
import time
import re
from urllib.parse import urljoin, urlparse, unquote
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
        found_domains: Set[str] = set()
        
        # 使用百度搜索（最有效）
        search_url = f"https://www.baidu.com/s?wd={requests.utils.quote(icp_number)}"
        
        try:
            logger.info(f"正在搜索: {search_url}")
            response = self.session.get(search_url, timeout=self.timeout)
            
            # 从搜索结果中提取真实网站链接（不是搜索引擎自己的链接）
            candidate_urls = self.extract_real_website_urls_from_baidu(response.text)
            
            logger.info(f"从搜索结果中提取到 {len(candidate_urls)} 个候选网站")
            
            # 验证每个候选网站是否真的包含该ICP备案号
            for url in candidate_urls:
                parsed = urlparse(url)
                domain = parsed.netloc
                
                # 跳过已处理的域名和搜索引擎域名
                if domain in found_domains or self.is_search_engine_domain(domain):
                    continue
                
                found_domains.add(domain)
                
                # 构建网站首页URL
                website_url = f"{parsed.scheme}://{domain}"
                
                logger.info(f"检查网站: {website_url}")
                
                # 访问网站并验证是否包含该ICP备案号
                if self.verify_icp_on_website(website_url, icp_number):
                    logger.info(f"✓ 找到匹配网站: {website_url}")
                    websites.append({
                        'domain': domain,
                        'icp_number': icp_number,
                        'website_url': website_url,
                        'status': '已验证'
                    })
                else:
                    logger.info(f"✗ 网站未包含该备案号: {website_url}")
                
                time.sleep(1)  # 避免请求过于频繁
                
        except Exception as e:
            logger.error(f"搜索失败: {e}")
        
        return websites
    
    def extract_real_website_urls_from_baidu(self, html_content: str) -> List[str]:
        """从百度搜索结果中提取真实网站链接"""
        urls = []
        
        # 改进的URL提取正则（匹配完整域名）
        url_pattern = r'https?://(?:www\.)?([a-zA-Z0-9][-a-zA-Z0-9]{0,62}\.)+[a-zA-Z]{2,}(?:/[^\s"\'<>]*)?'
        all_urls = re.findall(url_pattern, html_content)
        
        logger.info(f"正则提取到 {len(all_urls)} 个URL")
        
        # 过滤出真实网站链接
        seen_domains = set()
        for match in all_urls:
            # match 是元组，需要重新构建完整URL
            # 重新匹配完整URL
            pass
        
        # 使用更精确的正则
        full_url_pattern = r'(https?://(?:www\.)?[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(?:\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+)'
        all_full_urls = re.findall(full_url_pattern, html_content)
        
        for url in all_full_urls:
            # 清理URL
            url = re.sub(r'[.,;!?)\]]+$', '', url)
            
            if not self.is_valid_url(url):
                continue
                
            parsed = urlparse(url)
            domain = parsed.netloc
            
            # 跳过已处理的域名
            if domain in seen_domains:
                continue
            
            # 跳过搜索引擎和CDN域名
            if self.is_search_engine_domain(domain):
                continue
            
            # 跳过明显的资源文件
            if url.endswith(('.css', '.js', '.png', '.jpg', '.ico', '.svg', '.gif', '.woff', '.ttf')):
                continue
            
            # 构建主域名URL    
            main_url = f"{parsed.scheme}://{domain}"
            if main_url not in urls:
                urls.append(main_url)
                seen_domains.add(domain)
        
        logger.info(f"过滤后剩余 {len(urls)} 个候选网站")
        if urls:
            logger.info(f"候选网站示例: {urls[:3]}")
        
        # 限制数量
        return urls[:30]
    
    def is_search_engine_domain(self, domain: str) -> bool:
        """判断是否为搜索引擎或CDN域名"""
        search_domains = [
            'baidu.com', 'www.baidu.com', 'so.com', 'www.so.com', 
            'sogou.com', 'www.sogou.com', 'google.com', 'bing.com',
            'bcebos.com', 'qhimg.com', 'bdstatic.com', 'ssl.qhimg.com',
            'baidustatic.com', 'info.so.com', 'zhanzhang.so.com'
        ]
        
        for search_domain in search_domains:
            if search_domain in domain:
                return True
        return False
    
    def verify_icp_on_website(self, website_url: str, icp_number: str) -> bool:
        """访问网站并验证页面中是否包含该ICP备案号"""
        try:
            response = self.session.get(website_url, timeout=10, allow_redirects=True)
            
            if response.status_code != 200:
                return False
            
            # 检查页面内容是否包含ICP备案号
            content = response.text.lower()
            icp_clean = icp_number.replace('号', '').lower()
            
            # 多种匹配方式
            if (icp_number.lower() in content or 
                icp_clean in content or
                icp_number.replace('ICP备', 'icp').lower() in content):
                return True
                
            return False
            
        except Exception as e:
            logger.debug(f"验证网站失败 {website_url}: {e}")
            return False
    
    def is_valid_url(self, url: str) -> bool:
        """验证URL有效性"""
        try:
            result = urlparse(url)
            return all([result.scheme, result.netloc]) and result.scheme in ['http', 'https']
        except:
            return False
    
    def verify_website(self, website_url: str) -> str:
        """验证网站是否可访问（简化版，主要用于兼容）"""
        return "已验证"
    
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
