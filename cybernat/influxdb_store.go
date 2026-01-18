package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// InfluxDBConfig InfluxDB配置
type InfluxDBConfig struct {
	URL           string
	Token         string
	Organization  string
	Bucket        string
	BatchSize     int
	FlushInterval time.Duration
}

// InfluxDBStore InfluxDB存储管理器
type InfluxDBStore struct {
	client      influxdb2.Client
	config      InfluxDBConfig
	measurement string
}

// NewInfluxDBStore 创建新的InfluxDB存储管理器
func NewInfluxDBStore(config InfluxDBConfig, measurement string) (*InfluxDBStore, error) {
	// 创建InfluxDB客户端
	client := influxdb2.NewClient(config.URL, config.Token)

	return &InfluxDBStore{
		client:      client,
		config:      config,
		measurement: measurement,
	}, nil
}

// Close 关闭InfluxDB存储管理器
func (s *InfluxDBStore) Close() {
	// 关闭客户端
	s.client.Close()
	log.Println("InfluxDB连接已关闭")
}

// WriteScanResult 写入扫描结果到InfluxDB
func (s *InfluxDBStore) WriteScanResult(result *ScanResult, location *IPLocation) error {
	// 构建InfluxDB数据点
	point := influxdb2.NewPointWithMeasurement(s.measurement).
		AddTag("ip", result.IP).
		AddTag("country", location.Country).
		AddTag("region", location.Region).
		AddTag("city", location.City).
		AddTag("isp", location.ISP).
		AddTag("asn", strconv.FormatUint(uint64(location.ASN), 10)).
		AddTag("ip_segment", result.IPSegment).
		AddField("is_alive", result.IsAlive).
		AddField("scan_time", result.ScanTime).
		AddField("response_time", result.ResponseTime).
		AddField("open_ports_count", len(result.OpenPorts)).
		SetTime(result.ScanTimestamp)

	// 使用同步写入API
	writeAPI := s.client.WriteAPIBlocking(s.config.Organization, s.config.Bucket)
	return writeAPI.WritePoint(context.Background(), point)
}

// WriteScanResults 批量写入扫描结果到InfluxDB
func (s *InfluxDBStore) WriteScanResults(results []*ScanResult, locations map[string]*IPLocation) {
	for _, result := range results {
		// 只写入活跃IP的数据
		if result.IsAlive {
			location, exists := locations[result.IP]
			if exists {
				if err := s.WriteScanResult(result, location); err != nil {
					log.Printf("写入IP %s 到InfluxDB失败: %v", result.IP, err)
				} else {
					log.Printf("成功写入IP %s 到InfluxDB", result.IP)
				}
			} else {
				log.Printf("未找到IP %s 的位置信息，跳过写入", result.IP)
			}
		}
	}
}

// TestConnection 测试InfluxDB连接
func (s *InfluxDBStore) TestConnection() error {
	// 使用Ping测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.client.Ping(ctx)
	return err
}

// QueryActiveIPs 查询活跃IP数量
func (s *InfluxDBStore) QueryActiveIPs(timeRange time.Duration) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	queryAPI := s.client.QueryAPI(s.config.Organization)
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -%s)
		|> filter(fn: (r) => r._measurement == "%s")
		|> filter(fn: (r) => r._field == "is_alive")
		|> filter(fn: (r) => r._value == true)
		|> count()
	`, s.config.Bucket, timeRange, s.measurement)

	result, err := queryAPI.Query(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("查询InfluxDB失败: %v", err)
	}

	count := 0
	for result.Next() {
		if val, ok := result.Record().Value().(float64); ok {
			count = int(val)
		}
	}

	if result.Err() != nil {
		return 0, fmt.Errorf("处理查询结果失败: %v", result.Err())
	}

	return count, nil
}

// QueryIPScanResults 查询IP扫描结果
func (s *InfluxDBStore) QueryIPScanResults(timeRange time.Duration, limit int) ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	queryAPI := s.client.QueryAPI(s.config.Organization)
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -%s)
		|> filter(fn: (r) => r._measurement == "%s")
		|> filter(fn: (r) => r._field == "is_alive")
		|> sort(columns: ["_time"], desc: true)
		|> limit(n: %d)
	`, s.config.Bucket, timeRange, s.measurement, limit)

	result, err := queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询InfluxDB失败: %v", err)
	}

	results := make([]map[string]interface{}, 0)
	for result.Next() {
		record := result.Record()
		item := map[string]interface{}{
			"time":       record.Time(),
			"ip":         record.ValueByKey("ip"),
			"country":    record.ValueByKey("country"),
			"region":     record.ValueByKey("region"),
			"city":       record.ValueByKey("city"),
			"isp":        record.ValueByKey("isp"),
			"asn":        record.ValueByKey("asn"),
			"ip_segment": record.ValueByKey("ip_segment"),
			"is_alive":   record.Value(),
		}
		results = append(results, item)
	}

	if result.Err() != nil {
		return nil, fmt.Errorf("处理查询结果失败: %v", result.Err())
	}

	return results, nil
}

