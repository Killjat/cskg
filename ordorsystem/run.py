import os
import sys

# 添加backend目录到Python路径
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'backend'))

from app import create_app
from config import env_config

# 获取当前环境
env = os.environ.get('FLASK_ENV', 'default')

# 创建应用实例
app = create_app(env_config[env])

if __name__ == '__main__':
    # 启动应用
    app.run(debug=app.config['DEBUG'], host='0.0.0.0', port=5001)
