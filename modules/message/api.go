package message

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/base/event"
	"github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/channel"
	chservice "github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/channel/service"
	commonapi "github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/common"
	"github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/file"
	"github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/group"
	"github.com/TangSengDaoDao/TangSengDaoDaoServer/modules/user"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/common"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/log"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/network"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/util"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/wkevent"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/wkhttp"
	"github.com/gocraft/dbr/v2"
	"github.com/pkg/errors"
	"github.com/sendgrid/rest"
	"go.uber.org/zap"
)

// Message 消息相关API
type Message struct {
	ctx *config.Context
	log.Log
	db                  *DB
	messageReactionDB   *messageReactionDB
	userDB              *user.DB
	messageExtraDB      *messageExtraDB
	memberReadedDB      *memberReadedDB
	channelOffsetDB     *channelOffsetDB
	deviceOffsetDB      *deviceOffsetDB
	conversationExtradb *conversationExtraDB
	messageUserExtraDB  *messageUserExtraDB
	remindersDB         *remindersDB
	pinnedDB            *pinnedDB
	userService         user.IService
	groupService        group.IService
	commonService       commonapi.IService
	fileService         file.IService
	channelService      chservice.IService
	notificationService *commonapi.NotificationService
	adminConfigService  *commonapi.AdminConfigService
	mutex               sync.Mutex
}

// New New
func New(ctx *config.Context) *Message {

	m := &Message{

		ctx:                 ctx,
		Log:                 log.NewTLog("Message"),
		db:                  NewDB(ctx),
		userDB:              user.NewDB(ctx),
		messageExtraDB:      newMessageExtraDB(ctx),
		groupService:        group.NewService(ctx),
		memberReadedDB:      newMemberReadedDB(ctx),
		conversationExtradb: newConversationExtraDB(ctx),
		messageReactionDB:   newMessageReactionDB(ctx),
		messageUserExtraDB:  newMessageUserExtraDB(ctx),
		channelOffsetDB:     newChannelOffsetDB(ctx),
		deviceOffsetDB:      newDeviceOffsetDB(ctx.DB()),
		remindersDB:         newRemindersDB(ctx),
		pinnedDB:            newPinnedDB(ctx),
		userService:         user.NewService(ctx),
		commonService:       commonapi.NewService(ctx),
		fileService:         file.NewService(ctx),
		channelService:      channel.NewService(ctx),
		notificationService: commonapi.NewNotificationService(*ctx),
		adminConfigService:  commonapi.NewAdminConfigService(ctx.DB()),
	}
	m.ctx.AddEventListener(event.GroupMemberAdd, m.handleGroupMemberAddEvent)
	m.ctx.AddEventListener(event.GroupMemberScanJoin, m.handleGroupMemberScanJoinEvent)
	return m
}

// Route 路由配置
func (m *Message) Route(r *wkhttp.WKHttp) {
	message := r.Group("/v1/message", m.ctx.AuthMiddleware(r))
	{

		message.POST("/sync", m.sync)                             // 同步消息 (写模式才用到 TODO：此方法未来将弃用)
		message.POST("/syncack/:last_message_seq", m.syncack)     // 同步消息回执 （写模式才用到 TODO：此方法未来将弃用）
		message.DELETE("", m.delete)                              // 删除消息
		message.DELETE("/mutual", m.mutualDelete)                 // 双向删除消息
		message.POST("/revoke", m.revoke)                         // 撤回消息
		message.POST("/offset", m.offset)                         // 清除某频道消息
		message.PUT("/voicereaded", m.voiceReaded)                // 语音消息设置为已读
		message.POST("/search", m.search)                         // 消息搜索
		message.POST("/global", m.globalSearch)                   // 全局搜索
		message.POST("/typing", m.typing)                         // 发送typing消息
		message.POST("/channel/sync", m.syncChannelMessage)       // 同步频道消息
		message.POST("/extra/sync", m.syncMessageExtra)           // 同步消息扩展
		message.POST("/readed", m.messageReaded)                  // 消息已读
		message.GET("/sync/sensitivewords", m.syncSensitiveWords) // 同步敏感词
		message.POST("/edit", m.messageEdit)                      // 消息编辑
		message.POST("/reminder/sync", m.reminderSync)            // 同步提醒
		message.POST("/reminder/done", m.reminderDone)            // 提醒已处理完成
		message.GET("/prohibit_words/sync", m.syncProhibitWords)  // 同步违禁词
		message.POST("/pinned", m.pinnedMessage)                  // 置顶消息
		message.POST("/pinned/sync", m.syncPinnedMessage)         // 同步置顶消息
		message.POST("/pinned/clear", m.clearPinnedMessage)       // 删除所有置顶消息
	}
	messages := r.Group("/v1/messages", m.ctx.AuthMiddleware(r))
	{
		// messages.PUT("/:message_id/voicereaded", m.voiceReaded)
		messages.GET("/:message_id/receipt", m.messageReceiptList) // 消息回执列表
	}
	// 回应
	reactions := r.Group("/v1/reactions", m.ctx.AuthMiddleware(r))
	{
		reactions.POST("", m.addOrCancelReaction) // 添加或取消回应
	}
	reaction := r.Group("/v1/reaction", m.ctx.AuthMiddleware(r))
	{
		reaction.POST("/sync", m.syncReaction)
	}
	msg := r.Group("/v1/message")
	{
		msg.POST("/send", m.sendMsg) // 代发消息
	}
	m.ctx.AddMessagesListener(m.listenerMessages) // 监听消息
	m.syncMessageReadedCount()
}

func (m *Message) sendMsg(c *wkhttp.Context) {
	if !m.ctx.GetConfig().Message.SendMessageOn {
		c.ResponseError(errors.New("不支持代发消息"))
		return
	}
	var req struct {
		Token              string                 `json:"token"`                // 发送者
		ReceiveChannelID   string                 `json:"receive_channel_id"`   // 接受者id
		ReceiveChannelType uint8                  `json:"receive_channel_type"` // 接受类型
		Payload            map[string]interface{} `json:"payload"`              // 消息体
	}
	if err := c.BindJSON(&req); err != nil {
		c.ResponseErrorf("数据格式有误！", err)
		return
	}
	if req.Token == "" {
		c.ResponseError(errors.New("发送者token不能为空"))
		return
	}
	if req.ReceiveChannelID == "" {
		c.ResponseError(errors.New("接受channelID不能为空"))
		return
	}
	if req.Payload == nil {
		c.ResponseError(errors.New("消息不能为空"))
		return
	}
	uidAndName, err := m.ctx.Cache().Get(m.ctx.GetConfig().Cache.TokenCachePrefix + req.Token)
	if err != nil {
		m.Error("解析token错误", zap.Error(err))
		c.ResponseError(errors.New("解析token错误"))
		return
	}
	if strings.TrimSpace(uidAndName) == "" {
		c.ResponseError(errors.New("请先登录"))
		return
	}
	uidAndNames := strings.Split(uidAndName, "@")
	if len(uidAndNames) < 2 {
		c.ResponseError(errors.New("token错误"))
		return
	}
	uid := uidAndNames[0]
	if uid == "" {
		c.ResponseError(errors.New("发送者不能为空"))
		return
	}

	if req.ReceiveChannelType == common.ChannelTypePerson.Uint8() {
		sendUserIsFriend, err := m.userService.IsFriend(uid, req.ReceiveChannelID)
		if err != nil {
			m.Error("查询发送者与接受者好友关系错误", zap.Error(err))
			c.ResponseError(errors.New("查询好友关系错误"))
			return
		}
		if !sendUserIsFriend {
			c.ResponseError(errors.New("发送者与接受者不是好友"))
			return
		}
		recvUserIsFriend, err := m.userService.IsFriend(req.ReceiveChannelID, uid)
		if err != nil {
			m.Error("查询接受者与发送者好友关系错误", zap.Error(err))
			c.ResponseError(errors.New("查询接受者与发送者好友关系错误"))
			return
		}
		if !recvUserIsFriend {
			c.ResponseError(errors.New("接受者与发送者不是好友"))
			return
		}
	}
	if req.ReceiveChannelType == common.ChannelTypeGroup.Uint8() {
		isExist, err := m.groupService.ExistMember(req.ReceiveChannelID, uid)
		if err != nil {
			m.Error("查询发送者是否在群内错误", zap.Error(err))
			c.ResponseError(errors.New("查询发送者是否在群内错误"))
			return
		}
		if !isExist {
			c.ResponseError(errors.New("未在群内"))
			return
		}
	}
	err = m.sendMessage(req.ReceiveChannelID, req.ReceiveChannelType, uid, req.Payload)
	if err != nil {
		c.ResponseError(err)
		return
	}
	c.ResponseOK()
}

func (m *Message) sendMessage(channelID string, channelType uint8, fromUID string, payload map[string]interface{}) error {
	err := m.ctx.SendMessage(&config.MsgSendReq{
		Header: config.MsgHeader{
			RedDot: 1,
		},
		ChannelID:   channelID,
		ChannelType: channelType,
		FromUID:     fromUID,
		Payload:     []byte(util.ToJson(payload)),
	})
	if err != nil {
		m.Error("发送消息错误", zap.Error(err))
		return errors.New("发送消息错误")
	}
	return nil
}

// 消息编辑
func (m *Message) messageEdit(c *wkhttp.Context) {
	var req struct {
		MessageID   string `json:"message_id"`
		MessageSeq  uint32 `json:"message_seq"`
		ChannelID   string `json:"channel_id"`
		ChannelType uint8  `json:"channel_type"`
		ContentEdit string `json:"content_edit"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.ResponseErrorf("数据格式有误！", err)
		return
	}
	if req.MessageID == "" {
		c.ResponseError(errors.New("消息ID不能为空！"))
		return
	}
	if req.MessageSeq == 0 {
		c.ResponseError(errors.New("消息序号不能为空！"))
		return
	}
	if req.ChannelID == "" {
		c.ResponseError(errors.New("频道ID不能为空！"))
		return
	}
	contentEdit := dbr.NewNullString(req.ContentEdit).String
	contentMD5 := util.MD5(contentEdit)

	exist, err := m.messageExtraDB.existContentEdit(req.MessageID, contentMD5)
	if err != nil {
		m.Error("查询是否存在相同正文失败！", zap.Error(err))
		c.ResponseError(errors.New("查询是否存在相同正文失败！"))
		return
	}
	if exist {
		m.Warn("存在相同编辑正文，不再处理！")
		c.ResponseOK()
		return
	}

	tx, err := m.db.session.Begin()
	if err != nil {
		m.Error("开启事务失败！", zap.Error(err))
		c.ResponseError(errors.New("开启事务失败！"))
		return
	}
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
			panic(err)
		}
	}()
	fakeChannelID := req.ChannelID
	if req.ChannelType == common.ChannelTypePerson.Uint8() {
		fakeChannelID = common.GetFakeChannelIDWith(c.GetLoginUID(), req.ChannelID)
	}

	version := m.genMessageExtraSeq(fakeChannelID)
	err = m.messageExtraDB.insertOrUpdateContentEditTx(&messageExtraModel{
		MessageID:       req.MessageID,
		MessageSeq:      req.MessageSeq,
		ChannelID:       fakeChannelID,
		ChannelType:     req.ChannelType,
		ContentEdit:     dbr.NewNullString(req.ContentEdit),
		ContentEditHash: contentMD5,
		EditedAt:        int(time.Now().Unix()),
		Version:         version,
	}, tx)
	if err != nil {
		m.Error("添加或修改编辑内容失败！", zap.Error(err))
		c.ResponseError(errors.New("添加或修改编辑内容失败！"))
		return
	}
	msgIds := make([]string, 0)
	msgIds = append(msgIds, req.MessageID)
	// 发布编辑事件
	var eventID int64 = 0
	if m.ctx.GetConfig().ZincSearch.SearchOn {
		eventID, err = m.ctx.EventBegin(&wkevent.Data{
			Event: event.EventUpdateSearchMessage,
			Data: &config.UpdateSearchMessageReq{
				MessageIDs: msgIds,
				ChannelID:  req.ChannelID,
			},
			Type: wkevent.None,
		}, tx)
		if err != nil {
			tx.Rollback()
			m.Error("开启事件失败！", zap.Error(err))
			c.ResponseError(errors.New("开启事件失败！"))
			return
		}
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.ResponseErrorf("事务提交失败！", err)
		return
	}
	if eventID > 0 {
		m.ctx.EventCommit(eventID)
	}

	err = m.ctx.SendCMD(config.MsgCMDReq{
		NoPersist:   true,
		ChannelID:   req.ChannelID,
		ChannelType: req.ChannelType,
		FromUID:     c.GetLoginUID(),
		CMD:         common.CMDSyncMessageExtra,
	})

	if err != nil {
		m.Error("发送cmd失败！", zap.Error(err))
		c.ResponseError(err)
		return
	}
	c.ResponseOK()
}

// 消息已读
func (m *Message) messageReaded(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	var req struct {
		MessageIDs  []string `json:"message_ids"`
		ChannelID   string   `json:"channel_id"`
		ChannelType uint8    `json:"channel_type"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.ResponseErrorf("数据格式有误！", err)
		return
	}
	if len(req.MessageIDs) == 0 {
		c.ResponseError(errors.New("message_ids不能为空！"))
		return
	}
	// var cloneNo string
	fakeChannelID := req.ChannelID
	if req.ChannelType == common.ChannelTypePerson.Uint8() {
		fakeChannelID = common.GetFakeChannelIDWith(req.ChannelID, loginUID)
	}

	// 添加详细日志
	m.Info("消息已读请求参数",
		zap.String("loginUID", loginUID),
		zap.String("reqChannelID", req.ChannelID),
		zap.Uint8("reqChannelType", req.ChannelType),
		zap.String("fakeChannelID", fakeChannelID),
		zap.Strings("messageIDs", req.MessageIDs))

	if len(req.MessageIDs) <= 0 {
		c.ResponseOK()
		return
	}
	messageIDStrs := util.RemoveRepeatedElement(req.MessageIDs)

	messages, err := m.db.queryMessagesWithMessageIDs(fakeChannelID, messageIDStrs)
	if err != nil {
		m.Error("查询消息失败！",
			zap.String("fakeChannelID", fakeChannelID),
			zap.Strings("messageIDs", messageIDStrs),
			zap.String("loginUID", loginUID),
			zap.String("reqChannelID", req.ChannelID),
			zap.Uint8("reqChannelType", req.ChannelType))
		c.ResponseErrorf("查询消息失败！", err)
		return
	}
	if len(messages) <= 0 {
		m.Warn("没有读取到消息！",
			zap.String("fakeChannelID", fakeChannelID),
			zap.Strings("messageIDs", messageIDStrs),
			zap.String("loginUID", loginUID),
			zap.String("reqChannelID", req.ChannelID),
			zap.Uint8("reqChannelType", req.ChannelType))
		c.ResponseError(errors.New("没有读取到消息！"))
		return
	}

	tx, err := m.ctx.DB().Begin()
	if err != nil {
		m.Error("开启事务失败！", zap.Error(err))
		c.ResponseError(errors.New("开启事务失败！"))
		return
	}
	defer func() {
		if err := recover(); err != nil {
			tx.RollbackUnlessCommitted()
			panic(err)
		}
	}()

	// //	fromUIDs := make([]string, 0, len(messages)) // 消息发送者
	// for _, message := range messages {
	// 	//	fromUIDs = append(fromUIDs, message.FromUID)
	// 	err := m.memberReadedDB.insertOrUpdateTx(&memberReadedModel{
	// 		MessageID:   message.MessageID,
	// 		ChannelID:   fakeChannelID,
	// 		ChannelType: req.ChannelType,
	// 		UID:         loginUID,
	// 	}, tx)
	// 	if err != nil {
	// 		tx.Rollback()
	// 		c.ResponseErrorf("添加已读数据失败！", err)
	// 		return
	// 	}
	// }
	// 构建批量插入的数据
	readedModels := make([]*memberReadedModel, 0, len(messages))
	for _, message := range messages {
		readedModels = append(readedModels, &memberReadedModel{
			MessageID:   message.MessageID,
			ChannelID:   fakeChannelID,
			ChannelType: req.ChannelType,
			UID:         loginUID,
		})
	}
	// 批量插入或更新已读记录
	err = m.memberReadedDB.batchInsertOrUpdateTx(readedModels, tx)
	if err != nil {
		tx.Rollback()
		c.ResponseErrorf("添加已读数据失败！", err)
		return
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.ResponseErrorf("提交事务失败！", err)
		return
	}

	// 异步处理 Redis 缓存
	go func() {
		for _, message := range messages {
			messageIDStr := strconv.FormatInt(message.MessageID, 10)
			jsonStr, err := json.Marshal(&messageReadedCountModel{
				MessageIDStr:   messageIDStr,
				MessageID:      message.MessageID,
				MessageSeq:     message.MessageSeq,
				FromUID:        message.FromUID,
				ChannelID:      fakeChannelID,
				ChannelType:    req.ChannelType,
				LoginUID:       loginUID,
				ReqChannelID:   req.ChannelID,
				ReqChannelType: req.ChannelType,
			})
			if err != nil {
				m.Error("序列化消息错误", zap.Error(err))
				continue
			}

			m.mutex.Lock()
			err = m.ctx.GetRedisConn().SetAndExpire(
				fmt.Sprintf("%s%s", CacheReadedCountPrefix, messageIDStr),
				jsonStr,
				time.Hour*24*7,
			)
			m.mutex.Unlock()

			if err != nil {
				m.Error("添加消息扩展数据到缓存失败！",
					zap.Error(err),
					zap.Int64("messageID", message.MessageID),
					zap.String("channelID", fakeChannelID),
				)
			}
		}
	}()
	c.ResponseOK()

}

