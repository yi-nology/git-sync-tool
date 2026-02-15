package constants

// 时间格式常量
const (
	TimeFormatFull     = "2006-01-02 15:04:05"
	TimeFormatDate     = "2006-01-02"
	TimeFormatDateOnly = "20060102"
	TimeFormatTime     = "15:04:05"
)

// 分页常量
const (
	DefaultPageSize = 100
	MaxPageSize     = 1000
	DefaultPage     = 1
)

// Git 默认值
const (
	DefaultRemoteName = "origin"
	DefaultBranch     = "main"
	DefaultVersion    = "v0.0.0"
	InitialVersion    = "v0.1.0"
)

// 认证类型常量
const (
	AuthTypeSSH      = "ssh"
	AuthTypePassword = "password"
	AuthTypeToken    = "token"
	AuthTypeNone     = "none"
)

// 分支类型常量
const (
	BranchTypeLocal  = "local"
	BranchTypeRemote = "remote"
	BranchTypeAll    = "all"
)

// Context 键常量
const (
	ContextKeyRepo      = "repo"
	ContextKeyRequestID = "request_id"
	ContextKeyUserID    = "user_id"
)

// HTTP Header 常量
const (
	HeaderRequestID   = "X-Request-ID"
	HeaderContentType = "Content-Type"
)

// 同步状态常量
const (
	SyncStatusPending  = "pending"
	SyncStatusRunning  = "running"
	SyncStatusSuccess  = "success"
	SyncStatusFailed   = "failed"
	SyncStatusCanceled = "canceled"
)

// 操作类型常量 (用于审计日志)
const (
	OperationCreate = "create"
	OperationUpdate = "update"
	OperationDelete = "delete"
	OperationMerge  = "merge"
	OperationRebase = "rebase"
	OperationPush   = "push"
	OperationFetch  = "fetch"
	OperationClone  = "clone"
	OperationSync   = "sync"
)

// 实体类型常量 (用于审计日志)
const (
	EntityRepo    = "repo"
	EntityBranch  = "branch"
	EntityTag     = "tag"
	EntityCommit  = "commit"
	EntitySSHKey  = "ssh_key"
	EntitySync    = "sync"
	EntityWebhook = "webhook"
)