// QueryIPByCountry 查询不同国家的IP数量
func (s *InfluxDBStore) QueryIPByCountry(timeRange time.Duration) (map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	queryAPI := s.client.QueryAPI(s.config.Organization)
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -%s)
		|> filter(fn: (r) => r._measurement == "%s")
		|> filter(fn: (r) => r._field == "is_alive")
		|> filter(fn: (r) => r._value == true)
		|> group(columns: ["country"])
		|> count()
	`, s.config.Bucket, timeRange, s.measurement)

	result, err := queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询InfluxDB失败: %v", err)
	}

	countryCount := make(map[string]int)
	for result.Next() {
		if country, ok := result.Record().ValueByKey("country").(string); ok {
			if val, ok := result.Record().Value().(float64); ok {
				countryCount[country] = int(val)
			}
		}
	}

	if result.Err() != nil {
		return nil, fmt.Errorf("处理查询结果失败: %v", result.Err())
	}

	return countryCount, nil
}

// WriteIPSegmentStatus 写入IP段执行状态到InfluxDB
func (s *InfluxDBStore) WriteIPSegmentStatus(cidr string, status string, startTime time.Time, endTime time.Time) error {
	// 构建InfluxDB数据点
	point := influxdb2.NewPointWithMeasurement("ip_segment_status").
		AddTag("cidr", cidr).
		AddField("status", status).
		AddField("start_time", startTime.Unix()).
		AddField("end_time", endTime.Unix()).
		AddField("duration", endTime.Sub(startTime).Seconds()).
		SetTime(endTime)

	// 使用同步写入API
	writeAPI := s.client.WriteAPIBlocking(s.config.Organization, s.config.Bucket)
	return writeAPI.WritePoint(context.Background(), point)
}

// GetLastProcessedIPSegment 获取上一次处理的IP段
func (s *InfluxDBStore) GetLastProcessedIPSegment() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	queryAPI := s.client.QueryAPI(s.config.Organization)
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -7d)
		|> filter(fn: (r) => r._measurement == "ip_segment_status")
		|> filter(fn: (r) => r._field == "status")
		|> sort(columns: ["_time"], desc: true)
		|> limit(n: 1)
	`, s.config.Bucket)

	result, err := queryAPI.Query(ctx, query)
	if err != nil {
		return "", fmt.Errorf("查询InfluxDB失败: %v", err)
	}

	var lastCIDR string
	for result.Next() {
		if cidr, ok := result.Record().ValueByKey("cidr").(string); ok {
			lastCIDR = cidr
		}
	}

	if result.Err() != nil {
		return "", fmt.Errorf("处理查询结果失败: %v", result.Err())
	}

	return lastCIDR, nil
}

// IsIPSegmentProcessed 检查IP段是否已经处理过
func (s *InfluxDBStore) IsIPSegmentProcessed(cidr string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	queryAPI := s.client.QueryAPI(s.config.Organization)
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -7d)
		|> filter(fn: (r) => r._measurement == "ip_segment_status")
		|> filter(fn: (r) => r.cidr == "%s")
		|> filter(fn: (r) => r._field == "status")
		|> filter(fn: (r) => r._value == "completed")
		|> count()
	`, s.config.Bucket, cidr)

	result, err := queryAPI.Query(ctx, query)
	if err != nil {
		return false, fmt.Errorf("查询InfluxDB失败: %v", err)
	}

	count := 0
	for result.Next() {
		if val, ok := result.Record().Value().(float64); ok {
			count = int(val)
		}
	}

	if result.Err() != nil {
		return false, fmt.Errorf("处理查询结果失败: %v", result.Err())
	}

	return count > 0, nil
}

// Example 使用示例
func ExampleInfluxDBStore() {
	// 创建InfluxDB配置
	config := InfluxDBConfig{
		URL:           "http://localhost:8086",
		Token:         "your-token",
		Organization:  "your-org",
		Bucket:        "taiwan_ip_scan",
		BatchSize:     1000,
		FlushInterval: 1 * time.Second,
	}

	// 创建InfluxDB存储管理器
	store, err := NewInfluxDBStore(config, "taiwan_ip_scan")
	if err != nil {
		log.Fatalf("创建InfluxDB存储管理器失败: %v", err)
	}
	defer store.Close()

	// 测试连接（可选）
	// if err := store.TestConnection(); err != nil {
	// 	log.Printf("InfluxDB连接测试失败: %v", err)
	// 	// 继续执行，可能是测试环境没有InfluxDB
	// }

	// 创建示例扫描结果
	exampleResult := &ScanResult{
		IP:            "8.8.8.8",
		IsAlive:       true,
		ScanTime:      123.45,
		ResponseTime:  15.5,
		OpenPorts:     []int{80, 443},
		ScanTimestamp: time.Now(),
		IPSegment:     "8.8.8.0/24",
	}

	// 创建示例位置信息
	exampleLocation := &IPLocation{
		IP:        "8.8.8.8",
		Country:   "Taiwan",
		Region:    "Taipei",
		City:      "Taipei",
		ISP:       "Chunghwa Telecom",
		ASN:       3462,
		Latitude:  25.0330,
		Longitude: 121.5654,
	}

	// 写入示例数据
	if err := store.WriteScanResult(exampleResult, exampleLocation); err != nil {
		log.Printf("写入示例数据失败: %v", err)
	} else {
		log.Println("示例数据已写入InfluxDB")
	}
}
