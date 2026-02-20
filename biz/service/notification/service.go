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
	Title        string        `json:"title"`
	Content      string        `json:"content"`
	Status       string        `json:"status"`        // success, failure
	TriggerEvent string        `json:"trigger_event"` // 触发事件类型
	TaskKey      string        `json:"task_key"`
	RepoKey      string        `json:"repo_key"`
	Data         *TemplateData `json:"-"` // 模板渲染数据（可选）
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
		if !s.shouldNotify(&channel, msg) {
			continue
		}

		// 发送通知
		go func(ch po.NotificationChannel) {
			// 渲染模板：为每个渠道使用其自定义模板或默认模板
			renderedMsg := s.renderMessage(&ch, msg)

			sender, err := s.createSender(&ch)
			if err != nil {
				log.Printf("[Notification] Failed to create sender for channel %s: %v", ch.Name, err)
				return
			}

			if err := sender.Send(renderedMsg); err != nil {
				log.Printf("[Notification] Failed to send to channel %s: %v", ch.Name, err)
			} else {
				log.Printf("[Notification] Sent to channel %s successfully", ch.Name)
			}
		}(channel)
	}
}

// renderMessage 为指定渠道渲染消息模板
func (s *NotificationService) renderMessage(channel *po.NotificationChannel, msg *NotificationMessage) *NotificationMessage {
	// 如果没有模板数据，直接返回原始消息
	if msg.Data == nil {
		return msg
	}

	titleTmpl := channel.TitleTemplate
	contentTmpl := channel.ContentTemplate

	// 优先使用事件级模板（留空则回退到渠道级模板）
	if msg.TriggerEvent != "" {
		etDAO := db.NewNotificationEventTemplateDAO()
		et, err := etDAO.FindByChannelAndEvent(channel.ID, msg.TriggerEvent)
		if err == nil && et != nil {
			if et.TitleTemplate != "" {
				titleTmpl = et.TitleTemplate
			}
			if et.ContentTemplate != "" {
				contentTmpl = et.ContentTemplate
			}
		}
	}

	title, content := RenderTitleAndContent(titleTmpl, contentTmpl, msg.Data)

	return &NotificationMessage{
		Title:        title,
		Content:      content,
		Status:       msg.Status,
		TriggerEvent: msg.TriggerEvent,
		TaskKey:      msg.TaskKey,
		RepoKey:      msg.RepoKey,
	}
}

// shouldNotify 检查是否应该发送通知
func (s *NotificationService) shouldNotify(channel *po.NotificationChannel, msg *NotificationMessage) bool {
	// 优先使用 TriggerEvents 配置
	if channel.TriggerEvents != "" {
		var events []string
		if err := json.Unmarshal([]byte(channel.TriggerEvents), &events); err == nil && len(events) > 0 {
			// 如果消息指定了触发事件类型，检查是否匹配
			if msg.TriggerEvent != "" {
				for _, event := range events {
					if event == msg.TriggerEvent {
						return true
					}
				}
				return false
			}
			// 向后兼容：根据 status 推断触发事件
			for _, event := range events {
				if msg.Status == "success" && (event == po.TriggerSyncSuccess || event == po.TriggerBackupSuccess) {
					return true
				}
				if msg.Status == "failure" && (event == po.TriggerSyncFailure || event == po.TriggerBackupFailure) {
					return true
				}
			}
			return false
		}
	}

	// 向后兼容：使用旧的 NotifyOnSuccess/NotifyOnFailure 配置
	if msg.Status == "success" && !channel.NotifyOnSuccess {
		return false
	}
	if msg.Status == "failure" && !channel.NotifyOnFailure {
		return false
	}
	return true
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
