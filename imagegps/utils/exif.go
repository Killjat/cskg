package utils

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

// min 辅助函数，返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GPSInfo GPS信息结构
type GPSInfo struct {
	Latitude      float64 `json:"latitude"`       // 纬度
	Longitude     float64 `json:"longitude"`      // 经度
	Altitude      float64 `json:"altitude"`       // 海拔
	LatitudeRef   string  `json:"latitude_ref"`   // 纬度参考 (N/S)
	LongitudeRef  string  `json:"longitude_ref"`  // 经度参考 (E/W)
	HasGPS        bool    `json:"has_gps"`        // 是否包含GPS信息
	DateTime      string  `json:"datetime"`       // 拍摄时间
	Make          string  `json:"make"`           // 相机制造商
	Model         string  `json:"model"`          // 相机型号
	GoogleMapURL  string  `json:"google_map_url"` // Google地图链接
	BaiduMapURL   string  `json:"baidu_map_url"`  // 百度地图链接
}

// ExtractGPSFromFile 从上传的文件中提取GPS信息
func ExtractGPSFromFile(file *multipart.FileHeader) (*GPSInfo, error) {
	fmt.Printf("收到文件上传请求: 文件名=%s, 大小=%d字节, 类型=%s\n", file.Filename, file.Size, file.Header.Get("Content-Type"))

	// 快速检查是否为截图
	if strings.Contains(strings.ToLower(file.Filename), "screenshot") || strings.Contains(strings.ToLower(file.Filename), "截屏") {
		fmt.Printf("检测到截图文件，可能不包含GPS信息\n")
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		fmt.Printf("无法打开文件: %v\n", err)
		return nil, fmt.Errorf("无法打开文件: %v", err)
	}
	defer src.Close()

	// 验证文件大小
	if file.Size == 0 {
		fmt.Printf("文件大小为0，可能是上传不完整\n")
		return nil, fmt.Errorf("文件大小为0，可能是上传不完整")
	}

	// 初始化GPS信息
	info := &GPSInfo{
		HasGPS: false,
	}

	// 直接完整读取文件进行解析，确保所有EXIF数据都能被读取
	fmt.Println("开始完整解析文件...")
	x, err := exif.Decode(src)
	if err != nil {
		// 如果第一次解析失败，尝试重新读取文件
		fmt.Printf("第一次解析EXIF失败: %v，尝试重新读取文件\n", err)
		if _, err := src.Seek(0, io.SeekStart); err == nil {
			// 读取所有内容到缓冲区
			buffer, readErr := io.ReadAll(src)
			if readErr == nil {
				fmt.Printf("成功读取完整文件，大小: %d 字节\n", len(buffer))
				x, err = exif.Decode(bytes.NewReader(buffer))
				if err != nil {
					fmt.Printf("从缓冲区解析EXIF也失败: %v\n", err)
				}
			} else {
				fmt.Printf("重新读取文件失败: %v\n", readErr)
			}
		}
		if err != nil {
			fmt.Printf("完整解析EXIF失败: %v\n", err)
			// 如果无法解析EXIF数据，仍然返回基本信息
			return info, fmt.Errorf("无法解析EXIF数据: %v", err)
		}
	}

	fmt.Println("EXIF数据解析成功")

	// 处理EXIF数据
	processEXIFData(x, info)

	return info, nil
}

