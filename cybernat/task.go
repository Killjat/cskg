package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// Executor 任务执行器
type Executor struct {
	ResultCallback func(*TaskResult)
}

// NewExecutor 创建任务执行器
func NewExecutor(resultCallback func(*TaskResult)) *Executor {
	return &Executor{
		ResultCallback: resultCallback,
	}
}

// ExecuteTask 执行任务
func (e *Executor) ExecuteTask(task *Task) {
	// 创建结果对象
	result := &TaskResult{
		TaskID:    task.ID,
		NodeID:    task.NodeID,
		Output:    "",
		Error:     "",
		ExitCode:  0,
		Completed: false,
		Timestamp: time.Now(),
	}

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(task.Timeout)*time.Second)
	defer cancel()

	// 构建命令
	cmd := exec.CommandContext(ctx, task.Command, task.Args...)

	// 执行命令
	output, err := cmd.CombinedOutput()
	result.Output = string(output)

	// 检查错误
	if err != nil {
		// 检查是否是超时错误
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = "任务执行超时"
			result.ExitCode = -1
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			// 命令执行失败，但退出码有效
			result.Error = fmt.Sprintf("命令执行失败: %v", exitErr)
			result.ExitCode = exitErr.ExitCode()
		} else {
			// 其他错误
			result.Error = fmt.Sprintf("无法执行命令: %v", err)
			result.ExitCode = -1
		}
	} else {
		// 命令执行成功
		result.ExitCode = 0
	}

	// 标记任务完成
	result.Completed = true
	result.Timestamp = time.Now()

	// 调用回调函数
	if e.ResultCallback != nil {
		e.ResultCallback(result)
	}
}
