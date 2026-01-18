#!/usr/bin/env python3
import json
import os
import requests
import logging
import re
from datetime import datetime
import random
from urllib.parse import urlparse

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('smart_targets.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class SmartTargetGenerator:
    def __init__(self, config=None):
        self.config = config or {
            'max_targets': 100,
            'validation_timeout': 5,
            'sources': {
                'alexa': True,
                'ct_logs': False,
                'search_engines': False,
                'dns_enum': False,
                'subdomains': False
            },
            'filters': {
                'tld': [],  # 允许的顶级域名，空列表表示允许所有
                'country': [],  # 允许的国家/地区代码
                'min_estimated_visits': 0,  # 最低预估访问量
                'tech_category': [],  # 筛选特定技术类别，如['CMS']
                'industry': []  # 筛选特定行业，如['ecommerce', 'education', 'finance']
            }
        }
        
    def validate_domain(self, domain):
        """验证域名是否有效且可访问"""
        if not domain:
            return False
        
        # 基本格式验证
        domain_regex = r'^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$'
        if not re.match(domain_regex, domain):
            logger.debug(f"域名 {domain} 格式无效")
            return False
        
        # 尝试访问验证
        try:
            urls = [f"https://{domain}", f"http://{domain}"]
            for url in urls:
                response = requests.get(url, timeout=self.config['validation_timeout'], allow_redirects=True)
                if response.status_code < 400:
                    logger.debug(f"域名 {domain} 可访问")
                    return True
            logger.debug(f"域名 {domain} 无法访问")
            return False
        except Exception as e:
            logger.debug(f"验证域名 {domain} 时出错：{e}")
            return False
    
    def get_alexa_top_sites(self, count=100):
        """从Alexa获取顶级网站"""
        logger.info(f"从Alexa获取前 {count} 个顶级网站...")
        targets = []
        
        # 注意：Alexa API需要API密钥，这里使用一个示例数据源
        # 实际应用中，应该使用官方API或其他可靠数据源
        sample_urls = [
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
            "https://www.bing.com"
        ]
        
        # 随机选择指定数量的域名
        selected = random.sample(sample_urls, min(count, len(sample_urls)))
        
        # 验证域名
        for url in selected:
            domain = urlparse(url).netloc
            if self.validate_domain(domain):
                targets.append(url)
        
        logger.info(f"从Alexa成功获取 {len(targets)} 个有效网站")
        return targets
    
    def get_tranco_list(self, count=100):
        """从Tranco列表获取顶级网站"""
        logger.info(f"从Tranco列表获取前 {count} 个顶级网站...")
        targets = []
        
        # Tranco列表URL，每天更新
        tranco_url = "https://tranco-list.eu/download/latest/list"
        
        try:
            response = requests.get(tranco_url, timeout=30)
            response.raise_for_status()
            
            lines = response.text.strip().split('\n')
            for line in lines[:count]:
                if ',' in line:
                    rank, domain = line.split(',', 1)
                    domain = domain.strip()
                    if self.validate_domain(domain):
                        targets.append(f"https://{domain}")
        except Exception as e:
            logger.error(f"获取Tranco列表失败：{e}")
        
        logger.info(f"从Tranco列表成功获取 {len(targets)} 个有效网站")
        return targets
    
    def get_subdomains(self, domain, count=50):
        """获取指定域名的子域名"""
        logger.info(f"获取域名 {domain} 的子域名...")
        targets = []
        
        # 这里使用一个简单的子域名列表作为示例
        # 实际应用中，可以使用更复杂的子域名枚举技术
        common_subdomains = [
            'www', 'blog', 'api', 'dev', 'test', 'staging', 'admin', 'mail',
            'shop', 'store', 'forum', 'support', 'docs', 'help', 'status',
            'portal', 'app', 'mobile', 'm', 'cdn', 'static', 'media', 'images',
            'video', 'audio', 'files', 'download', 'upload', 'backup', 'monitor'
        ]
        
        # 生成子域名并验证
        for sub in common_subdomains[:count]:
            subdomain = f"{sub}.{domain}"
            if self.validate_domain(subdomain):
                targets.append(f"https://{subdomain}")
        
        logger.info(f"为域名 {domain} 成功获取 {len(targets)} 个有效子域名")
        return targets
    
    def get_related_domains(self, seed_domain, count=50):
        """获取与指定域名相关的域名"""
        logger.info(f"获取与 {seed_domain} 相关的域名...")
        targets = []
        
        # 这里使用一个简单的示例实现
        # 实际应用中，可以通过反向链接、共同IP等方式获取相关域名
        # 我们这里只是返回一些示例域名
        sample_related = [
            "https://www.example.com",
            "https://www.example.org",
            "https://www.sample.com",
            "https://www.sample.org",
            "https://www.demo.com",
            "https://www.demo.org"
        ]
        
        # 验证域名
        for url in sample_related:
            domain = urlparse(url).netloc
            if self.validate_domain(domain):
                targets.append(url)
        
        logger.info(f"为域名 {seed_domain} 成功获取 {len(targets)} 个相关域名")
        return targets
    
    def filter_targets(self, targets):
        """根据配置筛选目标"""
        logger.info(f"开始筛选 {len(targets)} 个目标...")
        filtered = []
        
        for target in targets:
            domain = urlparse(target).netloc
            tld = domain.split('.')[-1]
            
            # 筛选顶级域名
            if self.config['filters']['tld'] and tld not in self.config['filters']['tld']:
                continue
            
            filtered.append(target)
        
        logger.info(f"筛选后剩余 {len(filtered)} 个目标")
        return filtered
    
    def get_cms_websites(self, count=50):
        """获取CMS网站列表"""
        logger.info(f"开始获取 {count} 个CMS网站...")
        
        # 在实际应用中，这里可以：
        # 1. 使用搜索引擎API搜索CMS相关网站
        # 2. 从已知CMS网站列表中获取
        # 3. 使用DNS枚举查找使用特定CMS的网站
        
        # 示例CMS网站列表
        cms_websites = [
            "https://wordpress.org",
            "https://joomla.org",
            "https://drupal.org",
            "https://magento.com",
            "https://shopify.com",
            "https://wix.com",
            "https://squarespace.com",
            "https://blogger.com",
            "https://medium.com",
            "https://ghost.org",
            "https://typepad.com",
            "https://weebly.com",
            "https://jimdo.com",
            "https://mozilla.org",
            "https://apache.org"
        ]
        
        # 验证并筛选网站
        valid_cms_websites = []
        for url in cms_websites[:count]:
            domain = urlparse(url).netloc
            if self.validate_domain(domain):
                valid_cms_websites.append(url)
        
        logger.info(f"成功获取 {len(valid_cms_websites)} 个有效CMS网站")
        return valid_cms_websites
    
    def get_industry_websites(self, industry, count=50):
        """根据行业获取网站"""
        logger.info(f"开始获取 {count} 个{industry}行业网站...")
        
        # 行业网站映射表
        industry_websites = {
            'ecommerce': [
                "https://www.amazon.com",
                "https://www.taobao.com",
                "https://www.jd.com",
                "https://www.tmall.com",
                "https://www.alibaba.com",
                "https://www.ebay.com",
                "https://www.walmart.com",
                "https://www.target.com",
                "https://www.bestbuy.com",
                "https://www.newegg.com",
                "https://www.etsy.com",
                "https://www.shopify.com",
                "https://www.rakuten.com",
                "https://www.lazada.com",
                "https://www.flipkart.com"
            ],
            'education': [
                "https://www.mooc.org",
                "https://www.coursera.org",
                "https://www.udemy.com",
                "https://www.khanacademy.org",
                "https://www.edx.org",
                "https://www.coursera.org",
                "https://www.udacity.com",
                "https://www.open.edu",
                "https://www.saylor.org",
                "https://www.cambridge.org",
                "https://www.oxforduniversitypress.com",
                "https://www.princeton.edu",
                "https://www.harvard.edu",
                "https://www.mit.edu",
                "https://www.stanford.edu"
            ],
            'finance': [
                "https://www.jpmorgan.com",
                "https://www.chase.com",
                "https://www.bankofamerica.com",
                "https://www.wellsfargo.com",
                "https://www.citibank.com",
                "https://www.americanexpress.com",
                "https://www.mastercard.com",
                "https://www.visa.com",
                "https://www.paypal.com",
                "https://www.coinbase.com",
                "https://www.bloomberg.com",
                "https://www.reuters.com",
                "https://www.marketwatch.com",
                "https://www.fidelity.com",
                "https://www.schwab.com"
            ],
            'healthcare': [
                "https://www.mayoclinic.org",
                "https://www.webmd.com",
                "https://www.cdc.gov",
                "https://www.nih.gov",
                "https://www.who.int",
                "https://www.healthline.com",
                "https://www.medicalnewstoday.com",
                "https://www.merckmanuals.com",
                "https://www.cancer.org",
                "https://www.heart.org",
                "https://www.diabetes.org",
                "https://www.alz.org",
                "https://www.nih.gov",
                "https://www.fda.gov",
                "https://www.medscape.com"
            ],
            'technology': [
                "https://www.google.com",
                "https://www.microsoft.com",
                "https://www.apple.com",
                "https://www.amazon.com",
                "https://www.facebook.com",
                "https://www.twitter.com",
                "https://www.linkedin.com",
                "https://www.github.com",
                "https://www.stackoverflow.com",
                "https://www.reddit.com",
                "https://www.techcrunch.com",
                "https://www.wired.com",
                "https://www.theverge.com",
                "https://www.arstechnica.com",
                "https://www.cnet.com"
            ]
        }
        
        valid_industry_websites = []
        if industry in industry_websites:
            # 从行业网站列表中选择
            websites = industry_websites[industry]
            
            # 随机选择指定数量
            if len(websites) > count:
                websites = random.sample(websites, count)
            
            # 验证网站
            for url in websites:
                domain = urlparse(url).netloc
                if self.validate_domain(domain):
                    valid_industry_websites.append(url)
        else:
            # 对于未知行业，使用默认的目标生成
            logger.warning(f"未知行业: {industry}, 使用默认目标生成")
            valid_industry_websites = self.generate_targets(count)
        
        logger.info(f"成功获取 {len(valid_industry_websites)} 个{industry}行业网站")
        return valid_industry_websites
    
    def get_websites_by_tech_category(self, category, count=50, industry=None):
        """根据技术类别获取网站"""
        logger.info(f"开始获取 {count} 个{category}类别网站...")
        
        # 根据类别获取网站
        if category.lower() == 'cms':
            cms_websites = self.get_cms_websites(count)
            # 如果指定了行业，进一步筛选
            if industry:
                logger.info(f"对CMS网站进行{industry}行业筛选...")
                industry_sites = set(self.get_industry_websites(industry, count * 2))
                filtered = list(set(cms_websites) & industry_sites)
                if filtered:
                    return filtered[:count]
                else:
                    return cms_websites[:count]
            return cms_websites
        else:
            # 对于其他类别，使用默认的目标生成
            return self.generate_targets(count)
    
    def get_targets_from_file(self, file_path='education_sites.json', count=None):
        """从文件中读取目标网站"""
        logger.info(f"从文件 {file_path} 中读取目标网站...")
        targets = []
        
        if os.path.exists(file_path):
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    data = json.load(f)
                    
                # 检查文件格式
                if isinstance(data, dict) and 'sites' in data:
                    targets = data['sites']
                elif isinstance(data, list):
                    targets = data
                
                logger.info(f"从文件 {file_path} 中读取到 {len(targets)} 个网站")
                
                # 如果指定了数量，只返回前count个
                if count and isinstance(count, int) and count > 0:
                    targets = targets[:count]
                    logger.info(f"返回前 {count} 个网站")
                
            except Exception as e:
                logger.error(f"读取文件 {file_path} 时出错：{e}")
        else:
            logger.warning(f"文件 {file_path} 不存在")
        
        return targets
    
    def generate_targets(self, count=100, industry=None):
        """智能生成扫描目标"""
        logger.info(f"开始智能生成 {count} 个扫描目标...")
        all_targets = []
        
        # 1. 从文件中读取目标网站（优先）
        file_targets = self.get_targets_from_file('education_sites.json', count * 2)
        all_targets.extend(file_targets)
        
        # 2. 如果指定了行业，从行业网站获取
        if industry:
            logger.info(f"根据行业 {industry} 获取网站...")
            industry_targets = self.get_industry_websites(industry, count * 2)
            all_targets.extend(industry_targets)
        
        # 3. 从不同来源获取目标
        if self.config['sources']['alexa']:
            all_targets.extend(self.get_alexa_top_sites(count))
        
        # 4. 去重
        all_targets = list(set(all_targets))
        
        # 5. 筛选目标
        all_targets = self.filter_targets(all_targets)
        
        # 6. 限制数量
        if len(all_targets) > count:
            all_targets = random.sample(all_targets, count)
        
        # 7. 如果数量不足，使用其他来源补充
        if len(all_targets) < count:
            logger.warning(f"当前目标数量不足 {count}，正在补充...")
            # 可以添加更多的目标来源，例如：
            # - 使用内置的行业网站列表
            # - 生成随机域名
            # - 使用其他API获取
            
            # 这里使用内置的行业网站作为补充
            additional_targets = []
            for industry_name in ['education', 'technology', 'ecommerce']:
                additional_targets.extend(self.get_industry_websites(industry_name, count))
            
            # 去重并合并
            all_targets = list(set(all_targets + additional_targets))
            
            # 再次限制数量
            if len(all_targets) > count:
                all_targets = random.sample(all_targets, count)
        
        logger.info(f"成功生成 {len(all_targets)} 个扫描目标")
        return all_targets

if __name__ == "__main__":
    # 导入urlparse
    from urllib.parse import urlparse
    
    # 创建智能目标生成器实例
    generator = SmartTargetGenerator()
    
    # 生成10个扫描目标
    targets = generator.generate_targets(10)
    
    # 打印结果
    print("=== 智能生成的扫描目标 ===")
    for i, target in enumerate(targets, 1):
        print(f"{i}. {target}")
