package router

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/yi-nology/git-manage-service/biz/handler"
)

// GeneratedRegister registers all routes
func GeneratedRegister(h *server.Hertz) {
	// API Group
	api := h.Group("/api")

	// Repo Routes
	repos := api.Group("/repos")
	{
		repos.POST("", handler.RegisterRepo)
		repos.POST("/scan", handler.ScanRepo)
		repos.POST("/clone", handler.CloneRepo)
		repos.GET("", handler.ListRepos)
		repos.GET("/:key", handler.GetRepo)
		repos.PUT("/:key", handler.UpdateRepo)
		repos.DELETE("/:key", handler.DeleteRepo)

		// Branch Sub-Routes
		repos.GET("/:key/branches", handler.ListRepoBranches)
		repos.POST("/:key/branches", handler.CreateBranch)
		repos.DELETE("/:key/branches/:name", handler.DeleteBranch)
		repos.PUT("/:key/branches/:name", handler.UpdateBranch)
		repos.POST("/:key/branches/:name/checkout", handler.CheckoutBranch)
		repos.POST("/:key/branches/:name/push", handler.PushBranch)
		repos.POST("/:key/branches/:name/pull", handler.PullBranch)

		// Tag Routes
		repos.POST("/:key/tags", handler.CreateTag)
		repos.GET("/:key/tags", handler.ListTags)

		// Version Route
		repos.GET("/:key/version", handler.GetVersionInfo)
		repos.GET("/:key/versions", handler.ListVersions)
		repos.GET("/:key/version/next", handler.GetNextVersions)

		// Workspace/Submit Routes
		repos.GET("/:key/status", handler.GetRepoStatus)
		repos.GET("/:key/git-config", handler.GetRepoGitConfig)
		repos.POST("/:key/submit", handler.SubmitChanges)

		// Merge Routes
		repos.GET("/:key/compare", handler.CompareBranches)
		repos.GET("/:key/diff", handler.GetDiffContent)
		repos.GET("/:key/merge/check", handler.MergeCheck)
		repos.POST("/:key/merge", handler.ExecuteMerge)
		repos.GET("/:key/patch", handler.GetPatch)
	}

	// Task Routes
	api.GET("/tasks/:id", handler.GetCloneTask)

	// Config Routes
	api.GET("/config", handler.GetConfig)
	api.POST("/config", handler.UpdateConfig)

	// System Routes
	system := api.Group("/system")
	{
		system.GET("/dirs", handler.ListDirs)
		system.GET("/ssh-keys", handler.ListSSHKeys)
	}
	api.POST("/git/test-connection", handler.TestConnection)

	// Sync Routes
	sync := api.Group("/sync")
	{
		sync.POST("/tasks", handler.CreateTask)
		sync.GET("/tasks", handler.ListTasks)
		sync.GET("/tasks/:key", handler.GetTask)
		sync.PUT("/tasks/:key", handler.UpdateTask)
		sync.DELETE("/tasks/:key", handler.DeleteTask)
		sync.POST("/run", handler.RunSync)
		sync.POST("/execute", handler.ExecuteSync)
		sync.GET("/history", handler.ListHistory)
		sync.DELETE("/history/:id", handler.DeleteHistory)
	}

	// Stats Routes
	stats := api.Group("/stats")
	{
		stats.GET("/branches", handler.ListBranches)
		stats.GET("/commits", handler.ListCommits)
		stats.GET("/analyze", handler.GetStats)
		stats.GET("/export/csv", handler.ExportStatsCSV)
	}

	// Audit Routes
	api.GET("/audit/logs", handler.ListAuditLogs)
	api.GET("/audit/logs/:id", handler.GetAuditLog)

	// Swagger JSON
	h.StaticFile("/docs/swagger.json", "./docs/swagger.json")
	h.Static("/docs", "./docs")

	// Static Files (Frontend)
	h.Static("/", "./public")

	// Redirect root to index.html
	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.File("./public/index.html")
	})
}
