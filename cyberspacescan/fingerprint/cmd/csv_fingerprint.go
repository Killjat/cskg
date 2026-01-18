package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	
	"cskg/cyberspacescan/fingerprint"
)

func main() {
	// 命令行参数
	inputFile := flag.String("i", "", "输入CSV文件路径")
	outputFile := flag.String("o", "", "输出CSV文件路径（可选，默认在同目录下生成）")
	batchMode := flag.Bool("batch", false, "批量处理模式")
	inputDir := flag.String("dir", "", "批量处理：输入目录")
	outputDir := flag.String("outdir", "", "批量处理：输出目录")
	showStats := flag.Bool("stats", false, "显示详细统计信息")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "🔍 CSV指纹识别工具\n\n")
		fmt.Fprintf(os.Stderr, "用法:\n")
		fmt.Fprintf(os.Stderr, "  单文件处理:\n")
		fmt.Fprintf(os.Stderr, "    %s -i <输入文件> [-o <输出文件>] [-stats]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  批量处理:\n")
		fmt.Fprintf(os.Stderr, "    %s -batch -dir <输入目录> -outdir <输出目录>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "参数说明:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  # 处理单个文件\n")
		fmt.Fprintf(os.Stderr, "  %s -i scan_result.csv -o scan_result_fingerprint.csv\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 显示详细统计\n")
		fmt.Fprintf(os.Stderr, "  %s -i scan_result.csv -stats\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # 批量处理\n")
		fmt.Fprintf(os.Stderr, "  %s -batch -dir ./results -outdir ./fingerprint_results\n\n", os.Args[0])
	}
	
	flag.Parse()
	
	fmt.Println("🔍 CSV指纹识别工具")
	fmt.Println("=" + "================================")
	fmt.Println()
	
	// 批量处理模式
	if *batchMode {
		if *inputDir == "" {
			fmt.Println("❌ 错误: 批量模式需要指定输入目录 (-dir)")
			flag.Usage()
			os.Exit(1)
		}
		
		if *outputDir == "" {
			*outputDir = filepath.Join(*inputDir, "fingerprint_output")
		}
		
		fmt.Printf("📁 输入目录: %s\n", *inputDir)
		fmt.Printf("📁 输出目录: %s\n\n", *outputDir)
		
		if err := fingerprint.BatchProcessCSV(*inputDir, *outputDir); err != nil {
			fmt.Printf("❌ 批量处理失败: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println("✅ 批量处理完成！")
		return
	}
	
	// 单文件处理模式
	if *inputFile == "" {
		fmt.Println("❌ 错误: 请指定输入文件 (-i)")
		flag.Usage()
		os.Exit(1)
	}
	
	// 检查输入文件
	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		fmt.Printf("❌ 错误: 输入文件不存在: %s\n", *inputFile)
		os.Exit(1)
	}
	
	// 生成输出文件名
	if *outputFile == "" {
		dir := filepath.Dir(*inputFile)
		base := filepath.Base(*inputFile)
		ext := filepath.Ext(base)
		name := base[:len(base)-len(ext)]
		*outputFile = filepath.Join(dir, name+"_fingerprint"+ext)
	}
	
	fmt.Printf("📖 输入文件: %s\n", *inputFile)
	fmt.Printf("💾 输出文件: %s\n\n", *outputFile)
	
	// 处理CSV
	if *showStats {
		// 使用带统计信息的处理函数
		stats, err := fingerprint.ProcessCSVWithStats(*inputFile, *outputFile)
		if err != nil {
			fmt.Printf("❌ 处理失败: %v\n", err)
			os.Exit(1)
		}
		
		stats.PrintStats()
	} else {
		// 简单处理
		if err := fingerprint.ProcessCSV(*inputFile, *outputFile); err != nil {
			fmt.Printf("❌ 处理失败: %v\n", err)
			os.Exit(1)
		}
	}
	
	fmt.Println("\n✅ 处理完成！")
}
