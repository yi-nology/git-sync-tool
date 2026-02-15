// biz/model/biz/branch/custom_models.go - 手动添加的分支模型

package branch

// CherryPickRequest Cherry-pick请求
type CherryPickRequest struct {
	RepoKey    string `json:"repo_key" form:"repo_key"`
	CommitHash string `json:"commit_hash" form:"commit_hash"`
	NoCommit   bool   `json:"no_commit" form:"no_commit"`
}

// CherryPickResponse Cherry-pick响应
type CherryPickResponse struct {
	Success   bool     `json:"success"`
	NewCommit string   `json:"new_commit"`
	Conflicts []string `json:"conflicts"`
}

// RebaseRequest Rebase请求
type RebaseRequest struct {
	RepoKey     string `json:"repo_key" form:"repo_key"`
	Upstream    string `json:"upstream" form:"upstream"`
	Onto        string `json:"onto" form:"onto"`
	Interactive bool   `json:"interactive" form:"interactive"`
}

// RebaseResponse Rebase响应
type RebaseResponse struct {
	Success       bool     `json:"success"`
	InProgress    bool     `json:"in_progress"`
	Conflicts     []string `json:"conflicts"`
	CurrentCommit string   `json:"current_commit"`
}

// RebaseAbortRequest 中止Rebase请求
type RebaseAbortRequest struct {
	RepoKey string `json:"repo_key" form:"repo_key"`
}

// RebaseContinueRequest 继续Rebase请求
type RebaseContinueRequest struct {
	RepoKey string `json:"repo_key" form:"repo_key"`
}
