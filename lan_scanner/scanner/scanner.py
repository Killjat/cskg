import subprocess
import re
import time
import socket
from scapy.all import ARP, Ether, srp
from database.db import Database
import netifaces
import ipaddress

class LanScanner:
    def __init__(self):
        self.db = Database()
        # 常用端口服务映射表
        self.common_services = {
            21: 'ftp',
            22: 'ssh',
            23: 'telnet',
            25: 'smtp',
            53: 'dns',
            80: 'http',
            110: 'pop3',
            143: 'imap',
            443: 'https',
            3306: 'mysql',
            5432: 'postgresql',
            8080: 'http-proxy',
            8443: 'https-alt'
        }

    def get_local_network(self):
        """获取本地网络信息"""
        try:
            # 获取默认网关接口
            gws = netifaces.gateways()
            default_gateway = gws['default'][netifaces.AF_INET]
            interface = default_gateway[1]
            
            # 获取接口的IP地址和子网掩码
            addresses = netifaces.ifaddresses(interface)
            ipv4_info = addresses[netifaces.AF_INET][0]
            ip = ipv4_info['addr']
            netmask = ipv4_info['netmask']
            
            # 计算网络地址和广播地址
            network = ipaddress.IPv4Network(f'{ip}/{netmask}', strict=False)
            return str(network)
        except Exception as e:
            print(f"获取本地网络信息错误: {e}")
            return "192.168.1.0/24"  # 默认值

    def arp_scan(self, network):
        """使用ARP扫描发现局域网内的活跃设备"""
        print(f"开始ARP扫描: {network}")
        devices = []
        
        # 创建ARP请求数据包
        arp = ARP(pdst=network)
        ether = Ether(dst="ff:ff:ff:ff:ff:ff")
        packet = ether/arp
        
        # 发送数据包并接收响应
        result = srp(packet, timeout=3, verbose=0)[0]
        
        for sent, received in result:
            devices.append({
                'ip': received.psrc,
                'mac': received.hwsrc,
                'hostname': self._get_hostname(received.psrc)
            })
        
        print(f"ARP扫描完成，发现 {len(devices)} 个活跃设备")
        return devices

    def _get_hostname(self, ip):
        """获取IP对应的主机名"""
        try:
            hostname = socket.gethostbyaddr(ip)[0]
            return hostname
        except socket.herror:
            return None

    def _parse_ports(self, ports_str):
        """解析端口范围字符串，返回端口列表"""
        ports = []
        if '-' in ports_str:
            start, end = map(int, ports_str.split('-'))
            ports = range(start, end + 1)
        else:
            ports = [int(ports_str)]
        return ports
    
    def _scan_tcp_port(self, ip, port, timeout=1):
        """扫描单个TCP端口"""
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(timeout)
        try:
            result = sock.connect_ex((ip, port))
            if result == 0:
                # 尝试获取服务信息
                service = self.common_services.get(port, '')
                return {
                    'port': port,
                    'protocol': 'tcp',
                    'status': 'open',
                    'service': service,
                    'application': ''
                }
            else:
                return None
        except Exception as e:
            return None
        finally:
            sock.close()
    
    def _scan_udp_port(self, ip, port, timeout=1):
        """扫描单个UDP端口"""
        sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        sock.settimeout(timeout)
        try:
            # 发送空数据包
            sock.sendto(b'', (ip, port))
            # 尝试接收响应
            sock.recvfrom(1024)
            service = self.common_services.get(port, '')
            return {
                'port': port,
                'protocol': 'udp',
                'status': 'open',
                'service': service,
                'application': ''
            }
        except socket.timeout:
            # UDP端口可能是开放的但没有响应
            service = self.common_services.get(port, '')
            return {
                'port': port,
                'protocol': 'udp',
                'status': 'open|filtered',
                'service': service,
                'application': ''
            }
        except Exception as e:
            return None
        finally:
            sock.close()
    
    def port_scan(self, ip, ports='1-1024', scan_speed='T2'):
        """对单个IP进行端口扫描"""
        print(f"开始端口扫描: {ip}")
        ports_info = []
        
        # 根据扫描速度调整超时时间
        timeout_map = {'T1': 2, 'T2': 1, 'T3': 0.5, 'T4': 0.2, 'T5': 0.1}
        timeout = timeout_map.get(scan_speed, 1)
        
        # 解析端口范围
        port_list = self._parse_ports(ports)
        
        # 只扫描前1000个端口，避免扫描时间过长
        if len(port_list) > 1000:
            port_list = port_list[:1000]
        
        for port in port_list:
            try:
                # 扫描TCP端口
                tcp_result = self._scan_tcp_port(ip, port, timeout)
                if tcp_result:
                    ports_info.append(tcp_result)
                
                # 扫描UDP端口（只扫描常用端口）
                if port in self.common_services:
                    udp_result = self._scan_udp_port(ip, port, timeout)
                    if udp_result:
                        ports_info.append(udp_result)
                
                # 控制扫描速度
                time.sleep(timeout * 0.1)
            except Exception as e:
                continue
        
        print(f"端口扫描完成 {ip}，发现 {len(ports_info)} 个开放端口")
        return ports_info

    def scan_network(self, network=None, scan_speed='T2'):
        """扫描整个网络"""
        if not network:
            network = self.get_local_network()
        
        # 确保数据库连接
        self.db.connect()
        self.db.create_tables()
        
        # 扫描活跃设备
        devices = self.arp_scan(network)
        
        for device in devices:
            # 保存设备信息到数据库
            device_id = self.db.insert_device(
                ip=device['ip'],
                mac=device['mac'],
                hostname=device['hostname'],
                status='up'
            )
            
            if device_id:
                # 对设备进行端口扫描
                ports_info = self.port_scan(device['ip'], scan_speed=scan_speed)
                
                # 保存端口信息到数据库
                for port in ports_info:
                    self.db.insert_port(
                        device_id=device_id,
                        port=port['port'],
                        protocol=port['protocol'],
                        status=port['status'],
                        service=port['service'],
                        application=port['application'].strip()
                    )
            
            # 控制扫描速度
            time.sleep(1)
        
        # 关闭数据库连接
        self.db.close()
        print("网络扫描完成")

