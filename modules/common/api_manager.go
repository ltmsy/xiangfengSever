package common

import (
	"errors"
	"strings"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/wkhttp"
	"go.uber.org/zap"
)

// Manager 通用后台管理api
type Manager struct {
	ctx *config.Context
	log.Log
	db                 *db
	appconfigDB        *appConfigDB
	adminConfigService *AdminConfigService
}

// NewManager NewManager
func NewManager(ctx *config.Context) *Manager {
	return &Manager{
		ctx:                ctx,
		Log:                log.NewTLog("commonManager"),
		db:                 newDB(ctx.DB()),
		appconfigDB:        newAppConfigDB(ctx),
		adminConfigService: NewAdminConfigService(ctx.DB()),
	}
}

// Route 配置路由规则
func (m *Manager) Route(r *wkhttp.WKHttp) {
	auth := r.Group("/v1/manager", m.ctx.AuthMiddleware(r))
	{
		auth.GET("/common/appconfig", m.appconfig)               // 获取app配置
		auth.POST("/common/appconfig", m.updateConfig)           // 修改app配置
		auth.GET("/common/appmodule", m.getAppModule)            // 获取app模块
		auth.PUT("/common/appmodule", m.updateAppModule)         // 修改app模块
		auth.POST("/common/appmodule", m.addAppModule)           // 新增app模块
		auth.DELETE("/common/:sid/appmodule", m.deleteAppModule) // 删除app模块

		// 配置管理接口
		auth.GET("/config/categories", m.getConfigCategories)          // 获取配置分类列表
		auth.GET("/config/list", m.getAllConfigsGrouped)               // 获取所有配置（按分类分组）
		auth.GET("/config/category/:category", m.getConfigsByCategory) // 根据分类获取配置
		auth.PUT("/config/update", m.updateSystemConfig)               // 更新单个配置
		auth.PUT("/config/batch-update", m.batchUpdateConfigs)         // 批量更新配置
		auth.POST("/config/search", m.searchConfigs)                   // 搜索配置
		auth.POST("/config/create", m.createConfig)                    // 新增配置项
		auth.DELETE("/config/delete/:config_key", m.deleteConfig)      // 删除配置项
		auth.DELETE("/config/batch-delete", m.batchDeleteConfigs)      // 批量删除配置项
		auth.GET("/config/public", m.getPublicConfigs)                 // 获取公开配置
	}
}
func (m *Manager) deleteAppModule(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	sid := c.Param("sid")
	if strings.TrimSpace(sid) == "" {
		c.ResponseError(errors.New("sid不能为空！"))
		return
	}
	module, err := m.db.queryAppModuleWithSid(sid)
	if err != nil {
		m.Error("查询app模块错误", zap.Error(err))
		c.ResponseError(errors.New("查询app模块错误"))
		return
	}
	if module == nil {
		c.ResponseError(errors.New("删除的模块不存在"))
		return
	}
	err = m.db.deleteAppModule(sid)
	if err != nil {
		m.Error("删除app模块错误", zap.Error(err))
		c.ResponseError(errors.New("删除app模块错误"))
		return
	}
	c.ResponseOK()
}

// 新增app模块
func (m *Manager) addAppModule(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}
	type ReqVO struct {
		SID    string `json:"sid"`
		Name   string `json:"name"`
		Desc   string `json:"desc"`
		Status int    `json:"status"`
	}
	var req ReqVO
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}

	if strings.TrimSpace(req.SID) == "" || strings.TrimSpace(req.Desc) == "" || strings.TrimSpace(req.Name) == "" {
		c.ResponseError(errors.New("名称/ID/介绍不能为空！"))
		return
	}
	module, err := m.db.queryAppModuleWithSid(req.SID)
	if err != nil {
		m.Error("查询app模块错误", zap.Error(err))
		c.ResponseError(errors.New("查询app模块错误"))
		return
	}
	if module != nil && module.SID != "" {
		c.ResponseError(errors.New("该sid模块已存在"))
		return
	}
	_, err = m.db.insertAppModule(&appModuleModel{
		SID:    req.SID,
		Name:   req.Name,
		Desc:   req.Desc,
		Status: req.Status,
	})
	if err != nil {
		m.Error("新增app模块错误", zap.Error(err))
		c.ResponseError(errors.New("新增app模块错误"))
		return
	}
	c.ResponseOK()
}
func (m *Manager) updateAppModule(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}
	type ReqVO struct {
		SID    string `json:"sid"`
		Name   string `json:"name"`
		Desc   string `json:"desc"`
		Status int    `json:"status"`
	}
	var req ReqVO
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}

	if strings.TrimSpace(req.SID) == "" || strings.TrimSpace(req.Desc) == "" || strings.TrimSpace(req.Name) == "" {
		c.ResponseError(errors.New("名称/ID/介绍不能为空！"))
		return
	}
	module, err := m.db.queryAppModuleWithSid(req.SID)
	if err != nil {
		m.Error("查询app模块错误", zap.Error(err))
		c.ResponseError(errors.New("查询app模块错误"))
		return
	}
	if module == nil {
		c.ResponseError(errors.New("不存在该模块"))
		return
	}
	module.Name = req.Name
	module.Desc = req.Desc
	module.Status = req.Status
	err = m.db.updateAppModule(module)
	if err != nil {
		m.Error("修改app模块错误", zap.Error(err))
		c.ResponseError(errors.New("修改app模块错误"))
		return
	}
	c.ResponseOK()
}

