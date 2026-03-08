// biz/router/branch/custom_routes.go - 手动添加的分支路由

package branch

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	branch "github.com/yi-nology/git-manage-service/biz/handler/branch"
)

// RegisterCustomRoutes 注册自定义路由（cherry-pick, rebase等）
func RegisterCustomRoutes(r *server.Hertz) {
	root := r.Group("/")
	{
		_api := root.Group("/api")
		{
			_v1 := _api.Group("/v1")
			{
				_branch := _v1.Group("/branch")
				{
					// Cherry-pick
					_branch.POST("/cherry-pick", branch.CherryPick)

					// Rebase
					_branch.POST("/rebase", branch.Rebase)

					// Rebase子路由
					rebaseGroup := _branch.Group("/rebase")
					{
						rebaseGroup.POST("/abort", branch.RebaseAbort)
						rebaseGroup.POST("/continue", branch.RebaseContinue)
					}
				}
			}
		}
	}
}