class TrafficMonitor:
    def __init__(self):
        self.db = Database()

    def packet_callback(self, packet):
        """处理捕获的网络数据包"""
        try:
            if packet.haslayer('IP'):
                ip_layer = packet.getlayer('IP')
                source_ip = ip_layer.src
                destination_ip = ip_layer.dst
                protocol = ip_layer.proto
                
                # 获取传输层协议
                if packet.haslayer('TCP'):
                    tcp_layer = packet.getlayer('TCP')
                    source_port = tcp_layer.sport
                    destination_port = tcp_layer.dport
                    protocol_name = 'TCP'
                elif packet.haslayer('UDP'):
                    udp_layer = packet.getlayer('UDP')
                    source_port = udp_layer.sport
                    destination_port = udp_layer.dport
                    protocol_name = 'UDP'
                else:
                    source_port = 0
                    destination_port = 0
                    protocol_name = 'OTHER'
                
                # 数据包长度
                length = len(packet)
                
                # 保存到数据库
                self.db.insert_traffic(
                    source_ip=source_ip,
                    destination_ip=destination_ip,
                    source_port=source_port,
                    destination_port=destination_port,
                    protocol=protocol_name,
                    length=length
                )
        except Exception as e:
            print(f"处理数据包错误: {e}")

    def start_monitoring(self, interface=None):
        """开始流量监听"""
        from scapy.all import sniff
        
        # 确保数据库连接
        self.db.connect()
        self.db.create_tables()
        
        print("开始流量监听...")
        print("按 Ctrl+C 停止监听")
        
        try:
            # 开始捕获数据包
            sniff(iface=interface, prn=self.packet_callback, store=0)
        except KeyboardInterrupt:
            print("\n流量监听已停止")
        finally:
            # 关闭数据库连接
            self.db.close()
