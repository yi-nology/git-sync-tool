package errno

import (
	"fmt"
)

// ErrNo 错误码结构
type ErrNo struct {
	ErrCode int32
	ErrMsg  string
}

func (e ErrNo) Error() string {
	return fmt.Sprintf("err_code=%d, err_msg=%s", e.ErrCode, e.ErrMsg)
}

// NewErrNo 创建新的错误码
func NewErrNo(code int32, msg string) ErrNo {
	return ErrNo{
		ErrCode: code,
		ErrMsg:  msg,
	}
}

// WithMessage 携带自定义消息
func (e ErrNo) WithMessage(msg string) ErrNo {
	e.ErrMsg = msg
	return e
}

// ==================== 通用错误码 (0-999) ====================

var (
	Success      = NewErrNo(0, "success")
	ServiceErr   = NewErrNo(500, "internal server error")
	ParamErr     = NewErrNo(400, "invalid parameter")
	Unauthorized = NewErrNo(401, "unauthorized")
	Forbidden    = NewErrNo(403, "forbidden")
	NotFound     = NewErrNo(404, "not found")
	Conflict     = NewErrNo(409, "conflict")
)

// ==================== 仓库相关错误码 (10000-10999) ====================

var (
	RepoNotFound      = NewErrNo(10001, "repository not found")
	RepoAlreadyExists = NewErrNo(10002, "repository already exists")
	RepoPathInvalid   = NewErrNo(10003, "invalid repository path")
	RepoCloneFailed   = NewErrNo(10004, "clone repository failed")
	RepoFetchFailed   = NewErrNo(10005, "fetch repository failed")
	RepoScanFailed    = NewErrNo(10006, "scan repository failed")
	RepoNotGit        = NewErrNo(10007, "path is not a valid git repository")
	RepoInUse         = NewErrNo(10008, "repository is used by sync tasks")
)

// ==================== 分支相关错误码 (11000-11999) ====================

var (
	BranchNotFound      = NewErrNo(11001, "branch not found")
	BranchAlreadyExists = NewErrNo(11002, "branch already exists")
	BranchDeleteFailed  = NewErrNo(11003, "delete branch failed")
	BranchCreateFailed  = NewErrNo(11004, "create branch failed")
	BranchRenameFailed  = NewErrNo(11005, "rename branch failed")
	CheckoutFailed      = NewErrNo(11006, "checkout branch failed")
	PushFailed          = NewErrNo(11007, "push failed")
	PullFailed          = NewErrNo(11008, "pull failed")
	MergeFailed         = NewErrNo(11009, "merge failed")
	MergeConflict       = NewErrNo(11010, "merge conflict detected")
	DirtyWorktree       = NewErrNo(11011, "working tree has uncommitted changes")
)

// ==================== 同步任务相关错误码 (12000-12999) ====================

var (
	SyncTaskNotFound   = NewErrNo(12001, "sync task not found")
	SyncRunFailed      = NewErrNo(12002, "sync execution failed")
	CronConfigErr      = NewErrNo(12003, "invalid cron configuration")
	SyncTaskDisabled   = NewErrNo(12004, "sync task is disabled")
	SyncAlreadyRunning = NewErrNo(12005, "sync task is already running")
)

// ==================== 认证相关错误码 (13000-13999) ====================

var (
	AuthFailed          = NewErrNo(13001, "authentication failed")
	SSHKeyInvalid       = NewErrNo(13002, "invalid SSH key")
	SSHKeyNotFound      = NewErrNo(13003, "SSH key not found")
	RemoteConnectFailed = NewErrNo(13004, "remote connection failed")
	CredentialInvalid   = NewErrNo(13005, "invalid credentials")
)

// ==================== 标签相关错误码 (14000-14999) ====================

var (
	TagNotFound      = NewErrNo(14001, "tag not found")
	TagAlreadyExists = NewErrNo(14002, "tag already exists")
	TagCreateFailed  = NewErrNo(14003, "create tag failed")
	TagDeleteFailed  = NewErrNo(14004, "delete tag failed")
)

// ==================== 系统相关错误码 (15000-15999) ====================

var (
	ConfigLoadFailed   = NewErrNo(15001, "load configuration failed")
	ConfigSaveFailed   = NewErrNo(15002, "save configuration failed")
	DirNotFound        = NewErrNo(15003, "directory not found")
	DirAccessDenied    = NewErrNo(15004, "directory access denied")
	FileOperationError = NewErrNo(15005, "file operation error")
)

// ==================== 兼容旧错误码 ====================

var (
	// 保持向后兼容
	AuthorizationFailedErr = Unauthorized
	RecordNotFound         = NotFound
)

// ConvertErr 将 error 转换为 ErrNo
func ConvertErr(err error) ErrNo {
	if err == nil {
		return Success
	}
	if e, ok := err.(ErrNo); ok {
		return e
	}
	return ServiceErr.WithMessage(err.Error())
}
