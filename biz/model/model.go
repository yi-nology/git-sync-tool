package model

import (
	"encoding/json"
	"time"

	"github.com/yi-nology/git-manage-service/biz/utils"
	"gorm.io/gorm"
)

type AuthInfo struct {
	Type   string `json:"type"`   // ssh, http, none
	Key    string `json:"key"`    // SSH Key Path or Username
	Secret string `json:"secret"` // Passphrase or Password (Encrypted in DB)
}

type Repo struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Key          string `gorm:"uniqueIndex" json:"key"`
	Name         string `gorm:"uniqueIndex" json:"name"`
	Path         string `json:"path"`
	RemoteURL    string `json:"remote_url"`
	AuthType     string `json:"auth_type"`     // ssh, http, none
	AuthKey      string `json:"auth_key"`      // SSH Key Path or Username
	AuthSecret   string `json:"auth_secret"`   // Passphrase or Password (Encrypted in DB)
	ConfigSource string `json:"config_source"` // local, database

	RemoteAuthsJSON string              `json:"-"`                     // Stored in DB
	RemoteAuths     map[string]AuthInfo `gorm:"-" json:"remote_auths"` // Memory & API

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
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
		encryptedMap := make(map[string]AuthInfo)
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
		var encryptedMap map[string]AuthInfo
		if err := json.Unmarshal([]byte(r.RemoteAuthsJSON), &encryptedMap); err == nil {
			decryptedMap := make(map[string]AuthInfo)
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

// SyncTask structure used for persistent tasks
type SyncTask struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Key           string         `gorm:"uniqueIndex" json:"key"`
	SourceRepoKey string         `json:"source_repo_key"`
	SourceRemote  string         `json:"source_remote"`
	SourceBranch  string         `json:"source_branch"`
	TargetRepoKey string         `json:"target_repo_key"`
	TargetRemote  string         `json:"target_remote"`
	TargetBranch  string         `json:"target_branch"`
	PushOptions   string         `json:"push_options"` // e.g. "--force --no-verify"
	Cron          string         `json:"cron"`         // e.g. "0 2 * * *"
	Enabled       bool           `json:"enabled"`
	WebhookToken  string         `json:"webhook_token"` // For webhook triggering
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Associations
	SourceRepo Repo `gorm:"foreignKey:SourceRepoKey;references:Key" json:"source_repo"`
	TargetRepo Repo `gorm:"foreignKey:TargetRepoKey;references:Key" json:"target_repo"`
}

type SyncRun struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TaskKey      string    `json:"task_key"`
	Status       string    `json:"status"` // success, failed, conflict
	CommitRange  string    `json:"commit_range"`
	ErrorMessage string    `json:"error_message"`
	Details      string    `json:"details" gorm:"type:text"` // Execution logs
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`

	// Associations
	Task SyncTask `gorm:"foreignKey:TaskKey;references:Key" json:"-"`
}
