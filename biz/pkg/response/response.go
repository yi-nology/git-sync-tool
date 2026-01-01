package response

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// Response standard structure
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success response
func Success(c *app.RequestContext, data interface{}) {
	c.JSON(consts.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error response
func Error(c *app.RequestContext, httpCode int, errCode int, msg string) {
	c.JSON(httpCode, Response{
		Code:    errCode,
		Message: msg,
	})
}

// Common Errors
func BadRequest(c *app.RequestContext, msg string) {
	Error(c, consts.StatusBadRequest, 400, msg)
}

func InternalServerError(c *app.RequestContext, msg string) {
	Error(c, consts.StatusInternalServerError, 500, msg)
}

func NotFound(c *app.RequestContext, msg string) {
	Error(c, consts.StatusNotFound, 404, msg)
}
