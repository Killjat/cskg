#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
MQTT 客户端实现
"""

import paho.mqtt.client as mqtt
import time
import logging

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class MQTTClient:
    """MQTT 客户端类"""
    
    def __init__(self, client_id, broker="localhost", port=1883, username=None, password=None):
        """
        初始化 MQTT 客户端
        
        Args:
            client_id: 客户端 ID
            broker: MQTT 代理地址
            port: MQTT 端口
            username: 用户名
            password: 密码
        """
        self.client = mqtt.Client(client_id=client_id)
        self.broker = broker
        self.port = port
        
        # 设置认证信息
        if username and password:
            self.client.username_pw_set(username, password)
        
        # 注册回调函数
        self.client.on_connect = self._on_connect
        self.client.on_message = self._on_message
        self.client.on_publish = self._on_publish
        self.client.on_disconnect = self._on_disconnect
    
    def _on_connect(self, client, userdata, flags, rc):
        """连接回调"""
        if rc == 0:
            logger.info(f"Connected to MQTT broker at {self.broker}:{self.port}")
        else:
            logger.error(f"Connection failed with code {rc}")
    
    def _on_message(self, client, userdata, msg):
        """消息接收回调"""
        logger.info(f"Received message: {msg.payload.decode()} on topic {msg.topic}")
    
    def _on_publish(self, client, userdata, mid):
        """发布回调"""
        logger.info(f"Message published with mid {mid}")
    
    def _on_disconnect(self, client, userdata, rc):
        """断开连接回调"""
        logger.info(f"Disconnected with code {rc}")
    
    def connect(self):
        """连接到 MQTT 代理"""
        self.client.connect(self.broker, self.port, keepalive=60)
    
    def subscribe(self, topic, qos=0):
        """订阅主题"""
        self.client.subscribe(topic, qos)
    
    def publish(self, topic, payload, qos=0, retain=False):
        """发布消息"""
        self.client.publish(topic, payload, qos, retain)
    
    def start_loop(self):
        """启动消息循环"""
        self.client.loop_start()
    
    def stop_loop(self):
        """停止消息循环"""
        self.client.loop_stop()
    
    def disconnect(self):
        """断开连接"""
        self.client.disconnect()
