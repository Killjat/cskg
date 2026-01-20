#!/bin/bash

# CyberStroll 完整系统测试脚本

set -e

echo "🚀 开始 CyberStroll 完整系统测试..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    ((PASSED_TESTS++))
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    ((FAILED_TESTS++))
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 测试函数
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    ((TOTAL_TESTS++))
    log_info "执行测试: $test_name"
    
    if eval "$test_command"; then
        log_success "$test_name 通过"
        return 0
    else
        log_error "$test_name 失败"
        return 1
    fi
}

# 检查依赖服务
check_dependencies() {
    log_info "检查依赖服务..."
    
    # 检查Kafka
    if nc -z localhost 9092 2>/dev/null; then
        log_success "Kafka 服务正常 (localhost:9092)"
    else
        log_error "Kafka 服务不可用 (localhost:9092)"
        return 1
    fi
    
    # 检查MongoDB
    if nc -z localhost 27017 2>/dev/null; then
        log_success "MongoDB 服务正常 (localhost:27017)"
    else
        log_error "MongoDB 服务不可用 (localhost:27017)"
        return 1
    fi
    
    # 检查Elasticsearch
    if nc -z localhost 9200 2>/dev/null; then
        log_success "Elasticsearch 服务正常 (localhost:9200)"
    else
        log_error "Elasticsearch 服务不可用 (localhost:9200)"
        return 1
    fi
    
    return 0
}

# 构建项目
build_project() {
    log_info "构建项目..."
    
    if [ -f "scripts/build.sh" ]; then
        chmod +x scripts/build.sh
        if ./scripts/build.sh > /tmp/build.log 2>&1; then
            log_success "项目构建成功"
            return 0
        else
            log_error "项目构建失败，查看 /tmp/build.log"
            return 1
        fi
    else
        log_error "构建脚本不存在: scripts/build.sh"
        return 1
    fi
}

# 启动服务
start_services() {
    log_info "启动 CyberStroll 服务..."
    
    # 创建日志目录
    mkdir -p logs
    
    # 启动任务管理节点
    if [ -f "bin/task_manager" ]; then
        log_info "启动任务管理节点..."
        nohup ./bin/task_manager --config configs/task_manager.yaml > logs/task_manager.log 2>&1 &
        TASK_MANAGER_PID=$!
        echo $TASK_MANAGER_PID > logs/task_manager.pid
        sleep 3
        
        if kill -0 $TASK_MANAGER_PID 2>/dev/null; then
            log_success "任务管理节点启动成功 (PID: $TASK_MANAGER_PID)"
        else
            log_error "任务管理节点启动失败"
            return 1
        fi
    else
        log_warning "任务管理节点可执行文件不存在"
    fi
    
    # 启动扫描节点
    if [ -f "bin/scan_node" ]; then
        log_info "启动扫描节点..."
        nohup ./bin/scan_node --config configs/scan_node.yaml > logs/scan_node.log 2>&1 &
        SCAN_NODE_PID=$!
        echo $SCAN_NODE_PID > logs/scan_node.pid
        sleep 3
        
        if kill -0 $SCAN_NODE_PID 2>/dev/null; then
            log_success "扫描节点启动成功 (PID: $SCAN_NODE_PID)"
        else
            log_error "扫描节点启动失败"
            return 1
        fi
    else
        log_warning "扫描节点可执行文件不存在"
    fi
    
    # 启动处理节点
    if [ -f "bin/processor_node" ]; then
        log_info "启动处理节点..."
        nohup ./bin/processor_node --config configs/processor_node.yaml > logs/processor_node.log 2>&1 &
        PROCESSOR_NODE_PID=$!
        echo $PROCESSOR_NODE_PID > logs/processor_node.pid
        sleep 3
        
        if kill -0 $PROCESSOR_NODE_PID 2>/dev/null; then
            log_success "处理节点启动成功 (PID: $PROCESSOR_NODE_PID)"
        else
            log_error "处理节点启动失败"
            return 1
        fi
    else
        log_warning "处理节点可执行文件不存在"
    fi
    
    # 启动搜索节点
    if [ -f "bin/search_node" ]; then
        log_info "启动搜索节点..."
        nohup ./bin/search_node --config configs/search_node.yaml > logs/search_node.log 2>&1 &
        SEARCH_NODE_PID=$!
        echo $SEARCH_NODE_PID > logs/search_node.pid
        sleep 3
        
        if kill -0 $SEARCH_NODE_PID 2>/dev/null; then
            log_success "搜索节点启动成功 (PID: $SEARCH_NODE_PID)"
        else
            log_error "搜索节点启动失败"
            return 1
        fi
    else
        log_warning "搜索节点可执行文件不存在"
    fi
    
    # 等待服务完全启动
    log_info "等待服务完全启动..."
    sleep 10
    
    return 0
}

