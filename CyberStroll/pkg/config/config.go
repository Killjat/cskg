package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	"github.com/cskg/CyberStroll/internal/kafka"
	"github.com/cskg/CyberStroll/internal/scanner"
)

// ScanNodeConfig 扫描节点配置
type ScanNodeConfig struct {
	Node     NodeConfig           `yaml:"node"`
	Kafka    kafka.KafkaConfig    `yaml:"kafka"`
	Scanner  scanner.ScannerConfig `yaml:"scanner"`
	Storage  StorageConfig        `yaml:"storage"`
	Logging  LoggingConfig        `yaml:"logging"`
}

// TaskManagerConfig 任务管理节点配置
type TaskManagerConfig struct {
	Node    NodeConfig        `yaml:"node"`
	Kafka   kafka.KafkaConfig `yaml:"kafka"`
	Storage StorageConfig     `yaml:"storage"`
	Web     WebConfig         `yaml:"web"`
	Logging LoggingConfig     `yaml:"logging"`
}

// ProcessorNodeConfig 处理节点配置
type ProcessorNodeConfig struct {
	Node          NodeConfig        `yaml:"node"`
	Kafka         kafka.KafkaConfig `yaml:"kafka"`
	Storage       StorageConfig     `yaml:"storage"`
	Elasticsearch ESConfig          `yaml:"elasticsearch"`
	Processor     ProcessorConfig   `yaml:"processor"`
	Logging       LoggingConfig     `yaml:"logging"`
	Debug         bool              `yaml:"debug"`
}

// ProcessorConfig 处理器配置
type ProcessorConfig struct {
	BatchSize       int  `yaml:"batch_size"`
	BatchTimeout    int  `yaml:"batch_timeout"`
	MaxConcurrency  int  `yaml:"max_concurrency"`
	RetryCount      int  `yaml:"retry_count"`
	EnableGeoLookup bool `yaml:"enable_geo_lookup"`
}

// SearchNodeConfig 搜索节点配置
type SearchNodeConfig struct {
	Node          NodeConfig    `yaml:"node"`
	Elasticsearch ESConfig      `yaml:"elasticsearch"`
	Web           WebConfig     `yaml:"web"`
	Logging       LoggingConfig `yaml:"logging"`
}

// EnrichmentNodeConfig 富化节点配置
type EnrichmentNodeConfig struct {
	Node          NodeConfig        `yaml:"node"`
	Elasticsearch ESConfig          `yaml:"elasticsearch"`
	Enrichment    EnrichmentConfig  `yaml:"enrichment"`
	Logging       LoggingConfig     `yaml:"logging"`
}

// EnrichmentConfig 富化配置
type EnrichmentConfig struct {
	BatchSize         int           `yaml:"batch_size"`
	WorkerCount       int           `yaml:"worker_count"`
	ScanInterval      time.Duration `yaml:"scan_interval"`
	RequestTimeout    time.Duration `yaml:"request_timeout"`
	MaxRetries        int           `yaml:"max_retries"`
	EnableCert        bool          `yaml:"enable_cert"`
	EnableAPI         bool          `yaml:"enable_api"`
	EnableWebInfo     bool          `yaml:"enable_web_info"`
	EnableFingerprint bool          `yaml:"enable_fingerprint"`
	EnableContent     bool          `yaml:"enable_content"`
}

// NodeConfig 节点基本配置
type NodeConfig struct {
	ID     string `yaml:"id"`
	Name   string `yaml:"name"`
	Region string `yaml:"region"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	MongoDB MongoConfig `yaml:"mongodb"`
}

// MongoConfig MongoDB配置
type MongoConfig struct {
	URI      string `yaml:"uri"`
	Database string `yaml:"database"`
	Timeout  int    `yaml:"timeout"`
}

// ESConfig Elasticsearch配置
type ESConfig struct {
	URLs     []string `yaml:"urls"`
	Index    string   `yaml:"index"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	Timeout  int      `yaml:"timeout"`
}

// WebConfig Web服务配置
type WebConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	TLS  struct {
		Enabled  bool   `yaml:"enabled"`
		CertFile string `yaml:"cert_file"`
		KeyFile  string `yaml:"key_file"`
	} `yaml:"tls"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    string `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// LoadScanNodeConfig 加载扫描节点配置
func LoadScanNodeConfig(configFile string) (*ScanNodeConfig, error) {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")
	
	// 设置默认值
	setDefaultScanNodeConfig()
	
	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，使用默认配置
			fmt.Printf("配置文件不存在，使用默认配置: %s\n", configFile)
		} else {
			return nil, fmt.Errorf("读取配置文件失败: %v", err)
		}
	}
	
	var config ScanNodeConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %v", err)
	}
	
	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}
	
	return &config, nil
}

