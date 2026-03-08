// biz/router/patch/custom_routes.go - Patch 路由

package patch

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	patch "github.com/yi-nology/git-manage-service/biz/handler/patch"
)

// RegisterCustomRoutes 注册自定义 patch 路由
func RegisterCustomRoutes(r *server.Hertz) {
	root := r.Group("/")
	{
		_api := root.Group("/api")
		{
			_v1 := _api.Group("/v1")
			{
				_patch := _v1.Group("/patch")
				{
					_patch.POST("/generate", patch.GeneratePatch)
					_patch.POST("/save", patch.SavePatch)
					_patch.GET("/list", patch.ListPatches)
					_patch.GET("/content", patch.GetPatchContent)
					_patch.GET("/download", patch.DownloadPatch)
					_patch.POST("/apply", patch.ApplyPatch)
					_patch.POST("/check", patch.CheckPatch)
					_patch.POST("/delete", patch.DeletePatch)
				}
			}
		}
	}
}
