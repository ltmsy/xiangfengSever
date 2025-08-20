# 悟空IM消息接口文档

## 📑 目录

- [📋 文档概述](#-文档概述)
- [🚀 快速开始](#-快速开始)
  - [基础信息](#基础信息)
  - [认证方式](#认证方式)
- [📨 消息发送接口](#-消息发送接口)
  - [1. 发送单条消息](#1-发送单条消息)
  - [2. 批量发送消息](#2-批量发送消息)
- [🔍 消息查询接口](#-消息查询接口)
  - [3. 批量查询消息](#3-批量查询消息)
  - [4. 搜索单条消息](#4-搜索单条消息)
- [🔄 消息同步接口](#-消息同步接口)
  - [5. 频道消息同步（推荐）](#5-频道消息同步推荐)
  - [6. 获取频道最大消息序号](#6-获取频道最大消息序号)
- [🏢 管理端消息搜索接口](#-管理端消息搜索接口)
  - [7. 集群消息搜索](#7-集群消息搜索)
- [❌ 已废弃的接口](#-已废弃的接口)
  - [8. 消息同步（已废弃）](#8-消息同步已废弃)
  - [9. 消息同步回执（已废弃）](#9-消息同步回执已废弃)
- [🌊 流消息接口](#-流消息接口)
  - [10. 开启流消息](#10-开启流消息)
  - [11. 关闭流消息](#11-关闭流消息)
- [📊 消息数据结构](#-消息数据结构)
  - [消息响应格式](#消息响应格式)
  - [消息头部格式](#消息头部格式)
  - [频道类型说明](#频道类型说明)
- [💡 最佳实践](#-最佳实践)
  - [1. 消息同步策略](#1-消息同步策略)
  - [2. 消息发送优化](#2-消息发送优化)
  - [3. 错误处理](#3-错误处理)
- [🔧 配置说明](#-配置说明)
  - [Webhook配置](#webhook配置)
  - [集群配置](#集群配置)
- [📚 相关文档](#-相关文档)
- [🆘 技术支持](#-技术支持)
- [📝 更新日志](#-更新日志)

---

## 📋 文档概述

本文档详细介绍了悟空IM（WuKongIM）提供的所有消息相关接口，包括消息发送、查询、同步等功能。悟空IM是一个高性能的分布式即时通讯系统，支持个人聊天、群组聊天、客服聊天等多种场景。

## 🚀 快速开始

### 基础信息
- **服务地址**: `http://localhost:5001`
- **API文档**: `http://localhost:5001/docs`
- **健康检查**: `http://localhost:5001/health`
- **管理后台**: `http://localhost:5300/web`

### 认证方式
悟空IM支持多种认证方式：
- **Token认证**: 通过Header中的Authorization字段传递
- **无认证模式**: 部分接口支持无认证访问
- **Webhook认证**: 通过配置的密钥进行验证

---

## 📨 消息发送接口

### 1. 发送单条消息

**接口地址**: `POST /message/send`

**功能描述**: 向指定频道发送单条消息

**请求参数**:
```json
{
  "header": {
    "no_persist": 0,      // 是否不持久化 (0=持久化, 1=不持久化)
    "red_dot": 1,         // 是否显示红点 (0=不显示, 1=显示)
    "sync_once": 0        // 是否只同步一次 (0=否, 1=是)
  },
  "client_msg_no": "msg_001",     // 客户端消息号（可选）
  "stream_no": "",                // 流消息号（可选）
  "from_uid": "user123",          // 发送者用户ID
  "channel_id": "group123",       // 频道ID
  "channel_type": 2,              // 频道类型 (1=个人, 2=群组)
  "expire": 0,                    // 消息过期时间（秒，0=永不过期）
  "subscribers": [],              // 指定订阅者列表（可选）
  "payload": "SGVsbG8gV29ybGQ=", // Base64编码的消息内容
  "tag_key": "important"          // 消息标签（可选）
}
```

**响应示例**:
```json
{
  "message_id": 123456789,
  "message_seq": 1001,
  "client_msg_no": "msg_001"
}
```

**使用说明**:
- `payload`字段必须进行Base64编码
- `channel_type`为1时表示个人聊天，为2时表示群组聊天
- 个人聊天时，`channel_id`应为对方用户ID
- 群组聊天时，`channel_id`应为群组ID

### 2. 批量发送消息

**接口地址**: `POST /message/sendbatch`

**功能描述**: 批量发送多条消息

**请求参数**:
```json
[
  {
    "header": {"no_persist": 0, "red_dot": 1, "sync_once": 0},
    "from_uid": "user123",
    "channel_id": "group123",
    "channel_type": 2,
    "payload": "SGVsbG8gV29ybGQ="
  },
  {
    "header": {"no_persist": 0, "red_dot": 0, "sync_once": 0},
    "from_uid": "user123",
    "channel_id": "group456",
    "channel_type": 2,
    "payload": "V2VsY29tZQ=="
  }
]
```

**响应示例**:
```json
[
  {
    "message_id": 123456789,
    "message_seq": 1001,
    "client_msg_no": "msg_001"
  },
  {
    "message_id": 123456790,
    "message_seq": 1002,
    "client_msg_no": "msg_002"
  }
]
```

---

## 🔍 消息查询接口

### 3. 批量查询消息

**接口地址**: `POST /messages`

**功能描述**: 根据多种条件批量查询消息

**请求参数**:
```json
{
  "login_uid": "user123",                    // 当前登录用户ID
  "channel_id": "group123",                  // 频道ID
  "channel_type": 2,                         // 频道类型
  "message_seqs": [1001, 1002, 1003],       // 按消息序号查询（可选）
  "message_ids": [123456789, 123456790],     // 按消息ID查询（可选）
  "client_msg_nos": ["msg001", "msg002"]     // 按客户端消息号查询（可选）
}
```

**响应示例**:
```json
[
  {
    "message_id": 123456789,
    "message_seq": 1001,
    "client_msg_no": "msg001",
    "from_uid": "user123",
    "channel_id": "group123",
    "channel_type": 2,
    "target_uid": "group123",
    "timestamp": 1640995200,
    "payload": "SGVsbG8gV29ybGQ="
  }
]
```

**使用说明**:
- 三种查询条件可以组合使用
- 至少需要提供一种查询条件
- 返回的消息按时间倒序排列

### 4. 搜索单条消息

**接口地址**: `POST /message`

**功能描述**: 精确查询单条消息

**请求参数**:
```json
{
  "login_uid": "user123",        // 当前登录用户ID
  "channel_id": "group123",      // 频道ID
  "channel_type": 2,             // 频道类型
  "message_id": 123456789,       // 消息ID（二选一）
  "client_msg_no": "msg001"      // 客户端消息号（二选一）
}
```

**响应示例**:
```json
{
  "message_id": 123456789,
  "message_seq": 1001,
  "client_msg_no": "msg001",
  "from_uid": "user123",
  "channel_id": "group123",
  "channel_type": 2,
  "target_uid": "group123",
  "timestamp": 1640995200,
  "payload": "SGVsbG8gV29ybGQ="
}
```

**使用说明**:
- `message_id`和`client_msg_no`必须提供其中一个
- 个人聊天时必须提供`login_uid`
- 如果消息不存在，返回404状态码

---

## 🔄 消息同步接口

### 5. 频道消息同步（推荐）

**接口地址**: `POST /channel/messagesync`

**功能描述**: 同步指定频道的消息，支持分页和范围查询

**请求参数**:
```json
{
  "login_uid": "user123",           // 当前登录用户ID（必填）
  "channel_id": "group123",         // 频道ID（必填）
  "channel_type": 2,                // 频道类型（必填）
  "start_message_seq": 100,         // 开始消息序号（包含，可选）
  "end_message_seq": 200,           // 结束消息序号（不包含，可选）
  "limit": 50,                      // 每次同步数量限制（最大10000）
  "pull_mode": 0,                   // 拉取模式 (0=向下拉取, 1=向上拉取)
  "stream_v2": 0                    // 是否使用stream_v2 (0=否, 1=是)
}
```

**响应示例**:
```json
{
  "start_message_seq": 100,
  "end_message_seq": 200,
  "more": 1,                        // 是否有更多数据 (0=否, 1=是)
  "messages": [
    {
      "message_id": 123456789,
      "message_seq": 100,
      "client_msg_no": "msg001",
      "from_uid": "user123",
      "channel_id": "group123",
      "channel_type": 2,
      "target_uid": "group123",
      "timestamp": 1640995200,
      "payload": "SGVsbG8gV29ybGQ="
    }
  ]
}
```

**使用说明**:
- **推荐使用**：这是消息同步的主要接口
- 如果不提供`start_message_seq`和`end_message_seq`，则返回最新的消息
- `pull_mode`为0时向下拉取（获取更新的消息），为1时向上拉取（获取更早的消息）
- `more`字段用于判断是否还有更多数据，实现分页加载

### 6. 获取频道最大消息序号

**接口地址**: `GET /channel/max_message_seq`

**功能描述**: 获取指定频道最新的消息序号

**请求参数**:
```
GET /channel/max_message_seq?channel_id=group123&channel_type=2
```

**响应示例**:
```json
{
  "message_seq": 12345
}
```

**使用说明**:
- 用于判断频道是否有新消息
- 客户端可以保存上次同步的最大序号，与新序号比较判断是否需要同步

---

## 🏢 管理端消息搜索接口

### 7. 集群消息搜索

**接口地址**: `GET /cluster/messages`

**功能描述**: 在集群范围内搜索消息，支持多种搜索条件

**请求参数**:
```
GET /cluster/messages?node_id=1&from_uid=user123&channel_id=group123&channel_type=2&limit=20&offset_message_id=0&pre=0&payload=base64内容&message_id=123&client_msg_no=msg001
```

**参数说明**:
- `node_id`: 节点ID（0表示所有节点）
- `from_uid`: 发送者用户ID
- `channel_id`: 频道ID
- `channel_type`: 频道类型
- `limit`: 返回数量限制
- `offset_message_id`: 偏移消息ID
- `offset_message_seq`: 偏移消息序号
- `pre`: 是否向前搜索（1=是，0=否）
- `payload`: Base64编码的消息内容（用于内容搜索）
- `message_id`: 消息ID
- `client_msg_no`: 客户端消息号

**响应示例**:
```json
{
  "data": [
    {
      "message_id": 123456789,
      "message_seq": 1001,
      "client_msg_no": "msg001",
      "from_uid": "user123",
      "channel_id": "group123",
      "channel_type": 2,
      "target_uid": "group123",
      "timestamp": 1640995200,
      "payload": "SGVsbG8gV29ybGQ="
    }
  ],
  "total": 1000
}
```

**使用说明**:
- 支持跨节点搜索
- 支持消息内容模糊搜索（通过payload参数）
- 支持分页和偏移查询
- 返回总消息数量

---

## ❌ 已废弃的接口

### 8. 消息同步（已废弃）

**接口地址**: `POST /message/sync`

**状态**: 已废弃，不推荐使用

**废弃原因**: 后续不提供带存储的命令消息，业务端通过不存储的命令 + 调用业务端接口一样可以实现相同效果

**替代方案**: 使用 `/channel/messagesync` 接口

### 9. 消息同步回执（已废弃）

**接口地址**: `POST /message/syncack`

**状态**: 已废弃，不推荐使用

**废弃原因**: 与 `/message/sync` 配套的废弃接口

---

## 🌊 流消息接口

### 10. 开启流消息

**接口地址**: `POST /stream/open`

**功能描述**: 开启一个流消息会话

**请求参数**:
```json
{
  "header": {"no_persist": 0, "red_dot": 1, "sync_once": 0},
  "client_msg_no": "stream_msg_001",
  "from_uid": "user123",
  "channel_id": "group123",
  "channel_type": 2,
  "payload": "SGVsbG8gV29ybGQ="
}
```

**响应示例**:
```json
{
  "stream_no": "stream_abc123"
}
```

### 11. 关闭流消息

**接口地址**: `POST /stream/close`

**功能描述**: 关闭流消息会话

**请求参数**:
```json
{
  "stream_no": "stream_abc123"
}
```

---

## 📊 消息数据结构

### 消息响应格式

```json
{
  "message_id": 123456789,        // 消息唯一ID（int64）
  "message_seq": 1001,            // 消息序号（uint64）
  "client_msg_no": "msg001",      // 客户端消息号（string）
  "from_uid": "user123",          // 发送者用户ID（string）
  "channel_id": "group123",       // 频道ID（string）
  "channel_type": 2,              // 频道类型（uint8）
  "target_uid": "group123",       // 目标用户/群组ID（string）
  "timestamp": 1640995200,        // 消息时间戳（秒）
  "payload": "SGVsbG8gV29ybGQ=", // Base64编码的消息内容
  "setting": 0,                   // 消息设置（uint8）
  "expire": 0                     // 过期时间（秒）
}
```

### 消息头部格式

```json
{
  "no_persist": 0,    // 是否不持久化 (0=持久化, 1=不持久化)
  "red_dot": 1,       // 是否显示红点 (0=不显示, 1=显示)
  "sync_once": 0      // 是否只同步一次 (0=否, 1=是)
}
```

### 频道类型说明

| 类型值 | 说明 | 描述 |
|--------|------|------|
| 1 | 个人聊天 | 一对一私聊 |
| 2 | 群组聊天 | 群组聊天 |
| 3 | 客服聊天 | 客服系统 |
| 4 | 社区聊天 | 社区论坛 |

### target_uid 字段说明

`target_uid` 字段用于快速识别消息的目标对象，无需再解析 `channel_id` 和 `channel_type` 的组合。

**字段含义**:
- **字段名**: `target_uid`
- **类型**: `string`
- **说明**: 目标用户/群组ID

**使用场景**:

**私聊场景（`channel_type = 1`）**:
- `target_uid` = 对方用户ID
- 例如：用户A向用户B发送消息，`target_uid` 为 "userB"

**群聊场景（`channel_type = 2`）**:
- `target_uid` = 群组ID  
- 例如：在群组 "group123" 中发送消息，`target_uid` 为 "group123"

**业务应用**:
- 快速获取消息接收方信息
- 简化前端显示逻辑
- 减少业务端数据解析工作

---

## 💡 最佳实践

### 1. 消息同步策略

**增量同步**:
```javascript
// 1. 获取频道最大消息序号
const maxSeq = await getChannelMaxMessageSeq(channelId, channelType);

// 2. 与本地保存的序号比较
if (maxSeq > localLastSeq) {
  // 3. 同步新消息
  const messages = await syncChannelMessages(channelId, channelType, localLastSeq, maxSeq);
  // 4. 更新本地序号
  localLastSeq = maxSeq;
}
```

**分页加载**:
```javascript
// 1. 首次加载最新消息
let messages = await syncChannelMessages(channelId, channelType, 0, 0, 50);

// 2. 向上加载更多历史消息
if (messages.more) {
  const moreMessages = await syncChannelMessages(
    channelId, 
    channelType, 
    0, 
    messages.messages[0].message_seq, 
    50, 
    1  // pull_mode = 1 (向上拉取)
  );
}
```

### 2. 消息发送优化

**批量发送**:
```javascript
// 对于需要发送多条消息的场景，使用批量接口
const batchMessages = [
  { from_uid: "user123", channel_id: "group1", channel_type: 2, payload: "Hello" },
  { from_uid: "user123", channel_id: "group2", channel_type: 2, payload: "World" }
];

const results = await sendBatchMessages(batchMessages);
```

**消息去重**:
```javascript
// 使用client_msg_no避免重复发送
const clientMsgNo = generateUniqueId();
const result = await sendMessage({
  client_msg_no: clientMsgNo,
  from_uid: "user123",
  channel_id: "group123",
  channel_type: 2,
  payload: "Hello World"
});
```

### 3. 错误处理

**常见错误码**:
- `400`: 请求参数错误
- `401`: 认证失败
- `404`: 资源不存在
- `500`: 服务器内部错误

**重试策略**:
```javascript
async function sendMessageWithRetry(messageData, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await sendMessage(messageData);
    } catch (error) {
      if (i === maxRetries - 1) throw error;
      if (error.status === 500) {
        await delay(1000 * Math.pow(2, i)); // 指数退避
        continue;
      }
      throw error;
    }
  }
}
```

---

## 🔧 配置说明

### Webhook配置

悟空IM支持通过Webhook推送消息事件：

```yaml
webhook:
  grpcAddr: "localhost:6979"                    # gRPC地址
  httpAddr: "http://127.0.0.1:8080/webhooks"   # HTTP地址
  msgNotifyEventPushInterval: 500ms             # 推送间隔
  msgNotifyEventCountPerPush: 100               # 每次推送数量
  msgNotifyEventRetryMaxCount: 5                # 最大重试次数
  focusEvents:                                   # 关注的事件类型
    - "msg.offline"                             # 离线消息事件
    - "msg.notify"                              # 消息通知事件
    - "user.onlinestatus"                       # 用户在线状态变化事件
```

### 集群配置

```yaml
cluster:
  nodeId: 1                                     # 节点ID (0-1023)
  addr: "0.0.0.0:5100"                         # 节点地址
  initNodes: ["0.0.0.0:5100"]                  # 初始节点列表
  slotCount: 1024                               # 槽数量
  slotReplicaCount: 1                           # 槽副本数量
  channelReplicaCount: 1                        # 频道副本数量
```

---

## 📚 相关文档

- [悟空IM官方文档](https://github.com/WuKongIM/WuKongIM)
- [API接口文档](http://localhost:5001/docs)
- [协议文档](docs/protocol.md)
- [部署指南](README.md)

---

## 🆘 技术支持

如果您在使用过程中遇到问题，可以通过以下方式获取帮助：

1. **查看日志**: 检查服务运行日志
2. **API文档**: 访问 `http://localhost:5001/docs`
3. **健康检查**: 访问 `http://localhost:5001/health`
4. **GitHub Issues**: 提交问题到官方仓库

---

## 📝 更新日志

| 版本 | 日期 | 更新内容 |
|------|------|----------|
| 1.0.0 | 2025-08-18 | 初始版本，包含所有消息接口 |
| 1.0.1 | 2025-08-18 | 添加最佳实践和配置说明 |
| 1.0.2 | 2025-08-18 | 新增 target_uid 字段，优化消息查询体验 |

---

**文档版本**: 1.0.2  
**最后更新**: 2025-08-18  
**维护者**: 悟空IM团队 