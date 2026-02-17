// biz/router/system/custom_routes.go - 手动添加的系统路由

package system

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	system "github.com/yi-nology/git-manage-service/biz/handler/system"
)

// RegisterCustomRoutes 注册自定义系统路由
func RegisterCustomRoutes(r *server.Hertz) {
	systemGroup := r.Group("/api/v1/system")
	{
		systemGroup.GET("/app-info", system.GetAppInfo)
	}
}
