#!/usr/bin/env python3
"""
系统日志生成模块
用于记录系统和服务的运行日志
"""

import logging
import logging.handlers
import os
import datetime

class SystemLogger:
    """系统日志管理器"""
    def __init__(self, log_dir="logs"):
        self.log_dir = log_dir
        self.logger = self._setup_logger()
    
    def _setup_logger(self):
        """配置日志记录器"""
        # 创建日志目录
        if not os.path.exists(self.log_dir):
            os.makedirs(self.log_dir)
        
        # 日志文件名格式：system_YYYY-MM-DD.log
        log_file = os.path.join(self.log_dir, f"system_{datetime.datetime.now().strftime('%Y-%m-%d')}.log")
        
        # 创建日志记录器
        logger = logging.getLogger("SystemLogger")
        logger.setLevel(logging.INFO)
        
        # 创建文件处理器，按天滚动
        handler = logging.handlers.TimedRotatingFileHandler(
            log_file,
            when='midnight',
            backupCount=7
        )
        
        # 设置日志格式
        formatter = logging.Formatter(
            '%(asctime)s - %(levelname)s - %(module)s - %(message)s',
            datefmt='%Y-%m-%d %H:%M:%S'
        )
        handler.setFormatter(formatter)
        
        # 添加处理器
        logger.addHandler(handler)
        
        return logger
    
    def log_info(self, message):
        """记录信息日志"""
        self.logger.info(message)
    
    def log_warning(self, message):
        """记录警告日志"""
        self.logger.warning(message)
    
    def log_error(self, message):
        """记录错误日志"""
        self.logger.error(message)
    
    def log_critical(self, message):
        """记录严重错误日志"""
        self.logger.critical(message)
    
    def log_system_start(self, app_name, version):
        """记录系统启动日志"""
        self.log_info(f"System started: {app_name} v{version}")
    
    def log_system_stop(self, app_name, version):
        """记录系统停止日志"""
        self.log_info(f"System stopped: {app_name} v{version}")
    
    def log_service_start(self, service_name, version, status):
        """记录服务启动日志"""
        self.log_info(f"Service started: {service_name} v{version} - Status: {status}")
    
    def log_service_stop(self, service_name, version, status):
        """记录服务停止日志"""
        self.log_info(f"Service stopped: {service_name} v{version} - Status: {status}")
    
    def log_service_error(self, service_name, version, error):
        """记录服务错误日志"""
        self.log_error(f"Service error: {service_name} v{version} - Error: {error}")
    
    def log_access(self, ip, path, method, status_code):
        """记录访问日志"""
        self.log_info(f"Access: {ip} - {method} {path} - Status: {status_code}")

# 单例实例
logger = None

# 初始化日志器
def init_logger():
    """初始化日志器"""
    global logger
    if logger is None:
        logger = SystemLogger()
    return logger

# 获取日志器实例
def get_logger():
    """获取日志器实例"""
    global logger
    if logger is None:
        return init_logger()
    return logger

# 便捷方法
def log_info(message):
    """便捷的信息日志记录"""
    get_logger().log_info(message)

def log_warning(message):
    """便捷的警告日志记录"""
    get_logger().log_warning(message)

def log_error(message):
    """便捷的错误日志记录"""
    get_logger().log_error(message)

def log_critical(message):
    """便捷的严重错误日志记录"""
    get_logger().log_critical(message)

def log_service_start(service_name, version, status):
    """便捷的服务启动日志记录"""
    get_logger().log_service_start(service_name, version, status)

def log_service_stop(service_name, version, status):
    """便捷的服务停止日志记录"""
    get_logger().log_service_stop(service_name, version, status)

def log_service_error(service_name, version, error):
    """便捷的服务错误日志记录"""
    get_logger().log_service_error(service_name, version, error)

def log_access(ip, path, method, status_code):
    """便捷的访问日志记录"""
    get_logger().log_access(ip, path, method, status_code)

# 测试代码
if __name__ == "__main__":
    # 初始化日志器
    init_logger()
    
    # 记录测试日志
    log_info("Test info message")
    log_warning("Test warning message")
    log_error("Test error message")
    log_critical("Test critical message")
    
    # 记录服务日志
    log_service_start("TestService", "1.0", "running")
    log_service_stop("TestService", "1.0", "stopped")
    log_service_error("TestService", "1.0", "Connection refused")
    
    # 记录访问日志
    log_access("127.0.0.1", "/api/services", "GET", 200)
    
    print("系统日志生成测试完成！")
    print(f"日志文件位置: logs/system_{datetime.datetime.now().strftime('%Y-%m-%d')}.log")
