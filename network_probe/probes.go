package main

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ProbeLoader 探测载荷加载器
type ProbeLoader struct {
	probes map[string]*Probe
}

// NewProbeLoader 创建探测加载器
func NewProbeLoader() *ProbeLoader {
	return &ProbeLoader{
		probes: make(map[string]*Probe),
	}
}

// LoadBuiltinProbes 加载内置探测
func (pl *ProbeLoader) LoadBuiltinProbes() map[string]*Probe {
	builtinProbes := []*Probe{
		// NULL探测 - 只连接不发送数据
		{
			Name:        "NULL",
			Type:        ProbeTypeTCP,
			Payload:     []byte{},
			PayloadHex:  "",
			Ports:       []int{21, 22, 23, 25, 53, 80, 110, 143, 443, 993, 995, 3306, 5432, 6379},
			Protocol:    "tcp",
			Description: "TCP NULL probe - just connect",
			Timeout:     5,
			Rarity:      1,
		},
		
		// HTTP GET请求
		{
			Name:        "GetRequest",
			Type:        ProbeTypeTCP,
			Payload:     []byte("GET / HTTP/1.0\r\n\r\n"),
			PayloadHex:  "474554202f20485454502f312e300d0a0d0a",
			Ports:       []int{80, 443, 8080, 8443, 8000, 8888},
			Protocol:    "http",
			Description: "HTTP GET request",
			Timeout:     10,
			Rarity:      2,
		},
		
		// HTTP OPTIONS请求
		{
			Name:        "HTTPOptions",
			Type:        ProbeTypeTCP,
			Payload:     []byte("OPTIONS / HTTP/1.0\r\n\r\n"),
			PayloadHex:  "4f5054494f4e53202f20485454502f312e300d0a0d0a",
			Ports:       []int{80, 443, 8080},
			Protocol:    "http",
			Description: "HTTP OPTIONS request",
			Timeout:     10,
			Rarity:      3,
		},
		
		// FTP探测
		{
			Name:        "FTPBounce",
			Type:        ProbeTypeTCP,
			Payload:     []byte("USER anonymous\r\n"),
			PayloadHex:  "55534552206e6f6e796d6f75730d0a",
			Ports:       []int{21},
			Protocol:    "ftp",
			Description: "FTP user command",
			Timeout:     10,
			Rarity:      4,
		},
		
		// SMTP探测
		{
			Name:        "SMTPOptions",
			Type:        ProbeTypeTCP,
			Payload:     []byte("EHLO example.com\r\n"),
			PayloadHex:  "45484c4f206578616d706c652e636f6d0d0a",
			Ports:       []int{25, 465, 587},
			Protocol:    "smtp",
			Description: "SMTP EHLO command",
			Timeout:     10,
			Rarity:      3,
		},
		
		// SSH探测 (通过NULL探测获取banner)
		{
			Name:        "SSHVersionExchange",
			Type:        ProbeTypeTCP,
			Payload:     []byte{},
			PayloadHex:  "",
			Ports:       []int{22},
			Protocol:    "ssh",
			Description: "SSH version exchange",
			Timeout:     5,
			Rarity:      1,
		},
		
		// MySQL探测
		{
			Name:        "MySQLGreeting",
			Type:        ProbeTypeTCP,
			Payload:     []byte{},
			PayloadHex:  "",
			Ports:       []int{3306},
			Protocol:    "mysql",
			Description: "MySQL greeting message",
			Timeout:     5,
			Rarity:      2,
		},
		
		// Redis探测
		{
			Name:        "RedisPing",
			Type:        ProbeTypeTCP,
			Payload:     []byte("*1\r\n$4\r\nPING\r\n"),
			PayloadHex:  "2a310d0a24340d0a50494e470d0a",
			Ports:       []int{6379},
			Protocol:    "redis",
			Description: "Redis PING command",
			Timeout:     5,
			Rarity:      5,
		},
		
		// PostgreSQL探测
		{
			Name:        "PostgreSQLStartup",
			Type:        ProbeTypeTCP,
			Payload:     []byte("\x00\x00\x00\x17\x00\x03\x00\x00user\x00test\x00database\x00test\x00\x00"),
			PayloadHex:  "0000001700030000757365720074657374006461746162617365007465737400",
			Ports:       []int{5432},
			Protocol:    "postgresql",
			Description: "PostgreSQL startup message",
			Timeout:     5,
			Rarity:      6,
		},
		
		// DNS探测
		{
			Name:        "DNSStatusRequest",
			Type:        ProbeTypeUDP,
			Payload:     []byte("\x00\x00\x10\x00\x00\x00\x00\x00\x00\x00\x00\x00"),
			PayloadHex:  "000010000000000000000000",
			Ports:       []int{53},
			Protocol:    "dns",
			Description: "DNS status request",
			Timeout:     5,
			Rarity:      3,
		},
		
		// SNMP探测
		{
			Name:        "SNMPv1GetRequest",
			Type:        ProbeTypeUDP,
			Payload:     []byte("\x30\x26\x02\x01\x00\x04\x06\x70\x75\x62\x6c\x69\x63\xa0\x19\x02\x04\x00\x00\x00\x00\x02\x01\x00\x02\x01\x00\x30\x0b\x30\x09\x06\x05\x2b\x06\x01\x02\x01\x05\x00"),
			PayloadHex:  "3026020100040670756c696963a019020400000000020100020100300b300906052b060102010500",
			Ports:       []int{161},
			Protocol:    "snmp",
			Description: "SNMP v1 GetRequest",
			Timeout:     5,
			Rarity:      7,
		},
		
		// Telnet探测
		{
			Name:        "TelnetOptions",
			Type:        ProbeTypeTCP,
			Payload:     []byte("\xff\xfb\x01\xff\xfb\x03\xff\xfc\x27"),
			PayloadHex:  "fffb01fffb03fffc27",
			Ports:       []int{23},
			Protocol:    "telnet",
			Description: "Telnet option negotiation",
			Timeout:     5,
			Rarity:      4,
		},
		
		// POP3探测
		{
			Name:        "POP3Capabilities",
			Type:        ProbeTypeTCP,
			Payload:     []byte("CAPA\r\n"),
			PayloadHex:  "434150410d0a",
			Ports:       []int{110, 995},
			Protocol:    "pop3",
			Description: "POP3 capabilities command",
			Timeout:     10,
			Rarity:      4,
		},
		
		// IMAP探测
		{
			Name:        "IMAPCapabilities",
			Type:        ProbeTypeTCP,
			Payload:     []byte("A001 CAPABILITY\r\n"),
			PayloadHex:  "41303031204341504142494c4954590d0a",
			Ports:       []int{143, 993},
			Protocol:    "imap",
			Description: "IMAP capabilities command",
			Timeout:     10,
			Rarity:      4,
		},
		
		// 优化的HTTPS/TLS探测
		{
			Name:        "TLSClientHello",
			Type:        ProbeTypeTCP,
			Payload:     generateOptimizedTLSClientHello(),
			PayloadHex:  hex.EncodeToString(generateOptimizedTLSClientHello()),
			Ports:       []int{443, 8443, 993, 995, 465, 587, 636, 989, 990, 992, 5061},
			Protocol:    "tls",
			Description: "Optimized TLS Client Hello handshake",
			Timeout:     10,
			Rarity:      2,
		},
		
		// HTTPS GET请求 (通过TLS)
		{
			Name:        "HTTPSGetRequest",
			Type:        ProbeTypeTCP,
			Payload:     []byte{}, // 需要TLS握手后发送
			PayloadHex:  "",
			Ports:       []int{443, 8443},
			Protocol:    "https",
			Description: "HTTPS GET request over TLS",
			Timeout:     15,
			Rarity:      3,
		},
		
		// MQTT探测
		{
			Name:        "MQTTConnect",
			Type:        ProbeTypeTCP,
			Payload:     generateMQTTConnectPacket(),
			PayloadHex:  hex.EncodeToString(generateMQTTConnectPacket()),
			Ports:       []int{1883, 8883, 1884, 8884},
			Protocol:    "mqtt",
			Description: "MQTT CONNECT packet",
			Timeout:     10,
			Rarity:      4,
		},
		
		// MQTT over WebSocket探测
		{
			Name:        "MQTTWebSocket",
			Type:        ProbeTypeTCP,
			Payload:     []byte("GET /mqtt HTTP/1.1\r\nHost: localhost\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Protocol: mqtt\r\nSec-WebSocket-Version: 13\r\n\r\n"),
			PayloadHex:  "474554202f6d71747420485454502f312e310d0a486f73743a206c6f63616c686f73740d0a557067726164653a20776562736f636b65740d0a436f6e6e656374696f6e3a20557067726164650d0a5365632d576562536f636b65742d4b65793a2064476868494842686257787349484e76626d4e6c0d0a5365632d576562536f636b65742d50726f746f636f6c3a206d7174740d0a5365632d576562536f636b65742d56657273696f6e3a2031330d0a0d0a",
			Ports:       []int{8080, 9001, 8000},
			Protocol:    "mqtt-ws",
			Description: "MQTT over WebSocket",
			Timeout:     10,
			Rarity:      6,
		},
		
		// RTSP探测 (摄像头主流协议)
		{
			Name:        "RTSPOptions",
			Type:        ProbeTypeTCP,
			Payload:     []byte("OPTIONS rtsp://127.0.0.1/ RTSP/1.0\r\nCSeq: 1\r\nUser-Agent: NetworkProbe/1.0\r\n\r\n"),
			PayloadHex:  "4f5054494f4e532072747370733a2f2f3132372e302e302e312f20525453502f312e300d0a435365713a20310d0a557365722d4167656e743a204e6574776f726b50726f62652f312e300d0a0d0a",
			Ports:       []int{554, 8554, 1935, 8000, 8080},
			Protocol:    "rtsp",
			Description: "RTSP OPTIONS request for IP cameras",
			Timeout:     10,
			Rarity:      3,
		},
		
		// RTSP DESCRIBE探测
		{
			Name:        "RTSPDescribe",
			Type:        ProbeTypeTCP,
			Payload:     []byte("DESCRIBE rtsp://127.0.0.1/ RTSP/1.0\r\nCSeq: 2\r\nUser-Agent: NetworkProbe/1.0\r\nAccept: application/sdp\r\n\r\n"),
			PayloadHex:  "4445534352494245207274737073a2f2f3132372e302e302e312f20525453502f312e300d0a435365713a20320d0a557365722d4167656e743a204e6574776f726b50726f62652f312e300d0a4163636570743a206170706c69636174696f6e2f7364700d0a0d0a",
			Ports:       []int{554, 8554},
			Protocol:    "rtsp",
			Description: "RTSP DESCRIBE request for stream info",
			Timeout:     10,
			Rarity:      4,
		},
		
		// ONVIF设备发现 (WS-Discovery)
		{
			Name:        "ONVIFDiscovery",
			Type:        ProbeTypeUDP,
			Payload:     generateONVIFDiscoveryPacket(),
			PayloadHex:  hex.EncodeToString(generateONVIFDiscoveryPacket()),
			Ports:       []int{3702},
			Protocol:    "onvif",
			Description: "ONVIF WS-Discovery probe",
			Timeout:     5,
			Rarity:      5,
		},
		
		// ONVIF HTTP探测
		{
			Name:        "ONVIFDeviceService",
			Type:        ProbeTypeTCP,
			Payload:     generateONVIFDeviceServiceRequest(),
			PayloadHex:  hex.EncodeToString(generateONVIFDeviceServiceRequest()),
			Ports:       []int{80, 8080, 8000, 8899},
			Protocol:    "onvif-http",
			Description: "ONVIF Device Service request",
			Timeout:     10,
			Rarity:      4,
		},
		
		// 海康威视私有协议
		{
			Name:        "HikvisionISAPI",
			Type:        ProbeTypeTCP,
			Payload:     []byte("GET /ISAPI/System/deviceInfo HTTP/1.1\r\nHost: 127.0.0.1\r\nUser-Agent: NetworkProbe/1.0\r\nAuthorization: Basic YWRtaW46MTIzNDU2\r\n\r\n"),
			PayloadHex:  "474554202f49534150492f53797374656d2f6465766963654e666f20485454502f312e310d0a486f73743a203132372e302e302e310d0a557365722d4167656e743a204e6574776f726b50726f62652f312e300d0a417574686f72697a6174696f6e3a2042617369632059574674615735364d5449794e4455320d0a0d0a",
			Ports:       []int{80, 8000, 8080, 443},
			Protocol:    "hikvision",
			Description: "Hikvision ISAPI device info request",
			Timeout:     10,
			Rarity:      5,
		},
		
		// 大华私有协议
		{
			Name:        "DahuaLogin",
			Type:        ProbeTypeTCP,
			Payload:     generateDahuaLoginPacket(),
			PayloadHex:  hex.EncodeToString(generateDahuaLoginPacket()),
			Ports:       []int{37777, 37778, 80, 8000},
			Protocol:    "dahua",
			Description: "Dahua camera login probe",
			Timeout:     10,
			Rarity:      6,
		},
		
		// Modbus TCP探测 (工控协议)
		{
			Name:        "ModbusTCP",
			Type:        ProbeTypeTCP,
			Payload:     generateModbusTCPPacket(),
			PayloadHex:  hex.EncodeToString(generateModbusTCPPacket()),
			Ports:       []int{502},
			Protocol:    "modbus",
			Description: "Modbus TCP read coils request",
			Timeout:     5,
			Rarity:      4,
		},
		
		// DNP3探测 (电力系统)
		{
			Name:        "DNP3Request",
			Type:        ProbeTypeTCP,
			Payload:     generateDNP3Packet(),
			PayloadHex:  hex.EncodeToString(generateDNP3Packet()),
			Ports:       []int{20000, 19999},
			Protocol:    "dnp3",
			Description: "DNP3 link layer test request",
			Timeout:     5,
			Rarity:      5,
		},
		
		// BACnet探测 (楼宇自动化)
		{
			Name:        "BACnetWhoIs",
			Type:        ProbeTypeUDP,
			Payload:     generateBACnetWhoIsPacket(),
			PayloadHex:  hex.EncodeToString(generateBACnetWhoIsPacket()),
			Ports:       []int{47808},
			Protocol:    "bacnet",
			Description: "BACnet Who-Is broadcast request",
			Timeout:     5,
			Rarity:      5,
		},
		
		// OPC UA探测 (工业4.0)
		{
			Name:        "OPCUAHello",
			Type:        ProbeTypeTCP,
			Payload:     generateOPCUAHelloPacket(),
			PayloadHex:  hex.EncodeToString(generateOPCUAHelloPacket()),
			Ports:       []int{4840, 4843},
			Protocol:    "opcua",
			Description: "OPC UA Hello message",
			Timeout:     10,
			Rarity:      4,
		},
		
		// S7 Protocol探测 (西门子PLC)
		{
			Name:        "S7CommSetup",
			Type:        ProbeTypeTCP,
			Payload:     generateS7SetupPacket(),
			PayloadHex:  hex.EncodeToString(generateS7SetupPacket()),
			Ports:       []int{102},
			Protocol:    "s7",
			Description: "Siemens S7 communication setup",
			Timeout:     5,
			Rarity:      5,
		},
		
		// 数据库协议探测
		// 优化的SQL Server探测
		{
			Name:        "SQLServerPreLogin",
			Type:        ProbeTypeTCP,
			Payload:     generateOptimizedSQLServerTDS(),
			PayloadHex:  hex.EncodeToString(generateOptimizedSQLServerTDS()),
			Ports:       []int{1433, 1434},
			Protocol:    "sqlserver",
			Description: "Optimized Microsoft SQL Server TDS pre-login",
			Timeout:     10,
			Rarity:      4,
		},
		
		// 优化的Oracle探测
		{
			Name:        "OracleConnect",
			Type:        ProbeTypeTCP,
			Payload:     generateOptimizedOracleTNSConnect("localhost"),
			PayloadHex:  hex.EncodeToString(generateOptimizedOracleTNSConnect("localhost")),
			Ports:       []int{1521, 1522},
			Protocol:    "oracle",
			Description: "Optimized Oracle Database TNS connect",
			Timeout:     10,
			Rarity:      4,
		},
		
		// MongoDB探测
		{
			Name:        "MongoDBHandshake",
			Type:        ProbeTypeTCP,
			Payload:     generateMongoDBHandshakePacket(),
			PayloadHex:  hex.EncodeToString(generateMongoDBHandshakePacket()),
			Ports:       []int{27017, 27018, 27019},
			Protocol:    "mongodb",
			Description: "MongoDB wire protocol handshake",
			Timeout:     10,
			Rarity:      4,
		},
		
		// Elasticsearch探测
		{
			Name:        "ElasticsearchInfo",
			Type:        ProbeTypeTCP,
			Payload:     []byte("GET / HTTP/1.1\r\nHost: localhost\r\nUser-Agent: NetworkProbe/1.0\r\n\r\n"),
			PayloadHex:  "474554202f20485454502f312e310d0a486f73743a206c6f63616c686f73740d0a557365722d4167656e743a204e6574776f726b50726f62652f312e300d0a0d0a",
			Ports:       []int{9200, 9300},
			Protocol:    "elasticsearch",
			Description: "Elasticsearch cluster info",
			Timeout:     10,
			Rarity:      4,
		},
		
		// InfluxDB探测
		{
			Name:        "InfluxDBPing",
			Type:        ProbeTypeTCP,
			Payload:     []byte("GET /ping HTTP/1.1\r\nHost: localhost\r\nUser-Agent: NetworkProbe/1.0\r\n\r\n"),
			PayloadHex:  "474554202f70696e6720485454502f312e310d0a486f73743a206c6f63616c686f73740d0a557365722d4167656e743a204e6574776f726b50726f62652f312e300d0a0d0a",
			Ports:       []int{8086, 8088},
			Protocol:    "influxdb",
			Description: "InfluxDB ping endpoint",
			Timeout:     10,
			Rarity:      5,
		},
		
		// Cassandra探测
		{
			Name:        "CassandraOptions",
			Type:        ProbeTypeTCP,
			Payload:     generateCassandraOptionsPacket(),
			PayloadHex:  hex.EncodeToString(generateCassandraOptionsPacket()),
			Ports:       []int{9042, 9160},
			Protocol:    "cassandra",
			Description: "Cassandra CQL options request",
			Timeout:     10,
			Rarity:      5,
		},
		
		// Neo4j探测
		{
			Name:        "Neo4jBolt",
			Type:        ProbeTypeTCP,
			Payload:     generateNeo4jBoltHandshake(),
			PayloadHex:  hex.EncodeToString(generateNeo4jBoltHandshake()),
			Ports:       []int{7687, 7474},
			Protocol:    "neo4j",
			Description: "Neo4j Bolt protocol handshake",
			Timeout:     10,
			Rarity:      5,
		},
		
		// IoT协议探测
		// CoAP探测
		{
			Name:        "CoAPPing",
			Type:        ProbeTypeUDP,
			Payload:     generateCoAPPingPacket(),
			PayloadHex:  hex.EncodeToString(generateCoAPPingPacket()),
			Ports:       []int{5683, 5684},
			Protocol:    "coap",
			Description: "CoAP ping request",
			Timeout:     5,
			Rarity:      4,
		},
		
		// LoRaWAN探测
		{
			Name:        "LoRaWANGateway",
			Type:        ProbeTypeUDP,
			Payload:     generateLoRaWANPacket(),
			PayloadHex:  hex.EncodeToString(generateLoRaWANPacket()),
			Ports:       []int{1700},
			Protocol:    "lorawan",
			Description: "LoRaWAN gateway protocol",
			Timeout:     5,
			Rarity:      6,
		},
		
		// AMQP探测
		{
			Name:        "AMQPProtocolHeader",
			Type:        ProbeTypeTCP,
			Payload:     []byte("AMQP\x00\x00\x09\x01"),
			PayloadHex:  "414d515000000901",
			Ports:       []int{5672, 5671},
			Protocol:    "amqp",
			Description: "AMQP protocol header",
			Timeout:     10,
			Rarity:      5,
		},
		
		// 企业基础设施协议
		// LDAP探测
		{
			Name:        "LDAPBind",
			Type:        ProbeTypeTCP,
			Payload:     generateLDAPBindPacket(),
			PayloadHex:  hex.EncodeToString(generateLDAPBindPacket()),
			Ports:       []int{389, 636},
			Protocol:    "ldap",
			Description: "LDAP bind request",
			Timeout:     10,
			Rarity:      4,
		},
		
		// Kerberos探测
		{
			Name:        "KerberosASReq",
			Type:        ProbeTypeUDP,
			Payload:     generateKerberosASReqPacket(),
			PayloadHex:  hex.EncodeToString(generateKerberosASReqPacket()),
			Ports:       []int{88, 464},
			Protocol:    "kerberos",
			Description: "Kerberos AS-REQ",
			Timeout:     5,
			Rarity:      5,
		},
		
		// RADIUS探测
		{
			Name:        "RADIUSAccessRequest",
			Type:        ProbeTypeUDP,
			Payload:     generateRADIUSAccessRequestPacket(),
			PayloadHex:  hex.EncodeToString(generateRADIUSAccessRequestPacket()),
			Ports:       []int{1812, 1813},
			Protocol:    "radius",
			Description: "RADIUS Access-Request",
			Timeout:     5,
			Rarity:      5,
		},
		
		// NTP探测
		{
			Name:        "NTPRequest",
			Type:        ProbeTypeUDP,
			Payload:     generateNTPRequestPacket(),
			PayloadHex:  hex.EncodeToString(generateNTPRequestPacket()),
			Ports:       []int{123},
			Protocol:    "ntp",
			Description: "NTP time request",
			Timeout:     5,
			Rarity:      3,
		},
		
		// Syslog探测
		{
			Name:        "SyslogMessage",
			Type:        ProbeTypeUDP,
			Payload:     []byte("<34>1 2024-01-01T00:00:00Z localhost NetworkProbe - - - Test message"),
			PayloadHex:  "3c33343e3120323032342d30312d30315430303a30303a30305a206c6f63616c686f7374204e6574776f726b50726f6265202d202d202d2054657374206d657373616765",
			Ports:       []int{514, 6514},
			Protocol:    "syslog",
			Description: "Syslog test message",
			Timeout:     5,
			Rarity:      4,
		},
		
		// 安全协议
		// OpenVPN探测
		{
			Name:        "OpenVPNHandshake",
			Type:        ProbeTypeUDP,
			Payload:     generateOpenVPNHandshakePacket(),
			PayloadHex:  hex.EncodeToString(generateOpenVPNHandshakePacket()),
			Ports:       []int{1194},
			Protocol:    "openvpn",
			Description: "OpenVPN handshake",
			Timeout:     5,
			Rarity:      5,
		},
		
		// WireGuard探测
		{
			Name:        "WireGuardHandshake",
			Type:        ProbeTypeUDP,
			Payload:     generateWireGuardHandshakePacket(),
			PayloadHex:  hex.EncodeToString(generateWireGuardHandshakePacket()),
			Ports:       []int{51820},
			Protocol:    "wireguard",
			Description: "WireGuard handshake initiation",
			Timeout:     5,
			Rarity:      6,
		},
		
		// 电信协议
		// SIP探测
		{
			Name:        "SIPOptions",
			Type:        ProbeTypeTCP,
			Payload:     []byte("OPTIONS sip:user@localhost SIP/2.0\r\nVia: SIP/2.0/TCP localhost:5060;branch=z9hG4bK123456\r\nFrom: <sip:probe@localhost>;tag=123456\r\nTo: <sip:user@localhost>\r\nCall-ID: 123456@localhost\r\nCSeq: 1 OPTIONS\r\nContent-Length: 0\r\n\r\n"),
			PayloadHex:  "4f5054494f4e532073697a3a75736572406c6f63616c686f73742053495a2f322e300d0a5669613a2053495a2f322e302f544350206c6f63616c686f73743a353036303b6272616e63683d7a39684734624b3132333435360d0a46726f6d3a203c7369703a70726f6265406c6f63616c686f73743e3b7461673d3132333435360d0a546f3a203c7369703a75736572406c6f63616c686f73743e0d0a43616c6c2d49443a203132333435364c6f63616c686f73740d0a435365713a2031204f5054494f4e530d0a436f6e74656e742d4c656e6774683a20300d0a0d0a",
			Ports:       []int{5060, 5061},
			Protocol:    "sip",
			Description: "SIP OPTIONS request",
			Timeout:     10,
			Rarity:      4,
		},
		
		// 云服务协议
		// Docker API探测
		{
			Name:        "DockerVersion",
			Type:        ProbeTypeTCP,
			Payload:     []byte("GET /version HTTP/1.1\r\nHost: localhost\r\nUser-Agent: NetworkProbe/1.0\r\n\r\n"),
			PayloadHex:  "474554202f76657273696f6e20485454502f312e310d0a486f73743a206c6f63616c686f73740d0a557365722d4167656e743a204e6574776f726b50726f62652f312e300d0a0d0a",
			Ports:       []int{2375, 2376},
			Protocol:    "docker",
			Description: "Docker API version check",
			Timeout:     10,
			Rarity:      5,
		},
		
		// Kubernetes API探测
		{
			Name:        "KubernetesAPI",
			Type:        ProbeTypeTCP,
			Payload:     []byte("GET /api/v1 HTTP/1.1\r\nHost: localhost\r\nUser-Agent: NetworkProbe/1.0\r\n\r\n"),
			PayloadHex:  "474554202f6170692f763120485454502f312e310d0a486f73743a206c6f63616c686f73740d0a557365722d4167656e743a204e6574776f726b50726f62652f312e300d0a0d0a",
			Ports:       []int{6443, 8080},
			Protocol:    "kubernetes",
			Description: "Kubernetes API server",
			Timeout:     10,
			Rarity:      5,
		},
	}
	
	// 转换为map
	for _, probe := range builtinProbes {
		pl.probes[probe.Name] = probe
	}
	
	return pl.probes
}

