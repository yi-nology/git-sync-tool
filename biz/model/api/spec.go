package api

import "time"

type SpecFile struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	IsDir    bool       `json:"is_dir"`
	Children []SpecFile `json:"children,omitempty"`
	Size     int64      `json:"size,omitempty"`
	ModTime  time.Time  `json:"mod_time,omitempty"`
}

type FileContent struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type SaveRequest struct {
	Content    string `json:"content"`
	Message    string `json:"message"`
	AutoCommit bool   `json:"autoCommit"`
}

type LintRequest struct {
	Content string   `json:"content"`
	Rules   []string `json:"rules,omitempty"`
}

type LintResult struct {
	File   string      `json:"file"`
	Issues []LintIssue `json:"issues"`
	Stats  LintStats   `json:"stats"`
}

type LintIssue struct {
	RuleID    string `json:"ruleId"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Line      int    `json:"line"`
	Column    int    `json:"column,omitempty"`
	EndLine   int    `json:"endLine,omitempty"`
	EndColumn int    `json:"endColumn,omitempty"`
}

type LintStats struct {
	ErrorCount   int `json:"errorCount"`
	WarningCount int `json:"warningCount"`
	InfoCount    int `json:"infoCount"`
}

type LintRule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Severity    string    `json:"severity"`
	Pattern     string    `json:"pattern"`
	Enabled     bool      `json:"enabled"`
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type UpdateLintRuleReq struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Severity    string `json:"severity,omitempty"`
	Pattern     string `json:"pattern,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
	Priority    *int   `json:"priority,omitempty"`
}

type CreateLintRuleReq struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Severity    string `json:"severity"`
	Pattern     string `json:"pattern"`
	Enabled     bool   `json:"enabled"`
	Priority    int    `json:"priority"`
}

type CommitRequest struct {
	Message string `json:"message"`
	Content string `json:"content,omitempty"`
}

type GetSpecTreeReq struct {
	RepoKey string `form:"repo_key" query:"repo_key"`
}

type GetSpecContentReq struct {
	RepoKey string `form:"repo_key" query:"repo_key"`
	Path    string `form:"path" query:"path"`
}

type SaveSpecContentReq struct {
	RepoKey    string `json:"repo_key" form:"repo_key"`
	Path       string `json:"path" form:"path"`
	Content    string `json:"content" form:"content"`
	Message    string `json:"message" form:"message"`
	AutoCommit bool   `json:"autoCommit" form:"autoCommit"`
}

type CommitSpecReq struct {
	RepoKey string `json:"repo_key" form:"repo_key"`
	Path    string `json:"path" form:"path"`
	Message string `json:"message" form:"message"`
	Content string `json:"content" form:"content,omitempty"`
}

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
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // error, warning, info
	Rule     string `json:"rule"`
	RuleDesc string `json:"rule_desc"`
	QuickFix string `json:"quick_fix,omitempty"`
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
	Content string `json:"content" form:"content"` // 可选，如果提供则使用此内容，否则使用模板
}

// DeleteSpecFileReq 删除 spec 文件请求
type DeleteSpecFileReq struct {
	RepoKey       string `json:"repo_key" form:"repo_key"`
	Path          string `json:"path" form:"path"`
	CommitMessage string `json:"commit_message" form:"commit_message"`
}
