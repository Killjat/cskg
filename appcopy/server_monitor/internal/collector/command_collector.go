package collector

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/example/server-monitor/internal/model"
)

// commandCollector 命令信息采集器实现
type commandCollector struct {
	ticker   *time.Ticker
	done     chan bool
	commands []model.CommandData
}

// NewCommandCollector 创建命令采集器实例
func NewCommandCollector(interval int) CommandCollector {
	return &commandCollector{
		ticker:   time.NewTicker(time.Duration(interval) * time.Second),
		done:     make(chan bool),
		commands: []model.CommandData{},
	}
}

// Start 启动命令采集器
func (cc *commandCollector) Start() error {
	// 初始采集一次
	cc.collectCurrentCommands()
	
	// 启动定期采集
	go func() {
		for {
			select {
			case <-cc.ticker.C:
				cc.collectCurrentCommands()
			case <-cc.done:
				return
			}
		}
	}()
	
	return nil
}

// Stop 停止命令采集器
func (cc *commandCollector) Stop() error {
	cc.done <- true
	cc.ticker.Stop()
	return nil
}

// CollectCurrentCommands 采集当前执行的命令
func (cc *commandCollector) CollectCurrentCommands() ([]model.CommandData, error) {
	return cc.commands, nil
}

// CollectCommandHistory 采集命令历史记录
func (cc *commandCollector) CollectCommandHistory() ([]model.CommandData, error) {
	// 实现读取命令历史记录的逻辑
	// 从用户的.bash_history、.zsh_history等文件中读取
	return []model.CommandData{}, nil
}

// collectCurrentCommands 采集当前执行的命令
func (cc *commandCollector) collectCurrentCommands() {
	newCommands := []model.CommandData{}
	
	// 遍历/proc目录下的所有进程，获取命令信息
	procDir := "/proc"
	entries, err := ioutil.ReadDir(procDir)
	if err != nil {
		fmt.Printf("读取/proc目录失败: %v\n", err)
		return
	}
	
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		// 检查是否为数字目录（PID）
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		
		// 读取进程的命令行信息
		cmdlinePath := filepath.Join(procDir, entry.Name(), "cmdline")
		cmdlineContent, err := ioutil.ReadFile(cmdlinePath)
		if err != nil {
			continue
		}
		
		// 跳过空命令行
		if len(cmdlineContent) == 0 {
			continue
		}
		
		// 解析命令行（以null字符分隔）
		cmdParts := strings.Split(string(cmdlineContent), "\x00")
		command := strings.Join(cmdParts, " ")
		command = strings.TrimSpace(command)
		
		// 跳过空命令
		if command == "" {
			continue
		}
		
		// 读取进程的基本信息
		statPath := filepath.Join(procDir, entry.Name(), "stat")
		statContent, err := ioutil.ReadFile(statPath)
		if err != nil {
			continue
		}
		
		statFields := strings.Fields(string(statContent))
		if len(statFields) < 22 {
			continue
		}
		
		// 获取父进程ID
		ppid, _ := strconv.Atoi(statFields[3])
		
		// 获取进程状态
		status := statFields[2]
		running := status == "R" // 只有运行状态的进程才被视为正在执行命令
		
		if running {
			// 创建命令记录
			cmd := model.CommandData{
				Command:     command,
				PID:         pid,
				PPID:        ppid,
				StartTime:   time.Now(),
				Duration:    0, // 实际实现中需要计算命令执行时长
				Status:      "running",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			
			// 尝试获取命令的用户信息
			statusPath := filepath.Join(procDir, entry.Name(), "status")
			statusContent, err := ioutil.ReadFile(statusPath)
			if err == nil {
				lines := strings.Split(string(statusContent), "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "Uid:") {
						uidFields := strings.Fields(line)
						if len(uidFields) > 1 {
							uid, _ := strconv.Atoi(uidFields[1])
							cmd.Username = cc.getUsernameByUID(uid)
						}
						break
					}
				}
			}
			
			// 添加到命令列表
			newCommands = append(newCommands, cmd)
		}
	}
	
	cc.commands = newCommands
}

// getUsernameByUID 根据UID获取用户名
func (cc *commandCollector) getUsernameByUID(uid int) string {
	// 读取/etc/passwd文件查找用户名
	passwdFile, err := os.Open("/etc/passwd")
	if err != nil {
		return strconv.Itoa(uid)
	}
	defer passwdFile.Close()
	
	// 读取文件内容
	content, err := ioutil.ReadFile("/etc/passwd")
	if err != nil {
		return strconv.Itoa(uid)
	}
	
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 3 {
			continue
		}
		
		fileUID, err := strconv.Atoi(fields[2])
		if err == nil && fileUID == uid {
			return fields[0]
		}
	}
	
	return strconv.Itoa(uid)
}