// 消息回执列表
func (m *Message) messageReceiptList(c *wkhttp.Context) {
	messageIDStr := c.Param("message_id")

	readed := c.Query("readed") // 查询已读未读的消息，0.未读 1.已读
	pIndex, pSize := c.GetPage()

	resps := make([]memberReceiptResp, 0)
	uids := make([]string, 0)
	if readed == "1" {
		memberReadedModels, err := m.memberReadedDB.queryWithMessageIDAndPage(messageIDStr, uint64(pIndex), uint64(pSize))
		if err != nil {
			c.ResponseErrorf("查询已读列表失败！", err)
			return
		}
		if len(memberReadedModels) > 0 {
			for _, memberReadedM := range memberReadedModels {
				uids = append(uids, memberReadedM.UID)
			}
		}
	}
	userResps, err := m.userService.GetUsers(uids)
	if err != nil {
		c.ResponseErrorf("查询用户数据失败！", err)
		return
	}
	userMap := map[string]*user.Resp{}
	if len(userResps) > 0 {
		for _, userResp := range userResps {
			userMap[userResp.UID] = userResp
		}
	}

	for _, uid := range uids {
		userResp := userMap[uid]
		var name string
		if userResp != nil {
			name = userResp.Name
		}
		resps = append(resps, memberReceiptResp{
			UID:  uid,
			Name: name,
		})
	}
	c.Response(resps)

}

//	func (m *Message) getCacheMessageReactionSeq(uid, channelID string, channelType uint8) (int64, error) {
//		versionStr, err := m.ctx.GetRedisConn().Hget(fmt.Sprintf("messageReactionSeq:%s", uid), fmt.Sprintf("%s-%d", channelID, channelType))
//		if err != nil {
//			return 0, err
//		}
//		if versionStr == "" {
//			return 0, nil
//		}
//		version, _ := strconv.ParseInt(versionStr, 10, 64)
//		return version, nil
//	}
func (m *Message) getMessageExtraVersion(uid, source, channelID string, channelType uint8) (int64, error) {
	versionStr, err := m.ctx.GetRedisConn().Hget(fmt.Sprintf("messageExtraVersion:%s%s", uid, source), fmt.Sprintf("%s-%d", channelID, channelType))
	if err != nil {
		return 0, err
	}
	if versionStr == "" {
		return 0, nil
	}
	version, _ := strconv.ParseInt(versionStr, 10, 64)
	return version, nil

}

func (m *Message) setMessageExtraVersion(uid, channelID string, channelType uint8, source string, messageExtraVersion int64) error {
	err := m.ctx.GetRedisConn().Hset(fmt.Sprintf("messageExtraVersion:%s%s", uid, source), fmt.Sprintf("%s-%d", channelID, channelType), fmt.Sprintf("%d", messageExtraVersion))
	if err != nil {
		return err
	}
	return nil
}

// 同步扩展消息数据
func (m *Message) syncMessageExtra(c *wkhttp.Context) {
	var req struct {
		ChannelID    string `json:"channel_id"`
		ChannelType  uint8  `json:"channel_type"`
		ExtraVersion int64  `json:"extra_version"`
		Source       string `json:"source"` // 操作源
		Limit        int    `json:"limit"`  // 数据限制
	}
	if err := c.BindJSON(&req); err != nil {
		c.ResponseErrorf("数据格式有误！", err)
		return
	}
	fakeChannelID := req.ChannelID
	if req.ChannelType == common.ChannelTypePerson.Uint8() {
		fakeChannelID = common.GetFakeChannelIDWith(c.GetLoginUID(), req.ChannelID)
	}
	cacheExtraVersion, err := m.getMessageExtraVersion(c.GetLoginUID(), req.Source, fakeChannelID, req.ChannelType)
	if err != nil {
		c.ResponseErrorf("从缓存中获取消息扩展版本失败！", err)
		return
	}
	extraVersion := req.ExtraVersion
	if cacheExtraVersion >= extraVersion {
		extraVersion = cacheExtraVersion
	} else {
		err = m.setMessageExtraVersion(c.GetLoginUID(), fakeChannelID, req.ChannelType, req.Source, extraVersion)
		if err != nil {
			c.ResponseErrorf("缓存最大的消息扩展版本失败！", err)
			return
		}

	}
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 10000 {
		limit = 10000
	}
	if strings.TrimSpace(req.ChannelID) == "" {
		c.ResponseError(errors.New("频道ID不能为空！"))
		return
	}
	extraModels, err := m.messageExtraDB.sync(extraVersion, fakeChannelID, req.ChannelType, uint64(limit), c.GetLoginUID())
	if err != nil {
		c.ResponseErrorf("同步消息扩展数据失败！", err)
		return
	}
	resps := make([]*messageExtraResp, 0, len(extraModels))
	if len(extraModels) > 0 {
		for _, extraModel := range extraModels {
			resps = append(resps, newMessageExtraResp(extraModel))
		}
	}
	c.Response(resps)
}

// 同步频道消息
func (m *Message) syncChannelMessage(c *wkhttp.Context) {
	var req config.SyncChannelMessageReq
	if err := c.BindJSON(&req); err != nil {
		m.Error("数据格式有误！", zap.Error(err))
		c.ResponseError(errors.New("数据格式有误！"))
		return
	}

	// 如果当前用户不在群内，则直接返回空消息数组
	if req.ChannelType == common.ChannelTypeGroup.Uint8() {
		exist, err := m.groupService.ExistMember(req.ChannelID, c.GetLoginUID())
		if err != nil {
			m.Error("查询是否在群内存在失败！", zap.Error(err))
			c.ResponseError(errors.New("查询是否在群内存在失败！"))
			return
		}
		if !exist {
			c.JSON(http.StatusOK, &syncChannelMessageResp{
				StartMessageSeq: req.EndMessageSeq,
				EndMessageSeq:   req.EndMessageSeq,
				PullMode:        req.PullMode,
				Messages:        make([]*MsgSyncResp, 0),
			})
			return
		}
	}
	req.LoginUID = c.GetLoginUID()
	resp, err := m.ctx.IMSyncChannelMessage(req)
	if err != nil {
		m.Error("同步频道内的消息失败！", zap.Error(err), zap.String("req", util.ToJson(req)))
		c.ResponseError(errors.New("同步频道内的消息失败！"))
		return
	}
	fakeChannelID := req.ChannelID
	if req.ChannelType == common.ChannelTypePerson.Uint8() { // 如果是群则需要计算群成员是否变化 如果有变化则将群成员加入到克隆表
		fakeChannelID = common.GetFakeChannelIDWith(req.LoginUID, req.ChannelID)
	}
	channelIds := make([]string, 0)
	channelIds = append(channelIds, fakeChannelID)
	channelSettings, err := m.channelService.GetChannelSettings(channelIds)
	if err != nil {
		m.Error("查询频道设置错误", zap.Error(err), zap.String("req", util.ToJson(req)))
		c.ResponseError(errors.New("查询频道设置错误"))
		return
	}
	var channelOffsetMessageSeq uint32 = 0
	if len(channelSettings) > 0 && channelSettings[0].OffsetMessageSeq > 0 {
		channelOffsetMessageSeq = channelSettings[0].OffsetMessageSeq
	}
	c.Response(newSyncChannelMessageResp(resp, c.GetLoginUID(), req.DeviceUUID, req.ChannelID, req.ChannelType, m.messageExtraDB, m.messageUserExtraDB, m.messageReactionDB, m.channelOffsetDB, m.deviceOffsetDB, channelOffsetMessageSeq))
}

