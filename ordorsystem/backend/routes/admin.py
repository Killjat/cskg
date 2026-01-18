from flask import Blueprint, render_template, redirect, url_for, request, flash
from flask_login import login_required, current_user
from models.product import Category, Product
from models.order import Order, OrderItem
from models.user import User
from app import db
import os

# 创建蓝图
bp = Blueprint('admin', __name__, template_folder='../../frontend/templates')

def is_admin():
    """检查当前用户是否为管理员"""
    return current_user.is_authenticated and current_user.is_admin()

@bp.before_request
def check_admin():
    """管理员权限检查"""
    if not is_admin():
        flash('您无权访问此页面！', 'danger')
        return redirect(url_for('auth.login'))

@bp.route('/dashboard')
def dashboard():
    """管理员仪表盘"""
    # 统计数据
    total_orders = Order.query.count()
    total_products = Product.query.count()
    total_users = User.query.count()
    total_categories = Category.query.count()
    
    # 最近订单
    recent_orders = Order.query.order_by(Order.created_at.desc()).limit(5).all()
    
    return render_template('admin/dashboard.html', 
                         total_orders=total_orders,
                         total_products=total_products,
                         total_users=total_users,
                         total_categories=total_categories,
                         recent_orders=recent_orders)

# ========== 商品分类管理 ==========

@bp.route('/categories')
def categories():
    """分类列表"""
    categories = Category.query.order_by(Category.sort_order).all()
    return render_template('admin/categories.html', categories=categories)

@bp.route('/categories/add', methods=['GET', 'POST'])
def add_category():
    """添加分类"""
    if request.method == 'POST':
        name = request.form['name']
        description = request.form.get('description', '')
        sort_order = int(request.form.get('sort_order', 0))
        
        # 检查分类名称是否已存在
        if Category.query.filter_by(name=name).first():
            flash('分类名称已存在！', 'danger')
            return redirect(url_for('admin.add_category'))
        
        # 创建分类
        category = Category(
            name=name,
            description=description,
            sort_order=sort_order
        )
        category.save()
        
        flash('分类添加成功！', 'success')
        return redirect(url_for('admin.categories'))
    
    return render_template('admin/add_category.html')

@bp.route('/categories/edit/<int:category_id>', methods=['GET', 'POST'])
def edit_category(category_id):
    """编辑分类"""
    category = Category.query.get_or_404(category_id)
    
    if request.method == 'POST':
        name = request.form['name']
        description = request.form.get('description', '')
        sort_order = int(request.form.get('sort_order', 0))
        is_active = request.form.get('is_active') == 'on'
        
        # 检查分类名称是否已存在（排除当前分类）
        existing = Category.query.filter_by(name=name).filter(Category.id != category_id).first()
        if existing:
            flash('分类名称已存在！', 'danger')
            return redirect(url_for('admin.edit_category', category_id=category_id))
        
        # 更新分类
        category.update(
            name=name,
            description=description,
            sort_order=sort_order,
            is_active=is_active
        )
        
        flash('分类更新成功！', 'success')
        return redirect(url_for('admin.categories'))
    
    return render_template('admin/edit_category.html', category=category)

@bp.route('/categories/delete/<int:category_id>')
def delete_category(category_id):
    """删除分类"""
    category = Category.query.get_or_404(category_id)
    
    # 检查该分类下是否有商品
    if category.products:
        flash('该分类下存在商品，无法删除！', 'danger')
        return redirect(url_for('admin.categories'))
    
    # 删除分类
    category.delete()
    flash('分类删除成功！', 'success')
    return redirect(url_for('admin.categories'))

# ========== 商品管理 ==========

@bp.route('/products')
def products():
    """商品列表"""
    # 获取查询参数
    category_id = request.args.get('category_id')
    search = request.args.get('search')
    
    # 构建查询
    query = Product.query
    
    if category_id:
        query = query.filter_by(category_id=category_id)
    
    if search:
        query = query.filter(Product.name.ilike(f'%{search}%'))
    
    # 执行查询
    products = query.order_by(Product.category_id, Product.sort_order).all()
    
    # 获取所有分类
    categories = Category.query.all()
    
    return render_template('admin/products.html', products=products, categories=categories)

