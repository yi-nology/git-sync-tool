package po

import (
	"time"

	"gorm.io/gorm"
)

type LintRule struct {
	ID          string    `gorm:"primaryKey;size:64" json:"id"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Description string    `gorm:"size:512" json:"description"`
	Category    string    `gorm:"size:32;not null;index" json:"category"` // syntax, best_practice, custom, required, style
	Severity    string    `gorm:"size:16;not null" json:"severity"`       // error, warning, info
	Pattern     string    `gorm:"size:1024" json:"pattern"`               // regex pattern or builtin rule name
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	Priority    int       `gorm:"default:0" json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (LintRule) TableName() string {
	return "lint_rules"
}

func (r *LintRule) BeforeCreate(tx *gorm.DB) error {
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	return nil
}

func (r *LintRule) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = time.Now()
	return nil
}
