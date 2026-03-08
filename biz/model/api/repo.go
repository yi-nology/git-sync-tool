package api

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type RegisterRepoReq struct {
	Name        string                     `json:"name"`
	Path        string                     `json:"path"`
	RemoteURL   string                     `json:"remote_url"`
	AuthType    string                     `json:"auth_type"`
	AuthKey     string                     `json:"auth_key"`
	AuthSecret  string                     `json:"auth_secret"`
	Remotes     []domain.GitRemote         `json:"remotes"`      // Optional list of remotes to sync
	RemoteAuths map[string]domain.AuthInfo `json:"remote_auths"` // Optional auth per remote (deprecated)
	// 新凭证池字段
	DefaultCredentialID uint            `json:"default_credential_id"` // 默认凭证 ID
	RemoteCredentials   map[string]uint `json:"remote_credentials"`    // remote name -> credential ID
}

type ScanRepoReq struct {
	Path string `json:"path"`
}

type CloneRepoReq struct {
	RemoteURL    string `json:"remote_url"`
	LocalPath    string `json:"local_path"`
	AuthType     string `json:"auth_type"`
	AuthKey      string `json:"auth_key"`
	AuthSecret   string `json:"auth_secret"`
	SSHKeyID     uint   `json:"ssh_key_id"`    // 数据库SSH密钥ID，优先于AuthKey (deprecated)
	CredentialID uint   `json:"credential_id"` // 凭证 ID (新字段)
}

type TestConnectionReq struct {
	URL string `json:"url"`
}

// ScanDirectoryReq 扫描目录请求
type ScanDirectoryReq struct {
	Path     string `json:"path"`     // 要扫描的父目录
	Depth    int    `json:"depth"`    // 扫描深度，默认 2
	Recursive bool   `json:"recursive"` // 是否递归扫描
}

// ScannedRepo 扫描到的仓库信息
type ScannedRepo struct {
	Name          string           `json:"name"`
	Path          string           `json:"path"`
	Remotes       []domain.GitRemote `json:"remotes"`
	CurrentBranch string           `json:"current_branch"`
	LastCommit    string           `json:"last_commit"`
	HasChanges    bool             `json:"has_changes"`
}

// ScanDirectoryResp 扫描目录响应
type ScanDirectoryResp struct {
	Repos []ScannedRepo `json:"repos"`
	Total int           `json:"total"`
}

// BatchCreateRepoReq 批量注册仓库请求
type BatchCreateRepoReq struct {
	Repos []BatchRepoItem `json:"repos"`
}

// BatchRepoItem 批量注册的仓库项
type BatchRepoItem struct {
	Name                string `json:"name"`
	Path                string `json:"path"`
	DefaultCredentialID uint   `json:"default_credential_id,omitempty"`
}

// BatchCreateRepoResp 批量注册仓库响应
type BatchCreateRepoResp struct {
	Success []RepoDTO `json:"success"`
	Failed  []BatchFailedItem `json:"failed"`
}

// BatchFailedItem 失败项
type BatchFailedItem struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

type MergeReq struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	Message  string `json:"message"`
	Strategy string `json:"strategy"` // Not implemented yet
}

type RepoDTO struct {
	ID                  uint                       `json:"id"`
	Key                 string                     `json:"key"`
	Name                string                     `json:"name"`
	Path                string                     `json:"path"`
	RemoteURL           string                     `json:"remote_url"`
	AuthType            string                     `json:"auth_type"`
	AuthKey             string                     `json:"auth_key"`
	AuthSecret          string                     `json:"auth_secret"`
	RemoteAuths         map[string]domain.AuthInfo `json:"remote_auths"`
	DefaultCredentialID uint                       `json:"default_credential_id,omitempty"`
	RemoteCredentials   map[string]uint            `json:"remote_credentials,omitempty"`
	CreatedAt           time.Time                  `json:"created_at"`
	UpdatedAt           time.Time                  `json:"updated_at"`
}

func NewRepoDTO(r po.Repo) RepoDTO {
	return RepoDTO{
		ID:                  r.ID,
		Key:                 r.Key,
		Name:                r.Name,
		Path:                r.Path,
		RemoteURL:           r.RemoteURL,
		AuthType:            r.AuthType,
		AuthKey:             r.AuthKey,
		AuthSecret:          r.AuthSecret,
		RemoteAuths:         r.RemoteAuths,
		DefaultCredentialID: r.DefaultCredentialID,
		RemoteCredentials:   r.RemoteCredentials,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
	}
}
