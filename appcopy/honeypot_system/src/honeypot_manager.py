#!/usr/bin/env python3
"""
蜜罐系统管理器
负责运行各种协议服务器，处理设备指纹识别和流量分析
"""

import os
import sys
import json
import time
import threading
import subprocess
import socket
from datetime import datetime

class HoneypotManager:
    """蜜罐系统管理器类"""
    
    def __init__(self):
        self.config_path = os.path.join(os.path.dirname(__file__), '../config/config.json')
        self.log_path = os.path.join(os.path.dirname(__file__), '../logs')
        self.running_services = {}
        self.connections = {}
        self.device_fingerprints = {}
        
        # 加载配置
        self.config = self.load_config()
        
        # 创建日志目录
        os.makedirs(self.log_path, exist_ok=True)
        
    def load_config(self):
        """加载配置文件"""
        default_config = {
            "services": {
                "modbus": {
                    "enabled": True,
                    "port": 502,
                    "server_script": "../../modbus_server.py"
                },
                "mysql": {
                    "enabled": True,
                    "port": 3306,
                    "server_script": "../../mysql_server.py"
                },
                "redis": {
                    "enabled": True,
                    "port": 6379,
                    "server_script": "../../redis_server.py"
                },
                "kafka": {
                    "enabled": True,
                    "port": 9092,
                    "server_script": "../../kafka_server.py"
                }
            },
            "packetbeat": {
                "enabled": True,
                "config_path": "../config/packetbeat.yml"
            },
            "web": {
                "enabled": True,
                "port": 10000,
                "host": "0.0.0.0"
            }
        }
        
        if os.path.exists(self.config_path):
            try:
                with open(self.config_path, 'r') as f:
                    return json.load(f)
            except Exception as e:
                print(f"加载配置文件失败: {e}")
                return default_config
        else:
            # 保存默认配置
            with open(self.config_path, 'w') as f:
                json.dump(default_config, f, indent=2)
            return default_config
    
    def start_service(self, service_name, config):
        """启动服务"""
        if not config['enabled']:
            return
        
        print(f"启动服务: {service_name} (端口: {config['port']})")
        
        try:
            # 检查脚本是否存在
            script_path = os.path.join(os.path.dirname(__file__), config['server_script'])
            if not os.path.exists(script_path):
                print(f"警告: {service_name} 脚本不存在: {script_path}")
                return
            
            # 启动服务进程
            process = subprocess.Popen(
                [sys.executable, script_path],
                shell=False,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                universal_newlines=True
            )
            
            # 记录服务信息
            self.running_services[service_name] = {
                'process': process,
                'config': config,
                'start_time': time.time()
            }
            
        except Exception as e:
            print(f"启动 {service_name} 服务失败: {e}")
    
    def stop_service(self, service_name):
        """停止服务"""
        if service_name not in self.running_services:
            return
        
        print(f"停止服务: {service_name}")
        
        try:
            process = self.running_services[service_name]['process']
            process.terminate()
            process.wait(timeout=5)
            del self.running_services[service_name]
        except Exception as e:
            print(f"停止 {service_name} 服务失败: {e}")
    
    def start_packetbeat(self):
        """启动Packetbeat"""
        if not self.config['packetbeat']['enabled']:
            return
        
        print("启动Packetbeat...")
        
        try:
            # 检查Packetbeat是否已安装
            result = subprocess.run(['which', 'packetbeat'], capture_output=True, text=True)
            if result.returncode != 0:
                print("警告: Packetbeat未安装，跳过启动")
                return
            
            packetbeat_path = result.stdout.strip()
            config_path = os.path.join(os.path.dirname(__file__), self.config['packetbeat']['config_path'])
            
            # 启动Packetbeat
            process = subprocess.Popen(
                [packetbeat_path, '-c', config_path],
                shell=False,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                universal_newlines=True
            )
            
            self.running_services['packetbeat'] = {
                'process': process,
                'start_time': time.time()
            }
            
        except Exception as e:
            print(f"启动Packetbeat失败: {e}")
    
    def start_web_server(self):
        """启动Web服务器"""
        if not self.config['web']['enabled']:
            return
        
        print(f"启动Web服务器: http://{self.config['web']['host']}:{self.config['web']['port']}")
        
        try:
            # 启动Web服务器进程
            web_script = os.path.join(os.path.dirname(__file__), '../web/app.py')
            process = subprocess.Popen(
                [sys.executable, web_script, '--host', self.config['web']['host'], '--port', str(self.config['web']['port'])],
                shell=False,
