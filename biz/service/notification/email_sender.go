// biz/service/notification/email_sender.go - 邮件发送器

package notification

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

// EmailSender 邮件发送器
type EmailSender struct {
	config *po.EmailConfig
}

// NewEmailSender 创建邮件发送器
func NewEmailSender(config *po.EmailConfig) *EmailSender {
	return &EmailSender{config: config}
}

// Send 发送邮件
func (s *EmailSender) Send(msg *NotificationMessage) error {
	from := s.config.FromAddress
	if s.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromAddress)
	}

	subject := msg.Title
	body := msg.Content

	// 构建邮件内容
	message := fmt.Sprintf("From: %s\r\n", from)
	message += fmt.Sprintf("To: %s\r\n", strings.Join(s.config.ToAddresses, ", "))
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: text/plain; charset=\"UTF-8\"\r\n"
	message += "\r\n"
	message += body

	// 认证
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.SMTPHost)

	// 发送
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	err := smtp.SendMail(addr, auth, s.config.FromAddress, s.config.ToAddresses, []byte(message))

	return err
}
