package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置结构体
type Config struct {
	System       SystemConfig       `mapstructure:"system"`
	Kafka        KafkaConfig        `mapstructure:"kafka"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	Scan         ScanConfig         `mapstructure:"scan"`
	Task         TaskConfig         `mapstructure:"task"`
}

// SystemConfig 系统配置
type SystemConfig struct {
	Name     string `mapstructure:"name"`
	Version  string `mapstructure:"version"`
	LogLevel string `mapstructure:"log_level"`
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers  []string          `mapstructure:"brokers"`
	GroupID  string            `mapstructure:"group_id"`
	Topics   KafkaTopicsConfig `mapstructure:"topics"`
	Consumer KafkaConsumerConfig `mapstructure:"consumer"`
	Producer KafkaProducerConfig `mapstructure:"producer"`
}

// KafkaTopicsConfig Kafka主题配置
type KafkaTopicsConfig struct {
	SystemTask string `mapstructure:"system_task"`
	NormalTask string `mapstructure:"normal_task"`
}

// KafkaConsumerConfig Kafka消费者配置
type KafkaConsumerConfig struct {
	AutoOffsetReset  string `mapstructure:"auto_offset_reset"`
	SessionTimeout   string `mapstructure:"session_timeout"`
	HeartbeatInterval string `mapstructure:"heartbeat_interval"`
}

// KafkaProducerConfig Kafka生产者配置
type KafkaProducerConfig struct {
	Acks      string `mapstructure:"acks"`
	Retries   int    `mapstructure:"retries"`
	LingerMS  int    `mapstructure:"linger_ms"`
}

// ElasticsearchConfig Elasticsearch配置
type ElasticsearchConfig struct {
	Addresses []string                `mapstructure:"addresses"`
	Username  string                  `mapstructure:"username"`
	Password  string                  `mapstructure:"password"`
	Index     ElasticsearchIndexConfig `mapstructure:"index"`
	Timeout   string                  `mapstructure:"timeout"`
}

// ElasticsearchIndexConfig Elasticsearch索引配置
type ElasticsearchIndexConfig struct {
	ScanResult string `mapstructure:"scan_result"`
}

// ScanConfig 扫描配置
type ScanConfig struct {
	Timeout     int    `mapstructure:"timeout"`
	Threads     int    `mapstructure:"threads"`
	Retries     int    `mapstructure:"retries"`
	DefaultPorts string `mapstructure:"default_ports"`
}

// TaskConfig 任务配置
type TaskConfig struct {
	MaxConcurrentTasks int `mapstructure:"max_concurrent_tasks"`
	TaskExpireTime     int `mapstructure:"task_expire_time"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// 添加配置文件搜索路径
	viper.AddConfigPath(configPath)
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// 读取环境变量
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，使用默认值
			fmt.Println("Warning: Config file not found, using default values")
		} else {
			// 配置文件存在但读取错误
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 解析配置
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 特殊处理切片类型的环境变量
	// 处理KAFKA_BROKERS环境变量
	if kafkaBrokers := os.Getenv("KAFKA_BROKERS"); kafkaBrokers != "" {
		config.Kafka.Brokers = strings.Split(kafkaBrokers, ",")
	}

	// 处理ELASTICSEARCH_ADDRESSES环境变量
	if esAddresses := os.Getenv("ELASTICSEARCH_ADDRESSES"); esAddresses != "" {
		config.Elasticsearch.Addresses = strings.Split(esAddresses, ",")
	}

	return &config, nil
}

// GetDefaultConfigPath 获取默认配置文件路径
func GetDefaultConfigPath() string {
	// 检查当前目录下是否有config目录
	if _, err := os.Stat("./config"); err == nil {
		return "./config"
	}
	return "."
}
