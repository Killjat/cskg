package main

import (
	"encoding/binary"
	"fmt"
)

// generateOptimizedTLSClientHello 生成优化的TLS Client Hello握手包
func generateOptimizedTLSClientHello() []byte {
	// 优化的TLS 1.2 Client Hello包，提高兼容性和成功率
	clientHello := []byte{
		// TLS Record Header
		0x16,       // Content Type: Handshake (22)
		0x03, 0x03, // Version: TLS 1.2 (更好的兼容性)
		0x01, 0x00, // Length: 256 bytes (扩展长度)
		
		// Handshake Header
		0x01,       // Handshake Type: Client Hello (1)
		0x00, 0x00, 0xfc, // Length: 252 bytes
		
		// Client Hello
		0x03, 0x03, // Version: TLS 1.2
		
		// Random (32 bytes) - 固定随机数便于测试
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
		0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
		
		// Session ID Length
		0x00, // No session ID
		
		// Cipher Suites Length
		0x00, 0x20, // 32 bytes (16 cipher suites)
		
		// Cipher Suites (更广泛的兼容性)
		0xc0, 0x2c, // TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
		0xc0, 0x30, // TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
		0x00, 0x9f, // TLS_DHE_RSA_WITH_AES_256_GCM_SHA384
		0xcc, 0xa9, // TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256
		0xcc, 0xa8, // TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
		0xc0, 0x2b, // TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
		0xc0, 0x2f, // TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
		0x00, 0x9e, // TLS_DHE_RSA_WITH_AES_128_GCM_SHA256
		0xc0, 0x24, // TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384
		0xc0, 0x28, // TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384
		0xc0, 0x23, // TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256
		0xc0, 0x27, // TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256
		0xc0, 0x0a, // TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA
		0xc0, 0x14, // TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA
		0x00, 0x39, // TLS_DHE_RSA_WITH_AES_256_CBC_SHA
		0x00, 0x33, // TLS_DHE_RSA_WITH_AES_128_CBC_SHA
		
		// Compression Methods Length
		0x01, // 1 compression method
		
		// Compression Methods
		0x00, // No compression
		
		// Extensions Length
		0x00, 0x93, // 147 bytes of extensions
		
		// Server Name Indication (SNI) Extension
		0x00, 0x00, // Extension Type: server_name (0)
		0x00, 0x0e, // Extension Length: 14 bytes
		0x00, 0x0c, // Server Name List Length: 12 bytes
		0x00,       // Name Type: host_name (0)
		0x00, 0x09, // Host Name Length: 9 bytes
		0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x68, 0x6f, 0x73, 0x74, // "localhost"
		
		// Supported Groups Extension
		0x00, 0x0a, // Extension Type: supported_groups (10)
		0x00, 0x0c, // Extension Length: 12 bytes
		0x00, 0x0a, // Supported Groups List Length: 10 bytes
		0x00, 0x1d, // secp256r1
		0x00, 0x17, // secp256k1
		0x00, 0x1e, // secp384r1
		0x00, 0x19, // secp521r1
		0x00, 0x18, // secp256k1
		
		// EC Point Formats Extension
		0x00, 0x0b, // Extension Type: ec_point_formats (11)
		0x00, 0x04, // Extension Length: 4 bytes
		0x03,       // EC Point Formats Length: 3 bytes
		0x00,       // uncompressed
		0x01,       // ansiX962_compressed_prime
		0x02,       // ansiX962_compressed_char2
		
		// Signature Algorithms Extension
		0x00, 0x0d, // Extension Type: signature_algorithms (13)
		0x00, 0x20, // Extension Length: 32 bytes
		0x00, 0x1e, // Signature Hash Algorithms Length: 30 bytes
		0x06, 0x01, // rsa_pkcs1_sha512
		0x06, 0x02, // dsa_sha512
		0x06, 0x03, // ecdsa_sha512
		0x05, 0x01, // rsa_pkcs1_sha384
		0x05, 0x02, // dsa_sha384
		0x05, 0x03, // ecdsa_sha384
		0x04, 0x01, // rsa_pkcs1_sha256
		0x04, 0x02, // dsa_sha256
		0x04, 0x03, // ecdsa_sha256
		0x03, 0x01, // rsa_pkcs1_sha224
		0x03, 0x02, // dsa_sha224
		0x03, 0x03, // ecdsa_sha224
		0x02, 0x01, // rsa_pkcs1_sha1
		0x02, 0x02, // dsa_sha1
		0x02, 0x03, // ecdsa_sha1
		
		// Application Layer Protocol Negotiation (ALPN) Extension
		0x00, 0x10, // Extension Type: application_layer_protocol_negotiation (16)
		0x00, 0x0e, // Extension Length: 14 bytes
		0x00, 0x0c, // ALPN Extension Length: 12 bytes
		0x02, 0x68, 0x32, // h2 (HTTP/2)
		0x08, 0x68, 0x74, 0x74, 0x70, 0x2f, 0x31, 0x2e, 0x31, // http/1.1
		
		// Status Request Extension (OCSP)
		0x00, 0x05, // Extension Type: status_request (5)
		0x00, 0x05, // Extension Length: 5 bytes
		0x01,       // Certificate Status Type: ocsp (1)
		0x00, 0x00, // Responder ID list Length: 0
		0x00, 0x00, // Request Extensions Length: 0
		
		// Supported Versions Extension (TLS 1.3 compatibility)
		0x00, 0x2b, // Extension Type: supported_versions (43)
		0x00, 0x05, // Extension Length: 5 bytes
		0x04,       // Supported Versions Length: 4 bytes
		0x03, 0x04, // TLS 1.3
		0x03, 0x03, // TLS 1.2
		
		// Key Share Extension (for TLS 1.3)
		0x00, 0x33, // Extension Type: key_share (51)
		0x00, 0x26, // Extension Length: 38 bytes
		0x00, 0x24, // Client Key Share Length: 36 bytes
		0x00, 0x1d, // Group: secp256r1 (29)
		0x00, 0x20, // Key Exchange Length: 32 bytes
		// Dummy public key (32 bytes)
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
		0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
	}
	
	return clientHello
}

