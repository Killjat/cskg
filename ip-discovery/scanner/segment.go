package scanner

import (
	"fmt"
	"log"
	"sync"
	"time"

	"ip-discovery/storage"
)

// SegmentScanner IP段扫描器
type SegmentScanner struct {
	pingScanner   *PingScanner
	influxClient  *storage.InfluxDBClient
	workers       int
	ipsPerSegment int
	scanInterval  time.Duration
}

// NewSegmentScanner 创建新的IP段扫描器
func NewSegmentScanner(influxClient *storage.InfluxDBClient, workers int, timeout time.Duration, ipsPerSegment int, scanInterval time.Duration) *SegmentScanner {
	return &SegmentScanner{
		pingScanner:   NewPingScanner(timeout, workers),
		influxClient:  influxClient,
		workers:       workers,
		ipsPerSegment: ipsPerSegment,
		scanInterval:  scanInterval,
	}
}

// ScanSegments 扫描IP段列表
func (s *SegmentScanner) ScanSegments(segments []*storage.IPSegment) error {
	fmt.Printf("开始扫描 %d 个IP段...\n", len(segments))

	// 创建任务通道
	jobs := make(chan *storage.IPSegment, len(segments))
	results := make(chan []*storage.AliveResult, len(segments))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for segment := range jobs {
				fmt.Printf("Worker %d 扫描段: %s\n", workerID, segment.CIDR)
				
				// 扫描IP段
				segmentResults, err := s.scanSegment(segment)
				if err != nil {
					log.Printf("扫描段 %s 失败: %v", segment.CIDR, err)
					continue
				}

				results <- segmentResults

				// 扫描间隔
				if s.scanInterval > 0 {
					time.Sleep(s.scanInterval)
				}
			}
		}(i)
	}

	// 启动结果处理协程
	go func() {
		wg.Wait()
		close(results)
	}()

	// 发送任务
	for _, segment := range segments {
		jobs <- segment
	}
	close(jobs)

	// 处理结果
	totalScanned := 0
	totalAlive := 0

	for segmentResults := range results {
		// 统计
		totalScanned += len(segmentResults)
		for _, result := range segmentResults {
			if result.IsAlive {
				totalAlive++
			}
		}

		// 批量写入InfluxDB
		if err := s.influxClient.WriteAliveResults(segmentResults); err != nil {
			log.Printf("写入InfluxDB失败: %v", err)
		} else {
			aliveCount := 0
			for _, r := range segmentResults {
				if r.IsAlive {
					aliveCount++
				}
			}
			fmt.Printf("已写入 %d 个结果到InfluxDB，其中 %d 个存活\n", len(segmentResults), aliveCount)
		}
	}

	fmt.Printf("扫描完成！总共扫描 %d 个IP，发现 %d 个存活IP\n", totalScanned, totalAlive)
	return nil
}

// scanSegment 扫描单个IP段
func (s *SegmentScanner) scanSegment(segment *storage.IPSegment) ([]*storage.AliveResult, error) {
	// 扫描IP段
	results, err := s.pingScanner.ScanCIDR(segment.CIDR, s.ipsPerSegment)
	if err != nil {
		return nil, fmt.Errorf("扫描CIDR %s 失败: %v", segment.CIDR, err)
	}

	return results, nil
}

// ScanSingleSegment 扫描单个IP段（用于测试）
func (s *SegmentScanner) ScanSingleSegment(cidr string) ([]*storage.AliveResult, error) {
	fmt.Printf("扫描IP段: %s\n", cidr)

	results, err := s.pingScanner.ScanCIDR(cidr, s.ipsPerSegment)
	if err != nil {
		return nil, err
	}

	// 统计结果
	aliveCount := 0
	for _, result := range results {
		if result.IsAlive {
			aliveCount++
			fmt.Printf("发现存活IP: %s (响应时间: %v)\n", result.IP, result.ResponseTime)
		}
	}

	fmt.Printf("扫描完成，共扫描 %d 个IP，发现 %d 个存活\n", len(results), aliveCount)

	// 写入InfluxDB
	if err := s.influxClient.WriteAliveResults(results); err != nil {
		return results, fmt.Errorf("写入InfluxDB失败: %v", err)
	}

	fmt.Printf("结果已写入InfluxDB\n")
	return results, nil
}

// GetScanStats 获取扫描统计信息
func (s *SegmentScanner) GetScanStats() (*storage.ScanStats, error) {
	// 获取IP段总数
	segmentCount, err := s.influxClient.GetSegmentCount()
	if err != nil {
		return nil, fmt.Errorf("获取IP段数量失败: %v", err)
	}

	// 获取存活IP数量
	aliveCount, err := s.influxClient.GetAliveIPCount()
	if err != nil {
		return nil, fmt.Errorf("获取存活IP数量失败: %v", err)
	}

	stats := &storage.ScanStats{
		TotalSegments: segmentCount,
		AliveIPs:      aliveCount,
	}

	return stats, nil
}