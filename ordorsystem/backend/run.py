from app import create_app
import os

# 获取环境变量中的配置类型
config_type = os.environ.get('FLASK_CONFIG') or 'default'

# 创建应用实例
app = create_app()

if __name__ == '__main__':
    # 启动应用服务器
    app.run(
        host=os.environ.get('FLASK_HOST') or '0.0.0.0',
        port=int(os.environ.get('FLASK_PORT') or 5000),
        debug=app.config.get('DEBUG', True)
    )