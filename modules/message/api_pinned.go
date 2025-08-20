package message

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/common"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/util"
	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/pkg/wkhttp"
	"github.com/gocraft/dbr/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// 封装悟空IM服务调用
type wukongIMService struct {
	client  *http.Client
	baseURL string
}

var wukongIMInstance *wukongIMService

func getWukongIMService(ctx *config.Context) *wukongIMService {
	if wukongIMInstance == nil {
		wukongIMInstance = &wukongIMService{
			client:  &http.Client{},
			baseURL: ctx.GetConfig().WuKongIM.APIURL, // ✅ 从配置中读取
		}
	}
	return wukongIMInstance
}

// 查询单条消息详情
func (w *wukongIMService) querySingleMessage(loginUID, channelID string, channelType uint8, messageID string) (*config.MessageResp, error) {
	messageIDInt, err := strconv.ParseInt(messageID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("消息ID格式错误: %w", err)
	}

	imReqBody := map[string]interface{}{
		"login_uid":    loginUID,
		"channel_id":   channelID,
		"channel_type": int(channelType),
		"message_id":   messageIDInt,
	}

	jsonData, err := json.Marshal(imReqBody)
	if err != nil {
		return nil, fmt.Errorf("请求参数序列化失败: %w", err)
	}

	httpReq, err := http.NewRequest("POST", w.baseURL+"/message", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("悟空IM接口调用失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var errorResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResult)
		return nil, fmt.Errorf("悟空IM接口返回错误状态: %d, %v", resp.StatusCode, errorResult)
	}

	var imResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&imResult); err != nil {
		return nil, fmt.Errorf("解析悟空IM返回结果失败: %w", err)
	}

	// 构造MessageResp对象
	messageResp := &config.MessageResp{
		MessageID:   int64(imResult["message_id"].(float64)),
		FromUID:     imResult["from_uid"].(string),
		ChannelID:   imResult["channel_id"].(string),
		ChannelType: uint8(imResult["channel_type"].(float64)),
		MessageSeq:  uint32(imResult["message_seq"].(float64)),
		Payload:     []byte(imResult["payload"].(string)),
	}

	return messageResp, nil
}

