package main

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
)

var (
	// 配置文件路径
	configPath string

	// 全局配置
	globalConfig *Config

	// NATS客户端
	natsClient *NATSClient

	// 节点管理器
	nodeManager *Manager
)

func main() {
	// 创建根命令
	var rootCmd = &cobra.Command{
		Use:   "nodemanage",
		Short: "节点管理系统",
		Long:  `基于Go语言的节点管理系统，用于管理远程节点并下发任务。`,
	}

	// 配置全局标志
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config.yaml", "配置文件路径")

	// 启动服务端命令
	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "启动服务端",
		RunE:  runServer,
	}

	// 启动客户端命令
	var clientCmd = &cobra.Command{
		Use:   "client",
		Short: "启动客户端",
		RunE:  runClient,
	}

	// 客户端特定标志
	var nodeID string
	clientCmd.Flags().StringVarP(&nodeID, "node-id", "n", "", "节点ID")
	clientCmd.MarkFlagRequired("node-id")

	// 安装节点命令
	var installCmd = &cobra.Command{
		Use:   "install [node-id]",
		Short: "安装客户端到节点",
		Args:  cobra.ExactArgs(1),
		RunE:  runInstall,
	}

	// 删除节点命令
	var deleteCmd = &cobra.Command{
		Use:   "delete [node-id]",
		Short: "从节点删除客户端",
		Args:  cobra.ExactArgs(1),
		RunE:  runDelete,
	}

	// 发送任务命令
	var taskCmd = &cobra.Command{
		Use:   "task [node-id] [command] [args...]",
		Short: "向节点发送任务",
		Args:  cobra.MinimumNArgs(2),
		RunE:  runTask,
	}

	// 查看节点状态命令
	var statusCmd = &cobra.Command{
		Use:   "status [node-id]",
		Short: "查看节点状态",
		Args:  cobra.ExactArgs(1),
		RunE:  runStatus,
	}

	// 列出所有节点命令
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "列出所有节点",
		RunE:  runList,
	}

	// 添加子命令
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(clientCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(listCmd)

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// 运行服务端
func runServer(cmd *cobra.Command, args []string) error {
	fmt.Println("启动节点管理服务端...")

	// 加载配置
	var err error
	globalConfig, err = LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("无法加载配置: %v", err)
	}

	// 初始化NATS客户端
	natsClient, err = NewNATSClient(
		globalConfig.NATS.URL,
		globalConfig.NATS.TaskSubject,
		globalConfig.NATS.ResultSubject,
	)
	if err != nil {
		return fmt.Errorf("无法初始化NATS客户端: %v", err)
	}
	defer natsClient.Close()

	// 初始化安装器
	installer := NewInstaller("./nodemanage")

	// 初始化节点管理器
	nodeManager = NewManager(installer)

	// 打印节点信息
	fmt.Printf("配置中包含 %d 个节点:\n", len(globalConfig.Nodes))
	for _, node := range globalConfig.Nodes {
		fmt.Printf("- %s: %s (%s:%d)\n", node.ID, node.Status, node.Host, node.Port)
	}

	// 订阅任务结果
	err = natsClient.SubscribeResults(func(result *TaskResult) {
		fmt.Printf("收到任务结果:\n")
		fmt.Printf("  任务ID: %s\n", result.TaskID)
		fmt.Printf("  节点ID: %s\n", result.NodeID)
		fmt.Printf("  退出码: %d\n", result.ExitCode)
		fmt.Printf("  输出: %s\n", result.Output)
		if result.Error != "" {
			fmt.Printf("  错误: %s\n", result.Error)
		}
		fmt.Println("-----------------------------")
	})
	if err != nil {
		return fmt.Errorf("无法订阅任务结果: %v", err)
	}

	fmt.Println("服务端已启动，等待任务结果...")
	fmt.Println("按 Ctrl+C 退出")

	// 保持运行
	select {}
}

// 运行客户端
func runClient(cmd *cobra.Command, args []string) error {
	nodeID, _ := cmd.Flags().GetString("node-id")
	fmt.Printf("启动节点管理客户端，节点ID: %s...\n", nodeID)

	// 加载配置
	var err error
	globalConfig, err = LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("无法加载配置: %v", err)
	}

	// 初始化NATS客户端
	natsClient, err = NewNATSClient(
		globalConfig.NATS.URL,
		globalConfig.NATS.TaskSubject,
		globalConfig.NATS.ResultSubject,
	)
	if err != nil {
		return fmt.Errorf("无法初始化NATS客户端: %v", err)
	}
	defer natsClient.Close()

	// 创建任务执行器
	executor := NewExecutor(func(result *TaskResult) {
		// 设置节点ID
		result.NodeID = nodeID

		// 发布结果
		log.Printf("发布任务结果: %s", result.TaskID)
		err := natsClient.PublishResult(result)
		if err != nil {
			log.Printf("无法发布任务结果: %v", err)
		}
	})

	// 订阅任务
	log.Printf("节点 %s 开始订阅任务...", nodeID)
	err = natsClient.SubscribeTasks(func(task *Task) {
		// 检查任务是否属于本节点
		if task.NodeID != "" && task.NodeID != nodeID {
			return
		}

		log.Printf("收到任务: %s\n", task.ID)
		log.Printf("  命令: %s %v\n", task.Command, task.Args)

		// 执行任务
		executor.ExecuteTask(task)
	})
	if err != nil {
		return fmt.Errorf("无法订阅任务: %v", err)
	}

	fmt.Println("客户端已启动，等待任务...")
	fmt.Println("按 Ctrl+C 退出")

	// 保持运行
	select {}
}

