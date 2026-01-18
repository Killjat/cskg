#!/usr/bin/env python3
"""
Web服务管理器
用于通过Web界面控制各种服务的开启与关闭
"""

from flask import Flask, render_template, request, jsonify
import json
import subprocess
import os
import time
import threading

# 创建Flask应用
app = Flask(__name__)

# 配置文件路径
CONFIG_FILE = 'service_config.json'

# 服务状态存储
running_services = {}

class ServiceManager:
    """服务管理类"""
    def __init__(self):
        self.config = self.load_config()
    
    def load_config(self):
        """加载配置文件"""
        try:
            with open(CONFIG_FILE, 'r') as f:
                return json.load(f)
        except FileNotFoundError:
            return {"applications": {}}
    
    def get_services(self):
        """获取所有服务列表"""
        services = []
        for app_name, versions in self.config['applications'].items():
            for version, info in versions.items():
                service_key = f"{app_name}_{version}"
                status = "running" if service_key in running_services else "stopped"
                
                service_info = {
                    "app_name": app_name,
                    "version": version,
                    "name": info["name"],
                    "description": info["description"],
                    "status": status,
                    "key": service_key
                }
                
                if "port" in info:
                    service_info["port"] = info["port"]
                elif "ports" in info:
                    service_info["ports"] = info["ports"]
                
                services.append(service_info)
        
        return services
    
    def start_service(self, app_name, version):
        """启动服务"""
        # 检查应用和版本是否存在
        if app_name not in self.config['applications'] or version not in self.config['applications'][app_name]:
            return False, f"服务 {app_name} {version} 不存在"
        
        service_key = f"{app_name}_{version}"
        
        # 检查服务是否已在运行
        if service_key in running_services:
            return False, f"服务 {app_name} {version} 已在运行"
        
        # 获取服务信息
        service_info = self.config['applications'][app_name][version]
        
        try:
            # 启动服务
            process = subprocess.Popen(
                service_info['command'],
                shell=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                universal_newlines=True,
                cwd=os.getcwd()
            )
            
            # 记录服务信息
            running_services[service_key] = {
                'process': process,
                'info': service_info,
                'start_time': time.time(),
                'app_name': app_name,
                'version': version
            }
            
            # 等待1秒，检查是否有立即的错误
            time.sleep(1)
            if process.poll() is not None:
                # 进程已结束，读取错误信息
                stderr = process.stderr.read()
                del running_services[service_key]
                return False, f"服务启动失败: {stderr}"
            
            return True, f"服务 {app_name} {version} 启动成功"
            
        except Exception as e:
            return False, f"启动服务时出错: {e}"
    
    def stop_service(self, app_name, version):
        """停止服务"""
        service_key = f"{app_name}_{version}"
        
        # 检查服务是否在运行
        if service_key not in running_services:
            return False, f"服务 {app_name} {version} 未在运行"
        
        service = running_services[service_key]
        process = service['process']
        
        try:
            # 终止进程
            process.terminate()
            # 等待进程结束
            process.wait(timeout=5)
            del running_services[service_key]
            return True, f"服务 {app_name} {version} 已停止"
        except subprocess.TimeoutExpired:
            # 超时，强制杀死进程
            process.kill()
            del running_services[service_key]
            return True, f"服务 {app_name} {version} 已强制停止"
        except Exception as e:
            return False, f"停止服务时出错: {e}"

# 创建服务管理器实例
service_manager = ServiceManager()

# Web路由
@app.route('/')
def index():
    """主页"""
    services = service_manager.get_services()
    return render_template('index.html', services=services)

@app.route('/api/services')
def api_services():
    """获取服务列表API"""
    services = service_manager.get_services()
    return jsonify({"services": services})

@app.route('/api/start/<app_name>/<version>', methods=['POST'])
def api_start_service(app_name, version):
    """启动服务API"""
    success, message = service_manager.start_service(app_name, version)
    return jsonify({"success": success, "message": message})

