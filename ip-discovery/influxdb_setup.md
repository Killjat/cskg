# InfluxDB 设置指南

## 方法1：Web界面设置（推荐）

### 1. 访问InfluxDB Web界面
```bash
# 在浏览器中访问
http://your-server-ip:8086
```

### 2. 初始化设置
如果是首次访问，会看到初始化页面：

1. **设置初始用户**
   - Username: `admin` (或你喜欢的用户名)
   - Password: `设置一个强密码`
   - Confirm Password: `确认密码`

2. **设置初始组织**
   - Organization Name: `taiwan-ip-scan` (或你喜欢的名称)
   - Bucket Name: `default` (默认bucket，我们会创建自己的)

3. **点击 "Continue"**

### 3. 获取API Token
初始化完成后：

1. 点击左侧菜单的 **"Data"** → **"API Tokens"**
2. 点击 **"Generate API Token"** → **"All Access API Token"**
3. 输入描述：`IP Discovery System`
4. 点击 **"Save"**
5. **复制生成的token**（这个token只显示一次，请保存好）

### 4. 创建专用Buckets
1. 点击左侧菜单的 **"Data"** → **"Buckets"**
2. 点击 **"Create Bucket"**
3. 创建第一个bucket：
   - Name: `taiwan_ip_segments`
   - Retention: `30 days` (或根据需要调整)
   - 点击 **"Create"**
4. 重复步骤创建第二个bucket：
   - Name: `taiwan_ip_alive`
   - Retention: `7 days` (或根据需要调整)

## 方法2：命令行设置

### 1. 检查InfluxDB状态
```bash
# 检查服务状态
sudo systemctl status influxdb

# 如果未启动，启动服务
sudo systemctl start influxdb
sudo systemctl enable influxdb
```

### 2. 使用influx CLI初始化
```bash
# 初始化InfluxDB
influx setup \
  --username admin \
  --password your-password \
  --org taiwan-ip-scan \
  --bucket default \
  --force

# 创建专用buckets
influx bucket create \
  --name taiwan_ip_segments \
  --org taiwan-ip-scan \
  --retention 720h

influx bucket create \
  --name taiwan_ip_alive \
  --org taiwan-ip-scan \
  --retention 168h
```

### 3. 创建API Token
```bash
# 创建All Access token
influx auth create \
  --org taiwan-ip-scan \
  --all-access \
  --description "IP Discovery System"
```

## 方法3：检查现有配置

### 如果InfluxDB已经配置过
```bash
# 查看现有组织
influx org list

# 查看现有buckets
influx bucket list

# 查看现有tokens
influx auth list
```

## 配置文件更新

获得组织名和token后，更新 `config.yaml`：

```yaml
influxdb:
  url: "http://your-server-ip:8086"
  token: "your-actual-token-here"
  organization: "taiwan-ip-scan"
  segments_bucket: "taiwan_ip_segments"
  alive_bucket: "taiwan_ip_alive"
```

## 测试连接

```bash
# 测试配置是否正确
go run main.go test
```

## 常见问题解决

### 1. 端口未开放
```bash
# 检查端口是否监听
sudo netstat -tlnp | grep 8086

# 如果使用防火墙，开放端口
sudo ufw allow 8086
# 或者
sudo firewall-cmd --permanent --add-port=8086/tcp
sudo firewall-cmd --reload
```

### 2. 服务未启动
```bash
# 启动InfluxDB
sudo systemctl start influxdb

# 查看日志
sudo journalctl -u influxdb -f
```

### 3. 权限问题
```bash
# 检查InfluxDB数据目录权限
sudo chown -R influxdb:influxdb /var/lib/influxdb
```

### 4. 配置文件位置
```bash
# InfluxDB配置文件通常在
/etc/influxdb/influxdb.conf
# 或
/etc/influxdb/config.toml
```

## 安全建议

1. **修改默认端口**（可选）
2. **启用HTTPS**（生产环境推荐）
3. **设置强密码**
4. **定期轮换API Token**
5. **限制网络访问**（使用防火墙规则）

## 备份建议

```bash
# 备份InfluxDB数据
influx backup /path/to/backup --org taiwan-ip-scan

# 恢复数据
influx restore /path/to/backup --org taiwan-ip-scan
```