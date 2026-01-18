#!/usr/bin/env python3
"""
Redis客户端示例
用于连接Redis服务器并执行基本命令
"""

try:
    import redis
except ImportError:
    print("[-] 请先安装redis库: pip3 install redis")
    import sys
    sys.exit(1)

def redis_client_example():
    """Redis客户端示例函数"""
    try:
        # 连接到Redis服务器
        r = redis.Redis(
            host='localhost',
            port=6379,
            db=0,
            decode_responses=True
        )
        
        # 测试连接
        print("[+] 测试连接...")
        pong = r.ping()
        if pong:
            print(f"[+] 成功连接到Redis服务器: PONG={pong}")
        else:
            print("[-] 连接Redis服务器失败")
            return False
        
        # 执行SET命令
        print("\n[+] 执行命令: SET test_key test_value")
        set_result = r.set('test_key', 'test_value')
        print(f"[+] SET命令结果: {set_result}")
        
        # 执行GET命令
        print("\n[+] 执行命令: GET test_key")
        get_result = r.get('test_key')
        print(f"[+] GET命令结果: {get_result}")
        
        # 执行ECHO命令
        print("\n[+] 执行命令: ECHO Hello, Redis!")
        echo_result = r.echo('Hello, Redis!')
        print(f"[+] ECHO命令结果: {echo_result}")
        
        # 执行SET命令设置多个键值对
        print("\n[+] 执行命令: MSET key1 value1 key2 value2")
        mset_result = r.mset({'key1': 'value1', 'key2': 'value2'})
        print(f"[+] MSET命令结果: {mset_result}")
        
        # 执行KEYS命令
        print("\n[+] 执行命令: KEYS *")
        keys_result = r.keys('*')
        print(f"[+] KEYS命令结果: {keys_result}")
        
        # 执行DBSIZE命令
        print("\n[+] 执行命令: DBSIZE")
        dbsize_result = r.dbsize()
        print(f"[+] DBSIZE命令结果: {dbsize_result}")
        
        # 执行DEL命令
        print("\n[+] 执行命令: DEL test_key")
        del_result = r.delete('test_key')
        print(f"[+] DEL命令结果: {del_result}")
        
        # 再次执行GET命令检查键是否存在
        print("\n[+] 执行命令: GET test_key")
        get_result = r.get('test_key')
        print(f"[+] GET命令结果: {get_result}")
        
        # 执行FLUSHDB命令
        print("\n[+] 执行命令: FLUSHDB")
        flush_result = r.flushdb()
        print(f"[+] FLUSHDB命令结果: {flush_result}")
        
        # 再次执行DBSIZE命令检查数据库大小
        print("\n[+] 执行命令: DBSIZE")
        dbsize_result = r.dbsize()
        print(f"[+] DBSIZE命令结果: {dbsize_result}")
        
        print("\n[+] Redis客户端示例执行完成")
        return True
        
    except redis.ConnectionError as e:
        print(f"[-] 连接Redis服务器时出错: {e}")
        return False
    except redis.RedisError as e:
        print(f"[-] Redis命令执行错误: {e}")
        return False
    except Exception as e:
        print(f"[-] 执行Redis命令时出错: {e}")
        return False
    finally:
        # 关闭Redis连接
        if 'r' in locals():
            r.close()
            print("[+] 已关闭Redis连接")

if __name__ == "__main__":
    print("=== Redis客户端示例 ===")
    redis_client_example()
