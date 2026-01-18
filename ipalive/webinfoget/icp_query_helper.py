#!/usr/bin/env python3
"""
ICP备案号查询助手
由于自动查询限制，本工具提供手动查询指导和验证功能
"""

import argparse
import webbrowser
import sys


def print_header():
    print("\n" + "="*70)
    print("ICP备案号查询助手")
    print("="*70 + "\n")


def print_query_guide(icp_number: str):
    """打印查询指南"""
    print(f"📋 正在查询ICP备案号: {icp_number}\n")
    
    print("由于自动查询受到限制，请按以下步骤手动查询：\n")
    
    print("步骤1: 访问以下任一网站进行查询")
    print("-" * 70)
    
    query_sites = [
        {
            "name": "工信部备案管理系统（最权威）",
            "url": "https://beian.miit.gov.cn/",
            "note": "官方查询，最准确"
        },
        {
            "name": "天眼查",
            "url": f"https://www.tianyancha.com/search?key={icp_number}",
            "note": "可能需要登录"
        },
        {
            "name": "爱站网",
            "url": f"https://icp.aizhan.com/{icp_number}/",
            "note": "免费查询"
        },
        {
            "name": "站长之家",
            "url": f"https://icp.chinaz.com/{icp_number}",
            "note": "可能需要验证码"
        }
    ]
    
    for i, site in enumerate(query_sites, 1):
        print(f"\n{i}. {site['name']}")
        print(f"   URL: {site['url']}")
        print(f"   说明: {site['note']}")
    
    print("\n" + "="*70)
    print("\n步骤2: 从查询结果中获取域名列表")
    print("-" * 70)
    print("将查询到的域名保存到文本文件中，每行一个域名")
    print("例如保存为: domains.txt\n")
    
    print("示例文件内容:")
    print("-" * 70)
    print("example.com")
    print("test.cn")
    print("demo.com.cn")
    
    print("\n" + "="*70)
    print("\n步骤3: 使用验证工具验证域名")
    print("-" * 70)
    print(f"运行命令:")
    print(f"python3 icp_verify.py -icp={icp_number} -f=domains.txt -o=verified.csv")
    
    print("\n或验证单个域名:")
    print(f"python3 icp_verify.py -icp={icp_number} -d=example.com")
    
    print("\n" + "="*70)


def open_query_sites(icp_number: str):
    """在浏览器中打开查询网站"""
    print("\n是否在浏览器中打开查询网站？(y/n): ", end="")
    choice = input().strip().lower()
    
    if choice == 'y':
        urls = [
            "https://beian.miit.gov.cn/",
            f"https://www.tianyancha.com/search?key={icp_number}",
            f"https://icp.aizhan.com/{icp_number}/",
        ]
        
        print("\n正在打开浏览器...")
        for url in urls[:2]:  # 只打开前2个，避免打开太多
            try:
                webbrowser.open(url)
                print(f"✓ 已打开: {url}")
            except:
                print(f"✗ 无法打开: {url}")
        
        print("\n提示: 如果浏览器没有自动打开，请手动复制上面的URL访问")


def create_template_file():
    """创建域名列表模板文件"""
    template = """# ICP备案号对应的域名列表
# 每行一个域名，不需要 http:// 或 https://
# 以 # 开头的行是注释

# 示例：
# example.com
# test.cn
# demo.com.cn

# 请在下方添加你查询到的域名：

"""
    
    filename = "domains_template.txt"
    try:
        with open(filename, 'w', encoding='utf-8') as f:
            f.write(template)
        print(f"\n✓ 已创建域名列表模板文件: {filename}")
        print(f"  请编辑此文件，添加查询到的域名")
    except Exception as e:
        print(f"\n✗ 创建模板文件失败: {e}")


def main():
    parser = argparse.ArgumentParser(
        description='ICP备案号查询助手 - 提供手动查询指导',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
示例用法:
  python3 icp_query_helper.py -icp=闽ICP备06031865号
  python3 icp_query_helper.py -icp=京ICP证030173号 --open

完整流程:
  1. 使用本工具获取查询指导
  2. 手动访问推荐网站查询
  3. 将查询结果保存到文本文件
  4. 使用 icp_verify.py 验证域名
        """
    )
    
    parser.add_argument('-icp', '--icp_number', required=True, help='ICP备案号')
    parser.add_argument('--open', action='store_true', help='在浏览器中打开查询网站')
    parser.add_argument('--create-template', action='store_true', help='创建域名列表模板文件')
    
    args = parser.parse_args()
    
    print_header()
    print_query_guide(args.icp_number)
    
    if args.create_template:
        create_template_file()
    
    if args.open:
        open_query_sites(args.icp_number)
    else:
        print("\n💡 提示: 使用 --open 参数可以自动在浏览器中打开查询网站")
        print(f"   命令: python3 icp_query_helper.py -icp={args.icp_number} --open")
    
    print("\n" + "="*70)
    print("查询助手使用完毕")
    print("="*70 + "\n")


if __name__ == "__main__":
    main()
