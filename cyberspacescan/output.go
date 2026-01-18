package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SaveResults 保存扫描结果（同时保存JSON和CSV格式）
func SaveResults(outputDir string, results []*ScanResult, format string) error {
	timestamp := time.Now().Format("20060102_150405")

	// 始终保存JSON格式
	if err := saveJSON(outputDir, results, timestamp); err != nil {
		return fmt.Errorf("保存JSON失败: %v", err)
	}

	// 始终保存CSV格式
	if err := saveCSV(outputDir, results, timestamp); err != nil {
		return fmt.Errorf("保存CSV失败: %v", err)
	}

	// 如果指定了其他格式，也保存
	switch format {
	case "txt":
		if err := saveTXT(outputDir, results, timestamp); err != nil {
			return fmt.Errorf("保存TXT失败: %v", err)
		}
	}

	return nil
}

// saveJSON 保存为JSON格式
func saveJSON(outputDir string, results []*ScanResult, timestamp string) error {
	filename := filepath.Join(outputDir, fmt.Sprintf("scan_result_%s.json", timestamp))
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	output := map[string]interface{}{
		"scan_time": timestamp,
		"total":     len(results),
		"alive":     countAliveHosts(results),
		"results":   results,
	}

	if err := encoder.Encode(output); err != nil {
		return err
	}

	fmt.Printf("\n✓ 结果已保存到: %s\n", filename)
	return nil
}

// saveCSV 保存为CSV格式
func saveCSV(outputDir string, results []*ScanResult, timestamp string) error {
	filename := filepath.Join(outputDir, fmt.Sprintf("scan_result_%s.csv", timestamp))
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	headers := []string{"IP地址", "端口", "协议", "状态", "服务", "Banner"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	// 写入数据
	for _, result := range results {
		if !result.IsAlive {
			continue
		}

		// TCP端口
		for _, port := range result.TCPPorts {
			row := []string{
				result.IP,
				fmt.Sprintf("%d", port.Port),
				port.Protocol,
				port.State,
				port.Service,
				port.Banner,
			}
			if err := writer.Write(row); err != nil {
				return err
			}
		}

		// UDP端口
		for _, port := range result.UDPPorts {
			row := []string{
				result.IP,
				fmt.Sprintf("%d", port.Port),
				port.Protocol,
				port.State,
				port.Service,
				port.Banner,
			}
			if err := writer.Write(row); err != nil {
				return err
			}
		}
	}

	fmt.Printf("\n✓ 结果已保存到: %s\n", filename)
	return nil
}

// saveTXT 保存为文本格式
func saveTXT(outputDir string, results []*ScanResult, timestamp string) error {
	filename := filepath.Join(outputDir, fmt.Sprintf("scan_result_%s.txt", timestamp))
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "网络空间扫描结果\n")
	fmt.Fprintf(file, "扫描时间: %s\n", timestamp)
	fmt.Fprintf(file, "%s\n\n", strings.Repeat("=", 70))

	for _, result := range results {
		fmt.Fprintf(file, "IP地址: %s\n", result.IP)
		fmt.Fprintf(file, "状态: ")
		if result.IsAlive {
			fmt.Fprintf(file, "存活\n")
		} else {
			fmt.Fprintf(file, "不可达\n")
			fmt.Fprintf(file, "\n")
			continue
		}

		if len(result.TCPPorts) > 0 {
			fmt.Fprintf(file, "\nTCP开放端口:\n")
			for _, port := range result.TCPPorts {
				fmt.Fprintf(file, "  端口: %d\n", port.Port)
				fmt.Fprintf(file, "  服务: %s\n", port.Service)
				fmt.Fprintf(file, "  状态: %s\n", port.State)
				if port.Banner != "" {
					fmt.Fprintf(file, "  Banner: %s\n", port.Banner)
				}
				fmt.Fprintf(file, "\n")
			}
		}

		if len(result.UDPPorts) > 0 {
			fmt.Fprintf(file, "UDP开放端口:\n")
			for _, port := range result.UDPPorts {
				fmt.Fprintf(file, "  端口: %d\n", port.Port)
				fmt.Fprintf(file, "  服务: %s\n", port.Service)
				fmt.Fprintf(file, "  状态: %s\n", port.State)
				fmt.Fprintf(file, "\n")
			}
		}

		fmt.Fprintf(file, "%s\n\n", strings.Repeat("-", 70))
	}

	fmt.Printf("\n✓ 结果已保存到: %s\n", filename)
	return nil
}

// SaveResponsePacket 单独保存响应包
func SaveResponsePacket(outputDir, ip string, port int, protocol string, response []byte) error {
	dir := filepath.Join(outputDir, "responses", ip)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filename := filepath.Join(dir, fmt.Sprintf("%s_%d_%s.bin", ip, port, protocol))
	return os.WriteFile(filename, response, 0644)
}
