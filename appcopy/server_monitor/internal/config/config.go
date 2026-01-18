package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// Config 系统配置结构体
type Config struct {
	Server struct {
		Port         int    `yaml:"port"`
		Host         string `yaml:"host"`
		ReadTimeout  int    `yaml:"read_timeout"`
		WriteTimeout int    `yaml:"write_timeout"`
	} `yaml:"server"`

	Collector struct {
		LoginInterval    int    `yaml:"login_interval"`    // 登录信息采集间隔（秒）
		ProcessInterval  int    `yaml:"process_interval"`  // 进程信息采集间隔（秒）
		CommandInterval  int    `yaml:"command_interval"`  // 命令信息采集间隔（秒）
		FileWatchPaths   []string `yaml:"file_watch_paths"` // 文件监控路径
		RecursiveWatch   bool   `yaml:"recursive_watch"`   // 是否递归监控目录
	} `yaml:"collector"`

	Storage struct {
		Type     string `yaml:"type"`     // 存储类型：sqlite, mysql, file
		FilePath string `yaml:"file_path"` // 文件存储路径（如SQLite数据库文件）
		Host     string `yaml:"host"`     // 数据库主机
		Port     int    `yaml:"port"`     // 数据库端口
		Username string `yaml:"username"` // 数据库用户名
		Password string `yaml:"password"` // 数据库密码
		Database string `yaml:"database"` // 数据库名称
	} `yaml:"storage"`

	Alert struct {
		Enabled       bool     `yaml:"enabled"`       // 是否启用告警
		Levels        []string `yaml:"levels"`        // 告警级别：info, warning, error, critical
		Notification  []string `yaml:"notification"`  // 告警通知方式：email, sms, wechat
		SMTP          struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
			From     string `yaml:"from"`
			To       string `yaml:"to"`
		} `yaml:"smtp"`
	} `yaml:"alert"`
}

// LoadConfig 从文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", filePath)
	}

	// 读取配置文件内容
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析YAML配置
	config := &Config{}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 设置默认值
	SetDefaults(config)

	return config, nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, filePath string) error {
	// 转换为YAML格式
	content, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("转换配置为YAML失败: %v", err)
	}

	// 写入文件
	err = ioutil.WriteFile(filePath, content, 0644)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// SetDefaults 设置配置默认值
func SetDefaults(config *Config) {
	// Server默认配置
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 30
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 30
	}

	// Collector默认配置
	if config.Collector.LoginInterval == 0 {
		config.Collector.LoginInterval = 5
	}
	if config.Collector.ProcessInterval == 0 {
		config.Collector.ProcessInterval = 2
	}
	if config.Collector.CommandInterval == 0 {
		config.Collector.CommandInterval = 2
	}
	if len(config.Collector.FileWatchPaths) == 0 {
		config.Collector.FileWatchPaths = []string{"/var/log", "/etc"}
	}
	// 默认递归监控目录
	// config.Collector.RecursiveWatch = true // 默认值为false，由YAML解析决定

	// Storage默认配置
	if config.Storage.Type == "" {
		config.Storage.Type = "sqlite"
	}
	if config.Storage.FilePath == "" {
		config.Storage.FilePath = "./server_monitor.db"
	}
	if config.Storage.Port == 0 {
		config.Storage.Port = 3306
	}

	// Alert默认配置
	// config.Alert.Enabled = false // 默认值为false，由YAML解析决定
	if len(config.Alert.Levels) == 0 {
		config.Alert.Levels = []string{"warning", "error", "critical"}
	}
	if len(config.Alert.Notification) == 0 {
		config.Alert.Notification = []string{"email"}
	}
}
