#!/usr/bin/env python3
import json
import os
import logging
from datetime import datetime

# 导入各个组件
from main import SelfEvolvingWappalyzer
from smart_targets import SmartTargetGenerator
from ml_predictor import MLTechnologyPredictor

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('integrated_system.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class IntegratedWappalyzerSystem:
    def __init__(self):
        # 初始化各个组件
        logger.info("初始化集成Wappalyzer系统...")
        
        # 初始化主系统
        self.wappalyzer = SelfEvolvingWappalyzer()
        
        # 初始化智能目标生成器
        self.target_generator = SmartTargetGenerator({
            'max_targets': 50,
            'validation_timeout': 5,
            'sources': {
                'alexa': True
            },
            'filters': {
                'tld': [],
                'min_estimated_visits': 0
            }
        })
        
        # 初始化机器学习预测器
        self.ml_predictor = self.wappalyzer.ml_predictor
        
        logger.info("集成Wappalyzer系统初始化完成")
    
    def smart_collect_and_learn(self, target_count=20, min_confidence=0.7, tech_category=None, industry=None):
        """智能收集网站数据并进行学习"""
        logger.info(f"开始智能收集和学习流程，目标数量：{target_count}")
        
        # 1. 智能获取真实网站
        logger.info("步骤1：智能获取真实网站...")
        
        targets = []
        if tech_category:
            # 根据技术类别获取网站，支持行业筛选
            targets = self.target_generator.get_websites_by_tech_category(tech_category, target_count, industry)
        elif industry:
            # 根据行业获取网站
            targets = self.target_generator.get_industry_websites(industry, target_count)
        else:
            # 常规获取网站
            targets = self.target_generator.generate_targets(target_count)
            
        if not targets:
            logger.error("未能获取到任何有效网站，流程终止")
            return False
        logger.info(f"成功获取 {len(targets)} 个有效网站")
        
        # 2. 扫描网站并收集数据
        logger.info(f"步骤2：扫描 {len(targets)} 个网站...")
        training_data = []
        detected_techs = {}
        
        for url in targets:
            logger.info(f"扫描网站：{url}")
            
            try:
                # 扫描网站
                result = self.wappalyzer.scan(url)
                
                # 提取特征用于训练
                from urllib.parse import urlparse
                import requests
                response = requests.get(url, timeout=5)
                html = response.text[:10000]  # 截断HTML
                
                website_data = {
                    'url': url,
                    'headers': dict(response.headers),
                    'html': html
                }
                
                # 提取特征
                feature_str = self.ml_predictor.extract_features(website_data)
                
                # 如果有检测到的技术，添加到训练数据
                if result:
                    for tech_name, tech_info in result.items():
                        training_data.append({
                            'features': feature_str,
                            'technology': tech_name
                        })
                        
                        # 统计检测到的技术
                        if tech_name not in detected_techs:
                            detected_techs[tech_name] = 0
                        detected_techs[tech_name] += 1
                
            except Exception as e:
                logger.error(f"处理网站 {url} 时出错：{e}")
                continue
        
        logger.info(f"扫描完成，收集到 {len(training_data)} 条训练数据")
        logger.info(f"检测到的技术分布：{detected_techs}")
        
        # 3. 训练机器学习模型
        if training_data:
            logger.info(f"步骤3：使用 {len(training_data)} 条数据训练机器学习模型...")
            if self.ml_predictor.update_model(training_data):
                logger.info("机器学习模型训练成功")
            else:
                logger.error("机器学习模型训练失败")
        else:
            logger.warning("没有足够的训练数据，跳过模型训练")
        
        # 4. 更新规则
        logger.info("步骤4：更新检测规则...")
        
        # 基于检测结果和机器学习预测更新规则
        for url in targets[:5]:  # 只使用前5个网站进行规则更新
            logger.info(f"基于网站 {url} 更新规则...")
            try:
                response = requests.get(url, timeout=5)
                html = response.text
                
                # 提取可能的技术线索
                expected_techs = {}
                
                # 1. 从响应头提取
                for header, value in response.headers.items():
                    if 'server' in header.lower():
                        server = value.lower()
                        if 'nginx' in server:
                            expected_techs['Nginx'] = {'name': 'Nginx', 'category': 'Web Servers'}
                        elif 'bfe' in server:
                            expected_techs['BFE'] = {'name': 'BFE', 'category': 'Web Servers'}
                
                # 2. 从HTML提取
                if 'wordpress' in html.lower():
                    expected_techs['WordPress'] = {'name': 'WordPress', 'category': 'CMS'}
                if 'react' in html.lower():
                    expected_techs['React'] = {'name': 'React', 'category': 'JavaScript Frameworks'}
                if 'baidu' in html.lower() or 'bdorz' in html.lower():
                    expected_techs['Baidu'] = {'name': 'Baidu', 'category': 'Search Engines'}
                
                # 3. 使用机器学习预测结果
                website_data = {
                    'url': url,
                    'headers': dict(response.headers),
                    'html': html[:10000]
                }
                ml_result = self.ml_predictor.predict_technology(website_data)
                if ml_result['technology'] and ml_result['confidence'] > min_confidence:
                    expected_techs[ml_result['technology']] = {
                        'name': ml_result['technology'],
                        'category': 'Unknown',
                        'ml_detected': True
                    }
                
                # 更新规则
                if expected_techs:
                    logger.info(f"基于 {url} 发现 {len(expected_techs)} 种可能的技术，尝试更新规则")
                    self.wappalyzer.evolve(url, expected_techs)
                
            except Exception as e:
                logger.error(f"更新网站 {url} 的规则时出错：{e}")
                continue
        
        # 5. 保存最终规则
        logger.info("步骤5：保存最终规则...")
        self.wappalyzer.save_rules()
        logger.info("规则已保存")
        
        # 6. 显示最终统计
        logger.info("步骤6：显示最终统计信息...")
        self.wappalyzer.show_stats()
        
        logger.info("智能收集和学习流程完成")
        return True
    
    def run_continuous_learning(self, iterations=5, target_count=10, tech_category=None, industry=None):
        """连续运行学习流程"""
        logger.info(f"开始连续学习流程，共 {iterations} 轮，每轮 {target_count} 个网站")
        
        for i in range(iterations):
            logger.info(f"=== 第 {i+1}/{iterations} 轮学习 ===")
            self.smart_collect_and_learn(target_count, tech_category=tech_category, industry=industry)
        
        logger.info("连续学习流程完成")
        return True
    
    def cms_fingerprint_learning(self, target_count=30, iterations=3, industry=None):
        """CMS指纹学习功能"""
        logger.info("=== 开始CMS指纹学习 ===")
        
        if industry:
            logger.info(f"针对{industry}行业的CMS网站进行学习")
        
        # 连续学习CMS网站，支持行业筛选
        self.run_continuous_learning(iterations, target_count, tech_category='CMS', industry=industry)
        
        # 专门针对CMS技术更新规则
        logger.info("=== 专门针对CMS技术更新规则 ===")
        
        # 加载当前规则
        from main import SelfEvolvingWappalyzer
        cms_wappalyzer = SelfEvolvingWappalyzer()
        
        # 打印CMS相关技术统计
        cms_techs = {}
        for tech_name, tech_info in cms_wappalyzer.technologies.items():
            if tech_info['category'] == 'CMS':
                cms_techs[tech_name] = tech_info
        
        logger.info(f"当前规则库中包含 {len(cms_techs)} 种CMS技术")
        for tech_name, tech_info in cms_techs.items():
            logger.info(f"- {tech_name}: {tech_info['description']}")
        
        logger.info("=== CMS指纹学习完成 ===")
        return True

if __name__ == "__main__":
    # 创建集成系统实例
    integrated_system = IntegratedWappalyzerSystem()
    
    # 运行连续学习模式
    # 参数1：学习轮数
    # 参数2：每轮获取的网站数量
    integrated_system.run_continuous_learning(3, 5)
    
    # 可选：只运行一次
    # integrated_system.smart_collect_and_learn(10)
