package collector

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/example/server-monitor/internal/model"
)

// systemStatsCollector 系统统计信息采集器实现
type systemStatsCollector struct {
	ticker  *time.Ticker
	done    chan bool
	stats   model.SystemStats
}

// NewSystemStatsCollector 创建系统统计信息采集器实例
func NewSystemStatsCollector(interval int) SystemStatsCollector {
	return &systemStatsCollector{
		ticker:  time.NewTicker(time.Duration(interval) * time.Second),
		done:    make(chan bool),
		stats:   model.SystemStats{},
	}
}

// Start 启动系统统计信息采集器
func (ssc *systemStatsCollector) Start() error {
	// 初始采集一次
	ssc.collect()
	
	// 启动定期采集
	go func() {
		for {
			select {
			case <-ssc.ticker.C:
				ssc.collect()
			case <-ssc.done:
				return
			}
		}
	}()
	
	return nil
}

// Stop 停止系统统计信息采集器
func (ssc *systemStatsCollector) Stop() error {
	ssc.done <- true
	ssc.ticker.Stop()
	return nil
}

// CollectSystemStats 采集系统统计信息
func (ssc *systemStatsCollector) CollectSystemStats() (model.SystemStats, error) {
	return ssc.stats, nil
}

// collect 采集系统统计信息
func (ssc *systemStatsCollector) collect() {
	stats := model.SystemStats{
		Timestamp: time.Now(),
	}
	
	// 获取进程数量
	procCount, err := ssc.getProcessCount()
	if err == nil {
		stats.TotalProcesses = procCount
	}
	
	// 获取运行中的进程数量
	runningCount, err := ssc.getRunningProcessCount()
	if err == nil {
		stats.RunningProcesses = runningCount
	}
	
	// 获取CPU使用率（简化实现）
	cpuUsage, err := ssc.getCPUUsage()
	if err == nil {
		stats.CPUUsage = cpuUsage
	}
	
	// 获取内存使用率
	memUsage, err := ssc.getMemoryUsage()
	if err == nil {
		stats.MemoryUsage = memUsage
	}
	
	// 获取磁盘使用率（简化实现）
	diskUsage, err := ssc.getDiskUsage()
	if err == nil {
		stats.DiskUsage = diskUsage
	}
	
	// 设置创建和更新时间
	now := time.Now()
	stats.CreatedAt = now
	stats.UpdatedAt = now
	
	ssc.stats = stats
}

// getProcessCount 获取进程总数
func (ssc *systemStatsCollector) getProcessCount() (int, error) {
	procDir := "/proc"
	entries, err := ioutil.ReadDir(procDir)
	if err != nil {
		return 0, fmt.Errorf("读取/proc目录失败: %v", err)
	}
	
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			// 检查是否为数字目录（PID）
			if _, err := strconv.Atoi(entry.Name()); err == nil {
				count++
			}
		}
	}
	
	return count, nil
}

// getRunningProcessCount 获取运行中的进程数量
func (ssc *systemStatsCollector) getRunningProcessCount() (int, error) {
	procDir := "/proc"
	entries, err := ioutil.ReadDir(procDir)
	if err != nil {
		return 0, fmt.Errorf("读取/proc目录失败: %v", err)
	}
	
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			// 检查是否为数字目录（PID）
			pid, err := strconv.Atoi(entry.Name())
			if err != nil {
				continue
			}
			
			// 读取进程状态
			statPath := fmt.Sprintf("/proc/%d/stat", pid)
			statContent, err := ioutil.ReadFile(statPath)
			if err != nil {
				continue
			}
			
			statFields := strings.Fields(string(statContent))
			if len(statFields) >= 3 && statFields[2] == "R" {
				count++
			}
		}
	}
	
	return count, nil
}

// getCPUUsage 获取CPU使用率（简化实现）
func (ssc *systemStatsCollector) getCPUUsage() (float64, error) {
	// 简化实现，实际需要读取/proc/stat文件并计算CPU使用率
	// 这里返回一个模拟值
	return 0.0, nil
}

// getMemoryUsage 获取内存使用率
func (ssc *systemStatsCollector) getMemoryUsage() (float64, error) {
	memInfo, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return 0.0, fmt.Errorf("读取/proc/meminfo文件失败: %v", err)
	}
	
	var totalMem, freeMem, buffers, cached int64
	lines := strings.Split(string(memInfo), "\n")
	
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				totalMem, _ = strconv.ParseInt(fields[1], 10, 64)
			}
		} else if strings.HasPrefix(line, "MemFree:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				freeMem, _ = strconv.ParseInt(fields[1], 10, 64)
			}
		} else if strings.HasPrefix(line, "Buffers:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				buffers, _ = strconv.ParseInt(fields[1], 10, 64)
			}
		} else if strings.HasPrefix(line, "Cached:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				cached, _ = strconv.ParseInt(fields[1], 10, 64)
			}
		}
	}
	
	// 计算已使用内存（考虑buffers和cached）
	usedMem := totalMem - freeMem - buffers - cached
	memUsage := (float64(usedMem) / float64(totalMem)) * 100
	
	return memUsage, nil
}

// getDiskUsage 获取磁盘使用率（简化实现）
func (ssc *systemStatsCollector) getDiskUsage() (float64, error) {
	// 简化实现，实际需要读取/proc/diskstats或使用df命令
	// 这里返回一个模拟值
	return 0.0, nil
}
