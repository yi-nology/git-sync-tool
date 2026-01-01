package model

import (
	"time"

	"gorm.io/gorm"
)

// AuditLog records user operations for security and tracking
type AuditLog struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Action    string         `gorm:"index" json:"action"`    // CREATE, UPDATE, DELETE, SYNC, etc.
	Target    string         `gorm:"index" json:"target"`    // repo:1, task:abc, etc.
	Operator  string         `json:"operator"`               // User ID or IP (since we don't have full auth yet)
	Details   string         `json:"details" gorm:"type:text"` // JSON payload of changes or details
	IPAddress string         `json:"ip_address"`
	UserAgent string         `json:"user_agent"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
