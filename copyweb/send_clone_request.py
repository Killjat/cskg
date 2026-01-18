#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
向网页克隆工具发送URL克隆请求的脚本

功能特性：
1. 支持单个URL克隆
2. 支持从文件批量克隆URL
3. 支持重试机制
4. 支持保存结果到文件
5. 详细的错误信息和提示
6. 支持不同输出格式

使用方法：
  单个URL：python send_clone_request.py <程序部署URL> <目标克隆URL>
  批量处理：python send_clone_request.py <程序部署URL> --file <URL列表文件>
  保存结果：python send_clone_request.py <程序部署URL> <目标克隆URL> --output <输出文件>

示例：
  python send_clone_request.py http://localhost:5001 https://example.com
  python send_clone_request.py http://localhost:5001 --file urls.txt
  python send_clone_request.py http://localhost:5001 https://example.com --output result.txt
"""

import sys
import requests
import json
import time
import argparse
from typing import List, Dict, Optional


def send_clone_request(
    server_url: str, 
    target_url: str, 
    retries: int = 3, 
    delay: int = 2
) -> Dict[str, any]:
    """
    向克隆服务器发送URL请求并返回克隆结果
    
    Args:
        server_url: 克隆服务器URL
        target_url: 目标克隆URL
        retries: 重试次数
        delay: 重试间隔（秒）
        
    Returns:
        包含克隆结果的字典
    """
    for attempt in range(retries):
        try:
            # 构建API端点
            api_url = f"{server_url}/api/clone"
            
            print(f"🚀 正在向服务器发送请求... (尝试 {attempt + 1}/{retries})")
            print(f"🌐 服务器地址: {server_url}")
            print(f"🎯 目标URL: {target_url}")
            print()
            
            # 发送POST请求
            response = requests.post(
                api_url,
                json={"url": target_url},
                headers={"Content-Type": "application/json"},
                timeout=30
            )
            
            # 检查响应状态
            response.raise_for_status()
            
            # 解析响应
            result = response.json()
            
            if result.get("success"):
                # 服务器已经返回了直接可访问的URL
                return {
                    'success': True,
                    'result': result
                }
            else:
                return {
                    'success': False,
                    'error': result.get('error', '未知错误')
                }
                
        except requests.exceptions.RequestException as e:
            if attempt < retries - 1:
                print(f"⚠️ 请求失败 (尝试 {attempt + 1}/{retries}): {str(e)}")
                print(f"⏱️ 将在 {delay} 秒后重试...")
                time.sleep(delay)
                print()
            else:
                return {
                    'success': False,
                    'error': f"请求失败: {str(e)}"
                }
        except json.JSONDecodeError as e:
            return {
                'success': False,
                'error': f"响应解析失败: {str(e)}"
            }
    
    return {
        'success': False,
        'error': "超过最大重试次数"
    }


def format_result(result: Dict[str, any], verbose: bool = True) -> str:
    """格式化克隆结果"""
    if result['success']:
        info = result['result']
        
        # 直接使用服务器返回的URL
        access_url = info.get('access_url')
        full_url = info.get('full_url')
        info_url = info.get('info_url')
        
        if verbose:
            output = []
            output.append("✅ 克隆成功！")
            output.append(f"📄 标题: {info.get('title')}")
            output.append(f"💾 保存目录: {info.get('save_dir')}")
            output.append("")
            output.append("📋 克隆后的访问地址:")
            output.append("="*60)
            
            output.append(f"🌐 简化版: {access_url}")
            output.append(f"🌐 完整版: {full_url}")
            output.append(f"📊 提取信息: {info_url}")
            output.append("="*60)
            output.append("")
            output.append(f"✅ 操作完成！直接访问地址: {access_url}")
            return "\n".join(output)
        else:
            return access_url
    else:
        return f"❌ 克隆失败: {result['error']}"


def process_single_url(
    server_url: str, 
    target_url: str, 
    output_file: Optional[str] = None,
    retries: int = 3,
    delay: int = 2
) -> bool:
    """处理单个URL克隆"""
    result = send_clone_request(server_url, target_url, retries=retries, delay=delay)
    output = format_result(result)
    
    print(output)
    
    if output_file:
        with open(output_file, 'w', encoding='utf-8') as f:
            f.write(output)
        print(f"\n📄 结果已保存到: {output_file}")
    
    return result['success']


def process_batch_urls(
    server_url: str, 
    file_path: str, 
    output_file: Optional[str] = None,
    retries: int = 3,
    delay: int = 2
) -> bool:
    """处理批量URL克隆"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            urls = [line.strip() for line in f if line.strip() and not line.startswith('#')]
    except Exception as e:
        print(f"❌ 读取URL文件失败: {str(e)}")
        return False
    
    if not urls:
        print("❌ URL列表文件为空")
        return False
    
    print(f"📋 共加载 {len(urls)} 个URL")
    print()
    
    results = []
    success_count = 0
    fail_count = 0
    
    for i, url in enumerate(urls, 1):
        print(f"📌 处理 URL {i}/{len(urls)}: {url}")
        print("-" * 60)
        
        result = send_clone_request(server_url, url, retries=retries, delay=delay)
        results.append(result)
        
        if result['success']:
            success_count += 1
        else:
            fail_count += 1
        
        print(format_result(result))
        print()
    
    # 输出统计信息
    print("=" * 60)
    print("📊 批量处理统计:")
    print(f"✅ 成功: {success_count} 个")
    print(f"❌ 失败: {fail_count} 个")
    print(f"📈 成功率: {(success_count / len(urls) * 100):.1f}%")
    print("=" * 60)
    
    # 保存结果到文件
    if output_file:
        with open(output_file, 'w', encoding='utf-8') as f:
            for i, (url, result) in enumerate(zip(urls, results), 1):
                f.write(f"# URL {i}: {url}\n")
                if result['success']:
                    # 直接使用服务器返回的URL
                    f.write(f"成功: {result['result']['access_url']}\n")
                else:
                    f.write(f"失败: {result['error']}\n")
                f.write("\n")
            
            # 写入统计信息
            f.write("=" * 60 + "\n")
            f.write("批量处理统计:\n")
            f.write(f"成功: {success_count} 个\n")
            f.write(f"失败: {fail_count} 个\n")
            f.write(f"成功率: {(success_count / len(urls) * 100):.1f}%\n")
        
        print(f"\n� 结果已保存到: {output_file}")
    
    return success_count > 0


