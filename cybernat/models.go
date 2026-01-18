package main

import "time"

// Node 节点信息
type Node struct {
	ID          string    `json:"id" yaml:"id"`
	Name        string    `json:"name" yaml:"name"`
	Host        string    `json:"host" yaml:"host"`
	Port        int       `json:"port" yaml:"port"`
	User        string    `json:"user" yaml:"user"`
	Password    string    `json:"password" yaml:"password"`
	Status      string    `json:"status" yaml:"status"`
	LastContact time.Time `json:"last_contact" yaml:"last_contact"`
}

// Task 任务信息
type Task struct {
	ID        string    `json:"id" yaml:"id"`
	NodeID    string    `json:"node_id" yaml:"node_id"`
	Command   string    `json:"command" yaml:"command"`
	Args      []string  `json:"args" yaml:"args"`
	Timeout   int       `json:"timeout" yaml:"timeout"`
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
}

// TaskResult 任务结果
type TaskResult struct {
	TaskID    string    `json:"task_id" yaml:"task_id"`
	NodeID    string    `json:"node_id" yaml:"node_id"`
	Output    string    `json:"output" yaml:"output"`
	Error     string    `json:"error" yaml:"error"`
	ExitCode  int       `json:"exit_code" yaml:"exit_code"`
	Completed bool      `json:"completed" yaml:"completed"`
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`
}

// Config 配置信息
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	NATS     NATSConfig     `yaml:"nats"`
	Nodes    []Node         `yaml:"nodes"`
}

// ServerConfig 服务端配置
type ServerConfig struct {
	Port int `yaml:"port"`
}

// NATSConfig NATS配置
type NATSConfig struct {
	URL            string `yaml:"url"`
	TaskSubject    string `yaml:"task_subject"`
	ResultSubject  string `yaml:"result_subject"`
}
