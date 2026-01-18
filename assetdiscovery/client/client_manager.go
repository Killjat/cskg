package main

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

// ClientManager 客户端管理器结构体
type ClientManager struct {
	client *Client
}

// NewClientManager 创建新的客户端管理器
func NewClientManager(client *Client) *ClientManager {
	return &ClientManager{
		client: client,
	}
}

// Register 向服务端注册客户端
func (cm *ClientManager) Register() {
	// 这里简化处理，实际应该向服务端发送注册请求
	log.Printf("Client %s registered with server", cm.client.config.Client.ClientID)
}

// SendHeartbeats 定期发送心跳到服务端
func (cm *ClientManager) SendHeartbeats() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		cm.sendHeartbeat()
	}
}

// sendHeartbeat 发送心跳到服务端
func (cm *ClientManager) sendHeartbeat() {
	// 这里简化处理，实际应该向服务端发送HTTP请求
	log.Printf("Client %s sending heartbeat", cm.client.config.Client.ClientID)

	// 示例：尝试连接到服务端
	serverURL := url.URL{
		Scheme: "http",
		Host:   "localhost:8080",
		Path:   "/api/heartbeat",
	}

	// 创建心跳请求
	req, err := http.NewRequest("POST", serverURL.String(), nil)
	if err != nil {
		log.Printf("Error creating heartbeat request: %v", err)
		return
	}

	// 设置请求头
	req.Header.Set("Client-ID", cm.client.config.Client.ClientID)

	// 发送请求
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		// 服务端可能未启动，不报错
		return
	}
	defer resp.Body.Close()

	log.Printf("Client %s heartbeat sent successfully", cm.client.config.Client.ClientID)
}
