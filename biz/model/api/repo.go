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
	RemoteAuths map[string]domain.AuthInfo `json:"remote_auths"` // Optional auth per remote
}

type ScanRepoReq struct {
	Path string `json:"path"`
}

type CloneRepoReq struct {
	RemoteURL  string `json:"remote_url"`
	LocalPath  string `json:"local_path"`
	AuthType   string `json:"auth_type"`
	AuthKey    string `json:"auth_key"`
	AuthSecret string `json:"auth_secret"`
	SSHKeyID   uint   `json:"ssh_key_id"` // 数据库SSH密钥ID，优先于AuthKey
}

type TestConnectionReq struct {
	URL string `json:"url"`
}

type MergeReq struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	Message  string `json:"message"`
	Strategy string `json:"strategy"` // Not implemented yet
}

type RepoDTO struct {
	ID          uint                       `json:"id"`
	Key         string                     `json:"key"`
	Name        string                     `json:"name"`
	Path        string                     `json:"path"`
	RemoteURL   string                     `json:"remote_url"`
	AuthType    string                     `json:"auth_type"`
	AuthKey     string                     `json:"auth_key"`
	AuthSecret  string                     `json:"auth_secret"`
	RemoteAuths map[string]domain.AuthInfo `json:"remote_auths"`
	CreatedAt   time.Time                  `json:"created_at"`
	UpdatedAt   time.Time                  `json:"updated_at"`
}

func NewRepoDTO(r po.Repo) RepoDTO {
	return RepoDTO{
		ID:          r.ID,
		Key:         r.Key,
		Name:        r.Name,
		Path:        r.Path,
		RemoteURL:   r.RemoteURL,
		AuthType:    r.AuthType,
		AuthKey:     r.AuthKey,
		AuthSecret:  r.AuthSecret,
		RemoteAuths: r.RemoteAuths,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
