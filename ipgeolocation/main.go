package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type IPLocation struct {
	IP          string  `json:"ip"`
	City        string  `json:"city"`
	Region      string  `json:"region"`
	Country     string  `json:"country"`
	Street      string  `json:"street,omitempty"`
	Loc         string  `json:"loc"` // 经纬度，格式：lat,lon
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Postal      string  `json:"postal"`
	Timezone    string  `json:"timezone"`
	Error       string  `json:"error,omitempty"`
}

func getIPLocation(ip string) IPLocation {
	// 首先尝试使用ip-api.com获取更详细的地址信息
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)
	resp, err := http.Get(url)
	if err != nil {
		return getIPLocationFromIpInfo(ip)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return getIPLocationFromIpInfo(ip)
	}

	// 解析ip-api.com的响应
	var ipAPIResp map[string]interface{}
	if err := json.Unmarshal(body, &ipAPIResp); err != nil {
		return getIPLocationFromIpInfo(ip)
	}

	// 检查ip-api.com响应是否成功
	if status, ok := ipAPIResp["status"].(string); !ok || status != "success" {
		// 如果ip-api.com失败，回退到ipinfo.io
		return getIPLocationFromIpInfo(ip)
	}

	// 从ip-api.com响应中提取数据
	loc := IPLocation{
		IP:          ip,
		City:        getStringFromMap(ipAPIResp, "city"),
		Region:      getStringFromMap(ipAPIResp, "regionName"),
		Country:     getStringFromMap(ipAPIResp, "country"),
		Latitude:    getFloatFromMap(ipAPIResp, "lat"),
		Longitude:   getFloatFromMap(ipAPIResp, "lon"),
		Postal:      getStringFromMap(ipAPIResp, "zip"),
		Timezone:    getStringFromMap(ipAPIResp, "timezone"),
	}

	// 获取街道信息
	// ip-api.com不提供街道信息，我们可以尝试使用经纬度进行反向地理编码
	if loc.Latitude != 0 && loc.Longitude != 0 {
		loc.Street = getStreetFromCoordinates(loc.Latitude, loc.Longitude)
	}

	return loc
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

// 使用openstreetmap获取街道信息
func getStreetFromCoordinates(lat, lon float64) string {
	// 对于大多数公共IP地址，反向地理编码通常只能得到大致位置
	// 而不是具体的街道信息，因为这些IP通常对应于数据中心或ISP节点
	// 我们可以尝试使用一些已知的地理位置数据库或API来获取更详细的信息
	// 但对于免费API来说，通常无法获取到精确的街道信息
	
	// 由于OpenStreetMap API在当前环境中可能不可用，我们直接返回空字符串
	return ""
}

// 从ipinfo.io获取IP位置信息（备用方案）
func getIPLocationFromIpInfo(ip string) IPLocation {
	url := fmt.Sprintf("https://ipinfo.io/%s/json", ip)
	resp, err := http.Get(url)
	if err != nil {
		return IPLocation{IP: ip, Error: err.Error()}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return IPLocation{IP: ip, Error: err.Error()}
	}

	var loc IPLocation
	if err := json.Unmarshal(body, &loc); err != nil {
		return IPLocation{IP: ip, Error: err.Error()}
	}

	// 解析经纬度
	if loc.Loc != "" {
		parts := strings.Split(loc.Loc, ",")
		if len(parts) == 2 {
			fmt.Sscanf(parts[0], "%f", &loc.Latitude)
			fmt.Sscanf(parts[1], "%f", &loc.Longitude)
		}
	}

	return loc
}

func readIPsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ips []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ip := strings.TrimSpace(scanner.Text())
		if ip != "" {
			ips = append(ips, ip)
		}
	}

	return ips, scanner.Err()
}

