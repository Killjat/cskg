#!/usr/bin/env python3
from database.db import Database

# 测试数据库连接和表创建
db = Database()

# 连接数据库
if db.connect():
    print("数据库连接成功")
    
    # 创建表
    if db.create_tables():
        print("数据库表创建成功")
    else:
        print("数据库表创建失败")
    
    # 关闭连接
    db.close()
else:
    print("数据库连接失败")
