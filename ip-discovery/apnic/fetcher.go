package apnic

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Fetcher APNIC数据获取器
type Fetcher struct {
	URL       string
	CacheFile string
	CacheHours int
}

// NewFetcher 创建新的APNIC数据获取器
func NewFetcher(url, cacheFile string, cacheHours int) *Fetcher {
	return &Fetcher{
		URL:        url,
		CacheFile:  cacheFile,
		CacheHours: cacheHours,
	}
}

// FetchData 获取APNIC数据
func (f *Fetcher) FetchData() (string, error) {
	// 检查缓存文件是否存在且未过期
	if f.isCacheValid() {
		fmt.Printf("使用缓存文件: %s\n", f.CacheFile)
		return f.CacheFile, nil
	}

	// 创建缓存目录
	if err := os.MkdirAll(filepath.Dir(f.CacheFile), 0755); err != nil {
		return "", fmt.Errorf("创建缓存目录失败: %v", err)
	}

	// 下载数据
	fmt.Printf("从 %s 下载APNIC数据...\n", f.URL)
	if err := f.downloadFile(); err != nil {
		return "", fmt.Errorf("下载APNIC数据失败: %v", err)
	}

	fmt.Printf("APNIC数据已保存到: %s\n", f.CacheFile)
	return f.CacheFile, nil
}

// isCacheValid 检查缓存是否有效
func (f *Fetcher) isCacheValid() bool {
	info, err := os.Stat(f.CacheFile)
	if err != nil {
		return false
	}

	// 检查文件是否在有效期内
	expireTime := info.ModTime().Add(time.Duration(f.CacheHours) * time.Hour)
	return time.Now().Before(expireTime)
}

// downloadFile 下载文件
func (f *Fetcher) downloadFile() error {
	// 创建HTTP请求
	resp, err := http.Get(f.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP请求失败: %s", resp.Status)
	}

	// 创建文件
	file, err := os.Create(f.CacheFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// 复制数据
	_, err = io.Copy(file, resp.Body)
	return err
}

// GetCacheInfo 获取缓存信息
func (f *Fetcher) GetCacheInfo() (bool, time.Time, error) {
	info, err := os.Stat(f.CacheFile)
	if err != nil {
		return false, time.Time{}, err
	}

	return f.isCacheValid(), info.ModTime(), nil
}