from datetime import datetime
from app import db

class BaseModel(db.Model):
    """基础模型类，包含通用字段和方法"""
    __abstract__ = True
    
    id = db.Column(db.Integer, primary_key=True, autoincrement=True)
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    updated_at = db.Column(db.DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    def save(self):
        """保存对象到数据库"""
        db.session.add(self)
        db.session.commit()
        return self
    
    def delete(self):
        """从数据库中删除对象"""
        db.session.delete(self)
        db.session.commit()
        return self
    
    def update(self, **kwargs):
        """更新对象属性"""
        for key, value in kwargs.items():
            setattr(self, key, value)
        self.updated_at = datetime.utcnow()
        db.session.commit()
        return self