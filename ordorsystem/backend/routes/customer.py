from flask import Blueprint, render_template, redirect, url_for, request, flash, session
from flask_login import login_required, current_user
from models.product import Category, Product
from models.order import Order, OrderItem
from app import db

# 创建蓝图
bp = Blueprint('customer', __name__, template_folder='../../frontend/templates')

@bp.route('/home')
def home():
    """顾客端首页"""
    # 获取所有激活的分类
    categories = Category.query.filter_by(is_active=True).order_by(Category.sort_order).all()
    
    # 获取推荐商品（示例：库存大于50的商品）
    recommended_products = Product.query.filter_by(is_available=True)
    recommended_products = recommended_products.filter(Product.stock > 50).order_by(Product.sort_order).limit(8).all()
    
    return render_template('customer/home.html', categories=categories, recommended_products=recommended_products)

@bp.route('/categories/<int:category_id>')
def category_products(category_id):
    """分类商品列表"""
    # 获取指定分类
    category = Category.query.get_or_404(category_id)
    
    # 获取该分类下的所有商品
    products = Product.query.filter_by(category_id=category_id, is_available=True)
    products = products.order_by(Product.sort_order).all()
    
    # 获取所有激活的分类
    categories = Category.query.filter_by(is_active=True).order_by(Category.sort_order).all()
    
    return render_template('customer/category.html', category=category, products=products, categories=categories)

@bp.route('/products/<int:product_id>')
def product_detail(product_id):
    """商品详情页"""
    product = Product.query.get_or_404(product_id)
    
    # 获取相关商品（同一分类下的其他商品）
    related_products = Product.query.filter_by(category_id=product.category_id, is_available=True)
    related_products = related_products.filter(Product.id != product_id).order_by(Product.sort_order).limit(4).all()
    
    return render_template('customer/product_detail.html', product=product, related_products=related_products)

@bp.route('/cart')
def cart():
    """购物车页面"""
    # 从session中获取购物车数据
    cart_items = session.get('cart', {})
    
    # 获取商品详情并计算总价
    products = []
    total_amount = 0
    
    for product_id, quantity in cart_items.items():
        product = Product.query.get(int(product_id))
        if product and product.is_available:
            subtotal = product.price * quantity
            total_amount += subtotal
            products.append({
                'product': product,
                'quantity': quantity,
                'subtotal': subtotal
            })
    
    return render_template('customer/cart.html', products=products, total_amount=total_amount, cart_count=len(products))

@bp.route('/add-to-cart/<int:product_id>', methods=['POST'])
def add_to_cart(product_id):
    """添加商品到购物车"""
    product = Product.query.get_or_404(product_id)
    
    if not product.is_available:
        flash('该商品已下架！', 'danger')
        return redirect(url_for('customer.product_detail', product_id=product_id))
    
    quantity = int(request.form.get('quantity', 1))
    
    # 检查库存
    if product.stock < quantity:
        flash(f'该商品库存不足，当前库存：{product.stock}', 'danger')
        return redirect(url_for('customer.product_detail', product_id=product_id))
    
    # 从session中获取购物车
    cart = session.get('cart', {})
    
    # 更新购物车
    cart[str(product_id)] = cart.get(str(product_id), 0) + quantity
    session['cart'] = cart
    
    flash('商品已添加到购物车！', 'success')
    
    # 重定向到来源页面
    return redirect(request.referrer or url_for('customer.cart'))

@bp.route('/update-cart/<int:product_id>', methods=['POST'])
def update_cart(product_id):
    """更新购物车商品数量"""
    quantity = int(request.form.get('quantity', 0))
    
    if quantity <= 0:
        # 删除商品
        return remove_from_cart(product_id)
    
    # 更新购物车
    cart = session.get('cart', {})
    if str(product_id) in cart:
        cart[str(product_id)] = quantity
        session['cart'] = cart
    
    flash('购物车已更新！', 'success')
    return redirect(url_for('customer.cart'))

