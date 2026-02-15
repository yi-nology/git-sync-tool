package domain

import "time"

type GitRemote struct {
	Name       string   `json:"name"`
	FetchURL   string   `json:"fetch_url"`
	PushURL    string   `json:"push_url"`
	FetchSpecs []string `json:"fetch_specs"`
	PushSpecs  []string `json:"push_specs"`
	IsMirror   bool     `json:"is_mirror"`
}

type GitBranch struct {
	Name        string `json:"name"`
	Remote      string `json:"remote"`
	Merge       string `json:"merge"`        // refs/heads/xxx
	UpstreamRef string `json:"upstream_ref"` // e.g. origin/main
}

type GitRepoConfig struct {
	Remotes  []GitRemote `json:"remotes"`
	Branches []GitBranch `json:"branches"`
}

type BranchInfo struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"` // "local" or "remote"
	IsCurrent   bool      `json:"is_current"`
	Hash        string    `json:"hash"`
	Author      string    `json:"author"`
	AuthorEmail string    `json:"author_email"`
	Date        time.Time `json:"date"`
	Message     string    `json:"message"`

	// Sync Status
	Upstream string `json:"upstream"`
	Ahead    int    `json:"ahead"`
	Behind   int    `json:"behind"`
}
