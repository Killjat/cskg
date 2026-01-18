package collector

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/example/server-monitor/internal/model"
	"github.com/fsnotify/fsnotify"
)

// fileCollector 文件操作采集器实现
type fileCollector struct {
	watcher      *fsnotify.Watcher
	done         chan bool
	watchPaths   map[string]bool
	fileOps      []model.FileOperationData
	recursive    bool
}

// NewFileCollector 创建文件操作采集器实例
func NewFileCollector(recursive bool) (FileCollector, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("创建文件监控器失败: %v", err)
	}
	
	return &fileCollector{
		watcher:      watcher,
		done:         make(chan bool),
		watchPaths:   make(map[string]bool),
		fileOps:      []model.FileOperationData{},
		recursive:    recursive,
	}, nil
}

// Start 启动文件操作采集器
func (fc *fileCollector) Start() error {
	// 启动事件处理协程
	go fc.handleEvents()
	return nil
}

// Stop 停止文件操作采集器
func (fc *fileCollector) Stop() error {
	fc.done <- true
	return fc.watcher.Close()
}

// CollectFileOperations 采集文件操作记录
func (fc *fileCollector) CollectFileOperations() ([]model.FileOperationData, error) {
	return fc.fileOps, nil
}

// AddWatch 添加监控路径
func (fc *fileCollector) AddWatch(path string, recursive bool) error {
	// 检查路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("监控路径不存在: %s", path)
	}
	
	// 如果是目录且需要递归监控，添加所有子目录
	if info, err := os.Stat(path); err == nil && info.IsDir() && recursive {
		err = filepath.Walk(path, func(subPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if err := fc.watcher.Add(subPath); err != nil {
					return fmt.Errorf("添加监控目录失败: %s, 错误: %v", subPath, err)
				}
				fc.watchPaths[subPath] = true
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		// 添加单个路径监控
		if err := fc.watcher.Add(path); err != nil {
			return fmt.Errorf("添加监控路径失败: %s, 错误: %v", path, err)
		}
		fc.watchPaths[path] = true
	}
	
	return nil
}

// RemoveWatch 移除监控路径
func (fc *fileCollector) RemoveWatch(path string) error {
	// 如果是目录且需要递归监控，移除所有子目录
	if info, err := os.Stat(path); err == nil && info.IsDir() && fc.recursive {
		err = filepath.Walk(path, func(subPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if err := fc.watcher.Remove(subPath); err != nil {
					// 忽略不存在的路径错误
					if !os.IsNotExist(err) {
						return fmt.Errorf("移除监控目录失败: %s, 错误: %v", subPath, err)
					}
				}
				delete(fc.watchPaths, subPath)
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		// 移除单个路径监控
		if err := fc.watcher.Remove(path); err != nil {
			// 忽略不存在的路径错误
			if !os.IsNotExist(err) {
				return fmt.Errorf("移除监控路径失败: %s, 错误: %v", path, err)
			}
		}
		delete(fc.watchPaths, path)
	}
	
	return nil
}

// handleEvents 处理文件系统事件
func (fc *fileCollector) handleEvents() {
	for {
		select {
		case event, ok := <-fc.watcher.Events:
			if !ok {
				return
			}
			
			// 转换事件类型
			opType := ""
			switch {
			case event.Op&fsnotify.Create != 0:
				opType = "create"
			case event.Op&fsnotify.Write != 0:
				opType = "write"
			case event.Op&fsnotify.Remove != 0:
				opType = "delete"
			case event.Op&fsnotify.Rename != 0:
				opType = "move"
			case event.Op&fsnotify.Chmod != 0:
				opType = "chmod"
			}
			
			if opType != "" {
				// 创建文件操作记录
				fileOp := model.FileOperationData{
					Operation:     opType,
					FilePath:      event.Name,
					OperationTime: time.Now(),
					Result:        "success",
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
				
				// 尝试获取操作的进程信息（简化实现）
				fileOp.PID = os.Getpid() // 实际实现中需要通过inode查找对应的进程
				fileOp.ProcessName = "server-monitor"
				fileOp.Username = "root" // 实际实现中需要获取实际用户名
				
				// 添加到操作记录
				fc.fileOps = append(fc.fileOps, fileOp)
				
				// 限制记录数量，避免内存占用过高
				if len(fc.fileOps) > 1000 {
					fc.fileOps = fc.fileOps[len(fc.fileOps)-1000:]
				}
				
				// 如果是创建目录且需要递归监控，自动添加监控
				if opType == "create" {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() && fc.recursive {
						if err := fc.watcher.Add(event.Name); err == nil {
							fc.watchPaths[event.Name] = true
						}
					}
				}
			}
			
		case err, ok := <-fc.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("文件监控错误: %v\n", err)
			
		case <-fc.done:
			return
		}
	}
}

// getUsernameFromPID 根据PID获取用户名
func (fc *fileCollector) getUsernameFromPID(pid int) string {
	// 简化实现，实际需要读取/proc/[PID]/status文件
	return "root"
}

// getProcessNameFromPID 根据PID获取进程名称
func (fc *fileCollector) getProcessNameFromPID(pid int) string {
	// 简化实现，实际需要读取/proc/[PID]/comm文件
	return "unknown"
}
