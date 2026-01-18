from app import db
from .base import BaseModel

class Category(BaseModel):
    """商品分类模型"""
    __tablename__ = 'categories'
    
    name = db.Column(db.String(50), unique=True, nullable=False, index=True)
    description = db.Column(db.Text)
    sort_order = db.Column(db.Integer, default=0)
    is_active = db.Column(db.Boolean, default=True)
    
    # 关系定义
    products = db.relationship('Product', backref='category', lazy=True)
    
    def __repr__(self):
        return f'<Category {self.name}>'

class Product(BaseModel):
    """商品模型"""
    __tablename__ = 'products'
    
    name = db.Column(db.String(100), unique=True, nullable=False, index=True)
    description = db.Column(db.Text)
    price = db.Column(db.Numeric(10, 2), nullable=False)
    stock = db.Column(db.Integer, nullable=False, default=0)
    image = db.Column(db.String(255))
    is_available = db.Column(db.Boolean, nullable=False, default=True)
    sort_order = db.Column(db.Integer, default=0)
    
    # 外键
    category_id = db.Column(db.Integer, db.ForeignKey('categories.id'), nullable=False)
    
    # 关系定义
    order_items = db.relationship('OrderItem', backref='product', lazy=True)
    
    def __repr__(self):
        return f'<Product {self.name} (${self.price})>'
    
    def check_stock(self, quantity):
        """检查库存是否充足"""
        return self.stock >= quantity
    
    def reduce_stock(self, quantity):
        """减少库存"""
        if self.check_stock(quantity):
            self.stock -= quantity
            db.session.commit()
            return True
        return False
    
    def add_stock(self, quantity):
        """增加库存"""
        self.stock += quantity
        db.session.commit()
        return True