# 测试API接口
test_apis() {
    log_info "测试API接口..."
    
    # 测试任务管理节点API
    run_test "任务管理节点健康检查" "curl -s -f http://localhost:8080/api/stats > /dev/null"
    
    # 测试搜索节点API
    run_test "搜索节点健康检查" "curl -s -f http://localhost:8081/api/recent > /dev/null"
    
    # 测试任务提交
    run_test "任务提交API" 'curl -s -X POST http://localhost:8080/api/tasks/submit \
        -H "Content-Type: application/json" \
        -d "{\"initiator\":\"test\",\"targets\":[\"127.0.0.1\"],\"task_type\":\"port_scan_default\"}" \
        | grep -q "task_id"'
    
    return 0
}

# 测试Web界面
test_web_interfaces() {
    log_info "测试Web界面..."
    
    # 测试任务管理界面
    run_test "任务管理Web界面" "curl -s -f http://localhost:8080/ > /dev/null"
    
    # 测试搜索界面
    run_test "搜索Web界面" "curl -s -f http://localhost:8081/ > /dev/null"
    
    return 0
}

# 测试数据流
test_data_flow() {
    log_info "测试数据流..."
    
    # 提交测试任务
    log_info "提交测试扫描任务..."
    TASK_RESPONSE=$(curl -s -X POST http://localhost:8080/api/tasks/submit \
        -H "Content-Type: application/json" \
        -d '{"initiator":"system-test","targets":["127.0.0.1","8.8.8.8"],"task_type":"port_scan_default","timeout":10}')
    
    if echo "$TASK_RESPONSE" | grep -q "task_id"; then
        TASK_ID=$(echo "$TASK_RESPONSE" | grep -o '"task_id":"[^"]*"' | cut -d'"' -f4)
        log_success "任务提交成功，任务ID: $TASK_ID"
        
        # 等待任务处理
        log_info "等待任务处理..."
        sleep 30
        
        # 检查任务状态
        TASK_STATUS=$(curl -s "http://localhost:8080/api/tasks/status?task_id=$TASK_ID")
        if echo "$TASK_STATUS" | grep -q "completed\|processing\|pending"; then
            log_success "任务状态查询正常"
        else
            log_error "任务状态查询异常"
        fi
        
        # 检查搜索结果
        sleep 10
        SEARCH_RESULTS=$(curl -s "http://localhost:8081/api/search?ip=127.0.0.1")
        if echo "$SEARCH_RESULTS" | grep -q "results"; then
            log_success "搜索功能正常"
        else
            log_warning "搜索结果为空或异常"
        fi
    else
        log_error "任务提交失败"
    fi
    
    return 0
}

# 性能测试
test_performance() {
    log_info "执行性能测试..."
    
    # 并发任务提交测试
    log_info "并发任务提交测试..."
    for i in {1..5}; do
        curl -s -X POST http://localhost:8080/api/tasks/submit \
            -H "Content-Type: application/json" \
            -d "{\"initiator\":\"perf-test-$i\",\"targets\":[\"192.168.1.$i\"],\"task_type\":\"port_scan_default\"}" &
    done
    wait
    
    log_success "并发任务提交测试完成"
    
    # 搜索性能测试
    log_info "搜索性能测试..."
    for i in {1..10}; do
        curl -s "http://localhost:8081/api/search?page=$i&size=20" > /dev/null &
    done
    wait
    
    log_success "搜索性能测试完成"
    
    return 0
}

