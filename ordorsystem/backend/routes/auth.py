from flask import Blueprint, render_template, redirect, url_for, request, flash
from flask_login import login_user, logout_user, login_required
from models.user import User

# 创建蓝图
bp = Blueprint('auth', __name__, template_folder='../../frontend/templates')

@bp.route('/login', methods=['GET', 'POST'])
def login():
    """用户登录"""
    if request.method == 'POST':
        username = request.form['username']
        password = request.form['password']
        
        # 查找用户
        user = User.query.filter_by(username=username).first()
        
        if user and user.check_password(password):
            if user.is_active:
                # 登录用户
                login_user(user)
                # 更新最后登录时间
                user.update_last_login()
                flash('登录成功！', 'success')
                
                # 根据用户角色跳转到不同页面
                if user.is_admin():
                    return redirect(url_for('admin.dashboard'))
                else:
                    return redirect(url_for('customer.home'))
            else:
                flash('账号已被禁用！', 'danger')
        else:
            flash('用户名或密码错误！', 'danger')
    
    return render_template('auth/login.html')

@bp.route('/register', methods=['GET', 'POST'])
def register():
    """用户注册"""
    if request.method == 'POST':
        username = request.form['username']
        password = request.form['password']
        confirm_password = request.form['confirm_password']
        nickname = request.form.get('nickname')
        phone = request.form.get('phone')
        email = request.form.get('email')
        
        # 验证密码一致性
        if password != confirm_password:
            flash('两次输入的密码不一致！', 'danger')
            return redirect(url_for('auth.register'))
        
        # 检查用户名是否已存在
        if User.query.filter_by(username=username).first():
            flash('用户名已存在！', 'danger')
            return redirect(url_for('auth.register'))
        
        # 检查邮箱是否已存在
        if email and User.query.filter_by(email=email).first():
            flash('邮箱已被注册！', 'danger')
            return redirect(url_for('auth.register'))
        
        # 创建新用户
        user = User(
            username=username,
            password=password,
            nickname=nickname,
            phone=phone,
            email=email,
            role='customer'
        )
        user.save()
        
        flash('注册成功！请登录', 'success')
        return redirect(url_for('auth.login'))
    
    return render_template('auth/register.html')

@bp.route('/logout')
@login_required
def logout():
    """用户登出"""
    logout_user()
    flash('已成功登出！', 'success')
    return redirect(url_for('auth.login'))

@bp.route('/forgot-password', methods=['GET', 'POST'])
def forgot_password():
    """忘记密码"""
    if request.method == 'POST':
        # 这里可以添加密码重置逻辑
        email = request.form['email']
        # TODO: 实现密码重置功能
        flash('密码重置链接已发送到您的邮箱！', 'info')
        return redirect(url_for('auth.login'))
    
    return render_template('auth/forgot_password.html')