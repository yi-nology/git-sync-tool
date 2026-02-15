// biz/service/notification/wechat_sender.go - 企业微信机器人发送器

package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

// WeChatSender 企业微信发送器
type WeChatSender struct {
	config *po.WeChatConfig
}

// NewWeChatSender 创建企业微信发送器
func NewWeChatSender(config *po.WeChatConfig) *WeChatSender {
	return &WeChatSender{config: config}
}

// WeChatMessage 企业微信消息
type WeChatMessage struct {
	MsgType  string         `json:"msgtype"`
	Markdown WeChatMarkdown `json:"markdown"`
}

// WeChatMarkdown 企业微信Markdown消息
type WeChatMarkdown struct {
	Content string `json:"content"`
}

// Send 发送企业微信消息
func (s *WeChatSender) Send(msg *NotificationMessage) error {
	// 构建消息
	statusColor := "info"
	statusEmoji := "✅"
	if msg.Status == "failure" {
		statusColor = "warning"
		statusEmoji = "❌"
	}

	content := fmt.Sprintf("### %s %s\n%s\n> <font color=\"%s\">Task:</font> %s\n> <font color=\"%s\">Repo:</font> %s",
		statusEmoji, msg.Title, msg.Content, statusColor, msg.TaskKey, statusColor, msg.RepoKey)

	wechatMsg := WeChatMessage{
		MsgType: "markdown",
		Markdown: WeChatMarkdown{
			Content: content,
		},
	}

	body, err := json.Marshal(wechatMsg)
	if err != nil {
		return err
	}

	resp, err := http.Post(s.config.WebhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("wechat api error: %s", string(respBody))
	}

	// 检查响应
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
		if result.ErrCode != 0 {
			return fmt.Errorf("wechat error: %s", result.ErrMsg)
		}
	}

	return nil
}
