#!/usr/bin/env python3
import requests
import re

# 检查百度网站的响应头和基本内容
url = "https://www.baidu.com"
print(f"检查 {url}...")

# 获取响应
response = requests.get(url, timeout=10)
print(f"状态码: {response.status_code}")

# 打印响应头
print("\n响应头:")
for header, value in response.headers.items():
    print(f"{header}: {value}")

# 检查HTML内容中的关键信息
html = response.text
print("\nHTML内容分析:")

# 检查生成器
meta_generator = re.search(r'<meta name=["\']generator["\'] content=["\']([^"\']+)["\']', html, re.IGNORECASE)
if meta_generator:
    print(f"生成器: {meta_generator.group(1)}")

# 检查主要脚本
scripts = re.findall(r'<script[^>]+src=["\']([^"\']+)["\']', html)
print(f"找到 {len(scripts)} 个脚本标签")
if scripts:
    print("主要脚本:")
    for script in scripts[:5]:  # 只显示前5个
        print(f"  - {script}")

# 检查CSS文件
styles = re.findall(r'<link[^>]+href=["\']([^"\']+\.css)["\']', html)
print(f"找到 {len(styles)} 个CSS文件")
if styles:
    print("主要CSS文件:")
    for style in styles[:3]:  # 只显示前3个
        print(f"  - {style}")

# 检查主要HTML结构
if '<div id="wrapper">' in html:
    print("发现wrapper div结构")
if 'baidu' in html.lower():
    print("页面包含'baidu'关键词")

# 检查Cookie
if 'Set-Cookie' in response.headers:
    cookies = response.headers['Set-Cookie']
    print(f"\nCookie信息: {cookies[:100]}...")  # 只显示前100个字符