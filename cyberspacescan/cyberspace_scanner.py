#!/usr/bin/env python3
"""
网络空间扫描工具
功能：扫描目标的网络空间信息，包括端口、服务、Banner等
"""

import socket
import argparse
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed
from typing import List, Dict, Tuple
import time
import logging

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class CyberSpaceScanner:
    """网络空间扫描器"""
    
    # 常见端口及其服务
    COMMON_PORTS = {
        21: 'FTP',
        22: 'SSH',
        23: 'Telnet',
        25: 'SMTP',
        53: 'DNS',
        80: 'HTTP',
        110: 'POP3',
        143: 'IMAP',
        443: 'HTTPS',
        445: 'SMB',
        3306: 'MySQL',
        3389: 'RDP',
        5432: 'PostgreSQL',
        6379: 'Redis',
        8080: 'HTTP-Proxy',
        8443: 'HTTPS-Alt',
        27017: 'MongoDB',
        9200: 'Elasticsearch',
    }
    
    def __init__(self, timeout: float = 1.0, threads: int = 50):
        self.timeout = timeout
        self.threads = threads
        
    def scan_port(self, host: str, port: int) -> Tuple[int, bool, str, str]:
        """
        扫描单个端口
        
        返回: (端口号, 是否开放, 服务名称, Banner信息)
        """
        try:
            # 创建socket连接
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.settimeout(self.timeout)
            
            result = sock.connect_ex((host, port))
            
            if result == 0:
                # 端口开放，尝试获取Banner
                service = self.COMMON_PORTS.get(port, 'Unknown')
                banner = self.grab_banner(sock, port)
                sock.close()
                return (port, True, service, banner)
            else:
                sock.close()
                return (port, False, '', '')
                
        except socket.timeout:
            return (port, False, '', '')
        except socket.error as e:
            return (port, False, '', '')
        except Exception as e:
            logger.debug(f"扫描端口 {port} 时出错: {e}")
            return (port, False, '', '')
    
    def grab_banner(self, sock: socket.socket, port: int) -> str:
        """获取服务Banner信息"""
        try:
            # 对于某些服务，需要先发送数据
            if port in [80, 8080, 8443]:
                sock.send(b'GET / HTTP/1.1\r\nHost: localhost\r\n\r\n')
            elif port == 21:
                pass  # FTP会自动发送欢迎信息
            elif port == 22:
                pass  # SSH会自动发送版本信息
            elif port == 25:
                sock.send(b'EHLO test\r\n')
            
            # 接收响应
            sock.settimeout(0.5)
            banner = sock.recv(1024).decode('utf-8', errors='ignore').strip()
            return banner[:200]  # 限制长度
            
        except:
            return ''
    
    def scan_host(self, host: str, ports: List[int]) -> List[Dict]:
        """
        扫描主机的多个端口
        
        返回: 开放端口的详细信息列表
        """
        logger.info(f"开始扫描主机: {host}")
        logger.info(f"扫描端口数量: {len(ports)}")
        logger.info(f"使用线程数: {self.threads}")
        
        open_ports = []
        
        with ThreadPoolExecutor(max_workers=self.threads) as executor:
            # 提交所有扫描任务
            futures = {
                executor.submit(self.scan_port, host, port): port 
                for port in ports
            }
            
            # 收集结果
            completed = 0
            for future in as_completed(futures):
                completed += 1
                if completed % 100 == 0:
                    logger.info(f"扫描进度: {completed}/{len(ports)}")
                
                try:
                    port, is_open, service, banner = future.result()
                    
                    if is_open:
                        port_info = {
                            'port': port,
                            'service': service,
                            'banner': banner,
                            'state': 'open'
                        }
                        open_ports.append(port_info)
                        logger.info(f"✓ 发现开放端口: {port} ({service})")
                        
                except Exception as e:
                    logger.error(f"获取扫描结果时出错: {e}")
        
        return sorted(open_ports, key=lambda x: x['port'])
    
    def scan_range(self, host: str, start_port: int, end_port: int) -> List[Dict]:
        """扫描端口范围"""
        ports = list(range(start_port, end_port + 1))
        return self.scan_host(host, ports)
    
    def scan_common_ports(self, host: str) -> List[Dict]:
        """扫描常见端口"""
        ports = list(self.COMMON_PORTS.keys())
        return self.scan_host(host, ports)
    
    def resolve_host(self, host: str) -> str:
        """解析主机名到IP地址"""
        try:
            ip = socket.gethostbyname(host)
            logger.info(f"主机解析: {host} -> {ip}")
            return ip
        except socket.gaierror:
            logger.error(f"无法解析主机名: {host}")
            return None
    
    def get_host_info(self, host: str) -> Dict:
        """获取主机基本信息"""
        info = {
            'hostname': host,
            'ip': None,
            'ptr': None
        }
        
        try:
            # 获取IP地址
            info['ip'] = socket.gethostbyname(host)
            
            # 尝试反向DNS查询
            try:
                info['ptr'] = socket.gethostbyaddr(info['ip'])[0]
            except:
                pass
                
        except Exception as e:
            logger.error(f"获取主机信息失败: {e}")
        
        return info


