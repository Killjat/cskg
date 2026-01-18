#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
简易 MQTT 报文分析工具

使用 paho-mqtt 客户端库实现，不需要 root 权限
通过日志记录和分析 MQTT 客户端与服务器之间的交互
"""

import paho.mqtt.client as mqtt
import time
import json
import argparse
from datetime import datetime

class SimpleMQTTPacketAnalyzer:
    """简易 MQTT 报文分析器类"""
    
    def __init__(self, broker, port=1883, topic="#", client_id_prefix="analyzer"):
        """
        初始化报文分析器
        
        Args:
            broker: MQTT 服务器地址
            port: MQTT 端口
            topic: 订阅的主题
            client_id_prefix: 客户端 ID 前缀
        """
        self.broker = broker
        self.port = port
        self.topic = topic
        self.client_id = f"{client_id_prefix}-{int(time.time())}"
        self.packets = []
        self.start_time = None
        
        # 创建 MQTT 客户端
        self.client = mqtt.Client(client_id=self.client_id, protocol=mqtt.MQTTv311)
        
        # 注册回调函数
        self.client.on_connect = self._on_connect
        self.client.on_disconnect = self._on_disconnect
        self.client.on_publish = self._on_publish
        self.client.on_subscribe = self._on_subscribe
        self.client.on_unsubscribe = self._on_unsubscribe
        self.client.on_message = self._on_message
        self.client.on_log = self._on_log
    
    def _on_connect(self, client, userdata, flags, rc):
        """连接回调"""
        self._log_packet({
            "type": "CONNECT_ACK",
            "direction": "IN",
            "rc": rc,
            "flags": flags
        })
    
    def _on_disconnect(self, client, userdata, rc):
        """断开连接回调"""
        self._log_packet({
            "type": "DISCONNECT",
            "direction": "OUT",
            "rc": rc
        })
    
    def _on_publish(self, client, userdata, mid):
        """发布回调"""
        self._log_packet({
            "type": "PUBLISH_ACK",
            "direction": "IN",
            "mid": mid
        })
    
    def _on_subscribe(self, client, userdata, mid, granted_qos):
        """订阅回调"""
        self._log_packet({
            "type": "SUBSCRIBE_ACK",
            "direction": "IN",
            "mid": mid,
            "granted_qos": granted_qos
        })
    
    def _on_unsubscribe(self, client, userdata, mid):
        """取消订阅回调"""
        self._log_packet({
            "type": "UNSUBSCRIBE_ACK",
            "direction": "IN",
            "mid": mid
        })
    
    def _on_message(self, client, userdata, msg):
        """消息接收回调"""
        self._log_packet({
            "type": "PUBLISH",
            "direction": "IN",
            "topic": msg.topic,
            "payload": msg.payload.decode(),
            "qos": msg.qos,
            "retain": msg.retain
        })
    
    def _on_log(self, client, userdata, level, buf):
        """日志回调"""
        if level == mqtt.MQTT_LOG_DEBUG:
            # 解析 MQTT 调试日志，提取报文信息
            self._parse_debug_log(buf)
    
    def _parse_debug_log(self, log_msg):
        """解析 MQTT 调试日志
        
        Args:
            log_msg: 调试日志字符串
        """
        try:
            if "Sending PUBLISH" in log_msg:
                # 示例: Sending PUBLISH (d0, q0, r0, m1, 'test/topic', ... (13 bytes))
                parts = log_msg.split()
                topic = parts[5].strip("'")
                mid = int(parts[4].strip(','))
                qos = int(parts[2][1])
                retain = parts[3] == "r1"
                self._log_packet({
                    "type": "PUBLISH",
                    "direction": "OUT",
                    "topic": topic,
                    "mid": mid,
                    "qos": qos,
                    "retain": retain
                })
            elif "Sending CONNECT" in log_msg:
                self._log_packet({
                    "type": "CONNECT",
                    "direction": "OUT"
                })
            elif "Sending SUBSCRIBE" in log_msg:
                # 示例: Sending SUBSCRIBE (d0, m2) [(test/topic, 0)]
                parts = log_msg.split()
                mid = int(parts[4].strip(')'))
                self._log_packet({
                    "type": "SUBSCRIBE",
                    "direction": "OUT",
                    "mid": mid
                })
            elif "Sending DISCONNECT" in log_msg:
                self._log_packet({
                    "type": "DISCONNECT",
                    "direction": "OUT"
                })
        except Exception as e:
            # 忽略解析错误
            pass
    
    def _log_packet(self, packet_info):
        """记录报文信息
        
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
        direction = "→" if packet["direction"] == "OUT" else "←"
        print(f"[{packet['formatted_time']}] {direction} {packet['type']}")
        
        if packet["type"] == "PUBLISH":
            print(f"  Topic: {packet['topic']}")
            if "payload" in packet:
                payload = packet['payload']
                print(f"  Payload: {payload[:50]}{'...' if len(payload) > 50 else ''}")
            print(f"  QoS: {packet.get('qos', 0)}, Retain: {packet.get('retain', False)}")
        elif "rc" in packet:
            print(f"  Return Code: {packet['rc']}")
        elif "mid" in packet:
            print(f"  Message ID: {packet['mid']}")
        
        print("  " + "-" * 50)
    
    def start(self, duration=10):
        """启动报文分析
        
        Args:
            duration: 分析时长（秒）
        """
        print("=== 简易 MQTT 报文分析工具 ===")
        print(f"连接到 MQTT 服务器: {self.broker}:{self.port}")
        print(f"订阅主题: {self.topic}")
        print(f"分析时长: {duration} 秒")
        print(f"客户端 ID: {self.client_id}")
        print()
        
        self.start_time = time.time()
        
        try:
            # 连接到 MQTT 服务器
            self.client.connect(self.broker, self.port, keepalive=60)
            
            # 启用调试日志
            self.client.enable_logger()
            
            # 订阅主题
            self.client.subscribe(self.topic)
            
            # 启动消息循环
            self.client.loop_start()
            
            # 等待指定时长
            time.sleep(duration)
            
        except KeyboardInterrupt:
            print("\n分析被用户中断")
        except Exception as e:
            print(f"\n分析过程中发生错误: {e}")
        finally:
            # 停止消息循环
            self.client.loop_stop()
            # 断开连接
            self.client.disconnect()
    
    def save_results(self, output_file):
        """保存分析结果到文件
        
        Args:
            output_file: 输出文件路径
        """
        try:
            with open(output_file, 'w') as f:
                json.dump(self.packets, f, indent=2, default=str)
            print(f"\n分析结果已保存到: {output_file}")
            print(f"共捕获 {len(self.packets)} 个 MQTT 报文")
        except Exception as e:
            print(f"\n保存结果失败: {e}")
    
    def generate_report(self):
        """生成分析报告"""
        if not self.packets:
            print("没有捕获到 MQTT 报文")
            return
        
        print("\n=== MQTT 报文分析报告 ===")
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
            percentage = count / len(self.packets) * 100
            print(f"  {'发送' if direction == 'OUT' else '接收'}: {count} 个 ({percentage:.1f}%)")
        
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

def main():
    """主函数"""
    parser = argparse.ArgumentParser(description='简易 MQTT 报文分析工具')
    parser.add_argument('--broker', type=str, default='test.mosquitto.org',
                        help='MQTT 服务器地址')
    parser.add_argument('--port', type=int, default=1883,
                        help='MQTT 端口号')
    parser.add_argument('--topic', type=str, default='#',
                        help='订阅的主题（默认订阅所有主题）')
    parser.add_argument('--duration', type=int, default=15,
                        help='分析时长（秒）')
    parser.add_argument('--output', type=str, default='mqtt_simple_analysis.json',
                        help='输出文件路径')
    
    args = parser.parse_args()
    
    # 创建并启动分析器
    analyzer = SimpleMQTTPacketAnalyzer(
        broker=args.broker,
        port=args.port,
        topic=args.topic
    )
    
    analyzer.start(duration=args.duration)
    analyzer.generate_report()
    analyzer.save_results(args.output)

if __name__ == '__main__':
    main()
