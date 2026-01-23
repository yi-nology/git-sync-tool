package response

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/yi-nology/git-manage-service/pkg/errno"
)

// Response 标准响应结构（符合 AGENT.md 规范）
type Response struct {
	Code  int32       `json:"code"`            // 业务状态码，0 表示成功
	Msg   string      `json:"msg"`             // 操作提示消息
	Error string      `json:"error,omitempty"` // 详细错误信息（调试用）
	Data  interface{} `json:"data,omitempty"`  // 业务数据
}

// Success 成功响应
func Success(c *app.RequestContext, data interface{}) {
	c.JSON(consts.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

// Accepted 异步处理响应 (HTTP 202)
func Accepted(c *app.RequestContext, msg string, data interface{}) {
	c.JSON(consts.StatusAccepted, Response{
		Code: 0,
		Msg:  msg,
		Data: data,
	})
}

// Error 错误响应
func Error(c *app.RequestContext, err error) {
	e := errno.ConvertErr(err)
	c.JSON(consts.StatusOK, Response{
		Code:  e.ErrCode,
		Msg:   e.ErrMsg,
		Error: err.Error(),
	})
}

// ErrorWithCode 带自定义错误码的错误响应
func ErrorWithCode(c *app.RequestContext, code int32, msg string, err error) {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	c.JSON(consts.StatusOK, Response{
		Code:  code,
		Msg:   msg,
		Error: errStr,
	})
}

// BadRequest 参数错误响应
func BadRequest(c *app.RequestContext, msg string) {
	Error(c, errno.ParamErr.WithMessage(msg))
}

// NotFound 资源不存在响应
func NotFound(c *app.RequestContext, msg string) {
	Error(c, errno.NotFound.WithMessage(msg))
}

// InternalServerError 服务器内部错误响应
func InternalServerError(c *app.RequestContext, msg string) {
	Error(c, errno.ServiceErr.WithMessage(msg))
}

// Unauthorized 未授权响应
func Unauthorized(c *app.RequestContext, msg string) {
	Error(c, errno.Unauthorized.WithMessage(msg))
}

// Forbidden 禁止访问响应
func Forbidden(c *app.RequestContext, msg string) {
	Error(c, errno.Forbidden.WithMessage(msg))
}

// Conflict 冲突响应
func Conflict(c *app.RequestContext, msg string) {
	Error(c, errno.Conflict.WithMessage(msg))
}
