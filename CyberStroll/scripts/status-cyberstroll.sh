#!/bin/bash

# CyberStroll状态检查脚本
# 检查所有服务和节点的运行状态

echo "📊 CyberStroll系统状态检查"
echo "================================"

# 进入项目目录
cd "$(dirname "$0")/.."

# 检查Docker服务状态
echo ""
echo "🐳 Docker服务状态:"
echo "--------------------------------"
docker-compose ps 2>/dev/null || echo "❌ Docker Compose服务未运行"

# 检查基础服务健康状态
echo ""
echo "🏥 基础服务健康检查:"
echo "--------------------------------"

# 检查MongoDB
echo -n "MongoDB:       "
if docker exec cyberstroll-mongodb mongosh --eval "db.adminCommand('ping')" --quiet > /dev/null 2>&1; then
    echo "✅ 正常运行"
else
    echo "❌ 服务异常"
fi

# 检查Elasticsearch
echo -n "Elasticsearch: "
if curl -s http://localhost:9200/_cluster/health > /dev/null; then
    health=$(curl -s http://localhost:9200/_cluster/health | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
    echo "✅ 正常运行 (状态: $health)"
else
    echo "❌ 服务异常"
fi

# 检查Kafka
echo -n "Kafka:         "
if docker exec cyberstroll-kafka kafka-broker-api-versions --bootstrap-server localhost:9092 > /dev/null 2>&1; then
    echo "✅ 正常运行"
else
    echo "❌ 服务异常"
fi

# 检查Redis
echo -n "Redis:         "
if docker exec cyberstroll-redis redis-cli -a cyberstroll123 ping > /dev/null 2>&1; then
    echo "✅ 正常运行"
else
    echo "❌ 服务异常"
fi

# 检查CyberStroll应用节点
echo ""
echo "🚀 CyberStroll应用节点状态:"
echo "--------------------------------"

check_node() {
    local node_name=$1
    local pid_file="logs/${node_name}.pid"
    local port=$2
    
    echo -n "${node_name}: "
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if ps -p $pid > /dev/null 2>&1; then
            echo -n "✅ 运行中 (PID: $pid)"
            
            # 检查端口监听 (如果提供了端口)
            if [ ! -z "$port" ]; then
                if lsof -i :$port > /dev/null 2>&1; then
                    echo " - 端口 $port 正常"
                else
                    echo " - ⚠️ 端口 $port 未监听"
                fi
            else
                echo ""
            fi
        else
            echo "❌ 进程不存在 (PID: $pid)"
        fi
    else
        echo "❌ 未启动"
    fi
}

check_node "任务管理节点" 8080
check_node "扫描节点"
check_node "处理节点"
check_node "搜索节点" 8082
check_node "富化节点"

# 检查Kafka主题
echo ""
echo "📝 Kafka主题状态:"
echo "--------------------------------"
if docker exec cyberstroll-kafka kafka-topics --list --bootstrap-server localhost:9092 > /dev/null 2>&1; then
    docker exec cyberstroll-kafka kafka-topics --list --bootstrap-server localhost:9092 | while read topic; do
        echo "  ✅ $topic"
    done
else
    echo "❌ 无法获取Kafka主题列表"
fi

# 检查Elasticsearch索引
echo ""
echo "🔍 Elasticsearch索引状态:"
echo "--------------------------------"
if curl -s http://localhost:9200/_cat/indices/cyberstroll_* > /dev/null 2>&1; then
    curl -s http://localhost:9200/_cat/indices/cyberstroll_* | while read line; do
        index=$(echo $line | awk '{print $3}')
        docs=$(echo $line | awk '{print $7}')
        size=$(echo $line | awk '{print $9}')
        echo "  ✅ $index (文档数: $docs, 大小: $size)"
    done
else
    echo "❌ 无法获取Elasticsearch索引信息"
fi

# 检查MongoDB集合
echo ""
echo "🗄️ MongoDB集合状态:"
echo "--------------------------------"
if docker exec cyberstroll-mongodb mongosh cyberstroll --eval "db.getCollectionNames()" --quiet > /dev/null 2>&1; then
    docker exec cyberstroll-mongodb mongosh cyberstroll --eval "
        db.getCollectionNames().forEach(function(collection) {
            var count = db[collection].countDocuments();
            print('  ✅ ' + collection + ' (文档数: ' + count + ')');
        });
    " --quiet 2>/dev/null
else
    echo "❌ 无法获取MongoDB集合信息"
fi

# 显示系统资源使用情况
echo ""
echo "💻 系统资源使用:"
echo "--------------------------------"
echo "Docker容器资源使用:"
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" 2>/dev/null | grep cyberstroll || echo "❌ 无法获取资源使用信息"

# 显示日志文件大小
echo ""
echo "📝 日志文件状态:"
echo "--------------------------------"
if [ -d "logs" ]; then
    for log_file in logs/*.log; do
        if [ -f "$log_file" ]; then
            size=$(du -h "$log_file" | cut -f1)
            echo "  📄 $(basename $log_file): $size"
        fi
    done
else
    echo "❌ 日志目录不存在"
fi

# 显示访问地址
echo ""
echo "🌐 服务访问地址:"
echo "--------------------------------"
echo "  任务管理界面:    http://localhost:8080"
echo "  搜索界面:        http://localhost:8082"
echo "  Kafka UI:        http://localhost:8080"
echo "  MongoDB Express: http://localhost:8081"
echo "  Kibana:          http://localhost:5601"

echo ""
echo "================================"
echo "状态检查完成 ✅"