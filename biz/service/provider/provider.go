package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Platform string

const (
	PlatformGitLab Platform = "gitlab"
	PlatformGitHub Platform = "github"
	PlatformGitea  Platform = "gitea"
)

type Provider interface {
	Platform() Platform
	ListRepos(ctx context.Context, opts ListRepoOptions) ([]*PlatformRepo, error)
	GetRepo(ctx context.Context, owner, repo string) (*PlatformRepo, error)
	CreateCR(ctx context.Context, opts CreateCROptions) (*ChangeRequest, error)
	GetCR(ctx context.Context, owner, repo string, number int) (*ChangeRequest, error)
	ListCRs(ctx context.Context, opts ListCROptions) ([]*ChangeRequest, int, error)
	MergeCR(ctx context.Context, owner, repo string, number int, opts MergeCROptions) (*ChangeRequest, error)
	CloseCR(ctx context.Context, owner, repo string, number int) (*ChangeRequest, error)
	CreateWebhook(ctx context.Context, opts CreateWebhookOptions) (*PlatformWebhook, error)
	DeleteWebhook(ctx context.Context, owner, repo string, webhookID int64) error
	ListWebhooks(ctx context.Context, owner, repo string) ([]*PlatformWebhook, error)
	ParseWebhookEvent(r *http.Request, secret string) (*NormalizedEvent, error)
	ValidateWebhookSignature(r *http.Request, secret string) error
	TestConnection(ctx context.Context) (*TestConnectionResult, error)
}

type PlatformRepo struct {
	ID            int64    `json:"id"`
	FullName      string   `json:"full_name"`
	Name          string   `json:"name"`
	Owner         string   `json:"owner"`
	Description   string   `json:"description"`
	CloneURL      string   `json:"clone_url"`
	SSHURL        string   `json:"ssh_url"`
	DefaultBranch string   `json:"default_branch"`
	Private       bool     `json:"private"`
	Platform      Platform `json:"platform"`
}

type CRState string

const (
	CRStateOpened CRState = "opened"
	CRStateMerged CRState = "merged"
	CRStateClosed CRState = "closed"
)

type ChangeRequest struct {
	ID           int64     `json:"id"`
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	State        CRState   `json:"state"`
	SourceBranch string    `json:"source_branch"`
	TargetBranch string    `json:"target_branch"`
	Author       *CRUser   `json:"author"`
	Reviewers    []*CRUser `json:"reviewers"`
	Labels       []string  `json:"labels"`
	MergeStatus  string    `json:"merge_status"`
	WebURL       string    `json:"web_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CRUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type ListRepoOptions struct {
	Owner   string `json:"owner"`
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
}

type CreateCROptions struct {
	Owner              string   `json:"owner"`
	Repo               string   `json:"repo"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	SourceBranch       string   `json:"source_branch"`
	TargetBranch       string   `json:"target_branch"`
	Labels             []string `json:"labels"`
	RemoveSourceBranch bool     `json:"remove_source_branch"`
}

type ListCROptions struct {
	Owner        string  `json:"owner"`
	Repo         string  `json:"repo"`
	State        CRState `json:"state"`
	SourceBranch string  `json:"source_branch"`
	TargetBranch string  `json:"target_branch"`
	Page         int     `json:"page"`
	PerPage      int     `json:"per_page"`
}

type MergeCROptions struct {
	MergeCommitMessage string `json:"merge_commit_message"`
	Squash             bool   `json:"squash"`
	RemoveSourceBranch bool   `json:"remove_source_branch"`
}

type CreateWebhookOptions struct {
	Owner  string   `json:"owner"`
	Repo   string   `json:"repo"`
	URL    string   `json:"url"`
	Secret string   `json:"secret"`
	Events []string `json:"events"`
}

type PlatformWebhook struct {
	ID     int64    `json:"id"`
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

type NormalizedEvent struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Source     Platform        `json:"source"`
	Timestamp  time.Time       `json:"timestamp"`
	Actor      *CRUser         `json:"actor"`
	Repo       *EventRepo      `json:"repo"`
	CR         *ChangeRequest  `json:"cr,omitempty"`
	Branch     string          `json:"branch,omitempty"`
	Tag        string          `json:"tag,omitempty"`
	RawPayload json.RawMessage `json:"raw_payload"`
}

type EventRepo struct {
	FullName string `json:"full_name"`
	Owner    string `json:"owner"`
	Name     string `json:"name"`
}

type TestConnectionResult struct {
	Connected bool   `json:"connected"`
	Platform  string `json:"platform"`
	UserName  string `json:"user_name"`
	Message   string `json:"message,omitempty"`
}
