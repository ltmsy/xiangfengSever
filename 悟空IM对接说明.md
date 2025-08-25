# 悟空IM对接实现说明

## 概述

本文档描述了如何通过HTTP方式直接调用悟空IM接口，实现群组新成员历史消息查看权限控制功能。

## 功能说明

### 1. 新成员历史消息查看权限控制

根据悟空IM的业务端对接文档，悟空IM已经实现了群组历史消息权限控制功能：

- **接口**：`POST /channel/subscriber_add`
- **参数**：`allow_view_history_msg`
  - `0` = 不允许查看历史消息
  - `1` = 允许查看历史消息（默认值）

### 2. 自动消息过滤

悟空IM在 `POST /channel/messagesync` 接口中自动根据用户权限过滤消息：
- 如果设置为 `0`，只返回用户加入群组后的消息
- 如果设置为 `1`，返回所有历史消息

## 代码实现

### 1. 悟空IM客户端结构

```go
// WukongIMClient 悟空IM客户端
type WukongIMClient struct {
    baseURL    string
    httpClient *http.Client
}

// AddSubscriberRequest 添加订阅者请求
type AddSubscriberRequest struct {
    ChannelID           string   `json:"channel_id"`
    ChannelType         int      `json:"channel_type"`
    Subscribers         []string `json:"subscribers"`
    AllowViewHistoryMsg int      `json:"allow_view_history_msg"`
    Reset               int      `json:"reset"`
    TempSubscriber      int      `json:"temp_subscriber"`
}
```

### 2. 客户端创建

```go
func NewWukongIMClient(ctx *config.Context) *WukongIMClient {
    // 从配置中获取悟空IM的base URL
    baseURL := ctx.GetConfig().WuKongIM.APIURL
    if baseURL == "" {
        baseURL = "http://localhost:8090" // 默认悟空IM地址
    }

    return &WukongIMClient{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}
```

### 3. 添加订阅者接口调用

```go
func (w *WukongIMClient) AddSubscriber(req *AddSubscriberRequest) error {
    url := fmt.Sprintf("%s/channel/subscriber_add", w.baseURL)
    
    // 序列化请求数据
    jsonData, err := json.Marshal(req)
    if err != nil {
        return fmt.Errorf("序列化请求数据失败: %w", err)
    }

    // 创建HTTP请求
    httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return fmt.Errorf("创建HTTP请求失败: %w", err)
    }

    // 设置请求头
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("User-Agent", "TangSengDaoDao/1.0")

    // 发送请求
    resp, err := w.httpClient.Do(httpReq)
    if err != nil {
        return fmt.Errorf("发送HTTP请求失败: %w", err)
    }
    defer resp.Body.Close()

    // 读取响应
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("读取响应失败: %w", err)
    }

    // 检查HTTP状态码
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("悟空IM接口返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
    }

    // 解析响应
    var response AddSubscriberResponse
    if err := json.Unmarshal(body, &response); err != nil {
        return fmt.Errorf("解析响应失败: %w", err)
    }

    // 检查业务状态码
    if response.Code != 0 {
        return fmt.Errorf("悟空IM接口返回业务错误: %s", response.Msg)
    }

    return nil
}
```

## 使用方法

### 1. 在群组添加成员时调用

在 `modules/group/api.go` 的 `addMembersTx` 方法中：

```go
// 查询群组信息，获取历史消息查看权限设置
groupInfo, err := g.db.QueryWithGroupNo(groupNo)
if err != nil {
    g.Error("查询群组信息失败！", zap.Error(err))
    return nil, errors.New("查询群组信息失败！")
}

// 创建悟空IM客户端
wukongIMClient := NewWukongIMClient(g.ctx)

// 调用悟空IM的添加订阅者接口，传递群组历史消息查看权限设置
err = wukongIMClient.AddSubscriber(&AddSubscriberRequest{
    ChannelID:           groupNo,
    ChannelType:         2, // 群组类型
    Subscribers:         realMembers,
    AllowViewHistoryMsg: groupInfo.AllowViewHistoryMsg, // 传递群组权限设置
    Reset:               0,
    TempSubscriber:      0,
})
if err != nil {
    g.Error("调用悟空IM添加订阅者失败！", zap.Error(err))
    return nil, fmt.Errorf("调用悟空IM添加订阅者失败: %w", err)
}
```

### 2. 配置悟空IM地址

在 `configs/tsdd.yaml` 中配置悟空IM的API地址：

```yaml
##################### 悟空IM配置 ####################
wukongIM:
  apiURL: "http://112.121.164.130:5001" # 悟空IM的api地址
  managerToken: "" # 悟空IM的管理者token
```

## 工作流程

### 1. 新成员加入群组流程

1. 业务端调用群组添加成员接口
2. 查询群组的 `AllowViewHistoryMsg` 设置
3. 调用悟空IM的 `POST /channel/subscriber_add` 接口
4. 传递 `allow_view_history_msg` 参数
5. 悟空IM存储权限设置

### 2. 消息同步流程

1. 客户端调用悟空IM的 `POST /channel/messagesync` 接口
2. 悟空IM根据用户权限自动过滤消息
3. 返回符合权限的消息列表

## 优势

### 1. 无需本地过滤

- 悟空IM在服务端直接过滤消息
- 避免传输不必要的数据
- 提高性能和用户体验

### 2. 权限一致性

- 权限设置在悟空IM层面统一管理
- 避免本地和悟空IM权限不一致的问题

### 3. 配置灵活

- 每个群组可以独立设置权限
- 支持系统级别的默认配置

## 注意事项

### 1. 错误处理

- 悟空IM接口调用失败时，需要记录详细错误信息
- 考虑重试机制和降级策略

### 2. 性能考虑

- 悟空IM客户端使用连接池和超时设置
- 避免频繁创建和销毁HTTP客户端

### 3. 配置管理

- 确保悟空IM地址配置正确
- 支持环境变量和配置文件两种配置方式

## 测试验证

### 1. 单元测试

运行群组模块的测试：

```bash
cd modules/group
go test -v -run TestWukongIMClient
```

### 2. 集成测试

1. 创建群组并设置 `AllowViewHistoryMsg = 0`
2. 添加新成员
3. 验证新成员只能看到加入后的消息

### 3. 接口测试

使用Postman或curl测试悟空IM接口：

```bash
curl -X POST "http://your-wukongim-address/channel/subscriber_add" \
  -H "Content-Type: application/json" \
  -d '{
    "channel_id": "test_group_123",
    "channel_type": 2,
    "subscribers": ["user1", "user2"],
    "allow_view_history_msg": 0,
    "reset": 0,
    "temp_subscriber": 0
  }'
```

## 总结

通过HTTP方式直接调用悟空IM接口，我们成功实现了群组新成员历史消息查看权限控制功能。这种实现方式具有以下特点：

1. **简单直接**：无需额外的依赖包，直接通过HTTP调用
2. **功能完整**：悟空IM已经实现了完整的权限控制逻辑
3. **性能优秀**：在服务端直接过滤，避免数据传输开销
4. **配置灵活**：支持群组级别和系统级别的权限配置

该实现完全符合悟空IM的业务端对接文档要求，确保了功能的正确性和一致性。
