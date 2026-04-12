package api

import "time"

type ProviderConfigDTO struct {
	ID               uint      `json:"id"`
	Name             string    `json:"name"`
	Platform         string    `json:"platform"`
	BaseURL          string    `json:"base_url"`
	CredentialID     uint      `json:"credential_id"`
	CredentialName   string    `json:"credential_name,omitempty"`
	WebhookEndpoint  string    `json:"webhook_endpoint,omitempty"`
	HasWebhookSecret bool      `json:"has_webhook_secret"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateProviderConfigReq struct {
	Name          string `json:"name"`
	Platform      string `json:"platform"`
	BaseURL       string `json:"base_url"`
	CredentialID  uint   `json:"credential_id"`
	WebhookSecret string `json:"webhook_secret"`
}

type UpdateProviderConfigReq struct {
	Name          string `json:"name"`
	BaseURL       string `json:"base_url"`
	CredentialID  uint   `json:"credential_id"`
	WebhookSecret string `json:"webhook_secret"`
}

type TestProviderConfigResp struct {
	Connected bool   `json:"connected"`
	Platform  string `json:"platform"`
	UserName  string `json:"user_name"`
	Message   string `json:"message,omitempty"`
}

type CRDTO struct {
	ID             uint       `json:"id"`
	RepoID         uint       `json:"repo_id"`
	RepoName       string     `json:"repo_name,omitempty"`
	ProviderID     uint       `json:"provider_id"`
	Platform       string     `json:"platform,omitempty"`
	CRNumber       int        `json:"cr_number"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	State          string     `json:"state"`
	SourceBranch   string     `json:"source_branch"`
	TargetBranch   string     `json:"target_branch"`
	AuthorName     string     `json:"author_name"`
	AuthorUsername string     `json:"author_username"`
	WebURL         string     `json:"web_url"`
	MergeStatus    string     `json:"merge_status"`
	Labels         []string   `json:"labels"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	MergedAt       *time.Time `json:"merged_at,omitempty"`
}

type CreateCRReq struct {
	RepoKey            string   `json:"repo_key"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	SourceBranch       string   `json:"source_branch"`
	TargetBranch       string   `json:"target_branch"`
	Labels             []string `json:"labels"`
	RemoveSourceBranch bool     `json:"remove_source_branch"`
}

type MergeCRReq struct {
	RepoKey            string `json:"repo_key"`
	CRNumber           int    `json:"cr_number"`
	MergeCommitMessage string `json:"merge_commit_message"`
	Squash             bool   `json:"squash"`
	RemoveSourceBranch bool   `json:"remove_source_branch"`
}

type CloseCRReq struct {
	RepoKey  string `json:"repo_key"`
	CRNumber int    `json:"cr_number"`
}

type ListCRsReq struct {
	RepoKey      string `json:"repo_key" query:"repo_key"`
	State        string `json:"state" query:"state"`
	SourceBranch string `json:"source_branch" query:"source_branch"`
	TargetBranch string `json:"target_branch" query:"target_branch"`
	Page         int    `json:"page" query:"page"`
	PageSize     int    `json:"page_size" query:"page_size"`
}

type GetCRReq struct {
	RepoKey  string `json:"repo_key" query:"repo_key"`
	CRNumber int    `json:"cr_number" query:"cr_number"`
}

type SyncCRsReq struct {
	RepoKey string `json:"repo_key"`
	State   string `json:"state"`
}

type WebhookEventDTO struct {
	ID               uint       `json:"id"`
	EventID          string     `json:"event_id"`
	EventType        string     `json:"event_type"`
	Source           string     `json:"source"`
	RepoID           uint       `json:"repo_id,omitempty"`
	CRID             uint       `json:"cr_id,omitempty"`
	PlatformCRNumber int        `json:"platform_cr_number,omitempty"`
	ActorName        string     `json:"actor_name"`
	ActorUsername    string     `json:"actor_username"`
	Status           string     `json:"status"`
	ProcessedAt      *time.Time `json:"processed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	ErrorMessage     string     `json:"error_message,omitempty"`
}

type ListWebhookEventsReq struct {
	EventType string `json:"event_type" query:"event_type"`
	Source    string `json:"source" query:"source"`
	Status    string `json:"status" query:"status"`
	Page      int    `json:"page" query:"page"`
	PageSize  int    `json:"page_size" query:"page_size"`
}

type WebhookRuleDTO struct {
	ID               uint                   `json:"id"`
	Name             string                 `json:"name"`
	ProviderConfigID uint                   `json:"provider_config_id"`
	EventTypePattern string                 `json:"event_type_pattern"`
	RepoPattern      string                 `json:"repo_pattern"`
	Action           string                 `json:"action"`
	ActionConfig     map[string]interface{} `json:"action_config"`
	Enabled          bool                   `json:"enabled"`
	CreatedAt        time.Time              `json:"created_at"`
}

type CreateWebhookRuleReq struct {
	Name             string                 `json:"name"`
	ProviderConfigID uint                   `json:"provider_config_id"`
	EventTypePattern string                 `json:"event_type_pattern"`
	RepoPattern      string                 `json:"repo_pattern"`
	Action           string                 `json:"action"`
	ActionConfig     map[string]interface{} `json:"action_config"`
	Enabled          bool                   `json:"enabled"`
}
