package po

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type WebhookEvent struct {
	gorm.Model
	EventID          string                 `gorm:"uniqueIndex;size:100" json:"event_id"`
	ProviderConfigID uint                   `gorm:"index" json:"provider_config_id"`
	EventType        string                 `gorm:"size:50;index" json:"event_type"`
	Source           string                 `gorm:"size:20;index" json:"source"`
	RepoID           uint                   `gorm:"index" json:"repo_id"`
	CRID             uint                   `gorm:"index" json:"cr_id"`
	PlatformCRNumber int                    `json:"platform_cr_number"`
	ActorName        string                 `gorm:"size:200" json:"actor_name"`
	ActorUsername    string                 `gorm:"size:200" json:"actor_username"`
	PayloadJSON      string                 `gorm:"type:text" json:"-"`
	Payload          map[string]interface{} `gorm:"-" json:"payload"`
	Status           string                 `gorm:"size:20;index" json:"status"`
	ProcessedAt      *time.Time             `json:"processed_at"`
	ErrorMessage     string                 `gorm:"size:500" json:"error_message"`
}

func (WebhookEvent) TableName() string { return "webhook_events" }

func (e *WebhookEvent) BeforeSave(tx *gorm.DB) error {
	if e.Payload != nil {
		b, err := json.Marshal(e.Payload)
		if err != nil {
			return err
		}
		e.PayloadJSON = string(b)
	}
	return nil
}

func (e *WebhookEvent) AfterFind(tx *gorm.DB) error {
	if e.PayloadJSON != "" {
		json.Unmarshal([]byte(e.PayloadJSON), &e.Payload)
	}
	return nil
}

type WebhookRule struct {
	gorm.Model
	Name             string                 `gorm:"size:100" json:"name"`
	ProviderConfigID uint                   `gorm:"index" json:"provider_config_id"`
	EventTypePattern string                 `gorm:"size:100" json:"event_type_pattern"`
	RepoPattern      string                 `gorm:"size:200" json:"repo_pattern"`
	Action           string                 `gorm:"size:50" json:"action"`
	ActionConfigJSON string                 `gorm:"type:text" json:"-"`
	ActionConfig     map[string]interface{} `gorm:"-" json:"action_config"`
	Enabled          bool                   `gorm:"default:true" json:"enabled"`
}

func (WebhookRule) TableName() string { return "webhook_rules" }

func (r *WebhookRule) BeforeSave(tx *gorm.DB) error {
	if r.ActionConfig != nil {
		b, err := json.Marshal(r.ActionConfig)
		if err != nil {
			return err
		}
		r.ActionConfigJSON = string(b)
	}
	return nil
}

func (r *WebhookRule) AfterFind(tx *gorm.DB) error {
	if r.ActionConfigJSON != "" {
		json.Unmarshal([]byte(r.ActionConfigJSON), &r.ActionConfig)
	}
	return nil
}
