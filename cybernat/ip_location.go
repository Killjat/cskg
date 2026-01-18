package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang"
)

// IPLocation IP位置信息结构体
type IPLocation struct {
	IP        string  `json:"ip"`        // IP地址
	Country   string  `json:"country"`   // 国家/地区
	Region    string  `json:"region"`    // 省/州
	City      string  `json:"city"`      // 城市
	ISP       string  `json:"isp"`       // ISP
	ASN       uint    `json:"asn"`       // ASN编号
	Latitude  float64 `json:"latitude"`  // 纬度
	Longitude float64 `json:"longitude"` // 经度
}

// IPLocationManager IP位置管理器
type IPLocationManager struct {
	geoipDB     *geoip2.Reader    // GeoIP2数据库读取器
	cache       map[string]*IPLocation // IP位置缓存
	cacheMutex  sync.RWMutex           // 缓存互斥锁
	cacheTTL    time.Duration          // 缓存过期时间
	useMockData bool                   // 是否使用模拟数据
}

// NewIPLocationManager 创建新的IP位置管理器
func NewIPLocationManager(geoipDBPath string, cacheTTL time.Duration) *IPLocationManager {
	manager := &IPLocationManager{
		cache:       make(map[string]*IPLocation),
		cacheTTL:    cacheTTL,
		useMockData: false, // 默认不使用模拟数据，尝试使用真实GeoIP2数据
	}

	// 尝试打开GeoIP2数据库
	var dbPath string
	var err error

	if geoipDBPath != "" {
		// 如果指定了数据库路径，直接使用
		dbPath = geoipDBPath
	} else {
		// 否则，尝试自动检测常见的GeoLite2数据库路径
		commonPaths := []string{
			"./GeoLite2-City.mmdb",
			"./GeoLite2-City.mmdb.gz",
			"./GeoLite2-ASN.mmdb",
			"./GeoLite2-ASN.mmdb.gz",
			"/usr/share/GeoIP/GeoLite2-City.mmdb",
			"/usr/local/share/GeoIP/GeoLite2-City.mmdb",
			"/var/lib/GeoIP/GeoLite2-City.mmdb",
			"/usr/share/GeoIP/GeoLite2-ASN.mmdb",
			"/usr/local/share/GeoIP/GeoLite2-ASN.mmdb",
			"/var/lib/GeoIP/GeoLite2-ASN.mmdb",
		}

		// 遍历常见路径，尝试找到存在的GeoLite2数据库文件
		found := false
		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				dbPath = path
				found = true
				log.Printf("自动检测到GeoLite2数据库: %s", dbPath)
				break
			}
		}

		if !found {
			log.Println("未找到GeoLite2数据库文件，将跳过IP位置查询")
			log.Println("建议下载GeoLite2数据库文件:")
			log.Println("1. 访问 https://dev.maxmind.com/geoip/geoip2/geolite2/")
			log.Println("2. 注册并下载 GeoLite2-City.mmdb 和 GeoLite2-ASN.mmdb")
			log.Println("3. 将文件放置在程序目录或 /usr/share/GeoIP/ 目录下")
			return manager
		}
	}

	// 尝试打开GeoIP2数据库
	db, err := geoip2.Open(dbPath)
	if err != nil {
		log.Printf("无法打开GeoLite2数据库 %s: %v，将跳过IP位置查询", dbPath, err)
	} else {
		manager.geoipDB = db
		log.Printf("成功加载GeoLite2数据库: %s", dbPath)
	}

	return manager
}

// Close 关闭IP位置管理器
func (m *IPLocationManager) Close() {
	if m.geoipDB != nil {
		m.geoipDB.Close()
	}
}

