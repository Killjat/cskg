#!/usr/bin/env python3
"""
简易工业协议蜜罐系统
具有Web界面，支持Modbus、MySQL、Redis、Kafka协议
"""

import os
import sys
import json
import time
import threading
import socket
from datetime import datetime
from flask import Flask, render_template, jsonify

# 创建Flask应用
app = Flask(__name__)
app.config['SECRET_KEY'] = 'honeypot-secret-key'

# 全局变量
connections = []
device_fingerprints = {}
running = True

# 服务配置
services = {
    "modbus": {"port": 502, "name": "Modbus TCP", "device_model": "Siemens S7-1200"},
    "mysql": {"port": 3306, "name": "MySQL", "device_model": "MySQL Client"},
    "redis": {"port": 6379, "name": "Redis", "device_model": "Redis Client"},
    "kafka": {"port": 9092, "name": "Kafka", "device_model": "Kafka Client"}
}

class DeviceFingerprint:
    """设备指纹类"""
    def __init__(self, client_ip, client_port, server_port, protocol):
        self.client_ip = client_ip
        self.client_port = client_port
        self.server_port = server_port
        self.protocol = protocol
        self.first_seen = datetime.now()
        self.last_seen = datetime.now()
        self.connection_count = 1
        self.device_info = self.identify_device()
    
    def identify_device(self):
        """识别设备信息"""
        service = services.get(self.protocol, {})
        return {
            "os": "Unknown",
            "device_type": "Industrial Device",
            "manufacturer":