// ParseNmapProbe 解析Nmap格式的探测
func (pl *ProbeLoader) ParseNmapProbe(line string) (*Probe, error) {
	// 解析 Probe TCP GetRequest q|GET / HTTP/1.0\r\n\r\n|
	if !strings.HasPrefix(line, "Probe ") {
		return nil, fmt.Errorf("not a probe line")
	}
	
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid probe format")
	}
	
	protocol := strings.ToUpper(parts[1])
	probeName := parts[2]
	
	probe := &Probe{
		Name:        probeName,
		Type:        ProbeType(protocol),
		Protocol:    strings.ToLower(protocol),
		Description: fmt.Sprintf("Nmap %s probe", probeName),
		Timeout:     10,
		Rarity:      5,
		Ports:       []int{},
	}
	
	// 解析payload
	if len(parts) > 3 {
		payloadStr := strings.Join(parts[3:], " ")
		payload, err := pl.parsePayload(payloadStr)
		if err == nil {
			probe.Payload = payload
			probe.PayloadHex = hex.EncodeToString(payload)
		}
	}
	
	return probe, nil
}

// parsePayload 解析payload字符串
func (pl *ProbeLoader) parsePayload(payloadStr string) ([]byte, error) {
	// 处理 q|payload| 格式
	if strings.HasPrefix(payloadStr, "q|") && strings.HasSuffix(payloadStr, "|") {
		content := payloadStr[2 : len(payloadStr)-1]
		return pl.unescapePayload(content), nil
	}
	
	return []byte(payloadStr), nil
}