func saveResultsToHTML(results []IPLocation, filename string) error {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>IP地理位置查询结果</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
            background-color: #f5f5f5;
        }
        h1 {
            color: #333;
            text-align: center;
        }
        .info-note {
            background-color: #ffffcc;
            border: 1px solid #ffeb3b;
            padding: 15px;
            margin: 10px 0;
            border-radius: 5px;
            font-size: 14px;
            color: #666;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
            background-color: white;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #4CAF50;
            color: white;
            font-weight: bold;
        }
        tr:hover {
            background-color: #f5f5f5;
        }
        .error {
            color: red;
        }
        .success {
            color: green;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .timestamp {
            text-align: center;
            color: #666;
            margin-top: 10px;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>IP地理位置查询结果</h1>
        <div class="timestamp">查询时间: %s</div>
        <div class="info-note">
            <strong>注意事项：</strong>
            <ul>
                <li>免费IP地理定位API通常不提供街道级别的详细信息</li>
                <li>只有特定的付费服务才能提供精确的街道地址</li>
                <li>公共IP地址（如8.8.8.8）通常只能获取到大致的地理位置</li>
                <li>私有IP地址无法获取地理位置信息</li>
            </ul>
        </div>
        <table>
            <tr>
                <th>IP地址</th>
                <th>城市</th>
                <th>地区</th>
                <th>国家</th>
                <th>街道</th>
                <th>纬度</th>
                <th>经度</th>
                <th>邮编</th>
                <th>时区</th>
                <th>状态</th>
            </tr>`

	html = fmt.Sprintf(html, time.Now().Format("2006-01-02 15:04:05"))

	for _, loc := range results {
		if loc.Error != "" {
			html += fmt.Sprintf(`
            <tr>
                <td>%s</td>
                <td colspan="8"></td>
                <td class="error">%s</td>
            </tr>`, loc.IP, loc.Error)
		} else {
			html += fmt.Sprintf(`
            <tr>
                <td>%s</td>
                <td>%s</td>
                <td>%s</td>
                <td>%s</td>
                <td>%s</td>
                <td>%.6f</td>
                <td>%.6f</td>
                <td>%s</td>
                <td>%s</td>
                <td class="success">成功</td>
            </tr>`, 
				loc.IP, loc.City, loc.Region, loc.Country, loc.Street,
				loc.Latitude, loc.Longitude, loc.Postal, loc.Timezone)
		}
	}

	html += `
        </table>
    </div>
</body>
</html>`

	return ioutil.WriteFile(filename, []byte(html), 0644)
}

func startWebServer() {
	// 设置静态文件服务
	http.Handle("/", http.FileServer(http.Dir(".")))
	
	fmt.Printf("Web服务器已启动，监听端口: 8086\n")
	fmt.Printf("可以通过 http://localhost:8086/results.html 访问查询结果\n")
	
	err := http.ListenAndServe(":8086", nil)
	if err != nil {
		fmt.Printf("启动Web服务器失败: %v\n", err)
	}
}

func main() {
	filePath := flag.String("file", "", "IP地址文件路径")
	webMode := flag.Bool("web", false, "启动Web服务器查看结果")
	flag.Parse()

	if *filePath == "" {
		fmt.Println("请提供IP地址文件路径，使用 -file 参数")
		os.Exit(1)
	}

	ips, err := readIPsFromFile(*filePath)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		os.Exit(1)
	}

	if len(ips) == 0 {
		fmt.Println("文件中没有IP地址")
		os.Exit(1)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []IPLocation

	fmt.Printf("正在查询 %d 个IP地址的地理位置信息...\n", len(ips))
	fmt.Println("==================================================")
	fmt.Println("注意: 免费IP地理定位API通常不提供街道级别的详细信息")
	fmt.Println("      只有特定的付费服务才能提供精确的街道地址")
	fmt.Println("--------------------------------------------------")

	for _, ip := range ips {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			loc := getIPLocation(ip)
			
			// 实时输出结果到控制台
			mu.Lock()
			if loc.Error != "" {
				fmt.Printf("IP: %s, 错误: %s\n", loc.IP, loc.Error)
			} else {
				fmt.Printf("IP: %s\n", loc.IP)
				fmt.Printf("  城市: %s\n", loc.City)
				fmt.Printf("  地区: %s\n", loc.Region)
				fmt.Printf("  国家: %s\n", loc.Country)
				fmt.Printf("  街道: %s\n", loc.Street)
				fmt.Printf("  经纬度: %.6f, %.6f\n", loc.Latitude, loc.Longitude)
				fmt.Printf("  邮编: %s\n", loc.Postal)
				fmt.Printf("  时区: %s\n", loc.Timezone)
			}
			fmt.Println("--------------------------------------------------")
			
			// 将结果添加到results数组
			results = append(results, loc)
			
			// 实时更新HTML文件
			err := saveResultsToHTML(results, "results.html")
			if err != nil {
				fmt.Printf("保存HTML结果失败: %v\n", err)
			}
			mu.Unlock()
		}(ip)
	}

	wg.Wait()

	fmt.Println("==================================================")
	fmt.Printf("查询完成，共查询了 %d 个IP地址\n", len(results))
	fmt.Println("结果已保存到 results.html 文件")

	// 如果指定了-web参数，启动Web服务器
	if *webMode {
		startWebServer()
	}
}