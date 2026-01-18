#!/usr/bin/env python3
import argparse
import csv
import time
from scanner import WebsiteScanner
from utils import load_targets

def main():
    parser = argparse.ArgumentParser(description="网站扫描工具")
    parser.add_argument("-t", "--targets", required=True, help="包含网站URL的文件路径")
    parser.add_argument("-o", "--output", default="scan_results.csv", help="输出CSV文件路径")
    args = parser.parse_args()
    
    print("网站扫描工具启动...")
    print(f"加载目标文件: {args.targets}")
    
    targets = load_targets(args.targets)
    if not targets:
        print("没有找到有效的目标网站")
        return
    
    print(f"共加载 {len(targets)} 个目标网站")
    
    scanner = WebsiteScanner()
    results = []
    
    for idx, target in enumerate(targets, 1):
        print(f"\n[{idx}/{len(targets)}] 扫描: {target}")
        try:
            result = scanner.scan(target)
            results.append(result)
            print(f"  ✅ 扫描完成: {result['title']}")
        except Exception as e:
            print(f"  ❌ 扫描失败: {e}")
            results.append({
                'url': target,
                'title': '',
                'site_name': '',
                'frameworks': [],
                'services': [],
                'applications': [],
                'programming_languages': [],
                'icp': '',
                'has_login_form': False,
                'error': str(e)
            })
        
        # 避免请求过快
        time.sleep(1)
    
    # 保存结果到CSV
    print(f"\n保存结果到: {args.output}")
    save_to_csv(results, args.output)
    print(f"✅ 扫描完成，共扫描 {len(results)} 个网站")

def save_to_csv(results, output_file):
    # 定义CSV字段
    fields = [
        'url', 'title', 'site_name', 'frameworks', 'services', 
        'applications', 'programming_languages', 'icp', 'has_login_form', 'error'
    ]
    
    with open(output_file, mode='w', newline='', encoding='utf-8') as f:
        writer = csv.DictWriter(f, fieldnames=fields)
        writer.writeheader()
        
        for result in results:
            # 处理列表类型为字符串
            row = result.copy()
            for key in ['frameworks', 'services', 'applications', 'programming_languages']:
                if key in row:
                    row[key] = ', '.join(row[key]) if isinstance(row[key], list) else str(row[key])
            
            writer.writerow(row)

if __name__ == "__main__":
    main()