// unescapePayload 处理转义字符
func (pl *ProbeLoader) unescapePayload(payload string) []byte {
	// 处理常见转义字符
	payload = strings.ReplaceAll(payload, "\\r", "\r")
	payload = strings.ReplaceAll(payload, "\\n", "\n")
	payload = strings.ReplaceAll(payload, "\\t", "\t")
	payload = strings.ReplaceAll(payload, "\\0", "\x00")
	payload = strings.ReplaceAll(payload, "\\\\", "\\")
	
	// 处理十六进制转义 \x##
	hexRe := regexp.MustCompile(`\\x([0-9a-fA-F]{2})`)
	payload = hexRe.ReplaceAllStringFunc(payload, func(match string) string {
		hex := match[2:]
		if val, err := strconv.ParseInt(hex, 16, 8); err == nil {
			return string(byte(val))
		}
		return match
	})
	
	return []byte(payload)
}

// GetProbe 获取指定探测
func (pl *ProbeLoader) GetProbe(name string) (*Probe, bool) {
	probe, exists := pl.probes[name]
	return probe, exists
}

// GetProbesByPort 获取指定端口的探测
func (pl *ProbeLoader) GetProbesByPort(port int) []*Probe {
	var probes []*Probe
	
	for _, probe := range pl.probes {
		for _, p := range probe.Ports {
			if p == port {
				probes = append(probes, probe)
				break
			}
		}
	}
	
	// 按稀有度排序（稀有度低的优先）
	for i := 0; i < len(probes)-1; i++ {
		for j := i + 1; j < len(probes); j++ {
			if probes[i].Rarity > probes[j].Rarity {
				probes[i], probes[j] = probes[j], probes[i]
			}
		}
	}
	
	return probes
}