def print_results(host_info: Dict, open_ports: List[Dict]):
    """打印扫描结果"""
    print("\n" + "="*70)
    print("网络空间扫描结果")
    print("="*70)
    
    print(f"\n目标信息:")
    print(f"  主机名: {host_info['hostname']}")
    print(f"  IP地址: {host_info['ip']}")
    if host_info['ptr']:
        print(f"  PTR记录: {host_info['ptr']}")
    
    print(f"\n开放端口: {len(open_ports)}")
    print("-"*70)
    
    if open_ports:
        print(f"{'端口':<8} {'服务':<15} {'状态':<8} Banner")
        print("-"*70)
        
        for port_info in open_ports:
            banner = port_info['banner'][:50] + '...' if len(port_info['banner']) > 50 else port_info['banner']
            print(f"{port_info['port']:<8} {port_info['service']:<15} {port_info['state']:<8} {banner}")
    else:
        print("未发现开放端口")
    
    print("\n" + "="*70)


def save_results(filename: str, host_info: Dict, open_ports: List[Dict]):
    """保存结果到文件"""
    try:
        import json
        
        results = {
            'host_info': host_info,
            'open_ports': open_ports,
            'scan_time': time.strftime('%Y-%m-%d %H:%M:%S')
        }
        
        with open(filename, 'w', encoding='utf-8') as f:
            json.dump(results, f, indent=2, ensure_ascii=False)
        
        logger.info(f"结果已保存到: {filename}")
        
    except Exception as e:
        logger.error(f"保存结果失败: {e}")


def main():
    parser = argparse.ArgumentParser(
        description='网络空间扫描工具',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
使用示例:
  # 扫描常见端口
  python3 cyberspace_scanner.py -t example.com
  
  # 扫描指定端口范围
  python3 cyberspace_scanner.py -t 192.168.1.1 -p 1-1000
  
  # 扫描指定端口列表
  python3 cyberspace_scanner.py -t example.com -p 80,443,8080,3306
  
  # 快速扫描（减少超时时间）
  python3 cyberspace_scanner.py -t example.com --timeout 0.5 --threads 100
  
  # 保存结果到文件
  python3 cyberspace_scanner.py -t example.com -o scan_result.json

注意:
  - 请仅在授权范围内使用此工具
  - 未经授权的端口扫描可能违反法律法规
        """
    )
    
    parser.add_argument('-t', '--target', required=True, help='目标主机（IP或域名）')
    parser.add_argument('-p', '--ports', help='端口范围或列表（如: 1-1000 或 80,443,8080）')
    parser.add_argument('--timeout', type=float, default=1.0, help='连接超时时间（秒，默认1.0）')
    parser.add_argument('--threads', type=int, default=50, help='线程数（默认50）')
    parser.add_argument('-o', '--output', help='输出文件（JSON格式）')
    
    args = parser.parse_args()
    
    # 创建扫描器
    scanner = CyberSpaceScanner(timeout=args.timeout, threads=args.threads)
    
    # 获取主机信息
    host_info = scanner.get_host_info(args.target)
    
    if not host_info['ip']:
        logger.error("无法解析目标主机，退出")
        sys.exit(1)
    
    # 确定要扫描的端口
    if args.ports:
        if '-' in args.ports:
            # 端口范围
            start, end = map(int, args.ports.split('-'))
            logger.info(f"扫描端口范围: {start}-{end}")
            open_ports = scanner.scan_range(host_info['ip'], start, end)
        elif ',' in args.ports:
            # 端口列表
            ports = [int(p.strip()) for p in args.ports.split(',')]
            logger.info(f"扫描指定端口: {ports}")
            open_ports = scanner.scan_host(host_info['ip'], ports)
        else:
            # 单个端口
            port = int(args.ports)
            logger.info(f"扫描单个端口: {port}")
            open_ports = scanner.scan_host(host_info['ip'], [port])
    else:
        # 扫描常见端口
        logger.info("扫描常见端口")
        open_ports = scanner.scan_common_ports(host_info['ip'])
    
    # 打印结果
    print_results(host_info, open_ports)
    
    # 保存结果
    if args.output:
        save_results(args.output, host_info, open_ports)
    
    # 统计信息
    logger.info(f"扫描完成！共发现 {len(open_ports)} 个开放端口")


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n\n扫描被用户中断")
        sys.exit(0)
    except Exception as e:
        logger.error(f"程序出错: {e}")
        sys.exit(1)
