#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
MQTT 报文分析工具

使用 pyshark 库捕获和分析 MQTT 客户端与服务器之间的交互报文
"""

import argparse
import time
import pyshark
import json
from datetime import datetime

class MQTTPacketAnalyzer:
    """MQTT 报文分析器类"""
    
    def __init__(self, interface=None, host=None, port=1883, output_file=None):
        """
        初始化报文分析器
        
        Args:
            interface: 网络接口名称
            host: 过滤特定主机
            port: MQTT 端口号
            output_file: 输出文件路径
        """
        self.interface = interface
        self.host = host
        self.port = port
        self.output_file = output_file
        self.packets = []
        
    def get_filter(self):
        """生成 Wireshark 过滤表达式"""
        filters = [f"mqtt and tcp.port == {self.port}"]
        if self.host:
            filters.append(f"(ip.addr == {self.host})")
        return " and ".join(filters)
    
    def analyze_packet(self, packet):
        """分析单个 MQTT 报文
        
        Args:
            packet: pyshark 捕获的报文对象
            
        Returns:
            解析后的报文信息字典
        """
        try:
            # 基本信息
            packet_info = {
                "timestamp": float(packet.sniff_timestamp),
                "formatted_time": packet.sniff_time.strftime("%Y-%m-%d %H:%M:%S.%f"),
                "source": packet.ip.src,
                "destination": packet.ip.dst,
                "protocol": "MQTT",
                "length": int(packet.length)
            }
            
            # MQTT 特定信息
            mqtt_layer = packet.mqtt
            packet_info["mqtt_type"] = self.get_mqtt_type(int(mqtt_layer.msgtype))
            packet_info["mqtt_msgtype"] = int(mqtt_layer.msgtype)
            
            # 根据消息类型解析不同字段
            if hasattr(mqtt_layer, "msgtype"):
                msgtype = int(mqtt_layer.msgtype)
                
                # CONNECT 消息
                if msgtype == 1:
                    packet_info["client_id"] = getattr(mqtt_layer, "connect_clientid", "N/A")
                    packet_info["protocol_name"] = getattr(mqtt_layer, "connect_protoname", "N/A")
                    packet_info["protocol_version"] = getattr(mqtt_layer, "connect_protolen", "N/A")
                    packet_info["clean_session"] = getattr(mqtt_layer, "connect_cleansess", "N/A")
                    packet_info["keep_alive"] = getattr(mqtt_layer, "connect_keepalive", "N/A")
                
                # CONNACK 消息
                elif msgtype == 2:
                    packet_info["connack_flags"] = getattr(mqtt_layer, "connack_flags", "N/A")
                    packet_info["connack_returncode"] = getattr(mqtt_layer, "connack_returncode", "N/A")
                
                # PUBLISH 消息
                elif msgtype == 3:
                    packet_info["topic"] = getattr(mqtt_layer, "topic", "N/A")
                    packet_info["qos"] = getattr(mqtt_layer, "qos", "N/A")
                    packet_info["retain"] = getattr(mqtt_layer, "retain", "N/A")
                    packet_info["dup"] = getattr(mqtt_layer, "dup", "N/A")
                    if hasattr(mqtt_layer, "value"):
                        packet_info["payload"] = getattr(mqtt_layer, "value", "N/A")
                
                # PUBACK 消息
                elif msgtype == 4:
                    packet_info["message_id"] = getattr(mqtt_layer, "msgident", "N/A")
                
                # PUBREC 消息
                elif msgtype == 5:
                    packet_info["message_id"] = getattr(mqtt_layer, "msgident", "N/A")
                
                # PUBREL 消息
                elif msgtype == 6:
                    packet_info["message_id"] = getattr(mqtt_layer, "msgident", "N/A")
                
                # PUBCOMP 消息
                elif msgtype == 7:
                    packet_info["message_id"] = getattr(mqtt_layer, "msgident", "N/A")
                
                # SUBSCRIBE 消息
                elif msgtype == 8:
                    packet_info["message_id"] = getattr(mqtt_layer, "msgident", "N/A")
                    if hasattr(mqtt_layer, "topic"):
                        packet_info["topic"] = getattr(mqtt_layer, "topic", "N/A")
                    if hasattr(mqtt_layer, "qos"):
                        packet_info["qos"] = getattr(mqtt_layer, "qos", "N/A")
                
                # SUBACK 消息
                elif msgtype == 9:
                    packet_info["message_id"] = getattr(mqtt_layer, "msgident", "N/A")
                    packet_info["return_codes"] = getattr(mqtt_layer, "suback_returncodes", "N/A")
                
                # UNSUBSCRIBE 消息
                elif msgtype == 10:
                    packet_info["message_id"] = getattr(mqtt_layer, "msgident", "N/A")
                    if hasattr(mqtt_layer, "topic"):
                        packet_info["topic"] = getattr(mqtt_layer, "topic", "N/A")
                
                # UNSUBACK 消息
                elif msgtype == 11:
                    packet_info["message_id"] = getattr(mqtt_layer, "msgident", "N/A")
                
                # PINGREQ 消息
                elif msgtype == 12:
                    pass
                
                # PINGRESP 消息
                elif msgtype == 13:
                    pass
                
                # DISCONNECT 消息
                elif msgtype == 14:
                    pass
        
        except Exception as e:
            packet_info = {
                "timestamp": float(packet.sniff_timestamp),
                "formatted_time": packet.sniff_time.strftime("%Y-%m-%d %H:%M:%S.%f"),
                "source": packet.ip.src,
                "destination": packet.ip.dst,
                "protocol": "MQTT",
                "length": int(packet.length),
                "error": str(e)
            }
        
        return packet_info
    
    def get_mqtt_type(self, msgtype):
        """根据消息类型代码获取消息类型名称
        
        Args:
            msgtype: 消息类型代码
            
        Returns:
            消息类型名称
        """
        mqtt_types = {
            1: "CONNECT",
            2: "CONNACK",
            3: "PUBLISH",
            4: "PUBACK",
            5: "PUBREC",
            6: "PUBREL",
            7: "PUBCOMP",
            8: "SUBSCRIBE",
            9: "SUBACK",
            10: "UNSUBSCRIBE",
            11: "UNSUBACK",
            12: "PINGREQ",
            13: "PINGRESP",
            14: "DISCONNECT"
        }
        return mqtt_types.get(msgtype, f"UNKNOWN({msgtype})")
    
    def print_packet_summary(self, packet_info):
        """打印报文摘要
        
        Args:
            packet_info: 解析后的报文信息字典
        """
        direction = "→" if "error" not in packet_info else "⚠️"
        print(f"[{packet_info['formatted_time']}] {packet_info['source']} {direction} {packet_info['destination']} ")
        print(f"  Type: {packet_info.get('mqtt_type', 'UNKNOWN')} ({packet_info.get('mqtt_msgtype', 'N/A')})")
        print(f"  Length: {packet_info['length']} bytes")
        
        # 打印特定类型的详细信息
        if packet_info.get('mqtt_msgtype') == 1:  # CONNECT
            print(f"  Client ID: {packet_info.get('client_id', 'N/A')}")
            print(f"  Protocol: {packet_info.get('protocol_name', 'N/A')} v{packet_info.get('protocol_version', 'N/A')}")
        elif packet_info.get('mqtt_msgtype') == 3:  # PUBLISH
            print(f"  Topic: {packet_info.get('topic', 'N/A')}")
            print(f"  QoS: {packet_info.get('qos', 'N/A')}")
            print(f"  Payload: {packet_info.get('payload', 'N/A')[:50]}{'...' if len(packet_info.get('payload', '')) > 50 else ''}")
        elif packet_info.get('mqtt_msgtype') in [8, 10]:  # SUBSCRIBE, UNSUBSCRIBE
            print(f"  Topic: {packet_info.get('topic', 'N/A')}")
        
        print("  " + "-" * 50)
    
    def capture_packets(self, duration=10):
        """捕获 MQTT 报文
        
        Args:
            duration: 捕获时长（秒）
        """
        print(f"=== MQTT 报文分析工具 ===")
        print(f"捕获接口: {self.interface or '默认接口'}")
        print(f"过滤条件: {self.get_filter()}")
        print(f"捕获时长: {duration} 秒")
        print("开始捕获 MQTT 报文...\n")
        
        # 构建捕获参数
        capture_params = {
            'display_filter': self.get_filter(),
            'timeout': duration
        }
        
        if self.interface:
            capture_params['interface'] = self.interface
        
        try:
            # 开始捕获
            capture = pyshark.LiveCapture(**capture_params)
            
            start_time = time.time()
            
            for packet in capture.sniff_continuously():
                if time.time() - start_time > duration:
                    break
                
                if 'mqtt' in packet:
                    packet_info = self.analyze_packet(packet)
                    self.packets.append(packet_info)
                    self.print_packet_summary(packet_info)
        
        except KeyboardInterrupt:
            print("\n捕获被用户中断")
        except Exception as e:
            print(f"捕获过程中发生错误: {e}")
        
        print(f"\n捕获完成，共捕获 {len(self.packets)} 个 MQTT 报文")
        
        # 保存结果到文件
        if self.output_file:
            self.save_results()
    
    def save_results(self):
        """保存分析结果到文件"""
        try:
            with open(self.output_file, 'w') as f:
                json.dump(self.packets, f, indent=2, default=str)
            print(f"分析结果已保存到: {self.output_file}")
        except Exception as e:
            print(f"保存结果失败: {e}")
    
    def generate_report(self):
        """生成分析报告"""
        if not self.packets:
            print("没有捕获到 MQTT 报文")
            return
        
        print("\n=== MQTT 报文分析报告 ===")
        print(f"总报文数: {len(self.packets)}")
        
        # 统计不同类型的报文数量
        type_count = {}
        for packet in self.packets:
            msg_type = packet.get('mqtt_type', 'UNKNOWN')
            type_count[msg_type] = type_count.get(msg_type, 0) + 1
        
        print("\n报文类型分布:")
        for msg_type, count in type_count.items():
            print(f"  {msg_type}: {count} 个 ({count/len(self.packets)*100:.1f}%)")
        
        # 统计通信双方
        source_count = {}
        for packet in self.packets:
            source = packet['source']
            source_count[source] = source_count.get(source, 0) + 1
        
        print("\n通信方统计:")
        for source, count in source_count.items():
            print(f"  {source}: {count} 个报文")
        
        # 统计主题分布
        topic_count = {}
        for packet in self.packets:
            if packet.get('mqtt_msgtype') == 3:  # PUBLISH
                topic = packet.get('topic', 'N/A')
                topic_count[topic] = topic_count.get(topic, 0) + 1
        
        if topic_count:
            print("\n主题分布:")
            for topic, count in topic_count.items():
                print(f"  {topic}: {count} 条消息")

def main():
    """主函数"""
    parser = argparse.ArgumentParser(description='MQTT 报文分析工具')
    parser.add_argument('-i', '--interface', type=str, help='网络接口名称')
    parser.add_argument('-H', '--host', type=str, help='过滤特定主机')
    parser.add_argument('-p', '--port', type=int, default=1883, help='MQTT 端口号')
    parser.add_argument('-d', '--duration', type=int, default=10, help='捕获时长（秒）')
    parser.add_argument('-o', '--output', type=str, help='输出文件路径')
    
    args = parser.parse_args()
    
    # 创建报文分析器
    analyzer = MQTTPacketAnalyzer(
        interface=args.interface,
        host=args.host,
        port=args.port,
        output_file=args.output
    )
    
    # 捕获报文
    analyzer.capture_packets(duration=args.duration)
    
    # 生成报告
    analyzer.generate_report()

if __name__ == '__main__':
    main()
