// biz/service/notification/feishu_sender.go - 飞书机器人发送器

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
	"strconv"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

// FeishuSender 飞书发送器
type FeishuSender struct {
	config *po.FeishuConfig
}

// NewFeishuSender 创建飞书发送器
func NewFeishuSender(config *po.FeishuConfig) *FeishuSender {
	return &FeishuSender{config: config}
}

// FeishuMessage 飞书消息
type FeishuMessage struct {
	Timestamp string          `json:"timestamp,omitempty"`
	Sign      string          `json:"sign,omitempty"`
	MsgType   string          `json:"msg_type"`
	Content   json.RawMessage `json:"content"`
}

// FeishuPostContent 飞书富文本内容
type FeishuPostContent struct {
	Post map[string]FeishuPostLang `json:"post"`
}

// FeishuPostLang 飞书富文本语言内容
type FeishuPostLang struct {
	Title   string              `json:"title"`
	Content [][]FeishuPostEntry `json:"content"`
}

// FeishuPostEntry 飞书富文本条目
type FeishuPostEntry struct {
	Tag  string `json:"tag"`
	Text string `json:"text,omitempty"`
}

// Send 发送飞书消息
func (s *FeishuSender) Send(msg *NotificationMessage) error {
	// 确定安全模式
	securityType := s.config.SecurityType

	// 构建消息内容
	statusEmoji := "✅"
	if msg.Status == "failure" {
		statusEmoji = "❌"
	}

	titleText := fmt.Sprintf("%s %s", statusEmoji, msg.Title)
	contentText := msg.Content
	detailText := fmt.Sprintf("Task: %s | Repo: %s", msg.TaskKey, msg.RepoKey)

	// 关键字模式：在内容前添加关键字
	if securityType == "keyword" && s.config.Keywords != "" {
		contentText = s.config.Keywords + "\n" + contentText
	}

	// 构建飞书 post 富文本
	postContent := FeishuPostContent{
		Post: map[string]FeishuPostLang{
			"zh_cn": {
				Title: titleText,
				Content: [][]FeishuPostEntry{
					{{Tag: "text", Text: contentText}},
					{{Tag: "text", Text: detailText}},
				},
			},
		},
	}

	contentBytes, err := json.Marshal(postContent)
	if err != nil {
		return err
	}

	feishuMsg := FeishuMessage{
		MsgType: "post",
		Content: contentBytes,
	}

	// 签名模式：将 timestamp 和 sign 放入 body
	if securityType == "sign" && s.config.Secret != "" {
		timestamp := time.Now().Unix()
		sign := s.sign(timestamp, s.config.Secret)
		feishuMsg.Timestamp = strconv.FormatInt(timestamp, 10)
		feishuMsg.Sign = sign
	}

	body, err := json.Marshal(feishuMsg)
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
		return fmt.Errorf("feishu api error: %s", string(respBody))
	}

	return nil
}

// sign 生成飞书签名（timestamp + "\n" + secret -> HMAC-SHA256 -> Base64）
func (s *FeishuSender) sign(timestamp int64, secret string) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
