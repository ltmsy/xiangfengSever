package common

import (
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"go.uber.org/zap"
)

// 群组通知配置键常量
const (
	ConfigKeyNotifyMemberJoin    = "group.notify_member_join"     // 群成员加入通知
	ConfigKeyNotifyMemberLeave   = "group.notify_member_leave"    // 群成员离开通知
	ConfigKeyNotifyMemberKick    = "group.notify_member_kick"     // 群成员被踢通知
	ConfigKeyNewMemberSeeHistory = "group.new_member_see_history" // 新成员是否可见历史消息
)

// NotificationService 通知配置服务
type NotificationService struct {
	ctx           config.Context
	adminConfigDB *adminConfigDB
	log.Log
}

// NewNotificationService 创建通知配置服务实例
func NewNotificationService(ctx config.Context) *NotificationService {
	return &NotificationService{
		ctx:           ctx,
		adminConfigDB: newAdminConfigDB(ctx.DB()),
		Log:           log.NewTLog("NotificationService"),
	}
}

// IsGroupMemberJoinNotifyEnabled 检查群成员加入通知是否启用
func (n *NotificationService) IsGroupMemberJoinNotifyEnabled() bool {
	return n.isNotificationEnabled(ConfigKeyNotifyMemberJoin)
}

// IsGroupMemberLeaveNotifyEnabled 检查群成员离开通知是否启用
func (n *NotificationService) IsGroupMemberLeaveNotifyEnabled() bool {
	return n.isNotificationEnabled(ConfigKeyNotifyMemberLeave)
}

// IsGroupMemberKickNotifyEnabled 检查群成员被踢通知是否启用
func (n *NotificationService) IsGroupMemberKickNotifyEnabled() bool {
	return n.isNotificationEnabled(ConfigKeyNotifyMemberKick)
}

// IsNewMemberSeeHistoryEnabled 检查新成员是否可见历史消息
func (n *NotificationService) IsNewMemberSeeHistoryEnabled() bool {
	return n.isNotificationEnabled(ConfigKeyNewMemberSeeHistory)
}

// isNotificationEnabled 检查指定的通知配置是否启用
func (n *NotificationService) isNotificationEnabled(configKey string) bool {
	// 获取配置值，如果获取失败或配置不存在，默认返回true（保持原有行为）
	config, err := n.adminConfigDB.GetConfigByKey(configKey)
	if err != nil {
		n.Warn("获取通知配置失败，使用默认值", zap.String("configKey", configKey), zap.Error(err))
		return true // 默认启用通知
	}

	if config == nil {
		n.Warn("通知配置不存在，使用默认值", zap.String("configKey", configKey))
		return true // 默认启用通知
	}

	configValue := config.ConfigValue

	// 记录调试日志
	n.Debug("检查通知配置", zap.String("configKey", configKey), zap.String("configValue", configValue))

	// 配置值为"1"时启用通知，为"0"时禁用通知
	enabled := configValue == "1"

	if !enabled {
		n.Info("通知已禁用", zap.String("configKey", configKey))
	}

	return enabled
}
