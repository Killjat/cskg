#!/bin/bash

# CyberStroll 启动脚本
# 用于便捷地启动各个模块

echo "=== CyberStroll 启动脚本 ==="
echo "Cyber Space Security Stroll Tool"
echo "Version: 1.0.0"
echo ""

# 显示帮助信息
show_help() {
    echo "使用方法: ./start.sh [模块名称] [参数]"
    echo ""
    echo "模块名称:"
    echo "  task_manager   启动任务管理模块"
    echo "  scan_node      启动扫描节点模块（单个实例）"
    echo "  scan_nodes     启动所有扫描节点实例（三个实例）"
    echo "  search         启动一键搜索模块"
    echo "  backend        启动后台系统（同时启动所有三个组件）"
    echo "  help           显示帮助信息"
    echo ""
    echo "示例:"
    echo "  ./start.sh task_manager -targets 192.168.1.1 -type port_scan"
    echo "  ./start.sh scan_node"
    echo "  ./start.sh scan_nodes"
    echo "  ./start.sh search -query http"
    echo "  ./start.sh backend"
    echo ""
    exit 1
}

# 检查Go环境
check_go_env() {
    if ! command -v go &> /dev/null; then
        echo "错误: 未检测到 Go 环境，请先安装 Go 1.21+"
        exit 1
    fi
    echo "✓ 检测到 Go 环境: $(go version)"
    echo ""
}

# 启动任务管理模块
start_task_manager() {
    echo "正在启动任务管理模块..."
    go run cmd/task_manager/main.go "$@"
}

# 启动扫描节点模块
start_scan_node() {
    echo "正在启动扫描节点模块..."
    go run cmd/scan_node/main.go "$@"
}

# 启动搜索模块
start_search() {
    echo "正在启动一键搜索模块..."
    go run cmd/search/main.go "$@"
}

# 启动所有扫描节点实例（三个实例）
start_scan_nodes() {
    echo "正在启动所有扫描节点实例..."
    echo "使用Docker Compose启动三个扫描节点实例"
    echo ""
    
    # 启动所有扫描节点实例
    docker-compose up -d scan_node_1 scan_node_2 scan_node_3
    
    echo ""
    echo "所有扫描节点实例启动完成！"
    echo "使用以下命令查看日志:"
    echo "  docker-compose logs scan_node_1"
    echo "  docker-compose logs scan_node_2"
    echo "  docker-compose logs scan_node_3"
    echo "使用以下命令停止所有实例:"
    echo "  docker-compose down scan_node_1 scan_node_2 scan_node_3"
}

# 启动后台系统（同时启动所有三个组件）
start_backend() {
    echo "正在启动后台系统..."
    echo "同时启动所有三个组件: task_manager, scan_node, search"
    echo ""
    
    # 启动扫描节点（后台运行）
    echo "1. 启动扫描节点模块..."
    go run cmd/scan_node/main.go > scan_node.log 2>&1 &
    SCAN_NODE_PID=$!
    echo "   扫描节点已启动，PID: $SCAN_NODE_PID，日志: scan_node.log"
    
    # 等待1秒，确保扫描节点启动
    sleep 1
    
    # 启动任务管理模块示例任务
    echo "2. 启动任务管理模块，下发示例任务..."
    go run cmd/task_manager/main.go -targets www.baidu.com -type port_scan -protocol tcp -port 80,443
    
    # 启动搜索模块示例查询
    echo "3. 启动搜索模块，查询扫描结果..."
    go run cmd/search/main.go
    
    echo ""
    echo "后台系统启动完成！"
    echo "扫描节点仍在后台运行，PID: $SCAN_NODE_PID"
    echo "使用 'kill $SCAN_NODE_PID' 可以停止扫描节点"
}

# 主函数
main() {
    # 检查Go环境
    check_go_env

    # 如果没有参数，显示帮助信息
    if [ $# -eq 0 ]; then
        show_help
    fi

    # 解析模块名称
    module="$1"
    shift

    # 根据模块名称启动对应的模块
    case "$module" in
        "task_manager")
            start_task_manager "$@"
            ;;
        "scan_node")
            start_scan_node "$@"
            ;;
        "scan_nodes")
            start_scan_nodes "$@"
            ;;
        "search")
            start_search "$@"
            ;;
        "backend")
            start_backend "$@"
            ;;
        "help")
            show_help
            ;;
        *)
            echo "错误: 未知的模块名称: $module"
            show_help
            ;;
    esac
}

# 执行主函数
main "$@"