// LoadTaskManagerConfig 加载任务管理节点配置
func LoadTaskManagerConfig(configFile string) (*TaskManagerConfig, error) {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")
	
	setDefaultTaskManagerConfig()
	
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("配置文件不存在，使用默认配置: %s\n", configFile)
		} else {
			return nil, fmt.Errorf("读取配置文件失败: %v", err)
		}
	}
	
	var config TaskManagerConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %v", err)
	}
	
	return &config, nil
}

// setDefaultScanNodeConfig 设置扫描节点默认配置
func setDefaultScanNodeConfig() {
	// 节点配置
	viper.SetDefault("node.id", "scan-node-001")
	viper.SetDefault("node.name", "扫描节点1")
	viper.SetDefault("node.region", "default")
	
	// Kafka配置
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.system_task_topic", "system_tasks")
	viper.SetDefault("kafka.regular_task_topic", "regular_tasks")
	viper.SetDefault("kafka.result_topic", "scan_results")
	viper.SetDefault("kafka.group_id", "scan_nodes")
	viper.SetDefault("kafka.auto_offset_reset", "latest")
	viper.SetDefault("kafka.session_timeout", 30000)
	viper.SetDefault("kafka.heartbeat_interval", 3000)
	
	// 扫描器配置
	viper.SetDefault("scanner.max_concurrency", 100)
	viper.SetDefault("scanner.timeout", "10s")
	viper.SetDefault("scanner.retry_count", 3)
	viper.SetDefault("scanner.probe_delay", "100ms")
	viper.SetDefault("scanner.enable_logging", true)
	
	// 存储配置
	viper.SetDefault("storage.mongodb.uri", "mongodb://localhost:27017")
	viper.SetDefault("storage.mongodb.database", "cyberstroll")
	viper.SetDefault("storage.mongodb.timeout", 10)
	
	// 日志配置
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", "logs/scan_node.log")
	viper.SetDefault("logging.max_size", "100MB")
	viper.SetDefault("logging.max_backups", 10)
	viper.SetDefault("logging.max_age", 30)
	viper.SetDefault("logging.compress", true)
}

// setDefaultTaskManagerConfig 设置任务管理节点默认配置
func setDefaultTaskManagerConfig() {
	// 节点配置
	viper.SetDefault("node.id", "task-manager-001")
	viper.SetDefault("node.name", "任务管理节点1")
	viper.SetDefault("node.region", "default")
	
	// Kafka配置
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.system_task_topic", "system_tasks")
	viper.SetDefault("kafka.regular_task_topic", "regular_tasks")
	viper.SetDefault("kafka.result_topic", "scan_results")
	viper.SetDefault("kafka.group_id", "task_managers")
	
	// Web配置
	viper.SetDefault("web.host", "0.0.0.0")
	viper.SetDefault("web.port", 8080)
	viper.SetDefault("web.tls.enabled", false)
	
	// 存储配置
	viper.SetDefault("storage.mongodb.uri", "mongodb://localhost:27017")
	viper.SetDefault("storage.mongodb.database", "cyberstroll")
	viper.SetDefault("storage.mongodb.timeout", 10)
	
	// 日志配置
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", "logs/task_manager.log")
	viper.SetDefault("logging.max_size", "100MB")
	viper.SetDefault("logging.max_backups", 10)
	viper.SetDefault("logging.max_age", 30)
	viper.SetDefault("logging.compress", true)
}

// LoadSearchNodeConfig 加载搜索节点配置
func LoadSearchNodeConfig(configFile string) (*SearchNodeConfig, error) {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")
	
	setDefaultSearchNodeConfig()
	
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("配置文件不存在，使用默认配置: %s\n", configFile)
		} else {
			return nil, fmt.Errorf("读取配置文件失败: %v", err)
		}
	}
	
	var config SearchNodeConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %v", err)
	}
	
	return &config, nil
}

// setDefaultSearchNodeConfig 设置搜索节点默认配置
func setDefaultSearchNodeConfig() {
	// 节点配置
	viper.SetDefault("node.id", "search-node-001")
	viper.SetDefault("node.name", "搜索节点1")
	viper.SetDefault("node.region", "default")
	
	// Elasticsearch配置
	viper.SetDefault("elasticsearch.urls", []string{"http://localhost:9200"})
	viper.SetDefault("elasticsearch.index", "cyberstroll_ip_scan")
	viper.SetDefault("elasticsearch.username", "")
	viper.SetDefault("elasticsearch.password", "")
	viper.SetDefault("elasticsearch.timeout", 30)
	
	// Web配置
	viper.SetDefault("web.host", "0.0.0.0")
	viper.SetDefault("web.port", 8081)
	viper.SetDefault("web.tls.enabled", false)
	
	// 日志配置
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", "logs/search_node.log")
	viper.SetDefault("logging.max_size", "100MB")
	viper.SetDefault("logging.max_backups", 10)
	viper.SetDefault("logging.max_age", 30)
	viper.SetDefault("logging.compress", true)
}

