// biz/router/branch/custom_routes.go - 手动添加的分支路由

package branch

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	branch "github.com/yi-nology/git-manage-service/biz/handler/branch"
)

// RegisterCustomRoutes 注册自定义路由（cherry-pick, rebase等）
func RegisterCustomRoutes(r *server.Hertz) {
	branchGroup := r.Group("/api/v1/branch")
	{
		// Cherry-pick
		branchGroup.POST("/cherry-pick", branch.CherryPick)

		// Rebase
		branchGroup.POST("/rebase", branch.Rebase)

		// Rebase子路由
		rebaseGroup := branchGroup.Group("/rebase")
		{
			rebaseGroup.POST("/abort", branch.RebaseAbort)
			rebaseGroup.POST("/continue", branch.RebaseContinue)
		}
	}
}
