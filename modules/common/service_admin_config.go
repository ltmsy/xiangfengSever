package common

import (
	"errors"
	"strings"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"github.com/gocraft/dbr/v2"
	"go.uber.org/zap"
)

// AdminConfigService 配置管理服务
type AdminConfigService struct {
	adminConfigDB *adminConfigDB
	log.Log
}

// NewAdminConfigService 创建配置管理服务实例
func NewAdminConfigService(session *dbr.Session) *AdminConfigService {
	return &AdminConfigService{
		adminConfigDB: newAdminConfigDB(session),
		Log:           log.NewTLog("AdminConfigService"),
	}
}

// ConfigCategoryResponse 配置分类响应
type ConfigCategoryResponse struct {
	Category     string `json:"category"`
	CategoryName string `json:"category_name"`
	Count        int    `json:"count"`
}

// ConfigListResponse 配置列表响应
type ConfigListResponse struct {
	Category     string               `json:"category"`
	CategoryName string               `json:"category_name"`
	Configs      []*AdminSystemConfig `json:"configs"`
}

// ConfigUpdateRequest 配置更新请求
type ConfigUpdateRequest struct {
	ConfigKey   string `json:"config_key" binding:"required"`
	ConfigValue string `json:"config_value" binding:"required"`
}

// ConfigBatchUpdateRequest 批量配置更新请求
type ConfigBatchUpdateRequest struct {
	Configs []*ConfigUpdateRequest `json:"configs" binding:"required"`
}

// ConfigSearchRequest 配置搜索请求
type ConfigSearchRequest struct {
	Keyword  string `json:"keyword"`
	Category string `json:"category"`
	Page     int    `json:"page" binding:"required,min=1"`
	PageSize int    `json:"page_size" binding:"required,min=1,max=100"`
}

