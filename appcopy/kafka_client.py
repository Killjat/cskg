#!/usr/bin/env python3
"""
Kafka客户端示例
用于连接Kafka服务器并执行基本操作
"""

try:
    from kafka import KafkaProducer, KafkaConsumer
    from kafka.errors import KafkaError
except ImportError:
    print("[-] 请先安装kafka-python库: pip3 install kafka-python")
    import sys
    sys.exit(1)

def kafka_producer_example():
    """Kafka生产者示例"""
    try:
        # 创建Kafka生产者
        producer = KafkaProducer(
            bootstrap_servers=['localhost:9092'],
            client_id='test-producer',
            value_serializer=lambda v: str(v).encode('utf-8')
        )
        
        print("[+] Kafka生产者已创建")
        
        # 发送消息
        topic = 'test_topic'
        message = 'Hello, Kafka!'
        
        print(f"[+] 发送消息到主题 {topic}: {message}")
        future = producer.send(topic, value=message)
        
        # 等待消息发送完成
        record_metadata = future.get(timeout=10)
        print(f"[+] 消息发送成功: 主题={record_metadata.topic}, 分区={record_metadata.partition}, 偏移量={record_metadata.offset}")
        
        # 关闭生产者
        producer.close()
        print("[+] Kafka生产者已关闭")
        
        return True
        
    except KafkaError as e:
        print(f"[-] Kafka生产者错误: {e}")
        return False
    except Exception as e:
        print(f"[-] 发送消息时出错: {e}")
        return False

def kafka_consumer_example():
    """Kafka消费者示例"""
    try:
        # 创建Kafka消费者
        consumer = KafkaConsumer(
            'test_topic',
            bootstrap_servers=['localhost:9092'],
            client_id='test-consumer',
            group_id='test-group',
            auto_offset_reset='earliest',
            value_deserializer=lambda x: x.decode('utf-8')
        )
        
        print("[+] Kafka消费者已创建")
        print("[+] 开始消费消息 (5秒后自动停止)...")
        
        # 消费消息，最多消费5条或5秒后停止
        import time
        start_time = time.time()
        message_count = 0
        
        for message in consumer:
            print(f"[+] 收到消息: 主题={message.topic}, 分区={message.partition}, 偏移量={message.offset}, 键={message.key}, 值={message.value}")
            message_count += 1
            
            # 停止条件
            if message_count >= 5 or time.time() - start_time > 5:
                break
        
        # 关闭消费者
        consumer.close()
        print("[+] Kafka消费者已关闭")
        
        return True
        
    except KafkaError as e:
        print(f"[-] Kafka消费者错误: {e}")
        return False
    except Exception as e:
        print(f"[-] 消费消息时出错: {e}")
        return False

def kafka_client_example():
    """Kafka客户端示例主函数"""
    print("=== Kafka客户端示例 ===")
    
    # 运行生产者示例
    print("\n--- 生产者示例 ---")
    kafka_producer_example()
    
    # 运行消费者示例
    print("\n--- 消费者示例 ---")
    kafka_consumer_example()

if __name__ == "__main__":
    kafka_client_example()
