package po

import (
	"time"

	"gorm.io/gorm"
)

// BackupRecord 备份记录
type BackupRecord struct {
	gorm.Model
	RepoID      uint      `gorm:"column:repo_id;index" json:"repo_id"`               // 仓库ID
	RepoKey     string    `gorm:"column:repo_key;index" json:"repo_key"`             // 仓库Key
	StorageKey  string    `gorm:"column:storage_key" json:"storage_key"`             // 对象存储Key
	Size        int64     `gorm:"column:size" json:"size"`                           // 备份大小（字节）
	Status      string    `gorm:"column:status;index" json:"status"`                 // 状态: pending, success, failed
	ErrorMsg    string    `gorm:"column:error_msg" json:"error_msg,omitempty"`       // 错误信息
	StartedAt   time.Time `gorm:"column:started_at" json:"started_at"`               // 开始时间
	CompletedAt time.Time `gorm:"column:completed_at" json:"completed_at,omitempty"` // 完成时间
}

func (BackupRecord) TableName() string {
	return "backup_records"
}