// ConfigSearchResponse 配置搜索响应
type ConfigSearchResponse struct {
	Configs  []*AdminSystemConfig `json:"configs"`
	Total    int                  `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// GetConfigCategories 获取配置分类列表
func (s *AdminConfigService) GetConfigCategories() ([]*ConfigCategoryResponse, error) {
	categories, err := s.adminConfigDB.GetConfigsByCategory()
	if err != nil {
		s.Error("获取配置分类失败", zap.Error(err))
		return nil, err
	}

	var response []*ConfigCategoryResponse
	for _, category := range categories {
		response = append(response, &ConfigCategoryResponse{
			Category:     category.Category,
			CategoryName: s.getCategoryDisplayName(category.Category),
			Count:        category.Count,
		})
	}

	return response, nil
}

// GetConfigsByCategory 根据分类获取配置列表
func (s *AdminConfigService) GetConfigsByCategory(category string) (*ConfigListResponse, error) {
	if strings.TrimSpace(category) == "" {
		return nil, errors.New("分类不能为空")
	}

	configs, err := s.adminConfigDB.GetConfigsByCategoryKey(category)
	if err != nil {
		s.Error("根据分类获取配置失败", zap.Error(err))
		return nil, err
	}

	return &ConfigListResponse{
		Category:     category,
		CategoryName: s.getCategoryDisplayName(category),
		Configs:      configs,
	}, nil
}

// GetAllConfigsGrouped 获取所有配置，按分类分组
func (s *AdminConfigService) GetAllConfigsGrouped() ([]*ConfigListResponse, error) {
	categories, err := s.GetConfigCategories()
	if err != nil {
		return nil, err
	}

	var result []*ConfigListResponse
	for _, category := range categories {
		configs, err := s.adminConfigDB.GetConfigsByCategoryKey(category.Category)
		if err != nil {
			s.Error("获取分类配置失败", zap.Error(err))
			continue
		}

		result = append(result, &ConfigListResponse{
			Category:     category.Category,
			CategoryName: category.CategoryName,
			Configs:      configs,
		})
	}

	return result, nil
}

// UpdateConfig 更新单个配置
func (s *AdminConfigService) UpdateConfig(req *ConfigUpdateRequest, updatedBy string) error {
	if strings.TrimSpace(req.ConfigKey) == "" {
		return errors.New("配置键不能为空")
	}

	// 检查配置是否存在
	config, err := s.adminConfigDB.GetConfigByKey(req.ConfigKey)
	if err != nil {
		s.Error("查询配置失败", zap.Error(err))
		return err
	}

	if config == nil {
		return errors.New("配置不存在")
	}

	// 检查配置是否可编辑
	if config.IsEditable == 0 {
		return errors.New("该配置不可编辑")
	}

	// 验证配置值
	if err := s.validateConfigValue(config.ConfigType, req.ConfigValue); err != nil {
		return err
	}

	// 更新配置
	err = s.adminConfigDB.UpdateConfig(req.ConfigKey, req.ConfigValue, updatedBy)
	if err != nil {
		s.Error("更新配置失败", zap.Error(err))
		return err
	}

	s.Info("配置更新成功", zap.String("config_key", req.ConfigKey), zap.String("updated_by", updatedBy))
	return nil
}

// BatchUpdateConfigs 批量更新配置
func (s *AdminConfigService) BatchUpdateConfigs(req *ConfigBatchUpdateRequest, updatedBy string) error {
	if len(req.Configs) == 0 {
		return errors.New("配置列表不能为空")
	}

	// 验证所有配置
	configMap := make(map[string]string)
	for _, config := range req.Configs {
		if strings.TrimSpace(config.ConfigKey) == "" {
			return errors.New("配置键不能为空")
		}

		// 检查配置是否存在且可编辑
		existingConfig, err := s.adminConfigDB.GetConfigByKey(config.ConfigKey)
		if err != nil {
			s.Error("查询配置失败", zap.Error(err))
			return err
		}

		if existingConfig == nil {
			return errors.New("配置不存在: " + config.ConfigKey)
		}

		if existingConfig.IsEditable == 0 {
			return errors.New("配置不可编辑: " + config.ConfigKey)
		}

		// 验证配置值
		if err := s.validateConfigValue(existingConfig.ConfigType, config.ConfigValue); err != nil {
			return errors.New("配置值验证失败: " + config.ConfigKey + " - " + err.Error())
		}

		configMap[config.ConfigKey] = config.ConfigValue
	}

	// 批量更新
	err := s.adminConfigDB.BatchUpdateConfigs(configMap, updatedBy)
	if err != nil {
		s.Error("批量更新配置失败", zap.Error(err))
		return err
	}

	s.Info("批量配置更新成功", zap.Int("count", len(configMap)), zap.String("updated_by", updatedBy))
	return nil
}

// SearchConfigs 搜索配置
func (s *AdminConfigService) SearchConfigs(req *ConfigSearchRequest) (*ConfigSearchResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	configs, total, err := s.adminConfigDB.SearchConfigs(req.Keyword, req.Category, req.Page, req.PageSize)
	if err != nil {
		s.Error("搜索配置失败", zap.Error(err))
		return nil, err
	}

	return &ConfigSearchResponse{
		Configs:  configs,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetPublicConfigs 获取公开配置
func (s *AdminConfigService) GetPublicConfigs() (map[string]string, error) {
	return s.adminConfigDB.GetPublicConfigs()
}

// AddConfig 添加配置
func (s *AdminConfigService) AddConfig(config *AdminSystemConfig) error {
	if strings.TrimSpace(config.ConfigKey) == "" {
		return errors.New("配置键不能为空")
	}

	// 检查配置键是否已存在
	existingConfig, err := s.adminConfigDB.GetConfigByKey(config.ConfigKey)
	if err != nil {
		s.Error("查询配置失败", zap.Error(err))
		return err
	}

	if existingConfig != nil {
		return errors.New("配置键已存在")
	}

	// 验证配置值
	if err := s.validateConfigValue(config.ConfigType, config.ConfigValue); err != nil {
		return err
	}

	// 设置默认值
	if config.Version == 0 {
		config.Version = 1
	}
	if config.SortOrder == 0 {
		config.SortOrder = 999
	}

	err = s.adminConfigDB.AddConfig(config)
	if err != nil {
		s.Error("添加配置失败", zap.Error(err))
		return err
	}

	s.Info("配置添加成功", zap.String("config_key", config.ConfigKey))
	return nil
}

// DeleteConfig 删除配置
func (s *AdminConfigService) DeleteConfig(configKey string) error {
	if strings.TrimSpace(configKey) == "" {
		return errors.New("配置键不能为空")
	}

	// 检查配置是否存在
	config, err := s.adminConfigDB.GetConfigByKey(configKey)
	if err != nil {
		s.Error("查询配置失败", zap.Error(err))
		return err
	}

	if config == nil {
		return errors.New("配置不存在")
	}

	err = s.adminConfigDB.DeleteConfig(configKey)
	if err != nil {
		s.Error("删除配置失败", zap.Error(err))
		return err
	}

	s.Info("配置删除成功", zap.String("config_key", configKey))
	return nil
}

// validateConfigValue 验证配置值
func (s *AdminConfigService) validateConfigValue(configType, value string) error {
	switch configType {
	case "boolean":
		if value != "0" && value != "1" && value != "true" && value != "false" {
			return errors.New("布尔值配置只能是 0/1 或 true/false")
		}
	case "number":
		// 这里可以添加数字范围验证
		if strings.TrimSpace(value) == "" {
			return errors.New("数字配置值不能为空")
		}
	case "string":
		if strings.TrimSpace(value) == "" {
			return errors.New("字符串配置值不能为空")
		}
	case "json":
		// 这里可以添加JSON格式验证
		if strings.TrimSpace(value) == "" {
			return errors.New("JSON配置值不能为空")
		}
	default:
		// 未知类型，跳过验证
	}

	return nil
}

// getCategoryDisplayName 获取分类显示名称
func (s *AdminConfigService) getCategoryDisplayName(category string) string {
	categoryNames := map[string]string{
		"group":        "群组配置",
		"system":       "系统配置",
		"security":     "安全配置",
		"notification": "通知配置",
		"ui":           "UI配置",
		"message":      "消息配置",
		"file":         "文件配置",
		"robot":        "机器人配置",
		"statistics":   "统计配置",
	}

	if name, exists := categoryNames[category]; exists {
		return name
	}

	return category
}
