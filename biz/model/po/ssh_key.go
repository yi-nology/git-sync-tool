package po

import (
	"github.com/yi-nology/git-manage-service/biz/utils"
	"gorm.io/gorm"
)

// SSHKey 数据库SSH密钥模型
type SSHKey struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;size:100" json:"name"` // 密钥名称
	Description string `gorm:"size:500" json:"description"`      // 描述信息
	PrivateKey  string `gorm:"type:text" json:"-"`               // 私钥内容(加密存储)
	PublicKey   string `gorm:"type:text" json:"public_key"`      // 公钥内容(明文)
	Passphrase  string `gorm:"size:500" json:"-"`                // 密码短语(加密存储)
	KeyType     string `gorm:"size:20" json:"key_type"`          // 密钥类型: rsa/ed25519/ecdsa
	UserID      uint   `gorm:"index" json:"user_id"`             // 预留用户ID字段
}

func (SSHKey) TableName() string {
	return "ssh_keys"
}

// BeforeSave 保存前加密敏感字段
func (s *SSHKey) BeforeSave(tx *gorm.DB) (err error) {
	// 加密私钥
	if s.PrivateKey != "" {
		enc, err := utils.Encrypt(s.PrivateKey)
		if err != nil {
			return err
		}
		s.PrivateKey = enc
	}

	// 加密密码短语
	if s.Passphrase != "" {
		enc, err := utils.Encrypt(s.Passphrase)
		if err != nil {
			return err
		}
		s.Passphrase = enc
	}

	return nil
}

// AfterFind 查询后解密敏感字段
func (s *SSHKey) AfterFind(tx *gorm.DB) (err error) {
	// 解密私钥
	if s.PrivateKey != "" {
		dec, err := utils.Decrypt(s.PrivateKey)
		if err == nil {
			s.PrivateKey = dec
		}
	}

	// 解密密码短语
	if s.Passphrase != "" {
		dec, err := utils.Decrypt(s.Passphrase)
		if err == nil {
			s.Passphrase = dec
		}
	}

	return nil
}

// HasPassphraseSet 检查是否设置了密码短语
func (s *SSHKey) HasPassphraseSet() bool {
	return s.Passphrase != ""
}
