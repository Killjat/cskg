from datetime import datetime
import bcrypt
from app import db, login_manager
from flask_login import UserMixin
from .base import BaseModel

@login_manager.user_loader
def load_user(user_id):
    """根据用户ID加载用户对象"""
    return User.query.get(int(user_id))

class User(BaseModel, UserMixin):
    """用户模型，包含顾客和管理员"""
    __tablename__ = 'users'
    
    username = db.Column(db.String(50), unique=True, nullable=False, index=True)
    password_hash = db.Column(db.String(255), nullable=False)
    role = db.Column(db.String(20), nullable=False, default='customer')  # customer, admin
    nickname = db.Column(db.String(50))
    phone = db.Column(db.String(20))
    email = db.Column(db.String(100), unique=True, index=True)
    points = db.Column(db.Integer, default=0)
    is_active = db.Column(db.Boolean, default=True)
    last_login_at = db.Column(db.DateTime)
    
    # 关系定义
    orders = db.relationship('Order', backref='user', lazy=True)
    
    @property
    def password(self):
        """密码属性，不可直接读取"""
        raise AttributeError('password is not a readable attribute')
    
    @password.setter
    def password(self, password):
        """密码设置器，自动加密"""
        self.password_hash = bcrypt.hashpw(password.encode('utf-8'), bcrypt.gensalt()).decode('utf-8')
    
    def check_password(self, password):
        """验证密码是否正确"""
        return bcrypt.checkpw(password.encode('utf-8'), self.password_hash.encode('utf-8'))
    
    def is_admin(self):
        """检查用户是否为管理员"""
        return self.role == 'admin'
    
    def is_customer(self):
        """检查用户是否为顾客"""
        return self.role == 'customer'
    
    def update_last_login(self):
        """更新最后登录时间"""
        self.last_login_at = datetime.utcnow()
        db.session.commit()
    
    def __repr__(self):
        return f'<User {self.username} ({self.role})>'