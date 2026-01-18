#!/usr/bin/env python3
"""
Redis启动失败诊断脚本
用于检查Redis启动失败的可能原因
"""

import socket
import os
import subprocess

def check_port(port):
    """检查端口是否被占用"""
    print(f"\n1. 检查端口 {port} 是否被占用...")
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        s.bind(('0.0.0.0', port))
        print(f"   ✅ 端口 {port} 可用")
        s.close()
        return True
    except OSError as e:
        print(f"   ❌ 端口 {port} 被占用: {e}")
        return False

def check_file_permissions(file_path):
    """检查文件权限"""
    print(f"\n2. 检查文件 {file_path} 权限...")
    if os.path.exists(file_path):
        print(f"   ✅ 文件存在")
        print(f"   权限: {oct(os.stat(file_path).st_mode)[-3:]}")
        return True
    else:
        print(f"   ❌ 文件不存在")
        return False

def test_redis_server():
    """测试Redis服务器代码"""
    print("\n3. 测试Redis服务器代码...")
    try:
        import redis_server
        print("   ✅ Redis服务器模块导入成功")
        return True
    except Exception as e:
        print(f"   ❌ Redis服务器模块导入失败: {e}")
        return False

def check_python_version():
    """检查Python版本"""
    print("\n4. 检查Python版本...")
    import sys
    print(f"   Python版本: {sys.version}")
    if sys.version_info >= (3, 6):
        print("   ✅ Python版本符合要求 (>= 3.6)")
        return True
    else:
        print("   ❌ Python版本过低，需要 >= 3.6")
        return False

def run_redis_diagnostics():
    """运行Redis诊断"""
    print("=== Redis启动失败诊断报告 ===")
    print(f"诊断时间: {subprocess.check_output(['date']).decode('utf-8').strip()}")
    
    # 检查端口
    port_available = check_port(6379)
    
    # 检查文件权限
    check_file_permissions("redis_server.py")
    
    # 测试Redis服务器代码
    test_redis_server()
    
    # 检查Python版本
    check_python_version()
    
    print("\n=== 诊断建议 ===")
    if not port_available:
        print("1. 端口6379被占用，建议:")
        print("   - 查找并关闭占用该端口的进程")
        print("   - 或修改Redis服务器代码，使用其他端口")
    
    print("\n2. 查看Redis服务器日志，了解具体错误:")
    print("   python3 redis_server.py")
    
    print("\n3. 检查系统日志:")
    print("   tail -n 50 logs/system_*.log")
    
    print("\n=== 诊断完成 ===")

if __name__ == "__main__":
    run_redis_diagnostics()
