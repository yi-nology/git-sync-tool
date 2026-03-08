// biz/router/patch/custom_routes.go - Patch 路由

package patch

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	patch "github.com/yi-nology/git-manage-service/biz/handler/patch"
)

// RegisterCustomRoutes 注册自定义 patch 路由
func RegisterCustomRoutes(r *server.Hertz) {
	patchGroup := r.Group("/api/v1/patch")
	{
		patchGroup.POST("/generate", patch.GeneratePatch)
		patchGroup.POST("/save", patch.SavePatch)
		patchGroup.GET("/list", patch.ListPatches)
		patchGroup.GET("/content", patch.GetPatchContent)
		patchGroup.GET("/download", patch.DownloadPatch)
		patchGroup.POST("/apply", patch.ApplyPatch)
		patchGroup.POST("/check", patch.CheckPatch)
		patchGroup.POST("/delete", patch.DeletePatch)
	}
}
