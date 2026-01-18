package main

import (
	"log"
	"sync"
	"time"

	"github.com/cskg/assetdiscovery/common"
)

// ClientManager 客户端管理器结构体
type ClientManager struct {
	server   *Server
	clients  map[string]*common.ClientStatus
	clientsLock sync.RWMutex
	heartbeatTimeout int64
}

// NewClientManager 创建新的客户端管理器
func NewClientManager(server *Server) *ClientManager {
	return &ClientManager{
		server:           server,
		clients:          make(map[string]*common.ClientStatus),
		heartbeatTimeout: 300, // 5分钟
	}
}

// Start 启动客户端管理器
func (cm *ClientManager) Start() {
	log.Println("Starting client manager...")

	// 启动心跳检查协程
	go cm.checkHeartbeats()
}

// RegisterClient 注册客户端
func (cm *ClientManager) RegisterClient(clientID string) {
	cm.clientsLock.Lock()
	defer cm.clientsLock.Unlock()

	client := &common.ClientStatus{
		ClientID:     clientID,
		Status:       "online",
		LastSeen:     time.Now().Unix(),
		ActiveTasks:  0,
		CompletedTasks: 0,
	}

	cm.clients[clientID] = client
	log.Printf("Client registered: %s", clientID)
}

// UpdateClientHeartbeat 更新客户端心跳
func (cm *ClientManager) UpdateClientHeartbeat(clientID string) {
	cm.clientsLock.Lock()
	defer cm.clientsLock.Unlock()

	if client, exists := cm.clients[clientID]; exists {
		client.LastSeen = time.Now().Unix()
		client.Status = "online"
		log.Printf("Client heartbeat updated: %s", clientID)
	}
}

// UpdateClientStatus 更新客户端状态
func (cm *ClientManager) UpdateClientStatus(clientID string, activeTasks int, completedTasks int) {
	cm.clientsLock.Lock()
	defer cm.clientsLock.Unlock()

	if client, exists := cm.clients[clientID]; exists {
		client.ActiveTasks = activeTasks
		client.CompletedTasks = completedTasks
		client.LastSeen = time.Now().Unix()
	}
}

// GetClient 获取客户端信息
func (cm *ClientManager) GetClient(clientID string) (*common.ClientStatus, bool) {
	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()

	client, exists := cm.clients[clientID]
	return client, exists
}

// GetAllClients 获取所有客户端
func (cm *ClientManager) GetAllClients() []*common.ClientStatus {
	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()

	clients := make([]*common.ClientStatus, 0, len(cm.clients))
	for _, client := range cm.clients {
		clients = append(clients, client)
	}

	return clients
}

// GetAvailableClients 获取可用客户端数量
func (cm *ClientManager) GetAvailableClients() []string {
	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()

	var available []string
	currentTime := time.Now().Unix()

	for clientID, client := range cm.clients {
		// 检查客户端是否在线且活跃
		if client.Status == "online" && (currentTime-client.LastSeen) < cm.heartbeatTimeout {
			available = append(available, clientID)
		}
	}

	return available
}

// checkHeartbeats 检查客户端心跳
func (cm *ClientManager) checkHeartbeats() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		cm.clientsLock.Lock()

		currentTime := time.Now().Unix()
		for clientID, client := range cm.clients {
			if (currentTime - client.LastSeen) > cm.heartbeatTimeout {
				client.Status = "offline"
				log.Printf("Client %s marked as offline due to heartbeat timeout", clientID)
			}
		}

		cm.clientsLock.Unlock()
	}
}

// GetLeastLoadedClient 获取负载最低的客户端
func (cm *ClientManager) GetLeastLoadedClient() string {
	cm.clientsLock.RLock()
	defer cm.clientsLock.RUnlock()

	var leastLoadedClient string
	minActiveTasks := -1
	currentTime := time.Now().Unix()

	for clientID, client := range cm.clients {
		// 跳过离线客户端
		if client.Status != "online" || (currentTime-client.LastSeen) > cm.heartbeatTimeout {
			continue
		}

		// 选择活跃任务最少的客户端
		if minActiveTasks == -1 || client.ActiveTasks < minActiveTasks {
			minActiveTasks = client.ActiveTasks
			leastLoadedClient = clientID
		}
	}

	return leastLoadedClient
}
