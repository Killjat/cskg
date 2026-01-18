package main

import (
	"fmt"
)

// Manager 节点管理器
type Manager struct {
	// 安装器
	Installer *Installer
}

// NewManager 创建节点管理器
func NewManager(installer *Installer) *Manager {
	return &Manager{
		Installer: installer,
	}
}

// InstallNode 安装节点
func (m *Manager) InstallNode(node *Node) error {
	// 使用安装器安装客户端
	err := m.Installer.Install(
		node.Host,
		node.Port,
		node.User,
		node.Password,
		node.ID,
	)
	if err != nil {
		return fmt.Errorf("无法安装节点 %s: %v", node.ID, err)
	}

	// 更新节点状态
	node.Status = "installed"

	return nil
}

// DeleteNode 删除节点
func (m *Manager) DeleteNode(node *Node) error {
	// 使用安装器卸载客户端
	err := m.Installer.Uninstall(
		node.Host,
		node.Port,
		node.User,
		node.Password,
		node.ID,
	)
	if err != nil {
		return fmt.Errorf("无法删除节点 %s: %v", node.ID, err)
	}

	// 更新节点状态
	node.Status = "deleted"

	return nil
}

// GetNodeStatus 获取节点状态
func (m *Manager) GetNodeStatus(node *Node) (string, error) {
	// 这里可以添加实际的状态检查逻辑
	// 例如通过SSH检查服务是否运行
	return node.Status, nil
}

// UpdateNodeStatus 更新节点状态
func (m *Manager) UpdateNodeStatus(node *Node, status string) {
	node.Status = status
}
