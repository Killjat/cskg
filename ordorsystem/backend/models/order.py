from app import db
from .base import BaseModel

class Order(BaseModel):
    """订单模型"""
    __tablename__ = 'orders'
    
    order_no = db.Column(db.String(50), unique=True, nullable=False, index=True)
    total_amount = db.Column(db.Numeric(10, 2), nullable=False)
    status = db.Column(db.String(20), nullable=False, default='pending')  # pending, paid, preparing, completed, cancelled
    payment_method = db.Column(db.String(20))  # cash, wechat, alipay, card
    payment_status = db.Column(db.String(20), nullable=False, default='unpaid')  # unpaid, paid, failed, refunded
    table_number = db.Column(db.String(20))
    notes = db.Column(db.Text)
    created_by = db.Column(db.String(50))
    processed_by = db.Column(db.String(50))
    
    # 外键
    user_id = db.Column(db.Integer, db.ForeignKey('users.id'), nullable=False)
    
    # 关系定义
    order_items = db.relationship('OrderItem', backref='order', lazy=True, cascade='all, delete-orphan')
    payment = db.relationship('Payment', backref='order', lazy=True, uselist=False, cascade='all, delete-orphan')
    
    def __repr__(self):
        return f'<Order {self.order_no} (${self.total_amount}) - {self.status}>'
    
    def calculate_total(self):
        """计算订单总金额"""
        total = sum(item.total_price for item in self.order_items)
        self.total_amount = total
        db.session.commit()
        return total
    
    def is_paid(self):
        """检查订单是否已支付"""
        return self.payment_status == 'paid'
    
    def is_completed(self):
        """检查订单是否已完成"""
        return self.status == 'completed'
    
    def can_cancel(self):
        """检查订单是否可以取消"""
        return self.status in ['pending', 'paid']

class OrderItem(BaseModel):
    """订单明细模型"""
    __tablename__ = 'order_items'
    
    quantity = db.Column(db.Integer, nullable=False)
    unit_price = db.Column(db.Numeric(10, 2), nullable=False)
    total_price = db.Column(db.Numeric(10, 2), nullable=False)
    product_name = db.Column(db.String(100), nullable=False)
    
    # 外键
    order_id = db.Column(db.Integer, db.ForeignKey('orders.id'), nullable=False)
    product_id = db.Column(db.Integer, db.ForeignKey('products.id'), nullable=False)
    
    def __repr__(self):
        return f'<OrderItem {self.product_name} x {self.quantity}>'
    
    def update_total_price(self):
        """更新明细总价"""
        self.total_price = self.quantity * self.unit_price
        db.session.commit()
        return self.total_price

class Payment(BaseModel):
    """支付记录模型"""
    __tablename__ = 'payments'
    
    payment_no = db.Column(db.String(50), unique=True, nullable=False, index=True)
    amount = db.Column(db.Numeric(10, 2), nullable=False)
    payment_method = db.Column(db.String(20), nullable=False)  # cash, wechat, alipay, card
    payment_status = db.Column(db.String(20), nullable=False, default='pending')  # pending, paid, failed, refunded
    transaction_id = db.Column(db.String(100))
    
    # 外键
    order_id = db.Column(db.Integer, db.ForeignKey('orders.id'), nullable=False, unique=True)
    
    def __repr__(self):
        return f'<Payment {self.payment_no} (${self.amount}) - {self.payment_status}>'
    
    def is_successful(self):
        """检查支付是否成功"""
        return self.payment_status == 'paid'