package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"

	"honeypot-system/internal/device"
	"honeypot-system/internal/logger"
	"honeypot-system/internal/packet"
	"honeypot-system/internal/session"
	"honeypot-system/internal/web"
)

func main() {
	// 初始化配置
	initConfig()

	// 初始化日志
	logPath := viper.GetString("honeypot.log_path")
	logLevel := viper.GetString("honeypot.log_level")
	log := logger.NewLogger(logPath, logLevel)
	log.Info("Starting Industrial Protocol Honeypot System...")

	// 初始化设备指纹识别
	fpManager := device.NewFingerprintManager(
		viper.GetString("device_fingerprint.db_path"),
		viper.GetBool("device_fingerprint.rules.user_agent_analysis"),
		viper.GetBool("device_fingerprint.rules.ja3_fingerprinting"),
		viper.GetBool("device_fingerprint.rules.tcp_window_scaling"),
		viper.GetBool("device_fingerprint.rules.tls_extensions"),
	)
	defer fpManager.Close()

	// 初始化会话管理器
	sessionManager := session.NewSessionManager(
		log,
		viper.GetInt("session.timeout"),
		viper.GetInt("session.cleanup_interval"),
	)
	defer sessionManager.Stop()

	// 启动会话管理器
	sessionManager.Start()

	// 初始化数据包捕获
	// 注意：数据包捕获需要CGO_ENABLED=1
	var packetCapture *packet.PacketCapture

	// 检查是否启用了数据包捕获
	if viper.GetBool("packet_capture.enabled") {
		packetCapture = packet.NewPacketCapture(
			viper.GetStringSlice("packet_capture.interfaces"),
			viper.GetIntSlice("packet_capture.ports"),
			viper.GetBool("packet_capture.full_capture"),
			viper.GetString("packet_capture.save_path"),
			fpManager,
			sessionManager,
			log,
		)
		defer packetCapture.Stop()

		// 启动数据包捕获
		if err := packetCapture.Start(); err != nil {
			log.Error(fmt.Sprintf("Failed to start packet capture: %v", err))
			// 数据包捕获失败不影响系统启动
			log.Warn("Continuing without packet capture")
		}
	}

	// 初始化Web服务器
	webServer := web.NewServer(
		viper.GetString("web.host"),
		viper.GetInt("web.port"),
		viper.GetBool("web.https"),
		viper.GetInt("web.session_timeout"),
		fpManager,
		sessionManager,
		log,
	)
	defer webServer.Stop()

	// 启动Web服务器
	if err := webServer.Start(); err != nil {
		log.Error(fmt.Sprintf("Failed to start web server: %v", err))
		os.Exit(1)
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down honeypot system...")

	// 优雅关闭
	time.Sleep(2 * time.Second)

	log.Info("Honeypot system stopped successfully")
}

// initConfig 初始化配置
func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Fatal error reading config file: %v", err)
	}

	// 设置默认值
	viper.SetDefault("honeypot.log_level", "info")
	viper.SetDefault("honeypot.log_path", "./logs")
	viper.SetDefault("web.port", 8080)
	viper.SetDefault("web.host", "0.0.0.0")
}