package main

import (
	"context"
	"fmt"
	"log"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

// ClearInfluxData 清除InfluxDB中的所有数据
func ClearInfluxData() {
	log.Println("=== 清除InfluxDB中的所有数据 ===")

	// InfluxDB配置
	url := "http://121.43.143.169:8086"
	token := "rs1BHmkQZ21MGKq2OcIqdhSV123abK9--VfXT8AMLNEQVInd0-8xGEBIkRA5DW37cMpSFbj-E4QtKqPSMxiXWA=="
	org := "my-org"

	// 创建InfluxDB客户端
	client := influxdb2.NewClient(url, token)
	defer client.Close()

	// 获取查询API
	queryAPI := client.QueryAPI(org)

	// 构建删除查询 - 删除所有数据
	query := fmt.Sprintf(`DELETE FROM %s`, "taiwan_ip_scan")

	log.Printf("执行删除查询: %s", query)

	// 执行查询
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		log.Fatalf("执行删除查询失败: %v", err)
	}

	// 检查结果
	for result.Next() {
		log.Printf("删除操作结果: %v", result.Record())
	}

	if result.Err() != nil {
		log.Fatalf("处理删除结果失败: %v", result.Err())
	}

	// 删除IP段状态数据
	query = fmt.Sprintf(`DELETE FROM %s`, "ip_segment_status")
	log.Printf("执行删除查询: %s", query)

	result, err = queryAPI.Query(context.Background(), query)
	if err != nil {
		log.Printf("执行删除IP段状态查询失败: %v", err)
	} else {
		for result.Next() {
			log.Printf("删除IP段状态操作结果: %v", result.Record())
		}
		if result.Err() != nil {
			log.Printf("处理删除IP段状态结果失败: %v", result.Err())
		}
	}

	log.Println("=== 数据清除完成 ===")
}

