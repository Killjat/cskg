package collector

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/example/server-monitor/internal/model"
)

// processCollector 进程信息采集器实现
type processCollector struct {
	ticker  *time.Ticker
	done    chan bool
	processes []model.ProcessData
}

// NewProcessCollector 创建进程采集器实例
func NewProcessCollector(interval int) ProcessCollector {
	return &processCollector{
		ticker:  time.NewTicker(time.Duration(interval) * time.Second),
		done:    make(chan bool),
		processes: []model.ProcessData{},
	}
}

// Start 启动进程采集器
func (pc *processCollector) Start() error {
	// 初始采集一次
	pc.collect()
	
	// 启动定期采集
	go func() {
		for {
			select {
			case <-pc.ticker.C:
				pc.collect()
			case <-pc.done:
				return
			}
		}
	}()
	
	return nil
}

// Stop 停止进程采集器
func (pc *processCollector) Stop() error {
	pc.done <- true
	pc.ticker.Stop()
	return nil
}

// CollectProcesses 采集所有进程信息
func (pc *processCollector) CollectProcesses() ([]model.ProcessData, error) {
	return pc.processes, nil
}

// GetProcessByPID 根据PID获取进程信息
func (pc *processCollector) GetProcessByPID(pid int) (model.ProcessData, error) {
	for _, p := range pc.processes {
		if p.PID == pid {
			return p, nil
		}
	}
	return model.ProcessData{}, fmt.Errorf("进程不存在: PID=%d", pid)
}

// GetProcessCount 获取进程总数
func (pc *processCollector) GetProcessCount() (int, error) {
	return len(pc.processes), nil
}

// GetRunningProcessCount 获取运行中的进程数
func (pc *processCollector) GetRunningProcessCount() (int, error) {
	count := 0
	for _, p := range pc.processes {
		if p.Status == "running" {
			count++
		}
	}
	return count, nil
}

// collect 采集进程信息的内部方法
func (pc *processCollector) collect() {
	processes := []model.ProcessData{}
	
	// 遍历/proc目录下的数字目录，每个目录代表一个进程
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
		
		// 检查目录名是否为数字（PID）
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		
		// 读取进程信息
		process, err := pc.readProcessInfo(pid)
		if err != nil {
			// 跳过无法读取的进程
			continue
		}
		
		processes = append(processes, process)
	}
	
	pc.processes = processes
}

// readProcessInfo 读取单个进程的详细信息
func (pc *processCollector) readProcessInfo(pid int) (model.ProcessData, error) {
	process := model.ProcessData{
		PID: pid,
	}
	
	// 读取/proc/[PID]/stat文件获取基本状态信息
	statPath := filepath.Join("/proc", strconv.Itoa(pid), "stat")
	statContent, err := ioutil.ReadFile(statPath)
	if err != nil {
		return process, fmt.Errorf("读取进程stat文件失败: %v", err)
	}
	
	// 解析stat文件内容
	statFields := strings.Fields(string(statContent))
	if len(statFields) < 22 {
		return process, fmt.Errorf("stat文件格式不正确")
	}
	
	// 进程名称（去除括号）
	process.Name = strings.Trim(statFields[1], "()")
	
	// 进程状态
	status := statFields[2]
	switch status {
	case "R":
		process.Status = "running"
	case "S":
		process.Status = "sleeping"
	case "D":
		process.Status = "disk_sleep"
	case "Z":
		process.Status = "zombie"
	case "T":
		process.Status = "stopped"
	case "t":
		process.Status = "tracing_stop"
	case "X":
		process.Status = "dead"
	case "I":
		process.Status = "idle"
	default:
		process.Status = status
	}
	
	// 父进程ID
	ppid, _ := strconv.Atoi(statFields[3])
	process.PPID = ppid
	
	// 读取/proc/[PID]/cmdline文件获取命令行
	cmdlinePath := filepath.Join("/proc", strconv.Itoa(pid), "cmdline")
	cmdlineContent, err := ioutil.ReadFile(cmdlinePath)
	if err == nil {
		// cmdline文件中的参数以null字符分隔
		cmdline := strings.ReplaceAll(string(cmdlineContent), "\x00", " ")
		process.Command = strings.TrimSpace(cmdline)
	}
	
	// 读取/proc/[PID]/status文件获取更多信息
	statusPath := filepath.Join("/proc", strconv.Itoa(pid), "status")
	statusContent, err := ioutil.ReadFile(statusPath)
	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(statusContent)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "Uid:") {
				// 用户ID
				uidFields := strings.Fields(line)
				if len(uidFields) > 1 {
					uid, _ := strconv.Atoi(uidFields[1])
					// 根据UID获取用户名
					process.Username = pc.getUsernameByUID(uid)
				}
			}
		}
	}
	
	// 计算进程启动时间
	timestamp, _ := strconv.ParseInt(statFields[21], 10, 64)
	// 读取系统启动时间
	bootTime, err := pc.getBootTime()
	if err == nil {
		startTime := bootTime + (timestamp / 100) // 转换为秒
		process.StartTime = time.Unix(startTime, 0)
	}
	
	// 计算CPU使用率（简化实现，实际需要更复杂的计算）
	process.CPUUsage = 0.0
	
	// 计算内存使用率
	rss, _ := strconv.ParseInt(statFields[23], 10, 64)
	// 读取系统总内存
	totalMem, err := pc.getTotalMemory()
	if err == nil {
		process.MemoryUsage = (float64(rss*4096) / float64(totalMem)) * 100 // rss单位是页，每页4096字节
	}
	
	// 设置创建和更新时间
	now := time.Now()
	process.CreatedAt = now
	process.UpdatedAt = now
	
	return process, nil
}

// getUsernameByUID 根据UID获取用户名
func (pc *processCollector) getUsernameByUID(uid int) string {
	// 读取/etc/passwd文件查找用户名
	passwdFile, err := os.Open("/etc/passwd")
	if err != nil {
		return strconv.Itoa(uid)
	}
	defer passwdFile.Close()
	
	scanner := bufio.NewScanner(passwdFile)
	for scanner.Scan() {
		line := scanner.Text()
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

// getBootTime 获取系统启动时间
func (pc *processCollector) getBootTime() (int64, error) {
	bootTimeStr, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return 0, err
	}
	
	lines := strings.Split(string(bootTimeStr), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "btime ") {
			bootTime, err := strconv.ParseInt(strings.Fields(line)[1], 10, 64)
			if err != nil {
				return 0, err
			}
			return bootTime, nil
		}
	}
	
	return 0, fmt.Errorf("未找到系统启动时间")
}

// getTotalMemory 获取系统总内存（字节）
func (pc *processCollector) getTotalMemory() (int64, error) {
	memInfo, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	
	lines := strings.Split(string(memInfo), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			memFields := strings.Fields(line)
			if len(memFields) < 2 {
				return 0, fmt.Errorf("MemTotal格式不正确")
			}
			memKB, err := strconv.ParseInt(memFields[1], 10, 64)
			if err != nil {
				return 0, err
			}
			return memKB * 1024, nil // 转换为字节
		}
	}
	
	return 0, fmt.Errorf("未找到系统总内存信息")
}

