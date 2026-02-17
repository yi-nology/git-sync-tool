package po

import (
	"time"

	"gorm.io/gorm"
)

// 触发来源常量
const (
	TriggerSourceManual  = "manual"  // 手动触发
	TriggerSourceCron    = "cron"    // 定时任务触发
	TriggerSourceWebhook = "webhook" // Webhook 触发
)

type SyncRun struct {
	gorm.Model
	TaskKey       string    `json:"task_key"`
	TriggerSource string    `json:"trigger_source"` // 触发来源: manual, cron, webhook
	Status        string    `json:"status"`         // success, failed, conflict
	CommitRange   string    `json:"commit_range"`
	ErrorMessage  string    `json:"error_message"`
	Details       string    `json:"details" gorm:"type:text"` // Execution logs
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`

	// Associations
	Task SyncTask `gorm:"foreignKey:TaskKey;references:Key" json:"task"`
}

func (SyncRun) TableName() string {
	return "sync_runs"
}