// GetAllProbes 获取所有探测
func (pl *ProbeLoader) GetAllProbes() map[string]*Probe {
	return pl.probes
}

// GetProbesByProtocol 获取指定协议的探测
func (pl *ProbeLoader) GetProbesByProtocol(protocol string) []*Probe {
	var probes []*Probe
	
	for _, probe := range pl.probes {
		if probe.Protocol == strings.ToLower(protocol) {
			probes = append(probes, probe)
		}
	}
	
	return probes
}

// generateTLSClientHello 生成TLS Client Hello握手包
func generateTLSClientHello() []byte {
	// 简化的TLS 1.2 Client Hello包
	// 这是一个基本的Client Hello，用于触发服务器响应
	clientHello := []byte{
		// TLS Record Header
		0x16,       // Content Type: Handshake (22)
		0x03, 0x01, // Version: TLS 1.0 (for compatibility)
		0x00, 0x9c, // Length: 156 bytes
		
		// Handshake Header
		0x01,             // Handshake Type: Client Hello (1)
		0x00, 0x00, 0x98, // Length: 152 bytes
		
		// Client Hello
		0x03, 0x03, // Version: TLS 1.2
		
		// Random (32 bytes) - 当前时间戳 + 28字节随机数
		0x63, 0x82, 0x0a, 0x1c, // Unix timestamp (example)
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
		0x19, 0x1a, 0x1b, 0x1c, // 28 bytes random
		
		// Session ID Length
		0x00, // No session ID
		
		// Cipher Suites Length
		0x00, 0x20, // 32 bytes (16 cipher suites)
		
		// Cipher Suites (常见的安全套件)
		0xc0, 0x2c, // TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
		0xc0, 0x30, // TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
		0x00, 0x9f, // TLS_DHE_RSA_WITH_AES_256_GCM_SHA384
		0xcc, 0xa9, // TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256
		0xcc, 0xa8, // TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
		0xcc, 0xaa, // TLS_DHE_RSA_WITH_CHACHA20_POLY1305_SHA256
		0xc0, 0x2b, // TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
		0xc0, 0x2f, // TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
		0x00, 0x9e, // TLS_DHE_RSA_WITH_AES_128_GCM_SHA256
		0xc0, 0x24, // TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384
		0xc0, 0x28, // TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384
		0x00, 0x6b, // TLS_DHE_RSA_WITH_AES_256_CBC_SHA256
		0xc0, 0x23, // TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256
		0xc0, 0x27, // TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256
		0x00, 0x67, // TLS_DHE_RSA_WITH_AES_128_CBC_SHA256
		0xc0, 0x0a, // TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA
		
		// Compression Methods Length
		0x01, // 1 method
		0x00, // No compression
		
		// Extensions Length
		0x00, 0x49, // 73 bytes
		
		// Server Name Indication (SNI) Extension
		0x00, 0x00, // Extension Type: server_name (0)
		0x00, 0x0e, // Extension Length: 14
		0x00, 0x0c, // Server Name List Length: 12
		0x00,       // Name Type: host_name (0)
		0x00, 0x09, // Host Name Length: 9
		0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x68, 0x6f, 0x73, 0x74, // "localhost"
		
		// Supported Groups Extension
		0x00, 0x0a, // Extension Type: supported_groups (10)
		0x00, 0x08, // Extension Length: 8
		0x00, 0x06, // Supported Groups List Length: 6
		0x00, 0x1d, // secp256r1
		0x00, 0x17, // secp256r1
		0x00, 0x18, // secp384r1
		
		// EC Point Formats Extension
		0x00, 0x0b, // Extension Type: ec_point_formats (11)
		0x00, 0x02, // Extension Length: 2
		0x01,       // EC Point Formats Length: 1
		0x00,       // uncompressed
		
		// Signature Algorithms Extension
		0x00, 0x0d, // Extension Type: signature_algorithms (13)
		0x00, 0x20, // Extension Length: 32
		0x00, 0x1e, // Signature Hash Algorithms Length: 30
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
	}
	
	return clientHello
}
// generateMQTTConnectPacket 生成MQTT CONNECT数据包
func generateMQTTConnectPacket() []byte {
	// MQTT CONNECT包结构:
	// Fixed Header: [Message Type + Flags:1][Remaining Length:1-4]
	// Variable Header: [Protocol Name][Protocol Level][Connect Flags][Keep Alive]
	// Payload: [Client ID][Will Topic][Will Message][User Name][Password]
	
	// 构建MQTT 3.1.1 CONNECT包
	var packet []byte
	
	// Fixed Header
	// Message Type: CONNECT (1), DUP=0, QoS=0, RETAIN=0
	packet = append(packet, 0x10) // 0001 0000
	
	// Variable Header
	var variableHeader []byte
	
	// Protocol Name: "MQTT" (4字节长度 + 4字节内容)
	variableHeader = append(variableHeader, 0x00, 0x04) // Length MSB, LSB
	variableHeader = append(variableHeader, 'M', 'Q', 'T', 'T')
	
	// Protocol Level: 4 (MQTT 3.1.1)
	variableHeader = append(variableHeader, 0x04)
	
	// Connect Flags: Clean Session=1, others=0
	variableHeader = append(variableHeader, 0x02) // 0000 0010
	
	// Keep Alive: 60 seconds
	variableHeader = append(variableHeader, 0x00, 0x3C) // 60 seconds
	
	// Payload
	var payload []byte
	
	// Client ID: "probe_client"
	clientID := "probe_client"
	payload = append(payload, 0x00, byte(len(clientID))) // Length
	payload = append(payload, []byte(clientID)...)
	
	// 计算剩余长度 (Variable Header + Payload)
	remainingLength := len(variableHeader) + len(payload)
	
	// 编码剩余长度 (MQTT变长编码)
	remainingLengthBytes := encodeMQTTLength(remainingLength)
	packet = append(packet, remainingLengthBytes...)
	
	// 添加Variable Header和Payload
	packet = append(packet, variableHeader...)
	packet = append(packet, payload...)
	
	return packet
}

