# Docker环境下InfluxDB设置指南

## 方法1：通过docker exec进入容器操作

### 1. 查找InfluxDB容器
```bash
# 查看运行中的容器
docker ps | grep influx

# 或者查看所有容器
docker ps -a | grep influx
```

### 2. 进入InfluxDB容器
```bash
# 进入容器（替换container_name为实际容器名）
docker exec -it <container_name> bash

# 或者如果容器名包含influx
docker exec -it $(docker ps --format "table {{.Names}}" | grep influx) bash
```

### 3. 在容器内执行influx命令
```bash
# 在容器内执行
influx setup \
  --username admin \
  --password "your-password" \
  --org "taiwan-ip-scan" \
  --bucket "default" \
  --host http://localhost:8086 \
  --force

# 创建专用buckets
influx bucket create \
  --name "taiwan_ip_segments" \
  --org "taiwan-ip-scan" \
  --retention 720h

influx bucket create \
  --name "taiwan_ip_alive" \
  --org "taiwan-ip-scan" \
  --retention 168h

# 创建API token
influx auth create \
  --org "taiwan-ip-scan" \
  --all-access \
  --description "IP Discovery System"

# 查看创建的token
influx auth list
```

## 方法2：直接通过docker exec执行命令

### 1. 一键设置脚本
```bash
#!/bin/bash
# docker_influx_setup.sh

# 设置变量
CONTAINER_NAME=$(docker ps --format "{{.Names}}" | grep influx | head -1)
ORG_NAME="taiwan-ip-scan"
USERNAME="admin"
PASSWORD="your-password-here"  # 请修改密码

echo "找到InfluxDB容器: $CONTAINER_NAME"

# 初始化InfluxDB
echo "1. 初始化InfluxDB..."
docker exec $CONTAINER_NAME influx setup \
  --username "$USERNAME" \
  --password "$PASSWORD" \
  --org "$ORG_NAME" \
  --bucket "default" \
  --host http://localhost:8086 \
  --force

# 创建专用buckets
echo "2. 创建专用buckets..."
docker exec $CONTAINER_NAME influx bucket create \
  --name "taiwan_ip_segments" \
  --org "$ORG_NAME" \
  --retention 720h

docker exec $CONTAINER_NAME influx bucket create \
  --name "taiwan_ip_alive" \
  --org "$ORG_NAME" \
  --retention 168h

# 创建API token
echo "3. 创建API token..."
TOKEN=$(docker exec $CONTAINER_NAME influx auth create \
  --org "$ORG_NAME" \
  --all-access \
  --description "IP Discovery System - $(date)" \
  --json | jq -r '.token')

echo ""
echo "=== 配置信息 ==="
echo "容器名: $CONTAINER_NAME"
echo "组织名: $ORG_NAME"
echo "API Token: $TOKEN"
echo ""
echo "请将以下信息更新到 config.yaml:"
echo "influxdb:"
echo "  url: \"http://localhost:8086\"  # 或者你的服务器IP"
echo "  token: \"$TOKEN\""
echo "  organization: \"$ORG_NAME\""
echo "  segments_bucket: \"taiwan_ip_segments\""
echo "  alive_bucket: \"taiwan_ip_alive\""
```

### 2. 使用脚本
```bash
# 保存并运行脚本
chmod +x docker_influx_setup.sh
./docker_influx_setup.sh
```

## 方法3：查看现有配置

### 1. 查看现有组织和buckets
```bash
# 获取容器名
CONTAINER_NAME=$(docker ps --format "{{.Names}}" | grep influx | head -1)

# 查看组织
docker exec $CONTAINER_NAME influx org list

# 查看buckets
docker exec $CONTAINER_NAME influx bucket list

# 查看tokens
docker exec $CONTAINER_NAME influx auth list
```

### 2. 单条命令获取所有信息
```bash
CONTAINER_NAME=$(docker ps --format "{{.Names}}" | grep influx | head -1)

echo "=== InfluxDB配置信息 ==="
echo "容器名: $CONTAINER_NAME"
echo ""
echo "组织列表:"
docker exec $CONTAINER_NAME influx org list
echo ""
echo "Bucket列表:"
docker exec $CONTAINER_NAME influx bucket list
echo ""
echo "Token列表:"
docker exec $CONTAINER_NAME influx auth list
```

## 方法4：通过Web界面（推荐）

### 1. 获取容器端口映射
```bash
# 查看端口映射
docker port $(docker ps --format "{{.Names}}" | grep influx | head -1)

# 或者
docker ps | grep influx
```

### 2. 访问Web界面
```bash
# 在浏览器中访问（根据端口映射调整）
http://your-server-ip:8086
# 或者如果端口映射不同
http://your-server-ip:mapped-port
```

## 方法5：Docker Compose环境

### 1. 如果使用docker-compose
```bash
# 进入服务
docker-compose exec influxdb bash

# 或者直接执行命令
docker-compose exec influxdb influx setup \
  --username admin \
  --password "your-password" \
  --org "taiwan-ip-scan" \
  --bucket "default" \
  --force
```

### 2. 完整的docker-compose.yml示例
```yaml
version: '3.8'
services:
  influxdb:
    image: influxdb:2.7
    container_name: influxdb
    ports:
      - "8086:8086"
    volumes:
      - influxdb-data:/var/lib/influxdb2
      - influxdb-config:/etc/influxdb2
    environment:
      - DOCKER_INFLUXDB_INIT_MODE=setup
      - DOCKER_INFLUXDB_INIT_USERNAME=admin
      - DOCKER_INFLUXDB_INIT_PASSWORD=your-password
      - DOCKER_INFLUXDB_INIT_ORG=taiwan-ip-scan
      - DOCKER_INFLUXDB_INIT_BUCKET=default
      - DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=your-initial-token

volumes:
  influxdb-data:
  influxdb-config:
```

## 快速命令参考

### 获取容器名
```bash
INFLUX_CONTAINER=$(docker ps --format "{{.Names}}" | grep influx | head -1)
```

### 创建token
```bash
docker exec $INFLUX_CONTAINER influx auth create \
  --org "taiwan-ip-scan" \
  --all-access \
  --description "IP Discovery"
```

### 查看所有配置
```bash
echo "组织:" && docker exec $INFLUX_CONTAINER influx org list
echo "Buckets:" && docker exec $INFLUX_CONTAINER influx bucket list  
echo "Tokens:" && docker exec $INFLUX_CONTAINER influx auth list
```

## 故障排除

### 1. 容器未运行
```bash
# 启动容器
docker start <container_name>

# 查看日志
docker logs <container_name>
```

### 2. 端口未映射
```bash
# 检查端口映射
docker port <container_name>

# 如果没有映射8086端口，需要重新创建容器
docker run -d --name influxdb -p 8086:8086 influxdb:2.7
```

### 3. 权限问题
```bash
# 检查容器内的权限
docker exec $INFLUX_CONTAINER ls -la /var/lib/influxdb2/
```

## 验证设置

```bash
# 测试连接（从宿主机）
curl -I http://localhost:8086/ping

# 或者从容器内测试
docker exec $INFLUX_CONTAINER influx ping
```