// generateOptimizedOracleTNSConnect 生成优化的Oracle TNS连接包
func generateOptimizedOracleTNSConnect(host string) []byte {
	// Oracle TNS (Transparent Network Substrate) Connect包
	// 这是Oracle数据库的网络协议
	
	// 构建Connect Data字符串
	connectData := fmt.Sprintf(
		"(CONNECT_DATA=(SERVICE_NAME=XE)(CID=(PROGRAM=sqlplus)(HOST=%s)(USER=oracle)))",
		host)
	
	// TNS包头
	tnsHeader := []byte{
		0x00, 0x00, // Packet Length (will be filled later)
		0x00, 0x00, // Packet Checksum
		0x01,       // Packet Type: Connect (1)
		0x00,       // Reserved
		0x00, 0x00, // Header Checksum
	}
	
	// TNS Connect包数据
	connectPacket := []byte{
		0x01, 0x36, // Version
		0x01, 0x2c, // Version (compatible)
		0x00, 0x00, // Service Options
		0x08, 0x00, // Session Data Unit Size
		0x7f, 0xff, // Maximum Transmission Unit
		0xa0, 0x0a, // NT Protocol Characteristics
		0x00, 0x00, // Line Turnaround Value
		0x01, 0x00, // Value of 1 in Hardware
		0x00, 0x00, // Length of Connect Data
	}
	
	// 添加Connect Data长度
	connectDataLen := len(connectData)
	binary.BigEndian.PutUint16(connectPacket[len(connectPacket)-2:], uint16(connectDataLen))
	
	// 组装完整包
	fullPacket := append(tnsHeader, connectPacket...)
	fullPacket = append(fullPacket, []byte(connectData)...)
	
	// 设置包长度
	binary.BigEndian.PutUint16(fullPacket[0:2], uint16(len(fullPacket)))
	
	return fullPacket
}

