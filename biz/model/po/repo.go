package po

import (
	"encoding/json"

	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/utils"
	"gorm.io/gorm"
)

type Repo struct {
	gorm.Model
	Key        string `gorm:"uniqueIndex" json:"key"`
	Name       string `gorm:"uniqueIndex" json:"name"`
	Path       string `json:"path"`
	RemoteURL  string `json:"remote_url"`
	AuthType   string `json:"auth_type"`   // ssh, http, none (deprecated, kept for compatibility)
	AuthKey    string `json:"auth_key"`    // SSH Key Path or Username (deprecated)
	AuthSecret string `json:"auth_secret"` // Passphrase or Password (Encrypted in DB) (deprecated)

	RemoteAuthsJSON string                     `json:"-"`                     // Stored in DB (deprecated)
	RemoteAuths     map[string]domain.AuthInfo `gorm:"-" json:"remote_auths"` // Memory & API (deprecated)

	// 新凭证池字段
	DefaultCredentialID   uint            `json:"default_credential_id"`       // 默认凭证 ID
	RemoteCredentialsJSON string          `json:"-"`                           // Stored in DB: {"origin": 1, "upstream": 2}
	RemoteCredentials     map[string]uint `gorm:"-" json:"remote_credentials"` // Memory & API: remote name -> credential ID
}

func (Repo) TableName() string {
	return "repos"
}

func (r *Repo) BeforeSave(tx *gorm.DB) (err error) {
	// Encrypt main secret
	if r.AuthSecret != "" {
		enc, err := utils.Encrypt(r.AuthSecret)
		if err != nil {
			return err
		}
		r.AuthSecret = enc
	}

	// Handle RemoteAuths (deprecated, kept for compatibility)
	if r.RemoteAuths != nil {
		// Encrypt secrets in map
		encryptedMap := make(map[string]domain.AuthInfo)
		for k, v := range r.RemoteAuths {
			if v.Secret != "" {
				enc, err := utils.Encrypt(v.Secret)
				if err != nil {
					return err
				}
				v.Secret = enc
			}
			encryptedMap[k] = v
		}
		bytes, err := json.Marshal(encryptedMap)
		if err != nil {
			return err
		}
		r.RemoteAuthsJSON = string(bytes)
	}

	// Handle RemoteCredentials (new)
	if r.RemoteCredentials != nil {
		bytes, err := json.Marshal(r.RemoteCredentials)
		if err != nil {
			return err
		}
		r.RemoteCredentialsJSON = string(bytes)
	}

	return nil
}

func (r *Repo) AfterFind(tx *gorm.DB) (err error) {
	// Decrypt main secret
	if r.AuthSecret != "" {
		dec, err := utils.Decrypt(r.AuthSecret)
		if err == nil {
			r.AuthSecret = dec
		}
	}

	// Handle RemoteAuths (deprecated, kept for compatibility)
	if r.RemoteAuthsJSON != "" {
		var encryptedMap map[string]domain.AuthInfo
		if err := json.Unmarshal([]byte(r.RemoteAuthsJSON), &encryptedMap); err == nil {
			decryptedMap := make(map[string]domain.AuthInfo)
			for k, v := range encryptedMap {
				if v.Secret != "" {
					dec, err := utils.Decrypt(v.Secret)
					if err == nil {
						v.Secret = dec
					}
				}
				decryptedMap[k] = v
			}
			r.RemoteAuths = decryptedMap
		}
	}

	// Handle RemoteCredentials (new)
	if r.RemoteCredentialsJSON != "" {
		var remoteCreds map[string]uint
		if err := json.Unmarshal([]byte(r.RemoteCredentialsJSON), &remoteCreds); err == nil {
			r.RemoteCredentials = remoteCreds
		}
	}

	return nil
}
