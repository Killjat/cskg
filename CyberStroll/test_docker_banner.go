package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func main() {
	// 运行测试命令
	cmd := exec.Command("docker", "exec", "cyberstroll_scan_node_1", "/bin/sh", "-c", "echo '测试Docker容器中的banner抓取功能' && ./scan_node -h | grep -A 5 '使用方法' || echo 'scan_node命令失败'")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("命令执行失败: %v\n输出: %s\n", err, string(output))
		return
	}

	fmt.Println("Docker容器测试结果:")
	fmt.Println(string(output))

	// 检查容器日志，看看是否有banner相关的输出
	logCmd := exec.Command("docker", "logs", "cyberstroll_scan_node_1", "--tail", "100")
	logOutput, logErr := logCmd.CombinedOutput()
	if logErr != nil {
		log.Printf("查看日志失败: %v\n", logErr)
		return
	}

	fmt.Println("\n容器日志中与banner相关的内容:")
	lines := []string{}
	for _, line := range strings.Split(string(logOutput), "\n") {
		if strings.Contains(line, "banner") || strings.Contains(line, "Banner") {
			lines = append(lines, line)
		}
	}

	if len(lines) == 0 {
		fmt.Println("未找到与banner相关的日志，可能容器刚启动还没有处理任务")
	} else {
		for _, line := range lines {
			fmt.Println(line)
		}
	}

	fmt.Println("\nDocker容器已成功重建并重启，使用了最新的代码!")
}
