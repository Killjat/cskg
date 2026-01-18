#!/usr/bin/env python3
"""
模拟Redis服务器
实现基本的Redis协议和命令处理
"""

import socket
import threading

class RedisServer:
    def __init__(self, host='0.0.0.0', port=6379):
        self.host = host
        self.port = port
        self.server_socket = None
        self.running = False
        self.data_store = {}  # 简单的数据存储
    
    def start(self):
        """启动Redis服务器"""
        try:
            self.server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            self.server_socket.bind((self.host, self.port))
            self.server_socket.listen(5)
            self.running = True
            print(f"[+] Redis Server started on {self.host}:{self.port}")
            
            while self.running:
                client_socket, addr = self.server_socket.accept()
                print(f"[+] Redis client connected from {addr}")
                threading.Thread(target=self.handle_client, args=(client_socket, addr)).start()
                
        except Exception as e:
            print(f"[-] Redis server error: {e}")
            self.stop()
    
    def stop(self):
        """停止Redis服务器"""
        self.running = False
        if self.server_socket:
            self.server_socket.close()
        print("[+] Redis Server stopped")
    
    def handle_client(self, client_socket, addr):
        """处理Redis客户端连接"""
        try:
            buffer = b''
            
            while self.running:
                # 接收数据
                data = client_socket.recv(1024)
                if not data:
                    break
                
                buffer += data
                
                # 处理完整的Redis命令
                while b'\r\n' in buffer:
                    # 解析Redis命令
                    command, remaining = self._parse_command(buffer)
                    buffer = remaining
                    
                    if command:
                        print(f"[+] Received command from {addr}: {command}")
                        
                        # 处理命令
                        response = self._handle_command(command)
                        
                        # 发送响应
                        client_socket.send(response)
                        print(f"[+] Sent response to {addr}")
                        
        except Exception as e:
            print(f"[-] Error handling Redis client {addr}: {e}")
        finally:
            client_socket.close()
            print(f"[-] Redis client {addr} disconnected")
    
    def _parse_command(self, data):
        """解析Redis RESP协议命令"""
        # Redis RESP协议命令格式: *<number of arguments>\r\n$<length of argument>\r\n<argument>\r\n...
        
        # 检查是否包含完整的命令
        if data[0:1] != b'*':
            return None, data
        
        # 找到第一个\r\n
        end_pos = data.find(b'\r\n')
        if end_pos == -1:
            return None, data
        
        # 解析参数数量
        try:
            arg_count = int(data[1:end_pos])
        except ValueError:
            return None, data
        
        # 解析所有参数
        args = []
        pos = end_pos + 2  # 跳过\r\n
        for _ in range(arg_count):
            # 检查是否是$开头
            if pos >= len(data) or data[pos:pos+1] != b'$':
                return None, data
            
            # 找到参数长度结束位置
            end_len_pos = data.find(b'\r\n', pos)
            if end_len_pos == -1:
                return None, data
            
            # 解析参数长度
            try:
                arg_len = int(data[pos+1:end_len_pos])
            except ValueError:
                return None, data
            
            # 检查参数是否完整
            arg_start = end_len_pos + 2
            arg_end = arg_start + arg_len + 2  # +2 for \r\n
            if arg_end > len(data):
                return None, data
            
            # 提取参数
            arg = data[arg_start:arg_start+arg_len].decode('utf-8')
            args.append(arg.upper())
            
            # 更新位置
            pos = arg_end
        
        # 返回解析后的命令和剩余数据
        return args, data[pos:]
    
    def _handle_command(self, command):
        """处理Redis命令"""
        if not command:
            return b'-ERR invalid command\r\n'
        
        cmd = command[0]
        
        # 处理PING命令
        if cmd == 'PING':
            return b'+PONG\r\n'
        
        # 处理ECHO命令
        elif cmd == 'ECHO':
            if len(command) < 2:
                return b'-ERR wrong number of arguments for ECHO command\r\n'
            msg = command[1]
            return f'${len(msg)}\r\n{msg}\r\n'.encode('utf-8')
        
        # 处理SET命令
        elif cmd == 'SET':
            if len(command) < 3:
                return b'-ERR wrong number of arguments for SET command\r\n'
            key = command[1]
            value = command[2]
            self.data_store[key] = value
            return b'+OK\r\n'
        
        # 处理GET命令
        elif cmd == 'GET':
            if len(command) < 2:
                return b'-ERR wrong number of arguments for GET command\r\n'
            key = command[1]
            if key in self.data_store:
                value = self.data_store[key]
                return f'${len(value)}\r\n{value}\r\n'.encode('utf-8')
            else:
                return b'$-1\r\n'  # 键不存在
        
        # 处理DEL命令
        elif cmd == 'DEL':
            if len(command) < 2:
                return b'-ERR wrong number of arguments for DEL command\r\n'
            keys = command[1:]
            deleted = 0
            for key in keys:
                if key in self.data_store:
                    del self.data_store[key]
                    deleted += 1
            return f':{deleted}\r\n'.encode('utf-8')
        
        # 处理KEYS命令
        elif cmd == 'KEYS':
            if len(command) < 2:
                return b'-ERR wrong number of arguments for KEYS command\r\n'
            pattern = command[1]
            # 简单的模式匹配，只支持*通配符
            if pattern == '*':
                keys = list(self.data_store.keys())
            else:
                # 简单替换*为.*，使用正则匹配
                import re
                regex = pattern.replace('*', '.*')
                keys = [key for key in self.data_store.keys() if re.match(regex, key)]
            
            # 构造数组响应
            response = f'*{len(keys)}\r\n'.encode('utf-8')
            for key in keys:
                response += f'${len(key)}\r\n{key}\r\n'.encode('utf-8')
            
            return response
        
        # 处理DBSIZE命令
        elif cmd == 'DBSIZE':
            size = len(self.data_store)
            return f':{size}\r\n'.encode('utf-8')
        
        # 处理FLUSHDB命令
        elif cmd == 'FLUSHDB':
            self.data_store.clear()
            return b'+OK\r\n'
        
        # 不支持的命令
        else:
            return b'-ERR unknown command\r\n'

if __name__ == "__main__":
    try:
        redis_server = RedisServer()
        redis_server.start()
    except KeyboardInterrupt:
        redis_server.stop()