// encodeMQTTLength MQTT变长编码
func encodeMQTTLength(length int) []byte {
	var encoded []byte
	
	for {
		encodedByte := byte(length % 128)
		length = length / 128
		
		if length > 0 {
			encodedByte = encodedByte | 128
		}
		
		encoded = append(encoded, encodedByte)
		
		if length == 0 {
			break
		}
	}
	
	return encoded
}

// decodeMQTTLength MQTT变长解码
func decodeMQTTLength(data []byte) (int, int) {
	multiplier := 1
	length := 0
	index := 0
	
	for index < len(data) {
		encodedByte := data[index]
		length += int(encodedByte&127) * multiplier
		
		if (encodedByte & 128) == 0 {
			break
		}
		
		multiplier *= 128
		index++
		
		if multiplier > 128*128*128 {
			return -1, -1 // 错误：长度太大
		}
	}
	
	return length, index + 1
}
// generateONVIFDiscoveryPacket 生成ONVIF WS-Discovery探测包
func generateONVIFDiscoveryPacket() []byte {
	// ONVIF WS-Discovery Probe消息
	soapEnvelope := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:tns="http://schemas.xmlsoap.org/ws/2005/04/discovery">
    <soap:Header>
        <wsa:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</wsa:Action>
        <wsa:MessageID>urn:uuid:` + generateUUID() + `</wsa:MessageID>
        <wsa:To>urn:schemas-xmlsoap-org:ws:2005:04:discovery</wsa:To>
    </soap:Header>
    <soap:Body>
        <tns:Probe>
            <tns:Types>dn:NetworkVideoTransmitter</tns:Types>
        </tns:Probe>
    </soap:Body>
</soap:Envelope>`
	
	return []byte(soapEnvelope)
}

// generateONVIFDeviceServiceRequest 生成ONVIF设备服务请求
func generateONVIFDeviceServiceRequest() []byte {
	soapRequest := `POST /onvif/device_service HTTP/1.1
Host: 127.0.0.1
Content-Type: application/soap+xml; charset=utf-8
Content-Length: 500
SOAPAction: "http://www.onvif.org/ver10/device/wsdl/GetDeviceInformation"

<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
    <soap:Header/>
    <soap:Body>
        <tds:GetDeviceInformation/>
    </soap:Body>
</soap:Envelope>`
	
	return []byte(soapRequest)
}

// generateDahuaLoginPacket 生成大华登录探测包
func generateDahuaLoginPacket() []byte {
	// 大华私有协议登录包
	// 包头: 0xa0 (固定标识)
	// 包类型: 0x01 (登录请求)
	// 数据长度: 变长
	packet := []byte{
		0xa0, 0x00, 0x00, 0x60, // 包头和长度
		0x01, 0x00, 0x00, 0x00, // 命令类型: 登录
		0x00, 0x00, 0x00, 0x00, // 序列号
		0x00, 0x00, 0x00, 0x00, // 会话ID
		// 用户名 (32字节)
		0x61, 0x64, 0x6d, 0x69, 0x6e, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// 密码 (32字节)
		0x61, 0x64, 0x6d, 0x69, 0x6e, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// 其他字段
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	
	return packet
}

// generateUUID 生成简单的UUID
func generateUUID() string {
	return "12345678-1234-5678-9012-123456789012"
}
// generateModbusTCPPacket 生成Modbus TCP探测包
func generateModbusTCPPacket() []byte {
	// Modbus TCP ADU (Application Data Unit)
	// [Transaction ID:2][Protocol ID:2][Length:2][Unit ID:1][Function Code:1][Data:N]
	
	packet := []byte{
		// MBAP Header (Modbus Application Protocol Header)
		0x00, 0x01, // Transaction ID: 1
		0x00, 0x00, // Protocol ID: 0 (Modbus)
		0x00, 0x06, // Length: 6 bytes following
		0x01,       // Unit ID: 1
		
		// PDU (Protocol Data Unit)
		0x01,       // Function Code: Read Coils (0x01)
		0x00, 0x00, // Starting Address: 0
		0x00, 0x10, // Quantity of Coils: 16
	}
	
	return packet
}

// generateDNP3Packet 生成DNP3探测包
func generateDNP3Packet() []byte {
	// DNP3 Link Layer Frame
	// [Start:2][Length:1][Control:1][Dest:2][Src:2][CRC:2][Data:N][CRC:2]
	
	packet := []byte{
		// Link Header
		0x05, 0x64, // Start bytes
		0x05,       // Length: 5 bytes (minimum)
		0x44,       // Control: DIR=0, PRM=1, FCB=0, FCV=0, FUNC=4 (Reset Link)
		0x00, 0x00, // Destination: 0
		0x00, 0x01, // Source: 1
		0x00, 0x00, // CRC (simplified, should be calculated)
	}
	
	return packet
}

// generateBACnetWhoIsPacket 生成BACnet Who-Is探测包
func generateBACnetWhoIsPacket() []byte {
	// BACnet NPDU + APDU
	// Who-Is request for device discovery
	
	packet := []byte{
		// BVLC (BACnet Virtual Link Control)
		0x81,       // Type: BACnet/IP
		0x0A,       // Function: Original-Unicast-NPDU
		0x00, 0x0C, // Length: 12 bytes
		
		// NPDU (Network Protocol Data Unit)
		0x01,       // Version: 1
		0x00,       // Control: No destination, no source
		
		// APDU (Application Protocol Data Unit)
		0x10,       // PDU Type: Unconfirmed-REQ, Segmented: No
		0x08,       // Service Choice: Who-Is
		
		// Optional: Device Instance Range
		0x09, 0x00, // Context Tag 0: Unsigned Integer
		0x09, 0xFF, // Context Tag 1: Unsigned Integer (max)
	}
	
	return packet
}

// generateOPCUAHelloPacket 生成OPC UA Hello探测包
func generateOPCUAHelloPacket() []byte {
	// OPC UA Hello Message
	// [Message Type:3][Chunk Type:1][Message Size:4][Version:4][Receive Buffer Size:4][Send Buffer Size:4][Max Message Size:4][Max Chunk Count:4][Endpoint URL:String]
	
	endpointURL := "opc.tcp://localhost:4840"
	urlLength := len(endpointURL)
	
	packet := []byte{
		// Message Header
		'H', 'E', 'L', // Message Type: HEL
		'F',            // Chunk Type: Final
		
		// Message Size (will be calculated)
		0x00, 0x00, 0x00, 0x00, // Placeholder for message size
		
		// Hello Message Body
		0x00, 0x00, 0x00, 0x00, // Protocol Version: 0
		0x00, 0x00, 0x80, 0x00, // Receive Buffer Size: 32768
		0x00, 0x00, 0x80, 0x00, // Send Buffer Size: 32768
		0x00, 0x00, 0x00, 0x00, // Max Message Size: 0 (no limit)
		0x00, 0x00, 0x00, 0x00, // Max Chunk Count: 0 (no limit)
	}
	
	// Add Endpoint URL
	urlLengthBytes := []byte{
		byte(urlLength), byte(urlLength >> 8), byte(urlLength >> 16), byte(urlLength >> 24),
	}
	packet = append(packet, urlLengthBytes...)
	packet = append(packet, []byte(endpointURL)...)
	
	// Update message size
	messageSize := len(packet)
	packet[4] = byte(messageSize)
	packet[5] = byte(messageSize >> 8)
	packet[6] = byte(messageSize >> 16)
	packet[7] = byte(messageSize >> 24)
	
	return packet
}

// generateS7SetupPacket 生成西门子S7通信建立包
func generateS7SetupPacket() []byte {
	// S7 Communication Setup (COTP Connection Request)
	// Based on ISO 8073 (COTP) and RFC 1006 (TPKT)
	
	packet := []byte{
		// TPKT Header (RFC 1006)
		0x03, 0x00, // Version: 3, Reserved: 0
		0x00, 0x16, // Length: 22 bytes
		
		// COTP Header (ISO 8073)
		0x11,       // Length: 17 bytes
		0xE0,       // PDU Type: Connection Request (CR)
		0x00, 0x00, // Destination Reference: 0
		0x00, 0x01, // Source Reference: 1
		0x00,       // Class and Option: Class 0
		
		// COTP Parameters
		0xC1, 0x02, 0x01, 0x00, // Parameter: TPDU Size (256 bytes)
		0xC2, 0x02, 0x01, 0x02, // Parameter: Source TSAP
		0xC0, 0x01, 0x0A,       // Parameter: Destination TSAP
	}
	
	return packet
}
// generateSQLServerLoginPacket 生成SQL Server登录包
func generateSQLServerLoginPacket() []byte {
	// SQL Server TDS (Tabular Data Stream) Login7 packet
	// 这是一个简化的Login7包，用于触发服务器响应
	packet := []byte{
		// TDS Header
		0x10,       // Type: Login7
		0x01,       // Status: EOM (End of Message)
		0x00, 0x34, // Length: 52 bytes
		0x00, 0x00, // SPID: 0
		0x00,       // PacketID: 0
		0x00,       // Window: 0
		
		// Login7 Data
		0x00, 0x00, 0x00, 0x00, // Length: will be calculated
		0x71, 0x00, 0x00, 0x00, // TDS Version: 7.1
		0x00, 0x10, 0x00, 0x00, // Packet Size: 4096
		0x00, 0x00, 0x00, 0x00, // Client Program Version
		0x00, 0x00, 0x00, 0x00, // Client PID
		0x00, 0x00, 0x00, 0x00, // Connection ID
		0x00,                   // Option Flags 1
		0x00,                   // Option Flags 2
		0x00,                   // Type Flags
		0x00,                   // Option Flags 3
		0x00, 0x00, 0x00, 0x00, // Client Time Zone
		0x00, 0x00, 0x00, 0x00, // Client LCID
		
		// Offsets and lengths for variable data (simplified)
		0x5e, 0x00, // Hostname offset
		0x00, 0x00, // Hostname length
		0x5e, 0x00, // Username offset  
		0x00, 0x00, // Username length
		0x5e, 0x00, // Password offset
		0x00, 0x00, // Password length
		0x5e, 0x00, // App name offset
		0x00, 0x00, // App name length
		0x5e, 0x00, // Server name offset
		0x00, 0x00, // Server name length
		0x5e, 0x00, // Library name offset
		0x00, 0x00, // Library name length
		0x5e, 0x00, // Language offset
		0x00, 0x00, // Language length
		0x5e, 0x00, // Database offset
		0x00, 0x00, // Database length
	}
	
	return packet
}

// generateOracleConnectPacket 生成Oracle TNS连接包
func generateOracleConnectPacket() []byte {
	// Oracle TNS (Transparent Network Substrate) Connect packet
	connectData := "(CONNECT_DATA=(SERVICE_NAME=XE))"
	
	packet := []byte{
		// TNS Header
		0x00, 0x3A, // Packet Length (58 bytes)
		0x00, 0x00, // Packet Checksum
		0x01,       // Packet Type: Connect
		0x00,       // Reserved
		0x00, 0x00, // Header Checksum
		
		// Connect Data
		0x01, 0x36, // Version: 310
		0x01, 0x2C, // Version (compatible)
		0x00, 0x00, // Service Options
		0x08, 0x00, // Session Data Unit Size
		0x7F, 0xFF, // Maximum Transmission Unit
		0x7F, 0x08, // NT Protocol Characteristics
		0x00, 0x00, // Line Turnaround Value
		0x00, 0x01, // Value of 1 in Hardware
		0x00, 0x1D, // Length of Connect Data
	}
	
	// Add connect data
	packet = append(packet, []byte(connectData)...)
	
	// Update packet length
	totalLen := len(packet)
	packet[0] = byte(totalLen >> 8)
	packet[1] = byte(totalLen)
	
	return packet
}

// generateMongoDBHandshakePacket 生成MongoDB握手包
func generateMongoDBHandshakePacket() []byte {
	// MongoDB Wire Protocol OP_QUERY for isMaster command
	// This is used for handshake and server discovery
	
	// Message Header
	messageLength := int32(60) // Will be updated
	requestID := int32(1)
	responseTo := int32(0)
	opCode := int32(2004) // OP_QUERY
	
	// Query message
	flags := int32(0)
	fullCollectionName := "admin.$cmd\x00"
	numberToSkip := int32(0)
	numberToReturn := int32(1)
	
	// isMaster command document (BSON)
	query := []byte{
		0x17, 0x00, 0x00, 0x00, // Document length: 23 bytes
		0x10,                   // Type: int32
		'i', 's', 'M', 'a', 's', 't', 'e', 'r', 0x00, // Field name: "isMaster"
		0x01, 0x00, 0x00, 0x00, // Value: 1
		0x00, // End of document
	}
	
	// Build packet
	packet := make([]byte, 0, 100)
	
	// Add message header (16 bytes)
	packet = append(packet, byte(messageLength), byte(messageLength>>8), byte(messageLength>>16), byte(messageLength>>24))
	packet = append(packet, byte(requestID), byte(requestID>>8), byte(requestID>>16), byte(requestID>>24))
	packet = append(packet, byte(responseTo), byte(responseTo>>8), byte(responseTo>>16), byte(responseTo>>24))
	packet = append(packet, byte(opCode), byte(opCode>>8), byte(opCode>>16), byte(opCode>>24))
	
	// Add query message
	packet = append(packet, byte(flags), byte(flags>>8), byte(flags>>16), byte(flags>>24))
	packet = append(packet, []byte(fullCollectionName)...)
	packet = append(packet, byte(numberToSkip), byte(numberToSkip>>8), byte(numberToSkip>>16), byte(numberToSkip>>24))
	packet = append(packet, byte(numberToReturn), byte(numberToReturn>>8), byte(numberToReturn>>16), byte(numberToReturn>>24))
	packet = append(packet, query...)
	
	// Update message length
	actualLength := int32(len(packet))
	packet[0] = byte(actualLength)
	packet[1] = byte(actualLength >> 8)
	packet[2] = byte(actualLength >> 16)
	packet[3] = byte(actualLength >> 24)
	
	return packet
}

// generateCassandraOptionsPacket 生成Cassandra OPTIONS包
func generateCassandraOptionsPacket() []byte {
	// Cassandra Native Protocol v4 OPTIONS request
	packet := []byte{
		0x04,       // Version: 4
		0x00,       // Flags: 0
		0x00, 0x01, // Stream ID: 1
		0x05,       // Opcode: OPTIONS
		0x00, 0x00, 0x00, 0x00, // Length: 0 (no body)
	}
	
	return packet
}

// generateNeo4jBoltHandshake 生成Neo4j Bolt握手包
func generateNeo4jBoltHandshake() []byte {
	// Neo4j Bolt Protocol handshake
	// Magic preamble + supported versions
	packet := []byte{
		// Magic preamble
		0x60, 0x60, 0xB0, 0x17,
		
		// Supported protocol versions (4 versions, 4 bytes each)
		0x00, 0x00, 0x04, 0x04, // Version 4.4
		0x00, 0x00, 0x04, 0x03, // Version 4.3
		0x00, 0x00, 0x04, 0x02, // Version 4.2
		0x00, 0x00, 0x04, 0x01, // Version 4.1
	}
	
	return packet
}

// generateCoAPPingPacket 生成CoAP ping包
func generateCoAPPingPacket() []byte {
	// CoAP (Constrained Application Protocol) ping message
	// This is a CON (Confirmable) Empty message
	packet := []byte{
		0x40,       // Version: 1, Type: CON (0), Token Length: 0
		0x00,       // Code: 0.00 (Empty)
		0x00, 0x01, // Message ID: 1
		// No token, no options, no payload
	}
	
	return packet
}

// generateLoRaWANPacket 生成LoRaWAN网关包
func generateLoRaWANPacket() []byte {
	// LoRaWAN Semtech UDP Protocol
	// PULL_DATA packet for gateway communication
	packet := []byte{
		0x02,       // Protocol Version: 2
		0x00, 0x01, // Random Token
		0x02,       // PULL_DATA identifier
		// Gateway EUI (8 bytes) - using example EUI
		0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
	}
	
	return packet
}

// generateLDAPBindPacket 生成LDAP绑定包
func generateLDAPBindPacket() []byte {
	// LDAP Bind Request (Simple Authentication)
	// ASN.1 BER encoded
	packet := []byte{
		0x30, 0x0c, // SEQUENCE, length 12
		0x02, 0x01, 0x01, // messageID: 1
		0x60, 0x07, // bindRequest SEQUENCE, length 7
		0x02, 0x01, 0x03, // version: 3
		0x04, 0x00, // name: empty (anonymous)
		0x80, 0x00, // simple authentication, empty password
	}
	
	return packet
}

// generateKerberosASReqPacket 生成Kerberos AS-REQ包
func generateKerberosASReqPacket() []byte {
	// Simplified Kerberos AS-REQ (Authentication Server Request)
	// This is a minimal packet to trigger a response
	packet := []byte{
		0x6a, 0x81, 0x80, // APPLICATION 10, length 128 (AS-REQ)
		0x30, 0x7d,       // SEQUENCE, length 125
		0xa1, 0x03,       // [1] pvno
		0x02, 0x01, 0x05, // INTEGER 5 (Kerberos v5)
		0xa2, 0x03,       // [2] msg-type  
		0x02, 0x01, 0x0a, // INTEGER 10 (AS-REQ)
		0xa4, 0x11,       // [4] req-body
		0x30, 0x0f,       // SEQUENCE
		0xa0, 0x07,       // [0] kdc-options
		0x03, 0x05, 0x00, 0x40, 0x00, 0x00, 0x10, // BIT STRING
		0xa2, 0x04,       // [2] cname
		0x30, 0x02,       // SEQUENCE  
		0x04, 0x00,       // OCTET STRING (empty principal)
	}
	
	return packet
}

// generateRADIUSAccessRequestPacket 生成RADIUS Access-Request包
func generateRADIUSAccessRequestPacket() []byte {
	// RADIUS Access-Request packet
	packet := []byte{
		0x01,       // Code: Access-Request
		0x01,       // Identifier: 1
		0x00, 0x14, // Length: 20 bytes (header only)
		// Request Authenticator (16 bytes of zeros for simplicity)
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	
	return packet
}

// generateNTPRequestPacket 生成NTP请求包
func generateNTPRequestPacket() []byte {
	// NTP (Network Time Protocol) request packet
	packet := []byte{
		0x1b,       // LI: 0, VN: 3, Mode: 3 (client)
		0x00,       // Stratum: 0
		0x00,       // Poll: 0
		0x00,       // Precision: 0
		0x00, 0x00, 0x00, 0x00, // Root Delay
		0x00, 0x00, 0x00, 0x00, // Root Dispersion
		0x00, 0x00, 0x00, 0x00, // Reference ID
		// Reference Timestamp (8 bytes)
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// Origin Timestamp (8 bytes)
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// Receive Timestamp (8 bytes)
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// Transmit Timestamp (8 bytes) - current time would go here
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	
	return packet
}

// generateOpenVPNHandshakePacket 生成OpenVPN握手包
func generateOpenVPNHandshakePacket() []byte {
	// OpenVPN UDP handshake initiation
	// P_CONTROL_HARD_RESET_CLIENT_V2
	packet := []byte{
		0x38,       // Opcode: P_CONTROL_HARD_RESET_CLIENT_V2, Key ID: 0
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, // Session ID (8 bytes)
		0x00, 0x00, 0x00, 0x00, // Packet ID (4 bytes)
		// No additional data for initial handshake
	}
	
	return packet
}

// generateWireGuardHandshakePacket 生成WireGuard握手包
func generateWireGuardHandshakePacket() []byte {
	// WireGuard Handshake Initiation message
	packet := []byte{
		0x01, 0x00, 0x00, 0x00, // Message Type: Handshake Initiation (1)
		0x00, 0x00, 0x00, 0x00, // Reserved (4 bytes)
		// Sender Index (4 bytes) - random
		0x12, 0x34, 0x56, 0x78,
		// Unencrypted Ephemeral (32 bytes) - would be actual curve25519 public key
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// Encrypted Static (48 bytes) - would be encrypted static public key + auth tag
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// Encrypted Timestamp (28 bytes) - would be encrypted timestamp + auth tag
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}
	
	return packet
}