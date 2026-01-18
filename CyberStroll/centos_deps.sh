#!/bin/bash

# CentOS 依赖环境部署脚本 - Zookeeper、Kafka、Elasticsearch

set -e

echo "====================================="
echo "CentOS 依赖环境部署脚本"
echo "部署组件: Zookeeper、Kafka、Elasticsearch"
echo "====================================="

# 1. 系统更新与基础依赖安装
echo "\n1. 更新系统并安装基础依赖..."
# 清理无效的zookeeper仓库配置（如果存在）
if [ -f /etc/yum.repos.d/zookeeper.repo ]; then
    rm -f /etc/yum.repos.d/zookeeper.repo
fi
# 清理yum缓存
yum clean all
yum update -y
yum install -y wget java-11-openjdk java-11-openjdk-devel curl

# 2. 安装 Zookeeper
echo "\n2. 安装 Zookeeper..."
# 下载 Zookeeper 二进制包
ZOOKEEPER_VERSION="3.7.1"
wget -O /tmp/apache-zookeeper-${ZOOKEEPER_VERSION}-bin.tar.gz https://archive.apache.org/dist/zookeeper/zookeeper-${ZOOKEEPER_VERSION}/apache-zookeeper-${ZOOKEEPER_VERSION}-bin.tar.gz

# 解压并安装
mkdir -p /opt/zookeeper
tar -xzf /tmp/apache-zookeeper-${ZOOKEEPER_VERSION}-bin.tar.gz -C /opt/zookeeper --strip-components 1

# 创建数据目录
mkdir -p /var/lib/zookeeper

# 配置 Zookeeper
cp /opt/zookeeper/conf/zoo_sample.cfg /opt/zookeeper/conf/zoo.cfg
sed -i 's/dataDir=\/tmp\/zookeeper/dataDir=\/var\/lib\/zookeeper/' /opt/zookeeper/conf/zoo.cfg

# 创建 Zookeeper 系统服务
cat > /etc/systemd/system/zookeeper.service << EOF
[Unit]
Description=Apache Zookeeper Server
After=network.target

[Service]
Type=simple
User=root
ExecStart=/opt/zookeeper/bin/zkServer.sh start-foreground
ExecStop=/opt/zookeeper/bin/zkServer.sh stop
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

# 启动 Zookeeper
systemctl daemon-reload
systemctl enable zookeeper
systemctl start zookeeper

# 3. 安装 Kafka
echo "\n3. 安装 Kafka..."
# 下载 Kafka 二进制包
KAFKA_VERSION="3.5.1"
wget -O /tmp/kafka_2.13-${KAFKA_VERSION}.tgz https://archive.apache.org/dist/kafka/${KAFKA_VERSION}/kafka_2.13-${KAFKA_VERSION}.tgz

# 解压并安装
mkdir -p /opt/kafka
tar -xzf /tmp/kafka_2.13-${KAFKA_VERSION}.tgz -C /opt/kafka --strip-components 1

# 创建 Kafka 系统服务
cat > /etc/systemd/system/kafka.service << EOF
[Unit]
Description=Apache Kafka Server
Requires=zookeeper.service
After=zookeeper.service

[Service]
Type=simple
User=root
ExecStart=/opt/kafka/bin/kafka-server-start.sh /opt/kafka/config/server.properties
ExecStop=/opt/kafka/bin/kafka-server-stop.sh
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

# 启动 Kafka
systemctl enable kafka
systemctl start kafka

# 4. 安装 Elasticsearch
echo "\n4. 安装 Elasticsearch..."
# 导入 Elasticsearch GPG 密钥
rpm --import https://artifacts.elastic.co/GPG-KEY-elasticsearch

# 添加 Elasticsearch 仓库
cat > /etc/yum.repos.d/elasticsearch.repo << EOF
[elasticsearch-7.x]
name=Elasticsearch repository for 7.x packages
baseurl=https://artifacts.elastic.co/packages/7.x/yum
gpgcheck=1
gpgkey=https://artifacts.elastic.co/GPG-KEY-elasticsearch
enabled=1
autorefresh=1
type=rpm-md
EOF

# 安装 Elasticsearch
yum install -y elasticsearch

# 配置 Elasticsearch
sed -i 's/#cluster.name: my-application/cluster.name: cyberstroll-cluster/' /etc/elasticsearch/elasticsearch.yml
sed -i 's/#node.name: node-1/node.name: node-1/' /etc/elasticsearch/elasticsearch.yml
sed -i 's/#network.host: 192.168.0.1/network.host: 0.0.0.0/' /etc/elasticsearch/elasticsearch.yml
sed -i 's/#http.port: 9200/http.port: 9200/' /etc/elasticsearch/elasticsearch.yml
sed -i '/network.host/a discovery.type: single-node' /etc/elasticsearch/elasticsearch.yml

# 启动 Elasticsearch
systemctl enable elasticsearch
systemctl start elasticsearch

# 5. 验证安装
echo "\n5. 验证安装状态..."
echo "等待服务启动..."
sleep 10

echo "\nZookeeper 状态:"
systemctl status zookeeper --no-pager

echo "\nKafka 状态:"
systemctl status kafka --no-pager

echo "\nElasticsearch 状态:"
systemctl status elasticsearch --no-pager

echo "\nElasticsearch 健康检查:"
curl -s http://localhost:9200/_cluster/health

echo "\n====================================="
echo "部署完成！"
echo "组件访问信息:"
echo "- Zookeeper: localhost:2181"
echo "- Kafka: localhost:9092"
echo "- Elasticsearch: http://localhost:9200"
echo "====================================="
