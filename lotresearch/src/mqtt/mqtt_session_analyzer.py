#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
MQTT 会话分析工具

模拟 MQTT 客户端和服务器之间的完整会话，进行全流量解析和会话行为分析
"""

import paho.mqtt.client as mqtt
import time
import json
import argparse
from datetime import datetime
from collections import defaultdict

class MQTTMessage:
    """MQTT 消息类"""
    
    def __init__(self, msg_type, direction, timestamp, details):
        """
        初始化 MQTT 消息
        
        Args:
            msg_type: 消息类型
            direction: 消息方向（IN/OUT）
            timestamp: 消息时间戳
            details: 消息详细信息
        """
        self.msg_type = msg_type
        self.direction = direction
        self.timestamp = timestamp
        self.details = details
        self.formatted_time = datetime.fromtimestamp(timestamp).strftime("%Y-%m-%d %H:%M:%S.%f")[:-3]
        self.relative_time = 0.0
    
    def __str__(self):
        """字符串表示"""
        direction_symbol = "→" if self.direction == "OUT" else "←"
        return f"[{self.formatted_time}] {direction_symbol} {self.msg_type}"
    
    def to_dict(self):
        """转换为字典"""
        return {
            "msg_type": self.msg_type,
            "direction": self.direction,
            "timestamp": self.timestamp,
            "formatted_time": self.formatted_time,
            "relative_time": self.relative_time,
            "details": self.details
        }

class MQTTSessionAnalyzer:
    """MQTT 会话分析器类"""
    
    def __init__(self, broker, port=1883, topic="iot/test/topic", client_id_prefix="analyzer"):
        """
        初始化会话分析器
        
        Args:
            broker: MQTT 服务器地址
            port: MQTT 端口
            topic: 测试主题
            client_id_prefix: 客户端 ID 前缀
        """
        self.broker = broker
        self.port = port
        self.topic = topic
        self.client_id = f"{client_id_prefix}-{int(time.time())}"
        self.messages = []
        self.start_time = None
        self.session_start = None
        self.session_end = None
        self.session_duration = 0.0
        
        # 创建 MQTT 客户端
        self.client = mqtt.Client(client_id=self.client_id, protocol=mqtt.MQTTv311)
        
        # 注册回调函数
        self.client.on_connect = self._on_connect
        self.client.on_message = self._on_message
        self.client.on_publish = self._on_publish
        self.client.on_subscribe = self._on_subscribe
        self.client.on_unsubscribe = self._on_unsubscribe
        self.client.on_log = self._on_log
        self.client.on_disconnect = self._on_disconnect
        
        # 会话统计信息
        self.stats = {
            "total_messages": 0,
            "message_types": defaultdict(int),
            "message_directions": defaultdict(int),
            "topics": defaultdict(int),
            "qos_distribution": defaultdict(int),
            "retain_distribution": defaultdict(int),
            "dup_distribution": defaultdict(int),
            "session_events": []
        }
    
    def _on_connect(self, client, userdata, flags, rc):
        """连接回调"""
        timestamp = time.time()
        self._add_message(MQTTMessage(
            "CONNECT_ACK",
            "IN",
            timestamp,
            {
                "return_code": rc,
                "return_code_str": self._get_connack_rc_str(rc),
                "flags": flags,
                "clean_session": flags.get("clean_session", False),
                "session_present": flags.get("session_present", False)
            }
        ))
        
        self.stats["session_events"].append({
            "event": "CONNECTED",
            "timestamp": timestamp,
            "details": {
                "return_code": rc,
                "return_code_str": self._get_connack_rc_str(rc)
            }
        })
    
    def _on_message(self, client, userdata, msg):
        """消息接收回调"""
        timestamp = time.time()
        self._add_message(MQTTMessage(
            "PUBLISH",
            "IN",
            timestamp,
            {
                "topic": msg.topic,
                "payload": msg.payload.decode(),
                "payload_length": len(msg.payload),
                "qos": msg.qos,
                "retain": msg.retain,
                "dup": msg.dup
            }
        ))
        
        # 更新统计信息
        self.stats["topics"][msg.topic] += 1
        self.stats["qos_distribution"][msg.qos] += 1
        self.stats["retain_distribution"][msg.retain] += 1
        self.stats["dup_distribution"][msg.dup] += 1
    
    def _on_publish(self, client, userdata, mid):
        """发布回调"""
        timestamp = time.time()
        self._add_message(MQTTMessage(
            "PUBACK",
            "IN",
            timestamp,
            {
                "message_id": mid
            }
        ))
    
    def _on_subscribe(self, client, userdata, mid, granted_qos):
        """订阅回调"""
        timestamp = time.time()
        self._add_message(MQTTMessage(
            "SUBACK",
            "IN",
            timestamp,
            {
                "message_id": mid,
                "granted_qos": granted_qos
            }
        ))
    
    def _on_unsubscribe(self, client, userdata, mid):
        """取消订阅回调"""
        timestamp = time.time()
        self._add_message(MQTTMessage(
            "UNSUBACK",
            "IN",
            timestamp,
            {
                "message_id": mid
            }
        ))
    
    def _on_disconnect(self, client, userdata, rc):
        """断开连接回调"""
        timestamp = time.time()
        self._add_message(MQTTMessage(
            "DISCONNECT",
            "OUT",
            timestamp,
            {
                "return_code": rc
            }
        ))
        
        self.stats["session_events"].append({
            "event": "DISCONNECTED",
            "timestamp": timestamp,
            "details": {
                "return_code": rc
            }
        })
    
    def _on_log(self, client, userdata, level, buf):
        """日志回调"""
        if level == mqtt.MQTT_LOG_DEBUG:
            timestamp = time.time()
            if "Sending CONNECT" in buf:
                self._add_message(MQTTMessage(
                    "CONNECT",
                    "OUT",
                    timestamp,
                    {
                        "protocol_name": "MQTT",
                        "protocol_version": 4,  # MQTTv3.1.1
                        "protocol_version_str": "MQTTv3.1.1"
                    }
                ))
                
                self.stats["session_events"].append({
                    "event": "CONNECT_REQUEST",
                    "timestamp": timestamp
                })
            elif "Sending PUBLISH" in buf:
                # 解析发送的 PUBLISH 报文
                parts = buf.split()
                try:
                    dup = parts[2][1] == '1'
                    qos = int(parts[3][1])
                    retain = parts[4] == "r1"
                    mid = int(parts[5].strip(','))
                    topic = parts[6].strip("'")
                    payload_len = int(parts[8].strip('('))
                    
                    self._add_message(MQTTMessage(
                        "PUBLISH",
                        "OUT",
                        timestamp,
                        {
                            "topic": topic,
                            "message_id": mid,
                            "qos": qos,
                            "retain": retain,
                            "dup": dup,
                            "payload_length": payload_len
                        }
                    ))
                except Exception as e:
                    pass
            elif "Sending SUBSCRIBE" in buf:
                # 解析发送的 SUBSCRIBE 报文
                parts = buf.split()
                try:
                    mid = int(parts[4].strip(')'))
                    
                    self._add_message(MQTTMessage(
                        "SUBSCRIBE",
                        "OUT",
                        timestamp,
                        {
                            "message_id": mid
                        }
                    ))
                except Exception as e:
                    pass
            elif "Sending DISCONNECT" in buf:
                self.stats["session_events"].append({
                    "event": "DISCONNECT_REQUEST",
                    "timestamp": timestamp
                })
    
    def _add_message(self, message):
        """添加消息到会话"""
        if self.start_time is None:
            self.start_time = message.timestamp
            self.session_start = message.timestamp
        
        # 计算相对时间
        message.relative_time = message.timestamp - self.start_time
        
        self.messages.append(message)
        self.stats["total_messages"] += 1
        self.stats["message_types"][message.msg_type] += 1
        self.stats["message_directions"][message.direction] += 1
    
    def _get_connack_rc_str(self, rc):
        """获取 CONNACK 返回码的字符串描述"""
        rc_map = {
            0: "Connection accepted",
            1: "Connection refused, unacceptable protocol version",
            2: "Connection refused, identifier rejected",
            3: "Connection refused, server unavailable",
            4: "Connection refused, bad user name or password",
            5: "Connection refused, not authorized"
        }
        return rc_map.get(rc, f"Unknown return code: {rc}")
    
    def run_session(self, duration=15):
        """运行 MQTT 会话
        
        Args:
            duration: 会话持续时间（秒）
        """
        print("=== MQTT 会话分析工具 ===")
        print(f"MQTT 服务器: {self.broker}:{self.port}")
        print(f"测试主题: {self.topic}")
        print(f"会话时长: {duration} 秒")
        print(f"客户端 ID: {self.client_id}")
        print()
        print("🔄 开始 MQTT 会话...")
        
        try:
            # 连接到 MQTT 服务器
            self.client.connect(self.broker, self.port, keepalive=60)
            
            # 启用调试日志
            import logging
            logger = logging.getLogger(__name__)
            logger.setLevel(logging.DEBUG)
            self.client.enable_logger(logger)
            
            # 启动消息循环
            self.client.loop_start()
            
            # 订阅主题
            self.client.subscribe(self.topic, qos=0)
            
            # 模拟会话行为
            self._simulate_session_behavior(duration)
            
            # 记录会话结束时间
            self.session_end = time.time()
            self.session_duration = self.session_end - self.session_start
            
            # 断开连接
            self.client.disconnect()
            
            print("✅ MQTT 会话结束")
            print()
            
        except KeyboardInterrupt:
            print("\n⚠️  会话被用户中断")
        except Exception as e:
            print(f"\n❌ 会话过程中发生错误: {e}")
        finally:
            # 停止消息循环
            self.client.loop_stop()
    
    def _simulate_session_behavior(self, duration):
        """模拟 MQTT 会话行为"""
        # 发布测试消息
        test_messages = [
            "Session test message 1",
            "Session test message 2",
            "Session test message 3 with longer payload to test message handling",
            "Session test message 4",
            "Session test message 5"
        ]
        
        # 发送多条测试消息
        for i, msg in enumerate(test_messages):
            self.client.publish(self.topic, msg, qos=0)
            time.sleep(1)
        
        # 等待剩余时间
        remaining_time = duration - len(test_messages) - 2  # 减去已用时间和连接时间
        if remaining_time > 0:
            time.sleep(remaining_time)
    
    def analyze_session(self):
        """分析 MQTT 会话"""
        print("=== MQTT 会话分析报告 ===")
        print(f"会话时长: {self.session_duration:.2f} 秒")
        print(f"消息总数: {self.stats['total_messages']}")
        print(f"消息类型: {len(self.stats['message_types'])} 种")
        print(f"涉及主题: {len(self.stats['topics'])} 个")
        print()
        
        # 消息类型分布
        print("📊 消息类型分布:")
        for msg_type, count in sorted(self.stats['message_types'].items(), key=lambda x: x[1], reverse=True):
            percentage = (count / self.stats['total_messages']) * 100
            print(f"  {msg_type}: {count} 个 ({percentage:.1f}%)")
        print()
        
        # 消息方向分布
        print("🔄 消息方向分布:")
        total = self.stats['total_messages']
        for direction, count in self.stats['message_directions'].items():
            percentage = (count / total) * 100
            direction_str = "发送" if direction == "OUT" else "接收"
            print(f"  {direction_str}: {count} 个 ({percentage:.1f}%)")
        print()
        
        # QoS 分布
        if self.stats['qos_distribution']:
            print("🎯 QoS 分布:")
            for qos, count in sorted(self.stats['qos_distribution'].items()):
                print(f"  QoS {qos}: {count} 个")
            print()
        
        # 主题分布
        if self.stats['topics']:
            print("📋 主题分布:")
            for topic, count in sorted(self.stats['topics'].items(), key=lambda x: x[1], reverse=True):
                print(f"  {topic}: {count} 个")
            print()
        
        # 会话事件序列
        print("⏱️  会话事件序列:")
        for event in self.stats['session_events']:
            event_time = datetime.fromtimestamp(event['timestamp']).strftime("%H:%M:%S.%f")[:-3]
            print(f"  [{event_time}] {event['event']}")
            if 'details' in event:
                for key, value in event['details'].items():
                    print(f"    {key}: {value}")
        print()
        
        # 消息时序图
        print("📈 消息时序图:")
        for msg in self.messages:
            direction_symbol = "→" if msg.direction == "OUT" else "←"
            print(f"  [{msg.relative_time:.3f}s] {direction_symbol} {msg.msg_type}")
            if msg.msg_type == "PUBLISH":
                print(f"    Topic: {msg.details['topic']}")
                if 'payload' in msg.details:
                    payload = msg.details['payload']
                    print(f"    Payload: {payload[:50]}{'...' if len(payload) > 50 else ''}")
                print(f"    QoS: {msg.details['qos']}, Retain: {msg.details['retain']}, DUP: {msg.details['dup']}")
        print()
        
        # 会话行为分析
        self._analyze_session_behavior()
    
    def _analyze_session_behavior(self):
        """分析会话行为"""
        print("🔍 会话行为分析:")
        
        # 计算消息速率
        if self.session_duration > 0:
            message_rate = self.stats['total_messages'] / self.session_duration
            print(f"  消息速率: {message_rate:.2f} 条/秒")
        
        # 计算平均消息大小
        publish_messages = [msg for msg in self.messages if msg.msg_type == "PUBLISH" and 'payload_length' in msg.details]
        if publish_messages:
            avg_size = sum(msg.details['payload_length'] for msg in publish_messages) / len(publish_messages)
            print(f"  平均消息大小: {avg_size:.2f} 字节")
        
        # 检查消息完整性
        publish_out = [msg for msg in self.messages if msg.msg_type == "PUBLISH" and msg.direction == "OUT"]
        puback_in = [msg for msg in self.messages if msg.msg_type == "PUBACK" and msg.direction == "IN"]
        print(f"  发送的 PUBLISH 消息: {len(publish_out)} 个")
        print(f"  收到的 PUBACK 消息: {len(puback_in)} 个")
        
        # 检查会话状态
        connected_events = [event for event in self.stats['session_events'] if event['event'] == 'CONNECTED']
        disconnected_events = [event for event in self.stats['session_events'] if event['event'] == 'DISCONNECTED']
        print(f"  连接事件: {len(connected_events)} 次")
        print(f"  断开连接事件: {len(disconnected_events)} 次")
        
        print()
    
    def save_analysis(self, output_file):
        """保存分析结果到文件
        
        Args:
            output_file: 输出文件路径
        """
        analysis_data = {
            "session_info": {
                "client_id": self.client_id,
                "broker": self.broker,
                "port": self.port,
                "topic": self.topic,
                "session_start": self.session_start,
                "session_end": self.session_end,
                "session_duration": self.session_duration,
                "formatted_session_start": datetime.fromtimestamp(self.session_start).strftime("%Y-%m-%d %H:%M:%S.%f")[:-3] if self.session_start else "N/A",
                "formatted_session_end": datetime.fromtimestamp(self.session_end).strftime("%Y-%m-%d %H:%M:%S.%f")[:-3] if self.session_end else "N/A"
            },
            "messages": [msg.to_dict() for msg in self.messages],
            "statistics": {
                "total_messages": self.stats['total_messages'],
                "message_types": dict(self.stats['message_types']),
                "message_directions": dict(self.stats['message_directions']),
                "topics": dict(self.stats['topics']),
                "qos_distribution": dict(self.stats['qos_distribution']),
                "retain_distribution": dict(self.stats['retain_distribution']),
                "dup_distribution": dict(self.stats['dup_distribution'])
            },
            "session_events": self.stats['session_events']
        }
        
        try:
            with open(output_file, 'w') as f:
                json.dump(analysis_data, f, indent=2, default=str)
            print(f"📥 分析结果已保存到: {output_file}")
            print(f"📋 保存内容包括: 会话信息、{len(self.messages)} 条消息、统计数据、会话事件")
        except Exception as e:
            print(f"❌ 保存结果失败: {e}")

def main():
    """主函数"""
    parser = argparse.ArgumentParser(description='MQTT 会话分析工具')
    parser.add_argument('--broker', type=str, default='test.mosquitto.org',
                        help='MQTT 服务器地址')
    parser.add_argument('--port', type=int, default=1883,
                        help='MQTT 端口号')
    parser.add_argument('--topic', type=str, default='iot/test/topic',
                        help='测试主题')
    parser.add_argument('--duration', type=int, default=15,
                        help='会话持续时间（秒）')
    parser.add_argument('--output', type=str, default='mqtt_session_analysis.json',
                        help='输出文件路径')
    
    args = parser.parse_args()
    
    # 创建并启动会话分析器
    analyzer = MQTTSessionAnalyzer(
        broker=args.broker,
        port=args.port,
        topic=args.topic
    )
    
    analyzer.run_session(duration=args.duration)
    analyzer.analyze_session()
    analyzer.save_analysis(args.output)

if __name__ == '__main__':
    main()
