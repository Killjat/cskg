#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
MQTT 协议分析工具

基于 paho-mqtt 调试功能实现，不需要 root 权限
通过解析 MQTT 客户端的调试日志来模拟网络数据包分析
"""

import paho.mqtt.client as mqtt
import time
import json
import argparse
from datetime import datetime

class MQTTProtocolAnalyzer:
    """MQTT 协议分析器类"""
    
    def __init__(self, broker, port=1883, topic="iot/test/topic"):
        """
        初始化协议分析器
        
        Args:
            broker: MQTT 服务器地址
            port: MQTT 端口
            topic: 测试主题
        """
        self.broker = broker
        self.port = port
        self.topic = topic
        self.client_id = f"analyzer-{int(time.time())}"
        self.packets = []
        self.start_time = None
        self.protocol_versions = {
            3: "MQTTv3.1",
            4: "MQTTv3.1.1",
            5: "MQTTv5.0"
        }
        
        # 创建 MQTT 客户端
        self.client = mqtt.Client(client_id=self.client_id, protocol=mqtt.MQTTv311)
        
        # 注册回调函数
        self.client.on_connect = self._on_connect
        self.client.on_message = self._on_message
        self.client.on_publish = self._on_publish
        self.client.on_subscribe = self._on_subscribe
        self.client.on_log = self._on_log
        
        # 启用调试日志
        import logging
        self.logger = logging.getLogger(__name__)
        self.logger.setLevel(logging.DEBUG)
        
    def _on_connect(self, client, userdata, flags, rc):
        """连接回调"""
        self._log_protocol_packet({
            "type": "CONNACK",
            "direction": "IN",
            "packet_type": 2,
            "return_code": rc,
            "flags": flags,
            "return_code_str": self._get_connack_rc_str(rc)
        })
    
    def _on_message(self, client, userdata, msg):
        """消息接收回调"""
        self._log_protocol_packet({
            "type": "PUBLISH",
            "direction": "IN",
            "packet_type": 3,
            "topic": msg.topic,
            "payload": msg.payload.decode(),
            "qos": msg.qos,
            "retain": msg.retain,
            "dup": msg.dup
        })
    
    def _on_publish(self, client, userdata, mid):
        """发布回调"""
        self._log_protocol_packet({
            "type": "PUBACK",
            "direction": "IN",
            "packet_type": 4,
            "message_id": mid
        })
    
    def _on_subscribe(self, client, userdata, mid, granted_qos):
        """订阅回调"""
        self._log_protocol_packet({
            "type": "SUBACK",
            "direction": "IN",
            "packet_type": 9,
            "message_id": mid,
            "granted_qos": granted_qos
        })
    
    def _on_log(self, client, userdata, level, buf):
        """日志回调"""
        if level == mqtt.MQTT_LOG_DEBUG:
            # 解析 MQTT 调试日志，提取协议报文信息
            self._parse_debug_log(buf)
    
    def _parse_debug_log(self, log_msg):
        """解析 MQTT 调试日志
        
        Args:
            log_msg: 调试日志字符串
        """
        try:
            if "Sending PUBLISH" in log_msg:
                # 解析发送的 PUBLISH 报文
                # 示例: Sending PUBLISH (d0, q0, r0, m1, 'test/topic', ... (13 bytes))
                parts = log_msg.split()
                direction = "OUT"
                packet_type = 3
                dup = parts[2][1] == '1'
                qos = int(parts[3][1])
                retain = parts[4] == "r1"
                mid = int(parts[5].strip(','))
                topic = parts[6].strip("'")
                payload_len = int(parts[8].strip('('))
                
                self._log_protocol_packet({
                    "type": "PUBLISH",
                    "direction": direction,
                    "packet_type": packet_type,
                    "topic": topic,
                    "message_id": mid,
                    "qos": qos,
                    "retain": retain,
                    "dup": dup,
                    "payload_length": payload_len
                })
            
            elif "Sending CONNECT" in log_msg:
                # 解析发送的 CONNECT 报文
                direction = "OUT"
                packet_type = 1
                
                self._log_protocol_packet({
                    "type": "CONNECT",
                    "direction": direction,
                    "packet_type": packet_type,
                    "protocol_name": "MQTT",
                    "protocol_version": 4,  # MQTTv3.1.1
                    "protocol_version_str": "MQTTv3.1.1",
                    "clean_session": True,
                    "keep_alive": 60
                })
            
            elif "Sending SUBSCRIBE" in log_msg:
                # 解析发送的 SUBSCRIBE 报文
                # 示例: Sending SUBSCRIBE (d0, m2) [(test/topic, 0)]
                parts = log_msg.split()
                direction = "OUT"
                packet_type = 8
                mid = int(parts[4].strip(')'))
                
                self._log_protocol_packet({
                    "type": "SUBSCRIBE",
                    "direction": direction,
                    "packet_type": packet_type,
                    "message_id": mid
                })
            
            elif "Sending DISCONNECT" in log_msg:
                # 解析发送的 DISCONNECT 报文
                direction = "OUT"
                packet_type = 14
                
                self._log_protocol_packet({
                    "type": "DISCONNECT",
                    "direction": direction,
                    "packet_type": packet_type
                })
            
            elif "Received PUBLISH" in log_msg:
                # 解析接收的 PUBLISH 报文
                # 示例: Received PUBLISH (d0, q0, r0, m0, 'test/topic', ... (13 bytes))
                parts = log_msg.split()
                direction = "IN"
                packet_type = 3
                dup = parts[2][1] == '1'
                qos = int(parts[3][1])
                retain = parts[4] == "r1"
                mid = int(parts[5].strip(','))
                topic = parts[6].strip("'")
                payload_len = int(parts[8].strip('('))
                
                self._log_protocol_packet({
                    "type": "PUBLISH",
                    "direction": direction,
                    "packet_type": packet_type,
                    "topic": topic,
                    "message_id": mid,
                    "qos": qos,
                    "retain": retain,
                    "dup": dup,
                    "payload_length": payload_len
                })
        
        except Exception as e:
            # 忽略解析错误，记录原始日志
            self._log_protocol_packet({
                "type": "DEBUG_LOG",
                "direction": "LOG",
                "raw_log": log_msg
            })
    
    def _log_protocol_packet(self, packet_info):
        """记录协议报文信息
        
        Args:
            packet_info: 报文信息字典
        """
        timestamp = time.time()
        if self.start_time is None:
            self.start_time = timestamp
        
        packet = {
            "timestamp": timestamp,
            "relative_time": timestamp - self.start_time,
            "formatted_time": datetime.now().strftime("%Y-%m-%d %H:%M:%S.%f"),
            "broker": self.broker,
            "broker_port": self.port,
            "client_id": self.client_id
        }
        packet.update(packet_info)
        
        self.packets.append(packet)
        self._print_packet_summary(packet)
    
    def _print_packet_summary(self, packet):
        """打印报文摘要
        
        Args:
            packet: 报文信息字典
        """
        direction_symbol = "→" if packet["direction"] == "OUT" else "←"
        print(f"[{packet['formatted_time']}] {direction_symbol} {packet['type']} (Type: {packet.get('packet_type', 'N/A')})")
        
        if packet["type"] == "CONNECT":
            print(f"  Protocol: {packet.get('protocol_name', 'N/A')} {packet.get('protocol_version_str', 'N/A')}")
            print(f"  Clean Session: {packet.get('clean_session', 'N/A')}")
            print(f"  Keep Alive: {packet.get('keep_alive', 'N/A')} seconds")
        elif packet["type"] == "CONNACK":
            print(f"  Return Code: {packet.get('return_code', 'N/A')} ({packet.get('return_code_str', 'N/A')})")
        elif packet["type"] == "PUBLISH":
            print(f"  Topic: {packet.get('topic', 'N/A')}")
            print(f"  QoS: {packet.get('qos', 'N/A')}")
            print(f"  Retain: {packet.get('retain', 'N/A')}")
            print(f"  DUP: {packet.get('dup', 'N/A')}")
            if "payload" in packet:
                payload = packet['payload']
                print(f"  Payload: {payload[:50]}{'...' if len(payload) > 50 else ''}")
            if "payload_length" in packet:
                print(f"  Payload Length: {packet['payload_length']} bytes")
        elif "message_id" in packet:
            print(f"  Message ID: {packet['message_id']}")
        
        print("  " + "=" * 60)
    
    def _get_connack_rc_str(self, rc):
        """获取 CONNACK 返回码的字符串描述
        
        Args:
            rc: 返回码
            
        Returns:
            返回码描述
        """
        rc_map = {
            0: "Connection accepted",
            1: "Connection refused, unacceptable protocol version",
            2: "Connection refused, identifier rejected",
            3: "Connection refused, server unavailable",
            4: "Connection refused, bad user name or password",
            5: "Connection refused, not authorized"
        }
        return rc_map.get(rc, f"Unknown return code: {rc}")
    
    def connect_and_analyze(self, duration=10):
        """连接到 MQTT 服务器并开始分析
        
        Args:
            duration: 分析时长（秒）
        """
        print("=== MQTT 协议分析工具 ===")
        print(f"MQTT 服务器: {self.broker}:{self.port}")
        print(f"测试主题: {self.topic}")
        print(f"分析时长: {duration} 秒")
        print(f"客户端 ID: {self.client_id}")
        print()
        
        self.start_time = time.time()
        
        try:
            # 连接到 MQTT 服务器
            self.client.connect(self.broker, self.port, keepalive=60)
            
            # 启用调试日志
            self.client.enable_logger(self.logger)
            
            # 订阅主题
            self.client.subscribe(self.topic, qos=0)
            
            # 启动消息循环
            self.client.loop_start()
            
            # 发布测试消息
            self.client.publish(self.topic, "Protocol analyzer test message 1")
            time.sleep(2)
            self.client.publish(self.topic, "Protocol analyzer test message 2")
            time.sleep(2)
            self.client.publish(self.topic, "Protocol analyzer test message 3")
            
            # 等待指定时长
            remaining_time = duration - 6  # 减去已用时间
            if remaining_time > 0:
                time.sleep(remaining_time)
            
        except KeyboardInterrupt:
            print("\n分析被用户中断")
        except Exception as e:
            print(f"\n分析过程中发生错误: {e}")
        finally:
            # 停止消息循环
            self.client.loop_stop()
            # 断开连接
            self.client.disconnect()
    
    def generate_protocol_report(self):
        """生成协议分析报告"""
        if not self.packets:
            print("没有捕获到 MQTT 协议报文")
            return
        
        print("\n=== MQTT 协议分析报告 ===")
        print(f"总报文数: {len(self.packets)}")
        print(f"分析时长: {self.packets[-1]['relative_time']:.2f} 秒")
        print(f"MQTT 服务器: {self.broker}:{self.port}")
        
        # 统计不同类型的报文
        type_count = {}
        for packet in self.packets:
            msg_type = packet["type"]
            type_count[msg_type] = type_count.get(msg_type, 0) + 1
        
        print("\n报文类型分布:")
        for msg_type, count in type_count.items():
            percentage = count / len(self.packets) * 100
            print(f"  {msg_type}: {count} 个 ({percentage:.1f}%)")
        
        # 统计收发方向
        direction_count = {}
        for packet in self.packets:
            direction = packet["direction"]
            direction_count[direction] = direction_count.get(direction, 0) + 1
        
        print("\n报文方向分布:")
        for direction, count in direction_count.items():
            direction_str = "发送" if direction == "OUT" else "接收"
            percentage = count / len(self.packets) * 100
            print(f"  {direction_str}: {count} 个 ({percentage:.1f}%)")
        
        # 统计主题分布
        topic_count = {}
        for packet in self.packets:
            if "topic" in packet:
                topic = packet["topic"]
                topic_count[topic] = topic_count.get(topic, 0) + 1
        
        if topic_count:
            print("\n主题分布:")
            for topic, count in topic_count.items():
                print(f"  {topic}: {count} 个")
        
        # 统计 QoS 分布
        qos_count = {}
        for packet in self.packets:
            if "qos" in packet:
                qos = packet["qos"]
                qos_count[qos] = qos_count.get(qos, 0) + 1
        
        if qos_count:
            print("\nQoS 分布:")
            for qos, count in qos_count.items():
                print(f"  QoS {qos}: {count} 个")
    
    def save_analysis(self, output_file):
        """保存分析结果到文件
        
        Args:
            output_file: 输出文件路径
        """
        try:
            with open(output_file, 'w') as f:
                json.dump(self.packets, f, indent=2, default=str)
            print(f"\n分析结果已保存到: {output_file}")
            print(f"共捕获 {len(self.packets)} 个 MQTT 协议报文")
        except Exception as e:
            print(f"\n保存结果失败: {e}")

def main():
    """主函数"""
    parser = argparse.ArgumentParser(description='MQTT 协议分析工具')
    parser.add_argument('--broker', type=str, default='test.mosquitto.org',
                        help='MQTT 服务器地址')
    parser.add_argument('--port', type=int, default=1883,
                        help='MQTT 端口号')
    parser.add_argument('--topic', type=str, default='iot/test/topic',
                        help='测试主题')
    parser.add_argument('--duration', type=int, default=15,
                        help='分析时长（秒）')
    parser.add_argument('--output', type=str, default='mqtt_protocol_analysis.json',
                        help='输出文件路径')
    
    args = parser.parse_args()
    
    # 创建并启动协议分析器
    analyzer = MQTTProtocolAnalyzer(
        broker=args.broker,
        port=args.port,
        topic=args.topic
    )
    
    analyzer.connect_and_analyze(duration=args.duration)
    analyzer.generate_protocol_report()
    analyzer.save_analysis(args.output)

if __name__ == '__main__':
    main()
