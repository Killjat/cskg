#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
网页克隆工具 - 高可用版本
使用FastAPI + Uvicorn实现，支持epoll/kqueue高效I/O多路复用
"""

import os
import json
import requests
from datetime import datetime
from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import HTMLResponse, JSONResponse, FileResponse
from fastapi.staticfiles import StaticFiles
from fastapi.middleware.cors import CORSMiddleware
from bs4 import BeautifulSoup

# 创建FastAPI应用
app = FastAPI(
    title="网页克隆工具",
    description="高可用网页克隆服务，支持URL克隆和内容提取",
    version="1.0.0"
)

# 配置CORS，允许所有跨域请求
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # 允许所有来源
    allow_credentials=True,
    allow_methods=["*"],  # 允许所有HTTP方法
    allow_headers=["*"],  # 允许所有HTTP头
)

# 全局配置
CLONED_DIR = os.path.join(os.path.dirname(__file__), 'cloned_pages')
os.makedirs(CLONED_DIR, exist_ok=True)

# 初始化服务器配置
# 获取本地/公网IP地址
LOCAL_IP = None
try:
    import os
    import socket
    
    print("🔍 尝试获取公网IP...")
    # 对所有系统，尝试通过curl命令获取公网IP
    import subprocess
    try:
        # 使用curl获取公网IP
        result = subprocess.run(
            ['curl', '-s', 'icanhazip.com'],
            capture_output=True,
            text=True,
            timeout=10
        )
        if result.returncode == 0:
            public_ip = result.stdout.strip()
            # 验证获取到的是有效的IP地址
            socket.inet_aton(public_ip)  # 验证IP格式
            LOCAL_IP = public_ip
            print(f"✅ 成功获取公网IP: {LOCAL_IP}")
    except Exception as e:
        print(f"⚠️ 获取公网IP失败，尝试获取局域网IP: {str(e)}")
    
    # 如果获取公网IP失败，尝试获取局域网IP
    if not LOCAL_IP:
        print("🔍 尝试获取局域网IP...")
        # 尝试通过网络接口获取IP
        try:
            # 导入必要的模块
            import fcntl
            import struct
            
            # 获取所有网络接口
            def get_ip_address(ifname):
                s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
                return socket.inet_ntoa(fcntl.ioctl(
                    s.fileno(),
                    0x8915,  # SIOCGIFADDR
                    struct.pack('256s', ifname[:15].encode('utf-8'))
                )[20:24])
            
            # 尝试获取常见物理网络接口的IP
            interfaces = ['eth0', 'en0', 'en1', 'wlan0', 'wifi0']
            for iface in interfaces:
                try:
                    ip = get_ip_address(iface)
                    if ip and ip != '127.0.0.1':
                        LOCAL_IP = ip
                        print(f"✅ 成功获取局域网IP: {LOCAL_IP}")
                        break
                except Exception:
                    continue
        except Exception:
            pass
        
        # 如果没有找到物理网络接口，使用传统方法
        if not LOCAL_IP:
            s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            s.connect(("8.8.8.8", 80))
            LOCAL_IP = s.getsockname()[0]
            s.close()
            print(f"✅ 使用传统方法获取IP: {LOCAL_IP}")
except Exception as e:
    LOCAL_IP = "127.0.0.1"
    print(f"⚠️ 初始化IP失败: {str(e)}")

# 使用固定端口8080
SERVER_PORT = 8080

# 挂载静态文件目录
app.mount("/cloned", StaticFiles(directory=CLONED_DIR), name="cloned")


def clone_web_page(url: str):
    """克隆网页内容"""
    try:
        print(f"正在访问 URL: {url}")
        headers = {
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36'
        }
        response = requests.get(url, headers=headers, timeout=30)
        response.raise_for_status()

        html = response.text
        soup = BeautifulSoup(html, 'html.parser')

        # 提取关键信息
        title = soup.title.string.strip() if soup.title and soup.title.string else '无标题'
        header = str(soup.find('header')) if soup.find('header') else ''
        body_content = str(soup.body) if soup.body else ''
        head_content = str(soup.head) if soup.head else ''

        print(f"成功获取网页: {title}")

        # 创建保存目录
        safe_url = url.replace('://', '_').replace('/', '_').replace(':', '_').replace('?', '_').replace('&', '_')
        save_dir = os.path.join(CLONED_DIR, safe_url)
        os.makedirs(save_dir, exist_ok=True)

        # 保存完整HTML
        full_html_path = os.path.join(save_dir, 'full.html')
        with open(full_html_path, 'w', encoding='utf-8') as f:
            f.write(html)
        print(f"完整HTML已保存到: {full_html_path}")

        # 保存提取的信息
        extracted_info = {
            'url': url,
            'title': title,
            'timestamp': datetime.now().isoformat(),
            'head': head_content,
            'header': header,
            'body': body_content
        }

        info_path = os.path.join(save_dir, 'info.json')
        with open(info_path, 'w', encoding='utf-8') as f:
            json.dump(extracted_info, f, ensure_ascii=False, indent=2)
        print(f"提取的信息已保存到: {info_path}")

        # 保存简化版HTML
        simple_html = f'''<!DOCTYPE html>
<html>
<head>
  <title>{title}</title>
  {head_content}
</head>
<body>
  {header}
  {body_content}
</body>
</html>'''

        simple_path = os.path.join(save_dir, 'simple.html')
        with open(simple_path, 'w', encoding='utf-8') as f:
            f.write(simple_html)
        print(f"简化HTML已保存到: {simple_path}")

        # 返回克隆结果，不包含URL（URL将在API层生成）
        return {
            'success': True,
            'title': title,
            'save_dir': save_dir,
            'safe_url': safe_url
        }
    except Exception as e:
        print(f"克隆失败: {str(e)}")
        return {
            'success': False,
            'error': str(e)
        }


@app.get("/", response_class=HTMLResponse)
def index():
    """主页 - 提供克隆功能界面"""
    return '''
    <!DOCTYPE html>
    <html lang="zh-CN">
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <title>网页克隆工具</title>
      <style>
        body {
          font-family: Arial, sans-serif;
          max-width: 800px;
          margin: 0 auto;
          padding: 20px;
          background-color: #f5f5f5;
        }
        h1 {
          color: #333;
          text-align: center;
        }
        .container {
          background-color: white;
          padding: 20px;
          border-radius: 8px;
          box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
        }
        .form-group {
          margin-bottom: 20px;
        }
        label {
          display: block;
          margin-bottom: 8px;
          font-weight: bold;
        }
        input[type="url"] {
          width: 100%;
          padding: 10px;
          font-size: 16px;
          border: 1px solid #ddd;
          border-radius: 4px;
        }
        button {
          background-color: #4CAF50;
          color: white;
          padding: 10px 20px;
          border: none;
          border-radius: 4px;
          cursor: pointer;
          font-size: 16px;
        }
        button:hover {
          background-color: #45a049;
        }
        .result {
          margin-top: 20px;
          padding: 15px;
          border-radius: 4px;
        }
        .success {
          background-color: #d4edda;
          color: #155724;
          border: 1px solid #c3e6cb;
        }
        .error {
          background-color: #f8d7da;
          color: #721c24;
          border: 1px solid #f5c6cb;
        }
        .cloned-list {
          margin-top: 30px;
        }
        .cloned-item {
          margin-bottom: 15px;
          padding: 15px;
          background-color: #e9ecef;
          border-radius: 4px;
        }
        .cloned-item h3 {
          margin: 0 0 10px 0;
        }
        .cloned-item .links {
          margin-top: 10px;
        }
        .cloned-item a {
          margin-right: 15px;
          color: #007bff;
          text-decoration: none;
        }
        .cloned-item a:hover {
          text-decoration: underline;
        }
      </style>
    </head>
    <body>
      <div class="container">
        <h1>网页克隆工具</h1>
        <div class="form-group">
          <label for="url">输入要克隆的URL：</label>
          <input type="url" id="url" placeholder="https://example.com" required>
        </div>
        <button onclick="clonePage()">克隆网页</button>
        <div id="result" class="result" style="display: none;"></div>
        
        <div class="cloned-list">
          <h2>已克隆的页面</h2>
          <div id="clonedPages"></div>
        </div>
      </div>
      
      <script>
        // 克隆页面功能
        async function clonePage() {
          const url = document.getElementById('url').value;
          const resultDiv = document.getElementById('result');
          
          if (!url) {
            resultDiv.className = 'result error';
            resultDiv.innerHTML = '请输入有效的URL';
            resultDiv.style.display = 'block';
            return;
          }
          
          resultDiv.className = 'result success';
          resultDiv.innerHTML = '正在克隆...';
          resultDiv.style.display = 'block';
          
          try {
            const response = await fetch('/api/clone', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json'
              },
              body: JSON.stringify({ url })
            });
            
            const data = await response.json();
            
            if (data.success) {
              resultDiv.className = 'result success';
              resultDiv.innerHTML = `
                <h3>克隆成功！</h3>
                <p>标题：${data.title}</p>
                <p>保存目录：${data.save_dir}</p>
                <div class="links">
                  <a href="/cloned/${data.safe_url}/simple.html" target="_blank">查看简化版</a>
                  <a href="/cloned/${data.safe_url}/full.html" target="_blank">查看完整版</a>
                  <a href="/cloned/${data.safe_url}/info.json" target="_blank">查看提取信息</a>
                </div>
              `;
            } else {
              resultDiv.className = 'result error';
              resultDiv.innerHTML = `克隆失败：${data.error}`;
            }
          } catch (error) {
            resultDiv.className = 'result error';
            resultDiv.innerHTML = `克隆失败：${error.message}`;
          }
          
          // 刷新已克隆页面列表
          loadClonedPages();
        }
        
        // 加载已克隆页面列表
        async function loadClonedPages() {
          const response = await fetch('/api/cloned-pages');
          const pages = await response.json();
          const container = document.getElementById('clonedPages');
          
          if (pages.length === 0) {
            container.innerHTML = '<p>暂无克隆页面</p>';
            return;
          }
          
          container.innerHTML = pages.map(page => `
            <div class="cloned-item">
              <h3>${page.title}</h3>
              <p>URL: <a href="${page.url}" target="_blank">${page.url}</a></p>
              <p>克隆时间: ${new Date(page.timestamp).toLocaleString()}</p>
              <div class="links">
                <a href="/cloned/${page.safe_url}/simple.html" target="_blank">查看简化版</a>
                <a href="/cloned/${page.safe_url}/full.html" target="_blank">查看完整版</a>
                <a href="/cloned/${page.safe_url}/info.json" target="_blank">查看提取信息</a>
              </div>
            </div>
          `).join('');
        }
        
        // 页面加载时初始化
        window.onload = loadClonedPages;
      </script>
    </body>
    </html>
    '''


@app.post("/api/clone")
async def api_clone(request: Request):
    """API - 克隆网页"""
    data = await request.json()
    url = data.get('url')
    
    if not url:
        return JSONResponse(status_code=400, content={'success': False, 'error': '请提供URL'})
    
    result = clone_web_page(url)
    
    # 确保返回的结果包含正确的访问URL
    if result.get('success'):
        # 获取局域网IP用于本地访问
        local_network_ip = None
        try:
            import socket
            import fcntl
            import struct
            
            def get_local_ip(ifname):
                s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
                return socket.inet_ntoa(fcntl.ioctl(
                    s.fileno(),
                    0x8915,  # SIOCGIFADDR
                    struct.pack('256s', ifname[:15].encode('utf-8'))
                )[20:24])
            
            interfaces = ['eth0', 'en0', 'en1', 'wlan0', 'wifi0']
            for iface in interfaces:
                try:
                    ip = get_local_ip(iface)
                    if ip and ip != '127.0.0.1':
                        local_network_ip = ip
                        break
                except Exception:
                    continue
        except Exception:
            pass
        
        if not local_network_ip:
            # 使用传统方法获取局域网IP
            s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            s.connect(("8.8.8.8", 80))
            local_network_ip = s.getsockname()[0]
            s.close()
        
        safe_url = result.get('safe_url')
        
        # 构建本地访问URL
        local_access_url = f"http://{local_network_ip}:{SERVER_PORT}/cloned/{safe_url}/simple.html"
        local_full_url = f"http://{local_network_ip}:{SERVER_PORT}/cloned/{safe_url}/full.html"
        local_info_url = f"http://{local_network_ip}:{SERVER_PORT}/cloned/{safe_url}/info.json"
        
        # 构建外部访问URL
        external_access_url = f"http://{LOCAL_IP}:{SERVER_PORT}/cloned/{safe_url}/simple.html"
        external_full_url = f"http://{LOCAL_IP}:{SERVER_PORT}/cloned/{safe_url}/full.html"
        external_info_url = f"http://{LOCAL_IP}:{SERVER_PORT}/cloned/{safe_url}/info.json"
        
        # 构建返回结果
        return JSONResponse(content={
            'success': True,
            'title': result.get('title'),
            'save_dir': result.get('save_dir'),
            'safe_url': safe_url,
            # 本地网络访问地址
            'local_access': {
                'simple': local_access_url,
                'full': local_full_url,
                'info': local_info_url
            },
            # 外部网络访问地址
            'external_access': {
                'simple': external_access_url,
                'full': external_full_url,
                'info': external_info_url
            },
            'message': '克隆成功，本地网络使用local_access地址，外部网络使用external_access地址'
        })
    else:
        return JSONResponse(content=result)


@app.get("/api/info")
def api_info():
    """API - 获取服务器信息"""
    # 获取局域网IP用于本地访问
    local_network_ip = None
    try:
        import socket
        import fcntl
        import struct
        
        def get_local_ip(ifname):
            s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            return socket.inet_ntoa(fcntl.ioctl(
                s.fileno(),
                0x8915,  # SIOCGIFADDR
                struct.pack('256s', ifname[:15].encode('utf-8'))
            )[20:24])
        
        interfaces = ['eth0', 'en0', 'en1', 'wlan0', 'wifi0']
        for iface in interfaces:
            try:
                ip = get_local_ip(iface)
                if ip and ip != '127.0.0.1':
                    local_network_ip = ip
                    break
            except Exception:
                continue
    except Exception:
        pass
    
    if not local_network_ip:
        # 使用传统方法获取局域网IP
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.connect(("8.8.8.8", 80))
        local_network_ip = s.getsockname()[0]
        s.close()
    
    return JSONResponse(content={
        'success': True,
        'local_ip': local_network_ip,
        'public_ip': LOCAL_IP,
        'server_port': SERVER_PORT,
        'local_url': f"http://{local_network_ip}:{SERVER_PORT}",
        'external_url': f"http://{LOCAL_IP}:{SERVER_PORT}",
        'status': 'running',
        'message': '网页克隆服务正在运行',
        'access_info': {
            'local_access': f"http://{local_network_ip}:{SERVER_PORT}",
            'external_access': f"http://{LOCAL_IP}:{SERVER_PORT}",
            'note': '本地网络使用local_access地址，外部网络使用external_access地址'
        }
    })


@app.get("/api/cloned-pages")
def api_cloned_pages():
    """API - 获取已克隆页面列表"""
    pages = []
    
    try:
        for dir_name in os.listdir(CLONED_DIR):
            dir_path = os.path.join(CLONED_DIR, dir_name)
            if os.path.isdir(dir_path):
                info_path = os.path.join(dir_path, 'info.json')
                if os.path.exists(info_path):
                    try:
                        with open(info_path, 'r', encoding='utf-8') as f:
                            info = json.load(f)
                        info['safe_url'] = dir_name
                        pages.append(info)
                    except Exception as e:
                        print(f"读取{info_path}失败: {str(e)}")
    except Exception as e:
        print(f"获取克隆页面列表失败: {str(e)}")
    
    return JSONResponse(content=pages)


def get_local_ip():
    """获取本地网卡IP地址"""
    import socket
    try:
        # 创建套接字连接到外部地址，获取本地IP
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.connect(("8.8.8.8", 80))
        local_ip = s.getsockname()[0]
        s.close()
        return local_ip
    except Exception as e:
        print(f"⚠️ 获取本地IP失败: {str(e)}")
        return "127.0.0.1"


if __name__ == "__main__":
    import uvicorn
    import sys
    import threading
    import time
    import requests
    
    # 重新获取一次IP，对所有系统优先获取公网IP
    try:
        import os
        import socket
        
        print("🔍 尝试获取公网IP...")
        # 对所有系统，尝试通过curl命令获取公网IP
        import subprocess
        try:
            # 使用curl获取公网IP
            result = subprocess.run(
                ['curl', '-s', 'icanhazip.com'],
                capture_output=True,
                text=True,
                timeout=10
            )
            if result.returncode == 0:
                public_ip = result.stdout.strip()
                # 验证获取到的是有效的IP地址
                socket.inet_aton(public_ip)  # 验证IP格式
                LOCAL_IP = public_ip
                print(f"✅ 成功获取公网IP: {LOCAL_IP}")
        except Exception as e:
            print(f"⚠️ 获取公网IP失败，尝试获取局域网IP: {str(e)}")
        
        # 如果获取公网IP失败，尝试获取局域网IP
        if not LOCAL_IP:
            print("🔍 尝试获取局域网IP...")
            # 尝试通过网络接口获取IP
            try:
                # 导入必要的模块
                import fcntl
                import struct
                
                # 获取所有网络接口
                def get_ip_address(ifname):
                    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
                    return socket.inet_ntoa(fcntl.ioctl(
                        s.fileno(),
                        0x8915,  # SIOCGIFADDR
                        struct.pack('256s', ifname[:15].encode('utf-8'))
                    )[20:24])
                
                # 尝试获取常见物理网络接口的IP
                interfaces = ['eth0', 'en0', 'en1', 'wlan0', 'wifi0']
                for iface in interfaces:
                    try:
                        ip = get_ip_address(iface)
                        if ip and ip != '127.0.0.1':
                            LOCAL_IP = ip
                            print(f"✅ 成功获取局域网IP: {LOCAL_IP}")
                            break
                    except Exception:
                        continue
            except Exception:
                pass
            
            # 如果没有找到物理网络接口，使用传统方法
            if not LOCAL_IP:
                s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
                s.connect(("8.8.8.8", 80))
                LOCAL_IP = s.getsockname()[0]
                s.close()
                print(f"✅ 使用传统方法获取IP: {LOCAL_IP}")
    except Exception as e:
        print(f"⚠️ 更新IP失败，使用默认值: {LOCAL_IP}")
    
    def check_web_access():
        """检查WEB服务是否可以正常访问"""
        print("\n🔍 正在进行WEB服务自检...")
        time.sleep(2)  # 等待服务完全启动
        
        # 获取局域网IP（用于本地访问）
        local_network_ip = None
        try:
            import socket
            import fcntl
            import struct
            
            def get_local_ip(ifname):
                s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
                return socket.inet_ntoa(fcntl.ioctl(
                    s.fileno(),
                    0x8915,  # SIOCGIFADDR
                    struct.pack('256s', ifname[:15].encode('utf-8'))
                )[20:24])
            
            interfaces = ['eth0', 'en0', 'en1', 'wlan0', 'wifi0']
            for iface in interfaces:
                try:
                    ip = get_local_ip(iface)
                    if ip and ip != '127.0.0.1':
                        local_network_ip = ip
                        break
                except Exception:
                    continue
        except Exception:
            pass
        
        if not local_network_ip:
            # 使用传统方法获取局域网IP
            s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            s.connect(("8.8.8.8", 80))
            local_network_ip = s.getsockname()[0]
            s.close()
        
        # 测试本地访问（不测试公网IP，因为NAT网络本地无法直接访问公网IP）
        test_urls = [
            f"http://localhost:{SERVER_PORT}",
            f"http://127.0.0.1:{SERVER_PORT}",
            f"http://{local_network_ip}:{SERVER_PORT}",
            f"http://localhost:{SERVER_PORT}/api/info"
        ]
        
        for url in test_urls:
            try:
                response = requests.get(url, timeout=5)
                if response.status_code == 200:
                    print(f"✅ {url} - 访问成功 (状态码: {response.status_code})")
                else:
                    print(f"⚠️ {url} - 访问失败 (状态码: {response.status_code})")
            except requests.exceptions.RequestException as e:
                print(f"❌ {url} - 访问失败: {str(e)}")
        
        print("\n✅ WEB服务自检完成！")
        print("\n📋 访问地址说明：")
        print(f"🏠 本地访问地址: http://{local_network_ip}:{SERVER_PORT}")
        print(f"🌐 外部访问地址: http://{LOCAL_IP}:{SERVER_PORT}")
        print("💡 注意：本地无法直接测试外部访问地址（这是NAT网络的正常现象）")
        print("\n按 Ctrl+C 停止服务器\n")
    
    # 获取局域网IP用于显示
    local_network_ip = None
    try:
        import socket
        import fcntl
        import struct
        
        def get_local_ip(ifname):
            s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            return socket.inet_ntoa(fcntl.ioctl(
                s.fileno(),
                0x8915,  # SIOCGIFADDR
                struct.pack('256s', ifname[:15].encode('utf-8'))
            )[20:24])
        
        interfaces = ['eth0', 'en0', 'en1', 'wlan0', 'wifi0']
        for iface in interfaces:
            try:
                ip = get_local_ip(iface)
                if ip and ip != '127.0.0.1':
                    local_network_ip = ip
                    break
            except Exception:
                continue
    except Exception:
        pass
    
    if not local_network_ip:
        # 使用传统方法获取局域网IP
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.connect(("8.8.8.8", 80))
        local_network_ip = s.getsockname()[0]
        s.close()
    
    print("\n🚀 网页克隆工具（高可用版）正在启动...")
    print(f"🏠 本地访问地址: http://{local_network_ip}:{SERVER_PORT}")
    print(f"🌐 外部访问地址: http://{LOCAL_IP}:{SERVER_PORT}")
    print(f"📁 克隆页面保存目录: {CLONED_DIR}")
    print("🔧 服务器: Uvicorn (基于ASGI，支持epoll/kqueue)")
    print(f"🔍 监听地址: 0.0.0.0:{SERVER_PORT} (所有网络接口)")
    print(f"💡 注意：服务使用了固定端口 {SERVER_PORT}，避免端口冲突")
    print("💡 提示：本地网络使用本地访问地址，外部网络使用外部访问地址")
    
    # 启动自检线程
    check_thread = threading.Thread(target=check_web_access)
    check_thread.daemon = True
    check_thread.start()
    
    # 启动Uvicorn服务器，默认使用epoll/kqueue
    uvicorn.run(
        "app:app",
        host="0.0.0.0",
        port=SERVER_PORT,
        reload=False,
        workers=1,
        log_level="info"
    )
