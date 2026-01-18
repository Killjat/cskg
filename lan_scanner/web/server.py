#!/usr/bin/env python3
from flask import Flask, render_template, jsonify, request, Response
import csv
import os
import sys

# 添加父目录到Python路径
sys.path.append(os.path.dirname(os.path.abspath(__file__)) + '/..')

from database.db import Database

app = Flask(__name__, template_folder='templates', static_folder='static')

db = Database()

@app.route('/')
def index():
    """首页路由"""
    return render_template('index.html')

@app.route('/devices')
def devices():
    """设备列表页面"""
    # 确保数据库连接
    db.connect()
    
    # 获取所有设备
    devices = db.get_all_devices()
    
    return render_template('devices.html', devices=devices)

@app.route('/device/<int:device_id>')
def device_detail(device_id):
    """设备详细信息页面"""
    # 确保数据库连接
    db.connect()
    
    # 获取设备信息
    db.cursor.execute("SELECT * FROM devices WHERE id = %s", (device_id,))
    device = db.cursor.fetchone()
    
    # 获取设备的端口信息
    ports = db.get_device_ports(device_id)
    
    return render_template('device_detail.html', device=device, ports=ports)

@app.route('/traffic')
def traffic():
    """流量监控页面"""
    # 确保数据库连接
    db.connect()
    
    # 获取最近的流量信息
    traffic_data = db.get_recent_traffic(limit=1000)
    
    return render_template('traffic.html', traffic=traffic_data)

@app.route('/api/devices')
def api_devices():
    """获取所有设备的API"""
    # 确保数据库连接
    db.connect()
    
    # 获取所有设备
    devices = db.get_all_devices()
    
    # 转换为字典格式
    devices_list = []
    for device in devices:
        devices_list.append({
            'id': device[0],
            'ip': device[1],
            'mac': device[2],
            'hostname': device[3],
            'status': device[4],
            'scan_time': device[5]
        })
    
    return jsonify(devices_list)

@app.route('/api/device/<int:device_id>/ports')
def api_device_ports(device_id):
    """获取设备端口信息的API"""
    # 确保数据库连接
    db.connect()
    
    # 获取设备的端口信息
    ports = db.get_device_ports(device_id)
    
    # 转换为字典格式
    ports_list = []
    for port in ports:
        ports_list.append({
            'id': port[0],
            'device_id': port[1],
            'port': port[2],
            'protocol': port[3],
            'status': port[4],
            'service': port[5],
            'application': port[6],
            'scan_time': port[7]
        })
    
    return jsonify(ports_list)

@app.route('/api/traffic')
def api_traffic():
    """获取流量信息的API"""
    # 确保数据库连接
    db.connect()
    
    # 获取最近的流量信息
    limit = request.args.get('limit', 100, type=int)
    traffic_data = db.get_recent_traffic(limit=limit)
    
    # 转换为字典格式
    traffic_list = []
    for traffic in traffic_data:
        traffic_list.append({
            'id': traffic[0],
            'source_ip': traffic[1],
            'destination_ip': traffic[2],
            'source_port': traffic[3],
            'destination_port': traffic[4],
            'protocol': traffic[5],
            'length': traffic[6],
            'timestamp': traffic[7]
        })
    
    return jsonify(traffic_list)

@app.route('/download/devices.csv')
def download_devices_csv():
    """下载设备信息CSV"""
    # 确保数据库连接
    db.connect()
    
    # 获取所有设备
    devices = db.get_all_devices()
    
    # 生成CSV响应
    def generate():
        csv_writer = csv.writer(sys.stdout)
        # 写入表头
        csv_writer.writerow(['ID', 'IP地址', 'MAC地址', '主机名', '状态', '扫描时间'])
        yield ','.join(['ID', 'IP地址', 'MAC地址', '主机名', '状态', '扫描时间']) + '\n'
        
        # 写入数据
        for device in devices:
            row = [
                str(device[0]),
                device[1],
                device[2] if device[2] else '',
                device[3] if device[3] else '',
                device[4],
                str(device[5])
            ]
            csv_writer.writerow(row)
            yield ','.join(row) + '\n'
    
    return Response(generate(), mimetype='text/csv', headers={
        'Content-Disposition': 'attachment; filename=devices.csv'
    })

@app.route('/download/traffic.csv')
def download_traffic_csv():
    """下载流量信息CSV"""
    # 确保数据库连接
    db.connect()
    
    # 获取最近的流量信息
    traffic_data = db.get_recent_traffic(limit=10000)
    
    # 生成CSV响应
    def generate():
        csv_writer = csv.writer(sys.stdout)
        # 写入表头
        csv_writer.writerow(['ID', '源IP', '目标IP', '源端口', '目标端口', '协议', '长度', '时间戳'])
        yield ','.join(['ID', '源IP', '目标IP', '源端口', '目标端口', '协议', '长度', '时间戳']) + '\n'
        
        # 写入数据
        for traffic in traffic_data:
            row = [
                str(traffic[0]),
                traffic[1],
                traffic[2],
                str(traffic[3]),
                str(traffic[4]),
                traffic[5],
                str(traffic[6]),
                str(traffic[7])
            ]
            csv_writer.writerow(row)
            yield ','.join(row) + '\n'
    
    return Response(generate(), mimetype='text/csv', headers={
        'Content-Disposition': 'attachment; filename=traffic.csv'
    })

@app.route('/download/device/<int:device_id>/ports.csv')
def download_device_ports_csv(device_id):
    """下载设备端口信息CSV"""
    # 确保数据库连接
    db.connect()
    
    # 获取设备信息
    db.cursor.execute("SELECT ip FROM devices WHERE id = %s", (device_id,))
    device = db.cursor.fetchone()
    if not device:
        return "设备不存在", 404
    
    # 获取设备的端口信息
    ports = db.get_device_ports(device_id)
    
    # 生成CSV响应
    def generate():
        csv_writer = csv.writer(sys.stdout)
        # 写入表头
        csv_writer.writerow(['ID', '设备ID', '端口', '协议', '状态', '服务', '应用', '扫描时间'])
        yield ','.join(['ID', '设备ID', '端口', '协议', '状态', '服务', '应用', '扫描时间']) + '\n'
        
        # 写入数据
        for port in ports:
            row = [
                str(port[0]),
                str(port[1]),
                str(port[2]),
                port[3],
                port[4],
                port[5] if port[5] else '',
                port[6] if port[6] else '',
                str(port[7])
            ]
            csv_writer.writerow(row)
            yield ','.join(row) + '\n'
    
    return Response(generate(), mimetype='text/csv', headers={
        'Content-Disposition': f'attachment; filename={device[0]}_ports.csv'
    })

if __name__ == '__main__':
    print("🚀 局域网扫描WEB服务启动")
    print("📡 访问地址: http://localhost:5000")
    app.run(debug=True, host='0.0.0.0', port=5000)
