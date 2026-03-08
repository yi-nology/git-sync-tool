package api

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type SyncTaskDTO struct {
	ID            uint      `json:"id"`
	Key           string    `json:"key"`
	SourceRepoKey string    `json:"source_repo_key"`
	SourceRemote  string    `json:"source_remote"`
	SourceBranch  string    `json:"source_branch"`
	TargetRepoKey string    `json:"target_repo_key"`
	TargetRemote  string    `json:"target_remote"`
	TargetBranch  string    `json:"target_branch"`
	PushOptions   string    `json:"push_options"`
	Cron          string    `json:"cron"`
	Enabled       bool      `json:"enabled"`
	SyncMode      string    `json:"sync_mode"`
	GitTags       bool      `json:"git_tags"`
	GitForce      bool      `json:"git_force"`
	GitPrune      bool      `json:"git_prune"`
	GitNoVerify   bool      `json:"git_no_verify"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	SourceRepo RepoDTO `json:"source_repo"`
	TargetRepo RepoDTO `json:"target_repo"`
}

func NewSyncTaskDTO(t po.SyncTask) SyncTaskDTO {
	dto := SyncTaskDTO{
		ID:            t.ID,
		Key:           t.Key,
		SourceRepoKey: t.SourceRepoKey,
		SourceRemote:  t.SourceRemote,
		SourceBranch:  t.SourceBranch,
		TargetRepoKey: t.TargetRepoKey,
		TargetRemote:  t.TargetRemote,
		TargetBranch:  t.TargetBranch,
		PushOptions:   t.PushOptions,
		Cron:          t.Cron,
		Enabled:       t.Enabled,
		SyncMode:      t.SyncMode,
		GitTags:       t.GitTags,
		GitForce:      t.GitForce,
		GitPrune:      t.GitPrune,
		GitNoVerify:   t.GitNoVerify,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
	if t.SourceRepo.ID != 0 {
		dto.SourceRepo = NewRepoDTO(t.SourceRepo)
	}
	if t.TargetRepo.ID != 0 {
		dto.TargetRepo = NewRepoDTO(t.TargetRepo)
	}
	return dto
}