// processEXIFData 处理EXIF数据，提取GPS信息
func processEXIFData(x *exif.Exif, info *GPSInfo) {
	// 专门检查GPS相关标签
	fmt.Println("\n=== GPS相关标签 ===")

	// 检查所有可能的GPS标签，包括非标准标签
	fmt.Println("检查所有可能的GPS标签:")

	// 尝试获取GPSLatitude
	lat, err := x.Get(exif.GPSLatitude)
	if err != nil {
		fmt.Printf("GPSLatitude: 不存在\n")
		// 尝试使用字符串键获取
		lat, err = x.Get("GPSLatitude")
		if err != nil {
			fmt.Printf("GPSLatitude (字符串键): 不存在\n")
		} else {
			fmt.Printf("GPSLatitude (字符串键): %v\n", lat)
			// 尝试获取GPSLatitudeRef
			latRef, err := x.Get("GPSLatitudeRef")
			if err != nil {
				fmt.Printf("GPSLatitudeRef (字符串键): 不存在\n")
			} else {
				fmt.Printf("GPSLatitudeRef (字符串键): %v\n", latRef)
				if latRefStr, err := latRef.StringVal(); err == nil {
					info.Latitude = convertToDecimal(lat)
					info.LatitudeRef = latRefStr
					if info.LatitudeRef == "S" {
						info.Latitude = -info.Latitude
					}
					info.HasGPS = true
				}
			}
		}
	} else {
		fmt.Printf("GPSLatitude: %v\n", lat)
		// GPSLatitudeRef
		latRef, err := x.Get(exif.GPSLatitudeRef)
		if err != nil {
			fmt.Printf("GPSLatitudeRef: 不存在\n")
			// 尝试使用字符串键获取
			latRef, err = x.Get("GPSLatitudeRef")
			if err != nil {
				fmt.Printf("GPSLatitudeRef (字符串键): 不存在\n")
				// 尝试不使用Ref值解析（有些图片可能格式不标准）
				info.Latitude = convertToDecimal(lat)
				info.HasGPS = true
			} else {
				fmt.Printf("GPSLatitudeRef (字符串键): %v\n", latRef)
				if latRefStr, err := latRef.StringVal(); err == nil {
					info.Latitude = convertToDecimal(lat)
					info.LatitudeRef = latRefStr
					if info.LatitudeRef == "S" {
						info.Latitude = -info.Latitude
					}
					info.HasGPS = true
				}
			}
		} else {
			if latRefStr, err := latRef.StringVal(); err == nil {
				fmt.Printf("GPSLatitudeRef: %s\n", latRefStr)
				info.Latitude = convertToDecimal(lat)
				info.LatitudeRef = latRefStr
				if info.LatitudeRef == "S" {
					info.Latitude = -info.Latitude
				}
				info.HasGPS = true
			} else {
				fmt.Printf("获取GPSLatitudeRef字符串失败: %v\n", err)
			}
		}
	}

	// GPSLongitude
	lon, err := x.Get(exif.GPSLongitude)
	if err != nil {
		fmt.Printf("GPSLongitude: 不存在\n")
		// 尝试使用字符串键获取
		lon, err = x.Get("GPSLongitude")
		if err != nil {
			fmt.Printf("GPSLongitude (字符串键): 不存在\n")
		} else {
			fmt.Printf("GPSLongitude (字符串键): %v\n", lon)
			// 尝试获取GPSLongitudeRef
			lonRef, err := x.Get("GPSLongitudeRef")
			if err != nil {
				fmt.Printf("GPSLongitudeRef (字符串键): 不存在\n")
			} else {
				fmt.Printf("GPSLongitudeRef (字符串键): %v\n", lonRef)
				if lonRefStr, err := lonRef.StringVal(); err == nil {
					info.Longitude = convertToDecimal(lon)
					info.LongitudeRef = lonRefStr
					if info.LongitudeRef == "W" {
						info.Longitude = -info.Longitude
					}
					info.HasGPS = true
				}
			}
		}
	} else {
		fmt.Printf("GPSLongitude: %v\n", lon)
		// GPSLongitudeRef
		lonRef, err := x.Get(exif.GPSLongitudeRef)
		if err != nil {
			fmt.Printf("GPSLongitudeRef: 不存在\n")
			// 尝试使用字符串键获取
			lonRef, err = x.Get("GPSLongitudeRef")
			if err != nil {
				fmt.Printf("GPSLongitudeRef (字符串键): 不存在\n")
				// 尝试不使用Ref值解析（有些图片可能格式不标准）
				info.Longitude = convertToDecimal(lon)
				info.HasGPS = true
			} else {
				fmt.Printf("GPSLongitudeRef (字符串键): %v\n", lonRef)
				if lonRefStr, err := lonRef.StringVal(); err == nil {
					info.Longitude = convertToDecimal(lon)
					info.LongitudeRef = lonRefStr
					if info.LongitudeRef == "W" {
						info.Longitude = -info.Longitude
					}
					info.HasGPS = true
				}
			}
		} else {
			if lonRefStr, err := lonRef.StringVal(); err == nil {
				fmt.Printf("GPSLongitudeRef: %s\n", lonRefStr)
				info.Longitude = convertToDecimal(lon)
				info.LongitudeRef = lonRefStr
				if info.LongitudeRef == "W" {
					info.Longitude = -info.Longitude
				}
				info.HasGPS = true
			} else {
				fmt.Printf("获取GPSLongitudeRef字符串失败: %v\n", err)
			}
		}
	}

	// 提取海拔
	alt, err := x.Get(exif.GPSAltitude)
	if err != nil {
		fmt.Printf("GPSAltitude: 不存在\n")
		// 尝试使用字符串键获取
		alt, err = x.Get("GPSAltitude")
		if err != nil {
			fmt.Printf("GPSAltitude (字符串键): 不存在\n")
		} else {
			fmt.Printf("GPSAltitude (字符串键): %v\n", alt)
		}
	} else {
		fmt.Printf("GPSAltitude: %v\n", alt)
		num, denom, err := alt.Rat2(0)
		if err == nil && denom != 0 {
			info.Altitude = float64(num) / float64(denom)
			// 检查海拔参考
			altRef, err := x.Get(exif.GPSAltitudeRef)
			if err != nil {
				// 尝试使用字符串键获取
				altRef, err = x.Get("GPSAltitudeRef")
			}
			if err == nil {
				fmt.Printf("GPSAltitudeRef: %v\n", altRef)
				if altRefVal, err := altRef.Int(0); err == nil && altRefVal == 1 {
					// 如果海拔参考为1，表示低于海平面，海拔为负
					info.Altitude = -info.Altitude
				}
			}
		} else {
			fmt.Printf("解析海拔失败: %v\n", err)
		}
	}

	// 检查GPS日期时间
	gpsTime, err := x.Get("GPSDateTime")
	if err != nil {
		fmt.Printf("GPSDateTime: 不存在\n")
	} else {
		fmt.Printf("GPSDateTime: %v\n", gpsTime)
	}

	// 相机制造商和型号
	makeTag, err := x.Get("Make")
	if err == nil {
		if makeStr, err := makeTag.StringVal(); err == nil {
			info.Make = makeStr
			fmt.Printf("Make: %s\n", makeStr)
		}
	} else {
		fmt.Printf("Make: 不存在\n")
	}

	modelTag, err := x.Get("Model")
	if err == nil {
		if modelStr, err := modelTag.StringVal(); err == nil {
			info.Model = modelStr
			fmt.Printf("Model: %s\n", modelStr)
		}
	} else {
		fmt.Printf("Model: 不存在\n")
	}

	// 提取拍摄时间
	dateTime, err := x.DateTime()
	if err == nil {
		info.DateTime = dateTime.Format("2006-01-02 15:04:05")
		fmt.Printf("DateTime: %s\n", info.DateTime)
	} else {
		fmt.Printf("DateTime: 不存在\n")
		// 尝试直接获取DateTimeOriginal
		dateTimeTag, err := x.Get("DateTimeOriginal")
		if err == nil {
			info.DateTime, _ = dateTimeTag.StringVal()
			fmt.Printf("DateTimeOriginal: %s\n", info.DateTime)
		} else {
			fmt.Printf("DateTimeOriginal: 不存在\n")
			// 尝试获取普通DateTime
			dateTimeTag, err := x.Get("DateTime")
			if err == nil {
				info.DateTime, _ = dateTimeTag.StringVal()
				fmt.Printf("DateTime: %s\n", info.DateTime)
			} else {
				fmt.Printf("获取所有日期时间标签失败: %v\n", err)
			}
		}
	}

	// 生成地图链接
	if info.HasGPS {
		info.GoogleMapURL = fmt.Sprintf("https://www.google.com/maps?q=%.6f,%.6f", info.Latitude, info.Longitude)
		info.BaiduMapURL = fmt.Sprintf("https://api.map.baidu.com/marker?location=%.6f,%.6f&title=拍摄位置&content=从图片提取的位置&output=html", info.Latitude, info.Longitude)
		fmt.Printf("生成地图链接: GoogleMap=%s\n", info.GoogleMapURL)
	}
}

