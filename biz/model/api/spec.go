package api

// SpecFileInfo spec 文件信息
type SpecFileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size,omitempty"`
	ModTime string `json:"mod_time,omitempty"`
}

// SpecFileContent spec 文件内容
type SpecFileContent struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// SaveSpecReq 保存 spec 文件请求
type SaveSpecReq struct {
	RepoKey       string `json:"repo_key" form:"repo_key"`
	Path          string `json:"path" form:"path"`
	Content       string `json:"content" form:"content"`
	CommitMessage string `json:"commit_message" form:"commit_message"`
}

// SpecValidationResult 验证结果
type SpecValidationResult struct {
	Valid    bool              `json:"valid"`
	Issues   []SpecIssue       `json:"issues"`
	Warnings []SpecIssue       `json:"warnings"`
	Stats    map[string]string `json:"stats"`
}

// SpecIssue spec 问题
type SpecIssue struct {
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	Message   string `json:"message"`
	Severity  string `json:"severity"` // error, warning, info
	Rule      string `json:"rule"`
	RuleDesc  string `json:"rule_desc"`
	QuickFix  string `json:"quick_fix,omitempty"`
}

// SpecRule spec 规则
type SpecRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // error, warning, info
	Pattern     string `json:"pattern"`  // 正则或规则类型
	Enabled     bool   `json:"enabled"`
	Category    string `json:"category"` // required, style, security, best-practice
	AutoFix     bool   `json:"auto_fix"`
}

// CreateSpecFileReq 创建新 spec 文件请求
type CreateSpecFileReq struct {
	RepoKey string `json:"repo_key" form:"repo_key"`
	Path    string `json:"path" form:"path"`
	Name    string `json:"name" form:"name"`
}

// DeleteSpecFileReq 删除 spec 文件请求
type DeleteSpecFileReq struct {
	RepoKey       string `json:"repo_key" form:"repo_key"`
	Path          string `json:"path" form:"path"`
	CommitMessage string `json:"commit_message" form:"commit_message"`
}
