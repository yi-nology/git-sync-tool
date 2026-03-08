package api

// GeneratePatchReq 生成 patch 请求
type GeneratePatchReq struct {
	RepoKey string   `json:"repo_key" form:"repo_key"`
	Base    string   `json:"base" form:"base"`       // 基准分支/commit
	Target  string   `json:"target" form:"target"`   // 目标分支/commit
	Commits []string `json:"commits" form:"commits"` // 指定 commit 列表（可选）
}

// SavePatchReq 保存 patch 请求
type SavePatchReq struct {
	RepoKey       string `json:"repo_key" form:"repo_key"`
	PatchName     string `json:"patch_name" form:"patch_name"`         // patch 文件名
	PatchContent  string `json:"patch_content" form:"patch_content"`   // patch 内容
	CustomPath    string `json:"custom_path" form:"custom_path"`       // 自定义保存路径（可选）
	CommitMessage string `json:"commit_message" form:"commit_message"` // 提交消息（可选，为空则不提交）
	Sequence      int    `json:"sequence" form:"sequence"`             // 序号（用于排序）
}

// ApplyPatchReq 应用 patch 请求
type ApplyPatchReq struct {
	RepoKey       string `json:"repo_key" form:"repo_key"`
	PatchPath     string `json:"patch_path" form:"patch_path"`         // patch 文件路径
	PatchContent  string `json:"patch_content" form:"patch_content"`   // patch 内容（与 PatchPath 二选一）
	SignOff       bool   `json:"sign_off" form:"sign_off"`             // 是否添加 Signed-off-by
	CommitMessage string `json:"commit_message" form:"commit_message"` // 提交消息（为空则不自动提交）
	Force         bool   `json:"force" form:"force"`                   // 强制应用（跳过顺序检查）
}

// DeletePatchReq 删除 patch 请求
type DeletePatchReq struct {
	RepoKey       string `json:"repo_key" form:"repo_key"`
	PatchPath     string `json:"patch_path" form:"patch_path"`
	CommitMessage string `json:"commit_message" form:"commit_message"`
}

// PatchInfoDTO patch 信息
type PatchInfoDTO struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	ModTime     string `json:"mod_time"`
	Sequence    int    `json:"sequence"`     // 序号（从文件名解析）
	IsApplied   bool   `json:"is_applied"`   // 是否已应用
	CanApply    bool   `json:"can_apply"`    // 是否可以应用（考虑顺序）
	HasConflict bool   `json:"has_conflict"` // 是否有冲突
}

// PatchStatsDTO patch 统计信息
type PatchStatsDTO struct {
	Stat        string `json:"stat"`
	CanApply    bool   `json:"can_apply"`
	Error       string `json:"error,omitempty"`
	AppliedComm string `json:"applied_commit,omitempty"` // 如果已应用，记录应用的 commit
}

// PatchSeriesDTO patch 序列状态
type PatchSeriesDTO struct {
	RepoKey        string         `json:"repo_key"`
	TotalPatches   int            `json:"total_patches"`
	AppliedCount   int            `json:"applied_count"`
	PendingCount   int            `json:"pending_count"`
	ConflictCount  int            `json:"conflict_count"`
	Patches        []PatchInfoDTO `json:"patches"`
	CanApplyNext   bool           `json:"can_apply_next"`   // 是否可以应用下一个
	NextPatchIndex int            `json:"next_patch_index"` // 下一个待应用的 patch 索引
}

// ReorderPatchReq 重排 patch 顺序
type ReorderPatchReq struct {
	RepoKey    string   `json:"repo_key" form:"repo_key"`
	PatchOrder []string `json:"patch_order" form:"patch_order"` // 新的 patch 路径顺序
}
