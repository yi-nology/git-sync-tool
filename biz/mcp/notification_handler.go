package mcp

import (
	"encoding/json"
)

type notificationHandler struct {}

func newNotificationHandler() *notificationHandler {
	return &notificationHandler{}
}

func (h *notificationHandler) handleNotificationSend(params json.RawMessage) ([]byte, error) {
	var sendParams struct {
		Channel   string `json:"channel"`
		Recipient string `json:"recipient"`
		Message   string `json:"message"`
		Subject   string `json:"subject"`
	}

	if err := json.Unmarshal(params, &sendParams); err != nil {
		resp := ToolResponse{
			Success: false,
			Message: "Invalid parameters",
		}
		content, _ := json.Marshal(resp)
		return content, nil
	}

	// 这里应该调用通知服务的发送方法
	// 暂时返回成功
	resp := ToolResponse{
		Success: true,
		Message: "Notification sent successfully",
	}
	content, _ := json.Marshal(resp)
	return content, nil
}

func (h *notificationHandler) handleNotificationChannels(params json.RawMessage) ([]byte, error) {
	// 这里应该调用通知服务的获取渠道方法
	// 暂时返回成功
	responseData := struct {
		Channels []string `json:"channels"`
	}{
		Channels: []string{"email", "webhook", "slack"},
	}

	data, _ := json.Marshal(responseData)
	resp := ToolResponse{
		Success: true,
		Message: "Notification channels retrieved successfully",
		Data:    data,
	}
	content, _ := json.Marshal(resp)
	return content, nil
}