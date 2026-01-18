from flask import Blueprint, jsonify, request
from flask_login import login_required, current_user
from models.product import Category, Product
from models.order import Order, OrderItem
from models.user import User
from app import db

# 创建蓝图
bp = Blueprint('api', __name__)

# ========== 公共API ==========

@bp.route('/categories')
def api_categories():
    """获取所有分类"""
    categories = Category.query.filter_by(is_active=True).order_by(Category.sort_order).all()
    result = []
    for category in categories:
        result.append({
            'id': category.id,
            'name': category.name,
            'description': category.description,
            'sort_order': category.sort_order
        })
    return jsonify({
        'code': 200,
        'message': 'success',
        'data': result
    })

@bp.route('/products')
def api_products():
    """获取所有商品"""
    # 获取查询参数
    category_id = request.args.get('category_id')
    
    # 构建查询
    query = Product.query.filter_by(is_available=True)
    
    if category_id:
        query = query.filter_by(category_id=category_id)
    
    # 执行查询
    products = query.order_by(Product.sort_order).all()
    
    result = []
    for product in products:
        result.append({
            'id': product.id,
            'name': product.name,
            'description': product.description,
            'price': str(product.price),
            'stock': product.stock,
            'image': product.image,
            'category_id': product.category_id,
            'sort_order': product.sort_order
        })
    
    return jsonify({
        'code': 200,
        'message': 'success',
        'data': result
    })

@bp.route('/products/<int:product_id>')
def api_product_detail(product_id):
    """获取商品详情"""
    product = Product.query.get_or_404(product_id)
    
    if not product.is_available:
        return jsonify({
            'code': 404,
            'message': '商品不存在或已下架'
        })
    
    result = {
        'id': product.id,
        'name': product.name,
        'description': product.description,
        'price': str(product.price),
        'stock': product.stock,
        'image': product.image,
        'category_id': product.category_id,
        'category_name': product.category.name,
        'sort_order': product.sort_order
    }
    
    return jsonify({
        'code': 200,
        'message': 'success',
        'data': result
    })

# ========== 顾客API ==========

@bp.route('/customer/orders')
@login_required
def api_customer_orders():
    """获取当前顾客的订单"""
    if not current_user.is_customer():
        return jsonify({
            'code': 403,
            'message': '无权访问'
        })
    
    orders = Order.query.filter_by(user_id=current_user.id)
    orders = orders.order_by(Order.created_at.desc()).all()
    
    result = []
    for order in orders:
        order_items = []
        for item in order.order_items:
            order_items.append({
                'product_id': item.product_id,
                'product_name': item.product_name,
                'quantity': item.quantity,
                'unit_price': str(item.unit_price),
                'total_price': str(item.total_price)
            })
        
        result.append({
            'id': order.id,
            'order_no': order.order_no,
            'total_amount': str(order.total_amount),
            'status': order.status,
            'payment_status': order.payment_status,
            'payment_method': order.payment_method,
            'table_number': order.table_number,
            'notes': order.notes,
            'created_at': order.created_at.strftime('%Y-%m-%d %H:%M:%S'),
            'updated_at': order.updated_at.strftime('%Y-%m-%d %H:%M:%S'),
            'items': order_items
        })
    
    return jsonify({
        'code': 200,
        'message': 'success',
        'data': result
    })

@bp.route('/customer/orders/<int:order_id>')
@login_required
def api_customer_order_detail(order_id):
    """获取当前顾客的订单详情"""
    if not current_user.is_customer():
        return jsonify({
            'code': 403,
            'message': '无权访问'
        })
    
    order = Order.query.get_or_404(order_id)
    
    if order.user_id != current_user.id:
        return jsonify({
            'code': 403,
            'message': '无权访问此订单'
        })
    
    # 构建订单详情
    order_items = []
    for item in order.order_items:
        order_items.append({
            'product_id': item.product_id,
            'product_name': item.product_name,
            'quantity': item.quantity,
            'unit_price': str(item.unit_price),
            'total_price': str(item.total_price)
        })
    
    result = {
        'id': order.id,
        'order_no': order.order_no,
        'total_amount': str(order.total_amount),
        'status': order.status,
        'payment_status': order.payment_status,
        'payment_method': order.payment_method,
        'table_number': order.table_number,
        'notes': order.notes,
        'created_at': order.created_at.strftime('%Y-%m-%d %H:%M:%S'),
        'updated_at': order.updated_at.strftime('%Y-%m-%d %H:%M:%S'),
        'items': order_items
    }
    
    return jsonify({
        'code': 200,
        'message': 'success',
        'data': result
    })