# 停止服务
stop_services() {
    log_info "停止 CyberStroll 服务..."
    
    # 停止所有节点
    for pid_file in logs/*.pid; do
        if [ -f "$pid_file" ]; then
            PID=$(cat "$pid_file")
            if kill -0 "$PID" 2>/dev/null; then
                log_info "停止进程 $PID"
                kill "$PID"
                sleep 2
                if kill -0 "$PID" 2>/dev/null; then
                    log_warning "强制停止进程 $PID"
                    kill -9 "$PID"
                fi
            fi
            rm -f "$pid_file"
        fi
    done
    
    log_success "所有服务已停止"
}

# 清理测试数据
cleanup_test_data() {
    log_info "清理测试数据..."
    
    # 清理MongoDB测试数据
    if command -v mongo &> /dev/null; then
        mongo cyberstroll --eval "db.tasks.deleteMany({initiator: /test/})" > /dev/null 2>&1 || true
        mongo cyberstroll --eval "db.task_statistics.deleteMany({task_id: /test/})" > /dev/null 2>&1 || true
    fi
    
    # 清理Elasticsearch测试数据
    if command -v curl &> /dev/null; then
        curl -s -X DELETE "http://localhost:9200/cyberstroll_test" > /dev/null 2>&1 || true
    fi
    
    log_success "测试数据清理完成"
}

# 生成测试报告
generate_report() {
    local end_time=$(date)
    local success_rate=$(( PASSED_TESTS * 100 / TOTAL_TESTS ))
    
    echo ""
    echo "=========================================="
    echo "🎯 CyberStroll 系统测试报告"
    echo "=========================================="
    echo "📊 测试统计:"
    echo "  总测试数: $TOTAL_TESTS"
    echo "  成功测试: $PASSED_TESTS"
    echo "  失败测试: $FAILED_TESTS"
    echo "  成功率: $success_rate%"
    echo ""
    echo "🕒 测试时间: $end_time"
    echo ""
    
    if [ $success_rate -ge 90 ]; then
        echo "🟢 系统状态: 优秀 - 可以部署到生产环境"
    elif [ $success_rate -ge 80 ]; then
        echo "🟡 系统状态: 良好 - 建议修复失败项后部署"
    elif [ $success_rate -ge 70 ]; then
        echo "🟠 系统状态: 一般 - 需要修复关键问题"
    else
        echo "🔴 系统状态: 较差 - 不建议部署，需要大量修复"
    fi
    
    echo ""
    echo "📁 日志文件位置:"
    echo "  - logs/task_manager.log"
    echo "  - logs/scan_node.log"
    echo "  - logs/processor_node.log"
    echo "  - logs/search_node.log"
    echo ""
    echo "🌐 Web界面:"
    echo "  - 任务管理: http://localhost:8080"
    echo "  - 搜索界面: http://localhost:8081"
    echo ""
    echo "=========================================="
}

# 主测试流程
main() {
    local start_time=$(date)
    log_info "开始时间: $start_time"
    
    # 检查依赖
    if ! check_dependencies; then
        log_error "依赖服务检查失败，请确保 Kafka、MongoDB、Elasticsearch 已启动"
        exit 1
    fi
    
    # 构建项目
    if ! build_project; then
        log_error "项目构建失败"
        exit 1
    fi
    
    # 启动服务
    if ! start_services; then
        log_error "服务启动失败"
        stop_services
        exit 1
    fi
    
    # 执行测试
    test_apis
    test_web_interfaces
    test_data_flow
    test_performance
    
    # 停止服务
    stop_services
    
    # 清理测试数据
    cleanup_test_data
    
    # 生成报告
    generate_report
    
    # 返回结果
    if [ $FAILED_TESTS -eq 0 ]; then
        log_success "所有测试通过！"
        exit 0
    else
        log_error "有 $FAILED_TESTS 个测试失败"
        exit 1
    fi
}

# 信号处理
trap 'log_warning "测试被中断"; stop_services; exit 1' INT TERM

# 执行主流程
main "$@"