// 获取app模块
func (m *Manager) getAppModule(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}
	modules, err := m.db.queryAppModule()
	if err != nil {
		m.Error("查询app模块错误", zap.Error(err))
		c.ResponseError(errors.New("查询app模块错误"))
		return
	}
	list := make([]*managerAppModule, 0)
	if len(modules) > 0 {
		for _, module := range modules {
			list = append(list, &managerAppModule{
				SID:    module.SID,
				Name:   module.Name,
				Desc:   module.Desc,
				Status: module.Status,
			})
		}
	}
	c.Response(list)
}
func (m *Manager) updateConfig(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}
	type reqVO struct {
		RevokeSecond                   int    `json:"revoke_second"`
		WelcomeMessage                 string `json:"welcome_message"`
		NewUserJoinSystemGroup         int    `json:"new_user_join_system_group"`
		SearchByPhone                  int    `json:"search_by_phone"`
		RegisterInviteOn               int    `json:"register_invite_on"`                  // 开启注册邀请机制
		SendWelcomeMessageOn           int    `json:"send_welcome_message_on"`             // 开启注册登录发送欢迎语
		InviteSystemAccountJoinGroupOn int    `json:"invite_system_account_join_group_on"` // 开启系统账号加入群聊
		RegisterUserMustCompleteInfoOn int    `json:"register_user_must_complete_info_on"` // 注册用户必须填写完整信息
		ChannelPinnedMessageMaxCount   int    `json:"channel_pinned_message_max_count"`    // 频道置顶消息最大数量
		CanModifyApiUrl                int    `json:"can_modify_api_url"`                  // 是否可以修改api地址
	}
	var req reqVO
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误！"))
		return
	}
	appConfigM, err := m.appconfigDB.query()
	if err != nil {
		m.Error("查询应用配置失败！", zap.Error(err))
		c.ResponseError(errors.New("查询应用配置失败！"))
		return
	}
	configMap := map[string]interface{}{}
	configMap["revoke_second"] = req.RevokeSecond
	configMap["welcome_message"] = req.WelcomeMessage
	configMap["new_user_join_system_group"] = req.NewUserJoinSystemGroup
	configMap["search_by_phone"] = req.SearchByPhone
	configMap["register_invite_on"] = req.RegisterInviteOn
	configMap["send_welcome_message_on"] = req.SendWelcomeMessageOn
	configMap["invite_system_account_join_group_on"] = req.InviteSystemAccountJoinGroupOn
	configMap["register_user_must_complete_info_on"] = req.RegisterUserMustCompleteInfoOn
	configMap["channel_pinned_message_max_count"] = req.ChannelPinnedMessageMaxCount
	configMap["can_modify_api_url"] = req.CanModifyApiUrl
	err = m.appconfigDB.updateWithMap(configMap, appConfigM.Id)
	if err != nil {
		m.Error("修改app配置信息错误", zap.Error(err))
		c.ResponseError(errors.New("修改app配置信息错误"))
		return
	}
	c.ResponseOK()
}
func (m *Manager) appconfig(c *wkhttp.Context) {
	err := c.CheckLoginRole()
	if err != nil {
		c.ResponseError(err)
		return
	}
	appconfig, err := m.appconfigDB.query()
	if err != nil {
		m.Error("查询应用配置失败！", zap.Error(err))
		c.ResponseError(errors.New("查询应用配置失败！"))
		return
	}
	var revokeSecond = 0
	var newUserJoinSystemGroup = 1
	var welcomeMessage = ""
	var searchByPhone = 1
	var registerInviteOn = 0
	var sendWelcomeMessageOn = 0
	var inviteSystemAccountJoinGroupOn = 0
	var registerUserMustCompleteInfoOn = 0
	var channelPinnedMessageMaxCount = 10
	var canModifyApiUrl = 0
	if appconfig != nil {
		revokeSecond = appconfig.RevokeSecond
		welcomeMessage = appconfig.WelcomeMessage
		newUserJoinSystemGroup = appconfig.NewUserJoinSystemGroup
		searchByPhone = appconfig.SearchByPhone
		registerInviteOn = appconfig.RegisterInviteOn
		sendWelcomeMessageOn = appconfig.SendWelcomeMessageOn
		inviteSystemAccountJoinGroupOn = appconfig.InviteSystemAccountJoinGroupOn
		registerUserMustCompleteInfoOn = appconfig.RegisterUserMustCompleteInfoOn
		channelPinnedMessageMaxCount = appconfig.ChannelPinnedMessageMaxCount
		canModifyApiUrl = appconfig.CanModifyApiUrl
	}
	if revokeSecond == 0 {
		revokeSecond = 120
	}
	if welcomeMessage == "" {
		welcomeMessage = m.ctx.GetConfig().WelcomeMessage
	}
	c.Response(&managerAppConfigResp{
		RevokeSecond:                   revokeSecond,
		WelcomeMessage:                 welcomeMessage,
		NewUserJoinSystemGroup:         newUserJoinSystemGroup,
		SearchByPhone:                  searchByPhone,
		RegisterInviteOn:               registerInviteOn,
		SendWelcomeMessageOn:           sendWelcomeMessageOn,
		InviteSystemAccountJoinGroupOn: inviteSystemAccountJoinGroupOn,
		RegisterUserMustCompleteInfoOn: registerUserMustCompleteInfoOn,
		ChannelPinnedMessageMaxCount:   channelPinnedMessageMaxCount,
		CanModifyApiUrl:                canModifyApiUrl,
	})
}

