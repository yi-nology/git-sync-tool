// biz/service/notification/lanxin_sender.go - 蓝信机器人发送器

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

// LanxinSender 蓝信发送器
type LanxinSender struct {
	config *po.LanxinConfig
}

// NewLanxinSender 创建蓝信发送器
func NewLanxinSender(config *po.LanxinConfig) *LanxinSender {
	return &LanxinSender{config: config}
}

// LanxinMessage 蓝信消息
type LanxinMessage struct {
	MsgType  string         `json:"msgtype"`
	Markdown LanxinMarkdown `json:"markdown"`
}

// LanxinMarkdown 蓝信Markdown消息
type LanxinMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// Send 发送蓝信消息
func (s *LanxinSender) Send(msg *NotificationMessage) error {
	webhookURL := s.config.WebhookURL

	// 确定安全模式（向后兼容：SecurityType为空但Sign非空时按sign处理）
	securityType := s.config.SecurityType
	if securityType == "" && s.config.Sign != "" {
		securityType = "sign"
	}

	// 签名模式：添加URL参数签名
	if securityType == "sign" && s.config.Sign != "" {
		timestamp := time.Now().UnixMilli()
		sign := s.sign(timestamp, s.config.Sign)
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

	lanxinMsg := LanxinMessage{
		MsgType: "markdown",
		Markdown: LanxinMarkdown{
			Title: msg.Title,
			Text:  text,
		},
	}

	body, err := json.Marshal(lanxinMsg)
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
		return fmt.Errorf("lanxin api error: %s", string(respBody))
	}

	return nil
}

// sign 生成签名
func (s *LanxinSender) sign(timestamp int64, secret string) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
