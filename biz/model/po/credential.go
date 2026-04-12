package po

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/utils"
	"gorm.io/gorm"
)

// Credential 凭证模型 - 统一管理 SSH 密钥和 HTTP 认证信息
type Credential struct {
	gorm.Model
	Name          string     `gorm:"uniqueIndex;size:100" json:"name"` // 凭证名称
	Type          string     `gorm:"size:20;index" json:"type"`        // ssh_key, http_basic, http_token
	Description   string     `gorm:"size:500" json:"description"`      // 描述信息
	SSHKeyID      uint       `gorm:"index" json:"ssh_key_id"`          // 关联 ssh_keys 表 (当 Type=ssh_key, 数据库密钥)
	SSHKeyPath    string     `gorm:"size:500" json:"ssh_key_path"`     // 本地 SSH 文件路径 (当 Type=ssh_key, 本地密钥)
	Username      string     `gorm:"size:200" json:"username"`         // HTTP 用户名
	Secret        string     `gorm:"type:text" json:"-"`               // 加密存储: password/token/passphrase
	URLPattern    string     `gorm:"size:200" json:"url_pattern"`      // URL 自动匹配模式 如 *.github.com
	LastUsedAt    *time.Time `json:"last_used_at"`
	Platform      string     `gorm:"size:20" json:"platform"`
	PlatformScope string     `gorm:"size:200" json:"platform_scope"`
}

func (Credential) TableName() string {
	return "credentials"
}

// BeforeSave 保存前加密敏感字段
func (c *Credential) BeforeSave(tx *gorm.DB) (err error) {
	if c.Secret != "" {
		enc, err := utils.Encrypt(c.Secret)
		if err != nil {
			return err
		}
		c.Secret = enc
	}
	return nil
}

// AfterFind 查询后解密敏感字段
func (c *Credential) AfterFind(tx *gorm.DB) (err error) {
	if c.Secret != "" {
		dec, err := utils.Decrypt(c.Secret)
		if err == nil {
			c.Secret = dec
		}
	}
	return nil
}
