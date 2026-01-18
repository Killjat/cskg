#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
MQTT 客户端和服务端示例

该示例展示了如何使用 MQTTClient 和 MQTTServer 类
"""

import time
import threading
from mqtt.client import MQTTClient
from mqtt.server import MQTTServer

def run_server():
    """运行 MQTT 服务端"""
    server = MQTTServer(host="0.0.0.0", port=1883)
    server.start()
    return server

def run_publisher():
    """运行 MQTT 发布者"""
    client = MQTTClient(client_id="publisher-1", broker="localhost", port=1883)
    client.connect()
    client.start_loop()
    
    # 发布消息
    for i in range(5):
        topic = "iot/devices/data"
        payload = f"{{\"device_id\": \"device-{i}\", \"temperature\": {20 + i}, \"humidity\": {60 + i}}}"
        client.publish(topic, payload, qos=0)
        print(f"发布消息: {topic} -> {payload}")
        time.sleep(2)
    
    client.stop_loop()
    client.disconnect()

def run_subscriber():
    """运行 MQTT 订阅者"""
    client = MQTTClient(client_id="subscriber-1", broker="localhost", port=1883)
    client.connect()
    client.subscribe("iot/devices/#", qos=0)
    client.start_loop()
    
    # 运行 15 秒后停止
    time.sleep(15)
    
    client.stop_loop()
    client.disconnect()

def main():
    """主函数"""
    print("=== MQTT 客户端和服务端示例 ===")
    
    # 1. 启动 MQTT 服务端
    print("1. 启动 MQTT 服务端...")
    server = run_server()
    time.sleep(2)
    
    # 2. 启动订阅者线程
    print("2. 启动 MQTT 订阅者...")
    subscriber_thread = threading.Thread(target=run_subscriber)
    subscriber_thread.daemon = True
    subscriber_thread.start()
    
    # 3. 启动发布者
    print("3. 启动 MQTT 发布者...")
    publisher_thread = threading.Thread(target=run_publisher)
    publisher_thread.daemon = True
    publisher_thread.start()
    
    # 等待所有线程完成
    publisher_thread.join()
    time.sleep(5)  # 给订阅者一些时间接收消息
    
    # 4. 停止服务端
    print("4. 停止 MQTT 服务端...")
    server.stop()
    
    print("=== 示例完成 ===")

if __name__ == "__main__":
    main()
