package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type StreetLevelLocation struct {
	IP          string  `json:"ip"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	Country     string  `json:"country"`
	Street      string  `json:"street"`
	HouseNumber string  `json:"house_number"`
	Postal      string  `json:"postal"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
	Error       string  `json:"error,omitempty"`
}

func main() {
	// 测试IP地址
	ip := "8.8.8.8"
	
	fmt.Printf("正在查询IP地址 %s 的地理位置信息...\n\n", ip)
	
	// 使用ip-api.com获取基本信息
	basicInfo := getBasicInfoFromIpApi(ip)
	fmt.Println("=== 基本地理信息 ===")
	fmt.Printf("IP: %s\n", basicInfo.IP)
	fmt.Printf("城市: %s\n", basicInfo.City)
	fmt.Printf("地区: %s\n", basicInfo.Region)
	fmt.Printf("国家: %s\n", basicInfo.Country)
	fmt.Printf("经纬度: %.6f, %.6f\n", basicInfo.Latitude, basicInfo.Longitude)
	fmt.Printf("邮编: %s\n", basicInfo.Postal)
	fmt.Printf("时区: %s\n", basicInfo.Timezone)
	
	// 尝试使用OpenStreetMap获取街道信息
	streetInfo := getStreetInfoFromOpenStreetMap(basicInfo.Latitude, basicInfo.Longitude)
	fmt.Println("\n=== 街道级别信息 (基于经纬度) ===")
	if streetInfo.Street != "" {
		fmt.Printf("街道: %s\n", streetInfo.Street)
		fmt.Printf("门牌号: %s\n", streetInfo.HouseNumber)
	} else {
		fmt.Println("无法获取街道级别信息")
	}
	
	fmt.Println("\n=== 注意事项 ===")
	fmt.Println("1. 免费IP地理定位API通常不提供街道级别的详细信息")
	fmt.Println("2. 街道级别的IP定位需要更精确的数据库和更多的资源")
	fmt.Println("3. 对于公共IP地址（如8.8.8.8），只能获取到大致的地理位置")
	fmt.Println("4. 只有特定的付费IP地理定位服务才能提供街道级别的详细信息")
	fmt.Println("5. 对于私有IP地址，无法获取地理位置信息")
}

// 从ip-api.com获取基本信息
func getBasicInfoFromIpApi(ip string) StreetLevelLocation {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	resp, err := http.Get(url)
	if err != nil {
		return StreetLevelLocation{IP: ip, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return StreetLevelLocation{IP: ip, Error: err.Error()}
	}

	var ipAPIResp map[string]interface{}
	if err := json.Unmarshal(body, &ipAPIResp); err != nil {
		return StreetLevelLocation{IP: ip, Error: err.Error()}
	}

	// 检查ip-api.com响应是否成功
	if status, ok := ipAPIResp["status"].(string); !ok || status != "success" {
		return StreetLevelLocation{IP: ip, Error: "ip-api.com API调用失败"}
	}

	return StreetLevelLocation{
		IP:        ip,
		City:      getStringFromMap(ipAPIResp, "city"),
		Region:    getStringFromMap(ipAPIResp, "regionName"),
		Country:   getStringFromMap(ipAPIResp, "country"),
		Postal:    getStringFromMap(ipAPIResp, "zip"),
		Latitude:  getFloatFromMap(ipAPIResp, "lat"),
		Longitude: getFloatFromMap(ipAPIResp, "lon"),
		Timezone:  getStringFromMap(ipAPIResp, "timezone"),
	}
}

// 从OpenStreetMap获取街道信息
func getStreetInfoFromOpenStreetMap(lat, lon float64) StreetLevelLocation {
	if lat == 0 || lon == 0 {
		return StreetLevelLocation{Error: "无效的经纬度"}
	}
	
	reverseGeocodeURL := fmt.Sprintf(
		"https://nominatim.openstreetmap.org/reverse?format=json&lat=%.6f&lon=%.6f&zoom=18&addressdetails=1",
		lat, lon)
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", reverseGeocodeURL, nil)
	if err != nil {
		return StreetLevelLocation{Error: err.Error()}
	}
	req.Header.Set("User-Agent", "IPGeolocationTool/1.0")
	
	resp, err := client.Do(req)
	if err != nil {
		return StreetLevelLocation{Error: err.Error()}
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return StreetLevelLocation{Error: err.Error()}
	}
	
	var reverseGeocodeResult map[string]interface{}
	if err := json.Unmarshal(body, &reverseGeocodeResult); err != nil {
		return StreetLevelLocation{Error: err.Error()}
	}
	
	streetLoc := StreetLevelLocation{
		Latitude:  lat,
		Longitude: lon,
	}
	
	if address, ok := reverseGeocodeResult["address"].(map[string]interface{}); ok {
		streetLoc.Street = getStringFromMap(address, "road")
		if streetLoc.Street == "" {
			streetLoc.Street = getStringFromMap(address, "street")
		}
		streetLoc.HouseNumber = getStringFromMap(address, "house_number")
	}
	
	return streetLoc
}

// 从map中安全获取字符串值
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// 从map中安全获取浮点数值
func getFloatFromMap(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0
}