// GetIPLocation 获取IP位置信息
func (m *IPLocationManager) GetIPLocation(ipStr string) (*IPLocation, error) {
	// 解析IP地址
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("无效的IP地址: %s", ipStr)
	}

	// 检查缓存
	m.cacheMutex.RLock()
	if location, exists := m.cache[ipStr]; exists {
		m.cacheMutex.RUnlock()
		log.Printf("从缓存获取IP %s 的位置信息", ipStr)
		return location, nil
	}
	m.cacheMutex.RUnlock()

	// 查询IP位置
	location, err := m.queryIPLocation(ip)
	if err != nil {
		return nil, fmt.Errorf("查询IP位置失败: %v", err)
	}

	// 更新缓存
	m.cacheMutex.Lock()
	m.cache[ipStr] = location
	m.cacheMutex.Unlock()

	log.Printf("查询IP %s 的位置信息: %s, %s, %s", ipStr, location.Country, location.Region, location.City)
	return location, nil
}

// BatchGetIPLocation 批量获取IP位置信息
func (m *IPLocationManager) BatchGetIPLocation(ipStrs []string) (map[string]*IPLocation, error) {
	result := make(map[string]*IPLocation, len(ipStrs))

	for _, ipStr := range ipStrs {
		location, err := m.GetIPLocation(ipStr)
		if err != nil {
			log.Printf("批量查询IP %s 位置失败: %v", ipStr, err)
			continue
		}
		result[ipStr] = location
	}

	return result, nil
}

// queryIPLocation 查询IP位置信息
func (m *IPLocationManager) queryIPLocation(ip net.IP) (*IPLocation, error) {
	if m.geoipDB != nil {
		// 使用实际GeoIP2数据库查询
		return m.queryFromGeoIP2(ip)
	}

	// 如果没有GeoIP2数据库，返回基本信息
	return &IPLocation{
		IP:        ip.String(),
		Country:   "Unknown",
		Region:    "Unknown",
		City:      "Unknown",
		ISP:       "Unknown",
		ASN:       0,
		Latitude:  0.0,
		Longitude: 0.0,
	}, nil
}

// queryFromGeoIP2 从GeoIP2数据库查询IP位置
func (m *IPLocationManager) queryFromGeoIP2(ip net.IP) (*IPLocation, error) {
	// 查询城市信息
	cityRecord, err := m.geoipDB.City(ip)
	if err != nil {
		return nil, err
	}

	// 查询ASN信息
	asnRecord, err := m.geoipDB.ASN(ip)
	if err != nil {
		return nil, err
	}

	// 获取国家信息
	country := "Unknown"
	if cityRecord.Country.Names != nil && cityRecord.Country.Names["en"] != "" {
		country = cityRecord.Country.Names["en"]
	}

	// 获取地区信息
	region := "Unknown"
	if len(cityRecord.Subdivisions) > 0 && cityRecord.Subdivisions[0].Names != nil && cityRecord.Subdivisions[0].Names["en"] != "" {
		region = cityRecord.Subdivisions[0].Names["en"]
	}

	// 获取城市信息
	city := "Unknown"
	if cityRecord.City.Names != nil && cityRecord.City.Names["en"] != "" {
		city = cityRecord.City.Names["en"]
	}

	// 获取ISP信息
	isp := "Unknown"
	if asnRecord.AutonomousSystemOrganization != "" {
		isp = asnRecord.AutonomousSystemOrganization
	}

	return &IPLocation{
		IP:        ip.String(),
		Country:   country,
		Region:    region,
		City:      city,
		ISP:       isp,
		ASN:       uint(asnRecord.AutonomousSystemNumber),
		Latitude:  cityRecord.Location.Latitude,
		Longitude: cityRecord.Location.Longitude,
	}, nil
}

// ClearExpiredCache 清理过期缓存
func (m *IPLocationManager) ClearExpiredCache() {
	// 由于我们的缓存没有设置过期时间，这个函数暂时为空
	// 实际实现中应该为每个缓存项添加过期时间
}

// GetCacheSize 获取缓存大小
func (m *IPLocationManager) GetCacheSize() int {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()
	return len(m.cache)
}