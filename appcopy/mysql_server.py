#!/usr/bin/env python3
"""
模拟MySQL服务器
实现基本的MySQL协议握手和简单查询响应
"""

import socket
import threading
import struct

class MySQLServer:
    def __init__(self, host='0.0.0.0', port=3306):
        self.host = host
        self.port = port
        self.server_socket = None
        self.running = False
    
    def start(self):
        """启动MySQL服务器"""
        try:
            self.server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            self.server_socket.bind((self.host, self.port))
            self.server_socket.listen(5)
            self.running = True
            print(f"[+] MySQL Server started on {self.host}:{self.port}")
            
            while self.running:
                client_socket, addr = self.server_socket.accept()
                print(f"[+] MySQL client connected from {addr}")
                threading.Thread(target=self.handle_client, args=(client_socket, addr)).start()
                
        except Exception as e:
            print(f"[-] MySQL server error: {e}")
            self.stop()
    
    def stop(self):
        """停止MySQL服务器"""
        self.running = False
        if self.server_socket:
            self.server_socket.close()
        print("[+] MySQL Server stopped")
    
    def handle_client(self, client_socket, addr):
        """处理MySQL客户端连接"""
        try:
            # 1. 发送握手包
            handshake_packet = self._create_handshake_packet()
            client_socket.send(handshake_packet)
            print(f"[+] Sent handshake packet to {addr}")
            
            # 2. 接收客户端认证包
            auth_packet = client_socket.recv(1024)
            if not auth_packet:
                return
            print(f"[+] Received authentication packet from {addr}")
            
            # 3. 发送认证成功响应
            ok_packet = self._create_ok_packet()
            client_socket.send(ok_packet)
            print(f"[+] Sent authentication success to {addr}")
            
            # 4. 处理客户端查询
            while self.running:
                query_packet = client_socket.recv(1024)
                if not query_packet:
                    break
                
                # 解析查询
                query = self._parse_query_packet(query_packet)
                print(f"[+] Received query from {addr}: {query}")
                
                # 处理查询并发送响应
                response = self._handle_query(query)
                client_socket.send(response)
                print(f"[+] Sent response to {addr}")
                
        except Exception as e:
            print(f"[-] Error handling MySQL client {addr}: {e}")
        finally:
            client_socket.close()
            print(f"[-] MySQL client {addr} disconnected")
    
    def _create_handshake_packet(self):
        """创建MySQL握手包"""
        # 简单的握手包实现
        protocol_version = 0x0a  # MySQL 5.7+
        server_version = b'5.7.33-0ubuntu0.16.04.1'
        thread_id = 0x1234
        scramble_buff1 = b'\x01\x02\x03\x04\x05\x06\x07\x08'
        filler = b'\x00'
        server_capabilities = struct.pack('<I', 0xffffffef)
        server_language = 0x21  # utf8_general_ci
        server_status = 0x02
        scramble_buff2 = b'\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10'
        
        handshake = b''
        handshake += bytes([protocol_version])
        handshake += server_version + filler
        handshake += struct.pack('<I', thread_id)
        handshake += scramble_buff1
        handshake += filler
        handshake += server_capabilities
        handshake += bytes([server_language])
        handshake += struct.pack('<H', server_status)
        handshake += server_capabilities[2:4]  # 扩展能力标志
        handshake += b'\x15'  # 身份验证插件数据长度
        handshake += b'\x00' * 10  # 保留字段
        handshake += scramble_buff2
        handshake += b'\x00'  # 身份验证插件名称结束
        handshake += b'mysql_native_password\x00'  # 身份验证插件名称
        
        return handshake
    
    def _create_ok_packet(self):
        """创建OK响应包"""
        ok_packet = b''
        ok_packet += b'\x00'  # 状态标志
        ok_packet += struct.pack('<L', 0)  # 受影响的行数
        ok_packet += struct.pack('<H', 0)  # 最后插入的ID
        ok_packet += struct.pack('<H', 0)  # 状态标志
        ok_packet += struct.pack('<H', 0)  # 警告计数
        ok_packet += b'\x00'  # 状态信息
        
        return ok_packet
    
    def _parse_query_packet(self, packet):
        """解析查询包"""
        # 跳过包长度和序号
        if len(packet) < 5:
            return ""
        
        # 查询类型（0x03 = COM_QUERY）
        if packet[4] != 0x03:
            return ""
        
        # 提取查询字符串
        query = packet[5:].decode('utf-8', errors='ignore')
        return query
    
    def _handle_query(self, query):
        """处理客户端查询"""
        # 简单处理SELECT查询
        if query.upper().startswith('SELECT'):
            return self._create_result_set_packet()
        else:
            return self._create_ok_packet()
    
    def _create_result_set_packet(self):
        """创建结果集响应"""
        # 字段数量
        field_count = b'\x01'  # 1个字段
        
        # 字段定义
        field_def = b''
        field_def += b'\x03'  #  Catalog (def)
        field_def += b'def\x00'
        field_def += b'\x03'  #  Schema
        field_def += b'test\x00'
        field_def += b'\x03'  #  Table
        field_def += b'table\x00'
        field_def += b'\x03'  #  Original table
        field_def += b'table\x00'
        field_def += b'\x04'  #  Name
        field_def += b'column\x00'
        field_def += b'\x04'  #  Original name
        field_def += b'column\x00'
        field_def += b'\x0c'  #  Length of the following fields
        field_def += b'\x00\x0c'  #  Character set
        field_def += struct.pack('<L', 255)  #  Column length
        field_def += b'\x03'  #  Type
        field_def += b'\x00'  #  Flags
        field_def += b'\x00'  #  Decimals
        field_def += b'\x00\x00'  #  Filler
        
        # 字段结束包
        eof_packet = b'\xfe'  # EOF header
        eof_packet += struct.pack('<H', 0)  #  Warning count
        eof_packet += struct.pack('<H', 0)  #  Status flags
        
        # 数据行
        row_packet = b''
        row_packet += b'\x05'  # 长度编码的字符串长度
        row_packet += b'value'  # 字符串值
        
        # 结果集结束包
        result_eof_packet = eof_packet
        
        # 组合所有响应包
        result = field_count + field_def + eof_packet + row_packet + result_eof_packet
        
        return result

if __name__ == "__main__":
    try:
        mysql_server = MySQLServer()
        mysql_server.start()
    except KeyboardInterrupt:
        mysql_server.stop()