// generateOptimizedSQLServerTDS 生成优化的SQL Server TDS预登录包
func generateOptimizedSQLServerTDS() []byte {
	// SQL Server TDS (Tabular Data Stream) Pre-Login包
	// 用于SQL Server协议检测
	
	// TDS包头
	tdsHeader := []byte{
		0x12,       // Type: Pre-Login (0x12)
		0x01,       // Status: End of Message
		0x00, 0x2f, // Length: 47 bytes
		0x00, 0x00, // SPID: 0
		0x00,       // Packet ID: 0
		0x00,       // Window: 0
	}
	
	// Pre-Login选项
	preLoginOptions := []byte{
		// Option 1: Version
		0x00,       // Token: VERSION (0)
		0x00, 0x1a, // Offset: 26
		0x00, 0x06, // Length: 6
		
		// Option 2: Encryption
		0x01,       // Token: ENCRYPTION (1)
		0x00, 0x20, // Offset: 32
		0x00, 0x01, // Length: 1
		
		// Option 3: Instance
		0x02,       // Token: INSTOPT (2)
		0x00, 0x21, // Offset: 33
		0x00, 0x00, // Length: 0
		
		// Option 4: Thread ID
		0x03,       // Token: THREADID (3)
		0x00, 0x21, // Offset: 33
		0x00, 0x04, // Length: 4
		
		// Terminator
		0xff,
		
		// Version Data (6 bytes)
		0x0f, 0x00, 0x07, 0xd0, // Version: 15.0.2000
		0x00, 0x00, // Subbuild: 0
		
		// Encryption (1 byte)
		0x00, // ENCRYPT_OFF
		
		// Thread ID (4 bytes)
		0x00, 0x00, 0x00, 0x00, // Thread ID: 0
	}
	
	// 组装完整包
	fullPacket := append(tdsHeader, preLoginOptions...)
	
	return fullPacket
}

// generateOptimizedBACnetWhoIs 生成优化的BACnet Who-Is请求
func generateOptimizedBACnetWhoIs() []byte {
	// BACnet/IP Who-Is请求包
	// 用于楼宇自动化设备发现
	
	whoIsPacket := []byte{
		// BACnet/IP Header
		0x81,       // Type: BACnet/IP (0x81)
		0x0a,       // Function: Original-Unicast-NPDU (0x0a)
		0x00, 0x0c, // Length: 12 bytes
		
		// NPDU (Network Protocol Data Unit)
		0x01,       // Version: 1
		0x20,       // Control: Expecting Reply, Network Layer Message
		0xff, 0xff, // Destination Network: Global Broadcast
		0x00,       // Destination MAC Layer Address Length: 0
		0xff,       // Hop Count: 255 (no limit)
		
		// Network Layer Message Type
		0x00,       // Message Type: Who-Is-Router-To-Network (0)
		
		// APDU (Application Protocol Data Unit)
		0x10,       // PDU Type: Unconfirmed Request (1), Segmented: No (0)
		0x08,       // Service Choice: Who-Is (8)
		
		// Optional: Device Instance Range (uncomment for specific range)
		// 0x09, 0x22, // Context Tag: Unsigned Integer, Length: 2
		// 0x00, 0x00, // Low Limit: 0
		// 0x19, 0x22, // Context Tag: Unsigned Integer, Length: 2  
		// 0x3f, 0xff, // High Limit: 4194303
	}
	
	return whoIsPacket
}

