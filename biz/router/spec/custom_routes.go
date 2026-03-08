package spec

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	spec_handler "github.com/yi-nology/git-manage-service/biz/handler/spec"
)

// RegisterCustomRoutes 注册 spec 相关路由
func RegisterCustomRoutes(h *server.Hertz) {
	specGroup := h.Group("/api/v1/spec")
	{
		specGroup.GET("/list", spec_handler.ListSpecFiles)
		specGroup.GET("/content", spec_handler.GetSpecContent)
		specGroup.POST("/save", spec_handler.SaveSpecContent)
		specGroup.POST("/validate", spec_handler.ValidateSpec)
		specGroup.GET("/rules", spec_handler.GetSpecRules)
		specGroup.POST("/create", spec_handler.CreateSpecFile)
		specGroup.POST("/delete", spec_handler.DeleteSpecFile)
	}
}
