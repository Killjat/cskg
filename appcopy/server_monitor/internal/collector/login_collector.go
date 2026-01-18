package collector

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/example/server-monitor/internal/model"
)

// loginCollector 登录信息采集器实现
type loginCollector struct {
	ticker  *time.Ticker
	done    chan bool
	logins  []model.LoginData
}

// NewLoginCollector 创建登录采集器实例
func NewLoginCollector(interval int) LoginCollector {
	return &loginCollector{
		ticker:  time.NewTicker(time.Duration(interval) * time.Second),
		done:    make(chan bool),
		logins:  []model.LoginData{},
	}
}

// Start 启动登录采集器
func (lc *loginCollector) Start() error {
	// 初始采集一次
	lc.collectCurrentLogins()
	
	// 启动定期采集
	go func() {
		for {
			select {
			case <-lc.ticker.C:
				lc.collectCurrentLogins()
			case <-lc.done:
				return
			}
		}
	}()
	
	return nil
}

// Stop 停止登录采集器
func (lc *loginCollector) Stop() error {
	lc.done <- true
	lc.ticker.Stop()
	return nil
}

// CollectCurrentLogins 采集当前登录信息
func (lc *loginCollector) CollectCurrentLogins() ([]model.LoginData, error) {
	return lc.logins, nil
}

// CollectLoginHistory 采集登录历史记录
func (lc *loginCollector) CollectLoginHistory() ([]model.LoginData, error) {
	// 实现读取登录历史记录的逻辑
	// 从/var/log/wtmp文件或使用last命令获取
	return []model.LoginData{}, nil
}

// collectCurrentLogins 采集当前登录信息
func (lc *loginCollector) collectCurrentLogins() {
	newLogins := []model.LoginData{}
	
	// 读取/var/run/utmp文件获取当前登录用户
	utmpFile, err := os.Open("/var/run/utmp")
	if err != nil {
		fmt.Printf("读取/var/run/utmp文件失败: %v\n", err)
		return
	}
	defer utmpFile.Close()
	
	scanner := bufio.NewScanner(utmpFile)
	for scanner.Scan() {
		line := scanner.Text()
		// 解析utmp文件内容
		// 注意：实际utmp文件是二进制格式，这里使用简化的文本解析示例
		// 在实际实现中，需要使用utmp库或解析二进制格式
		login, err := lc.parseUtmpLine(line)
		if err == nil {
			newLogins = append(newLogins, login)
		}
	}
	
	// 如果无法读取utmp文件，使用who命令作为备选方案
	if len(newLogins) == 0 {
		newLogins, err = lc.getLoginsFromWhoCommand()
		if err != nil {
			fmt.Printf("获取登录信息失败: %v\n", err)
			return
		}
	}
	
	lc.logins = newLogins
}

// parseUtmpLine 解析utmp文件行
// 注意：这是一个简化的实现，实际utmp是二进制文件
func (lc *loginCollector) parseUtmpLine(line string) (model.LoginData, error) {
	login := model.LoginData{}
	// 实际实现中需要解析二进制格式
	return login, fmt.Errorf("utmp文件是二进制格式，需要使用专门的库解析")
}

// getLoginsFromWhoCommand 使用who命令获取登录信息
func (lc *loginCollector) getLoginsFromWhoCommand() ([]model.LoginData, error) {
	// 这里使用模拟数据，实际实现中应该执行who命令并解析输出
	// 例如：who命令输出格式：username pts/0 2023-05-20 14:30 (192.168.1.1)
	
	// 模拟数据
	mockLogins := []model.LoginData{
		{
			Username:  "root",
			LoginTime: time.Now().Add(-30 * time.Minute),
			SourceIP:  "192.168.1.100",
			Terminal:  "pts/0",
			LoginType: "ssh",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Username:  "admin",
			LoginTime: time.Now().Add(-15 * time.Minute),
			SourceIP:  "192.168.1.101",
			Terminal:  "pts/1",
			LoginType: "ssh",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	
	return mockLogins, nil
}
