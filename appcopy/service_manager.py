#!/usr/bin/env python3
"""
服务管理程序
根据用户输入的应用和版本，自动启动对应的服务端
"""

import json
import subprocess
import sys
import os
import time
from datetime import datetime

# 配置文件路径
CONFIG_FILE = 'service_config.json'

class ServiceManager:
    def __init__(self):
        self.config = self.load_config()
        self.running_services = {}
    
    def load_config(self):
        """加载配置文件"""
        try:
            with open(CONFIG_FILE, 'r') as f:
                return json.load(f)
        except FileNotFoundError:
            print(f"[-] 配置文件 {CONFIG_FILE} 不存在")
            sys.exit(1)
        except json.JSONDecodeError:
            print(f"[-] 配置文件 {CONFIG_FILE} 格式错误")
            sys.exit(1)
    
    def list_applications(self):
        """列出所有可用的应用和版本"""
        print("\n=== 可用应用列表 ===")
        for app_name, versions in self.config['applications'].items():
            print(f"\n{app_name}:")
            for version, info in versions.items():
                print(f"  {version}: {info['name']}")
                print(f"    描述: {info['description']}")
                if 'port' in info:
                    print(f"    端口: {info['port']}")
                elif 'ports' in info:
                    print(f"    端口: {', '.join(map(str, info['ports']))}")
    
    def start_service(self, app_name, version):
        """启动指定应用和版本的服务"""
        # 检查应用是否存在
        if app_name not in self.config['applications']:
            print(f"[-] 应用 '{app_name}' 不存在")
            self.list_applications()
            return False
        
        # 检查版本是否存在
        if version not in self.config['applications'][app_name]:
            print(f"[-] 应用 '{app_name}' 的版本 '{version}' 不存在")
            print(f"    可用版本: {', '.join(self.config['applications'][app_name].keys())}")
            return False
        
        # 获取服务信息
        service_info = self.config['applications'][app_name][version]
        
        # 检查服务是否已在运行
        service_key = f"{app_name}_{version}"
        if service_key in self.running_services:
            print(f"[-] 服务 {app_name} v{version} 已在运行")
            return False
        
        print(f"\n[+] 启动服务: {service_info['name']}")
        print(f"  描述: {service_info['description']}")
        if 'port' in service_info:
            print(f"  端口: {service_info['port']}")
        elif 'ports' in service_info:
            print(f"  端口: {', '.join(map(str, service_info['ports']))}")
        
        # 启动服务
        try:
            # 使用subprocess启动服务，设置为后台运行
            process = subprocess.Popen(
                service_info['command'],
                shell=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                universal_newlines=True,
                cwd=os.getcwd()
            )
            
            # 记录服务信息
            self.running_services[service_key] = {
                'process': process,
                'info': service_info,
                'start_time': datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
                'app_name': app_name,
                'version': version
            }
            
            print(f"  命令: {service_info['command']}")
            print(f"  进程ID: {process.pid}")
            print(f"  启动时间: {self.running_services[service_key]['start_time']}")
            print(f"[+] 服务 {app_name} v{version} 启动成功")
            
            # 等待1秒，检查是否有立即的错误
            time.sleep(1)
            if process.poll() is not None:
                # 进程已结束，读取错误信息
                stderr = process.stderr.read()
                print(f"[-] 服务启动失败: {stderr}")
                del self.running_services[service_key]
                return False
            
            return True
            
        except Exception as e:
            print(f"[-] 启动服务时出错: {e}")
            return False
    
    def stop_service(self, app_name, version):
        """停止指定应用和版本的服务"""
        service_key = f"{app_name}_{version}"
        if service_key not in self.running_services:
            print(f"[-] 服务 {app_name} v{version} 未在运行")
            return False
        
        service = self.running_services[service_key]
        process = service['process']
        
        print(f"[+] 停止服务: {service['info']['name']}")
        print(f"  进程ID: {process.pid}")
        
        try:
            # 终止进程
            process.terminate()
            # 等待进程结束
            process.wait(timeout=5)
            del self.running_services[service_key]
            print(f"[+] 服务 {app_name} v{version} 已停止")
            return True
        except subprocess.TimeoutExpired:
            # 超时，强制杀死进程
            process.kill()
            del self.running_services[service_key]
            print(f"[+] 服务 {app_name} v{version} 已强制停止")
            return True
        except Exception as e:
            print(f"[-] 停止服务时出错: {e}")
            return False
    
    def list_running_services(self):
        """列出所有正在运行的服务"""
        if not self.running_services:
            print("\n[+] 没有正在运行的服务")
            return
        
        print("\n=== 正在运行的服务 ===")
        for service_key, service in self.running_services.items():
            print(f"\n{service['app_name']} v{service['version']}:")
            print(f"  名称: {service['info']['name']}")
            print(f"  进程ID: {service['process'].pid}")
            print(f"  启动时间: {service['start_time']}")
            if 'port' in service['info']:
                print(f"  端口: {service['info']['port']}")
            elif 'ports' in service['info']:
                print(f"  端口: {', '.join(map(str, service['info']['ports']))}")
    
    def run(self):
        """主运行循环"""
        print("=== 服务管理程序 ===")
        print("\n可用命令:")
        print("  list      - 列出所有可用应用")
        print("  running   - 列出正在运行的服务")
        print("  start <app> <version> - 启动指定服务")
        print("  stop <app> <version>  - 停止指定服务")
        print("  exit      - 退出程序")
        
        while True:
            print("\n> ", end="")
            try:
                # 获取用户输入
                command = input().strip().split()
                if not command:
                    continue
                
                cmd = command[0].lower()
                
                if cmd == 'exit':
                    print("\n[+] 退出服务管理程序")
                    # 停止所有运行中的服务
                    for service_key in list(self.running_services.keys()):
                        service = self.running_services[service_key]
                        self.stop_service(service['app_name'], service['version'])
                    break
                
                elif cmd == 'list':
                    self.list_applications()
                
                elif cmd == 'running':
                    self.list_running_services()
                
                elif cmd == 'start':
                    if len(command) < 3:
                        print("[-] 请提供应用名称和版本")
                        print("    示例: start modbus 1.0")
                    else:
                        app_name = command[1]
                        version = command[2]
                        self.start_service(app_name, version)
                
                elif cmd == 'stop':
                    if len(command) < 3:
                        print("[-] 请提供应用名称和版本")
                        print("    示例: stop modbus 1.0")
                    else:
                        app_name = command[1]
                        version = command[2]
                        self.stop_service(app_name, version)
                
                else:
                    print(f"[-] 未知命令: {cmd}")
                    print("    可用命令: list, running, start, stop, exit")
                    
            except KeyboardInterrupt:
                print("\n\n[+] 退出服务管理程序")
                # 停止所有运行中的服务
                for service_key in list(self.running_services.keys()):
                    service = self.running_services[service_key]
                    self.stop_service(service['app_name'], service['version'])
                break
            except EOFError:
                print("\n[+] 退出服务管理程序")
                # 停止所有运行中的服务
                for service_key in list(self.running_services.keys()):
                    service = self.running_services[service_key]
                    self.stop_service(service['app_name'], service['version'])
                break

if __name__ == "__main__":
    manager = ServiceManager()
    manager.run()
