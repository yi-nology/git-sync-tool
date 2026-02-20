// biz/service/notification/lanxin_sender.go - 蓝信机器人发送器
// 参考文档：https://developer.lanxin.cn/official/article?article_id=646eda463d4e4adb7039c150

package notification

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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

// lanxinRequest 蓝信 webhook 请求体（官方文档格式）
// 字段说明：
//   - MsgType: 消息类型，支持 text / linkCard / appCard 等
//   - MsgData: 消息数据，根据 MsgType 填充对应结构
//   - Timestamp: 加签模式下的时间戳（秒级）
//   - Sign: 加签模式下的签名
type lanxinRequest struct {
	MsgType   string      `json:"msgType"`
	MsgData   interface{} `json:"msgData"`
	Timestamp string      `json:"timestamp,omitempty"`
	Sign      string      `json:"sign,omitempty"`
}

// lanxinTextData text 类型消息数据
type lanxinTextData struct {
	Text struct {
		Content string `json:"content"`
	} `json:"text"`
}

// lanxinAppCardData appCard 类型消息数据（支持富文本格式）
type lanxinAppCardData struct {
	AppCard struct {
		BodyTitle   string `json:"bodyTitle"`
		BodyContent string `json:"bodyContent,omitempty"`
		Signature   string `json:"signature,omitempty"`
	} `json:"appCard"`
}

// Send 发送蓝信消息
func (s *LanxinSender) Send(msg *NotificationMessage) error {
	webhookURL := s.config.WebhookURL

	// 构建消息内容
	statusEmoji := "✅"
	if msg.Status == "failure" {
		statusEmoji = "❌"
	}

	content := fmt.Sprintf("%s %s\n%s\nTask: %s\nRepo: %s",
		statusEmoji, msg.Title, msg.Content, msg.TaskKey, msg.RepoKey)

	// 关键字模式：在消息中添加关键字
	securityType := s.config.SecurityType
	if securityType == "" && s.config.Sign != "" {
		securityType = "sign"
	}
	if securityType == "keyword" && s.config.Keywords != "" {
		content = s.config.Keywords + "\n" + content
	}

	// 构建请求体
	textData := lanxinTextData{}
	textData.Text.Content = content

	req := &lanxinRequest{
		MsgType: "text",
		MsgData: textData,
	}

	// 加签模式：把 timestamp 和 sign 放在请求体中
	if securityType == "sign" && s.config.Sign != "" {
		ts := fmt.Sprintf("%d", time.Now().Unix())
		req.Timestamp = ts
		req.Sign = s.genSign(ts, s.config.Sign)
	}

	return s.doSend(webhookURL, req)
}

// doSend 执行实际的 HTTP 请求
func (s *LanxinSender) doSend(webhookURL string, req *lanxinRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	log.Printf("[Notification] Lanxin sending (%s): %s", req.MsgType, string(body))

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[Notification] Lanxin API response (status=%d): %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("lanxin api http error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// 检查 API 响应体中的业务错误码（官方文档字段名为 errCode/errMsg）
	var apiResp struct {
		ErrCode int    `json:"errCode"`
		ErrMsg  string `json:"errMsg"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err == nil {
		if apiResp.ErrCode != 0 {
			return fmt.Errorf("lanxin api error (errCode=%d): %s", apiResp.ErrCode, apiResp.ErrMsg)
		}
	}

	return nil
}

// genSign 生成加签签名（官方文档算法：timestamp@secret 作为 HMAC-SHA256 的 key）
func (s *LanxinSender) genSign(timestamp string, secret string) string {
	stringToSign := timestamp + "@" + secret
	h := hmac.New(sha256.New, []byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// sendAppCard 发送 appCard 格式消息（支持富文本，备用方法）
func (s *LanxinSender) sendAppCard(webhookURL string, title, content, signature string) error {
	cardData := lanxinAppCardData{}
	cardData.AppCard.BodyTitle = title
	cardData.AppCard.BodyContent = strings.ReplaceAll(content, "\n", "<br/>")
	cardData.AppCard.Signature = signature

	securityType := s.config.SecurityType
	if securityType == "" && s.config.Sign != "" {
		securityType = "sign"
	}

	req := &lanxinRequest{
		MsgType: "appCard",
		MsgData: cardData,
	}

	if securityType == "sign" && s.config.Sign != "" {
		ts := fmt.Sprintf("%d", time.Now().Unix())
		req.Timestamp = ts
		req.Sign = s.genSign(ts, s.config.Sign)
	}

	return s.doSend(webhookURL, req)
}
