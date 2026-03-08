package spec

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	spec_handler "github.com/yi-nology/git-manage-service/biz/handler/spec"
)

// RegisterCustomRoutes 注册 spec 相关路由
func RegisterCustomRoutes(h *server.Hertz) {
	root := h.Group("/")
	{
		_api := root.Group("/api")
		{
			_v1 := _api.Group("/v1")
			{
				_spec := _v1.Group("/spec")
				{
					// 文件操作 API
					_spec.GET("/tree", spec_handler.GetSpecTree)
					_spec.GET("/list", spec_handler.ListSpecFiles)
					_spec.GET("/content", spec_handler.GetSpecContent)
					_spec.GET("/content/:path", spec_handler.GetSpecContentByPath)
					_spec.POST("/save", spec_handler.SaveSpecContent)
					_spec.PUT("/content/:path", spec_handler.SaveSpecContentByPath)

					// Linting API
					_spec.POST("/lint", spec_handler.LintSpec)
					_spec.GET("/rules", spec_handler.GetLintRules)
					_spec.PUT("/rules/:id", spec_handler.UpdateLintRule)
					_spec.POST("/rules", spec_handler.CreateLintRule)

					// Git 操作 API
					_spec.POST("/commit/:path", spec_handler.CommitSpec)

					// 兼容旧 API
					_spec.POST("/validate", spec_handler.ValidateSpec)
					_spec.POST("/create", spec_handler.CreateSpecFile)
					_spec.POST("/delete", spec_handler.DeleteSpecFile)
				}
			}
		}
	}
}
