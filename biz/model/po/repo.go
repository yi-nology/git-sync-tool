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
	AuthType   string `json:"auth_type"`   // ssh, http, none
	AuthKey    string `json:"auth_key"`    // SSH Key Path or Username
	AuthSecret string `json:"auth_secret"` // Passphrase or Password (Encrypted in DB)

	RemoteAuthsJSON string                     `json:"-"`                     // Stored in DB
	RemoteAuths     map[string]domain.AuthInfo `gorm:"-" json:"remote_auths"` // Memory & API
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

	// Handle RemoteAuths
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

	// Handle RemoteAuths
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

	return nil
}
