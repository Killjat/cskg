#!/usr/bin/env python3
"""
MySQL客户端示例
用于连接MySQL服务器并执行简单查询
"""

import mysql.connector
from mysql.connector import Error

def mysql_client_example():
    """MySQL客户端示例函数"""
    try:
        # 连接到MySQL服务器
        connection = mysql.connector.connect(
            host='localhost',
            port=3306,
            user='test',
            password='test',
            database='test',
            charset='utf8'
        )
        
        if connection.is_connected():
            db_Info = connection.get_server_info()
            print(f"[+] 成功连接到MySQL服务器，版本: {db_Info}")
            
            # 创建游标对象
            cursor = connection.cursor()
            
            # 执行查询
            print("\n[+] 执行查询: SHOW DATABASES")
            cursor.execute("SHOW DATABASES")
            
            # 获取查询结果
            databases = cursor.fetchall()
            print("[+] 数据库列表:")
            for db in databases:
                print(f"   - {db[0]}")
            
            # 执行另一个查询
            print("\n[+] 执行查询: SELECT 'Hello, MySQL' as message")
            cursor.execute("SELECT 'Hello, MySQL' as message")
            
            # 获取查询结果
            result = cursor.fetchone()
            print(f"[+] 查询结果: {result[0]}")
            
    except Error as e:
        print(f"[-] 连接MySQL服务器时出错: {e}")
    finally:
        # 关闭数据库连接
        if 'connection' in locals() and connection.is_connected():
            cursor.close()
            connection.close()
            print("\n[+] 已关闭MySQL连接")

if __name__ == "__main__":
    print("=== MySQL客户端示例 ===")
    mysql_client_example()
