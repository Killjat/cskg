#!/usr/bin/env python3
import re
import requests
from bs4 import BeautifulSoup
from Wappalyzer import Wappalyzer, WebPage
import ssl
import certifi

session = requests.Session()
session.verify = certifi.where()
session.headers.update({
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
})

class WebsiteScanner:
    def __init__(self):
        self.wappalyzer = Wappalyzer.latest()
    
    def scan(self, url):
        """主扫描方法"""
        result = {
            'url': url,
            'title': '',
            'site_name': '',
            'frameworks': [],
            'services': [],
            'applications': [],
            'programming_languages': [],
            'icp': '',
            'has_login_form': False
        }
        
        # 获取网页内容
        response = session.get(url, timeout=10)
        response.raise_for_status()
        
        soup = BeautifulSoup(response.text, 'lxml')
        webpage = WebPage.new_from_response(response)
        
        # 获取标题
        result['title'] = self._get_title(soup)
        
        # 获取网站名称
        result['site_name'] = self._get_site_name(soup, response.url)
        
        # 识别技术栈
        technologies = self._identify_technologies(webpage)
        result.update(technologies)
        
        # 获取ICP备案号
        result['icp'] = self._get_icp(response.text)
        
        # 识别登录框
        result['has_login_form'] = self._has_login_form(soup)
        
        return result
    
    def _get_title(self, soup):
        """获取网站标题"""
        title = soup.title
        return title.text.strip() if title else ''
    
    def _get_site_name(self, soup, url):
        """获取网站名称"""
        # 尝试从meta标签获取
        meta_name = soup.find('meta', attrs={'property': 'og:site_name'})
        if meta_name:
            return meta_name.get('content', '').strip()
        
        meta_name = soup.find('meta', attrs={'name': 'application-name'})
        if meta_name:
            return meta_name.get('content', '').strip()
        
        # 尝试从title中提取
        title = self._get_title(soup)
        if title:
            return title.split(' - ')[0].strip()
        
        # 使用域名作为默认
        return url.split('//')[-1].split('/')[0]
    
    def _identify_technologies(self, webpage):
        """识别网站使用的技术栈"""
        results = {
            'frameworks': [],
            'services': [],
            'applications': [],
            'programming_languages': []
        }
        
        try:
            technologies = self.wappalyzer.analyze_with_versions_and_categories(webpage)
            
            for tech_name, tech_data in technologies.items():
                categories = tech_data.get('categories', [])
                
                if any('framework' in cat.lower() for cat in categories):
                    results['frameworks'].append(tech_name)
                elif any('service' in cat.lower() or 'hosting' in cat.lower() for cat in categories):
                    results['services'].append(tech_name)
                elif any('programming language' in cat.lower() for cat in categories):
                    results['programming_languages'].append(tech_name)
                else:
                    results['applications'].append(tech_name)
            
        except Exception as e:
            print(f"  ⚠️  技术栈识别出错: {e}")
        
        return results
    
    def _get_icp(self, html_content):
        """从网页内容中提取ICP备案号"""
        # 常见的ICP备案号格式正则
        icp_patterns = [
            r'ICP备案号:?\s*(\w+\-\d+)',
            r'ICP备(\d+)号',
            r'ICP备(\d+)\-(\d+)',
            r'京ICP备(\d+)号',
            r'\((?:京|沪|粤|浙|苏|鲁|豫|川|湘|鄂|闽|皖|赣|辽|吉|黑|陕|甘|宁|青|新|云|贵|桂|琼|晋|蒙|津|渝|港|澳|台)ICP备(\d+)号\-?(\d+)?\)',
        ]
        
        for pattern in icp_patterns:
            match = re.search(pattern, html_content, re.IGNORECASE)
            if match:
                # 组合完整的ICP号
                if len(match.groups()) == 2:
                    return f"ICP备{match.group(1)}-{match.group(2)}"
                elif len(match.groups()) == 1:
                    return f"ICP备{match.group(1)}"
                return match.group(1)
        
        return ''
    
    def _has_login_form(self, soup):
        """识别网站是否有登录框"""
        # 查找登录表单
        forms = soup.find_all('form')
        
        for form in forms:
            # 检查表单属性
            form_id = form.get('id', '').lower()
            form_class = form.get('class', [])
            form_action = form.get('action', '').lower()
            
            # 检查表单是否包含登录相关关键词
            if any(keyword in form_id for keyword in ['login', 'signin', 'auth']):
                return True
            
            if any(keyword in str(form_class) for keyword in ['login', 'signin', 'auth']):
                return True
            
            if any(keyword in form_action for keyword in ['login', 'signin', 'auth', 'submit']):
                return True
            
            # 检查表单内的输入字段
            inputs = form.find_all('input')
            has_username = any(input_tag.get('type') in ['text', 'email', 'username'] or 
                              any(keyword in input_tag.get('name', '').lower() for keyword in ['user', 'email', 'login', 'username']) or
                              any(keyword in input_tag.get('id', '').lower() for keyword in ['user', 'email', 'login', 'username']) 
                              for input_tag in inputs)
            
            has_password = any(input_tag.get('type') == 'password' or 
                              any(keyword in input_tag.get('name', '').lower() for keyword in ['pass', 'password']) or
                              any(keyword in input_tag.get('id', '').lower() for keyword in ['pass', 'password']) 
                              for input_tag in inputs)
            
            if has_username and has_password:
                return True
        
        # 检查是否有登录链接或按钮
        login_links = soup.find_all(['a', 'button'], text=re.compile(r'登录|Login|Sign in|Signin', re.IGNORECASE))
        if login_links:
            return True
        
        return False
