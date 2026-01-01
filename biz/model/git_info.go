package model

type GitRemote struct {
	Name       string   `json:"name"`
	FetchURL   string   `json:"fetch_url"`
	PushURL    string   `json:"push_url"`
	FetchSpecs []string `json:"fetch_specs"`
	PushSpecs  []string `json:"push_specs"`
	IsMirror   bool     `json:"is_mirror"`
}

type GitBranch struct {
	Name         string `json:"name"`
	Remote       string `json:"remote"`
	Merge        string `json:"merge"` // refs/heads/xxx
	UpstreamRef  string `json:"upstream_ref"` // e.g. origin/main
}

type GitRepoConfig struct {
	Remotes  []GitRemote `json:"remotes"`
	Branches []GitBranch `json:"branches"`
}
