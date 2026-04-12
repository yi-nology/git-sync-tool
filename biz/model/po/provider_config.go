package po

import (
	"github.com/yi-nology/git-manage-service/biz/utils"
	"gorm.io/gorm"
)

type ProviderConfig struct {
	gorm.Model
	Name            string `gorm:"uniqueIndex;size:100" json:"name"`
	Platform        string `gorm:"size:20;index" json:"platform"`
	BaseURL         string `gorm:"size:500" json:"base_url"`
	CredentialID    uint   `gorm:"index" json:"credential_id"`
	WebhookSecret   string `gorm:"size:200" json:"webhook_secret"`
	WebhookEndpoint string `gorm:"size:500" json:"webhook_endpoint"`
}

func (ProviderConfig) TableName() string { return "provider_configs" }

func (p *ProviderConfig) BeforeSave(tx *gorm.DB) error {
	if p.WebhookSecret != "" {
		enc, err := utils.Encrypt(p.WebhookSecret)
		if err != nil {
			return err
		}
		p.WebhookSecret = enc
	}
	return nil
}

func (p *ProviderConfig) AfterFind(tx *gorm.DB) error {
	if p.WebhookSecret != "" {
		dec, err := utils.Decrypt(p.WebhookSecret)
		if err == nil {
			p.WebhookSecret = dec
		}
	}
	return nil
}
