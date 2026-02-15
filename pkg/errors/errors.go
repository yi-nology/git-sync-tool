package errors

import (
	"fmt"
)

// AppError 应用错误类型
type AppError struct {
	Code     int32  // 业务错误码
	HTTPCode int    // HTTP 状态码
	Message  string // 错误消息
	Err      error  // 原始错误
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap 返回原始错误
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithMessage 设置错误消息
func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Code:     e.Code,
		HTTPCode: e.HTTPCode,
		Message:  msg,
		Err:      e.Err,
	}
}

// WithMessagef 格式化设置错误消息
func (e *AppError) WithMessagef(format string, args ...interface{}) *AppError {
	return &AppError{
		Code:     e.Code,
		HTTPCode: e.HTTPCode,
		Message:  fmt.Sprintf(format, args...),
		Err:      e.Err,
	}
}

// Wrap 包装原始错误
func (e *AppError) Wrap(err error) *AppError {
	return &AppError{
		Code:     e.Code,
		HTTPCode: e.HTTPCode,
		Message:  e.Message,
		Err:      err,
	}
}

// WrapWithMessage 包装错误并设置消息
func (e *AppError) WrapWithMessage(err error, msg string) *AppError {
	return &AppError{
		Code:     e.Code,
		HTTPCode: e.HTTPCode,
		Message:  msg,
		Err:      err,
	}
}

// New 创建新的应用错误
func New(code int32, httpCode int, message string) *AppError {
	return &AppError{
		Code:     code,
		HTTPCode: httpCode,
		Message:  message,
	}
}

// 预定义错误
var (
	// 通用错误
	ErrValidation = &AppError{Code: 400, HTTPCode: 400, Message: "validation error"}
	ErrBadRequest = &AppError{Code: 400, HTTPCode: 400, Message: "bad request"}
	ErrNotFound   = &AppError{Code: 404, HTTPCode: 404, Message: "not found"}
	ErrConflict   = &AppError{Code: 409, HTTPCode: 409, Message: "conflict"}
	ErrInternal   = &AppError{Code: 500, HTTPCode: 500, Message: "internal server error"}

	// 仓库相关错误
	ErrRepoNotFound    = &AppError{Code: 40401, HTTPCode: 404, Message: "repository not found"}
	ErrRepoKeyRequired = &AppError{Code: 40001, HTTPCode: 400, Message: "repo_key is required"}
	ErrRepoPathInvalid = &AppError{Code: 40002, HTTPCode: 400, Message: "invalid repository path"}

	// Git 操作错误
	ErrGitCommand  = &AppError{Code: 50001, HTTPCode: 500, Message: "git command execution failed"}
	ErrGitConflict = &AppError{Code: 40901, HTTPCode: 409, Message: "git conflict detected"}
	ErrGitMerge    = &AppError{Code: 50002, HTTPCode: 500, Message: "git merge failed"}
	ErrGitCheckout = &AppError{Code: 50003, HTTPCode: 500, Message: "git checkout failed"}
	ErrGitPush     = &AppError{Code: 50004, HTTPCode: 500, Message: "git push failed"}
	ErrGitFetch    = &AppError{Code: 50005, HTTPCode: 500, Message: "git fetch failed"}
	ErrGitClone    = &AppError{Code: 50006, HTTPCode: 500, Message: "git clone failed"}
	ErrGitAuth     = &AppError{Code: 40101, HTTPCode: 401, Message: "git authentication failed"}

	// 分支相关错误
	ErrBranchNotFound  = &AppError{Code: 40402, HTTPCode: 404, Message: "branch not found"}
	ErrBranchExists    = &AppError{Code: 40902, HTTPCode: 409, Message: "branch already exists"}
	ErrBranchProtected = &AppError{Code: 40301, HTTPCode: 403, Message: "branch is protected"}

	// SSH 密钥错误
	ErrSSHKeyNotFound = &AppError{Code: 40403, HTTPCode: 404, Message: "SSH key not found"}
	ErrSSHKeyInvalid  = &AppError{Code: 40003, HTTPCode: 400, Message: "invalid SSH key"}
	ErrSSHKeyExists   = &AppError{Code: 40903, HTTPCode: 409, Message: "SSH key already exists"}
)

// IsAppError 检查是否为应用错误
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// AsAppError 转换为应用错误
func AsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

// GetCode 获取错误码
func GetCode(err error) int32 {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return 500
}

// GetHTTPCode 获取 HTTP 状态码
func GetHTTPCode(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.HTTPCode
	}
	return 500
}
