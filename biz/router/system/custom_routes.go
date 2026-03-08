// biz/router/system/custom_routes.go - 手动添加的系统路由

package system

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	system "github.com/yi-nology/git-manage-service/biz/handler/system"
	repo "github.com/yi-nology/git-manage-service/biz/handler/repo"
)

// RegisterCustomRoutes 注册自定义系统路由
func RegisterCustomRoutes(r *server.Hertz) {
	systemGroup := r.Group("/api/v1/system")
	{
		systemGroup.GET("/app-info", system.GetAppInfo)
		systemGroup.POST("/select-directory", system.SelectDirectory)
	}

	repoGroup := r.Group("/api/v1/repo")
	{
		repoGroup.POST("/scan-directory", repo.ScanDirectory)
		repoGroup.POST("/batch-create", repo.BatchCreate)
	}
}
