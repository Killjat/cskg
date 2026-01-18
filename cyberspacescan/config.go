package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
	Scan struct {
		Workers int `yaml:"workers"`
		Timeout int `yaml:"timeout"`
		Retry   int `yaml:"retry"`
	} `yaml:"scan"`

	Ports struct {
		TCP []int `yaml:"tcp"`
		UDP []int `yaml:"udp"`
	} `yaml:"ports"`

	Output struct {
		Directory    string `yaml:"directory"`
		Format       string `yaml:"format"`
		SaveResponse bool   `yaml:"save_response"`
	} `yaml:"output"`

	Targets struct {
		File string `yaml:"file"`
	} `yaml:"targets"`
}

// LoadConfig 加载配置文件
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
