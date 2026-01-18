#!/usr/bin/env python3
"""
模拟CVE-2021-4161漏洞的Modbus网关
该漏洞存在于Moxa MGate MB3180/MB3280/MB3480系列网关中
漏洞类型：敏感信息明文传输
攻击者可通过嗅探网络流量窃取设备登录凭据
"""

import socket
import threading
import time
from http.server import BaseHTTPRequestHandler, HTTPServer

class VulnerableModbusGateway:
    def __init__(self, host='0.0.0.0', modbus_port=502, web_port=80):
        self.host = host
        self.modbus_port = modbus_port
        self.web_port = web_port
        self.modbus_server = None
        self.web_server = None
        
    def start_modbus_server(self):
        """启动模拟的Modbus TCP服务器"""
        try:
            # 创建TCP套接字
            self.modbus_server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.modbus_server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            self.modbus_server.bind((self.host, self.modbus_port))
            self.modbus_server.listen(5)
            print(f"[+] Modbus TCP Server started on {self.host}:{self.modbus_port}")
            
            while True:
                client_socket, addr = self.modbus_server.accept()
                print(f"[+] Modbus client connected from {addr}")
                # 简单处理Modbus请求（仅响应基本帧）
                threading.Thread(target=self.handle_modbus_client, args=(client_socket, addr)).start()
                
        except Exception as e:
            print(f"[-] Modbus server error: {e}")
            if self.modbus_server:
                self.modbus_server.close()
    
    def handle_modbus_client(self, client_socket, addr):
        """处理Modbus客户端连接"""
        try:
            while True:
                # 接收Modbus请求
                data = client_socket.recv(1024)
                if not data:
                    break
                
                # 简单响应Modbus请求
                # 仅处理保持寄存器读取请求（功能码03）
                if len(data) >= 8 and data[7] == 0x03:
                    # 构造简单响应
                    response = data[:8] + b'\x04\x00\x01\x00\x02'
                    client_socket.send(response)
                    print(f"[+] Responded to Modbus request from {addr}")
                    
        except Exception as e:
            print(f"[-] Error handling Modbus client {addr}: {e}")
        finally:
            client_socket.close()
            print(f"[-] Modbus client {addr} disconnected")
    
    def start_web_server(self):
        """启动有漏洞的Web管理界面服务器"""
        class VulnerableWebHandler(BaseHTTPRequestHandler):
            """有漏洞的Web处理程序，明文传输登录凭据"""
            
            def do_GET(self):
                """处理GET请求"""
                if self.path == '/':
                    # 返回登录页面
                    self.send_response(200)
                    self.send_header('Content-type', 'text/html')
                    self.end_headers()
                    login_page = """
                    <html>
                    <body>
                        <h1>Moxa MGate Gateway Login</h1>
                        <form action="/login" method="POST">
                            <label for="username">Username:</label><br>
                            <input type="text" id="username" name="username"><br>
                            <label for="password">Password:</label><br>
                            <input type="password" id="password" name="password"><br><br>
                            <input type="submit" value="Login">
                        </form>
                    </body>
                    </html>
                    """
                    self.wfile.write(login_page.encode('utf-8'))
                else:
                    self.send_response(404)
                    self.end_headers()
            
            def do_POST(self):
                """处理POST请求 - 漏洞点：明文传输凭据"""
                if self.path == '/login':
                    # 获取请求体长度
                    content_length = int(self.headers['Content-Length'])
                    # 读取请求体（明文传输的凭据）
                    post_data = self.rfile.read(content_length).decode('utf-8')
                    
                    print(f"[!] VULNERABLE: Login credentials captured in plaintext: {post_data}")
                    
                    # 解析用户名和密码
                    import urllib.parse
                    credentials = urllib.parse.parse_qs(post_data)
                    username = credentials.get('username', [''])[0]
                    password = credentials.get('password', [''])[0]
                    
                    # 简单验证（总是失败）
                    self.send_response(200)
                    self.send_header('Content-type', 'text/html')
                    self.end_headers()
                    response = f"""
                    <html>
                    <body>
                        <h1>Login Failed</h1>
                        <p>Username: {username}</p>
                        <p>Password: {password}</p>
                        <p>Invalid credentials. Please try again.</p>
                        <a href="/">Back to login</a>
                    </body>
                    </html>
                    """
                    self.wfile.write(response.encode('utf-8'))
                else:
                    self.send_response(404)
                    self.end_headers()
        
        try:
            server_address = (self.host, self.web_port)
            self.web_server = HTTPServer(server_address, VulnerableWebHandler)
            print(f"[+] Web Management Server started on {self.host}:{self.web_port}")
            print(f"[!] VULNERABILITY: Web login credentials are transmitted in PLAINTEXT")
            self.web_server.serve_forever()
        except Exception as e:
            print(f"[-] Web server error: {e}")
            if self.web_server:
                self.web_server.shutdown()
    
    def start(self):
        """启动所有服务"""
        # 启动Modbus服务器线程
        modbus_thread = threading.Thread(target=self.start_modbus_server)
        modbus_thread.daemon = True
        modbus_thread.start()
        
        # 启动Web服务器线程
        web_thread = threading.Thread(target=self.start_web_server)
        web_thread.daemon = True
        web_thread.start()
        
        try:
            while True:
                time.sleep(1)
        except KeyboardInterrupt:
            print("\n[+] Shutting down gateway...")
            if self.modbus_server:
                self.modbus_server.close()
            if self.web_server:
                self.web_server.shutdown()

if __name__ == "__main__":
    print("=== CVE-2021-4161 Vulnerable Modbus Gateway ===")
    print("This simulates the Moxa MGate gateway vulnerability where login credentials are transmitted in plaintext")
    print("\nUsage:")
    print("1. Run this gateway")
    print("2. Use the POC to capture credentials")
    print("3. Open http://localhost in browser and try to login")
    print("\nPress Ctrl+C to exit")
    
    gateway = VulnerableModbusGateway()
    gateway.start()
