#!/usr/bin/env python3
"""
自动化执行脚本
不需要用户输入，通过读取配置文件执行任务
持续运行，全网寻找目标继续执行
"""

import os
import json
import logging
from datetime import datetime
import threading
import time
import subprocess

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('auto_run.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class AutoRunner:
    def __init__(self):
        """初始化自动化运行器"""
        self.config = self.load_config()
        self.web_dashboard_process = None
        self.is_running = True
        
    def load_config(self):
        """加载配置文件"""
        config_path = 'config.json'
        if os.path.exists(config_path):
            with open(config_path, 'r', encoding='utf-8') as f:
                return json.load(f)
        else:
            logger.error(f"配置文件 {config_path} 不存在，使用默认配置")
            return {
                'self_evolving_wappalyzer': {
                    'rules_file': 'technologies.json',
                    'timeout': 10,
                    'confidence_threshold': 0.7
                },
                'scan_targets': {
                    'sources': {
                        'file': {
                            'enabled': True,
                            'file_path': 'education_sites.json'
                        },
                        'crawler': {
                            'enabled': True
                        }
                    },
                    'filters': {
                        'industry': ['education']
                    }
                }
            }
    
    def start_web_dashboard(self):
        """启动Web仪表盘"""
        logger.info("启动Web仪表盘...")
        try:
            # 使用subprocess启动Web仪表盘，作为后台进程运行
            self.web_dashboard_process = subprocess.Popen(
                ['python3', 'web_dashboard.py'],
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                universal_newlines=True
            )
            logger.info("Web仪表盘已启动，访问地址：http://localhost:5001")
        except Exception as e:
            logger.error(f"启动Web仪表盘失败：{e}")
    
    def stop_web_dashboard(self):
        """停止Web仪表盘"""
        if self.web_dashboard_process:
            logger.info("停止Web仪表盘...")
            self.web_dashboard_process.terminate()
            try:
                self.web_dashboard_process.wait(timeout=5)
                logger.info("Web仪表盘已停止")
            except subprocess.TimeoutExpired:
                self.web_dashboard_process.kill()
                logger.info("Web仪表盘已强制停止")
            self.web_dashboard_process = None
    
    def run_learning_cycle(self):
        """运行一轮学习周期"""
        logger.info("开始新一轮学习周期...")
        
        try:
            # 导入并使用集成系统
            from integrated_system import IntegratedWappalyzerSystem
            
            # 创建集成系统实例
            integrated_system = IntegratedWappalyzerSystem()
            
            # 获取配置参数
            scan_config = self.config.get('scan_targets', {})
            target_count = scan_config.get('sources', {}).get('file', {}).get('limit', 10)
            
            # 运行智能收集和学习
            success = integrated_system.smart_collect_and_learn(
                target_count=target_count,
                min_confidence=self.config['self_evolving_wappalyzer']['confidence_threshold'],
                industry=scan_config.get('filters', {}).get('industry', ['education'])[0] if scan_config.get('filters', {}).get('industry') else None
            )
            
            if success:
                logger.info("学习周期完成")
            else:
                logger.warning("学习周期失败")
                
        except Exception as e:
            logger.error(f"学习周期执行失败：{e}")
    
    def run_continuous_learning(self):
        """持续运行学习流程"""
        logger.info("开始持续学习流程...")
        
        while self.is_running:
            try:
                # 运行一轮学习周期
                self.run_learning_cycle()
                
                # 等待一段时间后继续下一轮
                wait_time = 300  # 默认5分钟
                logger.info(f"等待 {wait_time} 秒后开始下一轮学习...")
                time.sleep(wait_time)
                
            except KeyboardInterrupt:
                logger.info("收到中断信号，停止持续学习")
                self.is_running = False
            except Exception as e:
                logger.error(f"持续学习过程中出错：{e}")
                # 出错后等待一段时间后继续
                time.sleep(60)
    
    def run(self):
        """主运行方法"""
        logger.info("=== 自进化Wappalyzer自动化系统启动 ===")
        logger.info(f"启动时间：{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        
        try:
            # 启动Web仪表盘
            self.start_web_dashboard()
            
            # 启动持续学习线程
            learning_thread = threading.Thread(target=self.run_continuous_learning)
            learning_thread.daemon = True  # 设置为守护线程，主程序退出时自动退出
            learning_thread.start()
            
            # 主线程等待，直到收到中断信号
            logger.info("自动化系统已启动，按 Ctrl+C 停止")
            learning_thread.join()
            
        except KeyboardInterrupt:
            logger.info("收到中断信号，停止自动化系统")
        finally:
            # 停止所有服务
            self.stop_web_dashboard()
            logger.info("=== 自进化Wappalyzer自动化系统停止 ===")

if __name__ == "__main__":
    # 创建并运行自动化系统
    auto_runner = AutoRunner()
    auto_runner.run()
