#!/usr/bin/env python3
import argparse
from scanner.scanner import LanScanner, TrafficMonitor

def main():
    parser = argparse.ArgumentParser(description="局域网扫描和流量监听工具")
    
    # 创建子命令解析器
    subparsers = parser.add_subparsers(dest='command', help='可用命令')
    
    # 扫描命令
    scan_parser = subparsers.add_parser('scan', help='扫描局域网')
    scan_parser.add_argument('-n', '--network', help='指定网络地址，例如 192.168.1.0/24')
    scan_parser.add_argument('-s', '--speed', default='T2', choices=['T1', 'T2', 'T3', 'T4', 'T5'], help='扫描速度，T1最慢最准确，T5最快')
    
    # 流量监听命令
    traffic_parser = subparsers.add_parser('traffic', help='监听局域网流量')
    traffic_parser.add_argument('-i', '--interface', help='指定网络接口')
    
    args = parser.parse_args()
    
    if args.command == 'scan':
        scanner = LanScanner()
        scanner.scan_network(network=args.network, scan_speed=args.speed)
    elif args.command == 'traffic':
        monitor = TrafficMonitor()
        monitor.start_monitoring(interface=args.interface)
    else:
        parser.print_help()

if __name__ == "__main__":
    main()
