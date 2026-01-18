#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
MQTT 自动化测试脚本

同时运行 MQTT 客户端和报文分析工具，分析客户端与服务器之间的交互报文
"""

import subprocess
import time
import os
import sys

# 配置参数
MQTT_BROKER = "test.mosquitto.org"
MQTT_PORT = 1883
MQTT_TOPIC = "iot/test/topic"
CAPTURE_DURATION = 30  # 捕获时长（秒）
ANALYZER_OUTPUT = "mqtt_analysis_result.json"

class MQTTTestAutomator:
    """MQTT 测试自动化类"""
    
    def __init__(self):
        """初始化测试自动化类"""
        self.client_process = None
        self.analyzer_process = None
    
    def start_client(self):
        """启动 MQTT 客户端"""
        print("🔄 启动 MQTT 客户端...")
        client_cmd = [
            sys.executable,
            "src/mqtt/client_test.py",
            "--broker", MQTT_BROKER,
            "--port", str(MQTT_PORT),
            "--topic", MQTT_TOPIC,
            "--duration", str(CAPTURE_DURATION + 5)  # 客户端运行时间比捕获时长多5秒
        ]
        
        self.client_process = subprocess.Popen(
            client_cmd,
            cwd=os.path.dirname(os.path.dirname(os.path.dirname(__file__))),
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        
        return self.client_process
    
    def start_analyzer(self):
        """启动 MQTT 报文分析器"""
        print("🔄 启动 MQTT 报文分析器...")
        analyzer_cmd = [
            sys.executable,
            "src/mqtt/packet_analyzer.py",
            "-H", MQTT_BROKER,
            "-p", str(MQTT_PORT),
            "-d", str(CAPTURE_DURATION),
            "-o", ANALYZER_OUTPUT
        ]
        
        self.analyzer_process = subprocess.Popen(
            analyzer_cmd,
            cwd=os.path.dirname(os.path.dirname(os.path.dirname(__file__))),
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        
        return self.analyzer_process
    
    def run_test(self):
        """运行完整的测试流程"""
        print("=== MQTT 客户端-服务器交互测试 ===")
        print(f"MQTT 服务器: {MQTT_BROKER}:{MQTT_PORT}")
        print(f"测试主题: {MQTT_TOPIC}")
        print(f"捕获时长: {CAPTURE_DURATION} 秒")
        print()
        
        try:
            # 1. 启动客户端
            client_proc = self.start_client()
            time.sleep(2)  # 等待客户端启动
            
            # 2. 启动报文分析器
            analyzer_proc = self.start_analyzer()
            
            # 3. 等待分析器完成
            print("\n⏱️  正在运行测试，捕获 MQTT 报文...")
            analyzer_proc.wait()
            
            # 4. 等待客户端完成
            if self.client_process:
                self.client_process.terminate()
                self.client_process.wait()
            
            # 5. 查看分析结果
            print("\n📊 测试完成，查看分析结果...")
            if os.path.exists(ANALYZER_OUTPUT):
                print(f"分析结果已保存到: {ANALYZER_OUTPUT}")
                print("\n=== 报文分析报告 ===")
                # 打印报告的前几行
                with open(ANALYZER_OUTPUT, 'r') as f:
                    import json
                    data = json.load(f)
                    print(f"总捕获报文数: {len(data)}")
                    
                    # 统计不同类型的报文
                    type_count = {}
                    for packet in data:
                        msg_type = packet.get('mqtt_type', 'UNKNOWN')
                        type_count[msg_type] = type_count.get(msg_type, 0) + 1
                    
                    print("\n报文类型分布:")
                    for msg_type, count in type_count.items():
                        print(f"  {msg_type}: {count} 个")
                    
                    # 展示前几个报文的摘要
                    print(f"\n前 3 个报文摘要:")
                    for i, packet in enumerate(data[:3]):
                        print(f"\n报文 {i+1}:")
                        print(f"  时间: {packet['formatted_time']}")
                        print(f"  方向: {packet['source']} → {packet['destination']}")
                        print(f"  类型: {packet.get('mqtt_type', 'UNKNOWN')}")
                        if packet.get('mqtt_msgtype') == 3:  # PUBLISH
                            print(f"  主题: {packet.get('topic', 'N/A')}")
                            print(f"  负载: {packet.get('payload', 'N/A')[:30]}{'...' if len(packet.get('payload', '')) > 30 else ''}")
            
        except KeyboardInterrupt:
            print("\n\n⚠️  测试被用户中断")
        except Exception as e:
            print(f"\n\n❌ 测试过程中发生错误: {e}")
        finally:
            # 清理资源
            if self.client_process:
                self.client_process.terminate()
                self.client_process.wait()
            if self.analyzer_process:
                self.analyzer_process.terminate()
                self.analyzer_process.wait()
            
            print("\n=== 测试结束 ===")

def main():
    """主函数"""
    automator = MQTTTestAutomator()
    automator.run_test()

if __name__ == "__main__":
    main()