def main():
    """主函数"""
    # 解析命令行参数
    parser = argparse.ArgumentParser(
        description='向网页克隆工具发送URL克隆请求的脚本',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""\n示例用法:\n  python send_clone_request.py http://localhost:5001 https://example.com\n  python send_clone_request.py http://localhost:5001 --file urls.txt\n  python send_clone_request.py http://localhost:5001 https://example.com --output result.txt\n"""
    )
    
    parser.add_argument('server_url', help='克隆服务器的URL')
    parser.add_argument('target_url', nargs='?', help='目标克隆URL（与--file二选一）')
    parser.add_argument('--file', '-f', help='包含URL列表的文件路径')
    parser.add_argument('--output', '-o', help='保存结果的文件路径')
    parser.add_argument('--retries', '-r', type=int, default=3, help='请求失败重试次数')
    parser.add_argument('--delay', '-d', type=int, default=2, help='重试间隔（秒）')
    
    args = parser.parse_args()
    
    # 检查参数有效性
    if not (args.target_url or args.file):
        parser.error('必须提供 target_url 或 --file 参数')
    
    if args.target_url and args.file:
        parser.error('target_url 和 --file 参数不能同时使用')
    
    success = False
    
    if args.target_url:
        # 处理单个URL
        success = process_single_url(
            args.server_url, 
            args.target_url, 
            args.output,
            retries=args.retries,
            delay=args.delay
        )
    else:
        # 处理批量URL
        success = process_batch_urls(
            args.server_url, 
            args.file, 
            args.output,
            retries=args.retries,
            delay=args.delay
        )
    
    # 根据结果设置退出码
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
