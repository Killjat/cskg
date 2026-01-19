package storage

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// InfluxDBClient InfluxDB客户端
type InfluxDBClient struct {
	client         influxdb2.Client
	organization   string
	segmentsBucket string
	aliveBucket    string
}

// NewInfluxDBClient 创建新的InfluxDB客户端
func NewInfluxDBClient(url, token, org, segmentsBucket, aliveBucket string) *InfluxDBClient {
	client := influxdb2.NewClient(url, token)

	return &InfluxDBClient{
		client:         client,
		organization:   org,
		segmentsBucket: segmentsBucket,
		aliveBucket:    aliveBucket,
	}
}

// Close 关闭客户端
func (c *InfluxDBClient) Close() {
	c.client.Close()
}

// TestConnection 测试连接
func (c *InfluxDBClient) TestConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.Ping(ctx)
	return err
}

// WriteIPSegment 写入IP段信息
func (c *InfluxDBClient) WriteIPSegment(segment *IPSegment) error {
	writeAPI := c.client.WriteAPIBlocking(c.organization, c.segmentsBucket)

	point := influxdb2.NewPointWithMeasurement("ip_segments").
		AddTag("cidr", segment.CIDR).
		AddTag("country", segment.Country).
		AddTag("type", segment.Type).
		AddTag("status", segment.Status).
		AddTag("registry", segment.Registry).
		AddField("start_ip", segment.StartIP).
		AddField("end_ip", segment.EndIP).
		AddField("ip_count", segment.IPCount).
		AddField("date", segment.Date).
		SetTime(segment.CreatedAt)

	return writeAPI.WritePoint(context.Background(), point)
}

// WriteIPSegments 批量写入IP段信息
func (c *InfluxDBClient) WriteIPSegments(segments []*IPSegment) error {
	writeAPI := c.client.WriteAPIBlocking(c.organization, c.segmentsBucket)

	for _, segment := range segments {
		point := influxdb2.NewPointWithMeasurement("ip_segments").
			AddTag("cidr", segment.CIDR).
			AddTag("country", segment.Country).
			AddTag("type", segment.Type).
			AddTag("status", segment.Status).
			AddTag("registry", segment.Registry).
			AddField("start_ip", segment.StartIP).
			AddField("end_ip", segment.EndIP).
			AddField("ip_count", segment.IPCount).
			AddField("date", segment.Date).
			SetTime(segment.CreatedAt)

		if err := writeAPI.WritePoint(context.Background(), point); err != nil {
			return err
		}
	}

	return nil
}

// WriteAliveResult 写入探活结果
func (c *InfluxDBClient) WriteAliveResult(result *AliveResult) error {
	writeAPI := c.client.WriteAPIBlocking(c.organization, c.aliveBucket)

	point := influxdb2.NewPointWithMeasurement("ip_alive").
		AddTag("ip", result.IP).
		AddTag("cidr", result.CIDR).
		AddField("is_alive", result.IsAlive).
		AddField("response_time_ms", float64(result.ResponseTime.Nanoseconds())/1e6).
		AddField("packet_loss", result.PacketLoss).
		AddField("error_message", result.ErrorMessage).
		SetTime(result.ScanTime)

	return writeAPI.WritePoint(context.Background(), point)
}

// WriteAliveResults 批量写入探活结果
func (c *InfluxDBClient) WriteAliveResults(results []*AliveResult) error {
	writeAPI := c.client.WriteAPIBlocking(c.organization, c.aliveBucket)

	for _, result := range results {
		point := influxdb2.NewPointWithMeasurement("ip_alive").
			AddTag("ip", result.IP).
			AddTag("cidr", result.CIDR).
			AddField("is_alive", result.IsAlive).
			AddField("response_time_ms", float64(result.ResponseTime.Nanoseconds())/1e6).
			AddField("packet_loss", result.PacketLoss).
			AddField("error_message", result.ErrorMessage).
			SetTime(result.ScanTime)

		if err := writeAPI.WritePoint(context.Background(), point); err != nil {
			return err
		}
	}

	return nil
}

// GetSegmentCount 获取IP段总数
func (c *InfluxDBClient) GetSegmentCount() (int, error) {
	queryAPI := c.client.QueryAPI(c.organization)

	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -30d)
		|> filter(fn: (r) => r._measurement == "ip_segments")
		|> filter(fn: (r) => r._field == "ip_count")
		|> count()
	`, c.segmentsBucket)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return 0, err
	}

	count := 0
	for result.Next() {
		if val, ok := result.Record().Value().(int64); ok {
			count = int(val)
		}
	}

	return count, result.Err()
}

// GetAliveIPCount 获取存活IP数量
func (c *InfluxDBClient) GetAliveIPCount() (int, error) {
	queryAPI := c.client.QueryAPI(c.organization)

	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -1d)
		|> filter(fn: (r) => r._measurement == "ip_alive")
		|> filter(fn: (r) => r._field == "is_alive")
		|> filter(fn: (r) => r._value == true)
		|> count()
	`, c.aliveBucket)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return 0, err
	}

	count := 0
	for result.Next() {
		if val, ok := result.Record().Value().(int64); ok {
			count = int(val)
		}
	}

	return count, result.Err()
}

// GetRecentAliveIPs 获取最近的存活IP列表
func (c *InfluxDBClient) GetRecentAliveIPs(limit int) ([]*AliveResult, error) {
	queryAPI := c.client.QueryAPI(c.organization)

	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -1d)
		|> filter(fn: (r) => r._measurement == "ip_alive")
		|> filter(fn: (r) => r._field == "is_alive")
		|> filter(fn: (r) => r._value == true)
		|> sort(columns: ["_time"], desc: true)
		|> limit(n: %d)
	`, c.aliveBucket, limit)

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	var results []*AliveResult
	for result.Next() {
		record := result.Record()
		aliveResult := &AliveResult{
			IP:       record.ValueByKey("ip").(string),
			CIDR:     record.ValueByKey("cidr").(string),
			IsAlive:  record.Value().(bool),
			ScanTime: record.Time(),
		}
		results = append(results, aliveResult)
	}

	return results, result.Err()
}

// DeleteOldData 删除旧数据
func (c *InfluxDBClient) DeleteOldData(bucket string, days int) error {
	deleteAPI := c.client.DeleteAPI()

	start := time.Now().AddDate(0, 0, -days)
	stop := time.Now()

	return deleteAPI.DeleteWithName(context.Background(), c.organization, bucket, start, stop, "")
}
