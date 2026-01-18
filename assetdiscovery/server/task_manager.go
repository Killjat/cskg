package main

import (
	"log"
	"sync"
	"time"

	"github.com/cskg/assetdiscovery/common"
)

// TaskManager 任务管理器结构体
type TaskManager struct {
	server    *Server
	tasks     map[string]*common.Task
	tasksLock sync.RWMutex
}

// NewTaskManager 创建新的任务管理器
func NewTaskManager(server *Server) *TaskManager {
	return &TaskManager{
		server: server,
		tasks:  make(map[string]*common.Task),
	}
}

// AddTask 添加任务到管理器
func (tm *TaskManager) AddTask(task *common.Task) {
	tm.tasksLock.Lock()
	defer tm.tasksLock.Unlock()

	tm.tasks[task.TaskID] = task
	log.Printf("Added task %s: %s - %s", task.TaskID, task.TaskType, task.Target)
}

// GetTask 获取任务
func (tm *TaskManager) GetTask(taskID string) (*common.Task, bool) {
	tm.tasksLock.RLock()
	defer tm.tasksLock.RUnlock()

	task, exists := tm.tasks[taskID]
	return task, exists
}

// RemoveTask 移除任务
func (tm *TaskManager) RemoveTask(taskID string) {
	tm.tasksLock.Lock()
	defer tm.tasksLock.Unlock()

	if _, exists := tm.tasks[taskID]; exists {
		delete(tm.tasks, taskID)
		log.Printf("Removed task %s", taskID)
	}
}

// GetAllTasks 获取所有任务
func (tm *TaskManager) GetAllTasks() []*common.Task {
	tm.tasksLock.RLock()
	defer tm.tasksLock.RUnlock()

	tasks := make([]*common.Task, 0, len(tm.tasks))
	for _, task := range tm.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// GetTasksCount 获取任务数量
func (tm *TaskManager) GetTasksCount() int {
	tm.tasksLock.RLock()
	defer tm.tasksLock.RUnlock()

	return len(tm.tasks)
}

// CleanupExpiredTasks 清理过期任务
func (tm *TaskManager) CleanupExpiredTasks(expiryTime int64) {
	tm.tasksLock.Lock()
	defer tm.tasksLock.Unlock()

	currentTime := time.Now().Unix()
	for taskID, task := range tm.tasks {
		if currentTime-task.Timestamp > expiryTime {
			delete(tm.tasks, taskID)
			log.Printf("Cleaned up expired task %s", taskID)
		}
	}
}