@app.route('/api/stop/<app_name>/<version>', methods=['POST'])
def api_stop_service(app_name, version):
    """停止服务API"""
    success, message = service_manager.stop_service(app_name, version)
    return jsonify({"success": success, "message": message})

# 创建templates目录
if not os.path.exists('templates'):
    os.makedirs('templates')

# 创建HTML模板
index_html = '''
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>服务管理控制台</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background-color: #f5f5f5;
            color: #333;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        
        h1 {
            text-align: center;
            color: #2c3e50;
            margin-bottom: 30px;
            padding: 20px;
            background-color: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        
        .service-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
            gap: 20px;
        }
        
        .service-card {
            background-color: white;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            padding: 20px;
            transition: transform 0.2s ease, box-shadow 0.2s ease;
        }
        
        .service-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        }
        
        .service-header {
            margin-bottom: 15px;
            padding-bottom: 15px;
            border-bottom: 1px solid #eee;
        }
        
        .service-name {
            font-size: 18px;
            font-weight: 600;
            color: #2c3e50;
            margin-bottom: 5px;
        }
        
        .service-version {
            font-size: 14px;
            color: #7f8c8d;
        }
        
        .service-status {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 500;
            margin-top: 8px;
        }
        
        .status-running {
            background-color: #d4edda;
            color: #155724;
        }
        
        .status-stopped {
            background-color: #f8d7da;
            color: #721c24;
        }
        
        .service-info {
            margin-bottom: 15px;
        }
        
        .service-description {
            font-size: 14px;
            color: #666;
            margin-bottom: 10px;
        }
        
        .service-port {
            font-size: 13px;
            color: #95a5a6;
        }
        
        .service-actions {
            display: flex;
            gap: 10px;
        }
        
        .btn {
            flex: 1;
            padding: 10px;
            border: none;
            border-radius: 4px;
            font-size: 14px;
            font-weight: 500;
            cursor: pointer;
            transition: background-color 0.2s ease;
        }
        
        .btn-start {
            background-color: #27ae60;
            color: white;
        }
        
        .btn-start:hover {
            background-color: #229954;
        }
        
        .btn-stop {
            background-color: #e74c3c;
            color: white;
        }
        
        .btn-stop:hover {
            background-color: #c0392b;
        }
        
        .btn:disabled {
            background-color: #bdc3c7;
            cursor: not-allowed;
        }
        
        .message {
            margin-top: 15px;
            padding: 10px;
            border-radius: 4px;
            font-size: 14px;
        }
        
        .message-success {
            background-color: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        
        .message-error {
            background-color: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
        
        .refresh-btn {
            display: block;
            margin: 0 auto 20px;
            padding: 10px 20px;
            background-color: #3498db;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
        }
        
        .refresh-btn:hover {
            background-color: #2980b9;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>服务管理控制台</h1>
        <button class="refresh-btn" onclick="refreshServices()">刷新服务列表</button>
        
        <div class="service-grid" id="serviceGrid">
            <!-- 服务卡片将通过JavaScript动态生成 -->
        </div>
    </div>
    
    <script>
        // 页面加载时获取服务列表
        document.addEventListener('DOMContentLoaded', function() {
            loadServices();
        });
        
        // 加载服务列表
        function loadServices() {
            fetch('/api/services')
                .then(response => response.json())
                .then(data => {
                    renderServices(data.services);
                })
                .catch(error => {
                    console.error('Error loading services:', error);
                });
        }
        
        // 渲染服务卡片
        function renderServices(services) {
            const grid = document.getElementById('serviceGrid');
            grid.innerHTML = '';
            
            services.forEach(service => {
                const card = createServiceCard(service);
                grid.appendChild(card);
            });
        }
        
        // 创建服务卡片
        function createServiceCard(service) {
            const card = document.createElement('div');
            card.className = 'service-card';
            
            const statusClass = service.status === 'running' ? 'status-running' : 'status-stopped';
            const statusText = service.status === 'running' ? '运行中' : '已停止';
            
            let portHtml = '';
            if (service.port) {
                portHtml = `<div class="service-port">端口: ${service.port}</div>`;
            } else if (service.ports) {
                portHtml = `<div class="service-port">端口: ${service.ports.join(', ')}</div>`;
            }
            
            let actionsHtml = '';
            if (service.status === 'stopped') {
                actionsHtml = `
                    <div class="service-actions">
                        <button class="btn btn-start" onclick="startService('${service.app_name}', '${service.version}', this)">启动服务</button>
                    </div>
                `;
            } else {
                actionsHtml = `
                    <div class="service-actions">
                        <button class="btn btn-stop" onclick="stopService('${service.app_name}', '${service.version}', this)">停止服务</button>
                    </div>
                `;
            }
            
            card.innerHTML = `
                <div class="service-header">
                    <div class="service-name">${service.name}</div>
                    <div class="service-version">版本: ${service.version}</div>
                    <span class="service-status ${statusClass}">${statusText}</span>
                </div>
                
                <div class="service-info">
                    <div class="service-description">${service.description}</div>
                    ${portHtml}
                </div>
                
                ${actionsHtml}
                
                <div id="message-${service.key}"></div>
            `;
            
            return card;
        }
        
        // 启动服务
        function startService(appName, version, button) {
            const serviceKey = `${appName}_${version}`;
            const messageDiv = document.getElementById(`message-${serviceKey}`);
            
            // 禁用按钮
            button.disabled = true;
            messageDiv.innerHTML = '';
            
            fetch(`/api/start/${appName}/${version}`, {
                method: 'POST'
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    messageDiv.className = 'message message-success';
                    messageDiv.textContent = data.message;
                    // 刷新服务列表
                    setTimeout(() => {
                        loadServices();
                    }, 500);
                } else {
                    messageDiv.className = 'message message-error';
                    messageDiv.textContent = data.message;
                    button.disabled = false;
                }
            })
            .catch(error => {
                messageDiv.className = 'message message-error';
                messageDiv.textContent = '启动服务时出错';
                button.disabled = false;
            });
        }
        
        // 停止服务
        function stopService(appName, version, button) {
            const serviceKey = `${appName}_${version}`;
            const messageDiv = document.getElementById(`message-${serviceKey}`);
            
            // 禁用按钮
            button.disabled = true;
            messageDiv.innerHTML = '';
            
            fetch(`/api/stop/${appName}/${version}`, {
                method: 'POST'
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    messageDiv.className = 'message message-success';
                    messageDiv.textContent = data.message;
                    // 刷新服务列表
                    setTimeout(() => {
                        loadServices();
                    }, 500);
                } else {
                    messageDiv.className = 'message message-error';
                    messageDiv.textContent = data.message;
                    button.disabled = false;
                }
            })
            .catch(error => {
                messageDiv.className = 'message message-error';
                messageDiv.textContent = '停止服务时出错';
                button.disabled = false;
            });
        }
        
        // 刷新服务列表
        function refreshServices() {
            loadServices();
        }
    </script>
</body>
</html>
'''

# 写入HTML模板文件
with open('templates/index.html', 'w') as f:
    f.write(index_html)

# 主入口
if __name__ == '__main__':
    # 获取公网IP
    import subprocess
    try:
        public_ip = subprocess.check_output(['curl', '-s', 'icanhazip.com']).decode('utf-8').strip()
    except Exception as e:
        public_ip = "无法获取公网IP"
    
    print("=== Web服务管理器 ===")
    print("服务启动成功!")
    print(f"公网访问地址: http://{public_ip}:9999")
    print("内网访问地址: http://localhost:9999")
    print("按 Ctrl+C 停止服务")
    print()
    
    # 启动Flask应用
    app.run(host='0.0.0.0', port=9999, debug=False)