@bp.route('/remove-from-cart/<int:product_id>')
def remove_from_cart(product_id):
    """从购物车移除商品"""
    cart = session.get('cart', {})
    if str(product_id) in cart:
        del cart[str(product_id)]
        session['cart'] = cart
    
    flash('商品已从购物车移除！', 'success')
    return redirect(url_for('customer.cart'))

@bp.route('/clear-cart')
def clear_cart():
    """清空购物车"""
    session.pop('cart', None)
    flash('购物车已清空！', 'success')
    return redirect(url_for('customer.cart'))

@bp.route('/checkout', methods=['GET', 'POST'])
@login_required
def checkout():
    """结算页面"""
    # 获取购物车数据
    cart_items = session.get('cart', {})
    
    if not cart_items:
        flash('购物车为空，无法结算！', 'danger')
        return redirect(url_for('customer.home'))
    
    # 获取商品详情并计算总价
    products = []
    total_amount = 0
    
    for product_id, quantity in cart_items.items():
        product = Product.query.get(int(product_id))
        if product and product.is_available:
            if product.stock < quantity:
                flash(f'商品 {product.name} 库存不足，当前库存：{product.stock}', 'danger')
                return redirect(url_for('customer.cart'))
            
            subtotal = product.price * quantity
            total_amount += subtotal
            products.append({
                'product': product,
                'quantity': quantity,
                'subtotal': subtotal
            })
    
    if request.method == 'POST':
        # 创建订单
        table_number = request.form.get('table_number', '')
        notes = request.form.get('notes', '')
        
        # 生成订单号（简单实现，实际应该更复杂）
        from datetime import datetime
        order_no = f'ORD{datetime.now().strftime("%Y%m%d%H%M%S")}{current_user.id}'
        
        # 创建订单
        order = Order(
            order_no=order_no,
            total_amount=total_amount,
            status='pending',
            payment_status='unpaid',
            table_number=table_number,
            notes=notes,
            created_by=current_user.username,
            user_id=current_user.id
        )
        db.session.add(order)
        db.session.flush()  # 获取订单ID但不提交
        
        # 创建订单明细
        for product_id, quantity in cart_items.items():
            product = Product.query.get(int(product_id))
            if product:
                order_item = OrderItem(
                    order_id=order.id,
                    product_id=product.id,
                    product_name=product.name,
                    quantity=quantity,
                    unit_price=product.price,
                    total_price=product.price * quantity
                )
                db.session.add(order_item)
                
                # 减少库存
                product.reduce_stock(quantity)
        
        # 提交事务
        db.session.commit()
        
        # 清空购物车
        session.pop('cart', None)
        
        flash('订单创建成功！请尽快支付', 'success')
        return redirect(url_for('customer.order_detail', order_id=order.id))
    
    return render_template('customer/checkout.html', products=products, total_amount=total_amount)

@bp.route('/orders')
@login_required
def orders():
    """顾客订单列表"""
    # 获取当前用户的所有订单
    orders = Order.query.filter_by(user_id=current_user.id)
    orders = orders.order_by(Order.created_at.desc()).all()
    
    return render_template('customer/orders.html', orders=orders)

@bp.route('/orders/<int:order_id>')
@login_required
def order_detail(order_id):
    """订单详情页"""
    order = Order.query.get_or_404(order_id)
    
    # 检查订单归属
    if order.user_id != current_user.id:
        flash('您无权查看此订单！', 'danger')
        return redirect(url_for('customer.orders'))
    
    return render_template('customer/order_detail.html', order=order)

@bp.route('/profile')
@login_required
def profile():
    """个人中心"""
    return render_template('customer/profile.html', user=current_user)

@bp.route('/profile/edit', methods=['GET', 'POST'])
@login_required
def edit_profile():
    """编辑个人资料"""
    if request.method == 'POST':
        # 更新用户信息
        current_user.nickname = request.form.get('nickname')
        current_user.phone = request.form.get('phone')
        current_user.email = request.form.get('email')
        
        # 保存更改
        current_user.save()
        
        flash('个人资料更新成功！', 'success')
        return redirect(url_for('customer.profile'))
    
    return render_template('customer/edit_profile.html', user=current_user)