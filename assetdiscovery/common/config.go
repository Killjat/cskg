package common

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 系统配置结构体
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	Client   ClientConfig   `yaml:"client"`
	Scan     ScanConfig     `yaml:"scan"`
	Database DatabaseConfig `yaml:"database"`
	Web      WebConfig      `yaml:"web"`
}

// ServerConfig 服务端配置
type ServerConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Workers int    `yaml:"workers"`
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers     []string `yaml:"brokers"`
	TaskTopic   string   `yaml:"task_topic"`
	ResultTopic string   `yaml:"result_topic"`
	GroupID     string   `yaml:"group_id"`
}

// ClientConfig 客户端配置
type ClientConfig struct {
	ClientID     string `yaml:"client_id"`
	WorkerCount  int    `yaml:"worker_count"`
	ScanTimeout  int    `yaml:"scan_timeout"`
}

// ScanConfig 扫描配置
type ScanConfig struct {
	PortRange       string `yaml:"port_range"`
	Timeout         int    `yaml:"timeout"`
	ConcurrentScans int    `yaml:"concurrent_scans"`
	HTTPTimeout     int    `yaml:"http_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

// WebConfig Web配置
type WebConfig struct {
	TemplateDir string `yaml:"template_dir"`
	StaticDir   string `yaml:"static_dir"`
}

// LoadConfig 从文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &config, nil
}