// generateOptimizedS7COTP 生成优化的S7 COTP连接请求
func generateOptimizedS7COTP() []byte {
	// Siemens S7 Communication over COTP (Connection Oriented Transport Protocol)
	// 用于西门子PLC通信
	
	// TPKT Header (RFC 1006)
	tpktHeader := []byte{
		0x03, // Version: 3
		0x00, // Reserved: 0
		0x00, 0x16, // Length: 22 bytes
	}
	
	// COTP Connection Request
	cotpHeader := []byte{
		0x11, // Length Indicator: 17
		0xe0, // PDU Type: Connection Request (0xe0)
		0x00, 0x00, // Destination Reference: 0
		0x00, 0x01, // Source Reference: 1
		0x00, // Class and Option: Class 0
		
		// Parameters
		0xc0, // Parameter Code: tpdu-size (0xc0)
		0x01, // Parameter Length: 1
		0x0a, // TPDU Size: 1024 bytes (2^10)
		
		0xc1, // Parameter Code: src-tsap (0xc1)
		0x02, // Parameter Length: 2
		0x01, 0x00, // Source TSAP: 0x0100
		
		0xc2, // Parameter Code: dst-tsap (0xc2)
		0x02, // Parameter Length: 2
		0x01, 0x02, // Destination TSAP: 0x0102 (S7 Communication)
	}
	
	// 组装完整包
	fullPacket := append(tpktHeader, cotpHeader...)
	
	return fullPacket
}

// generateOptimizedLDAPBind 生成优化的LDAP绑定请求
func generateOptimizedLDAPBind() []byte {
	// LDAP Simple Bind Request (匿名绑定)
	// 用于LDAP目录服务检测
	
	// LDAP消息结构 (ASN.1 BER编码)
	ldapBind := []byte{
		// LDAP Message Sequence
		0x30, 0x0c, // SEQUENCE, Length: 12
		
		// Message ID
		0x02, 0x01, 0x01, // INTEGER: 1
		
		// Bind Request
		0x60, 0x07, // [APPLICATION 0] SEQUENCE, Length: 7
		
		// Version
		0x02, 0x01, 0x03, // INTEGER: 3 (LDAP v3)
		
		// Name (DN) - Empty for anonymous bind
		0x04, 0x00, // OCTET STRING, Length: 0
		
		// Authentication - Simple (password)
		0x80, 0x00, // [CONTEXT 0] OCTET STRING, Length: 0 (empty password)
	}
	
	return ldapBind
}

// generateOptimizedDockerAPI 生成优化的Docker API请求
func generateOptimizedDockerAPI() []byte {
	// Docker API HTTP请求 (无认证版本)
	// 用于检测Docker守护进程
	
	dockerRequest := []byte(
		"GET /version HTTP/1.1\r\n" +
		"Host: localhost\r\n" +
		"User-Agent: Docker-Client\r\n" +
		"Accept: application/json\r\n" +
		"Connection: close\r\n" +
		"\r\n")
	
	return dockerRequest
}

// generateOptimizedOPCUAHello 生成优化的OPC UA Hello消息
func generateOptimizedOPCUAHello() []byte {
	// OPC UA Hello Message
	// 用于工业4.0 OPC UA服务器检测
	
	// OPC UA消息头
	opcuaHeader := []byte{
		'H', 'E', 'L', 'F', // Message Type: "HELF" (Hello Final)
		0x38, 0x00, 0x00, 0x00, // Message Size: 56 bytes
	}
	
	// Hello消息体
	helloBody := []byte{
		0x00, 0x00, 0x00, 0x00, // Protocol Version: 0
		0x00, 0x80, 0x00, 0x00, // Receive Buffer Size: 32768
		0x00, 0x80, 0x00, 0x00, // Send Buffer Size: 32768
		0x00, 0x00, 0x10, 0x00, // Max Message Size: 1048576
		0x00, 0x00, 0x00, 0x00, // Max Chunk Count: 0
		
		// Endpoint URL Length
		0x20, 0x00, 0x00, 0x00, // Length: 32
		
		// Endpoint URL: "opc.tcp://localhost:4840/server"
		0x6f, 0x70, 0x63, 0x2e, 0x74, 0x63, 0x70, 0x3a,
		0x2f, 0x2f, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x68,
		0x6f, 0x73, 0x74, 0x3a, 0x34, 0x38, 0x34, 0x30,
		0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x00,
	}
	
	// 组装完整包
	fullPacket := append(opcuaHeader, helloBody...)
	
	return fullPacket
}