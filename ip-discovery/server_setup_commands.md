# 服务器上InfluxDB命令行设置

## 1. 检查InfluxDB状态

```bash
# 检查InfluxDB是否运行
sudo systemctl status influxdb

# 如果未运行，启动它
sudo systemctl start influxdb
sudo systemctl enable influxdb

# 检查端口是否监听
sudo netstat -tlnp | grep 8086
```

## 2. 安装influx CLI工具

```bash
# Ubuntu/Debian
wget https://dl.influxdata.com/influxdb/releases/influxdb2-client-2.7.3-linux-amd64.tar.gz
tar xvzf influxdb2-client-2.7.3-linux-amd64.tar.gz
sudo cp influx /usr/local/bin/

# CentOS/RHEL
wget https://dl.influxdata.com/influxdb/releases/influxdb2-client-2.7.3-linux-amd64.tar.gz
tar xvzf influxdb2-client-2.7.3-linux-amd64.tar.gz
sudo cp influx /usr/local/bin/

# 验证安装
influx version
```

## 3. 初始化InfluxDB（如果是全新安装）

```bash
# 初始化InfluxDB
influx setup \
  --username admin \
  --password "your-strong-password" \
  --org "taiwan-ip-scan" \
  --bucket "default" \
  --host http://localhost:8086 \
  --force
```

## 4. 查看现有配置（如果已经配置过）

```bash
# 查看所有组织
influx org list

# 查看所有buckets
influx bucket list

# 查看所有tokens
influx auth list
```

## 5. 创建新的组织和token（如果需要）

```bash
# 创建新组织
influx org create --name "taiwan-ip-scan"

# 创建专用buckets
influx bucket create \
  --name "taiwan_ip_segments" \
  --org "taiwan-ip-scan" \
  --retention 720h

influx bucket create \
  --name "taiwan_ip_alive" \
  --org "taiwan-ip-scan" \
  --retention 168h

# 创建All Access API Token
influx auth create \
  --org "taiwan-ip-scan" \
  --all-access \
  --description "IP Discovery System Token"
```

## 6. 获取现有token信息

```bash
# 列出所有tokens并显示详细信息
influx auth list --json

# 或者更简洁的显示
influx auth list
```

## 7. 一键设置脚本

创建一个自动化脚本：

```bash
#!/bin/bash
# setup_influxdb.sh

echo "=== InfluxDB 自动设置脚本 ==="

# 设置变量
ORG_NAME="taiwan-ip-scan"
USERNAME="admin"
PASSWORD="your-password-here"  # 请修改为你的密码
HOST="http://localhost:8086"

echo "1. 检查InfluxDB状态..."
if ! systemctl is-active --quiet influxdb; then
    echo "启动InfluxDB..."
    sudo systemctl start influxdb
    sleep 5
fi

echo "2. 检查是否已初始化..."
if influx org list &>/dev/null; then
    echo "InfluxDB已初始化"
    echo "现有组织:"
    influx org list
    echo "现有buckets:"
    influx bucket list
    echo "现有tokens:"
    influx auth list
else
    echo "3. 初始化InfluxDB..."
    influx setup \
        --username "$USERNAME" \
        --password "$PASSWORD" \
        --org "$ORG_NAME" \
        --bucket "default" \
        --host "$HOST" \
        --force
fi

echo "4. 创建专用buckets..."
# 检查bucket是否存在，不存在则创建
if ! influx bucket list --name "taiwan_ip_segments" &>/dev/null; then
    influx bucket create \
        --name "taiwan_ip_segments" \
        --org "$ORG_NAME" \
        --retention 720h
    echo "创建了 taiwan_ip_segments bucket"
fi

if ! influx bucket list --name "taiwan_ip_alive" &>/dev/null; then
    influx bucket create \
        --name "taiwan_ip_alive" \
        --org "$ORG_NAME" \
        --retention 168h
    echo "创建了 taiwan_ip_alive bucket"
fi

echo "5. 创建API Token..."
TOKEN=$(influx auth create \
    --org "$ORG_NAME" \
    --all-access \
    --description "IP Discovery System - $(date)" \
    --json | jq -r '.token')

echo ""
echo "=== 配置信息 ==="
echo "组织名: $ORG_NAME"
echo "API Token: $TOKEN"
echo "InfluxDB URL: $HOST"
echo ""
echo "请将以下信息更新到 config.yaml:"
echo "influxdb:"
echo "  url: \"$HOST\""
echo "  token: \"$TOKEN\""
echo "  organization: \"$ORG_NAME\""
echo "  segments_bucket: \"taiwan_ip_segments\""
echo "  alive_bucket: \"taiwan_ip_alive\""
```

## 8. 使用脚本

```bash
# 保存脚本
nano setup_influxdb.sh

# 修改密码
# 将 "your-password-here" 改为你想要的密码

# 给脚本执行权限
chmod +x setup_influxdb.sh

# 运行脚本
./setup_influxdb.sh
```

## 9. 手动获取token的简单方法

```bash
# 如果你只需要快速获取一个token
influx auth create --org "taiwan-ip-scan" --all-access --description "Quick Token"
```

## 10. 验证设置

```bash
# 使用token测试连接
export INFLUX_TOKEN="your-token-here"
export INFLUX_ORG="taiwan-ip-scan"
export INFLUX_HOST="http://localhost:8086"

# 测试连接
influx ping

# 列出buckets验证
influx bucket list
```

## 故障排除

### 如果influx命令不存在
```bash
# 检查是否安装了InfluxDB CLI
which influx

# 如果没有，下载并安装
cd /tmp
wget https://dl.influxdata.com/influxdb/releases/influxdb2-client-2.7.3-linux-amd64.tar.gz
tar xvzf influxdb2-client-2.7.3-linux-amd64.tar.gz
sudo mv influx /usr/local/bin/
```

### 如果端口8086未监听
```bash
# 检查InfluxDB配置
sudo cat /etc/influxdb/config.toml | grep -A5 "\[http\]"

# 重启服务
sudo systemctl restart influxdb
```

### 如果权限被拒绝
```bash
# 检查InfluxDB数据目录权限
sudo chown -R influxdb:influxdb /var/lib/influxdb/
sudo systemctl restart influxdb
```