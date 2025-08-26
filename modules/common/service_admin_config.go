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

// ConfigCreateRequest 配置创建请求
type ConfigCreateRequest struct {
	ConfigKey   string `json:"config_key" binding:"required"`
	ConfigValue string `json:"config_value" binding:"required"`
	ConfigType  string `json:"config_type" binding:"required"`
	Description string `json:"description" binding:"required"`
	Category    string `json:"category" binding:"required"`
	IsEditable  int    `json:"is_editable"`
	IsPublic    int    `json:"is_public"`
	SortOrder   int    `json:"sort_order"`
}

// ConfigBatchDeleteRequest 批量删除配置请求
type ConfigBatchDeleteRequest struct {
	ConfigKeys []string `json:"config_keys" binding:"required"`
}

// ConfigBatchDeleteResponse 批量删除配置响应
type ConfigBatchDeleteResponse struct {
	DeletedCount int      `json:"deleted_count"`
	FailedKeys   []string `json:"failed_keys"`
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

// GetConfigByKey 根据配置键获取配置
func (s *AdminConfigService) GetConfigByKey(configKey string) (*AdminSystemConfig, error) {
	return s.adminConfigDB.GetConfigByKey(configKey)
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
	case "image":
		// 图片类型验证：检查是否为有效的图片URL或路径
		if strings.TrimSpace(value) == "" {
			return errors.New("图片配置值不能为空")
		}
		// 可以添加图片格式验证（.jpg, .png, .gif, .webp等）
		if !s.isValidImagePath(value) {
			return errors.New("图片配置值格式不正确，应为有效的图片路径或URL")
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

// isValidImagePath 验证图片路径是否有效
func (s *AdminConfigService) isValidImagePath(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}

	// 支持的图片格式
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg"}
	pathLower := strings.ToLower(path)

	// 检查是否为HTTP/HTTPS URL
	if strings.HasPrefix(pathLower, "http://") || strings.HasPrefix(pathLower, "https://") {
		// URL格式，检查是否包含图片扩展名
		for _, ext := range validExtensions {
			if strings.HasSuffix(pathLower, ext) {
				return true
			}
		}
		return false
	}

	// 本地路径格式，检查是否包含图片扩展名
	for _, ext := range validExtensions {
		if strings.HasSuffix(pathLower, ext) {
			return true
		}
	}

	return false
}

// CreateConfig 创建新配置项
func (s *AdminConfigService) CreateConfig(req *ConfigCreateRequest, createdBy string) (*AdminSystemConfig, error) {
	// 参数验证
	if strings.TrimSpace(req.ConfigKey) == "" {
		return nil, errors.New("配置键不能为空")
	}
	if strings.TrimSpace(req.ConfigValue) == "" {
		return nil, errors.New("配置值不能为空")
	}
	if strings.TrimSpace(req.ConfigType) == "" {
		return nil, errors.New("配置类型不能为空")
	}
	if strings.TrimSpace(req.Description) == "" {
		return nil, errors.New("配置描述不能为空")
	}
	if strings.TrimSpace(req.Category) == "" {
		return nil, errors.New("配置分类不能为空")
	}

	// 检查配置键是否已存在
	existingConfig, err := s.adminConfigDB.GetConfigByKey(req.ConfigKey)
	if err != nil {
		s.Error("检查配置键是否存在失败", zap.Error(err))
		return nil, errors.New("检查配置键是否存在失败")
	}
	if existingConfig != nil {
		return nil, errors.New("配置键已存在")
	}

	// 验证配置值
	err = s.validateConfigValue(req.ConfigType, req.ConfigValue)
	if err != nil {
		return nil, err
	}

	// 设置默认值
	if req.IsEditable == 0 {
		req.IsEditable = 1
	}
	if req.IsPublic == 0 {
		req.IsPublic = 0
	}
	if req.SortOrder == 0 {
		req.SortOrder = 0
	}

	// 创建配置项
	config := &AdminSystemConfig{
		ConfigKey:   req.ConfigKey,
		ConfigValue: req.ConfigValue,
		ConfigType:  req.ConfigType,
		Description: req.Description,
		Category:    req.Category,
		IsEditable:  req.IsEditable,
		IsPublic:    req.IsPublic,
		SortOrder:   req.SortOrder,
		Version:     1,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}

	err = s.adminConfigDB.CreateConfig(config)
	if err != nil {
		s.Error("创建配置失败", zap.Error(err))
		return nil, errors.New("创建配置失败")
	}

	s.Info("配置创建成功", zap.String("config_key", req.ConfigKey))
	return config, nil
}

// DeleteConfigWithAuth 删除配置项（带权限检查）
func (s *AdminConfigService) DeleteConfigWithAuth(configKey string, deletedBy string) error {
	// 检查配置是否存在
	config, err := s.adminConfigDB.GetConfigByKey(configKey)
	if err != nil {
		s.Error("查询配置失败", zap.Error(err))
		return errors.New("查询配置失败")
	}
	if config == nil {
		return errors.New("配置不存在")
	}

	// 检查是否可编辑
	if config.IsEditable == 0 {
		return errors.New("该配置项不可删除")
	}

	// 检查是否为系统核心配置（不允许删除）
	coreConfigs := []string{
		"system.site_name", "system.site_logo", "system.site_favicon",
		"system.default_avatar", "system.default_group_avatar",
		"security.password_min_length", "security.allow_register",
	}
	for _, coreKey := range coreConfigs {
		if configKey == coreKey {
			return errors.New("系统核心配置不允许删除")
		}
	}

	// 执行删除
	err = s.adminConfigDB.DeleteConfig(configKey)
	if err != nil {
		s.Error("删除配置失败", zap.Error(err))
		return errors.New("删除配置失败")
	}

	s.Info("配置删除成功", zap.String("config_key", configKey), zap.String("deleted_by", deletedBy))
	return nil
}

// BatchDeleteConfigs 批量删除配置项
func (s *AdminConfigService) BatchDeleteConfigs(req *ConfigBatchDeleteRequest, deletedBy string) (*ConfigBatchDeleteResponse, error) {
	if len(req.ConfigKeys) == 0 {
		return nil, errors.New("配置键列表不能为空")
	}

	var deletedCount int
	var failedKeys []string

	// 使用事务进行批量删除
	tx, err := s.adminConfigDB.BeginTx()
	if err != nil {
		s.Error("开启事务失败", zap.Error(err))
		return nil, errors.New("开启事务失败")
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for _, configKey := range req.ConfigKeys {
		err = s.DeleteConfigWithAuth(configKey, deletedBy)
		if err != nil {
			failedKeys = append(failedKeys, configKey)
			s.Warn("删除配置失败", zap.String("config_key", configKey), zap.Error(err))
			continue
		}
		deletedCount++
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		s.Error("提交事务失败", zap.Error(err))
		return nil, errors.New("提交事务失败")
	}

	response := &ConfigBatchDeleteResponse{
		DeletedCount: deletedCount,
		FailedKeys:   failedKeys,
	}

	s.Info("批量删除配置完成", zap.Int("deleted_count", deletedCount), zap.Int("failed_count", len(failedKeys)))
	return response, nil
}