@bp.route('/customer/profile')
@login_required
def api_customer_profile():
    """获取当前顾客的个人信息"""
    if not current_user.is_customer():
        return jsonify({
            'code': 403,
            'message': '无权访问'
        })
    
    result = {
        'id': current_user.id,
        'username': current_user.username,
        'nickname': current_user.nickname,
        'phone': current_user.phone,
        'email': current_user.email,
        'points': current_user.points,
        'last_login_at': current_user.last_login_at.strftime('%Y-%m-%d %H:%M:%S') if current_user.last_login_at else None
    }
    
    return jsonify({
        'code': 200,
        'message': 'success',
        'data': result
    })

# ========== 管理员API ==========

@bp.route('/admin/dashboard')
@login_required
def api_admin_dashboard():
    """获取管理员仪表盘数据"""
    if not current_user.is_admin():
        return jsonify({
            'code': 403,
            'message': '无权访问'
        })
    
    # 统计数据
    total_orders = Order.query.count()
    total_products = Product.query.count()
    total_users = User.query.count()
    total_categories = Category.query.count()
    
    return jsonify({
        'code': 200,
        'message': 'success',
        'data': {
            'total_orders': total_orders,
            'total_products': total_products,
            'total_users': total_users,
            'total_categories': total_categories
        }
    })

@bp.route('/admin/orders')
@login_required
def api_admin_orders():
    """获取所有订单"""
    if not current_user.is_admin():
        return jsonify({
            'code': 403,
            'message': '无权访问'
        })
    
    # 获取查询参数
    status = request.args.get('status')
    
    # 构建查询
    query = Order.query
    
    if status:
        query = query.filter_by(status=status)
    
    # 执行查询
    orders = query.order_by(Order.created_at.desc()).limit(20).all()
    
    result = []
    for order in orders:
        result.append({
            'id': order.id,
            'order_no': order.order_no,
            'user_id': order.user_id,
            'username': order.user.username,
            'total_amount': str(order.total_amount),
            'status': order.status,
            'payment_status': order.payment_status,
            'created_at': order.created_at.strftime('%Y-%m-%d %H:%M:%S')
        })
    
    return jsonify({
        'code': 200,
        'message': 'success',
        'data': result
    })

@bp.route('/admin/orders/<int:order_id>/status', methods=['PUT'])
@login_required
def api_update_order_status(order_id):
    """更新订单状态"""
    if not current_user.is_admin():
        return jsonify({
            'code': 403,
            'message': '无权访问'
        })
    
    order = Order.query.get_or_404(order_id)
    data = request.get_json()
    new_status = data.get('status')
    
    if not new_status:
        return jsonify({
            'code': 400,
            'message': '缺少状态参数'
        })
    
    # 更新订单状态
    order.update(
        status=new_status,
        processed_by=current_user.username
    )
    
    return jsonify({
        'code': 200,
        'message': '订单状态更新成功',
        'data': {
            'order_id': order.id,
            'status': order.status
        }
    })

# ========== 错误处理 ==========

@bp.errorhandler(404)
def not_found(error):
    """404错误处理"""
    return jsonify({
        'code': 404,
        'message': '资源不存在'
    })

@bp.errorhandler(500)
def internal_error(error):
    """500错误处理"""
    return jsonify({
        'code': 500,
        'message': '服务器内部错误'
    })

@bp.errorhandler(401)
def unauthorized(error):
    """401错误处理"""
    return jsonify({
        'code': 401,
        'message': '未授权访问'
    })

@bp.errorhandler(403)
def forbidden(error):
    """403错误处理"""
    return jsonify({
        'code': 403,
        'message': '禁止访问'
    })