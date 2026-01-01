package main

import (
	"context"

	"github.com/yi-nology/git-manage-service/biz/config"
	"github.com/yi-nology/git-manage-service/biz/dal"
	"github.com/yi-nology/git-manage-service/biz/handler"
	"github.com/yi-nology/git-manage-service/biz/service"
	"github.com/yi-nology/git-manage-service/biz/utils"

	_ "github.com/yi-nology/git-manage-service/docs"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	// "github.com/hertz-contrib/swagger"
	// swaggerFiles "github.com/swaggo/files"
)

// @title Branch Management Tool API
// @version 1.0
// @description API documentation for Branch Management Tool

// @host localhost:8080
// @BasePath /

func main() {
	// 0. Init Config
	config.Init()

	// 1. Init DB
	dal.Init()

	// 2. Init Cron
	service.InitCronService()
	service.InitStatsService()
	service.InitAuditService()
	utils.InitEncryption()

	// 3. Init Server
	h := server.Default(server.WithHostPorts(":8080"))

	// 4. Register Routes
	h.POST("/api/repos", handler.RegisterRepo)
	h.POST("/api/repos/scan", handler.ScanRepo) // New Scan Endpoint
	h.POST("/api/repos/clone", handler.CloneRepo)
	h.GET("/api/repos", handler.ListRepos)
	h.PUT("/api/repos/:id", handler.UpdateRepo)
	h.DELETE("/api/repos/:id", handler.DeleteRepo)

	h.GET("/api/tasks/:id", handler.GetCloneTask) // New Task Endpoint

	h.GET("/api/config", handler.GetConfig)
	h.POST("/api/config", handler.UpdateConfig)

	// System Routes
	h.GET("/api/system/dirs", handler.ListDirs)
	h.GET("/api/system/ssh-keys", handler.ListSSHKeys)
	h.POST("/api/git/test-connection", handler.TestConnection)

	h.POST("/api/sync/tasks", handler.CreateTask)
	h.GET("/api/sync/tasks", handler.ListTasks)
	h.GET("/api/sync/tasks/:id", handler.GetTask)
	h.PUT("/api/sync/tasks/:id", handler.UpdateTask)
	h.DELETE("/api/sync/tasks/:id", handler.DeleteTask)
	h.POST("/api/sync/run", handler.RunSync)
	h.POST("/api/sync/execute", handler.ExecuteSync) // New Ad-hoc Sync
	h.GET("/api/sync/history", handler.ListHistory)
	h.DELETE("/api/sync/history/:id", handler.DeleteHistory)

	// Stats Routes
	h.GET("/api/stats/branches", handler.ListBranches)
	h.GET("/api/stats/commits", handler.ListCommits)
	h.GET("/api/stats/analyze", handler.GetStats)
	h.GET("/api/stats/export/csv", handler.ExportStatsCSV)

	// Webhook
	h.POST("/api/webhooks/trigger", handler.HandleWebhookTrigger)

	// Audit Routes
	h.GET("/api/audit/logs", handler.ListAuditLogs)

	// Swagger JSON
	h.StaticFile("/docs/swagger.json", "./docs/swagger.json")
	h.Static("/docs", "./docs")

	// 5. Static Files (Frontend)
	h.Static("/", "./public")

	// Redirect root to index.html if needed, but Static usually handles index.html
	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.File("./public/index.html")
	})

	h.Spin()
}