// 检查对方消息在悟空IM中的状态
func (w *wukongIMService) checkOtherUserMessageStatus(targetUID string, channelType uint8, messageID string) (bool, error) {
	// 修复：message_id 应该是 int64 类型
	messageIDInt, err := strconv.ParseInt(messageID, 10, 64)
	if err != nil {
		return false, fmt.Errorf("解析message_id失败: %v", err)
	}

	imReqBody := map[string]interface{}{
		"message_id":   messageIDInt,     // 🔧 修复：转换为int64
		"channel_id":   targetUID,        // ✅ 正确
		"channel_type": int(channelType), // ✅ 正确
		"login_uid":    targetUID,        // ✅ 正确
	}

	jsonData, err := json.Marshal(imReqBody)
	if err != nil {
		return false, fmt.Errorf("请求参数序列化失败: %w", err)
	}

	// 打印请求参数
	fmt.Printf("【悟空IM】检查对方消息状态请求 - targetUID: %s, channelType: %d, messageID: %s, requestBody: %s\n",
		targetUID, channelType, messageID, string(jsonData))

	httpReq, err := http.NewRequest("POST", w.baseURL+"/message", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(httpReq)
	if err != nil {
		return false, fmt.Errorf("悟空IM接口调用失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("读取响应体失败: %w", err)
	}

	// 打印完整响应信息
	fmt.Printf("【悟空IM】检查对方消息状态响应 - targetUID: %s, messageID: %s, statusCode: %d, responseBody: %s\n",
		targetUID, messageID, resp.StatusCode, string(respBody))

	// 如果返回404或消息不存在，说明对方没有这条消息
	if resp.StatusCode == 404 {
		fmt.Printf("【悟空IM】对方没有这条消息（404） - targetUID: %s, messageID: %s\n", targetUID, messageID)
		return false, nil
	}

	if resp.StatusCode != 200 {
		var errorResult map[string]interface{}
		json.Unmarshal(respBody, &errorResult)
		fmt.Printf("【悟空IM】接口返回错误状态 - targetUID: %s, messageID: %s, statusCode: %d, errorResult: %v\n",
			targetUID, messageID, resp.StatusCode, errorResult)
		return false, fmt.Errorf("悟空IM接口返回错误状态: %d, %v", resp.StatusCode, errorResult)
	}

	// 200状态码，解析响应内容
	var messageResp config.MessageResp
	if err := json.Unmarshal(respBody, &messageResp); err != nil {
		fmt.Printf("【悟空IM】解析响应JSON失败 - targetUID: %s, messageID: %s, error: %v, responseBody: %s\n",
			targetUID, messageID, err, string(respBody))
		return false, fmt.Errorf("解析响应JSON失败: %w", err)
	}

	// 检查是否有实际的消息数据
	hasMessage := messageResp.MessageID != 0 // 检查MessageID字段是否存在且不为0
	fmt.Printf("【悟空IM】检查对方消息状态结果 - targetUID: %s, messageID: %s, hasMessage: %t, messageResp: %+v\n",
		targetUID, messageID, hasMessage, messageResp)

	if hasMessage {
		return true, nil // 有消息数据
	} else {
		return false, nil // 没有消息数据
	}
}

// 置顶或取消置顶消息
func (m *Message) pinnedMessage(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	loginName := c.GetLoginName()
	type reqVO struct {
		MessageID   string `json:"message_id"`   // 消息唯一ID
		MessageSeq  uint32 `json:"message_seq"`  // 消息序列号
		ChannelID   string `json:"channel_id"`   // 频道唯一ID
		ChannelType uint8  `json:"channel_type"` // 频道类型
	}
	var req reqVO
	if err := c.BindJSON(&req); err != nil {
		m.Error(common.ErrData.Error(), zap.Error(err))
		c.ResponseError(common.ErrData)
		return
	}
	if req.ChannelID == "" {
		c.ResponseError(errors.New("频道ID不能为空"))
		return
	}
	if req.MessageID == "" {
		c.ResponseError(errors.New("消息ID不能为空"))
		return
	}
	if req.MessageSeq <= 0 {
		c.ResponseError(errors.New("消息seq不合法"))
		return
	}

	var messageFromIM *config.MessageResp // 从悟空IM获取的消息

	// 群聊场景：权限检查和消息查询
	if req.ChannelType == common.ChannelTypeGroup.Uint8() {
		groupInfo, err := m.groupService.GetGroupDetail(req.ChannelID, loginUID)
		if err != nil {
			m.Error("查询群组信息错误", zap.Error(err))
			c.ResponseError(errors.New("查询群组信息错误"))
			return
		}
		if groupInfo == nil || groupInfo.Status != 1 {
			c.ResponseError(errors.New("群不存在或已删除"))
			return
		}
		isCreatorOrManager, err := m.groupService.IsCreatorOrManager(req.ChannelID, loginUID)
		if err != nil {
			m.Error("查询用户在群内权限错误", zap.Error(err))
			c.ResponseError(errors.New("查询用户在群内权限错误"))
			return
		}
		if !isCreatorOrManager && groupInfo.AllowMemberPinnedMessage == 0 {
			c.ResponseError(errors.New("普通成员不允许置顶消息"))
			return
		}
	}

	messageExtra, err := m.messageExtraDB.queryWithMessageID(req.MessageID)
	if err != nil {
		m.Error("查询消息扩展信息错误", zap.Error(err))
		c.ResponseError(errors.New("查询消息扩展信息错误"))
		return
	}
	if messageExtra != nil && messageExtra.IsDeleted == 1 {
		c.ResponseError(errors.New("该消息不存在或已删除"))
		return
	}

	// 通过悟空IM获取消息详情，确定对话对象
	var targetUID string
	var originalChannelID string
	var currentChannelID string

	if req.ChannelType == common.ChannelTypePerson.Uint8() {
		// 私聊场景：通过悟空IM获取消息详情
		messageResp, err := getWukongIMService(m.ctx).querySingleMessage(loginUID, req.ChannelID, req.ChannelType, req.MessageID)
		if err != nil {
			m.Error("悟空IM搜索单条消息接口调用失败",
				zap.Error(err),
				zap.String("messageID", req.MessageID),
				zap.String("channelID", req.ChannelID))
			c.ResponseError(errors.New("无法获取消息详情，请稍后重试"))
			return
		}

		messageFromIM = messageResp
		m.Info("悟空IM查询成功",
			zap.String("messageID", req.MessageID),
			zap.String("fromUID", messageFromIM.FromUID),
			zap.String("channelID", messageFromIM.ChannelID),
			zap.Uint8("channelType", messageFromIM.ChannelType))

		// 计算频道ID和目标UID
		if messageFromIM.FromUID == loginUID {
			// 我发出的消息：对端=前端传入的channel_id（对方UID）
			targetUID = req.ChannelID
			originalChannelID = common.GetFakeChannelIDWith(loginUID, targetUID)
			currentChannelID = common.GetFakeChannelIDWith(loginUID, targetUID)
			m.Info("置顶自己的消息",
				zap.String("messageID", req.MessageID),
				zap.String("fromUID", messageFromIM.FromUID),
				zap.String("targetUID", targetUID),
				zap.String("originalChannelID", originalChannelID),
				zap.String("currentChannelID", currentChannelID))
		} else {
			// 对方发给我的消息：对端=from_uid
			targetUID = messageFromIM.FromUID
			originalChannelID = common.GetFakeChannelIDWith(messageFromIM.FromUID, loginUID)
			currentChannelID = common.GetFakeChannelIDWith(loginUID, targetUID)
			m.Info("置顶对方的消息",
				zap.String("messageID", req.MessageID),
				zap.String("fromUID", messageFromIM.FromUID),
				zap.String("targetUID", targetUID),
				zap.String("originalChannelID", originalChannelID),
				zap.String("currentChannelID", currentChannelID))
		}

		m.Info("通过悟空IM计算频道ID完成",
			zap.String("messageID", req.MessageID),
			zap.String("targetUID", targetUID),
			zap.String("originalChannelID", originalChannelID),
			zap.String("currentChannelID", currentChannelID))
	} else {
		// 群聊场景：两个频道ID相同
		targetUID = req.ChannelID
		originalChannelID = req.ChannelID
		currentChannelID = req.ChannelID
	}

	// 检查对方消息状态（私聊场景）
	shouldNotify := true
	if req.ChannelType == common.ChannelTypePerson.Uint8() {
		// 获取对方的UID
		otherUID := targetUID
		m.Debug("检查对方消息状态",
			zap.String("messageID", req.MessageID),
			zap.String("loginUID", loginUID),
			zap.String("otherUID", otherUID),
			zap.Bool("isOwnMessage", otherUID == loginUID))

		if otherUID == loginUID {
			// 如果是自己给自己发消息，不需要检查
			shouldNotify = true
			m.Debug("自己给自己发消息，shouldNotify=true")
		} else {
			// 通过本地数据库检查对方是否已删除这条消息
			otherMessageUserExtras, err := m.messageUserExtraDB.queryWithMessageIDsAndUID([]string{req.MessageID}, otherUID)
			if err != nil {
				m.Warn("查询对方消息状态失败", zap.Error(err), zap.String("messageID", req.MessageID), zap.String("otherUID", otherUID))
				// 查询失败时默认通知，避免影响正常功能
				shouldNotify = true
				m.Debug("查询对方消息状态失败，shouldNotify=true")
			} else {
				// 检查对方是否已删除该消息
				isDeleted := false
				if len(otherMessageUserExtras) > 0 {
					isDeleted = otherMessageUserExtras[0].MessageIsDeleted == 1
				}

				if isDeleted {
					// 对方已删除该消息，不发送通知
					shouldNotify = false
					m.Debug("对方已删除该消息，shouldNotify=false", zap.String("messageID", req.MessageID), zap.String("otherUID", otherUID))
				} else {
					// 对方未删除该消息，可以发送通知
					shouldNotify = true
					m.Debug("对方未删除该消息，shouldNotify=true", zap.String("messageID", req.MessageID), zap.String("otherUID", otherUID))
				}
			}
		}
	}

	m.Debug("最终shouldNotify值",
		zap.String("messageID", req.MessageID),
		zap.Bool("shouldNotify", shouldNotify),
		zap.Uint8("channelType", req.ChannelType))

	appConfig, err := m.commonService.GetAppConfig()
	if err != nil {
		m.Error("查询配置错误", zap.Error(err))
		c.ResponseError(errors.New("查询配置错误"))
		return
	}
	var maxCount = 10
	if appConfig != nil {
		maxCount = appConfig.ChannelPinnedMessageMaxCount
	}

	// 使用当前用户频道ID查询置顶消息数量
	currentCount, err := m.pinnedDB.queryCountWithCurrentChannel(currentChannelID, req.ChannelType)
	if err != nil {
		m.Error("查询当前置顶消息数量错误", zap.Error(err))
		c.ResponseError(errors.New("查询当前置顶消息数量错误"))
		return
	}

	// 使用消息ID和当前用户频道ID查询置顶消息
	pinnedMessage, err := m.pinnedDB.queryWithMessageIDAndCurrentChannel(req.MessageID, currentChannelID)
	if err != nil {
		m.Error("查询置顶消息错误", zap.Error(err))
		c.ResponseError(errors.New("查询置顶消息错误"))
		return
	}

	if currentCount >= int64(maxCount) && (pinnedMessage == nil || pinnedMessage.IsDeleted == 1) {
		c.ResponseError(errors.New("置顶数量已达到上限"))
		return
	}

	tx, err := m.db.session.Begin()
	if err != nil {
		m.Error("开启事务错误", zap.Error(err))
		c.ResponseError(errors.New("开启事务错误"))
		return
	}
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
			panic(err)
		}
	}()
	isPinned := 0
	isSendSystemMsg := false
	if pinnedMessage == nil {
		// 新增置顶消息，私聊场景根据对方状态决定插入记录数量
		if req.ChannelType == common.ChannelTypePerson.Uint8() {
			// 私聊场景：存储当前用户的置顶记录
			currentUserRecord := &pinnedMessageModel{
				MessageId:         req.MessageID,
				ChannelID:         currentChannelID,  // 使用当前用户频道ID，避免唯一约束冲突
				OriginalChannelID: originalChannelID, // 原始频道ID
				CurrentChannelID:  currentChannelID,  // 当前用户频道ID
				ChannelType:       req.ChannelType,
				IsDeleted:         0,
				MessageSeq:        req.MessageSeq,
				Version:           time.Now().UnixMilli(),
			}

			err = m.pinnedDB.insert(currentUserRecord)
			if err != nil {
				tx.Rollback()
				m.Error("新增当前用户置顶消息错误", zap.Error(err), zap.String("messageID", req.MessageID))
				c.ResponseError(errors.New("新增置顶消息错误"))
				return
			}

			// 根据对方消息状态决定是否插入对方的置顶记录
			if shouldNotify {
				// 对方未删除消息，插入对方的置顶记录（用于对方取消置顶）
				otherChannelID := ""
				if messageFromIM.FromUID == loginUID {
					// 置顶自己的消息：使用一个不同的频道ID，避免唯一约束冲突
					// 由于GetFakeChannelIDWith对参数顺序不敏感，我们使用一个后缀来区分
					otherChannelID = common.GetFakeChannelIDWith(targetUID, loginUID) + "_other"
				} else {
					// 置顶对方的消息：对方频道ID
					otherChannelID = common.GetFakeChannelIDWith(messageFromIM.FromUID, targetUID)
				}

				otherUserRecord := &pinnedMessageModel{
					MessageId:         req.MessageID,
					ChannelID:         otherChannelID,    // 使用对方频道ID，避免唯一约束冲突
					OriginalChannelID: originalChannelID, // 原始频道ID
					CurrentChannelID:  otherChannelID,    // 对方频道ID
					ChannelType:       req.ChannelType,
					IsDeleted:         0,
					MessageSeq:        req.MessageSeq,
					Version:           time.Now().UnixMilli(),
				}

				err = m.pinnedDB.insert(otherUserRecord)
				if err != nil {
					tx.Rollback()
					m.Error("新增对方置顶消息错误", zap.Error(err))
					c.ResponseError(errors.New("新增置顶消息错误"))
					return
				}

				m.Info("私聊场景存储两条置顶记录完成（对方未删除消息）",
					zap.String("messageID", req.MessageID),
					zap.String("currentChannelID", currentChannelID),
					zap.String("otherChannelID", otherChannelID))
			} else {
				// 对方已删除消息，只存储当前用户的置顶记录
				m.Info("私聊场景存储一条置顶记录（对方已删除消息）",
					zap.String("messageID", req.MessageID),
					zap.String("currentChannelID", currentChannelID))
			}
		} else {
			// 群聊场景：存储一条记录
			err = m.pinnedDB.insert(&pinnedMessageModel{
				MessageId:         req.MessageID,
				ChannelID:         req.ChannelID,     // 保持兼容性
				OriginalChannelID: originalChannelID, // 原始频道ID
				CurrentChannelID:  currentChannelID,  // 当前用户频道ID
				ChannelType:       req.ChannelType,
				IsDeleted:         0,
				MessageSeq:        req.MessageSeq,
				Version:           time.Now().UnixMilli(),
			})
			if err != nil {
				tx.Rollback()
				m.Error("新增群聊置顶消息错误", zap.Error(err))
				c.ResponseError(errors.New("新增置顶消息错误"))
				return
			}
		}

		isSendSystemMsg = true
		isPinned = 1
	} else {
		// 更新现有置顶消息状态
		if pinnedMessage.IsDeleted == 1 {
			pinnedMessage.IsDeleted = 0
			isPinned = 1
		} else {
			pinnedMessage.IsDeleted = 1
			isPinned = 0
		}
		pinnedMessage.Version = time.Now().UnixMilli()
		err = m.pinnedDB.update(pinnedMessage)
		if err != nil {
			tx.Rollback()
			m.Error("取消置顶消息错误", zap.Error(err))
			c.ResponseError(errors.New("取消置顶消息错误"))
			return
		}
	}
	// 使用原始频道ID更新消息扩展
	version := m.genMessageExtraSeq(originalChannelID)
	err = m.messageExtraDB.insertOrUpdatePinnedTx(&messageExtraModel{
		MessageID:   req.MessageID,
		MessageSeq:  req.MessageSeq,
		ChannelID:   originalChannelID,
		ChannelType: req.ChannelType,
		IsPinned:    isPinned,
		Version:     version,
	}, tx)
	if err != nil {
		tx.Rollback()
		c.ResponseErrorf("更新消息置顶状态失败！", err)
		return
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.ResponseErrorf("事务提交失败！", err)
		return
	}

	// 根据对方消息状态决定是否发送通知
	// 私聊场景：如果是置顶别人的消息，CMD应该发送到对方的频道
	cmdChannelID := req.ChannelID
	if req.ChannelType == common.ChannelTypePerson.Uint8() && targetUID != "" && targetUID != c.GetLoginUID() {
		// 置顶别人的消息，CMD发送到对方频道
		cmdChannelID = targetUID
		m.Info("私聊置顶别人消息，CMD发送到对方频道",
			zap.String("messageID", req.MessageID),
			zap.String("originalChannelID", req.ChannelID),
			zap.String("cmdChannelID", cmdChannelID),
			zap.String("targetUID", targetUID))
	}

	m.Debug("准备发送CMD通知",
		zap.String("messageID", req.MessageID),
		zap.Bool("shouldNotify", shouldNotify),
		zap.String("originalChannelID", req.ChannelID),
		zap.String("cmdChannelID", cmdChannelID),
		zap.Uint8("channelType", req.ChannelType),
		zap.String("fromUID", c.GetLoginUID()))

	if shouldNotify {
		m.Debug("开始发送CMD",
			zap.String("messageID", req.MessageID),
			zap.String("loginUID", c.GetLoginUID()),
			zap.String("originalChannelID", req.ChannelID),
			zap.String("cmdChannelID", cmdChannelID),
			zap.String("targetUID", targetUID),
			zap.String("originalChannelID", originalChannelID),
			zap.String("currentChannelID", currentChannelID))

		err = m.ctx.SendCMD(config.MsgCMDReq{
			NoPersist:   true,
			ChannelID:   cmdChannelID,
			ChannelType: req.ChannelType,
			FromUID:     c.GetLoginUID(),
			CMD:         common.CMDSyncPinnedMessage,
		})

		if err != nil {
			m.Error("发送cmd失败！", zap.Error(err))
			c.ResponseError(err)
			return
		} else {
			m.Debug("CMD发送成功",
				zap.String("messageID", req.MessageID),
				zap.String("channelID", req.ChannelID),
				zap.String("fromUID", c.GetLoginUID()),
				zap.String("cmd", common.CMDSyncPinnedMessage))
		}
	} else {
		m.Debug("对方已删除消息，跳过CMD通知",
			zap.String("messageID", req.MessageID),
			zap.String("channelID", req.ChannelID))
	}

	// 系统消息通知也根据状态决定
	if isSendSystemMsg && shouldNotify {
		var payloadMap map[string]interface{}
		var payload []byte

		// 根据场景选择消息内容来源
		if req.ChannelType == common.ChannelTypePerson.Uint8() && messageFromIM != nil {
			// 私聊场景：使用从悟空IM获取的消息内容
			payload = messageFromIM.Payload
			m.Info("使用悟空IM消息内容生成系统消息", zap.String("messageID", req.MessageID))
		} else if req.ChannelType == common.ChannelTypeGroup.Uint8() {
			// 群聊场景：需要从本地数据库查询消息内容
			// 由于我们删除了message变量，这里需要重新查询
			groupMessage, err := m.db.queryMessageWithMessageID(req.ChannelID, req.MessageID)
			if err != nil || groupMessage == nil {
				m.Warn("无法获取群聊消息内容，跳过系统消息通知", zap.String("messageID", req.MessageID))
				c.ResponseOK()
				return
			}
			payload = groupMessage.Payload
			m.Info("使用本地数据库消息内容生成系统消息", zap.String("messageID", req.MessageID))
		} else {
			m.Warn("无法获取消息内容，跳过系统消息通知", zap.String("messageID", req.MessageID))
			c.ResponseOK()
			return
		}

		// payload 可能是Base64，需要解码后再解析JSON
		decoded, decErr := base64.StdEncoding.DecodeString(string(payload))
		if decErr == nil {
			payload = decoded
		}

		err := util.ReadJsonByByte(payload, &payloadMap)
		if err != nil {
			m.Warn("负荷数据不是json格式！", zap.Error(err), zap.String("payload", string(payload)))
			c.ResponseOK()
			return
		}
		var contentType int = 0
		var content string = ""
		if payloadMap["type"] != nil {
			contentTypeI, _ := payloadMap["type"].(json.Number).Int64()
			contentType = int(contentTypeI)
		}
		if contentType == common.Text.Int() {
			content = payloadMap["content"].(string)
			content = fmt.Sprintf("`%s`", content)
		} else {
			content = common.GetDisplayText(contentType)
		}
		mesageContent := fmt.Sprintf("{0} 置顶了%s", content)
		// 系统消息的发送目标与CMD通知保持一致
		systemMsgChannelID := req.ChannelID
		if req.ChannelType == common.ChannelTypePerson.Uint8() && targetUID != "" && targetUID != c.GetLoginUID() {
			// 置顶别人的消息，系统消息发送到对方频道
			systemMsgChannelID = targetUID
		}

		err = m.ctx.SendMessage(&config.MsgSendReq{
			Header: config.MsgHeader{
				NoPersist: 0,
				RedDot:    1,
				SyncOnce:  0, // 只同步一次
			},
			ChannelID:   systemMsgChannelID,
			ChannelType: req.ChannelType,
			FromUID:     loginUID,
			Payload: []byte(util.ToJson(map[string]interface{}{
				"from_uid":  loginUID,
				"from_name": loginName,
				"content":   mesageContent,
				"extra": []config.UserBaseVo{
					{
						UID:  loginUID,
						Name: loginName,
					},
				},
				"type": common.Tip,
			})),
		})
		if err != nil {
			m.Warn("发送解散群消息错误", zap.Error(err))
		}
	} else if isSendSystemMsg && !shouldNotify {
		m.Info("对方已删除消息，跳过系统消息通知", zap.String("messageID", req.MessageID), zap.String("channelID", req.ChannelID))
	}
	c.ResponseOK()
}

