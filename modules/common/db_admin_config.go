package common

import (
	"errors"
	"strings"
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"github.com/gocraft/dbr/v2"
	"go.uber.org/zap"
)

// AdminSystemConfig 后台管理系统配置表
type AdminSystemConfig struct {
	ID          int64  `db:"id" json:"id"`
	ConfigKey   string `db:"config_key" json:"config_key"`
	ConfigValue string `db:"config_value" json:"config_value"`
	ConfigType  string `db:"config_type" json:"config_type"`
	Description string `db:"description" json:"description"`
	Category    string `db:"category" json:"category"`
	IsEditable  int    `db:"is_editable" json:"is_editable"`
	IsPublic    int    `db:"is_public" json:"is_public"`
	SortOrder   int    `db:"sort_order" json:"sort_order"`
	Version     int64  `db:"version" json:"version"`
	CreatedAt   string `db:"created_at" json:"created_at"`
	UpdatedAt   string `db:"updated_at" json:"updated_at"`
	CreatedBy   string `db:"created_by" json:"created_by"`
	UpdatedBy   string `db:"updated_by" json:"updated_by"`
}

// ConfigCategory 配置分类信息
type ConfigCategory struct {
	Category     string `db:"category" json:"category"`
	CategoryName string `db:"category_name" json:"category_name"`
	Count        int    `db:"count" json:"count"`
}

// adminConfigDB 配置管理数据库操作
type adminConfigDB struct {
	session *dbr.Session
	log.Log
}

// newAdminConfigDB 创建配置管理数据库操作实例
func newAdminConfigDB(session *dbr.Session) *adminConfigDB {
	return &adminConfigDB{
		session: session,
		Log:     log.NewTLog("adminConfigDB"),
	}
}

// GetConfigsByCategory 按分类获取配置列表
func (a *adminConfigDB) GetConfigsByCategory() ([]*ConfigCategory, error) {
	var categories []*ConfigCategory
	_, err := a.session.Select("category, COUNT(*) as count").From("admin_system_config").GroupBy("category").OrderBy("MIN(sort_order)").Load(&categories)
	if err != nil {
		a.Error("获取配置分类失败", zap.Error(err))
		return nil, err
	}

	return categories, nil
}

// GetConfigsByCategoryKey 根据分类键获取配置列表
func (a *adminConfigDB) GetConfigsByCategoryKey(category string) ([]*AdminSystemConfig, error) {
	var configs []*AdminSystemConfig
	_, err := a.session.Select("*").From("admin_system_config").Where("category = ?", category).OrderBy("sort_order ASC, id ASC").Load(&configs)
	if err != nil {
		a.Error("根据分类获取配置失败", zap.Error(err))
		return nil, err
	}

	return configs, nil
}

// GetConfigByKey 根据配置键获取配置
func (a *adminConfigDB) GetConfigByKey(configKey string) (*AdminSystemConfig, error) {
	var config AdminSystemConfig
	_, err := a.session.Select("*").From("admin_system_config").Where("config_key = ?", configKey).Load(&config)
	if err != nil {
		a.Error("根据键获取配置失败", zap.Error(err))
		return nil, err
	}

	if config.ID == 0 {
		return nil, nil
	}

	return &config, nil
}

