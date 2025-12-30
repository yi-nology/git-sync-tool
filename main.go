package main

import (
	"context"

	"github.com/yi-nology/git-sync-tool/biz/config"
	"github.com/yi-nology/git-sync-tool/biz/dal"
	"github.com/yi-nology/git-sync-tool/biz/handler"
	"github.com/yi-nology/git-sync-tool/biz/middleware"
	"github.com/yi-nology/git-sync-tool/biz/service"

	_ "github.com/yi-nology/git-sync-tool/docs"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	// "github.com/hertz-contrib/swagger"
	// swaggerFiles "github.com/swaggo/files"
)

// @title Git Sync Tool API
// @version 1.0
// @description API documentation for Git Sync Tool

// @host localhost:8080
// @BasePath /

func main() {
	// 0. Init Config
	config.Init()

	// 1. Init DB
	dal.Init()

	// 2. Init Cron
	service.InitCronService()

	// 3. Init Server
	h := server.Default(server.WithHostPorts(":8080"))

	// 4. Register Routes
	h.POST("/api/repos", handler.RegisterRepo)
	h.GET("/api/repos", handler.ListRepos)

	h.GET("/api/config", handler.GetConfig)
	h.POST("/api/config", handler.UpdateConfig)

	h.POST("/api/sync/tasks", handler.CreateTask)
	h.GET("/api/sync/tasks", handler.ListTasks)
	h.GET("/api/sync/tasks/:id", handler.GetTask)
	h.PUT("/api/sync/tasks/:id", handler.UpdateTask)
	h.POST("/api/sync/run", handler.RunSync)
	h.GET("/api/sync/history", handler.ListHistory)

	// Webhook
	h.POST("/api/webhooks/task-sync", middleware.WebhookAuth(), handler.HandleWebhookTrigger)

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
