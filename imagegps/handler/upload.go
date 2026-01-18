package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"imagegps/utils"

	"github.com/gin-gonic/gin"
)

// UploadResponse 上传响应结构
type UploadResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    *utils.GPSInfo   `json:"data,omitempty"`
}

// UploadImageHandler 处理图片上传
func UploadImageHandler(c *gin.Context) {
	// 设置请求上下文超时
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
	defer cancel()

	// 使用超时上下文替换原始请求上下文
	c.Request = c.Request.WithContext(ctx)

	// 快速获取上传的文件
	file, err := c.FormFile("image")
	if err != nil {
		fmt.Printf("获取上传文件失败: %v\n", err)
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Message: "未检测到上传的图片文件",
		})
		return
	}

	// 快速验证文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".tiff": true,
		".tif":  true,
	}

	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Message: fmt.Sprintf("不支持的文件格式: %s，仅支持 JPG, JPEG, PNG, TIFF", ext),
		})
		return
	}

	// 快速验证文件大小 (最大20MB)
	if file.Size > 20*1024*1024 {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Message: "文件大小超过限制（最大20MB）",
		})
		return
	}

	// 提取GPS信息（并行处理文件保存和GPS提取）
	go func() {
		// 创建保存上传文件的目录
		uploadDir := "./uploads"
		if err := os.MkdirAll(uploadDir, 0755); err == nil {
			// 生成唯一的文件名
			timestamp := time.Now().UnixNano() / int64(time.Millisecond)
			newFilename := fmt.Sprintf("%d%s", timestamp, ext)
			uploadPath := filepath.Join(uploadDir, newFilename)

			// 保存文件
			if err := c.SaveUploadedFile(file, uploadPath); err != nil {
				fmt.Printf("保存文件失败: %v\n", err)
			} else {
				fmt.Printf("文件已保存: %s\n", uploadPath)
			}
		}
	}()

	// 提取GPS信息
	gpsInfo, err := utils.ExtractGPSFromFile(file)
	if err != nil {
		fmt.Printf("提取GPS信息失败: %v\n", err)
		c.JSON(http.StatusOK, UploadResponse{
			Success: false,
			Message: fmt.Sprintf("无法提取GPS信息: %v", err),
		})
		return
	}

	// 检查是否包含GPS信息
	if !gpsInfo.HasGPS {
		c.JSON(http.StatusOK, UploadResponse{
			Success: false,
			Message: "该图片不包含GPS地理位置信息，可能原因：\n1. 拍摄时未开启GPS定位\n2. 图片经过处理，EXIF信息被删除\n3. 设备不支持GPS定位",
			Data:    gpsInfo,
		})
		return
	}

	// 成功提取GPS信息
	c.JSON(http.StatusOK, UploadResponse{
		Success: true,
		Message: "成功提取GPS位置信息",
		Data:    gpsInfo,
	})
}

// IndexHandler 首页处理器
func IndexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "图片GPS地理位置提取系统",
	})
}

// HealthHandler 健康检查
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"service": "imagegps",
	})
}