@bp.route('/products/add', methods=['GET', 'POST'])
def add_product():
    """添加商品"""
    categories = Category.query.filter_by(is_active=True).all()
    
    if request.method == 'POST':
        name = request.form['name']
        description = request.form.get('description', '')
        price = request.form['price']
        stock = int(request.form['stock'])
        category_id = int(request.form['category_id'])
        sort_order = int(request.form.get('sort_order', 0))
        is_available = request.form.get('is_available') == 'on'
        
        # 检查商品名称是否已存在
        if Product.query.filter_by(name=name).first():
            flash('商品名称已存在！', 'danger')
            return redirect(url_for('admin.add_product'))
        
        # 简单图片处理（暂时不实现文件上传，使用默认图片）
        image = None
        
        # 创建商品
        product = Product(
            name=name,
            description=description,
            price=price,
            stock=stock,
            category_id=category_id,
            image=image,
            sort_order=sort_order,
            is_available=is_available
        )
        product.save()
        
        flash('商品添加成功！', 'success')
        return redirect(url_for('admin.products'))
    
    return render_template('admin/add_product.html', categories=categories)

@bp.route('/products/edit/<int:product_id>', methods=['GET', 'POST'])
def edit_product(product_id):
    """编辑商品"""
    product = Product.query.get_or_404(product_id)
    categories = Category.query.filter_by(is_active=True).all()
    
    if request.method == 'POST':
        name = request.form['name']
        description = request.form.get('description', '')
        price = request.form['price']
        stock = int(request.form['stock'])
        category_id = int(request.form['category_id'])
        sort_order = int(request.form.get('sort_order', 0))
        is_available = request.form.get('is_available') == 'on'
        
        # 检查商品名称是否已存在（排除当前商品）
        existing = Product.query.filter_by(name=name).filter(Product.id != product_id).first()
        if existing:
            flash('商品名称已存在！', 'danger')
            return redirect(url_for('admin.edit_product', product_id=product_id))
        
        # 简单图片处理（暂时不实现文件上传）
        image = product.image
        
        # 更新商品
        product.update(
            name=name,
            description=description,
            price=price,
            stock=stock,
            category_id=category_id,
            image=image,
            sort_order=sort_order,
            is_available=is_available
        )
        
        flash('商品更新成功！', 'success')
        return redirect(url_for('admin.products'))
    
    return render_template('admin/edit_product.html', product=product, categories=categories)

@bp.route('/products/delete/<int:product_id>')
def delete_product(product_id):
    """删除商品"""
    product = Product.query.get_or_404(product_id)
    
    # 检查该商品是否有订单
    if product.order_items:
        flash('该商品已被订单使用，无法删除！', 'danger')
        return redirect(url_for('admin.products'))
    
    # 删除商品图片
    if product.image and os.path.exists(os.path.join(photos.config['UPLOAD_FOLDER'], product.image)):
        os.remove(os.path.join(photos.config['UPLOAD_FOLDER'], product.image))
    
    # 删除商品
    product.delete()
    flash('商品删除成功！', 'success')
    return redirect(url_for('admin.products'))

# ========== 订单管理 ==========

@bp.route('/orders')
def orders():
    """订单列表"""
    # 获取查询参数
    status = request.args.get('status')
    order_no = request.args.get('order_no')
    username = request.args.get('username')
    
    # 构建查询
    query = Order.query
    
    if status:
        query = query.filter_by(status=status)
    
    if order_no:
        query = query.filter(Order.order_no.ilike(f'%{order_no}%'))
    
    if username:
        query = query.join(User).filter(User.username.ilike(f'%{username}%'))
    
    # 执行查询
    orders = query.order_by(Order.created_at.desc()).all()
    
    return render_template('admin/orders.html', orders=orders)

@bp.route('/orders/<int:order_id>')
def order_detail(order_id):
    """订单详情"""
    order = Order.query.get_or_404(order_id)
    return render_template('admin/order_detail.html', order=order)

@bp.route('/orders/update-status/<int:order_id>', methods=['POST'])
def update_order_status(order_id):
    """更新订单状态"""
    order = Order.query.get_or_404(order_id)
    new_status = request.form['status']
    
    # 更新订单状态
    order.update(
        status=new_status,
        processed_by=current_user.username
    )
    
    flash('订单状态更新成功！', 'success')
    return redirect(url_for('admin.order_detail', order_id=order_id))

