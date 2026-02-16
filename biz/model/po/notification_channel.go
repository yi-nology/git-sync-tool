// biz/model/po/notification_channel.go - 通知渠道PO

package po

import (
	"gorm.io/gorm"
)

// NotificationChannel 通知渠道
type NotificationChannel struct {
	gorm.Model
	Name            string `gorm:"size:100" json:"name"`
	Type            string `gorm:"size:50;index" json:"type"` // email, dingtalk, wechat, webhook, lanxin, feishu
	Config          string `gorm:"type:text" json:"config"`   // JSON配置
	Enabled         bool   `json:"enabled"`
	NotifyOnSuccess bool   `json:"notify_on_success"`                 // 向后兼容
	NotifyOnFailure bool   `json:"notify_on_failure"`                 // 向后兼容
	TriggerEvents   string `gorm:"type:text" json:"trigger_events"`   // JSON数组，触发事件列表
	TitleTemplate   string `gorm:"type:text" json:"title_template"`   // 自定义标题模板
	ContentTemplate string `gorm:"type:text" json:"content_template"` // 自定义内容模板
}

// TriggerEvent 触发事件类型
const (
	TriggerSyncSuccess     = "sync_success"     // 同步成功
	TriggerSyncFailure     = "sync_failure"     // 同步失败
	TriggerSyncConflict    = "sync_conflict"    // 同步冲突
	TriggerWebhookReceived = "webhook_received" // Webhook 接收
	TriggerWebhookError    = "webhook_error"    // Webhook 处理错误
	TriggerCronTriggered   = "cron_triggered"   // 定时任务触发
	TriggerBackupSuccess   = "backup_success"   // 备份成功
	TriggerBackupFailure   = "backup_failure"   // 备份失败
)

func (NotificationChannel) TableName() string {
	return "notification_channels"
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost    string   `json:"smtp_host"`
	SMTPPort    int      `json:"smtp_port"`
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	FromAddress string   `json:"from_address"`
	FromName    string   `json:"from_name"`
	ToAddresses []string `json:"to_addresses"`
	UseTLS      bool     `json:"use_tls"`
}

// DingTalkConfig 钉钉机器人配置
type DingTalkConfig struct {
	WebhookURL   string `json:"webhook_url"`
	SecurityType string `json:"security_type"` // none, sign, keyword
	Secret       string `json:"secret"`        // 签名密钥（sign模式）
	Keywords     string `json:"keywords"`      // 关键字（keyword模式）
}

// WeChatConfig 企业微信机器人配置
type WeChatConfig struct {
	WebhookURL string `json:"webhook_url"`
}

// WebhookConfig 自定义Webhook配置
type WebhookConfig struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"` // POST, PUT
	Headers     map[string]string `json:"headers"`
	ContentType string            `json:"content_type"` // application/json, application/x-www-form-urlencoded
}

// LanxinConfig 蓝信机器人配置
type LanxinConfig struct {
	WebhookURL   string `json:"webhook_url"`
	SecurityType string `json:"security_type"` // none, sign, keyword
	Sign         string `json:"sign"`          // 签名密钥（sign模式）
	Keywords     string `json:"keywords"`      // 关键字（keyword模式）
}

// FeishuConfig 飞书机器人配置
type FeishuConfig struct {
	WebhookURL   string `json:"webhook_url"`
	SecurityType string `json:"security_type"` // none, sign, keyword
	Secret       string `json:"secret"`        // 签名密钥（sign模式）
	Keywords     string `json:"keywords"`      // 关键字（keyword模式）
}
