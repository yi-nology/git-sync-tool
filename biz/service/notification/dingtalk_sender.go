// biz/service/notification/dingtalk_sender.go - 钉钉机器人发送器

package notification

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

// DingTalkSender 钉钉发送器
type DingTalkSender struct {
	config *po.DingTalkConfig
}

// NewDingTalkSender 创建钉钉发送器
func NewDingTalkSender(config *po.DingTalkConfig) *DingTalkSender {
	return &DingTalkSender{config: config}
}

// DingTalkMessage 钉钉消息
type DingTalkMessage struct {
	MsgType  string           `json:"msgtype"`
	Markdown DingTalkMarkdown `json:"markdown"`
}

// DingTalkMarkdown 钉钉Markdown消息
type DingTalkMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// Send 发送钉钉消息
func (s *DingTalkSender) Send(msg *NotificationMessage) error {
	webhookURL := s.config.WebhookURL

	// 确定安全模式（向后兼容：SecurityType为空但Secret非空时按sign处理）
	securityType := s.config.SecurityType
	if securityType == "" && s.config.Secret != "" {
		securityType = "sign"
	}

	// 签名模式：添加URL参数签名
	if securityType == "sign" && s.config.Secret != "" {
		timestamp := time.Now().UnixMilli()
		sign := s.sign(timestamp, s.config.Secret)
		webhookURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhookURL, timestamp, url.QueryEscape(sign))
	}

	// 构建消息
	statusEmoji := "✅"
	if msg.Status == "failure" {
		statusEmoji = "❌"
	}

	text := fmt.Sprintf("### %s %s\n\n%s\n\n> Task: %s\n> Repo: %s",
		statusEmoji, msg.Title, msg.Content, msg.TaskKey, msg.RepoKey)

	// 关键字模式：在消息开头添加关键字
	if securityType == "keyword" && s.config.Keywords != "" {
		text = s.config.Keywords + "\n" + text
	}

	dingMsg := DingTalkMessage{
		MsgType: "markdown",
		Markdown: DingTalkMarkdown{
			Title: msg.Title,
			Text:  text,
		},
	}

	body, err := json.Marshal(dingMsg)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("dingtalk api error: %s", string(respBody))
	}

	return nil
}

// sign 生成签名
func (s *DingTalkSender) sign(timestamp int64, secret string) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
