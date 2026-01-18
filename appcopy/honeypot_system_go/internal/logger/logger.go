package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	// DebugLevel 调试级别
	DebugLevel LogLevel = iota
	// InfoLevel 信息级别
	InfoLevel
	// WarnLevel 警告级别
	WarnLevel
	// ErrorLevel 错误级别
	ErrorLevel
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	levels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	if int(l) < len(levels) {
		return levels[l]
	}
	return "UNKNOWN"
}

// Logger 日志接口
type Logger interface {
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	Close()
}

// logger 日志实现
 type logger struct {
	logFile   *os.File
	debugLog  *log.Logger
	infoLog   *log.Logger
	warnLog   *log.Logger
	errorLog  *log.Logger
	logLevel  LogLevel
	mutex     sync.Mutex
}

// NewLogger 创建新的日志实例
func NewLogger(logPath, logLevel string) Logger {
	// 创建日志目录
	os.MkdirAll(logPath, 0755)

	// 日志文件名
	logFileName := filepath.Join(logPath, fmt.Sprintf("honeypot_%s.log", time.Now().Format("2006-01-02")))

	// 打开日志文件
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// 解析日志级别
	level := InfoLevel
	switch logLevel {
	case "debug":
		level = DebugLevel
	case "info":
		level = InfoLevel
	case "warn":
		level = WarnLevel
	case "error":
		level = ErrorLevel
	}

	return &logger{
		logFile:  logFile,
		debugLog: log.New(logFile, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile),
		infoLog:  log.New(logFile, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLog:  log.New(logFile, "[WARN] ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLog: log.New(logFile, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
		logLevel: level,
	}
}

// Debug 记录调试日志
func (l *logger) Debug(format string, v ...interface{}) {
	if l.logLevel <= DebugLevel {
		l.mutex.Lock()
		defer l.mutex.Unlock()
		l.debugLog.Printf(format, v...)
		fmt.Fprintf(os.Stdout, "[DEBUG] %s\n", fmt.Sprintf(format, v...))
	}
}

// Info 记录信息日志
func (l *logger) Info(format string, v ...interface{}) {
	if l.logLevel <= InfoLevel {
		l.mutex.Lock()
		defer l.mutex.Unlock()
		l.infoLog.Printf(format, v...)
		fmt.Fprintf(os.Stdout, "[INFO] %s\n", fmt.Sprintf(format, v...))
	}
}

// Warn 记录警告日志
func (l *logger) Warn(format string, v ...interface{}) {
	if l.logLevel <= WarnLevel {
		l.mutex.Lock()
		defer l.mutex.Unlock()
		l.warnLog.Printf(format, v...)
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", fmt.Sprintf(format, v...))
	}
}

// Error 记录错误日志
func (l *logger) Error(format string, v ...interface{}) {
	if l.logLevel <= ErrorLevel {
		l.mutex.Lock()
		defer l.mutex.Unlock()
		l.errorLog.Printf(format, v...)
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", fmt.Sprintf(format, v...))
	}
}

// Close 关闭日志文件
func (l *logger) Close() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.logFile != nil {
		l.logFile.Close()
	}
}