// GetNodeID 获取节点ID
func GetNodeID() string {
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "default-node"
	}
	return nodeID
}

// GetKafkaBrokers 获取Kafka brokers
func GetKafkaBrokers() []string {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		return []string{"localhost:9092"}
	}
	return []string{brokers}
}

// GetMongoURI 获取MongoDB URI
func GetMongoURI() string {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		return "mongodb://localhost:27017"
	}
	return uri
}

// GetElasticsearchURL 获取Elasticsearch URL
func GetElasticsearchURL() []string {
	url := os.Getenv("ES_URL")
	if url == "" {
		return []string{"http://localhost:9200"}
	}
	return []string{url}
}

// validateConfig 验证配置
func validateConfig(config interface{}) error {
	// 基本验证，可以根据需要扩展
	return nil
}
// LoadConfig 通用配置加载函数
func LoadConfig(configFile string) (*TaskManagerConfig, error) {
	return LoadTaskManagerConfig(configFile)
}

// LoadProcessorNodeConfig 加载处理节点配置
func LoadProcessorNodeConfig(configFile string) (*ProcessorNodeConfig, error) {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")
	
	setDefaultProcessorNodeConfig()
	
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("配置文件不存在，使用默认配置: %s\n", configFile)
		} else {
			return nil, fmt.Errorf("读取配置文件失败: %v", err)
		}
	}
	
	var config ProcessorNodeConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %v", err)
	}
	
	return &config, nil
}

// setDefaultProcessorNodeConfig 设置处理节点默认配置
func setDefaultProcessorNodeConfig() {
	// 节点配置
	viper.SetDefault("node.id", "processor-node-001")
	viper.SetDefault("node.name", "处理节点1")
	viper.SetDefault("node.region", "default")
	
	// Kafka配置
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.result_topic", "scan_results")
	viper.SetDefault("kafka.group_id", "processors")
	
	// Elasticsearch配置
	viper.SetDefault("elasticsearch.urls", []string{"http://localhost:9200"})
	viper.SetDefault("elasticsearch.index", "cyberstroll_ip_scan")
	viper.SetDefault("elasticsearch.username", "")
	viper.SetDefault("elasticsearch.password", "")
	viper.SetDefault("elasticsearch.timeout", 30)
	
	// 存储配置
	viper.SetDefault("storage.mongodb.uri", "mongodb://localhost:27017")
	viper.SetDefault("storage.mongodb.database", "cyberstroll")
	viper.SetDefault("storage.mongodb.timeout", 10)
	
	// 日志配置
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", "logs/processor_node.log")
	viper.SetDefault("logging.max_size", "100MB")
	viper.SetDefault("logging.max_backups", 10)
	viper.SetDefault("logging.max_age", 30)
	viper.SetDefault("logging.compress", true)
}
func LoadEnrichmentNodeConfig(configFile string) (*EnrichmentNodeConfig, error) {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")
	
	setDefaultEnrichmentNodeConfig()
	
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("配置文件不存在，使用默认配置: %s\n", configFile)
		} else {
			return nil, fmt.Errorf("读取配置文件失败: %v", err)
		}
	}
	
	var config EnrichmentNodeConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %v", err)
	}
	
	return &config, nil
}

// setDefaultEnrichmentNodeConfig 设置富化节点默认配置
func setDefaultEnrichmentNodeConfig() {
	// 节点配置
	viper.SetDefault("node.id", "enrichment-node-001")
	viper.SetDefault("node.name", "网站数据富化节点1")
	viper.SetDefault("node.region", "default")
	
	// Elasticsearch配置
	viper.SetDefault("elasticsearch.urls", []string{"http://localhost:9200"})
	viper.SetDefault("elasticsearch.index", "cyberstroll_ip_scan")
	viper.SetDefault("elasticsearch.username", "")
	viper.SetDefault("elasticsearch.password", "")
	viper.SetDefault("elasticsearch.timeout", 30)
	
	// 富化配置
	viper.SetDefault("enrichment.batch_size", 50)
	viper.SetDefault("enrichment.worker_count", 5)
	viper.SetDefault("enrichment.scan_interval", "5m")
	viper.SetDefault("enrichment.request_timeout", "30s")
	viper.SetDefault("enrichment.max_retries", 3)
	viper.SetDefault("enrichment.enable_cert", true)
	viper.SetDefault("enrichment.enable_api", true)
	viper.SetDefault("enrichment.enable_web_info", true)
	viper.SetDefault("enrichment.enable_fingerprint", true)
	viper.SetDefault("enrichment.enable_content", true)
	
	// 日志配置
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", "logs/enrichment_node.log")
	viper.SetDefault("logging.max_size", "100MB")
	viper.SetDefault("logging.max_backups", 10)
	viper.SetDefault("logging.max_age", 30)
	viper.SetDefault("logging.compress", true)
}