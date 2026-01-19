package storage

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type            string        `yaml:"type"`
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	SQLitePath      string        `yaml:"sqlite_path"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// Database 数据库操作类
type Database struct {
	db *gorm.DB
}

// NewDatabase 创建数据库连接
func NewDatabase(config DatabaseConfig) (*Database, error) {
	var dialector gorm.Dialector

	switch config.Type {
	case "sqlite":
		dialector = sqlite.Open(config.SQLitePath)
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Username, config.Password, config.Host, config.Port, config.Database)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			config.Host, config.Username, config.Password, config.Database, config.Port)
		dialector = postgres.Open(dsn)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	// 自动迁移
	if err := db.AutoMigrate(
		&APIEndpoint{},
		&CrawlSession{},
		&CrawledPage{},
		&JSFile{},
		&FormInfo{},
		&APIPattern{},
	); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %v", err)
	}

	return &Database{db: db}, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// CreateSession 创建爬取会话
func (d *Database) CreateSession(session *CrawlSession) error {
	return d.db.Create(session).Error
}

// UpdateSession 更新会话状态
func (d *Database) UpdateSession(sessionID, status string, endTime *time.Time) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if endTime != nil {
		updates["end_time"] = endTime
	}
	return d.db.Model(&CrawlSession{}).Where("session_id = ?", sessionID).Updates(updates).Error
}

// GetSessions 获取会话列表
func (d *Database) GetSessions(limit, offset int) ([]CrawlSession, error) {
	var sessions []CrawlSession
	err := d.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&sessions).Error
	return sessions, err
}

// GetSession 获取单个会话
func (d *Database) GetSession(sessionID string) (*CrawlSession, error) {
	var session CrawlSession
	err := d.db.Where("session_id = ?", sessionID).First(&session).Error
	return &session, err
}

// SaveAPIEndpoint 保存API端点
func (d *Database) SaveAPIEndpoint(api *APIEndpoint) error {
	// 检查是否已存在
	var existing APIEndpoint
	if err := d.db.Where("url = ? AND method = ?", api.URL, api.Method).First(&existing).Error; err == nil {
		// 更新现有记录
		api.ID = existing.ID
		return d.db.Save(api).Error
	}
	return d.db.Create(api).Error
}

// GetAPIEndpoints 获取API端点列表
func (d *Database) GetAPIEndpoints(sessionID string, limit, offset int) ([]APIEndpoint, error) {
	var apis []APIEndpoint
	query := d.db.Order("created_at DESC").Limit(limit).Offset(offset)
	
	if sessionID != "" {
		// 通过关联的页面查找API
		query = query.Joins("JOIN crawled_pages ON crawled_pages.url LIKE CONCAT('%', api_endpoints.domain, '%')").
			Where("crawled_pages.session_id = ?", sessionID)
	}
	
	err := query.Find(&apis).Error
	return apis, err
}

// SaveCrawledPage 保存已爬取页面
func (d *Database) SaveCrawledPage(page *CrawledPage) error {
	return d.db.Create(page).Error
}

// IsPageCrawled 检查页面是否已爬取
func (d *Database) IsPageCrawled(sessionID, url string) bool {
	var count int64
	d.db.Model(&CrawledPage{}).Where("session_id = ? AND url = ?", sessionID, url).Count(&count)
	return count > 0
}

// GetCrawledPages 获取已爬取页面列表
func (d *Database) GetCrawledPages(sessionID string, limit, offset int) ([]CrawledPage, error) {
	var pages []CrawledPage
	err := d.db.Where("session_id = ?", sessionID).Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&pages).Error
	return pages, err
}

// SaveJSFile 保存JavaScript文件
func (d *Database) SaveJSFile(jsFile *JSFile) error {
	// 检查是否已存在
	var existing JSFile
	if err := d.db.Where("session_id = ? AND url = ?", jsFile.SessionID, jsFile.URL).First(&existing).Error; err == nil {
		jsFile.ID = existing.ID
		return d.db.Save(jsFile).Error
	}
	return d.db.Create(jsFile).Error
}

// GetUnanalyzedJSFiles 获取未分析的JS文件
func (d *Database) GetUnanalyzedJSFiles(sessionID string, limit int) ([]JSFile, error) {
	var jsFiles []JSFile
	query := d.db.Where("analyzed = ?", false).Limit(limit)
	
	if sessionID != "" {
		query = query.Where("session_id = ?", sessionID)
	}
	
	err := query.Find(&jsFiles).Error
	return jsFiles, err
}

// MarkJSFileAnalyzed 标记JS文件为已分析
func (d *Database) MarkJSFileAnalyzed(id uint) error {
	return d.db.Model(&JSFile{}).Where("id = ?", id).Update("analyzed", true).Error
}

// SaveFormInfo 保存表单信息
func (d *Database) SaveFormInfo(form *FormInfo) error {
	return d.db.Create(form).Error
}

// GetStatistics 获取统计信息
func (d *Database) GetStatistics(sessionID string) (*ScanStatistics, error) {
	stats := &ScanStatistics{SessionID: sessionID}

	// 获取会话信息
	session, err := d.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	stats.StartTime = session.StartTime
	
	if session.EndTime != nil {
		duration := session.EndTime.Sub(session.StartTime)
		stats.Duration = duration.String()
	}

	// 统计页面数
	var totalPages int64
	d.db.Model(&CrawledPage{}).Where("session_id = ?", sessionID).Count(&totalPages)
	stats.TotalPages = int(totalPages)

	// 统计API数量
	var apiCount int64
	d.db.Model(&APIEndpoint{}).
		Joins("JOIN crawled_pages ON crawled_pages.url LIKE CONCAT('%', api_endpoints.domain, '%')").
		Where("crawled_pages.session_id = ?", sessionID).
		Count(&apiCount)
	stats.TotalAPIs = int(apiCount)

	// 按类型统计API
	var restCount, graphqlCount, wsCount int64
	d.db.Model(&APIEndpoint{}).Where("type = ?", "REST").
		Joins("JOIN crawled_pages ON crawled_pages.url LIKE CONCAT('%', api_endpoints.domain, '%')").
		Where("crawled_pages.session_id = ?", sessionID).Count(&restCount)
	stats.RESTAPIs = int(restCount)

	d.db.Model(&APIEndpoint{}).Where("type = ?", "GraphQL").
		Joins("JOIN crawled_pages ON crawled_pages.url LIKE CONCAT('%', api_endpoints.domain, '%')").
		Where("crawled_pages.session_id = ?", sessionID).Count(&graphqlCount)
	stats.GraphQLAPIs = int(graphqlCount)

	d.db.Model(&APIEndpoint{}).Where("type = ?", "WebSocket").
		Joins("JOIN crawled_pages ON crawled_pages.url LIKE CONCAT('%', api_endpoints.domain, '%')").
		Where("crawled_pages.session_id = ?", sessionID).Count(&wsCount)
	stats.WebSocketAPIs = int(wsCount)

	// 统计JS文件数
	var jsCount int64
	d.db.Model(&JSFile{}).Where("session_id = ?", sessionID).Count(&jsCount)
	stats.JSFiles = int(jsCount)

	// 统计表单数
	var formCount int64
	d.db.Model(&FormInfo{}).Where("session_id = ?", sessionID).Count(&formCount)
	stats.Forms = int(formCount)

	// 统计域名
	var domains []string
	d.db.Model(&CrawledPage{}).Where("session_id = ?", sessionID).
		Distinct("SUBSTRING_INDEX(SUBSTRING_INDEX(url, '/', 3), '://', -1)").
		Pluck("SUBSTRING_INDEX(SUBSTRING_INDEX(url, '/', 3), '://', -1)", &domains)
	stats.Domains = domains

	return stats, nil
}

// GetAPIsByDomain 按域名获取API
func (d *Database) GetAPIsByDomain(domain string, limit, offset int) ([]APIEndpoint, error) {
	var apis []APIEndpoint
	err := d.db.Where("domain = ?", domain).Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&apis).Error
	return apis, err
}

// GetAPIsByType 按类型获取API
func (d *Database) GetAPIsByType(apiType string, limit, offset int) ([]APIEndpoint, error) {
	var apis []APIEndpoint
	err := d.db.Where("type = ?", apiType).Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&apis).Error
	return apis, err
}

// SearchAPIs 搜索API
func (d *Database) SearchAPIs(keyword string, limit, offset int) ([]APIEndpoint, error) {
	var apis []APIEndpoint
	err := d.db.Where("url LIKE ? OR path LIKE ?", "%"+keyword+"%", "%"+keyword+"%").
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&apis).Error
	return apis, err
}

// GetDomainStats 获取域名统计
func (d *Database) GetDomainStats() ([]DomainInfo, error) {
	var domains []DomainInfo
	err := d.db.Model(&APIEndpoint{}).
		Select("domain, COUNT(*) as api_count").
		Group("domain").
		Order("api_count DESC").
		Find(&domains).Error
	return domains, err
}

// DeleteSession 删除会话及相关数据
func (d *Database) DeleteSession(sessionID string) error {
	tx := d.db.Begin()
	
	// 删除相关数据
	if err := tx.Where("session_id = ?", sessionID).Delete(&CrawledPage{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	if err := tx.Where("session_id = ?", sessionID).Delete(&JSFile{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	if err := tx.Where("session_id = ?", sessionID).Delete(&FormInfo{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	if err := tx.Where("session_id = ?", sessionID).Delete(&CrawlSession{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	return tx.Commit().Error
}