@bp.route('/orders/delete/<int:order_id>')
def delete_order(order_id):
    """删除订单"""
    order = Order.query.get_or_404(order_id)
    
    # 删除订单
    order.delete()
    flash('订单删除成功！', 'success')
    return redirect(url_for('admin.orders'))

# ========== 用户管理 ==========

@bp.route('/users')
def users():
    """用户列表"""
    # 获取查询参数
    role = request.args.get('role')
    search = request.args.get('search')
    
    # 构建查询
    query = User.query
    
    if role:
        query = query.filter_by(role=role)
    
    if search:
        query = query.filter(User.username.ilike(f'%{search}%') | User.email.ilike(f'%{search}%'))
    
    # 执行查询
    users = query.order_by(User.created_at.desc()).all()
    
    return render_template('admin/users.html', users=users)

@bp.route('/users/edit/<int:user_id>', methods=['GET', 'POST'])
def edit_user(user_id):
    """编辑用户"""
    user = User.query.get_or_404(user_id)
    
    if request.method == 'POST':
        username = request.form['username']
        nickname = request.form.get('nickname', '')
        phone = request.form.get('phone', '')
        email = request.form.get('email', '')
        role = request.form['role']
        is_active = request.form.get('is_active') == 'on'
        points = int(request.form.get('points', 0))
        
        # 检查用户名是否已存在（排除当前用户）
        existing = User.query.filter_by(username=username).filter(User.id != user_id).first()
        if existing:
            flash('用户名已存在！', 'danger')
            return redirect(url_for('admin.edit_user', user_id=user_id))
        
        # 检查邮箱是否已存在（排除当前用户）
        if email:
            existing_email = User.query.filter_by(email=email).filter(User.id != user_id).first()
            if existing_email:
                flash('邮箱已被使用！', 'danger')
                return redirect(url_for('admin.edit_user', user_id=user_id))
        
        # 更新用户
        user.update(
            username=username,
            nickname=nickname,
            phone=phone,
            email=email,
            role=role,
            is_active=is_active,
            points=points
        )
        
        flash('用户信息更新成功！', 'success')
        return redirect(url_for('admin.users'))
    
    return render_template('admin/edit_user.html', user=user)

@bp.route('/users/reset-password/<int:user_id>', methods=['POST'])
def reset_password(user_id):
    """重置用户密码"""
    user = User.query.get_or_404(user_id)
    new_password = request.form['new_password']
    
    # 更新密码
    user.password = new_password
    db.session.commit()
    
    flash('密码重置成功！', 'success')
    return redirect(url_for('admin.edit_user', user_id=user_id))

@bp.route('/users/delete/<int:user_id>')
def delete_user(user_id):
    """删除用户"""
    user = User.query.get_or_404(user_id)
    
    # 不能删除自己
    if user.id == current_user.id:
        flash('不能删除自己的账号！', 'danger')
        return redirect(url_for('admin.users'))
    
    # 检查该用户是否有订单
    if user.orders:
        flash('该用户已产生订单，无法删除！', 'danger')
        return redirect(url_for('admin.users'))
    
    # 删除用户
    user.delete()
    flash('用户删除成功！', 'success')
    return redirect(url_for('admin.users'))

# ========== 系统设置 ==========

@bp.route('/settings')
def settings():
    """系统设置"""
    return render_template('admin/settings.html')

@bp.route('/profile')
def admin_profile():
    """管理员个人资料"""
    return render_template('admin/profile.html', user=current_user)

@bp.route('/profile/edit', methods=['GET', 'POST'])
def edit_admin_profile():
    """编辑管理员个人资料"""
    if request.method == 'POST':
        nickname = request.form.get('nickname')
        phone = request.form.get('phone')
        email = request.form.get('email')
        
        # 更新用户信息
        current_user.update(
            nickname=nickname,
            phone=phone,
            email=email
        )
        
        flash('个人资料更新成功！', 'success')
        return redirect(url_for('admin.admin_profile'))
    
    return render_template('admin/edit_profile.html', user=current_user)