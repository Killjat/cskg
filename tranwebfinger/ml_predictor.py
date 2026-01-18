#!/usr/bin/env python3
import json
import os
import logging
import numpy as np
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.ensemble import RandomForestClassifier
from sklearn.model_selection import train_test_split
from sklearn.metrics import accuracy_score, classification_report
import pickle

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('ml_predictor.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class MLTechnologyPredictor:
    def __init__(self, model_path='tech_predictor_model.pkl', vectorizer_path='vectorizer.pkl'):
        self.model_path = model_path
        self.vectorizer_path = vectorizer_path
        self.model = None
        self.vectorizer = None
        self.load_model()
    
    def load_model(self):
        """加载已训练的模型和向量器"""
        try:
            if os.path.exists(self.model_path) and os.path.exists(self.vectorizer_path):
                with open(self.model_path, 'rb') as f:
                    self.model = pickle.load(f)
                with open(self.vectorizer_path, 'rb') as f:
                    self.vectorizer = pickle.load(f)
                logger.info("成功加载已训练的模型和向量器")
                return True
            else:
                logger.info("未找到已训练的模型，将使用默认模型")
                self.create_default_model()
                return False
        except Exception as e:
            logger.error(f"加载模型时出错：{e}")
            self.create_default_model()
            return False
    
    def create_default_model(self):
        """创建默认模型"""
        logger.info("创建默认随机森林模型")
        # 创建默认模型和向量器
        self.model = RandomForestClassifier(n_estimators=100, random_state=42)
        self.vectorizer = TfidfVectorizer(max_features=1000, ngram_range=(1, 2))
        
        # 使用示例数据训练默认模型
        self.train_default_model()
    
    def train_default_model(self):
        """使用示例数据训练默认模型"""
        logger.info("使用示例数据训练默认模型")
        
        # 示例数据：网站特征和对应的技术
        sample_data = [
            # WordPress网站特征
            ("wordpress wp-content wp-includes generator:WordPress", "WordPress"),
            ("wp-content plugins themes wp-json", "WordPress"),
            
            # React网站特征
            ("react react-dom root create-react-app", "React"),
            ("react hooks useState useEffect", "React"),
            
            # Nginx网站特征
            ("nginx server nginx/1.20.0", "Nginx"),
            ("server nginx x-powered-by", "Nginx"),
            
            # BFE网站特征
            ("bfe server bfe/1.0.0", "BFE"),
            ("server bfe", "BFE"),
            
            # 百度网站特征
            ("baidu bdorz baidu.com", "Baidu"),
            ("bdstatic.com baidu baidustatic", "Baidu")
        ]
        
        # 准备训练数据
        X = [features for features, _ in sample_data]
        y = [tech for _, tech in sample_data]
        
        # 训练向量器
        self.vectorizer.fit(X)
        X_vectorized = self.vectorizer.transform(X)
        
        # 训练模型
        self.model.fit(X_vectorized, y)
        
        # 保存模型
        self.save_model()
        
        # 评估模型
        y_pred = self.model.predict(X_vectorized)
        accuracy = accuracy_score(y, y_pred)
        logger.info(f"默认模型训练完成，准确率：{accuracy:.2f}")
    
    def save_model(self):
        """保存模型和向量器"""
        try:
            with open(self.model_path, 'wb') as f:
                pickle.dump(self.model, f)
            with open(self.vectorizer_path, 'wb') as f:
                pickle.dump(self.vectorizer, f)
            logger.info("模型和向量器已保存")
            return True
        except Exception as e:
            logger.error(f"保存模型时出错：{e}")
            return False
    
    def extract_features(self, website_data):
        """从网站数据中提取特征"""
        logger.info("开始提取网站特征")
        
        features = []
        
        # 提取响应头特征
        if 'headers' in website_data:
            for header, value in website_data['headers'].items():
                features.append(f"{header.lower()}:{value.lower()}")
        
        # 提取HTML特征
        if 'html' in website_data:
            html = website_data['html'].lower()
            
            # 提取脚本标签内容
            import re
            scripts = re.findall(r'<script[^>]*>(.*?)</script>', html, re.DOTALL)
            for script in scripts[:10]:  # 只取前10个脚本
                features.extend(script.split()[:50])  # 每个脚本只取前50个词
            
            # 提取链接标签
            links = re.findall(r'<link[^>]*>', html)
            for link in links[:10]:
                features.extend(link.split()[:10])
            
            # 提取meta标签
            metas = re.findall(r'<meta[^>]*>', html)
            for meta in metas[:10]:
                features.extend(meta.split()[:10])
        
        # 提取URL特征
        if 'url' in website_data:
            url = website_data['url'].lower()
            features.extend(url.split('/'))
        
        # 清理特征
        cleaned_features = []
        for feature in features:
            # 移除特殊字符
            feature = re.sub(r'[^a-zA-Z0-9_-]', ' ', feature)
            # 移除多余空格
            feature = ' '.join(feature.split())
            if feature and len(feature) > 2:  # 只保留长度大于2的特征
                cleaned_features.append(feature)
        
        # 合并特征为字符串
        feature_str = ' '.join(cleaned_features)
        logger.debug(f"提取的特征：{feature_str[:100]}...")
        
        return feature_str
    
    def predict_technology(self, website_data):
        """预测网站使用的技术"""
        logger.info("开始预测网站使用的技术")
        
        try:
            # 提取特征
            feature_str = self.extract_features(website_data)
            
            # 向量化特征
            feature_vector = self.vectorizer.transform([feature_str])
            
            # 预测技术
            prediction = self.model.predict(feature_vector)
            
            # 获取预测概率
            probabilities = self.model.predict_proba(feature_vector)[0]
            max_prob = max(probabilities)
            
            logger.info(f"预测结果：{prediction[0]} (置信度：{max_prob:.2f})")
            
            return {
                'technology': prediction[0],
                'confidence': max_prob
            }
            
        except Exception as e:
            logger.error(f"预测时出错：{e}")
            return {
                'technology': None,
                'confidence': 0.0
            }
    
    def batch_predict(self, websites_data):
        """批量预测多个网站使用的技术"""
        logger.info(f"开始批量预测 {len(websites_data)} 个网站")
        results = []
        
        for website_data in websites_data:
            result = self.predict_technology(website_data)
            results.append(result)
        
        logger.info(f"批量预测完成，共预测 {len(results)} 个网站")
        return results
    
    def update_model(self, new_data):
        """使用新数据更新模型"""
        logger.info(f"使用 {len(new_data)} 条新数据更新模型")
        
        if not new_data:
            logger.warning("没有新数据用于更新模型")
            return False
        
        try:
            # 准备新数据
            X_new = [data['features'] for data in new_data]
            y_new = [data['technology'] for data in new_data]
            
            # 向量化新特征
            X_new_vectorized = self.vectorizer.transform(X_new)
            
            # 增量训练模型
            self.model.fit(X_new_vectorized, y_new)
            
            # 保存更新后的模型
            self.save_model()
            
            logger.info("模型更新完成")
            return True
        except Exception as e:
            logger.error(f"更新模型时出错：{e}")
            return False

if __name__ == "__main__":
    # 创建预测器实例
    predictor = MLTechnologyPredictor()
    
    # 示例网站数据
    sample_website = {
        'url': 'https://www.wordpress.org',
        'headers': {
            'Server': 'nginx',
            'X-Powered-By': 'PHP/7.4.33',
            'Link': '</wp-content/themes/twentytwentyone/style.css>; rel=preload; as=style'
        },
        'html': '<meta name="generator" content="WordPress 6.1.1">\n<link rel="stylesheet" href="/wp-content/themes/twentytwentyone/style.css">'
    }
    
    # 预测技术
    result = predictor.predict_technology(sample_website)
    print(f"预测结果：{result['technology']} (置信度：{result['confidence']:.2f})")
