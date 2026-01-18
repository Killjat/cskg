import os
from dotenv import load_dotenv

# 加载环境变量
load_dotenv()

class Config:
    """应用配置类"""
    # 密钥配置
    SECRET_KEY = os.environ.get('SECRET_KEY') or 'dev-secret-key-change-in-production'
    
    # 数据库配置
    BASE_DIR = os.path.abspath(os.path.dirname(__file__))
    DATABASE_PATH = os.path.join(BASE_DIR, '../database/ordersystem.db')
    SQLALCHEMY_DATABASE_URI = f'sqlite:///{DATABASE_PATH}'
    SQLALCHEMY_TRACK_MODIFICATIONS = False
    
    # 文件上传配置
    UPLOAD_FOLDER = os.path.join(BASE_DIR, '../frontend/static/images/products')
    MAX_CONTENT_LENGTH = 16 * 1024 * 1024  # 16MB
    ALLOWED_EXTENSIONS = {'png', 'jpg', 'jpeg', 'gif'}
    
    # 登录配置
    REMEMBER_COOKIE_DURATION = 3600  # 1小时
    SESSION_COOKIE_SECURE = False  # 生产环境设置为True
    SESSION_COOKIE_HTTPONLY = True
    SESSION_COOKIE_SAMESITE = 'Lax'
    
    # 模板配置
    TEMPLATES_AUTO_RELOAD = True
    
    # 其他配置
    DEBUG = True
    TESTING = False

class DevelopmentConfig(Config):
    """开发环境配置"""
    DEBUG = True

class ProductionConfig(Config):
    """生产环境配置"""
    DEBUG = False
    TESTING = False
    SECRET_KEY = os.environ.get('SECRET_KEY') or 'prod-secret-key-must-be-set-in-env'
    SQLALCHEMY_DATABASE_URI = os.environ.get('DATABASE_URL') or Config.SQLALCHEMY_DATABASE_URI
    SESSION_COOKIE_SECURE = True

class TestingConfig(Config):
    """测试环境配置"""
    TESTING = True
    SQLALCHEMY_DATABASE_URI = 'sqlite:///:memory:'
    WTF_CSRF_ENABLED = False

# 配置映射
env_config = {
    'development': DevelopmentConfig,
    'production': ProductionConfig,
    'testing': TestingConfig,
    'default': DevelopmentConfig
}