// 输入中
func (m *Message) typing(c *wkhttp.Context) {
	loginUID := c.MustGet("uid").(string)
	loginName := c.MustGet("name").(string)
	var req struct {
		ChannelID   string `json:"channel_id"`
		ChannelType uint8  `json:"channel_type"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.ResponseError(err)
		return
	}
	channelID := req.ChannelID
	channelType := req.ChannelType
	if req.ChannelType == common.ChannelTypePerson.Uint8() {
		channelID = loginUID
	}
	// 发送输入中的命令
	err := m.ctx.SendCMD(config.MsgCMDReq{
		NoPersist:   true,
		CMD:         common.CMDTyping,
		ChannelID:   req.ChannelID,
		ChannelType: req.ChannelType,
		Param: map[string]interface{}{
			"from_uid":     loginUID,
			"from_name":    loginName,
			"channel_id":   channelID,
			"channel_type": channelType,
		},
	})
	if err != nil {
		c.ResponseError(err)
		return
	}
	c.ResponseOK()
}

// 搜索消息
func (m *Message) search(c *wkhttp.Context) {
	var req struct {
		UID         string `json:"uid"` // 搜索的消息限定这某个用户内
		ChannelID   string `json:"channel_id"`
		ChannelType uint8  `json:"channel_type"`
		ContentType int    `json:"content_type"` // 正文类型
		Keyword     string `json:"keyword"`
	}
	if err := c.BindJSON(&req); err != nil {
		m.Error("数据格式有误！", zap.Error(err))
		c.ResponseError(err)
		return
	}
	uid := c.MustGet("uid").(string)
	req.UID = uid
	fmt.Println("req->", req)
	resp, err := network.Post(fmt.Sprintf("%s/message/search", m.ctx.GetConfig().WuKongIM.APIURL), []byte(util.ToJson(req)), nil)
	if err != nil {
		m.Error("调用搜索失败！", zap.Error(err))
		c.ResponseError(errors.New("调用搜索失败！"))
		return
	}
	err = m.handlerIMError(resp)
	if err != nil {
		m.Error("调用搜索错误！", zap.Error(err))
		c.ResponseError(errors.New("调用搜索错误！"))
		return
	}
	var results []map[string]interface{}
	err = util.ReadJsonByByte([]byte(resp.Body), &results)
	if err != nil {
		m.Error("解析搜索数据失败！", zap.Error(err))
		c.ResponseError(errors.New("解析搜索数据失败！"))
		return
	}
	c.JSON(http.StatusOK, results)
}

// 语音消息设置为已读
func (m *Message) voiceReaded(c *wkhttp.Context) {
	var req *voiceReadedReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseErrorf("数据格式有误！", err)
		return
	}
	if err := req.check(); err != nil {
		c.ResponseError(err)
		return
	}
	loginUID := c.GetLoginUID()

	err := m.messageUserExtraDB.insertOrUpdateVoiceRead(&messageUserExtraModel{
		UID:         loginUID,
		MessageID:   req.MessageID,
		MessageSeq:  req.MessageSeq,
		ChannelID:   req.ChannelID,
		ChannelType: req.ChannelType,
		VoiceReaded: 1,
	})
	if err != nil {
		c.ResponseErrorf("修改语音已读状态失败！", err)
		return
	}
	c.ResponseOK()
}

// 同步回应数据
func (m *Message) syncReaction(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	var req struct {
		ChannelID   string `json:"channel_id"`
		ChannelType uint8  `json:"channel_type"`
		Seq         int64  `json:"seq"` // 同步序列号
		Limit       uint64 `json:"limit"`
	}
	if err := c.BindJSON(&req); err != nil {
		m.Error("数据格式有误！", zap.Error(err))
		c.ResponseError(errors.New("数据格式有误！"))
		return
	}
	fakeChannelID := req.ChannelID
	if req.ChannelType == common.ChannelTypePerson.Uint8() {
		if !strings.Contains(req.ChannelID, "@") {
			fakeChannelID = common.GetFakeChannelIDWith(loginUID, req.ChannelID)
		}
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	// cacheReactionSeq, err := m.getCacheMessageReactionSeq(loginUID, req.ChannelID, req.ChannelType)
	// if err != nil {
	// 	m.Error("获取缓存messageSeq失败", zap.Error(err))
	// 	c.ResponseError(errors.New("获取缓存messageSeq失败"))
	// 	return
	// }
	// if req.Seq <= cacheReactionSeq {
	// 	req.Seq = cacheReactionSeq
	// }
	list, err := m.messageReactionDB.queryReactionWithChannelAndSeq(fakeChannelID, req.ChannelType, req.Seq, limit)
	if err != nil {
		m.Error("获取缓存seq错误", zap.Error(err))
		c.ResponseError(errors.New("获取缓存seq错误"))
		return
	}

	toChannelID := common.GetToChannelIDWithFakeChannelID(fakeChannelID, loginUID)

	reactions := make([]*reactionResp, 0)
	if len(list) > 0 {
		for _, model := range list {
			reactions = append(reactions, &reactionResp{
				UID:         model.UID,
				Name:        model.Name,
				ChannelID:   toChannelID,
				ChannelType: model.ChannelType,
				Seq:         model.Seq,
				MessageID:   model.MessageID,
				CreatedAt:   model.CreatedAt.String(),
				Emoji:       model.Emoji,
				IsDeleted:   model.IsDeleted,
			})
		}
	}
	c.JSON(http.StatusOK, reactions)
}

// 添加或取消回应
func (m *Message) addOrCancelReaction(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	loginName := c.GetLoginName()
	var req struct {
		MessageID   string `json:"message_id"`   // 消息唯一ID
		ChannelID   string `json:"channel_id"`   // 频道唯一ID
		ChannelType uint8  `json:"channel_type"` // 频道类型
		Emoji       string `json:"emoji"`        // 回应的emoji
	}
	if err := c.BindJSON(&req); err != nil {
		m.Error("数据格式有误！", zap.Error(err))
		c.ResponseError(errors.New("数据格式有误！"))
		return
	}
	model, err := m.messageReactionDB.queryReactionWithUIDAndMessageID(loginUID, req.MessageID)
	if err != nil {
		m.Error("查询登录用户是否回应消息错误", zap.Error(err))
		c.ResponseError(errors.New("查询登录用户是否回应消息错误"))
		return
	}
	fakeChannelID := req.ChannelID
	if req.ChannelType == common.ChannelTypePerson.Uint8() { // 如果是群则需要计算群成员是否变化 如果有变化则将群成员加入到克隆表
		fakeChannelID = common.GetFakeChannelIDWith(req.ChannelID, loginUID)
	}
	seq := m.genMessageReactionSeq(fakeChannelID) // 下次回复seq
	if model == nil {
		//新增回应
		err = m.messageReactionDB.insertReaction(&reactionModel{
			ChannelID:   fakeChannelID,
			ChannelType: req.ChannelType,
			UID:         loginUID,
			Name:        loginName,
			MessageID:   req.MessageID,
			Emoji:       req.Emoji,
			Seq:         seq,
			IsDeleted:   0,
		})
		if err != nil {
			m.Error("新增消息回应错误", zap.Error(err))
			c.ResponseError(errors.New("新增消息回应错误"))
			return
		}
	} else {
		model.Seq = seq
		if model.IsDeleted == 1 {
			model.IsDeleted = 0
			if model.Emoji != req.Emoji {
				model.Emoji = req.Emoji
			}
		} else {
			if model.Emoji == req.Emoji {
				model.IsDeleted = 1
			} else {
				model.Emoji = req.Emoji
			}
		}
		err = m.messageReactionDB.updateReactionStatus(model)
		if err != nil {
			m.Error("修改消息回应错误", zap.Error(err))
			c.ResponseError(errors.New("修改消息回应错误"))
			return
		}
	}

	//发送同步消息cmd
	err = m.ctx.SendCMD(config.MsgCMDReq{
		NoPersist:   true,
		ChannelID:   req.ChannelID,
		ChannelType: uint8(req.ChannelType),
		CMD:         common.CMDSyncMessageReaction,
		FromUID:     loginUID,
	})
	if err != nil {
		m.Error("发送同步命令失败！", zap.Error(err))
		c.ResponseErrorf("发送同步命令失败！", err)
		return
	}

	c.ResponseOK()
}
func (m *Message) handlerIMError(resp *rest.Response) error {
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest {
			resultMap, err := util.JsonToMap(resp.Body)
			if err != nil {
				return err
			}
			if resultMap != nil && resultMap["msg"] != nil {
				return fmt.Errorf("IM Extend服务失败！ -> %s", resultMap["msg"])
			}
		}
		return fmt.Errorf("IM Extend服务返回状态[%d]失败！", resp.StatusCode)
	}
	return nil
}

// 同步消息回执
func (m *Message) syncack(c *wkhttp.Context) {
	uid := c.MustGet("uid").(string)
	lastMessageSeqStr := c.Param("last_message_seq")
	lastMessageSeq, err := strconv.ParseUint(lastMessageSeqStr, 10, 64)
	if err != nil {
		m.Error("last_message_seq格式有误！", zap.String("last_message_seq", lastMessageSeqStr))
		c.ResponseError(errors.New("last_message_seq格式有误！"))
		return
	}
	err = m.ctx.IMSyncMessageAck(&config.SyncackReq{
		UID:            uid,
		LastMessageSeq: uint32(lastMessageSeq),
	})
	if err != nil {
		m.Error("同步消息回执失败！", zap.Error(err), zap.String("uid", uid), zap.String("last_message_seq", lastMessageSeqStr))
		c.ResponseError(errors.New("同步消息回执失败！"))
		return
	}
	c.ResponseOK()
}

// 同步消息
func (m *Message) sync(c *wkhttp.Context) {
	uid := c.MustGet("uid").(string)
	var req syncReq
	if err := c.BindJSON(&req); err != nil {
		m.Error(common.ErrData.Error(), zap.Error(err))
		c.ResponseError(common.ErrData)
		return
	}
	resps, err := m.ctx.IMSyncMessage(&config.MsgSyncReq{
		UID:        uid,
		MessageSeq: req.MaxMessageSeq,
		Limit:      req.Limit,
	})
	if err != nil {
		m.Error("同步消息失败！", zap.Error(err), zap.String("uid", uid))
		c.ResponseError(errors.New("同步消息失败！"))
		return
	}
	messageIDs := make([]string, 0, len(resps))
	for _, message := range resps {
		messageIDs = append(messageIDs, fmt.Sprintf("%d", message.MessageID))
	}

	// 全局扩充数据
	messageExtras, err := m.messageExtraDB.queryWithMessageIDsAndUID(messageIDs, c.GetLoginUID())
	if err != nil {
		log.Error("查询消息扩展字段失败！", zap.Error(err))
	}
	messageExtraMap := map[string]*messageExtraDetailModel{}
	if len(messageExtras) > 0 {
		for _, messageExtra := range messageExtras {
			messageExtraMap[messageExtra.MessageID] = messageExtra
		}
	}
	// 用户扩充数据
	messageUserExtras, err := m.messageUserExtraDB.queryWithMessageIDsAndUID(messageIDs, c.GetLoginUID())
	if err != nil {
		log.Error("查询用户消息扩展字段失败！", zap.Error(err))
	}
	messageUserExtraMap := map[string]*messageUserExtraModel{}
	if len(messageUserExtras) > 0 {
		for _, messageUserExtraM := range messageUserExtras {
			messageUserExtraMap[messageUserExtraM.MessageID] = messageUserExtraM
		}
	}
	// 用户频道偏移
	channelOffsetM, err := m.channelOffsetDB.queryWithUIDAndChannel(c.GetLoginUID(), req.ChannelID, req.ChannelType)
	if err != nil {
		m.Error("查询偏移量失败！", zap.Error(err))
		c.ResponseError(errors.New("查询偏移量失败！"))
		return
	}
	// 频道偏移
	fakeChannelID := req.ChannelID
	if req.ChannelType == common.ChannelTypePerson.Uint8() {
		fakeChannelID = common.GetFakeChannelIDWith(req.ChannelID, uid)
	}
	channelIds := make([]string, 0)
	channelIds = append(channelIds, fakeChannelID)
	channelSettings, err := m.channelService.GetChannelSettings(channelIds)
	if err != nil {
		m.Error("查询频道设置错误", zap.Error(err), zap.String("req", util.ToJson(req)))
		c.ResponseError(errors.New("查询频道设置错误"))
		return
	}
	var channelOffsetMessageSeq uint32 = 0
	if len(channelSettings) > 0 && channelSettings[0].OffsetMessageSeq > 0 {
		channelOffsetMessageSeq = channelSettings[0].OffsetMessageSeq
	}
	respVos := make([]*MsgSyncResp, 0)
	for _, resp := range resps {
		if channelOffsetM != nil && resp.MessageSeq <= channelOffsetM.MessageSeq {
			continue
		}
		messageIDStr := strconv.FormatInt(resp.MessageID, 10)
		messageExtraM := messageExtraMap[messageIDStr]
		messageUserExtraM := messageUserExtraMap[messageIDStr]
		respVo := &MsgSyncResp{}
		respVo.from(resp, c.GetLoginUID(), messageExtraM, messageUserExtraM, nil, channelOffsetMessageSeq)
		respVos = append(respVos, respVo)
	}

	c.JSON(http.StatusOK, respVos)
}

// 双向删除
func (m *Message) mutualDelete(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	var req deleteReq
	if err := c.BindJSON(&req); err != nil {
		m.Error("数据格式有误！", zap.Error(err))
		c.ResponseError(errors.New("数据格式有误！"))
		return
	}
	if err := req.check(); err != nil {
		c.ResponseError(err)
		return
	}

	// 读取双向删除配置
	doubleDeleteEnabled := false
	config, err := m.adminConfigService.GetConfigByKey("message.double_delete_enabled")
	if err == nil && config != nil {
		doubleDeleteEnabled = config.ConfigValue == "1" || config.ConfigValue == "true"
	}

	messageSeqs := make([]uint32, 0)
	messageSeqs = append(messageSeqs, req.MessageSeq)
	fakeChannelID := req.ChannelID
	if req.ChannelType == common.ChannelTypePerson.Uint8() {
		fakeChannelID = common.GetFakeChannelIDWith(loginUID, req.ChannelID)
	}
	resp, err := m.ctx.IMGetWithChannelAndSeqs(req.ChannelID, req.ChannelType, loginUID, messageSeqs)
	if err != nil {
		m.Error("查询消息错误", zap.Error(err), zap.String("req", util.ToJson(req)))
		c.ResponseError(errors.New("查询消息错误"))
		return
	}

	if resp == nil || len(resp.Messages) == 0 {
		c.ResponseError(errors.New("消息不存在"))
		return
	}
	isCanDelete := true
	if req.ChannelType == common.ChannelTypeGroup.Uint8() {
		isManager, err := m.groupService.IsCreatorOrManager(req.ChannelID, loginUID)
		if err != nil {
			m.Error("查询登录用户群内权限错误", zap.Error(err))
			c.ResponseError(errors.New("查询登录用户群内权限错误"))
			return
		}
		if resp.Messages[0].FromUID != loginUID && !isManager {
			isCanDelete = false
		}
	}
	if !isCanDelete {
		c.ResponseError(errors.New("用户无权删除此消息"))
		return
	}

	// 如果启用双向删除，需要通知其他可见该消息的用户也删除该条消息
	if doubleDeleteEnabled {
		m.Info("双向删除功能已启用，将通知其他用户删除消息",
			zap.String("message_id", req.MessageID),
			zap.String("channel_id", req.ChannelID),
			zap.Uint8("channel_type", req.ChannelType))

		// TODO: 这里可以添加逻辑来获取所有可见该消息的用户
		// 并通知他们删除该条消息
		// 例如：通过悟空IM的API发送删除命令
	}

	version := m.genMessageExtraSeq(fakeChannelID)
	err = m.messageExtraDB.insertOrUpdateDeleted(&messageExtraModel{
		MessageID:   req.MessageID,
		ChannelID:   fakeChannelID,
		ChannelType: req.ChannelType,
		IsDeleted:   1,
		Version:     version,
	})
	if err != nil {
		m.Error("删除消息错误", zap.Error(err))
		c.ResponseError(errors.New("删除消息错误"))
		return
	}
	err = m.ctx.SendCMD(config.MsgCMDReq{
		NoPersist:   true,
		ChannelID:   req.ChannelID,
		ChannelType: req.ChannelType,
		FromUID:     c.GetLoginUID(),
		CMD:         common.CMDSyncMessageExtra,
	})

	if err != nil {
		m.Error("发送cmd失败！", zap.Error(err))
		c.ResponseError(err)
		return
	}
	c.ResponseOK()
}

// 删除消息
func (m *Message) delete(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	var reqs []*deleteReq
	if err := c.BindJSON(&reqs); err != nil {
		m.Error("数据格式有误！", zap.Error(err))
		c.ResponseError(errors.New("数据格式有误！"))
		return
	}
	if len(reqs) == 0 {
		c.ResponseError(errors.New("参数不能为空！"))
		return
	}
	for _, req := range reqs {
		if err := req.check(); err != nil {
			c.ResponseError(err)
			return
		}
	}

	tx, err := m.ctx.DB().Begin()
	if err != nil {
		m.Error("开启事务失败！", zap.Error(err))
		c.ResponseError(errors.New("开启事务失败！"))
		return
	}
	defer func() {
		if err := recover(); err != nil {
			tx.RollbackUnlessCommitted()
			panic(err)
		}
	}()
	for _, req := range reqs {
		err := m.messageUserExtraDB.insertOrUpdateDeletedTx(&messageUserExtraModel{
			UID:              loginUID,
			MessageID:        req.MessageID,
			MessageSeq:       req.MessageSeq,
			ChannelID:        req.ChannelID,
			ChannelType:      req.ChannelType,
			MessageIsDeleted: 1,
		}, tx)
		if err != nil {
			tx.Rollback()
			m.Error("删除消息失败！", zap.Error(err))
			c.ResponseError(errors.New("删除消息失败！"))
			return
		}
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		m.Error("提交事务失败！", zap.Error(err))
		c.ResponseError(errors.New("提交事务失败！"))
		return
	}

	err = m.ctx.SendCMD(config.MsgCMDReq{
		NoPersist:   true,
		ChannelID:   loginUID,
		ChannelType: common.ChannelTypePerson.Uint8(),
		CMD:         CMDMessageDeleted,
		Param: map[string]interface{}{
			"messages": reqs,
		},
	})
	if err != nil {
		m.Error("发送命令失败", zap.Error(err))
		c.ResponseError(errors.New("发送命令失败"))
		return
	}

	c.ResponseOK()
}

func (m *Message) genMessageExtraSeq(channelID string) int64 {
	return time.Now().UnixNano() / 1e3
	// return m.ctx.GenSeq(fmt.Sprintf("%s:%s", common.MessageExtraSeqKey, channelID))
}
func (m *Message) genMessageReactionSeq(channelID string) int64 {
	return m.ctx.GenSeq(fmt.Sprintf("%s:%s", common.MessageReactionSeqKey, channelID))
}

// 消息偏移
func (m *Message) offset(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	var req struct {
		ChannelID   string `json:"channel_id"`
		ChannelType uint8  `json:"channel_type"`
		MessageSeq  uint32 `json:"message_seq"`
	}
	if err := c.BindJSON(&req); err != nil {
		m.Error("数据格式有误！", zap.Error(err))
		c.ResponseError(errors.New("数据格式有误！"))
		return
	}
	channelOffsetM, err := m.channelOffsetDB.queryWithUIDAndChannel(c.GetLoginUID(), req.ChannelID, req.ChannelType)
	if err != nil {
		m.Error("查询频道偏移数据失败！", zap.Error(err))
		c.ResponseError(errors.New("查询频道偏移数据失败！"))
		return
	}
	if channelOffsetM != nil {
		if channelOffsetM.MessageSeq >= req.MessageSeq {
			c.ResponseOK()
			return
		}
	}

	err = m.channelOffsetDB.insertOrUpdate(&channelOffsetModel{
		UID:         c.GetLoginUID(),
		ChannelID:   req.ChannelID,
		ChannelType: req.ChannelType,
		MessageSeq:  req.MessageSeq,
	})
	if err != nil {
		m.Error("清除失败！", zap.Error(err))
		c.ResponseError(errors.New("清除失败！"))
		return
	}
	// 清除最近会话的未读数（这里不管有没有未读数都调用清除）
	err = m.ctx.IMClearConversationUnread(config.ClearConversationUnreadReq{
		UID:         c.GetLoginUID(),
		ChannelID:   req.ChannelID,
		ChannelType: req.ChannelType,
		MessageSeq:  req.MessageSeq,
		Unread:      0,
	})
	if err != nil {
		m.Error("清除最近会话未读数失败！", zap.Error(err), zap.String("uid", c.GetLoginUID()), zap.String("channelID", req.ChannelID), zap.Uint8("channelType", req.ChannelType))
	}
	// 清空提醒项
	reminders, err := m.remindersDB.queryWithUIDAndChannel(loginUID, req.ChannelID, req.ChannelType, req.MessageSeq)
	if err != nil {
		m.Error("查询用户提醒项失败！", zap.Error(err))
		c.ResponseError(errors.New("查询用户提醒项失败！"))
		return
	}
	reminderIds := make([]int64, 0)
	if len(reminders) > 0 {
		for _, reminder := range reminders {
			if reminder.MessageSeq <= req.MessageSeq && reminder.Done == 0 {
				reminderIds = append(reminderIds, reminder.Id)
			}
		}
	}

	if len(reminderIds) > 0 {
		tx, err := m.ctx.DB().Begin()
		if err != nil {
			m.Error("开启事务失败！", zap.Error(err))
			c.ResponseError(errors.New("开启事务失败！"))
			return
		}
		defer func() {
			if err := recover(); err != nil {
				tx.RollbackUnlessCommitted()
				panic(err)
			}
		}()
		err = m.remindersDB.insertDonesTx(reminderIds, loginUID, tx)
		if err != nil {
			tx.Rollback()
			m.Error("更新提醒项状态失败！", zap.Error(err))
			c.ResponseError(errors.New("更新提醒项状态失败！"))
			return
		}
		for _, id := range reminderIds {
			version := m.ctx.GenSeq(common.RemindersKey)
			err = m.remindersDB.updateVersionTx(version, id, tx)
			if err != nil {
				tx.Rollback()
				m.Error("更新提醒项版本失败！", zap.Error(err))
				c.ResponseError(errors.New("更新提醒项版本失败！"))
				return
			}
		}
		if err := tx.Commit(); err != nil {
			tx.Rollback()
			m.Error("提交事务失败！", zap.Error(err))
			c.ResponseError(errors.New("提交事务失败！"))
			return
		}
		err = m.ctx.SendCMD(config.MsgCMDReq{
			NoPersist:   true,
			ChannelID:   req.ChannelID,
			ChannelType: req.ChannelType,
			CMD:         common.CMDSyncReminders,
		})
		if err != nil {
			m.Error("发送cmd[CMDSyncReminders]失败！", zap.Error(err))
		}
	}
	// 发送清空红点的命令
	err = m.ctx.SendCMD(config.MsgCMDReq{
		NoPersist:   true,
		ChannelID:   c.GetLoginUID(),
		ChannelType: common.ChannelTypePerson.Uint8(),
		CMD:         common.CMDConversationUnreadClear,
		Param: map[string]interface{}{
			"channel_id":   req.ChannelID,
			"channel_type": req.ChannelType,
			"unread":       0,
		},
	})
	if err != nil {
		m.Error("命令发送失败！", zap.String("cmd", common.CMDConversationUnreadClear), zap.String("uid", c.GetLoginUID()), zap.String("channelID", req.ChannelID), zap.Uint8("channelType", req.ChannelType))
	}

	c.ResponseOK()
}

// 是否有撤回的权限
func (m *Message) hasRevokePermission(messageM *messageModel, loginUID string) (bool, error) {
	if messageM.FromUID == "" { // 没有fromUID的消息一般是命令类的消息，不被允许撤回
		return false, nil
	}
	if messageM.FromUID == loginUID { // 自己发的消息允许被撤回
		return true, nil
	}
	if messageM.ChannelType == common.ChannelTypeGroup.Uint8() { // 管理者或创建者可以撤回其他成员的消息
		loginMember, err := m.groupService.GetMember(messageM.ChannelID, loginUID)
		if err != nil {
			return false, err
		}
		if loginMember == nil {
			return false, nil
		}
		fromMember, err := m.groupService.GetMember(messageM.ChannelID, messageM.FromUID)
		if err != nil {
			return false, err
		}
		if fromMember == nil && loginMember.Role != int(common.GroupMemberRoleNormal) {
			return true, nil
		}
		if fromMember.Role == int(common.GroupMemberRoleCreater) || loginMember.Role == int(common.GroupMemberRoleNormal) {
			return false, nil
		}
		if loginMember.Role == int(common.GroupMemberRoleCreater) || (loginMember.Role == int(common.GroupMemberRoleManager) && fromMember.Role == int(common.GroupMemberRoleNormal)) {
			return true, nil
		}

	}

	return false, nil
}

func (m *Message) cancelMentionReminderIfNeed(message *messageModel) {
	setting := config.SettingFromUint8(message.Setting)
	//  如果撤回的是@消息，需要取消提醒
	if !setting.Signal {
		var payloadMap map[string]interface{}
		if err := util.ReadJsonByByte(message.Payload, &payloadMap); err != nil {
			m.Warn("解码消息内容失败！", zap.Error(err))
		}
		if payloadMap != nil {
			if m.hasMention(payloadMap) {
				all, uids := m.getMention(payloadMap)
				if all {
					version := m.ctx.GenSeq(common.RemindersKey)
					err := m.remindersDB.deleteWithChannel(message.ChannelID, message.ChannelType, message.MessageID, version)
					if err != nil {
						m.Error("删除提醒项失败！", zap.Error(err))
					} else {
						err = m.ctx.SendCMD(config.MsgCMDReq{
							NoPersist:   true,
							ChannelID:   message.ChannelID,
							ChannelType: message.ChannelType,
							CMD:         common.CMDSyncReminders,
						})
						if err != nil {
							m.Error("发送cmd[CMDSyncReminders]失败！", zap.Error(err))
						}
					}
				} else if len(uids) > 0 {
					tx, err := m.ctx.DB().Begin()
					if err != nil {
						m.Error("开启事务失败！", zap.Error(err))
						return
					}
					defer func() {
						if err := recover(); err != nil {
							tx.RollbackUnlessCommitted()
							panic(err)
						}
					}()
					for _, uid := range uids {
						version := m.ctx.GenSeq(common.RemindersKey)
						err := m.remindersDB.deleteWithChannelAndUIDTx(message.ChannelID, message.ChannelType, uid, message.MessageID, version, tx)
						if err != nil {
							m.Error("删除用户提醒项失败！", zap.Error(err))
							tx.Rollback()
							return
						}
					}
					if err := tx.Commit(); err != nil {
						m.Error("提交事务失败！", zap.Error(err))
						tx.RollbackUnlessCommitted()
						return
					}
					err = m.ctx.SendCMD(config.MsgCMDReq{
						NoPersist:   true,
						Subscribers: uids,
						CMD:         common.CMDSyncReminders,
					})
					if err != nil {
						m.Error("发送cmd[CMDSyncReminders]失败！", zap.Error(err))
					}
				}
			}
		}
	}
}

// 撤回消息
func (m *Message) revoke(c *wkhttp.Context) {
	loginUID := c.MustGet("uid").(string)
	messageID := c.Query("message_id")
	clientMsgNo := c.Query("client_msg_no") // TODO：后续版本不再使用messageID撤回，使用client_msg_no撤回，因为存在重试消息，clientMsgNo一样 但是messageID不一样
	channelID := c.Query("channel_id")
	channelType := c.Query("channel_type")

	if strings.TrimSpace(clientMsgNo) == "" {
		c.ResponseError(errors.New("撤回主键参数错误！"))
		return
	}

	//删除消息
	channelTypeI, _ := strconv.ParseUint(channelType, 10, 64)

	fakeChannelID := channelID
	if uint8(channelTypeI) == common.ChannelTypePerson.Uint8() {
		fakeChannelID = common.GetFakeChannelIDWith(c.GetLoginUID(), channelID)
	}

	var messageIDs = []string{}
	var err error
	var messages []*messageModel
	if clientMsgNo != "" {
		messages, err = m.db.queryMessagesWithChannelClientMsgNo(fakeChannelID, uint8(channelTypeI), clientMsgNo)
		if err != nil {
			m.Error("撤回失败！", zap.String("fakeChannelID", fakeChannelID), zap.String("clientMsgNo", clientMsgNo), zap.String("loginUID", c.GetLoginUID()))
			c.ResponseErrorf("查询消息失败！", err)
			return
		}
		if len(messages) == 0 {
			c.ResponseError(errors.New("撤回失败！"))
			return
		}
		var message *messageModel
		if len(messages) > 0 {
			message = messages[0]
			for _, message := range messages {
				messageIDs = append(messageIDs, fmt.Sprintf("%d", message.MessageID))
			}
		}
		if message != nil {
			allow, err := m.hasRevokePermission(message, c.GetLoginUID())
			if err != nil {
				m.Error("权限判断失败！", zap.Error(err))
				c.ResponseError(errors.New("权限判断失败！"))
				return
			}
			if !allow {
				c.ResponseError(errors.New("无权限撤回此消息！"))
				return
			}

			m.cancelMentionReminderIfNeed(message)

		}
	}

	if len(messageIDs) == 0 {
		messageIDs = append(messageIDs, messageID)
	}
	messageExtras, err := m.messageExtraDB.queryWithMessageIDs(messageIDs)
	if err != nil {
		m.Error("查询消息扩展错误", zap.Error(err))
		c.ResponseError(errors.New("查询消息扩展错误"))
		return
	}

	tx, err := m.db.session.Begin()
	if err != nil {
		m.Error("开启事务失败！", zap.Error(err))
		c.ResponseError(errors.New("开启事务失败！"))
		return
	}
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
			panic(err)
		}
	}()
	for _, msgID := range messageIDs {
		version := m.genMessageExtraSeq(fakeChannelID)
		// err = m.messageExtraDB.insertOrUpdateRevokeTx(&messageExtraModel{
		// 	MessageID:   msgID,
		// 	ChannelID:   fakeChannelID,
		// 	ChannelType: uint8(channelTypeI),
		// 	Revoke:      1,
		// 	Version:     version,
		// 	Revoker:     loginUID,
		// }, tx)
		// if err != nil {
		// 	tx.Rollback()
		// 	c.ResponseErrorf("更新消息为撤回状态失败！", err)
		// 	return
		// }
		var tempMsgExtra *messageExtraModel
		for _, messageExtra := range messageExtras {
			if messageExtra.MessageID == msgID {
				tempMsgExtra = messageExtra
				tempMsgExtra.Revoke = 1
				tempMsgExtra.Revoker = loginUID
				tempMsgExtra.Version = version
				break
			}
		}
		if tempMsgExtra != nil {
			err = m.messageExtraDB.updateTx(tempMsgExtra, tx)
			if err != nil {
				tx.Rollback()
				m.Error("更新消息扩展数据失败！", zap.Error(err), zap.String("messageID", msgID), zap.String("channelID", fakeChannelID))
				return
			}
		} else {
			fromUID := ""
			msgSeq := uint32(0)
			if len(messages) > 0 {
				for _, msg := range messages {
					messageIDStr := strconv.FormatInt(msg.MessageID, 10)
					if messageIDStr == msgID {
						fromUID = msg.FromUID
						msgSeq = msg.MessageSeq
						break
					}
				}
			}
			err = m.messageExtraDB.insertTx(&messageExtraModel{
				MessageID:   msgID,
				MessageSeq:  msgSeq,
				FromUID:     fromUID,
				ChannelID:   fakeChannelID,
				ChannelType: uint8(channelTypeI),
				ReadedCount: 0,
				Version:     version,
				Revoke:      1,
				Revoker:     loginUID,
			}, tx)
			if err != nil {
				tx.Rollback()
				m.Error("新增消息扩展数据失败！", zap.Error(err), zap.String("messageID", msgID), zap.String("channelID", fakeChannelID))
				return
			}
		}
	}
	msgIds := make([]string, 0)
	msgIds = append(msgIds, messageID)
	// 发布撤回消息事件
	eventID, err := m.ctx.EventBegin(&wkevent.Data{
		Event: event.EventUpdateSearchMessage,
		Data: &config.UpdateSearchMessageReq{
			MessageIDs: msgIds,
			ChannelID:  channelID,
		},
		Type: wkevent.None,
	}, tx)
	if err != nil {
		tx.Rollback()
		m.Error("开启事件失败！", zap.Error(err))
		c.ResponseError(errors.New("开启事件失败！"))
		return
	}
	err = m.deletePinnedMessage(channelID, uint8(channelTypeI), messageIDs, loginUID, tx)
	if err != nil {
		c.ResponseError(err)
		return
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.ResponseErrorf("事务提交失败！", err)
		return
	}
	m.ctx.EventCommit(eventID)
	// err = m.ctx.SendCMD(config.MsgCMDReq{
	// 	NoPersist:   true,
	// 	ChannelID:   channelID,
	// 	ChannelType: uint8(channelTypeI),
	// 	FromUID:     loginUID,
	// 	CMD:         common.CMDSyncMessageExtra,
	// })
	// if err != nil {
	// 	c.ResponseErrorf("发送同步命令失败！", err)
	// 	return
	// }

	for _, msgID := range messageIDs {
		messageIDI, _ := strconv.ParseInt(msgID, 10, 64)
		// 发给指定频道
		err = m.ctx.SendRevoke(&config.MsgRevokeReq{
			Operator:     loginUID,
			OperatorName: c.GetLoginName(),
			FromUID:      loginUID,
			ChannelID:    channelID,
			ChannelType:  uint8(channelTypeI),
			MessageID:    messageIDI,
		})
		if err != nil {
			m.Error("发送撤回消息失败！", zap.Error(err))
			c.ResponseError(errors.New("发送撤回消息失败！"))
			return
		}
	}

	c.ResponseOK()

}

// 同步违禁词
func (m *Message) syncProhibitWords(c *wkhttp.Context) {
	version := c.Query("version")
	maxVersion, _ := strconv.ParseInt(version, 10, 64)
	list, err := m.db.queryProhibitWordsWithVersion(maxVersion)
	if err != nil {
		m.Error("同步违禁词错误", zap.Error(err))
		c.ResponseError(errors.New("同步违禁词错误"))
		return
	}
	result := make([]*ProhibitWordResp, 0)
	if len(list) > 0 {
		for _, word := range list {
			result = append(result, &ProhibitWordResp{
				Id:        word.Id,
				Content:   word.Content,
				IsDeleted: word.IsDeleted,
				CreatedAt: word.CreatedAt.String(),
				Version:   word.Version,
			})
		}
	}
	c.Response(result)
}

// 同步敏感词
func (m *Message) syncSensitiveWords(c *wkhttp.Context) {
	type resp struct {
		Tips    string   `json:"tips"`
		List    []string `json:"list"`
		Version int64    `json:"version"`
	}
	reqVersion, _ := strconv.ParseInt(c.Query("version"), 10, 64)
	resultList := make([]string, 0)
	tips := ""
	if reqVersion < sensitiveWordsVersion {
		resultList = sensitive_words
		tips = "涉及私下交易、转账等资金问题，谨慎对待，谨防上当受骗，点击标题栏头像可投诉！"
	}
	c.Response(&resp{
		Tips:    tips,
		List:    resultList,
		Version: sensitiveWordsVersion,
	})
}

// // 接受IM的消息
// func (m *Message) notify(c *wkhttp.Context) {
// 	data, err := c.GetRawData()
// 	if err != nil {
// 		m.Error("notify读取数据失败！", zap.Error(err))
// 		c.ResponseError(err)
// 		return
// 	}
// 	var msgResps []msgResp
// 	err = util.ReadJsonByByte(data, &msgResps)
// 	if err != nil {
// 		m.Error("读取消息数据失败！", zap.Error(err))
// 		c.ResponseError(err)
// 		return
// 	}
// 	tx, _ := m.db.session.Begin()
// 	defer func() {
// 		if err := recover(); err != nil {
// 			tx.Rollback()
// 			panic(err)
// 		}
// 	}()
// 	messageIDS := make([]string, 0, len(msgResps))
// 	for _, msgResp := range msgResps {
// 		messageIDS = append(messageIDS, strconv.FormatUint(msgResp.MessageID, 10))
// 		messageModel := msgResp.ToModel()
// 		err = m.db.InsertTx(messageModel, tx)
// 		if err != nil {
// 			tx.Rollback()
// 			m.Error("添加消息失败！", zap.Any("msg", msgResp), zap.Error(err))
// 			c.ResponseError(err)
// 			return
// 		}
// 	}
// 	if err := tx.Commit(); err != nil {
// 		tx.Rollback()
// 		m.Error("提交事务失败！", zap.Error(err))
// 		c.ResponseError(err)
// 		return
// 	}
// 	c.Response(messageIDS)
// }

// ---------- vo ----------

type syncChannelMessageResp struct {
	StartMessageSeq uint32          `json:"start_message_seq"` // 开始序列号
	EndMessageSeq   uint32          `json:"end_message_seq"`   // 结束序列号
	PullMode        config.PullMode `json:"pull_mode"`         // 拉取模式
	More            int             `json:"more"`              // 是否还有更多 1.是 0.否
	Messages        []*MsgSyncResp  `json:"messages"`          // 消息数据
}

func newSyncChannelMessageResp(resp *config.SyncChannelMessageResp, loginUID string, deviceUUID string, channelID string, channelType uint8, messageExtraDB *messageExtraDB, messageUserExtraDB *messageUserExtraDB, messageReactionDB *messageReactionDB, channelOffsetDB *channelOffsetDB, deviceOffsetDB *deviceOffsetDB, channelOffsetMessageSeq uint32) *syncChannelMessageResp {
	messages := make([]*MsgSyncResp, 0, len(resp.Messages))
	if len(resp.Messages) > 0 {
		messageIDs := make([]string, 0, len(resp.Messages))
		for _, message := range resp.Messages {
			var payloadMap map[string]interface{}
			err := util.ReadJsonByByte(message.Payload, &payloadMap)
			if err != nil {
				log.Warn("负荷数据不是json格式！", zap.Error(err), zap.String("payload", string(message.Payload)))
			}
			if len(payloadMap) > 0 {
				replyJson := payloadMap["reply"]
				if replyJson != nil {
					msgId := replyJson.(map[string]interface{})["message_id"].(string)
					messageIDs = append(messageIDs, msgId)
				}
			}
			messageIDs = append(messageIDs, fmt.Sprintf("%d", message.MessageID))
		}

		// 消息全局扩张
		messageExtras, err := messageExtraDB.queryWithMessageIDsAndUID(messageIDs, loginUID)
		if err != nil {
			log.Error("查询消息扩展字段失败！", zap.Error(err))
		}
		// 修改消息扩展字段
		for _, message := range resp.Messages {
			var payloadMap map[string]interface{}
			err := util.ReadJsonByByte(message.Payload, &payloadMap)
			if err != nil {
				log.Warn("负荷数据不是json格式！", zap.Error(err), zap.String("payload", string(message.Payload)))
			}
			if len(payloadMap) > 0 {
				replyJson := payloadMap["reply"]
				if replyJson == nil {
					continue
				}
				msgId := replyJson.(map[string]interface{})["message_id"].(string)
				for _, messageExtra := range messageExtras {
					if messageExtra.MessageID == msgId {
						var contentEditMap map[string]interface{}
						if messageExtra.ContentEdit.String != "" {
							err := util.ReadJsonByByte([]byte(messageExtra.ContentEdit.String), &contentEditMap)
							if err != nil {
								log.Warn("负荷数据不是json格式！", zap.Error(err), zap.String("payload", string(messageExtra.ContentEdit.String)))
								continue
							}
							replyJson.(map[string]interface{})["payload"] = contentEditMap
							payloadMap["reply"] = replyJson
							message.Payload = []byte(util.ToJson(payloadMap))
						}
						break
					}
				}
			}
		}
		messageExtraMap := map[string]*messageExtraDetailModel{}
		if len(messageExtras) > 0 {
			for _, messageExtra := range messageExtras {
				messageExtraMap[messageExtra.MessageID] = messageExtra
			}
		}

		// 消息用户扩张
		messageUserExtras, err := messageUserExtraDB.queryWithMessageIDsAndUID(messageIDs, loginUID)
		if err != nil {
			log.Error("查询用户消息扩展字段失败！", zap.Error(err))
		}
		messageUserExtraMap := map[string]*messageUserExtraModel{}
		if len(messageUserExtras) > 0 {
			for _, messageUserExtraM := range messageUserExtras {
				messageUserExtraMap[messageUserExtraM.MessageID] = messageUserExtraM
			}
		}

		// 查询消息回应
		messageReaction, err := messageReactionDB.queryWithMessageIDs(messageIDs)
		if err != nil {
			log.Error("查询消息回应数据错误", zap.Error(err))
		}
		messageReactionMap := map[string][]*reactionModel{}
		if len(messageReaction) > 0 {
			for _, reaction := range messageReaction {
				msgReactionList := messageReactionMap[reaction.MessageID]
				if msgReactionList == nil {
					msgReactionList = make([]*reactionModel, 0)
				}
				msgReactionList = append(msgReactionList, reaction)
				messageReactionMap[reaction.MessageID] = msgReactionList
			}
		}

		// 用户频道偏移
		channelOffsetM, err := channelOffsetDB.queryWithUIDAndChannel(loginUID, channelID, channelType)
		if err != nil {
			log.Error("查询频道偏移量失败！", zap.Error(err))
		}

		// 设备偏移
		deviceLastMessageSeq, err := deviceOffsetDB.queryMessageSeq(loginUID, deviceUUID, channelID, channelType)
		if err != nil {
			log.Error("查询设备消息偏移量失败！", zap.Error(err))
		}
		for _, message := range resp.Messages {
			if channelOffsetM != nil && message.MessageSeq <= channelOffsetM.MessageSeq {
				continue
			}
			if message.MessageSeq <= uint32(deviceLastMessageSeq) {
				continue
			}
			messageIDStr := strconv.FormatInt(message.MessageID, 10)
			messageExtra := messageExtraMap[messageIDStr]
			messageUserExtra := messageUserExtraMap[messageIDStr]
			msgResp := &MsgSyncResp{}
			msgResp.from(message, loginUID, messageExtra, messageUserExtra, messageReactionMap[strconv.FormatInt(message.MessageID, 10)], channelOffsetMessageSeq)
			messages = append(messages, msgResp)
		}
	}
	return &syncChannelMessageResp{
		StartMessageSeq: resp.StartMessageSeq,
		EndMessageSeq:   resp.EndMessageSeq,
		PullMode:        resp.PullMode,
		Messages:        messages,
	}
}

// 消息头
type messageHeader struct {
	NoPersist int `json:"no_persist"` // 是否不持久化
	RedDot    int `json:"red_dot"`    // 是否显示红点
	SyncOnce  int `json:"sync_once"`  // 此消息只被同步或被消费一次
}

type syncReq struct {
	MaxMessageSeq uint32 `json:"max_message_seq"` // 客户端最大消息序列号
	Limit         int    `json:"limit"`           // 消息数量限制
	ChannelID     string `json:"channel_id"`      // 频道ID
	ChannelType   uint8  `json:"channel_type"`    // 频道类型
	Reverse       int    `json:"reverse"`         // 是否倒序
	Offset        int64  `json:"offset"`          // 偏移量
}

// type msgResp struct {
// 	MessageID   uint64 `json:"message_id"`   // 服务端的消息ID(全局唯一)
// 	FromUID     string `json:"from_uid"`     // 发送者UID
// 	ChannelID   string `json:"channel_id"`   // 频道ID
// 	ChannelType uint8  `json:"channel_type"` // 频道类型
// 	Timestamp   int64  `json:"timestamp"`    // 服务器消息时间戳(10位，到秒)
// 	Payload     []byte `json:"payload"`      // 消息内容
// }

// func (m msgResp) ToModel() *messageModel {
// 	var payloadMap map[string]interface{}
// 	err := util.ReadJsonByByte(m.Payload, &payloadMap)
// 	if err != nil {
// 		log.Warn("负荷数据不是json格式！", zap.Error(err), zap.String("payload", string(m.Payload)))
// 	}
// 	contentType := 0
// 	if payloadMap != nil {
// 		if payloadMap["type"] != nil {
// 			contentTypeInt64, _ := payloadMap["type"].(json.Number).Int64()
// 			contentType = int(contentTypeInt64)
// 		}
// 		// if payloadMap["content"] != nil {
// 		// 	keyword = payloadMap["content"].(string)
// 		// }
// 	}
// 	return &messageModel{
// 		MessageID:   int64(m.MessageID),
// 		FromUID:     m.FromUID,
// 		ChannelID:   m.ChannelID,
// 		ChannelType: m.ChannelType,
// 		Timestamp:   m.Timestamp,
// 		Payload:     m.Payload,
// 		Type:        contentType,
// 	}
// }

// type replyMsgSyncResp struct {
// 	Root     *config.MessageResp   `json:"root"`
// 	Messages []*config.MessageResp `json:"messages"`
// }

// MgSyncResp 消息同步请求
type MsgSyncResp struct {
	Header        messageHeader          `json:"header"`                    // 消息头部
	Setting       uint8                  `json:"setting"`                   // 设置
	MessageID     int64                  `json:"message_id"`                // 服务端的消息ID(全局唯一)
	MessageIDStr  string                 `json:"message_idstr"`             // 服务端的消息ID(全局唯一)字符串形式
	MessageSeq    uint32                 `json:"message_seq"`               // 消息序列号 （用户唯一，有序递增）
	ClientMsgNo   string                 `json:"client_msg_no"`             // 客户端消息唯一编号
	StreamNo      string                 `json:"stream_no,omitempty"`       // 流编号
	FromUID       string                 `json:"from_uid"`                  // 发送者UID
	ToUID         string                 `json:"to_uid,omitempty"`          // 接受者uid
	ChannelID     string                 `json:"channel_id"`                // 频道ID
	ChannelType   uint8                  `json:"channel_type"`              // 频道类型
	Expire        uint32                 `json:"expire,omitempty"`          // expire
	Timestamp     int32                  `json:"timestamp"`                 // 服务器消息时间戳(10位，到秒)
	Payload       map[string]interface{} `json:"payload"`                   // 消息内容
	SignalPayload string                 `json:"signal_payload"`            // signal 加密后的payload base64编码,TODO: 这里为了兼容没加密的版本，所以新用SignalPayload字段
	ReplyCount    int                    `json:"reply_count,omitempty"`     // 回复集合
	ReplyCountSeq string                 `json:"reply_count_seq,omitempty"` // 回复数量seq
	ReplySeq      string                 `json:"reply_seq,omitempty"`       // 回复seq
	Reactions     []*reactionSimpleResp  `json:"reactions,omitempty"`       // 回应数据
	IsDeleted     int                    `json:"is_deleted"`                // 是否已删除
	VoiceStatus   int                    `json:"voice_status,omitempty"`    // 语音状态 0.未读 1.已读
	Streams       []*streamItemResp      `json:"streams,omitempty"`         // 流数据
	// ---------- 旧字段 这些字段都放到MessageExtra对象里了 ----------
	Readed       int    `json:"readed"`                 // 是否已读（针对于自己）
	Revoke       int    `json:"revoke,omitempty"`       // 是否撤回
	Revoker      string `json:"revoker,omitempty"`      // 消息撤回者
	ReadedCount  int    `json:"readed_count,omitempty"` // 已读数量
	UnreadCount  int    `json:"unread_count,omitempty"` // 未读数量
	ExtraVersion int64  `json:"extra_version"`          // 扩展数据版本号

	// 消息扩展字段
	MessageExtra *messageExtraResp `json:"message_extra,omitempty"` // 消息扩展

}

func (m *MsgSyncResp) from(msgResp *config.MessageResp, loginUID string, messageExtraM *messageExtraDetailModel, messageUserExtraM *messageUserExtraModel, reactionModels []*reactionModel, channelOffsetMessageSeq uint32) {
	m.Header.NoPersist = msgResp.Header.NoPersist
	m.Header.RedDot = msgResp.Header.RedDot
	m.Header.SyncOnce = msgResp.Header.SyncOnce
	m.Setting = msgResp.Setting
	m.MessageID = msgResp.MessageID
	m.MessageIDStr = strconv.FormatInt(msgResp.MessageID, 10)
	m.MessageSeq = msgResp.MessageSeq
	m.ClientMsgNo = msgResp.ClientMsgNo
	m.StreamNo = msgResp.StreamNo
	m.FromUID = msgResp.FromUID
	m.ToUID = msgResp.ToUID
	m.ChannelID = msgResp.ChannelID
	m.ChannelType = msgResp.ChannelType
	m.Expire = msgResp.Expire
	m.Timestamp = msgResp.Timestamp
	if messageExtraM != nil {
		// TODO: 后续这些字段可以废除了 都放MessageExtra对象里了
		m.IsDeleted = messageExtraM.IsDeleted
		m.Revoke = messageExtraM.Revoke
		m.Revoker = messageExtraM.Revoker
		m.ReadedCount = messageExtraM.ReadedCount
		m.Readed = messageExtraM.Readed
		m.ExtraVersion = messageExtraM.Version

		m.MessageExtra = newMessageExtraResp(messageExtraM)
	}

	setting := config.SettingFromUint8(msgResp.Setting)
	var payloadMap map[string]interface{}
	if setting.Signal {
		m.SignalPayload = base64.StdEncoding.EncodeToString(msgResp.Payload)
		payloadMap = map[string]interface{}{
			"type": common.SignalError.Int(),
		}
	} else {
		err := util.ReadJsonByByte(msgResp.Payload, &payloadMap)
		if err != nil {
			log.Warn("负荷数据不是json格式！", zap.Error(err), zap.String("payload", string(msgResp.Payload)))
		}
		if len(payloadMap) > 0 {
			visibles := payloadMap["visibles"]
			if visibles != nil {
				visiblesArray := visibles.([]interface{})
				if len(visiblesArray) > 0 {
					m.IsDeleted = 1
					for _, limitUID := range visiblesArray {
						if limitUID == loginUID {
							m.IsDeleted = 0
						}
					}
				}
			}
		} else {
			payloadMap = map[string]interface{}{
				"type": common.ContentError.Int(),
			}
		}
	}

	if messageUserExtraM != nil {
		if m.IsDeleted == 0 {
			m.IsDeleted = messageUserExtraM.MessageIsDeleted
		}
		m.VoiceStatus = messageUserExtraM.VoiceReaded
	}

	if msgResp.Expire > 0 {
		if time.Now().Unix()-int64(msgResp.Expire) >= int64(msgResp.Timestamp) {
			m.IsDeleted = 1
		}
	}
	if channelOffsetMessageSeq != 0 && msgResp.MessageSeq <= channelOffsetMessageSeq {
		m.IsDeleted = 1
	}
	m.Payload = payloadMap

	msgReactionList := make([]*reactionSimpleResp, 0, len(reactionModels))
	if len(reactionModels) > 0 {
		for _, reaction := range reactionModels {
			msgReactionList = append(msgReactionList, &reactionSimpleResp{
				UID:       reaction.UID,
				Name:      reaction.Name,
				Seq:       reaction.Seq,
				IsDeleted: reaction.IsDeleted,
				Emoji:     reaction.Emoji,
				CreatedAt: reaction.CreatedAt.String(),
			})
		}
	}
	m.Reactions = msgReactionList

	if len(msgResp.Streams) > 0 {
		streams := make([]*streamItemResp, 0, len(msgResp.Streams))
		for _, streamItem := range msgResp.Streams {
			streams = append(streams, newStreamItemResp(streamItem))
		}
		m.Streams = streams
	}

}

type streamItemResp struct {
	StreamSeq   uint32         `json:"stream_seq"`    // 流序号
	ClientMsgNo string         `json:"client_msg_no"` // 客户端消息唯一编号
	Blob        map[string]any `json:"blob"`          // 消息内容
}

func newStreamItemResp(streamItem *config.StreamItemResp) *streamItemResp {
	var blobMap map[string]any
	err := util.ReadJsonByByte(streamItem.Blob, &blobMap)
	if err != nil {
		log.Warn("blob不是json格式！", zap.Error(err), zap.String("blob", string(streamItem.Blob)))
	}
	return &streamItemResp{
		ClientMsgNo: streamItem.ClientMsgNo,
		StreamSeq:   streamItem.StreamSeq,
		Blob:        blobMap,
	}
}

// 回应返回
type reactionResp struct {
	MessageID   string `json:"message_id"`   // 消息编号
	ChannelID   string `json:"channel_id"`   // 频道ID
	ChannelType uint8  `json:"channel_type"` // 频道类型
	Seq         int64  `json:"seq"`          // 回复序列号
	UID         string `json:"uid"`          // 回应用户ID
	Name        string `json:"name"`         // 回应用户名
	Emoji       string `json:"emoji"`        // 回应的emoji
	IsDeleted   int    `json:"is_deleted"`   // 是否删除
	CreatedAt   string `json:"created_at"`
}

// 回应返回
type reactionSimpleResp struct {
	Seq       int64  `json:"seq"`        // 回复序列号
	UID       string `json:"uid"`        // 回应用户ID
	Name      string `json:"name"`       // 回应用户名
	Emoji     string `json:"emoji"`      // 回应的emoji
	IsDeleted int    `json:"is_deleted"` // 是否删除
	CreatedAt string `json:"created_at"`
}

// type userResp struct {
// 	UID       string `json:"uid"`
// 	Name      string `json:"name"`
// 	IsDeleted int    `json:"is_deleted"`
// }

// type syncTotalResp struct {
// 	MessageID   string `json:"message_id"`   // 消息唯一ID
// 	Seq         string `json:"seq"`          // 回复序列号
// 	ChannelID   string `json:"channel_id"`   // 频道唯一ID
// 	ChannelType uint8  `json:"channel_type"` // 频道类型
// 	Count       int    `json:"count"`        // 回复数量
// }

type messageExtraResp struct {
	MessageID       int64                  `json:"message_id"`
	MessageIDStr    string                 `json:"message_id_str"`
	Revoke          int                    `json:"revoke,omitempty"`
	Revoker         string                 `json:"revoker,omitempty"`
	VoiceStatus     int                    `json:"voice_status,omitempty"`
	Readed          int                    `json:"readed,omitempty"`            // 是否已读（针对于自己）
	ReadedCount     int                    `json:"readed_count,omitempty"`      // 已读数量
	ReadedAt        int64                  `json:"readed_at,omitempty"`         // 已读时间
	IsMutualDeleted int                    `json:"is_mutual_deleted,omitempty"` // 双向删除
	IsPinned        int                    `json:"is_pinned,omitempty"`         // 是否置顶
	ContentEdit     map[string]interface{} `json:"content_edit,omitempty"`      // 编辑后的正文
	EditedAt        int                    `json:"edited_at,omitempty"`         // 编辑时间 例如 12:23
	ExtraVersion    int64                  `json:"extra_version"`               // 数据版本
}

func newMessageExtraResp(m *messageExtraDetailModel) *messageExtraResp {

	messageID, _ := strconv.ParseInt(m.MessageID, 10, 64)

	var contentEditMap map[string]interface{}
	if m.ContentEdit.String != "" {
		err := util.ReadJsonByByte([]byte(m.ContentEdit.String), &contentEditMap)
		if err != nil {
			log.Warn("负荷数据不是json格式！", zap.Error(err), zap.String("payload", string(m.ContentEdit.String)))
		}
	}

	var readedAt int64 = 0
	if m.ReadedAt.Valid {
		readedAt = m.ReadedAt.Time.Unix()
	}

	return &messageExtraResp{
		MessageID:       messageID,
		MessageIDStr:    m.MessageID,
		Revoke:          m.Revoke,
		Revoker:         m.Revoker,
		Readed:          m.Readed,
		ReadedAt:        readedAt,
		ReadedCount:     m.ReadedCount,
		ContentEdit:     contentEditMap,
		EditedAt:        m.EditedAt,
		IsMutualDeleted: m.IsDeleted,
		IsPinned:        m.IsPinned,
		ExtraVersion:    m.Version,
	}
}

type memberReceiptResp struct {
	UID  string `json:"uid"`  // 成员uid
	Name string `json:"name"` // 成员名称
}

type ProhibitWordResp struct {
	Id        int64  `json:"id"`
	Content   string `json:"content"`    // 违禁词
	IsDeleted int    `json:"is_deleted"` // 是否删除
	Version   int64  `json:"version"`    // 版本
	CreatedAt string `json:"created_at"` // 时间
}

// 全局搜索请求参数
type globalSearchReq struct {
	Keyword     string   `json:"keyword"`      // 搜索关键词
	SearchTypes []string `json:"search_types"` // 搜索类型数组
	Filters     struct {
		ChannelID   string `json:"channel_id"`   // 限定频道ID
		ChannelType uint8  `json:"channel_type"` // 频道类型
		ContentType uint8  `json:"content_type"` // 内容类型
		DateRange   struct {
			StartTime string `json:"start_time"` // 开始时间
			EndTime   string `json:"end_time"`   // 结束时间
		} `json:"date_range"`
		FileTypes []string `json:"file_types"` // 文件类型过滤
	} `json:"filters"`
	Pagination struct {
		Page     uint32 `json:"page"`      // 页码
		PageSize uint32 `json:"page_size"` // 每页数量
	} `json:"pagination"`
	Sort struct {
		Field string `json:"field"` // 排序字段
		Order string `json:"order"` // 排序方向
	} `json:"sort"`
}

// 全局搜索结果
type globalSearchResp struct {
	Total   int64                  `json:"total"`
	Results map[string]interface{} `json:"results"`
}

// 聊天搜索结果
type chatSearchResult struct {
	Type           string  `json:"type"`
	MessageID      string  `json:"message_id"`
	ChannelID      string  `json:"channel_id"`
	ChannelType    uint8   `json:"channel_type"`
	FromUID        string  `json:"from_uid"`
	FromName       string  `json:"from_name"`
	Content        string  `json:"content"`
	ContentType    uint8   `json:"content_type"`
	Timestamp      string  `json:"timestamp"`
	RelevanceScore float64 `json:"relevance_score"`
	Highlight      struct {
		Content string `json:"content"`
	} `json:"highlight"`
}

// 用户搜索结果
type userSearchResult struct {
	Type           string  `json:"type"`
	UID            string  `json:"uid"`
	Name           string  `json:"name"`
	Username       string  `json:"username"`
	ShortNo        string  `json:"short_no"`
	Avatar         string  `json:"avatar"`
	RelevanceScore float64 `json:"relevance_score"`
	Highlight      struct {
		Name string `json:"name"`
	} `json:"highlight"`
}

// 群组搜索结果
type groupSearchResult struct {
	Type           string  `json:"type"`
	GroupNo        string  `json:"group_no"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	MemberCount    uint32  `json:"member_count"`
	RelevanceScore float64 `json:"relevance_score"`
	Highlight      struct {
		Name string `json:"name"`
	} `json:"highlight"`
}

// 应用搜索结果
type appSearchResult struct {
	Type           string  `json:"type"`
	AppID          string  `json:"app_id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	Icon           string  `json:"icon"`
	RelevanceScore float64 `json:"relevance_score"`
	Highlight      struct {
		Name string `json:"name"`
	} `json:"highlight"`
}

// 全局搜索
func (m *Message) globalSearch(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	var req globalSearchReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseErrorf("数据格式有误！", err)
		return
	}

	// 验证必填参数
	if req.Keyword == "" {
		c.ResponseError(errors.New("搜索关键词不能为空"))
		return
	}
	if len(req.SearchTypes) == 0 {
		c.ResponseError(errors.New("搜索类型不能为空"))
		return
	}

	// 设置默认分页参数
	if req.Pagination.Page == 0 {
		req.Pagination.Page = 1
	}
	if req.Pagination.PageSize == 0 {
		req.Pagination.PageSize = 20
	}
	if req.Pagination.PageSize > 100 {
		req.Pagination.PageSize = 100
	}

	// 设置默认排序
	if req.Sort.Field == "" {
		req.Sort.Field = "relevance"
	}
	if req.Sort.Order == "" {
		req.Sort.Order = "desc"
	}

	m.Info("全局搜索请求",
		zap.String("loginUID", loginUID),
		zap.String("keyword", req.Keyword),
		zap.Strings("searchTypes", req.SearchTypes))

	results := make(map[string]interface{})
	total := int64(0)

	// 并发执行各类型搜索
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, searchType := range req.SearchTypes {
		wg.Add(1)
		go func(searchType string) {
			defer wg.Done()
			var result interface{}
			var count int64
			var err error

			switch searchType {
			case "chat":
				result, count, err = m.searchChatMessages(req.Keyword, req.Filters, req.Pagination)
			case "user":
				result, count, err = m.searchUsers(req.Keyword, req.Pagination)
			case "group":
				result, count, err = m.searchGroups(req.Keyword, req.Pagination)
			case "app":
				result, count, err = m.searchApps(req.Keyword, req.Pagination)
			default:
				m.Warn("不支持的搜索类型", zap.String("searchType", searchType))
				return
			}

			if err != nil {
				m.Error("搜索失败", zap.String("searchType", searchType), zap.Error(err))
				return
			}

			mu.Lock()
			results[searchType] = map[string]interface{}{
				"total": count,
				"items": result,
			}
			total += count
			mu.Unlock()
		}(searchType)
	}

	wg.Wait()

	response := &globalSearchResp{
		Total:   total,
		Results: results,
	}

	c.Response(response)
}

// 搜索聊天消息 - 直接查询本地数据库，处理分表
func (m *Message) searchChatMessages(keyword string, filters struct {
	ChannelID   string `json:"channel_id"`
	ChannelType uint8  `json:"channel_type"`
	ContentType uint8  `json:"content_type"`
	DateRange   struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	} `json:"date_range"`
	FileTypes []string `json:"file_types"`
}, pagination struct {
	Page     uint32 `json:"page"`
	PageSize uint32 `json:"page_size"`
}) (interface{}, int64, error) {
	// 获取所有消息表名
	tableNames := m.getAllMessageTableNames()

	// 构建搜索条件
	whereConditions := []string{"is_deleted = 0"}
	args := []interface{}{}

	// 添加频道过滤
	if filters.ChannelID != "" {
		whereConditions = append(whereConditions, "channel_id = ?")
		args = append(args, filters.ChannelID)
	}
	if filters.ChannelType != 0 {
		whereConditions = append(whereConditions, "channel_type = ?")
		args = append(args, filters.ChannelType)
	}

	// 添加时间范围过滤
	if filters.DateRange.StartTime != "" {
		startTime, err := time.Parse("2006-01-02 15:04:05", filters.DateRange.StartTime)
		if err == nil {
			whereConditions = append(whereConditions, "timestamp >= ?")
			args = append(args, startTime.Unix())
		}
	}
	if filters.DateRange.EndTime != "" {
		endTime, err := time.Parse("2006-01-02 15:04:05", filters.DateRange.EndTime)
		if err == nil {
			whereConditions = append(whereConditions, "timestamp <= ?")
			args = append(args, endTime.Unix())
		}
	}

	// 添加内容类型过滤
	if filters.ContentType != 0 {
		whereConditions = append(whereConditions, "setting = ?")
		args = append(args, filters.ContentType)
	}

	// 构建WHERE子句
	whereClause := strings.Join(whereConditions, " AND ")

	// 分页参数
	offset := (pagination.Page - 1) * pagination.PageSize
	limit := pagination.PageSize

	// 在所有表中搜索
	allResults := make([]*chatSearchResult, 0)
	totalCount := int64(0)

	for _, tableName := range tableNames {
		// 查询总数
		var count int64
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s AND (payload LIKE ? OR header LIKE ?)",
			tableName, whereClause)
		err := m.ctx.DB().QueryRow(countQuery, append(args, "%"+keyword+"%", "%"+keyword+"%")...).Scan(&count)
		if err != nil {
			m.Warn("查询表总数失败", zap.String("table", tableName), zap.Error(err))
			continue
		}
		totalCount += count

		// 查询消息内容
		query := fmt.Sprintf(`
			SELECT message_id, channel_id, channel_type, from_uid, timestamp, payload, header, setting
			FROM %s 
			WHERE %s AND (payload LIKE ? OR header LIKE ?)
			ORDER BY timestamp DESC
			LIMIT ? OFFSET ?
		`, tableName, whereClause)

		rows, err := m.ctx.DB().Query(query, append(args, "%"+keyword+"%", "%"+keyword+"%", limit, offset)...)
		if err != nil {
			m.Warn("查询表数据失败", zap.String("table", tableName), zap.Error(err))
			continue
		}
		defer rows.Close()

		for rows.Next() {
			var messageID, channelID, fromUID, header string
			var channelType, setting uint8
			var timestamp int64
			var payload []byte

			err := rows.Scan(&messageID, &channelID, &channelType, &fromUID, &timestamp, &payload, &header, &setting)
			if err != nil {
				continue
			}

			// 解析payload获取消息内容
			content := m.extractMessageContent(payload, header)

			// 计算相关度分数
			relevanceScore := m.calculateRelevanceScore(keyword, content, header)

			// 获取发送者姓名
			fromName := m.getUserName(fromUID)

			chatResult := &chatSearchResult{
				Type:           "chat",
				MessageID:      messageID,
				ChannelID:      channelID,
				ChannelType:    channelType,
				FromUID:        fromUID,
				FromName:       fromName,
				Content:        content,
				ContentType:    setting,
				Timestamp:      time.Unix(timestamp, 0).Format("2006-01-02 15:04:05"),
				RelevanceScore: relevanceScore,
			}

			// 高亮关键词
			chatResult.Highlight.Content = m.highlightKeyword(keyword, content)

			allResults = append(allResults, chatResult)
		}
	}

	// 按相关度排序
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].RelevanceScore > allResults[j].RelevanceScore
	})

	// 应用分页
	start := int(offset)
	end := start + int(limit)
	if start >= len(allResults) {
		start = len(allResults)
	}
	if end > len(allResults) {
		end = len(allResults)
	}

	if start < end {
		allResults = allResults[start:end]
	} else {
		allResults = []*chatSearchResult{}
	}

	return allResults, totalCount, nil
}

// 搜索用户 - 直接查询本地数据库
func (m *Message) searchUsers(keyword string, pagination struct {
	Page     uint32 `json:"page"`
	PageSize uint32 `json:"page_size"`
}) (interface{}, int64, error) {
	// 构建搜索条件
	whereConditions := []string{"is_deleted = 0"}
	args := []interface{}{}

	// 支持按姓名、用户名、短编号、手机号搜索
	searchConditions := []string{
		"name LIKE ?",
		"username LIKE ?",
		"short_no LIKE ?",
		"phone LIKE ?",
	}

	searchArgs := []interface{}{
		"%" + keyword + "%",
		"%" + keyword + "%",
		"%" + keyword + "%",
		"%" + keyword + "%",
	}

	// 分页参数
	offset := (pagination.Page - 1) * pagination.PageSize
	limit := pagination.PageSize

	// 查询总数
	var totalCount int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM user WHERE %s AND (%s)",
		strings.Join(whereConditions, " AND "), strings.Join(searchConditions, " OR "))

	err := m.ctx.DB().QueryRow(countQuery, append(args, searchArgs...)...).Scan(&totalCount)
	if err != nil {
		m.Error("查询用户总数失败", zap.Error(err))
		return nil, 0, err
	}

	// 查询用户数据
	query := fmt.Sprintf(`
		SELECT uid, name, username, short_no, phone, zone, sex, category, role
		FROM user 
		WHERE %s AND (%s)
		ORDER BY 
			CASE 
				WHEN name LIKE ? THEN 1
				WHEN username LIKE ? THEN 2
				WHEN short_no LIKE ? THEN 3
				WHEN phone LIKE ? THEN 4
				ELSE 5
			END,
			name ASC
		LIMIT ? OFFSET ?
	`, strings.Join(whereConditions, " AND "), strings.Join(searchConditions, " OR "))

	// 添加排序参数和分页参数
	queryArgs := append(args, searchArgs...)
	queryArgs = append(queryArgs, searchArgs...) // 排序条件
	queryArgs = append(queryArgs, limit, offset)

	rows, err := m.ctx.DB().Query(query, queryArgs...)
	if err != nil {
		m.Error("查询用户数据失败", zap.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	userResults := make([]*userSearchResult, 0)
	for rows.Next() {
		var uid, name, username, shortNo, phone, zone, category, role string
		var sex uint8

		err := rows.Scan(&uid, &name, &username, &shortNo, &phone, &zone, &category, &role, &sex)
		if err != nil {
			continue
		}

		// 计算相关度分数
		relevanceScore := m.calculateUserRelevanceScore(keyword, name, username, shortNo, phone)

		userResult := &userSearchResult{
			Type:           "user",
			UID:            uid,
			Name:           name,
			Username:       username,
			ShortNo:        shortNo,
			Avatar:         "", // 暂时为空，后续可以从用户设置表获取
			RelevanceScore: relevanceScore,
		}

		// 高亮关键词
		userResult.Highlight.Name = m.highlightKeyword(keyword, name)

		userResults = append(userResults, userResult)
	}

	return userResults, totalCount, nil
}

// 计算用户相关度分数
func (m *Message) calculateUserRelevanceScore(keyword, name, username, shortNo, phone string) float64 {
	score := 0.0
	keywordLower := strings.ToLower(keyword)

	// 姓名匹配分数
	if strings.Contains(strings.ToLower(name), keywordLower) {
		score += 0.8
		if strings.HasPrefix(strings.ToLower(name), keywordLower) {
			score += 0.2
		}
	}

	// 用户名匹配分数
	if strings.Contains(strings.ToLower(username), keywordLower) {
		score += 0.6
	}

	// 短编号匹配分数
	if strings.Contains(strings.ToLower(shortNo), keywordLower) {
		score += 0.5
	}

	// 手机号匹配分数
	if strings.Contains(strings.ToLower(phone), keywordLower) {
		score += 0.4
	}

	// 完全匹配加分
	if strings.EqualFold(name, keyword) {
		score += 0.3
	}

	return math.Min(score, 1.0)
}

// 搜索群组 - 直接查询本地数据库
func (m *Message) searchGroups(keyword string, pagination struct {
	Page     uint32 `json:"page"`
	PageSize uint32 `json:"page_size"`
}) (interface{}, int64, error) {
	// 构建搜索条件
	whereConditions := []string{"status = 1"} // 只搜索正常状态的群组
	args := []interface{}{}

	// 支持按群名称、群编号、群分类搜索
	searchConditions := []string{
		"name LIKE ?",
		"group_no LIKE ?",
		"category LIKE ?",
	}

	searchArgs := []interface{}{
		"%" + keyword + "%",
		"%" + keyword + "%",
		"%" + keyword + "%",
	}

	// 分页参数
	offset := (pagination.Page - 1) * pagination.PageSize
	limit := pagination.PageSize

	// 查询总数
	var totalCount int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `group` WHERE %s AND (%s)",
		strings.Join(whereConditions, " AND "), strings.Join(searchConditions, " OR "))

	err := m.ctx.DB().QueryRow(countQuery, append(args, searchArgs...)...).Scan(&totalCount)
	if err != nil {
		m.Error("查询群组总数失败", zap.Error(err))
		return nil, 0, err
	}

	// 查询群组数据
	query := fmt.Sprintf(`
		SELECT g.group_no, g.name, g.notice, g.avatar, g.group_type, g.category, g.created_at,
		       COUNT(gm.uid) as member_count
		FROM `+"`group`"+` g
		LEFT JOIN group_member gm ON g.group_no = gm.group_no AND gm.is_deleted = 0
		WHERE %s AND (%s)
		GROUP BY g.group_no, g.name, g.notice, g.avatar, g.group_type, g.category, g.created_at
		ORDER BY 
			CASE 
				WHEN g.name LIKE ? THEN 1
				WHEN g.group_no LIKE ? THEN 2
				WHEN g.category LIKE ? THEN 3
				ELSE 4
			END,
			g.name ASC
		LIMIT ? OFFSET ?
	`, strings.Join(whereConditions, " AND "), strings.Join(searchConditions, " OR "))

	// 添加排序参数和分页参数
	queryArgs := append(args, searchArgs...)
	queryArgs = append(queryArgs, searchArgs...) // 排序条件
	queryArgs = append(queryArgs, limit, offset)

	rows, err := m.ctx.DB().Query(query, queryArgs...)
	if err != nil {
		m.Error("查询群组数据失败", zap.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	groupResults := make([]*groupSearchResult, 0)
	for rows.Next() {
		var groupNo, name, notice, avatar, category string
		var groupType uint8
		var createdAt time.Time
		var memberCount int64

		err := rows.Scan(&groupNo, &name, &notice, &avatar, &groupType, &category, &createdAt, &memberCount)
		if err != nil {
			continue
		}

		// 计算相关度分数
		relevanceScore := m.calculateGroupRelevanceScore(keyword, name, groupNo, category)

		groupResult := &groupSearchResult{
			Type:           "group",
			GroupNo:        groupNo,
			Name:           name,
			Description:    notice,
			MemberCount:    uint32(memberCount),
			RelevanceScore: relevanceScore,
		}

		// 高亮关键词
		groupResult.Highlight.Name = m.highlightKeyword(keyword, name)

		groupResults = append(groupResults, groupResult)
	}

	return groupResults, totalCount, nil
}

// 计算群组相关度分数
func (m *Message) calculateGroupRelevanceScore(keyword, name, groupNo, category string) float64 {
	score := 0.0
	keywordLower := strings.ToLower(keyword)

	// 群名称匹配分数
	if strings.Contains(strings.ToLower(name), keywordLower) {
		score += 0.8
		if strings.HasPrefix(strings.ToLower(name), keywordLower) {
			score += 0.2
		}
	}

	// 群编号匹配分数
	if strings.Contains(strings.ToLower(groupNo), keywordLower) {
		score += 0.6
	}

	// 群分类匹配分数
	if strings.Contains(strings.ToLower(category), keywordLower) {
		score += 0.4
	}

	// 完全匹配加分
	if strings.EqualFold(name, keyword) {
		score += 0.3
	}

	return math.Min(score, 1.0)
}

// 搜索应用 - 直接查询本地数据库
func (m *Message) searchApps(keyword string, pagination struct {
	Page     uint32 `json:"page"`
	PageSize uint32 `json:"page_size"`
}) (interface{}, int64, error) {
	// 构建搜索条件
	whereConditions := []string{"status = 1"} // 只搜索正常状态的应用
	args := []interface{}{}

	// 支持按应用名称、描述、分类搜索
	searchConditions := []string{
		"name LIKE ?",
		"description LIKE ?",
		"app_category LIKE ?",
	}

	searchArgs := []interface{}{
		"%" + keyword + "%",
		"%" + keyword + "%",
		"%" + keyword + "%",
	}

	// 分页参数
	offset := (pagination.Page - 1) * pagination.PageSize
	limit := pagination.PageSize

	// 查询总数
	var totalCount int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM workplace_app WHERE %s AND (%s)",
		strings.Join(whereConditions, " AND "), strings.Join(searchConditions, " OR "))

	err := m.ctx.DB().QueryRow(countQuery, append(args, searchArgs...)...).Scan(&totalCount)
	if err != nil {
		m.Error("查询应用总数失败", zap.Error(err))
		return nil, 0, err
	}

	// 查询应用数据
	query := fmt.Sprintf(`
		SELECT app_id, name, description, icon, app_category, jump_type, app_route, web_route, is_paid_app
		FROM workplace_app 
		WHERE %s AND (%s)
		ORDER BY 
			CASE 
				WHEN name LIKE ? THEN 1
				WHEN description LIKE ? THEN 2
				WHEN app_category LIKE ? THEN 3
				ELSE 4
			END,
			name ASC
		LIMIT ? OFFSET ?
	`, strings.Join(whereConditions, " AND "), strings.Join(searchConditions, " OR "))

	// 添加排序参数和分页参数
	queryArgs := append(args, searchArgs...)
	queryArgs = append(queryArgs, searchArgs...) // 排序条件
	queryArgs = append(queryArgs, limit, offset)

	rows, err := m.ctx.DB().Query(query, queryArgs...)
	if err != nil {
		m.Error("查询应用数据失败", zap.Error(err))
		return nil, 0, err
	}
	defer rows.Close()

	appResults := make([]*appSearchResult, 0)
	for rows.Next() {
		var appID, name, description, icon, appCategory string
		var jumpType, isPaidApp uint8
		var appRoute, webRoute string

		err := rows.Scan(&appID, &name, &description, &icon, &appCategory, &jumpType, &appRoute, &webRoute, &isPaidApp)
		if err != nil {
			continue
		}

		// 计算相关度分数
		relevanceScore := m.calculateAppRelevanceScore(keyword, name, description, appCategory)

		appResult := &appSearchResult{
			Type:           "app",
			AppID:          appID,
			Name:           name,
			Description:    description,
			Icon:           icon,
			RelevanceScore: relevanceScore,
		}

		// 高亮关键词
		appResult.Highlight.Name = m.highlightKeyword(keyword, name)

		appResults = append(appResults, appResult)
	}

	return appResults, totalCount, nil
}

// 计算应用相关度分数
func (m *Message) calculateAppRelevanceScore(keyword, name, description, appCategory string) float64 {
	score := 0.0
	keywordLower := strings.ToLower(keyword)

	// 应用名称匹配分数
	if strings.Contains(strings.ToLower(name), keywordLower) {
		score += 0.8
		if strings.HasPrefix(strings.ToLower(name), keywordLower) {
			score += 0.2
		}
	}

	// 应用描述匹配分数
	if strings.Contains(strings.ToLower(description), keywordLower) {
		score += 0.6
	}

	// 应用分类匹配分数
	if strings.Contains(strings.ToLower(appCategory), keywordLower) {
		score += 0.4
	}

	// 完全匹配加分
	if strings.EqualFold(name, keyword) {
		score += 0.3
	}

	return math.Min(score, 1.0)
}

// 获取所有消息表名
func (m *Message) getAllMessageTableNames() []string {
	tableCount := m.ctx.GetConfig().TablePartitionConfig.MessageTableCount
	tableNames := make([]string, tableCount)

	for i := 0; i < int(tableCount); i++ {
		if i == 0 {
			tableNames[i] = "message"
		} else {
			tableNames[i] = fmt.Sprintf("message%d", i)
		}
	}

	return tableNames
}

// 提取消息内容
func (m *Message) extractMessageContent(payload []byte, header string) string {
	// 尝试解析payload为JSON
	var payloadMap map[string]interface{}
	if err := util.ReadJsonByByte(payload, &payloadMap); err == nil {
		// 如果是文本消息
		if content, ok := payloadMap["content"].(string); ok {
			return content
		}
		// 如果是其他类型消息，返回类型描述
		if msgType, ok := payloadMap["type"].(string); ok {
			return fmt.Sprintf("[%s消息]", msgType)
		}
	}

	// 如果payload解析失败，返回header
	if header != "" {
		return header
	}

	return "[未知消息类型]"
}

// 计算相关度分数
func (m *Message) calculateRelevanceScore(keyword, content, header string) float64 {
	score := 0.0

	// 内容匹配分数
	if strings.Contains(strings.ToLower(content), strings.ToLower(keyword)) {
		score += 0.6
		// 开头匹配加分
		if strings.HasPrefix(strings.ToLower(content), strings.ToLower(keyword)) {
			score += 0.2
		}
	}

	// header匹配分数
	if strings.Contains(strings.ToLower(header), strings.ToLower(keyword)) {
		score += 0.4
	}

	// 完全匹配加分
	if strings.EqualFold(content, keyword) {
		score += 0.3
	}

	return math.Min(score, 1.0)
}

// 高亮关键词
func (m *Message) highlightKeyword(keyword, content string) string {
	if keyword == "" || content == "" {
		return content
	}

	// 使用正则表达式进行大小写不敏感的替换
	re := regexp.MustCompile(fmt.Sprintf("(?i)(%s)", regexp.QuoteMeta(keyword)))
	return re.ReplaceAllString(content, "<em>$1</em>")
}

// 获取用户姓名
func (m *Message) getUserName(uid string) string {
	user, err := m.userService.GetUser(uid)
	if err != nil || user == nil {
		return "未知用户"
	}
	return user.Name
}

// 会话内消息搜索请求参数
type conversationSearchReq struct {
	Keyword     string `json:"keyword"`      // 搜索关键词
	ChannelID   string `json:"channel_id"`   // 会话ID（必填）
	ChannelType uint8  `json:"channel_type"` // 会话类型：1-私聊，2-群聊（必填）
	ContentType uint8  `json:"content_type"` // 内容类型：0-全部，1-文本，2-图片，3-语音，4-视频，5-文件
	DateRange   struct {
		StartTime string `json:"start_time"` // 开始时间（可选）
		EndTime   string `json:"end_time"`   // 结束时间（可选）
	} `json:"date_range"`
	Pagination struct {
		Page     uint32 `json:"page"`      // 页码
		PageSize uint32 `json:"page_size"` // 每页数量
	} `json:"pagination"`
	Sort struct {
		Field string `json:"field"` // 排序字段：time/relevance
		Order string `json:"order"` // 排序方向：asc/desc
	} `json:"sort"`
}

// 会话内消息搜索结果
type conversationSearchResp struct {
	Total   int64                        `json:"total"`
	Results []*conversationMessageResult `json:"results"`
}

// 会话内消息搜索结果项
type conversationMessageResult struct {
	MessageID      string  `json:"message_id"`
	FromUID        string  `json:"from_uid"`
	FromName       string  `json:"from_name"`
	Content        string  `json:"content"`
	ContentType    uint8   `json:"content_type"`
	Timestamp      string  `json:"timestamp"`
	RelevanceScore float64 `json:"relevance_score"`
	Highlight      struct {
		Content string `json:"content"`
	} `json:"highlight"`
	// 消息状态信息
	IsRead    bool `json:"is_read"`    // 是否已读
	IsDeleted bool `json:"is_deleted"` // 是否已删除
	IsRevoked bool `json:"is_revoked"` // 是否已撤回
	// 消息位置信息（便于前端定位）
	Position struct {
		TableName   string `json:"table_name"`   // 所在表名
		TableIndex  int    `json:"table_index"`  // 在表中的索引
		GlobalIndex int64  `json:"global_index"` // 全局消息索引
	} `json:"position"`
}

// 会话内消息搜索
func (m *Message) conversationSearch(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	var req conversationSearchReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseErrorf("数据格式有误！", err)
		return
	}

	// 参数验证
	if req.Keyword == "" {
		c.ResponseError(errors.New("搜索关键词不能为空"))
		return
	}
	if req.ChannelID == "" {
		c.ResponseError(errors.New("会话ID不能为空"))
		return
	}
	if req.ChannelType == 0 {
		c.ResponseError(errors.New("会话类型不能为空"))
		return
	}

	// 验证用户是否有权限访问该会话
	if !m.hasConversationAccess(loginUID, req.ChannelID, req.ChannelType) {
		c.ResponseError(errors.New("无权限访问该会话"))
		return
	}

	// 设置默认值
	if req.Pagination.Page == 0 {
		req.Pagination.Page = 1
	}
	if req.Pagination.PageSize == 0 {
		req.Pagination.PageSize = 20
	}
	if req.Pagination.PageSize > 100 {
		req.Pagination.PageSize = 100
	}
	if req.Sort.Field == "" {
		req.Sort.Field = "time"
	}
	if req.Sort.Order == "" {
		req.Sort.Order = "desc"
	}

	// 执行搜索
	results, total, err := m.searchConversationMessages(req)
	if err != nil {
		m.Error("会话内消息搜索失败", zap.Error(err))
		c.ResponseError(errors.New("搜索失败，请稍后重试"))
		return
	}

	response := &conversationSearchResp{
		Total:   total,
		Results: results,
	}

	c.Response(response)
}

// 搜索会话内消息
func (m *Message) searchConversationMessages(req conversationSearchReq) ([]*conversationMessageResult, int64, error) {
	// 获取消息表名
	tableNames := m.getAllMessageTableNames()

	// 构建搜索条件
	whereConditions := []string{
		"is_deleted = 0",
		"channel_id = ?",
		"channel_type = ?",
	}
	args := []interface{}{
		req.ChannelID,
		req.ChannelType,
	}

	// 添加内容类型过滤
	if req.ContentType != 0 {
		whereConditions = append(whereConditions, "setting = ?")
		args = append(args, req.ContentType)
	}

	// 添加时间范围过滤
	if req.DateRange.StartTime != "" {
		startTime, err := time.Parse("2006-01-02 15:04:05", req.DateRange.StartTime)
		if err == nil {
			whereConditions = append(whereConditions, "timestamp >= ?")
			args = append(args, startTime.Unix())
		}
	}
	if req.DateRange.EndTime != "" {
		endTime, err := time.Parse("2006-01-02 15:04:05", req.DateRange.EndTime)
		if err == nil {
			whereConditions = append(whereConditions, "timestamp <= ?")
			args = append(args, endTime.Unix())
		}
	}

	// 构建WHERE子句
	whereClause := strings.Join(whereConditions, " AND ")

	// 分页参数
	offset := (req.Pagination.Page - 1) * req.Pagination.PageSize
	limit := req.Pagination.PageSize

	// 在所有表中搜索
	allResults := make([]*conversationMessageResult, 0)
	totalCount := int64(0)
	globalIndex := int64(0)

	// 搜索关键词参数
	searchArgs := append(args, "%"+req.Keyword+"%", "%"+req.Keyword+"%")

	for tableIndex, tableName := range tableNames {
		// 查询总数
		var count int64
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s AND (payload LIKE ? OR header LIKE ?)",
			tableName, whereClause)
		err := m.ctx.DB().QueryRow(countQuery, searchArgs...).Scan(&count)
		if err != nil {
			m.Warn("查询表总数失败", zap.String("table", tableName), zap.Error(err))
			continue
		}
		totalCount += count

		// 查询消息内容
		query := fmt.Sprintf(`
			SELECT message_id, from_uid, timestamp, payload, header, setting, is_deleted, is_revoked
			FROM %s 
			WHERE %s AND (payload LIKE ? OR header LIKE ?)
			ORDER BY timestamp DESC
			LIMIT ? OFFSET ?
		`, tableName, whereClause)

		// 添加分页参数
		queryArgs := append(searchArgs, limit, offset)

		rows, err := m.ctx.DB().Query(query, queryArgs...)
		if err != nil {
			m.Warn("查询表数据失败", zap.String("table", tableName), zap.Error(err))
			continue
		}
		defer rows.Close()

		// 处理查询结果
		tableResults := m.processSearchResults(rows, req.Keyword, tableName, tableIndex, globalIndex)
		allResults = append(allResults, tableResults...)
		globalIndex += int64(len(tableResults))
	}

	// 按相关度或时间排序
	m.sortSearchResults(allResults, req.Sort)

	// 应用分页
	allResults = m.applyPagination(allResults, offset, limit)

	return allResults, totalCount, nil
}

// 处理搜索结果
func (m *Message) processSearchResults(rows *sql.Rows, keyword, tableName string, tableIndex int, globalIndex int64) []*conversationMessageResult {
	var results []*conversationMessageResult

	for rows.Next() {
		var messageID, fromUID, header string
		var timestamp int64
		var payload []byte
		var setting uint8
		var isDeleted, isRevoked bool

		err := rows.Scan(&messageID, &fromUID, &timestamp, &payload, &header, &setting, &isDeleted, &isRevoked)
		if err != nil {
			continue
		}

		// 解析payload获取消息内容
		content := m.extractMessageContent(payload, header)

		// 计算相关度分数
		relevanceScore := m.calculateRelevanceScore(keyword, content, header)

		// 获取发送者姓名
		fromName := m.getUserName(fromUID)

		// 检查消息是否已读
		isRead := m.isMessageRead(messageID, fromUID)

		messageResult := &conversationMessageResult{
			MessageID:      messageID,
			FromUID:        fromUID,
			FromName:       fromName,
			Content:        content,
			ContentType:    setting,
			Timestamp:      time.Unix(timestamp, 0).Format("2006-01-02 15:04:05"),
			RelevanceScore: relevanceScore,
			IsRead:         isRead,
			IsDeleted:      isDeleted,
			IsRevoked:      isRevoked,
		}

		// 设置位置信息
		messageResult.Position.TableName = tableName
		messageResult.Position.TableIndex = tableIndex
		messageResult.Position.GlobalIndex = globalIndex

		// 高亮关键词
		messageResult.Highlight.Content = m.highlightKeyword(keyword, content)

		results = append(results, messageResult)
	}

	return results
}

// 排序搜索结果
func (m *Message) sortSearchResults(results []*conversationMessageResult, sortConfig struct {
	Field string `json:"field"`
	Order string `json:"order"`
}) {
	if sortConfig.Field == "relevance" {
		sort.Slice(results, func(i, j int) bool {
			if sortConfig.Order == "asc" {
				return results[i].RelevanceScore < results[j].RelevanceScore
			}
			return results[i].RelevanceScore > results[j].RelevanceScore
		})
	} else {
		// 默认按时间排序
		sort.Slice(results, func(i, j int) bool {
			if sortConfig.Order == "asc" {
				return results[i].Timestamp < results[j].Timestamp
			}
			return results[i].Timestamp > results[j].Timestamp
		})
	}
}

// 应用分页
func (m *Message) applyPagination(results []*conversationMessageResult, offset, limit uint32) []*conversationMessageResult {
	start := int(offset)
	end := start + int(limit)

	if start >= len(results) {
		start = len(results)
	}
	if end > len(results) {
		end = len(results)
	}

	if start < end {
		return results[start:end]
	}

	return []*conversationMessageResult{}
}

// 验证用户是否有权限访问会话
func (m *Message) hasConversationAccess(uid, channelID string, channelType uint8) bool {
	if channelType == 1 {
		// 私聊：检查是否是会话的参与者
		return uid == channelID || m.isUserInConversation(uid, channelID)
	} else if channelType == 2 {
		// 群聊：检查是否是群成员
		return m.isGroupMember(uid, channelID)
	}
	return false
}

// 检查用户是否在会话中
func (m *Message) isUserInConversation(uid, channelID string) bool {
	// 这里可以根据实际业务逻辑实现
	// 比如检查好友关系、会话记录等
	return true // 暂时返回true，实际需要根据业务逻辑判断
}

// 检查用户是否是群成员
func (m *Message) isGroupMember(uid, groupNo string) bool {
	var count int64
	query := "SELECT COUNT(*) FROM group_member WHERE group_no = ? AND uid = ? AND is_deleted = 0"
	err := m.ctx.DB().QueryRow(query, groupNo, uid).Scan(&count)
	return err == nil && count > 0
}

// 检查消息是否已读
func (m *Message) isMessageRead(messageID, fromUID string) bool {
	// 如果消息是自己发送的，直接返回已读
	// 注意：这里需要传入当前用户ID，暂时返回false，实际使用时需要修改函数签名
	// if fromUID == currentUID {
	// 	return true
	// }

	// 查询消息已读状态
	var count int64
	query := "SELECT COUNT(*) FROM member_readed WHERE message_id = ? AND uid = ?"
	err := m.ctx.DB().QueryRow(query, messageID, fromUID).Scan(&count)
	return err == nil && count > 0
}

// 消息定位请求参数
type messageLocateReq struct {
	MessageID   string `json:"message_id"`   // 消息ID（必填）
	ChannelID   string `json:"channel_id"`   // 会话ID（必填）
	ChannelType uint8  `json:"channel_type"` // 会话类型：1-私聊，2-群聊（必填）
}

// 消息位置信息结构体
type messagePosition struct {
	Page       int    `json:"page"`        // 所在页码
	Index      int    `json:"index"`       // 在页面中的索引
	TotalCount int64  `json:"total_count"` // 会话总消息数
	TableName  string `json:"table_name"`  // 所在表名
	TableIndex int    `json:"table_index"` // 在表中的索引
}

// 消息定位响应
type messageLocateResp struct {
	Message  *conversationMessageResult `json:"message"`  // 消息详情
	Position messagePosition            `json:"position"` // 位置信息
}

// 定位消息
func (m *Message) locateMessage(req messageLocateReq) (*conversationMessageResult, messagePosition, error) {
	// 获取消息表名
	tableNames := m.getAllMessageTableNames()

	// 在所有表中查找消息
	for tableIndex, tableName := range tableNames {
		// 查询消息
		query := fmt.Sprintf(`
			SELECT message_id, from_uid, timestamp, payload, header, setting, is_deleted, is_revoked
			FROM %s 
			WHERE message_id = ? AND channel_id = ? AND channel_type = ?
		`, tableName)

		var messageID, fromUID, header string
		var timestamp int64
		var payload []byte
		var setting uint8
		var isDeleted, isRevoked bool

		err := m.ctx.DB().QueryRow(query, req.MessageID, req.ChannelID, req.ChannelType).Scan(
			&messageID, &fromUID, &timestamp, &payload, &header, &setting, &isDeleted, &isRevoked)

		if err != nil {
			if err == sql.ErrNoRows {
				continue // 继续查找下一个表
			}
			return nil, messagePosition{}, err
		}

		// 找到消息，构建结果
		content := m.extractMessageContent(payload, header)
		fromName := m.getUserName(fromUID)
		isRead := m.isMessageRead(messageID, fromUID)

		message := &conversationMessageResult{
			MessageID:      messageID,
			FromUID:        fromUID,
			FromName:       fromName,
			Content:        content,
			ContentType:    setting,
			Timestamp:      time.Unix(timestamp, 0).Format("2006-01-02 15:04:05"),
			RelevanceScore: 1.0, // 定位到的消息相关度为1.0
			IsRead:         isRead,
			IsDeleted:      isDeleted,
			IsRevoked:      isRevoked,
		}

		// 设置位置信息
		message.Position.TableName = tableName
		message.Position.TableIndex = tableIndex
		message.Position.GlobalIndex = 0 // 暂时设为0，实际需要计算

		// 计算消息在会话中的位置
		position := m.calculateMessagePosition(req.ChannelID, req.ChannelType, timestamp, tableName, tableIndex)

		return message, position, nil
	}

	// 未找到消息
	return nil, messagePosition{}, nil
}

// 计算消息在会话中的位置
func (m *Message) calculateMessagePosition(channelID string, channelType uint8, messageTimestamp int64, tableName string, tableIndex int) messagePosition {
	// 计算会话总消息数和消息位置（优化：一次遍历完成两个计算）
	var totalCount, messageIndex int64
	tableNames := m.getAllMessageTableNames()

	for _, table := range tableNames {
		// 计算总消息数
		var count int64
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE channel_id = ? AND channel_type = ? AND is_deleted = 0", table)
		err := m.ctx.DB().QueryRow(countQuery, channelID, channelType).Scan(&count)
		if err == nil {
			totalCount += count
		}

		// 计算消息位置（按时间排序）
		positionQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE channel_id = ? AND channel_type = ? AND is_deleted = 0 AND timestamp > ?", table)
		err = m.ctx.DB().QueryRow(positionQuery, channelID, channelType, messageTimestamp).Scan(&messageIndex)
		if err == nil {
			break
		}
	}

	// 计算页码和索引（假设每页20条）
	pageSize := int64(20)
	page := int(messageIndex/pageSize) + 1
	index := int(messageIndex % pageSize)

	return messagePosition{
		Page:       page,
		Index:      index,
		TotalCount: totalCount,
		TableName:  tableName,
		TableIndex: tableIndex,
	}
}

// 消息定位
func (m *Message) messageLocate(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	var req messageLocateReq
	if err := c.BindJSON(&req); err != nil {
		c.ResponseErrorf("数据格式有误！", err)
		return
	}

	// 参数验证
	if req.MessageID == "" {
		c.ResponseError(errors.New("消息ID不能为空"))
		return
	}
	if req.ChannelID == "" {
		c.ResponseError(errors.New("会话ID不能为空"))
		return
	}
	if req.ChannelType == 0 {
		c.ResponseError(errors.New("会话类型不能为空"))
		return
	}

	// 验证用户是否有权限访问该会话
	if !m.hasConversationAccess(loginUID, req.ChannelID, req.ChannelType) {
		c.ResponseError(errors.New("无权限访问该会话"))
		return
	}

	// 定位消息
	message, position, err := m.locateMessage(req)
	if err != nil {
		m.Error("消息定位失败", zap.Error(err))
		c.ResponseError(errors.New("消息定位失败，请稍后重试"))
		return
	}

	if message == nil {
		c.ResponseError(errors.New("消息不存在或已被删除"))
		return
	}

	response := &messageLocateResp{
		Message:  message,
		Position: position,
	}

	c.Response(response)
}
