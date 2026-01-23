package router

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/yi-nology/git-manage-service/biz/router/audit"
	"github.com/yi-nology/git-manage-service/biz/router/branch"
	"github.com/yi-nology/git-manage-service/biz/router/repo"
	"github.com/yi-nology/git-manage-service/biz/router/stats"
	"github.com/yi-nology/git-manage-service/biz/router/sync"
	"github.com/yi-nology/git-manage-service/biz/router/system"
	"github.com/yi-nology/git-manage-service/biz/router/tag"
	"github.com/yi-nology/git-manage-service/biz/router/version"
)

// GeneratedRegister registers all routes
func GeneratedRegister(h *server.Hertz) {
	// 注册各模块路由（/api/v1）
	repo.Register(h)
	branch.Register(h)
	tag.Register(h)
	version.Register(h)
	system.Register(h)
	sync.Register(h)
	stats.Register(h)
	audit.Register(h)

	// 静态资源
	h.StaticFile("/docs/swagger.json", "./docs/swagger.json")
	h.Static("/docs", "./docs")

	// Static Files (Frontend)
	h.Static("/", "./public")

	// Redirect root to index.html
	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.File("./public/index.html")
	})
}