type managerAppConfigResp struct {
	RevokeSecond                   int    `json:"revoke_second"`
	WelcomeMessage                 string `json:"welcome_message"`
	NewUserJoinSystemGroup         int    `json:"new_user_join_system_group"`
	SearchByPhone                  int    `json:"search_by_phone"`
	RegisterInviteOn               int    `json:"register_invite_on"`                  // 开启注册邀请机制
	SendWelcomeMessageOn           int    `json:"send_welcome_message_on"`             // 开启注册登录发送欢迎语
	InviteSystemAccountJoinGroupOn int    `json:"invite_system_account_join_group_on"` // 开启系统账号加入群聊
	RegisterUserMustCompleteInfoOn int    `json:"register_user_must_complete_info_on"` // 注册用户必须填写完整信息
	ChannelPinnedMessageMaxCount   int    `json:"channel_pinned_message_max_count"`    // 频道置顶消息最大数量
	CanModifyApiUrl                int    `json:"can_modify_api_url"`                  // 是否可以修改api地址
}

type managerAppModule struct {
	SID    string `json:"sid"`
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Status int    `json:"status"` // 模块状态 1.可选 0.不可选 2.选中不可编辑
}

// ==================== 配置管理接口实现 ====================

// getConfigCategories 获取配置分类列表
func (m *Manager) getConfigCategories(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	categories, err := m.adminConfigService.GetConfigCategories()
	if err != nil {
		m.Error("获取配置分类失败", zap.Error(err))
		c.ResponseError(errors.New("获取配置分类失败"))
		return
	}

	c.Response(categories)
}

// getAllConfigsGrouped 获取所有配置（按分类分组）
func (m *Manager) getAllConfigsGrouped(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	configs, err := m.adminConfigService.GetAllConfigsGrouped()
	if err != nil {
		m.Error("获取所有配置失败", zap.Error(err))
		c.ResponseError(errors.New("获取所有配置失败"))
		return
	}

	c.Response(configs)
}

