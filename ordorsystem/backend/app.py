from flask import Flask, redirect, url_for
from flask_sqlalchemy import SQLAlchemy
from flask_login import LoginManager

# 初始化数据库
db = SQLAlchemy()

# 初始化登录管理器
login_manager = LoginManager()
login_manager.login_view = 'auth.login'
login_manager.login_message_category = 'info'

def create_app(config_class=None):
    app = Flask(__name__)
    
    # 加载配置
    if config_class is None:
        app.config.from_pyfile('config.py', silent=True)
    else:
        app.config.from_object(config_class)
    
    # 初始化扩展
    db.init_app(app)
    login_manager.init_app(app)
    
    # 注册蓝图
    try:
        from routes.auth import bp as auth_bp
        app.register_blueprint(auth_bp, url_prefix='/auth')
        
        from routes.customer import bp as customer_bp
        app.register_blueprint(customer_bp, url_prefix='/customer')
        
        from routes.admin import bp as admin_bp
        app.register_blueprint(admin_bp, url_prefix='/admin')
        
        from routes.api import bp as api_bp
        app.register_blueprint(api_bp, url_prefix='/api')
    except ImportError:
        pass
    
    # 主页面路由
    @app.route('/')
    def index():
        return redirect(url_for('customer.home'))
    
    return app

# 导入模型，确保它们被注册到数据库
try:
    from models import user, product, order
except ImportError:
    pass