func (m *Message) clearPinnedMessage(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	var req struct {
		ChannelID   string `json:"channel_id"`
		ChannelType uint8  `json:"channel_type"`
	}
	if err := c.BindJSON(&req); err != nil {
		m.Error("数据格式有误！", zap.Error(err))
		c.ResponseError(errors.New("数据格式有误！"))
		return
	}
	if req.ChannelID == "" {
		c.ResponseError(errors.New("频道ID不能为空"))
		return
	}
	// 计算当前用户的频道ID
	currentChannelID := req.ChannelID
	if req.ChannelType == common.ChannelTypePerson.Uint8() {
		currentChannelID = common.GetFakeChannelIDWith(loginUID, req.ChannelID)
	} else {
		// 查询权限
		isCreatorOrManager, err := m.groupService.IsCreatorOrManager(req.ChannelID, loginUID)
		if err != nil {
			m.Error("查询用户在群内权限错误", zap.Error(err))
			c.ResponseError(errors.New("查询用户在群内权限错误"))
			return
		}
		if !isCreatorOrManager {
			c.ResponseError(errors.New("用户无权清空置顶消息"))
			return
		}
	}

	// 使用当前用户频道ID查询置顶消息
	pinnedMsgs, err := m.pinnedDB.queryAllWithCurrentChannel(currentChannelID, req.ChannelType)
	if err != nil {
		m.Error("查询置顶消息错误", zap.Error(err))
		c.ResponseError(errors.New("查询置顶消息错误"))
		return
	}
	messageIds := make([]string, 0)
	if len(pinnedMsgs) <= 0 {
		c.ResponseOK()
		return
	}

	for _, msg := range pinnedMsgs {
		messageIds = append(messageIds, msg.MessageId)
	}
	messageUserExtras, err := m.messageUserExtraDB.queryWithMessageIDsAndUID(messageIds, loginUID)
	if err != nil {
		m.Error("查询用户消息扩展字段失败！", zap.Error(err))
		c.ResponseError(errors.New("查询用户消息扩展字段失败！"))
		return
	}
	channelOffsetM, err := m.channelOffsetDB.queryWithUIDAndChannel(loginUID, currentChannelID, req.ChannelType)
	if err != nil {
		m.Error("查询频道偏移量失败！", zap.Error(err))
		c.ResponseError(errors.New("查询频道偏移量失败！"))
		return
	}
	updateModel := make([]*pinnedMessageModel, 0)
	for _, msg := range pinnedMsgs {
		isAdd := true
		if len(messageUserExtras) > 0 {
			for _, extra := range messageUserExtras {
				if extra.MessageID == msg.MessageId && extra.MessageIsDeleted == 1 {
					isAdd = false
					break
				}
			}
		}
		if channelOffsetM != nil && msg.MessageSeq <= channelOffsetM.MessageSeq {
			isAdd = false
		}
		if isAdd {
			msg.IsDeleted = 1
			msg.Version = time.Now().UnixMilli()
			updateModel = append(updateModel, msg)
		}
	}
	if len(updateModel) == 0 {
		c.ResponseOK()
		return
	}
	tx, err := m.db.session.Begin()
	if err != nil {
		m.Error("开启事务错误", zap.Error(err))
		c.ResponseError(errors.New("开启事务错误"))
		return
	}
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
			panic(err)
		}
	}()
	for _, msg := range updateModel {
		err = m.pinnedDB.updateTx(msg, tx)
		if err != nil {
			tx.Rollback()
			m.Error("删除置顶消息错误", zap.Error(err))
			c.ResponseError(errors.New("删除置顶消息错误"))
			return
		}

		// 使用原始频道ID更新消息扩展
		version := m.genMessageExtraSeq(msg.OriginalChannelID)
		err = m.messageExtraDB.insertOrUpdatePinnedTx(&messageExtraModel{
			MessageID:   msg.MessageId,
			MessageSeq:  msg.MessageSeq,
			ChannelID:   msg.OriginalChannelID,
			ChannelType: req.ChannelType,
			IsPinned:    0,
			Version:     version,
		}, tx)
		if err != nil {
			tx.Rollback()
			m.Error("修改消息扩展置顶状态错误", zap.Error(err))
			c.ResponseErrorf("修改消息扩展置顶状态错误", err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		c.ResponseErrorf("事务提交失败！", err)
		return
	}

	// 检查是否需要发送通知（私聊场景）
	shouldNotify := true
	if req.ChannelType == common.ChannelTypePerson.Uint8() {
		// 获取对方的UID
		otherUID := req.ChannelID
		if otherUID != loginUID {
			// 检查对方是否已删除相关消息
			otherMessageUserExtras, err := m.messageUserExtraDB.queryWithMessageIDsAndUID(messageIds, otherUID)
			if err != nil {
				m.Warn("查询对方消息状态失败", zap.Error(err), zap.Strings("messageIDs", messageIds), zap.String("otherUID", otherUID))
				// 查询失败时默认通知，避免影响正常功能
				shouldNotify = true
			} else {
				// 如果对方已删除所有相关消息，则不发送通知
				deletedCount := 0
				for _, extra := range otherMessageUserExtras {
					if extra.MessageIsDeleted == 1 {
						deletedCount++
					}
				}
				if deletedCount == len(messageIds) {
					shouldNotify = false
					m.Info("对方已删除所有相关消息，不发送清空置顶通知", zap.Strings("messageIDs", messageIds), zap.String("otherUID", otherUID))
				}
			}
		}
	}

	if shouldNotify {
		err = m.ctx.SendCMD(config.MsgCMDReq{
			NoPersist:   true,
			ChannelID:   req.ChannelID,
			ChannelType: req.ChannelType,
			FromUID:     c.GetLoginUID(),
			CMD:         common.CMDSyncPinnedMessage,
		})

		if err != nil {
			m.Error("发送cmd失败！", zap.Error(err))
			c.ResponseError(err)
			return
		}
	} else {
		m.Info("对方已删除相关消息，跳过清空置顶通知", zap.Strings("messageIDs", messageIds), zap.String("channelID", req.ChannelID))
	}
	c.ResponseOK()
}

func (m *Message) syncPinnedMessage(c *wkhttp.Context) {
	loginUID := c.GetLoginUID()
	var req struct {
		Version     int64  `json:"version"`
		ChannelID   string `json:"channel_id"`
		ChannelType uint8  `json:"channel_type"`
	}
	if err := c.BindJSON(&req); err != nil {
		m.Error("数据格式有误！", zap.Error(err))
		c.ResponseError(errors.New("数据格式有误！"))
		return
	}
	if req.ChannelID == "" {
		c.ResponseError(errors.New("频道ID不能为空"))
		return
	}
	// 计算当前用户的频道ID
	currentChannelID := req.ChannelID
	if req.ChannelType == common.ChannelTypePerson.Uint8() {
		currentChannelID = common.GetFakeChannelIDWith(loginUID, req.ChannelID)
	}

	// 使用当前用户频道ID查询置顶消息
	pinnedMsgs, err := m.pinnedDB.queryWithCurrentChannelID(currentChannelID, req.ChannelType, req.Version)
	if err != nil {
		m.Error("查询置顶消息错误", zap.Error(err))
		c.ResponseError(errors.New("查询置顶消息错误"))
		return
	}
	messageSeqs := make([]uint32, 0)
	messageIds := make([]string, 0)
	list := make([]*MsgSyncResp, 0)
	pinnedMessageList := make([]*pinnedMessageResp, 0)
	if len(pinnedMsgs) <= 0 {
		c.Response(map[string]interface{}{
			"pinned_messages": pinnedMessageList,
			"messages":        list,
		})
		return
	}

	for _, msg := range pinnedMsgs {
		messageSeqs = append(messageSeqs, msg.MessageSeq)
		messageIds = append(messageIds, msg.MessageId)
	}

	resp, err := m.ctx.IMGetWithChannelAndSeqs(req.ChannelID, req.ChannelType, loginUID, messageSeqs)
	if err != nil {
		m.Error("查询频道内的消息失败！", zap.Error(err), zap.String("req", util.ToJson(req)))
		c.ResponseError(errors.New("查询频道内的消息失败！"))
		return
	}

	if resp == nil || len(resp.Messages) == 0 {
		c.Response(map[string]interface{}{
			"pinned_messages": pinnedMessageList,
			"messages":        list,
		})
		return
	}
	// 消息全局扩张
	messageExtras, err := m.messageExtraDB.queryWithMessageIDsAndUID(messageIds, loginUID)
	if err != nil {
		m.Error("查询消息扩展字段失败！", zap.Error(err))
		c.ResponseError(errors.New("查询用户消息扩展错误"))
		return
	}
	messageExtraMap := map[string]*messageExtraDetailModel{}
	if len(messageExtras) > 0 {
		for _, messageExtra := range messageExtras {
			messageExtraMap[messageExtra.MessageID] = messageExtra
		}
	}
	// 消息用户扩张
	messageUserExtras, err := m.messageUserExtraDB.queryWithMessageIDsAndUID(messageIds, loginUID)
	if err != nil {
		m.Error("查询用户消息扩展字段失败！", zap.Error(err))
		c.ResponseError(errors.New("查询用户消息扩展字段失败！"))
		return
	}
	messageUserExtraMap := map[string]*messageUserExtraModel{}
	if len(messageUserExtras) > 0 {
		for _, messageUserExtraM := range messageUserExtras {
			messageUserExtraMap[messageUserExtraM.MessageID] = messageUserExtraM
		}
	}
	// 查询消息回应
	messageReaction, err := m.messageReactionDB.queryWithMessageIDs(messageIds)
	if err != nil {
		m.Error("查询消息回应数据错误", zap.Error(err))
		c.ResponseError(errors.New("查询消息回应数据错误"))
		return
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
	channelOffsetM, err := m.channelOffsetDB.queryWithUIDAndChannel(loginUID, currentChannelID, req.ChannelType)
	if err != nil {
		m.Error("查询频道偏移量失败！", zap.Error(err))
		c.ResponseError(errors.New("查询频道偏移量失败！"))
		return
	}
	// 频道偏移
	channelIds := make([]string, 0)
	channelIds = append(channelIds, currentChannelID)
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
	for _, message := range resp.Messages {
		if channelOffsetM != nil && message.MessageSeq <= channelOffsetM.MessageSeq {
			continue
		}
		msgResp := &MsgSyncResp{}
		messageIDStr := strconv.FormatInt(message.MessageID, 10)
		messageExtra := messageExtraMap[messageIDStr]
		messageUserExtra := messageUserExtraMap[messageIDStr]
		msgResp.from(message, loginUID, messageExtra, messageUserExtra, messageReactionMap[messageIDStr], channelOffsetMessageSeq)
		list = append(list, msgResp)
	}

	for _, msg := range pinnedMsgs {
		messageUserExtra := messageUserExtraMap[msg.MessageId]
		if messageUserExtra != nil && messageUserExtra.MessageIsDeleted == 1 {
			msg.IsDeleted = 1
		}
		if channelOffsetM != nil && msg.MessageSeq <= channelOffsetM.MessageSeq {
			msg.IsDeleted = 1
		}

		// 移除根据对方状态修改is_deleted的逻辑
		// 置顶记录的is_deleted应该表示我是否取消置顶，不受对方状态影响

		toChannelID := common.GetToChannelIDWithFakeChannelID(msg.ChannelID, loginUID)
		pinnedMessageList = append(pinnedMessageList, &pinnedMessageResp{
			MessageID:   msg.MessageId,
			MessageSeq:  msg.MessageSeq,
			ChannelID:   toChannelID,
			ChannelType: msg.ChannelType,
			IsDeleted:   msg.IsDeleted,
			Version:     msg.Version,
			CreatedAt:   msg.CreatedAt.String(),
			UpdatedAt:   msg.UpdatedAt.String(),
		})
	}
	c.Response(map[string]interface{}{
		"pinned_messages": pinnedMessageList,
		"messages":        list,
	})
}

type pinnedMessageResp struct {
	MessageID   string `json:"message_id"`
	MessageSeq  uint32 `json:"message_seq"`
	ChannelID   string `json:"channel_id"`
	ChannelType uint8  `json:"channel_type"`
	IsDeleted   int8   `json:"is_deleted"`
	Version     int64  `json:"version"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (m *Message) deletePinnedMessage(channelID string, channelType uint8, messageIds []string, loginUID string, tx *dbr.Tx) error {
	// 计算当前用户的频道ID
	currentChannelID := channelID
	if channelType == common.ChannelTypePerson.Uint8() {
		currentChannelID = common.GetFakeChannelIDWith(loginUID, channelID)
	}

	// 使用当前用户频道ID查询置顶消息
	pinnedMessages, err := m.pinnedDB.queryWithMessageIds(currentChannelID, channelType, messageIds)
	if err != nil {
		m.Error("查询置顶消息错误", zap.Error(err))
		return errors.New("查询置顶消息错误")
	}
	if len(pinnedMessages) == 0 {
		return nil
	}
	for _, pinnedMessage := range pinnedMessages {
		pinnedMessage.IsDeleted = 1
		pinnedMessage.Version = time.Now().UnixMilli()
		err = m.pinnedDB.updateTx(pinnedMessage, tx)
		if err != nil {
			tx.Rollback()
			m.Error("取消置顶消息错误", zap.Error(err))
			return errors.New("取消置顶消息错误")
		}
	}

	// 检查是否需要发送通知（私聊场景）
	shouldNotify := true
	if channelType == common.ChannelTypePerson.Uint8() {
		// 获取对方的UID
		otherUID := channelID
		if otherUID != loginUID {
			// 检查对方是否已删除相关消息
			otherMessageUserExtras, err := m.messageUserExtraDB.queryWithMessageIDsAndUID(messageIds, otherUID)
			if err != nil {
				m.Warn("查询对方消息状态失败", zap.Error(err), zap.Strings("messageIDs", messageIds), zap.String("otherUID", otherUID))
				// 查询失败时默认通知，避免影响正常功能
				shouldNotify = true
			} else {
				// 如果对方已删除所有相关消息，则不发送通知
				deletedCount := 0
				for _, extra := range otherMessageUserExtras {
					if extra.MessageIsDeleted == 1 {
						deletedCount++
					}
				}
				if deletedCount == len(messageIds) {
					shouldNotify = false
					m.Info("对方已删除所有相关消息，不发送取消置顶通知", zap.Strings("messageIDs", messageIds), zap.String("otherUID", otherUID))
				}
			}
		}
	}

	if shouldNotify {
		err = m.ctx.SendCMD(config.MsgCMDReq{
			NoPersist:   true,
			ChannelID:   channelID,
			ChannelType: channelType,
			FromUID:     loginUID,
			CMD:         common.CMDSyncPinnedMessage,
		})

		if err != nil {
			m.Warn("发送cmd失败！", zap.Error(err))
		}
	} else {
		m.Info("对方已删除相关消息，跳过取消置顶通知", zap.Strings("messageIDs", messageIds), zap.String("channelID", channelID))
	}
	return nil
}