// getConfigsByCategory 根据分类获取配置
func (m *Manager) getConfigsByCategory(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	category := c.Param("category")
	if strings.TrimSpace(category) == "" {
		c.ResponseError(errors.New("分类参数不能为空"))
		return
	}

	configs, err := m.adminConfigService.GetConfigsByCategory(category)
	if err != nil {
		m.Error("根据分类获取配置失败", zap.Error(err))
		c.ResponseError(errors.New("根据分类获取配置失败"))
		return
	}

	c.Response(configs)
}

// updateSystemConfig 更新单个配置
func (m *Manager) updateSystemConfig(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	var req ConfigUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	// 获取当前登录用户
	loginUID := c.GetLoginUID()
	if loginUID == "" {
		c.ResponseError(errors.New("用户未登录"))
		return
	}

	err = m.adminConfigService.UpdateConfig(&req, loginUID)
	if err != nil {
		m.Error("更新配置失败", zap.Error(err))
		c.ResponseError(err)
		return
	}

	c.ResponseOK()
}

// batchUpdateConfigs 批量更新配置
func (m *Manager) batchUpdateConfigs(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	var req ConfigBatchUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	// 获取当前登录用户
	loginUID := c.GetLoginUID()
	if loginUID == "" {
		c.ResponseError(errors.New("用户未登录"))
		return
	}

	err = m.adminConfigService.BatchUpdateConfigs(&req, loginUID)
	if err != nil {
		m.Error("批量更新配置失败", zap.Error(err))
		c.ResponseError(err)
		return
	}

	c.ResponseOK()
}

// searchConfigs 搜索配置
func (m *Manager) searchConfigs(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	var req ConfigSearchRequest
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	result, err := m.adminConfigService.SearchConfigs(&req)
	if err != nil {
		m.Error("搜索配置失败", zap.Error(err))
		c.ResponseError(errors.New("搜索配置失败"))
		return
	}

	c.Response(result)
}

// getPublicConfigs 获取公开配置
func (m *Manager) getPublicConfigs(c *wkhttp.Context) {
	// 公开接口，无需权限验证
	configs, err := m.adminConfigService.GetPublicConfigs()
	if err != nil {
		m.Error("获取公开配置失败", zap.Error(err))
		c.ResponseError(errors.New("获取公开配置失败"))
		return
	}

	c.Response(configs)
}

// createConfig 新增配置项
func (m *Manager) createConfig(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	var req ConfigCreateRequest
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	// 获取当前登录用户
	loginUID := c.GetLoginUID()
	if loginUID == "" {
		c.ResponseError(errors.New("用户未登录"))
		return
	}

	config, err := m.adminConfigService.CreateConfig(&req, loginUID)
	if err != nil {
		m.Error("新增配置失败", zap.Error(err))
		c.ResponseError(err)
		return
	}

	c.Response(config)
}

// deleteConfig 删除配置项
func (m *Manager) deleteConfig(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	configKey := c.Param("config_key")
	if strings.TrimSpace(configKey) == "" {
		c.ResponseError(errors.New("配置键不能为空"))
		return
	}

	// 获取当前登录用户
	loginUID := c.GetLoginUID()
	if loginUID == "" {
		c.ResponseError(errors.New("用户未登录"))
		return
	}

	err = m.adminConfigService.DeleteConfigWithAuth(configKey, loginUID)
	if err != nil {
		m.Error("删除配置失败", zap.Error(err))
		c.ResponseError(err)
		return
	}

	c.ResponseOK()
}

// batchDeleteConfigs 批量删除配置项
func (m *Manager) batchDeleteConfigs(c *wkhttp.Context) {
	err := c.CheckLoginRoleIsSuperAdmin()
	if err != nil {
		c.ResponseError(err)
		return
	}

	var req ConfigBatchDeleteRequest
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(errors.New("请求数据格式有误"))
		return
	}

	// 获取当前登录用户
	loginUID := c.GetLoginUID()
	if loginUID == "" {
		c.ResponseError(errors.New("用户未登录"))
		return
	}

	result, err := m.adminConfigService.BatchDeleteConfigs(&req, loginUID)
	if err != nil {
		m.Error("批量删除配置失败", zap.Error(err))
		c.ResponseError(err)
		return
	}

	c.Response(result)
}
