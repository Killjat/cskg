from app import create_app, db
from config import Config
from models.user import User
from models.product import Category, Product
from models.order import Order, OrderItem, Payment
import random

# 创建应用实例并明确加载配置
app = create_app(Config)

with app.app_context():
    # 删除所有表（如果存在）
    db.drop_all()
    
    # 创建所有表
    db.create_all()
    
    print('Database tables created successfully!')
    
    # 创建初始管理员用户
    admin = User(
        username='admin',
        password='admin123',
        role='admin',
        nickname='管理员',
        email='admin@example.com'
    )
    admin.save()
    print(f'Created admin user: {admin.username}')
    
    # 创建初始顾客用户
    customer = User(
        username='customer',
        password='customer123',
        role='customer',
        nickname='测试顾客',
        phone='13800138000',
        email='customer@example.com'
    )
    customer.save()
    print(f'Created customer user: {customer.username}')
    
    # 创建商品分类
    categories = [
        {'name': '饮料', 'description': '各种饮品'},
        {'name': '零食', 'description': '各种小吃'},
        {'name': '快餐', 'description': '快捷餐品'},
        {'name': '其他', 'description': '其他商品'}
    ]
    
    created_categories = []
    for cat_data in categories:
        category = Category(**cat_data)
        category.save()
        created_categories.append(category)
        print(f'Created category: {category.name}')
    
    # 创建示例商品
    products_data = [
        # 饮料类
        {'name': '可乐', 'description': '瓶装可乐500ml', 'price': 3.50, 'stock': 100, 'category_id': 1, 'is_available': True},
        {'name': '雪碧', 'description': '瓶装雪碧500ml', 'price': 3.50, 'stock': 80, 'category_id': 1, 'is_available': True},
        {'name': '芬达', 'description': '瓶装芬达500ml', 'price': 3.50, 'stock': 60, 'category_id': 1, 'is_available': True},
        {'name': '矿泉水', 'description': '瓶装矿泉水550ml', 'price': 2.00, 'stock': 200, 'category_id': 1, 'is_available': True},
        {'name': '绿茶', 'description': '瓶装绿茶500ml', 'price': 4.00, 'stock': 120, 'category_id': 1, 'is_available': True},
        
        # 零食类
        {'name': '薯片', 'description': '原味薯片100g', 'price': 5.00, 'stock': 150, 'category_id': 2, 'is_available': True},
        {'name': '花生', 'description': '椒盐花生200g', 'price': 6.00, 'stock': 90, 'category_id': 2, 'is_available': True},
        {'name': '饼干', 'description': '奶油饼干250g', 'price': 8.00, 'stock': 70, 'category_id': 2, 'is_available': True},
        {'name': '巧克力', 'description': '牛奶巧克力100g', 'price': 12.00, 'stock': 50, 'category_id': 2, 'is_available': True},
        {'name': '糖果', 'description': '水果糖果500g', 'price': 10.00, 'stock': 180, 'category_id': 2, 'is_available': True},
        
        # 快餐类
        {'name': '汉堡', 'description': '牛肉汉堡', 'price': 15.00, 'stock': 40, 'category_id': 3, 'is_available': True},
        {'name': '炸鸡', 'description': '香辣炸鸡', 'price': 12.00, 'stock': 30, 'category_id': 3, 'is_available': True},
        {'name': '薯条', 'description': '炸薯条', 'price': 8.00, 'stock': 50, 'category_id': 3, 'is_available': True},
        {'name': '炒饭', 'description': '扬州炒饭', 'price': 18.00, 'stock': 25, 'category_id': 3, 'is_available': True},
        {'name': '炒面', 'description': '鸡蛋炒面', 'price': 16.00, 'stock': 35, 'category_id': 3, 'is_available': True}
    ]
    
    for prod_data in products_data:
        product = Product(**prod_data)
        product.save()
        print(f'Created product: {product.name} - ${product.price}')
    
    print('\nInitial data created successfully!')
    print('=' * 50)
    print('Database initialized. You can now run the application.')
    print('Use the following credentials to login:')
    print('Admin: username=admin, password=admin123')
    print('Customer: username=customer, password=customer123')
    print('=' * 50)