package session

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"honeypot-system/internal/logger"
)

// SessionInfo 会话信息结构体
type SessionInfo struct {
	// 会话ID
	ID string `json:"id"`
	// 源IP
	SrcIP string `json:"src_ip"`
	// 源端口
	SrcPort int `json:"src_port"`
	// 目标IP
	DstIP string `json:"dst_ip"`
	// 目标端口
	DstPort int `json:"dst_port"`
	// 协议类型
	Protocol string `json:"protocol"`
	// 会话开始时间
	StartTime time.Time `json:"start_time"`
	// 会话结束时间
	EndTime time.Time `json:"end_time"`
	// 会话持续时间
	Duration int `json:"duration"`
	// 发送的数据包数量
	PacketsSent int `json:"packets_sent"`
	// 接收的数据包数量
	PacketsReceived int `json:"packets_received"`
	// 发送的字节数
	BytesSent int `json:"bytes_sent"`
	// 接收的字节数
	BytesReceived int `json:"bytes_received"`
	// 会话中的数据包列表
	Packets []PacketInfo `json:"packets"`
	// 会话状态
	Status string `json:"status"`
	// 设备指纹ID
	DeviceFingerprintID string `json:"device_fingerprint_id"`
}

// PacketInfo 数据包信息结构体
type PacketInfo struct {
	// 数据包ID
	ID string `json:"id"`
	// 数据包类型（请求/响应）
	Type string `json:"type"`
	// 数据包时间戳
	Timestamp time.Time `json:"timestamp"`
	// 数据包长度
	Length int `json:"length"`
	// 数据包内容（解析后）
	Content map[string]interface{} `json:"content"`
	// 原始数据包内容（十六进制）
	RawData string `json:"raw_data"`
	// 协议特定信息
	ProtocolInfo map[string]interface{} `json:"protocol_info"`
}

// SessionManager 会话管理器
type SessionManager struct {
	// 会话存储
	sessions map[string]*SessionInfo
	// 日志记录器
	logger logger.Logger
	// 互斥锁
	mutex sync.RWMutex
	// 会话超时时间（秒）
	timeout int
	// 清理间隔（秒）
	cleanupInterval int
	// 停止信号
	stopChan chan struct{}
	// 是否运行中
	running bool
}

// NewSessionManager 创建新的会话管理器
func NewSessionManager(logger logger.Logger, timeout int, cleanupInterval int) *SessionManager {
	return &SessionManager{
		sessions:        make(map[string]*SessionInfo),
		logger:          logger,
		timeout:         timeout,
		cleanupInterval: cleanupInterval,
		stopChan:        make(chan struct{}),
		running:         false,
	}
}

// Start 启动会话管理器
func (sm *SessionManager) Start() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.running {
		return
	}

	sm.running = true
	
	// 启动清理协程
	go sm.cleanupExpiredSessions()
	
	sm.logger.Info("Session manager started successfully")
}

// Stop 停止会话管理器
func (sm *SessionManager) Stop() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if !sm.running {
		return
	}

	// 发送停止信号
	close(sm.stopChan)
	
	sm.running = false
	
	sm.logger.Info("Session manager stopped successfully")
}

// GenerateSessionID 生成会话ID
func (sm *SessionManager) GenerateSessionID(srcIP string, srcPort int, dstIP string, dstPort int, protocol string) string {
	data := fmt.Sprintf("%s:%d:%s:%d:%s:%d", srcIP, srcPort, dstIP, dstPort, protocol, time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// CreateSession 创建新会话
func (sm *SessionManager) CreateSession(srcIP string, srcPort int, dstIP string, dstPort int, protocol string) *SessionInfo {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 生成会话ID
	sessionID := sm.GenerateSessionID(srcIP, srcPort, dstIP, dstPort, protocol)
	
	// 创建新会话
	session := &SessionInfo{
		ID:                 sessionID,
		SrcIP:              srcIP,
		SrcPort:            srcPort,
		DstIP:              dstIP,
		DstPort:            dstPort,
		Protocol:           protocol,
		StartTime:          time.Now(),
		Packets:            make([]PacketInfo, 0),
		Status:             "active",
	}
	
	// 存储会话
	sm.sessions[sessionID] = session
	
	return session
}

// GetSession 获取会话
func (sm *SessionManager) GetSession(sessionID string) (*SessionInfo, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	session, exists := sm.sessions[sessionID]
	return session, exists
}

// GetAllSessions 获取所有会话
func (sm *SessionManager) GetAllSessions() []*SessionInfo {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	sessions := make([]*SessionInfo, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	
	return sessions
}

// UpdateSession 更新会话
func (sm *SessionManager) UpdateSession(sessionID string, updateFn func(*SessionInfo) error) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}
	
	return updateFn(session)
}

// AddPacket 添加数据包到会话
func (sm *SessionManager) AddPacket(sessionID string, packet PacketInfo) error {
	return sm.UpdateSession(sessionID, func(session *SessionInfo) error {
		// 添加数据包
		session.Packets = append(session.Packets, packet)
		
		// 更新统计信息
		if packet.Type == "request" {
			session.PacketsSent++
			session.BytesSent += packet.Length
		} else if packet.Type == "response" {
			session.PacketsReceived++
			session.BytesReceived += packet.Length
		}
		
		// 更新会话最后活动时间
		session.EndTime = time.Now()
		session.Duration = int(time.Since(session.StartTime).Seconds())
		
		return nil
	})
}

// CloseSession 关闭会话
func (sm *SessionManager) CloseSession(sessionID string) error {
	return sm.UpdateSession(sessionID, func(session *SessionInfo) error {
		session.Status = "closed"
		session.EndTime = time.Now()
		session.Duration = int(time.Since(session.StartTime).Seconds())
		return nil
	})
}

// cleanupExpiredSessions 清理过期会话
func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(time.Duration(sm.cleanupInterval) * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			sm.mutex.Lock()
			now := time.Now()
			count := 0
			
			// 清理过期会话
			for sessionID, session := range sm.sessions {
				// 检查会话是否已关闭或过期
				if session.Status == "closed" || now.Sub(session.EndTime) > time.Duration(sm.timeout)*time.Second {
					delete(sm.sessions, sessionID)
					count++
				}
			}
			
			if count > 0 {
				sm.logger.Info(fmt.Sprintf("Cleaned up %d expired sessions", count))
			}
			
			sm.mutex.Unlock()
			
		case <-sm.stopChan:
			return
		}
	}
}
