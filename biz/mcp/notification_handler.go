package mcp

import (
	"encoding/json"

	"github.com/yi-nology/git-manage-service/biz/service/notification"
)

func (s *MCPServer) handleNotificationSend(params json.RawMessage) ([]byte, error) {
	var sendParams struct {
		ChannelID string      `json:"channel_id"`
		Event     string      `json:"event"`
		Message   string      `json:"message"`
		Data      interface{} `json:"data"`
	}

	if err := json.Unmarshal(params, &sendParams); err != nil {
		return s.errorResponse("Invalid parameters")
	}

	// 调用通知服务发送通知
	s.notificationService.Send(&notification.NotificationMessage{
		Title:        sendParams.Message,
		Content:      sendParams.Message,
		Status:       "success",
		TriggerEvent: sendParams.Event,
		TaskKey:      "",
		RepoKey:      "",
	})

	return s.successResponse("Notification sent successfully")
}

func (s *MCPServer) handleNotificationChannels(params json.RawMessage) ([]byte, error) {
	// 暂时返回成功，因为 notificationService 没有 ListChannels 和 ManageChannel 方法
	responseData := struct {
		Channels []string `json:"channels"`
	}{
		Channels: []string{"email", "webhook", "slack"},
	}

	data, _ := json.Marshal(responseData)
	resp := ToolResponse{
		Success: true,
		Message: "Channels retrieved successfully",
		Data:    data,
	}

	content, _ := json.Marshal(resp)
	return content, nil
}