package po

import (
	"gorm.io/gorm"
)

// SyncTask structure used for persistent tasks
type SyncTask struct {
	gorm.Model
	Key           string `gorm:"uniqueIndex" json:"key"`
	SourceRepoKey string `json:"source_repo_key"`
	SourceRemote  string `json:"source_remote"`
	SourceBranch  string `json:"source_branch"`
	TargetRepoKey string `json:"target_repo_key"`
	TargetRemote  string `json:"target_remote"`
	TargetBranch  string `json:"target_branch"`
	PushOptions   string `json:"push_options"` // e.g. "--force --no-verify"
	Cron          string `json:"cron"`         // e.g. "0 2 * * *"
	Enabled       bool   `json:"enabled"`
	WebhookToken  string `gorm:"index" json:"webhook_token"`      // 用于Webhook触发的Token
	SyncMode      string `gorm:"default:single" json:"sync_mode"` // single: 单分支同步, all-branch: 全分支同步

	// Associations
	SourceRepo Repo `gorm:"foreignKey:SourceRepoKey;references:Key" json:"source_repo"`
	TargetRepo Repo `gorm:"foreignKey:TargetRepoKey;references:Key" json:"target_repo"`
}

func (SyncTask) TableName() string {
	return "sync_tasks"
}
