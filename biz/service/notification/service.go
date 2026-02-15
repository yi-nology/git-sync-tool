// biz/service/notification/service.go - 通知服务

package notification

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

// NotificationService 通知服务
type NotificationService struct {
	dao *db.NotificationChannelDAO
}

// NewNotificationService 创建通知服务
func NewNotificationService() *NotificationService {
	return &NotificationService{
		dao: db.NewNotificationChannelDAO(),
	}
}

// NotificationMessage 通知消息
type NotificationMessage struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Status  string `json:"status"` // success, failure
	TaskKey string `json:"task_key"`
	RepoKey string `json:"repo_key"`
}

// Sender 发送器接口
type Sender interface {
	Send(msg *NotificationMessage) error
}

// Send 发送通知到所有启用的渠道
func (s *NotificationService) Send(msg *NotificationMessage) {
	channels, err := s.dao.FindEnabled()
	if err != nil {
		log.Printf("[Notification] Failed to get enabled channels: %v", err)
		return
	}

	for _, channel := range channels {
		// 检查是否需要通知
		if msg.Status == "success" && !channel.NotifyOnSuccess {
			continue
		}
		if msg.Status == "failure" && !channel.NotifyOnFailure {
			continue
		}

		// 发送通知
		go func(ch po.NotificationChannel) {
			sender, err := s.createSender(&ch)
			if err != nil {
				log.Printf("[Notification] Failed to create sender for channel %s: %v", ch.Name, err)
				return
			}

			if err := sender.Send(msg); err != nil {
				log.Printf("[Notification] Failed to send to channel %s: %v", ch.Name, err)
			} else {
				log.Printf("[Notification] Sent to channel %s successfully", ch.Name)
			}
		}(channel)
	}
}

// Test 测试通知渠道
func (s *NotificationService) Test(channelID uint, message string) error {
	channel, err := s.dao.FindByID(channelID)
	if err != nil {
		return err
	}

	sender, err := s.createSender(channel)
	if err != nil {
		return err
	}

	msg := &NotificationMessage{
		Title:   "Git Manage Service - 测试通知",
		Content: message,
		Status:  "success",
	}

	return sender.Send(msg)
}

// createSender 根据渠道类型创建发送器
func (s *NotificationService) createSender(channel *po.NotificationChannel) (Sender, error) {
	switch channel.Type {
	case "email":
		var config po.EmailConfig
		if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
			return nil, err
		}
		return NewEmailSender(&config), nil

	case "dingtalk":
		var config po.DingTalkConfig
		if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
			return nil, err
		}
		return NewDingTalkSender(&config), nil

	case "wechat":
		var config po.WeChatConfig
		if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
			return nil, err
		}
		return NewWeChatSender(&config), nil

	case "webhook":
		var config po.WebhookConfig
		if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
			return nil, err
		}
		return NewWebhookSender(&config), nil

	case "lanxin":
		var config po.LanxinConfig
		if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
			return nil, err
		}
		return NewLanxinSender(&config), nil

	case "feishu":
		var config po.FeishuConfig
		if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
			return nil, err
		}
		return NewFeishuSender(&config), nil

	default:
		return nil, fmt.Errorf("unknown channel type: %s", channel.Type)
	}
}

// NotifySvc 全局通知服务实例
var NotifySvc = NewNotificationService()
