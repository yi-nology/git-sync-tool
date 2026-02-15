// biz/service/notification/webhook_sender.go - 自定义Webhook发送器

package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

// WebhookSender 自定义Webhook发送器
type WebhookSender struct {
	config *po.WebhookConfig
}

// NewWebhookSender 创建Webhook发送器
func NewWebhookSender(config *po.WebhookConfig) *WebhookSender {
	return &WebhookSender{config: config}
}

// Send 发送Webhook
func (s *WebhookSender) Send(msg *NotificationMessage) error {
	method := s.config.Method
	if method == "" {
		method = "POST"
	}

	contentType := s.config.ContentType
	if contentType == "" {
		contentType = "application/json"
	}

	var body io.Reader

	if contentType == "application/json" {
		data, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(data)
	} else if contentType == "application/x-www-form-urlencoded" {
		form := url.Values{}
		form.Set("title", msg.Title)
		form.Set("content", msg.Content)
		form.Set("status", msg.Status)
		form.Set("task_key", msg.TaskKey)
		form.Set("repo_key", msg.RepoKey)
		body = strings.NewReader(form.Encode())
	} else {
		data, _ := json.Marshal(msg)
		body = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, s.config.URL, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", contentType)

	// 添加自定义headers
	for k, v := range s.config.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}