// UpdateConfig 更新配置
func (a *adminConfigDB) UpdateConfig(configKey, configValue, updatedBy string) error {
	// 先获取当前版本号
	currentConfig, err := a.GetConfigByKey(configKey)
	if err != nil {
		return err
	}
	if currentConfig == nil {
		return errors.New("配置不存在")
	}

	result, err := a.session.Update("admin_system_config").SetMap(map[string]interface{}{
		"config_value": configValue,
		"updated_by":   updatedBy,
		"updated_at":   time.Now().Format("2006-01-02 15:04:05"),
		"version":      currentConfig.Version + 1,
	}).Where("config_key = ?", configKey).Exec()

	if err != nil {
		a.Error("更新配置失败", zap.Error(err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("配置不存在或已被删除")
	}

	return nil
}

// BatchUpdateConfigs 批量更新配置
func (a *adminConfigDB) BatchUpdateConfigs(configs map[string]string, updatedBy string) error {
	if len(configs) == 0 {
		return nil
	}

	// 开始事务
	tx, err := a.session.Begin()
	if err != nil {
		a.Error("开始事务失败", zap.Error(err))
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 批量更新
	for configKey, configValue := range configs {
		// 先获取当前版本号
		currentConfig, err := a.GetConfigByKey(configKey)
		if err != nil {
			a.Error("获取配置版本失败", zap.Error(err))
			return err
		}
		if currentConfig == nil {
			a.Error("配置不存在", zap.String("config_key", configKey))
			return errors.New("配置不存在: " + configKey)
		}

		_, err = tx.Update("admin_system_config").SetMap(map[string]interface{}{
			"config_value": configValue,
			"updated_by":   updatedBy,
			"updated_at":   time.Now().Format("2006-01-02 15:04:05"),
			"version":      currentConfig.Version + 1,
		}).Where("config_key = ?", configKey).Exec()

		if err != nil {
			a.Error("批量更新配置失败", zap.Error(err))
			return err
		}
	}

	// 提交事务
	return tx.Commit()
}

// GetPublicConfigs 获取公开配置
func (a *adminConfigDB) GetPublicConfigs() (map[string]string, error) {
	var configs []*AdminSystemConfig
	_, err := a.session.Select("config_key, config_value").From("admin_system_config").Where("is_public = 1").Load(&configs)
	if err != nil {
		a.Error("获取公开配置失败", zap.Error(err))
		return nil, err
	}

	result := make(map[string]string)
	for _, config := range configs {
		result[config.ConfigKey] = config.ConfigValue
	}

	return result, nil
}

// SearchConfigs 搜索配置
func (a *adminConfigDB) SearchConfigs(keyword string, category string, page, pageSize int) ([]*AdminSystemConfig, int, error) {
	var whereClause []string
	var args []interface{}

	if strings.TrimSpace(keyword) != "" {
		whereClause = append(whereClause, "(config_key LIKE ? OR description LIKE ?)")
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	if strings.TrimSpace(category) != "" {
		whereClause = append(whereClause, "category = ?")
		args = append(args, category)
	}

	// whereClause = append(whereClause, "is_deleted = 0")

	whereSQL := strings.Join(whereClause, " AND ")

	// 获取总数
	var total int
	countQuery := a.session.Select("COUNT(*)").From("admin_system_config").Where(whereSQL, args...)
	_, err := countQuery.Load(&total)
	if err != nil {
		a.Error("获取配置总数失败", zap.Error(err))
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	var configs []*AdminSystemConfig
	query := a.session.Select("*").From("admin_system_config").Where(whereSQL, args...).OrderBy("sort_order ASC, id ASC").Limit(uint64(pageSize)).Offset(uint64(offset))
	_, err = query.Load(&configs)
	if err != nil {
		a.Error("搜索配置失败", zap.Error(err))
		return nil, 0, err
	}

	return configs, total, nil
}

// AddConfig 添加配置
func (a *adminConfigDB) AddConfig(config *AdminSystemConfig) error {
	_, err := a.session.InsertInto("admin_system_config").Columns("config_key", "config_value", "config_type", "description", "category", "is_editable", "is_public", "sort_order", "version", "created_by", "updated_by").Record(config).Exec()

	if err != nil {
		a.Error("添加配置失败", zap.Error(err))
		return err
	}

	return nil
}

// DeleteConfig 删除配置（硬删除）
func (a *adminConfigDB) DeleteConfig(configKey string) error {
	_, err := a.session.DeleteFrom("admin_system_config").Where("config_key = ?", configKey).Exec()

	if err != nil {
		a.Error("删除配置失败", zap.Error(err))
		return err
	}

	return nil
}

// CreateConfig 创建新配置项
func (a *adminConfigDB) CreateConfig(config *AdminSystemConfig) error {
	_, err := a.session.InsertInto("admin_system_config").Columns(
		"config_key", "config_value", "config_type", "description",
		"category", "is_editable", "is_public", "sort_order",
		"version", "created_by", "updated_by",
	).Record(config).Exec()

	if err != nil {
		a.Error("创建配置失败", zap.Error(err))
		return err
	}

	return nil
}

// BeginTx 开启事务
func (a *adminConfigDB) BeginTx() (*dbr.Tx, error) {
	return a.session.Begin()
}
