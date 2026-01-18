package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConfig 加载配置文件
func LoadConfig(filePath string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法读取配置文件: %v", err)
	}

	// 解析YAML配置
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("无法解析配置文件: %v", err)
	}

	return &config, nil
}

// GetNodeByID 根据ID获取节点信息
func GetNodeByID(config *Config, nodeID string) (*Node, error) {
	for _, node := range config.Nodes {
		if node.ID == nodeID {
			return &node, nil
		}
	}
	return nil, fmt.Errorf("找不到节点: %s", nodeID)
}
