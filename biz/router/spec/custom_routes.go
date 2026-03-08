package spec

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	spec_handler "github.com/yi-nology/git-manage-service/biz/handler/spec"
)

// RegisterCustomRoutes 注册 spec 相关路由
func RegisterCustomRoutes(h *server.Hertz) {
	specGroup := h.Group("/api/v1/spec")
	{
		// 文件操作 API
		specGroup.GET("/tree", spec_handler.GetSpecTree)
		specGroup.GET("/list", spec_handler.ListSpecFiles)
		specGroup.GET("/content", spec_handler.GetSpecContent)
		specGroup.GET("/content/:path", spec_handler.GetSpecContentByPath)
		specGroup.POST("/save", spec_handler.SaveSpecContent)
		specGroup.PUT("/content/:path", spec_handler.SaveSpecContentByPath)

		// Linting API
		specGroup.POST("/lint", spec_handler.LintSpec)
		specGroup.GET("/rules", spec_handler.GetLintRules)
		specGroup.PUT("/rules/:id", spec_handler.UpdateLintRule)
		specGroup.POST("/rules", spec_handler.CreateLintRule)

		// Git 操作 API
		specGroup.POST("/commit/:path", spec_handler.CommitSpec)

		// 兼容旧 API
		specGroup.POST("/validate", spec_handler.ValidateSpec)
		specGroup.POST("/create", spec_handler.CreateSpecFile)
		specGroup.POST("/delete", spec_handler.DeleteSpecFile)
	}
}
