package group

import (
	"testing"

	"github.com/TangSengDaoDao/TangSengDaoDaoServerLib/config"
)

func TestWukongIMClient(t *testing.T) {
	// 创建测试配置
	ctx := &config.Context{}

	// 创建悟空IM客户端
	client := NewWukongIMClient(ctx)

	// 测试客户端创建
	if client == nil {
		t.Fatal("悟空IM客户端创建失败")
	}

	// 测试默认baseURL
	if client.baseURL != "http://localhost:8090" {
		t.Errorf("期望默认baseURL为 http://localhost:8090，实际为 %s", client.baseURL)
	}

	// 测试HTTP客户端
	if client.httpClient == nil {
		t.Fatal("HTTP客户端未初始化")
	}

	// 测试超时设置
	if client.httpClient.Timeout != 30*1000000000 { // 30秒转换为纳秒
		t.Errorf("期望超时时间为30秒，实际为 %v", client.httpClient.Timeout)
	}
}

func TestAddSubscriberRequest(t *testing.T) {
	// 测试请求结构体
	req := &AddSubscriberRequest{
		ChannelID:           "test_group_123",
		ChannelType:         2,
		Subscribers:         []string{"user1", "user2"},
		AllowViewHistoryMsg: 0, // 不允许查看历史消息
		Reset:               0,
		TempSubscriber:      0,
	}

	if req.ChannelID != "test_group_123" {
		t.Errorf("期望ChannelID为 test_group_123，实际为 %s", req.ChannelID)
	}

	if req.ChannelType != 2 {
		t.Errorf("期望ChannelType为 2，实际为 %d", req.ChannelType)
	}

	if len(req.Subscribers) != 2 {
		t.Errorf("期望Subscribers长度为 2，实际为 %d", len(req.Subscribers))
	}

	if req.AllowViewHistoryMsg != 0 {
		t.Errorf("期望AllowViewHistoryMsg为 0，实际为 %d", req.AllowViewHistoryMsg)
	}
}
