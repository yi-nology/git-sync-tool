package po

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type ChangeRequest struct {
	gorm.Model
	RepoID           uint       `gorm:"index" json:"repo_id"`
	ProviderConfigID uint       `gorm:"index" json:"provider_config_id"`
	PlatformCRID     int64      `gorm:"index" json:"platform_cr_id"`
	CRNumber         int        `gorm:"index" json:"cr_number"`
	Title            string     `gorm:"size:500" json:"title"`
	Description      string     `gorm:"type:text" json:"description"`
	State            string     `gorm:"size:20;index" json:"state"`
	SourceBranch     string     `gorm:"size:200;index" json:"source_branch"`
	TargetBranch     string     `gorm:"size:200;index" json:"target_branch"`
	AuthorName       string     `gorm:"size:200" json:"author_name"`
	AuthorUsername   string     `gorm:"size:200" json:"author_username"`
	WebURL           string     `gorm:"size:500" json:"web_url"`
	MergeStatus      string     `gorm:"size:30" json:"merge_status"`
	LabelsJSON       string     `gorm:"type:text" json:"-"`
	Labels           []string   `gorm:"-" json:"labels"`
	MergedAt         *time.Time `json:"merged_at"`
	ClosedAt         *time.Time `json:"closed_at"`
}

func (ChangeRequest) TableName() string { return "change_requests" }

func (c *ChangeRequest) BeforeSave(tx *gorm.DB) error {
	if c.Labels != nil {
		b, err := json.Marshal(c.Labels)
		if err != nil {
			return err
		}
		c.LabelsJSON = string(b)
	}
	return nil
}

func (c *ChangeRequest) AfterFind(tx *gorm.DB) error {
	if c.LabelsJSON != "" {
		json.Unmarshal([]byte(c.LabelsJSON), &c.Labels)
	}
	return nil
}