// 运行安装命令
func runInstall(cmd *cobra.Command, args []string) error {
	nodeID := args[0]

	// 加载配置
	var err error
	globalConfig, err = LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("无法加载配置: %v", err)
	}

	// 初始化安装器
	installer := NewInstaller("./nodemanage")

	// 初始化节点管理器
	nodeManager = NewManager(installer)

	// 查找节点
	var targetNode *Node
	for i, node := range globalConfig.Nodes {
		if node.ID == nodeID {
			targetNode = &globalConfig.Nodes[i]
			break
		}
	}

	if targetNode == nil {
		return fmt.Errorf("找不到节点 %s", nodeID)
	}

	// 安装节点
	log.Printf("开始安装节点 %s", nodeID)
	err = nodeManager.InstallNode(targetNode)
	if err != nil {
		return err
	}

	log.Printf("节点 %s 安装成功", nodeID)
	return nil
}

// 运行删除命令
func runDelete(cmd *cobra.Command, args []string) error {
	nodeID := args[0]

	// 加载配置
	var err error
	globalConfig, err = LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("无法加载配置: %v", err)
	}

	// 初始化安装器
	installer := NewInstaller("./nodemanage")

	// 初始化节点管理器
	nodeManager = NewManager(installer)

	// 查找节点
	var targetNode *Node
	for i, node := range globalConfig.Nodes {
		if node.ID == nodeID {
			targetNode = &globalConfig.Nodes[i]
			break
		}
	}

	if targetNode == nil {
		return fmt.Errorf("找不到节点 %s", nodeID)
	}

	// 删除节点
	log.Printf("开始删除节点 %s", nodeID)
	err = nodeManager.DeleteNode(targetNode)
	if err != nil {
		return err
	}

	log.Printf("节点 %s 删除成功", nodeID)
	return nil
}

// 运行任务命令
func runTask(cmd *cobra.Command, args []string) error {
	nodeID := args[0]
	command := args[1]
	cmdArgs := args[2:]

	// 加载配置
	var err error
	globalConfig, err = LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("无法加载配置: %v", err)
	}

	// 初始化NATS客户端
	natsClient, err = NewNATSClient(
		globalConfig.NATS.URL,
		globalConfig.NATS.TaskSubject,
		globalConfig.NATS.ResultSubject,
	)
	if err != nil {
		return fmt.Errorf("无法初始化NATS客户端: %v", err)
	}
	defer natsClient.Close()

	// 创建任务
	task := &Task{
		ID:        fmt.Sprintf("task-%d", time.Now().UnixNano()),
		NodeID:    nodeID,
		Command:   command,
		Args:      cmdArgs,
		Timeout:   30,
		CreatedAt: time.Now(),
	}

	// 发布任务
	log.Printf("发送任务 %s 到节点 %s", task.ID, nodeID)
	err = natsClient.PublishTask(task)
	if err != nil {
		return fmt.Errorf("无法发布任务: %v", err)
	}

	log.Printf("任务 %s 已发送", task.ID)
	return nil
}

// 运行状态命令
func runStatus(cmd *cobra.Command, args []string) error {
	nodeID := args[0]

	// 加载配置
	var err error
	globalConfig, err = LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("无法加载配置: %v", err)
	}

	// 查找节点
	var targetNode *Node
	for _, node := range globalConfig.Nodes {
		if node.ID == nodeID {
			targetNode = &node
			break
		}
	}

	if targetNode == nil {
		return fmt.Errorf("找不到节点 %s", nodeID)
	}

	// 获取状态
	fmt.Printf("节点 %s 状态: %s\n", nodeID, targetNode.Status)
	fmt.Printf("  主机: %s:%d\n", targetNode.Host, targetNode.Port)
	fmt.Printf("  用户: %s\n", targetNode.User)
	return nil
}

// 运行列表命令
func runList(cmd *cobra.Command, args []string) error {
	// 加载配置
	var err error
	globalConfig, err = LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("无法加载配置: %v", err)
	}

	// 列出所有节点
	fmt.Printf("共有 %d 个节点:\n", len(globalConfig.Nodes))
	for _, node := range globalConfig.Nodes {
		fmt.Printf("- %s: %s (%s:%d)\n", node.ID, node.Status, node.Host, node.Port)
	}
	return nil
}
