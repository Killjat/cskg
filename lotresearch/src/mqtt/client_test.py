#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
MQTT 客户端测试脚本

使用公共MQTT服务器测试客户端连接、订阅和发布功能
"""

import paho.mqtt.client as mqtt
import time
import argparse

# 解析命令行参数
parser = argparse.ArgumentParser(description='MQTT 客户端测试脚本')
parser.add_argument('--broker', type=str, default='test.mosquitto.org',
                    help='MQTT 服务器地址')
parser.add_argument('--port', type=int, default=1883,
                    help='MQTT 服务器端口')
parser.add_argument('--topic', type=str, default='iot/test/topic',
                    help='MQTT 测试主题')
parser.add_argument('--duration', type=int, default=15,
                    help='客户端运行时长（秒）')

args = parser.parse_args()

# 公共MQTT服务器配置
MQTT_BROKER = args.broker
MQTT_PORT = args.port
MQTT_TOPIC = args.topic
CLIENT_DURATION = args.duration

# 回调函数定义

def on_connect(client, userdata, flags, rc):
    """连接回调"""
    if rc == 0:
        print(f"✅ 成功连接到MQTT服务器: {MQTT_BROKER}:{MQTT_PORT}")
        # 连接成功后订阅主题
        client.subscribe(MQTT_TOPIC)
        print(f"✅ 已订阅主题: {MQTT_TOPIC}")
    else:
        print(f"❌ 连接失败，错误代码: {rc}")

def on_message(client, userdata, msg):
    """消息接收回调"""
    print(f"📩 收到消息: 主题={msg.topic}, 内容={msg.payload.decode()}")

def on_publish(client, userdata, mid):
    """发布回调"""
    print(f"📤 消息发布成功，消息ID: {mid}")

def main():
    """主函数"""
    print("=== MQTT 客户端测试 ===")
    print(f"连接到公共MQTT服务器: {MQTT_BROKER}:{MQTT_PORT}")
    
    # 创建MQTT客户端实例
    client = mqtt.Client()
    
    # 注册回调函数
    client.on_connect = on_connect
    client.on_message = on_message
    client.on_publish = on_publish
    
    try:
        # 连接到MQTT服务器
        client.connect(MQTT_BROKER, MQTT_PORT, keepalive=60)
        
        # 启动消息循环
        client.loop_start()
        
        # 等待1秒确保连接成功
        time.sleep(1)
        
        # 记录开始时间
        start_time = time.time()
        
        # 发布测试消息
        test_message = "Hello, MQTT! This is a test message from our client."
        client.publish(MQTT_TOPIC, test_message)
        
        # 定期发布消息
        message_count = 1
        while time.time() - start_time < CLIENT_DURATION:
            # 每3秒发布一条消息
            if int(time.time() - start_time) % 3 == 0:
                message_count += 1
                test_message = f"Test message #{message_count} at {time.strftime('%H:%M:%S')}"
                client.publish(MQTT_TOPIC, test_message)
            # 短暂休眠，减少CPU占用
            time.sleep(0.5)
        
    except KeyboardInterrupt:
        print("\n用户中断程序")
    except Exception as e:
        print(f"❌ 发生错误: {e}")
    finally:
        # 停止消息循环并断开连接
        client.loop_stop()
        client.disconnect()
        print("\n✅ 已断开与MQTT服务器的连接")

if __name__ == "__main__":
    main()
