#!/usr/bin/env python3
"""
模拟Kafka服务器
实现基本的Kafka协议握手和元数据请求响应
"""

import socket
import threading
import struct

class KafkaServer:
    def __init__(self, host='0.0.0.0', port=9092):
        self.host = host
        self.port = port
        self.server_socket = None
        self.running = False
    
    def start(self):
        """启动Kafka服务器"""
        try:
            self.server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            self.server_socket.bind((self.host, self.port))
            self.server_socket.listen(5)
            self.running = True
            print(f"[+] Kafka Server started on {self.host}:{self.port}")
            
            while self.running:
                client_socket, addr = self.server_socket.accept()
                print(f"[+] Kafka client connected from {addr}")
                threading.Thread(target=self.handle_client, args=(client_socket, addr)).start()
                
        except Exception as e:
            print(f"[-] Kafka server error: {e}")
            self.stop()
    
    def stop(self):
        """停止Kafka服务器"""
        self.running = False
        if self.server_socket:
            self.server_socket.close()
        print("[+] Kafka Server stopped")
    
    def handle_client(self, client_socket, addr):
        """处理Kafka客户端连接"""
        try:
            while self.running:
                # 读取Kafka请求头
                request_header = client_socket.recv(4)
                if not request_header:
                    break
                
                # 解析请求长度
                request_len = struct.unpack('>i', request_header)[0]
                
                # 读取请求内容
                request_data = client_socket.recv(request_len)
                if not request_data:
                    break
                
                # 解析请求类型
                api_key = struct.unpack('>h', request_data[0:2])[0]
                api_version = struct.unpack('>h', request_data[2:4])[0]
                
                print(f"[+] Received request from {addr}: api_key={api_key}, api_version={api_version}")
                
                # 处理不同类型的请求
                if api_key == 3:  # MetadataRequest
                    response = self._handle_metadata_request(request_data)
                elif api_key == 0:  # ProduceRequest
                    response = self._handle_produce_request(request_data)
                elif api_key == 1:  # FetchRequest
                    response = self._handle_fetch_request(request_data)
                else:
                    response = self._create_error_response(request_data, "Unsupported api_key")
                
                # 发送响应
                client_socket.send(response)
                print(f"[+] Sent response to {addr}")
                
        except Exception as e:
            print(f"[-] Error handling Kafka client {addr}: {e}")
        finally:
            client_socket.close()
            print(f"[-] Kafka client {addr} disconnected")
    
    def _handle_metadata_request(self, request_data):
        """处理元数据请求"""
        # 解析请求
        correlation_id = struct.unpack('>i', request_data[4:8])[0]
        
        # 创建元数据响应
        # 响应格式: [response_len][correlation_id][brokers_len][brokers][topic_metadata_len][topic_metadata]
        
        # Brokers部分
        brokers_len = 1
        broker_id = 0
        host = b'localhost'
        port = 9092
        
        brokers = struct.pack('>i', brokers_len)
        brokers += struct.pack('>i', broker_id)
        brokers += struct.pack('>h', len(host)) + host
        brokers += struct.pack('>i', port)
        
        # Topic元数据部分
        topic_metadata_len = 1
        topic_error_code = 0
        topic_name = b'test_topic'
        topic_is_internal = 0
        partition_metadata_len = 1
        partition_error_code = 0
        partition_id = 0
        leader_id = 0
        replicas_len = 1
        isr_len = 1
        
        topic_metadata = struct.pack('>i', topic_metadata_len)
        topic_metadata += struct.pack('>h', topic_error_code)
        topic_metadata += struct.pack('>h', len(topic_name)) + topic_name
        topic_metadata += struct.pack('>b', topic_is_internal)
        topic_metadata += struct.pack('>i', partition_metadata_len)
        topic_metadata += struct.pack('>h', partition_error_code)
        topic_metadata += struct.pack('>i', partition_id)
        topic_metadata += struct.pack('>i', leader_id)
        topic_metadata += struct.pack('>i', replicas_len) + struct.pack('>i', broker_id)
        topic_metadata += struct.pack('>i', isr_len) + struct.pack('>i', broker_id)
        
        # 组合响应
        response_body = struct.pack('>i', correlation_id)
        response_body += brokers
        response_body += topic_metadata
        
        # 添加响应长度
        response = struct.pack('>i', len(response_body)) + response_body
        
        return response
    
    def _handle_produce_request(self, request_data):
        """处理生产请求"""
        correlation_id = struct.unpack('>i', request_data[4:8])[0]
        
        # 创建成功响应
        response_body = struct.pack('>i', correlation_id)
        response_body += struct.pack('>h', 0)  # Throttle time
        response_body += struct.pack('>i', 0)  # Topic count
        
        response = struct.pack('>i', len(response_body)) + response_body
        return response
    
    def _handle_fetch_request(self, request_data):
        """处理获取请求"""
        correlation_id = struct.unpack('>i', request_data[4:8])[0]
        
        # 创建空响应
        response_body = struct.pack('>i', correlation_id)
        response_body += struct.pack('>h', 0)  # Throttle time
        response_body += struct.pack('>i', 0)  # Topic count
        
        response = struct.pack('>i', len(response_body)) + response_body
        return response
    
    def _create_error_response(self, request_data, error_msg):
        """创建错误响应"""
        correlation_id = struct.unpack('>i', request_data[4:8])[0]
        
        response_body = struct.pack('>i', correlation_id)
        response_body += struct.pack('>h', 1)  # Error code: UnknownServerError
        
        response = struct.pack('>i', len(response_body)) + response_body
        return response

if __name__ == "__main__":
    try:
        kafka_server = KafkaServer()
        kafka_server.start()
    except KeyboardInterrupt:
        kafka_server.stop()
