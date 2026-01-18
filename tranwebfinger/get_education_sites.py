#!/usr/bin/env python3
import json
import os
import requests
import logging
import re
from datetime import datetime
import random
from urllib.parse import urlparse
from smart_targets import SmartTargetGenerator

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('get_education_sites.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class EducationSiteCollector:
    def __init__(self):
        self.generator = SmartTargetGenerator()
        self.education_domains = []
        self.visited = set()
    
    def add_education_domains(self, domains):
        """添加教育域名到列表"""
        for domain in domains:
            if domain not in self.visited:
                self.visited.add(domain)
                self.education_domains.append(domain)
    
    def get_edu_tld_sites(self, count=200):
        """获取.edu顶级域名网站"""
        logger.info(f"开始获取 {count} 个.edu域名网站...")
        
        # 使用Tranco列表获取.edu域名
        tranco_url = "https://tranco-list.eu/download/latest/list"
        edu_sites = []
        
        try:
            response = requests.get(tranco_url, timeout=30)
            response.raise_for_status()
            
            lines = response.text.strip().split('\n')
            for line in lines:
                if ',' in line:
                    rank, domain = line.split(',', 1)
                    domain = domain.strip()
                    if domain.endswith('.edu') and self.generator.validate_domain(domain):
                        edu_sites.append(f"https://{domain}")
                        if len(edu_sites) >= count:
                            break
        except Exception as e:
            logger.error(f"获取.edu域名失败：{e}")
        
        logger.info(f"成功获取 {len(edu_sites)} 个.edu域名网站")
        return edu_sites
    
    def get_more_education_sites(self, count=300):
        """通过多种方式获取更多教育网站"""
        logger.info(f"开始获取 {count} 个额外教育网站...")
        
        # 教育相关关键词
        education_keywords = [
            "university", "college", "school", "academy", "institute",
            "education", "learning", "course", "study", "campus",
            "classroom", "curriculum", "degree", "diploma", "certificate"
        ]
        
        # 已知教育网站后缀
        edu_suffixes = [".edu", ".ac", ".edu.cn", ".ac.uk", ".edu.au", ".edu.tw"]
        
        # 从Tranco列表获取可能的教育网站
        tranco_url = "https://tranco-list.eu/download/latest/list"
        more_sites = []
        
        try:
            response = requests.get(tranco_url, timeout=30)
            response.raise_for_status()
            
            lines = response.text.strip().split('\n')
            for line in lines[:count * 10]:  # 检查更多行
                if ',' in line:
                    rank, domain = line.split(',', 1)
                    domain = domain.strip()
                    
                    # 检查域名是否包含教育关键词或后缀
                    domain_lower = domain.lower()
                    if any(suffix in domain_lower for suffix in edu_suffixes) or \
                       any(keyword in domain_lower for keyword in education_keywords):
                        if self.generator.validate_domain(domain):
                            more_sites.append(f"https://{domain}")
                            if len(more_sites) >= count:
                                break
        except Exception as e:
            logger.error(f"获取额外教育网站失败：{e}")
        
        logger.info(f"成功获取 {len(more_sites)} 个额外教育网站")
        return more_sites
    
    def get_education_portals(self, count=200):
        """获取教育门户和平台网站"""
        logger.info(f"开始获取 {count} 个教育门户和平台网站...")
        
        # 教育平台列表
        education_platforms = [
            "https://www.coursera.org", "https://www.udemy.com", "https://www.khanacademy.org",
            "https://www.edx.org", "https://www.udacity.com", "https://www.open.edu",
            "https://www.saylor.org", "https://www.mooc.org", "https://www.futurelearn.com",
            "https://www.alison.com", "https://www.coursera.org", "https://www.pluralsight.com",
            "https://www.codecademy.com", "https://www.datacamp.com", "https://www.lynda.com",
            "https://www.teachable.com", "https://www.thinkific.com", "https://www.creativelive.com",
            "https://www.masterclass.com", "https://www.coursera.org", "https://www.udemy.com",
            "https://www.skillshare.com", "https://www.udacity.com", "https://www.edx.org",
            "https://www.khanacademy.org", "https://www.coursera.org", "https://www.udemy.com",
            "https://www.khanacademy.org", "https://www.edx.org", "https://www.udacity.com"
        ]
        
        # 验证网站
        valid_platforms = []
        for url in education_platforms:
            domain = urlparse(url).netloc
            if self.generator.validate_domain(domain) and url not in valid_platforms:
                valid_platforms.append(url)
        
        # 使用Tranco列表扩展
        tranco_url = "https://tranco-list.eu/download/latest/list"
        try:
            response = requests.get(tranco_url, timeout=30)
            response.raise_for_status()
            
            lines = response.text.strip().split('\n')
            for line in lines[:count * 5]:
                if ',' in line:
                    rank, domain = line.split(',', 1)
                    domain = domain.strip()
                    if self.generator.validate_domain(domain):
                        url = f"https://{domain}"
                        if url not in valid_platforms and url not in education_platforms:
                            # 简单检查网站标题是否包含教育相关内容
                            try:
                                response = requests.get(url, timeout=5)
                                if response.status_code < 400:
                                    title = re.search(r'<title>(.*?)</title>', response.text, re.IGNORECASE)
                                    if title:
                                        title_lower = title.group(1).lower()
                                        if any(keyword in title_lower for keyword in ["education", "learn", "course", "university"]):
                                            valid_platforms.append(url)
                                            if len(valid_platforms) >= count:
                                                break
                            except:
                                pass
        except Exception as e:
            logger.error(f"扩展教育平台列表失败：{e}")
        
        logger.info(f"成功获取 {len(valid_platforms)} 个教育门户和平台网站")
        return valid_platforms[:count]
    
    def get_education_sites_from_seed(self, seed_sites, count=200):
        """从种子网站获取相关教育网站"""
        logger.info(f"开始从种子网站获取 {count} 个相关教育网站...")
        related_sites = []
        
        # 这里使用简单的示例实现
        # 在实际应用中，可以通过反向链接、共同IP等方式获取相关网站
        for seed in seed_sites:
            related = self.generator.get_related_domains(urlparse(seed).netloc, 10)
            for site in related:
                if site not in related_sites:
                    related_sites.append(site)
                    if len(related_sites) >= count:
                        return related_sites
        
        logger.info(f"成功获取 {len(related_sites)} 个相关教育网站")
        return related_sites
    
    def get_large_education_list(self):
        """获取大型教育网站列表"""
        logger.info(f"开始从内置列表获取教育网站...")
        
        # 大型教育网站列表（包含大学、学院、教育平台等）
        large_education_list = [
            "https://www.coursera.org",
            "https://www.udemy.com",
            "https://www.khanacademy.org",
            "https://www.edx.org",
            "https://www.udacity.com",
            "https://www.open.edu",
            "https://www.saylor.org",
            "https://www.mooc.org",
            "https://www.futurelearn.com",
            "https://www.alison.com",
            "https://www.pluralsight.com",
            "https://www.codecademy.com",
            "https://www.datacamp.com",
            "https://www.lynda.com",
            "https://www.teachable.com",
            "https://www.thinkific.com",
            "https://www.creativelive.com",
            "https://www.masterclass.com",
            "https://www.skillshare.com",
            "https://www.coursera.org",
            "https://www.udemy.com",
            "https://www.khanacademy.org",
            "https://www.edx.org",
            "https://www.udacity.com",
            "https://www.harvard.edu",
            "https://www.stanford.edu",
            "https://www.mit.edu",
            "https://www.princeton.edu",
            "https://www.yale.edu",
            "https://www.columbia.edu",
            "https://www.chicago.edu",
            "https://www.caltech.edu",
            "https://www.ox.ac.uk",
            "https://www.cam.ac.uk",
            "https://www.imperial.ac.uk",
            "https://www.lse.ac.uk",
            "https://www.ucl.ac.uk",
            "https://www.manchester.ac.uk",
            "https://www.birmingham.ac.uk",
            "https://www.leeds.ac.uk",
            "https://www.sheffield.ac.uk",
            "https://www.nottingham.ac.uk",
            "https://www.liv.ac.uk",
            "https://www.newcastle.ac.uk",
            "https://www.ed.ac.uk",
            "https://www.gla.ac.uk",
            "https://www.strath.ac.uk",
            "https://www.dur.ac.uk",
            "https://www.st-andrews.ac.uk",
            "https://www.bristol.ac.uk",
            "https://www.bath.ac.uk",
            "https://www.exeter.ac.uk",
            "https://www.essex.ac.uk",
            "https://www.reading.ac.uk",
            "https://www.southampton.ac.uk",
            "https://www.sussex.ac.uk",
            "https://www.warwick.ac.uk",
            "https://www.uea.ac.uk",
            "https://www.lancaster.ac.uk",
            "https://www.aston.ac.uk",
            "https://www.bangor.ac.uk",
            "https://www.bournemouth.ac.uk",
            "https://www.brad.ac.uk",
            "https://www.brookes.ac.uk",
            "https://www.canterbury.ac.uk",
            "https://www.coventry.ac.uk",
            "https://www.derby.ac.uk",
            "https://www.dundee.ac.uk",
            "https://www.edgehill.ac.uk",
            "https://www.falmouth.ac.uk",
            "https://www.gold.ac.uk",
            "https://www.gre.ac.uk",
            "https://www.hud.ac.uk",
            "https://www.hull.ac.uk",
            "https://www.kent.ac.uk",
            "https://www.keele.ac.uk",
            "https://www.lincoln.ac.uk",
            "https://www.londonmet.ac.uk",
            "https://www.loughborough.ac.uk",
            "https://www.mmu.ac.uk",
            "https://www.napier.ac.uk",
            "https://www.northumbria.ac.uk",
            "https://www.ntu.ac.uk",
            "https://www.open.ac.uk",
            "https://www.plymouth.ac.uk",
            "https://www.port.ac.uk",
            "https://www.roehampton.ac.uk",
            "https://www.salford.ac.uk",
            "https://www.solent.ac.uk",
            "https://www.stir.ac.uk",
            "https://www.surrey.ac.uk",
            "https://www.swansea.ac.uk",
            "https://www.tees.ac.uk",
            "https://www.westminster.ac.uk",
            "https://www.worc.ac.uk",
            "https://www.york.ac.uk",
            "https://www.abdn.ac.uk",
            "https://www.aber.ac.uk",
            "https://www.anglia.ac.uk",
            "https://www.aru.ac.uk",
            "https://www.beds.ac.uk",
            "https://www.bcu.ac.uk",
            "https://www.uel.ac.uk",
            "https://www.mdx.ac.uk",
            "https://www.ncl.ac.uk",
            "https://www.qub.ac.uk",
            "https://www.royalholloway.ac.uk",
            "https://www.rhul.ac.uk",
            "https://www.soas.ac.uk",
            "https://www.soton.ac.uk",
            "https://www.staffs.ac.uk",
            "https://www.uwtsd.ac.uk",
            "https://www.wlv.ac.uk",
            "https://www.manchester.ac.uk",
            "https://www.birmingham.ac.uk",
            "https://www.leeds.ac.uk",
            "https://www.sheffield.ac.uk",
            "https://www.nottingham.ac.uk",
            "https://www.liv.ac.uk",
            "https://www.newcastle.ac.uk",
            "https://www.ed.ac.uk",
            "https://www.gla.ac.uk",
            "https://www.strath.ac.uk",
            "https://www.dur.ac.uk",
            "https://www.st-andrews.ac.uk",
            "https://www.bristol.ac.uk",
            "https://www.bath.ac.uk",
            "https://www.exeter.ac.uk",
            "https://www.essex.ac.uk",
            "https://www.reading.ac.uk",
            "https://www.southampton.ac.uk",
            "https://www.sussex.ac.uk",
            "https://www.warwick.ac.uk",
            "https://www.uea.ac.uk",
            "https://www.lancaster.ac.uk",
            "https://www.aston.ac.uk",
            "https://www.bangor.ac.uk",
            "https://www.bournemouth.ac.uk",
            "https://www.brad.ac.uk",
            "https://www.brookes.ac.uk",
            "https://www.canterbury.ac.uk",
            "https://www.coventry.ac.uk",
            "https://www.derby.ac.uk",
            "https://www.dundee.ac.uk",
            "https://www.edgehill.ac.uk",
            "https://www.falmouth.ac.uk",
            "https://www.gold.ac.uk",
            "https://www.gre.ac.uk",
            "https://www.hud.ac.uk",
            "https://www.hull.ac.uk",
            "https://www.kent.ac.uk",
            "https://www.keele.ac.uk",
            "https://www.lincoln.ac.uk",
            "https://www.londonmet.ac.uk",
            "https://www.loughborough.ac.uk",
            "https://www.mmu.ac.uk",
            "https://www.napier.ac.uk",
            "https://www.northumbria.ac.uk",
            "https://www.ntu.ac.uk",
            "https://www.open.ac.uk",
            "https://www.plymouth.ac.uk",
            "https://www.port.ac.uk",
            "https://www.roehampton.ac.uk",
            "https://www.salford.ac.uk",
            "https://www.solent.ac.uk",
            "https://www.stir.ac.uk",
            "https://www.surrey.ac.uk",
            "https://www.swansea.ac.uk",
            "https://www.tees.ac.uk",
            "https://www.westminster.ac.uk",
            "https://www.worc.ac.uk",
            "https://www.york.ac.uk",
            "https://www.abdn.ac.uk",
            "https://www.aber.ac.uk",
            "https://www.anglia.ac.uk",
            "https://www.aru.ac.uk",
            "https://www.beds.ac.uk",
            "https://www.bcu.ac.uk",
            "https://www.uel.ac.uk",
            "https://www.mdx.ac.uk",
            "https://www.ncl.ac.uk",
            "https://www.qub.ac.uk",
            "https://www.royalholloway.ac.uk",
            "https://www.rhul.ac.uk",
            "https://www.soas.ac.uk",
            "https://www.soton.ac.uk",
            "https://www.staffs.ac.uk",
            "https://www.uwtsd.ac.uk",
            "https://www.wlv.ac.uk",
            "https://www.manchester.ac.uk",
            "https://www.birmingham.ac.uk",
            "https://www.leeds.ac.uk",
            "https://www.sheffield.ac.uk",
            "https://www.nottingham.ac.uk",
            "https://www.liv.ac.uk",
            "https://www.newcastle.ac.uk",
            "https://www.ed.ac.uk",
            "https://www.gla.ac.uk",
            "https://www.strath.ac.uk",
            "https://www.dur.ac.uk",
            "https://www.st-andrews.ac.uk",
            "https://www.bristol.ac.uk",
            "https://www.bath.ac.uk",
            "https://www.exeter.ac.uk",
            "https://www.essex.ac.uk",
            "https://www.reading.ac.uk",
            "https://www.southampton.ac.uk",
            "https://www.sussex.ac.uk",
            "https://www.warwick.ac.uk",
            "https://www.uea.ac.uk",
            "https://www.lancaster.ac.uk",
            "https://www.aston.ac.uk",
            "https://www.bangor.ac.uk",
            "https://www.bournemouth.ac.uk",
            "https://www.brad.ac.uk",
            "https://www.brookes.ac.uk",
            "https://www.canterbury.ac.uk",
            "https://www.coventry.ac.uk",
            "https://www.derby.ac.uk",
            "https://www.dundee.ac.uk",
            "https://www.edgehill.ac.uk",
            "https://www.falmouth.ac.uk",
            "https://www.gold.ac.uk",
            "https://www.gre.ac.uk",
            "https://www.hud.ac.uk",
            "https://www.hull.ac.uk",
            "https://www.kent.ac.uk",
            "https://www.keele.ac.uk",
            "https://www.lincoln.ac.uk",
            "https://www.londonmet.ac.uk",
            "https://www.loughborough.ac.uk",
            "https://www.mmu.ac.uk",
            "https://www.napier.ac.uk",
            "https://www.northumbria.ac.uk",
            "https://www.ntu.ac.uk",
            "https://www.open.ac.uk",
            "https://www.plymouth.ac.uk",
            "https://www.port.ac.uk",
            "https://www.roehampton.ac.uk",
            "https://www.salford.ac.uk",
            "https://www.solent.ac.uk",
            "https://www.stir.ac.uk",
            "https://www.surrey.ac.uk",
            "https://www.swansea.ac.uk",
            "https://www.tees.ac.uk",
            "https://www.westminster.ac.uk",
            "https://www.worc.ac.uk",
            "https://www.york.ac.uk",
            "https://www.abdn.ac.uk",
            "https://www.aber.ac.uk",
            "https://www.anglia.ac.uk",
            "https://www.aru.ac.uk",
            "https://www.beds.ac.uk",
            "https://www.bcu.ac.uk",
            "https://www.uel.ac.uk",
            "https://www.mdx.ac.uk",
            "https://www.ncl.ac.uk",
            "https://www.qub.ac.uk",
            "https://www.royalholloway.ac.uk",
            "https://www.rhul.ac.uk",
            "https://www.soas.ac.uk",
            "https://www.soton.ac.uk",
            "https://www.staffs.ac.uk",
            "https://www.uwtsd.ac.uk",
            "https://www.wlv.ac.uk",
            "https://www.manchester.ac.uk",
            "https://www.birmingham.ac.uk",
            "https://www.leeds.ac.uk",
            "https://www.sheffield.ac.uk",
            "https://www.nottingham.ac.uk",
            "https://www.liv.ac.uk",
            "https://www.newcastle.ac.uk",
            "https://www.ed.ac.uk",
            "https://www.gla.ac.uk",
            "https://www.strath.ac.uk",
            "https://www.dur.ac.uk",
            "https://www.st-andrews.ac.uk",
            "https://www.bristol.ac.uk",
            "https://www.bath.ac.uk",
            "https://www.exeter.ac.uk",
            "https://www.essex.ac.uk",
            "https://www.reading.ac.uk",
            "https://www.southampton.ac.uk",
            "https://www.sussex.ac.uk",
            "https://www.warwick.ac.uk",
            "https://www.uea.ac.uk",
            "https://www.lancaster.ac.uk",
            "https://www.aston.ac.uk",
            "https://www.bangor.ac.uk",
            "https://www.bournemouth.ac.uk",
            "https://www.brad.ac.uk",
            "https://www.brookes.ac.uk",
            "https://www.canterbury.ac.uk",
            "https://www.coventry.ac.uk",
            "https://www.derby.ac.uk",
            "https://www.dundee.ac.uk",
            "https://www.edgehill.ac.uk",
            "https://www.falmouth.ac.uk",
            "https://www.gold.ac.uk",
            "https://www.gre.ac.uk",
            "https://www.hud.ac.uk",
            "https://www.hull.ac.uk",
            "https://www.kent.ac.uk",
            "https://www.keele.ac.uk",
            "https://www.lincoln.ac.uk",
            "https://www.londonmet.ac.uk",
            "https://www.loughborough.ac.uk",
            "https://www.mmu.ac.uk",
            "https://www.napier.ac.uk",
            "https://www.northumbria.ac.uk",
            "https://www.ntu.ac.uk",
            "https://www.open.ac.uk",
            "https://www.plymouth.ac.uk",
            "https://www.port.ac.uk",
            "https://www.roehampton.ac.uk",
            "https://www.salford.ac.uk",
            "https://www.solent.ac.uk",
            "https://www.stir.ac.uk",
            "https://www.surrey.ac.uk",
            "https://www.swansea.ac.uk",
            "https://www.tees.ac.uk",
            "https://www.westminster.ac.uk",
            "https://www.worc.ac.uk",
            "https://www.york.ac.uk",
            "https://www.abdn.ac.uk",
            "https://www.aber.ac.uk",
            "https://www.anglia.ac.uk",
            "https://www.aru.ac.uk",
            "https://www.beds.ac.uk",
            "https://www.bcu.ac.uk",
            "https://www.uel.ac.uk",
            "https://www.mdx.ac.uk",
            "https://www.ncl.ac.uk",
            "https://www.qub.ac.uk",
            "https://www.royalholloway.ac.uk",
            "https://www.rhul.ac.uk",
            "https://www.soas.ac.uk",
            "https://www.soton.ac.uk",
            "https://www.staffs.ac.uk",
            "https://www.uwtsd.ac.uk",
            "https://www.wlv.ac.uk",
            "https://www.manchester.ac.uk",
            "https://www.birmingham.ac.uk",
            "https://www.leeds.ac.uk",
            "https://www.sheffield.ac.uk",
            "https://www.nottingham.ac.uk",
            "https://www.liv.ac.uk",
            "https://www.newcastle.ac.uk",
            "https://www.ed.ac.uk",
            "https://www.gla.ac.uk",
            "https://www.strath.ac.uk",
            "https://www.dur.ac.uk",
            "https://www.st-andrews.ac.uk",
            "https://www.bristol.ac.uk",
            "https://www.bath.ac.uk",
            "https://www.exeter.ac.uk",
            "https://www.essex.ac.uk",
            "https://www.reading.ac.uk",
            "https://www.southampton.ac.uk",
            "https://www.sussex.ac.uk",
            "https://www.warwick.ac.uk",
            "https://www.uea.ac.uk",
            "https://www.lancaster.ac.uk",
            "https://www.aston.ac.uk",
            "https://www.bangor.ac.uk",
            "https://www.bournemouth.ac.uk",
            "https://www.brad.ac.uk",
            "https://www.brookes.ac.uk",
            "https://www.canterbury.ac.uk",
            "https://www.coventry.ac.uk",
            "https://www.derby.ac.uk",
            "https://www.dundee.ac.uk",
            "https://www.edgehill.ac.uk",
            "https://www.falmouth.ac.uk",
            "https://www.gold.ac.uk",
            "https://www.gre.ac.uk",
            "https://www.hud.ac.uk",
            "https://www.hull.ac.uk",
            "https://www.kent.ac.uk",
            "https://www.keele.ac.uk",
            "https://www.lincoln.ac.uk",
            "https://www.londonmet.ac.uk",
            "https://www.loughborough.ac.uk",
            "https://www.mmu.ac.uk",
            "https://www.napier.ac.uk",
            "https://www.northumbria.ac.uk",
            "https://www.ntu.ac.uk",
            "https://www.open.ac.uk",
            "https://www.plymouth.ac.uk",
            "https://www.port.ac.uk",
            "https://www.roehampton.ac.uk",
            "https://www.salford.ac.uk",
            "https://www.solent.ac.uk",
            "https://www.stir.ac.uk",
            "https://www.surrey.ac.uk",
            "https://www.swansea.ac.uk",
            "https://www.tees.ac.uk",
            "https://www.westminster.ac.uk",
            "https://www.worc.ac.uk",
            "https://www.york.ac.uk",
            "https://www.abdn.ac.uk",
            "https://www.aber.ac.uk",
            "https://www.anglia.ac.uk",
            "https://www.aru.ac.uk",
            "https://www.beds.ac.uk",
            "https://www.bcu.ac.uk",
            "https://www.uel.ac.uk",
            "https://www.mdx.ac.uk",
            "https://www.ncl.ac.uk",
            "https://www.qub.ac.uk",
            "https://www.royalholloway.ac.uk",
            "https://www.rhul.ac.uk",
            "https://www.soas.ac.uk",
            "https://www.soton.ac.uk",
            "https://www.staffs.ac.uk",
            "https://www.uwtsd.ac.uk",
            "https://www.wlv.ac.uk",
            "https://www.manchester.ac.uk",
            "https://www.birmingham.ac.uk",
            "https://www.leeds.ac.uk",
            "https://www.sheffield.ac.uk",
            "https://www.nottingham.ac.uk",
            "https://www.liv.ac.uk",
            "https://www.newcastle.ac.uk",
            "https://www.ed.ac.uk",
            "https://www.gla.ac.uk",
            "https://www.strath.ac.uk",
            "https://www.dur.ac.uk",
            "https://www.st-andrews.ac.uk",
            "https://www.bristol.ac.uk",
            "https://www.bath.ac.uk",
            "https://www.exeter.ac.uk",
            "https://www.essex.ac.uk",
            "https://www.reading.ac.uk",
            "https://www.southampton.ac.uk",
            "https://www.sussex.ac.uk",
            "https://www.warwick.ac.uk",
            "https://www.uea.ac.uk",
            "https://www.lancaster.ac.uk",
            "https://www.aston.ac.uk",
            "https://www.bangor.ac.uk",
            "https://www.bournemouth.ac.uk",
            "https://www.brad.ac.uk",
            "https://www.brookes.ac.uk",
            "https://www.canterbury.ac.uk",
            "https://www.coventry.ac.uk",
            "https://www.derby.ac.uk",
            "https://www.dundee.ac.uk",
            "https://www.edgehill.ac.uk",
            "https://www.falmouth.ac.uk",
            "https://www.gold.ac.uk",
            "https://www.gre.ac.uk",
            "https://www.hud.ac.uk",
            "https://www.hull.ac.uk",
            "https://www.kent.ac.uk",
            "https://www.keele.ac.uk",
            "https://www.lincoln.ac.uk",
            "https://www.londonmet.ac.uk",
            "https://www.loughborough.ac.uk",
            "https://www.mmu.ac.uk",
            "https://www.napier.ac.uk",
            "https://www.northumbria.ac.uk",
            "https://www.ntu.ac.uk",
            "https://www.open.ac.uk",
            "https://www.plymouth.ac.uk",
            "https://www.port.ac.uk",
            "https://www.roehampton.ac.uk",
            "https://www.salford.ac.uk",
            "https://www.solent.ac.uk",
            "https://www.stir.ac.uk",
            "https://www.surrey.ac.uk",
            "https://www.swansea.ac.uk",
            "https://www.tees.ac.uk",
            "https://www.westminster.ac.uk",
            "https://www.worc.ac.uk",
            "https://www.york.ac.uk",
            "https://www.abdn.ac.uk",
            "https://www.aber.ac.uk",
            "https://www.anglia.ac.uk",
            "https://www.aru.ac.uk",
            "https://www.beds.ac.uk",
            "https://www.bcu.ac.uk",
            "https://www.uel.ac.uk",
            "https://www.mdx.ac.uk",
            "https://www.ncl.ac.uk",
            "https://www.qub.ac.uk",
            "https://www.royalholloway.ac.uk",
            "https://www.rhul.ac.uk",
            "https://www.soas.ac.uk",
            "https://www.soton.ac.uk",
            "https://www.staffs.ac.uk",
            "https://www.uwtsd.ac.uk",
            "https://www.wlv.ac.uk",
            "https://www.tsinghua.edu.cn",
            "https://www.pku.edu.cn",
            "https://www.zju.edu.cn",
            "https://www.fudan.edu.cn",
            "https://www.nju.edu.cn",
            "https://www.whu.edu.cn",
            "https://www.sjtu.edu.cn",
            "https://www.ecnu.edu.cn",
            "https://www.hust.edu.cn",
            "https://www.csu.edu.cn",
            "https://www.xjtu.edu.cn",
            "https://www.ntu.edu.sg",
            "https://www.nus.edu.sg",
            "https://www.ubc.ca",
            "https://www.toronto.edu",
            "https://www.mcgill.ca",
            "https://www.utoronto.ca",
            "https://www.mcmaster.ca",
            "https://www.ualberta.ca",
            "https://www.sfu.ca",
            "https://www.mun.ca",
            "https://www.uwaterloo.ca",
            "https://www.queensu.ca",
            "https://www.westernu.ca",
            "https://www.yorku.ca",
            "https://www.concordia.ca",
            "https://www.uottawa.ca",
            "https://www.calgary.ca",
            "https://www.saskatoon.ca",
            "https://www.regina.ca",
            "https://www.uq.edu.au",
            "https://www.unimelb.edu.au",
            "https://www.usyd.edu.au",
            "https://www.mq.edu.au",
            "https://www.unsw.edu.au",
            "https://www.adelaide.edu.au",
            "https://www.monash.edu",
            "https://www.uwa.edu.au",
            "https://www.curtin.edu.au",
            "https://www.anu.edu.au",
            "https://www.griffith.edu.au",
            "https://www.jcu.edu.au",
            "https://www.deakin.edu.au",
            "https://www.utas.edu.au",
            "https://www.uni-sydney.edu.au",
            "https://www.uni-melbourne.edu.au",
            "https://www.uni-brisbane.edu.au",
            "https://www.uni-perth.edu.au",
            "https://www.uni-adelaide.edu.au",
            "https://www.uni-canberra.edu.au",
            "https://www.uni-hobart.edu.au",
            "https://www.uni-darwin.edu.au",
            "https://www.uni-newcastle.edu.au",
            "https://www.uni-wollongong.edu.au",
            "https://www.uni-newcastle.edu.au",
            "https://www.uni-wollongong.edu.au",
            "https://www.uni-newcastle.edu.au",
            "https://www.uni-wollongong.edu.au",
            "https://www.uni-newcastle.edu.au",
            "https://www.uni-wollongong.edu.au",
            "https://www.uni-newcastle.edu.au",
            "https://www.uni-wollongong.edu.au",
            "https://www.uni-newcastle.edu.au",
            "https://www.uni-wollongong.edu.au"
        ]
        
        # 验证并去重
        valid_sites = []
        seen = set()
        
        for url in large_education_list:
            if url not in seen:
                seen.add(url)
                domain = urlparse(url).netloc
                if self.generator.validate_domain(domain):
                    valid_sites.append(url)
                    if len(valid_sites) >= 1000:
                        break
        
        logger.info(f"从内置列表成功获取 {len(valid_sites)} 个有效教育网站")
        return valid_sites
    
    def collect(self, total=1000):
        """收集指定数量的教育网站"""
        logger.info(f"开始收集 {total} 个教育网站...")
        
        # 1. 从大型内置列表获取（主要来源）
        large_list = self.get_large_education_list()
        self.add_education_domains(large_list)
        
        # 2. 从现有的行业网站列表获取
        existing = self.generator.get_industry_websites('education', 200)
        self.add_education_domains(existing)
        
        # 3. 获取教育门户和平台
        portals = self.get_education_portals(300)
        self.add_education_domains(portals)
        
        # 4. 从种子网站获取相关网站
        if self.education_domains:
            related = self.get_education_sites_from_seed(self.education_domains[:20], 200)
            self.add_education_domains(related)
        
        # 去重并限制数量
        result = list(set(self.education_domains))[:total]
        
        logger.info(f"最终收集到 {len(result)} 个教育网站")
        
        return result
    
    def save_to_file(self, sites, filename="education_sites.json"):
        """保存教育网站到文件"""
        logger.info(f"保存 {len(sites)} 个教育网站到 {filename}...")
        data = {
            "total": len(sites),
            "timestamp": datetime.now().isoformat(),
            "sites": sites
        }
        
        with open(filename, 'w', encoding='utf-8') as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
        
        logger.info(f"成功保存到 {filename}")

def main():
    collector = EducationSiteCollector()
    education_sites = collector.collect(1000)
    collector.save_to_file(education_sites)
    
    print(f"\n=== 教育网站收集完成 ===")
    print(f"总共收集到 {len(education_sites)} 个教育网站")
    print(f"已保存到 education_sites.json 文件")
    print(f"前10个网站：")
    for site in education_sites[:10]:
        print(f"- {site}")

if __name__ == "__main__":
    main()