// ExtractGPS 从io.Reader中提取GPS信息
func ExtractGPS(reader io.Reader) (*GPSInfo, error) {
	info := &GPSInfo{
		HasGPS: false,
	}

	fmt.Println("开始解析EXIF数据")

	// 多次尝试解析EXIF数据，最多3次
	maxRetries := 3
	var x *exif.Exif
	var decodeErr error

	// 首先读取所有内容到缓冲区，以便可以多次尝试解析
	buffer, err := io.ReadAll(reader)
	if err != nil {
		fmt.Printf("读取数据失败: %v\n", err)
		return info, fmt.Errorf("读取数据失败: %v", err)
	}

	if len(buffer) == 0 {
		fmt.Println("读取到的数据为空")
		return info, fmt.Errorf("读取到的数据为空")
	}

	fmt.Printf("成功读取数据，大小: %d 字节\n", len(buffer))

	for i := 0; i < maxRetries; i++ {
		fmt.Printf("第 %d 次尝试解析EXIF数据\n", i+1)
		// 创建新的reader
		r := bytes.NewReader(buffer)
		// 解析EXIF数据
		x, decodeErr = exif.Decode(r)
		if decodeErr == nil {
			fmt.Println("EXIF数据解析成功")
			break
		}
		fmt.Printf("第 %d 次解析EXIF失败: %v\n", i+1, decodeErr)
	}

	if decodeErr != nil {
		fmt.Printf("所有尝试都失败，无法解析EXIF数据: %v\n", decodeErr)
		// 如果无法解析EXIF数据，仍然返回基本信息
		return info, fmt.Errorf("无法解析EXIF数据: %v", decodeErr)
	}

	// 处理EXIF数据
	processEXIFData(x, info)

	return info, nil
}

// convertToDecimal 将度分秒格式转换为十进制度数
func convertToDecimal(tag *tiff.Tag) float64 {
	if tag == nil {
		return 0
	}

	// GPS坐标通常以 [度, 分, 秒] 的形式存储
	degrees, degreeDenom, _ := tag.Rat2(0)
	minutes, minuteDenom, _ := tag.Rat2(1)
	seconds, secondDenom, _ := tag.Rat2(2)

	degreesDecimal := 0.0
	if degreeDenom != 0 {
		degreesDecimal = float64(degrees) / float64(degreeDenom)
	}

	minutesDecimal := 0.0
	if minuteDenom != 0 {
		minutesDecimal = float64(minutes) / float64(minuteDenom)
	}

	secondsDecimal := 0.0
	if secondDenom != 0 {
		secondsDecimal = float64(seconds) / float64(secondDenom)
	}

	// 转换为十进制度数
	result := degreesDecimal + (minutesDecimal / 60.0) + (secondsDecimal / 3600.0)
	return math.Round(result*1000000) / 1000000 // 保留6位小数
}
