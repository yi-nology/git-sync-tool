//go:build desktop

package main

import (
	"context"
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	hserver "github.com/cloudwego/hertz/pkg/app/server"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/router"
	"github.com/yi-nology/git-manage-service/biz/service/audit"
	"github.com/yi-nology/git-manage-service/biz/service/stats"
	"github.com/yi-nology/git-manage-service/biz/service/sync"
	"github.com/yi-nology/git-manage-service/biz/utils"
	"github.com/yi-nology/git-manage-service/pkg/configs"
)

//go:embed all:frontend/dist
var assets embed.FS

// App 应用结构
type App struct {
	ctx       context.Context
	version   string
	buildTime string
	gitCommit string
}

// NewApp 创建新的应用实例
func NewApp(version, buildTime, gitCommit string) *App {
	return &App{
		version:   version,
		buildTime: buildTime,
		gitCommit: gitCommit,
	}
}

// startup 应用启动时调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// 在后台启动后端服务
	go a.startBackend()
}

// shutdown 应用关闭时调用
func (a *App) shutdown(ctx context.Context) {
	log.Println("Application shutting down...")
}

// startBackend 启动后端服务
func (a *App) startBackend() {
	// 检查是否在构建时（通过环境变量或命令行参数判断）
	// 在生成绑定时跳过后端初始化
	if os.Getenv("WAILS_BUILD") != "" {
		log.Println("Skipping backend initialization during build")
		return
	}
	
	log.Println("Initializing backend services...")
	
	// 加载配置
	configs.Init()
	
	// 初始化数据库
	db.Init()
	db.InitLintRules()
	
	// 初始化加密工具
	utils.InitEncryption()
	
	// 初始化业务服务
	sync.InitCronService()
	stats.InitStatsService()
	audit.InitAuditService()
	
	// 启动 HTTP 服务器
	log.Println("Starting HTTP server on :38080...")
	hServer := hserver.Default(hserver.WithHostPorts(":38080"))
	
	// 注册路由
	router.GeneratedRegister(hServer)
	
	log.Println("Backend services started successfully")
	
	if err := hServer.Run(); err != nil {
		log.Printf("HTTP server error: %v\n", err)
	}
}

// GetVersion 获取版本信息
func (a *App) GetVersion() string {
	return a.version
}

// GetBuildTime 获取构建时间
func (a *App) GetBuildTime() string {
	return a.buildTime
}

// GetGitCommit 获取 Git 提交
func (a *App) GetGitCommit() string {
	return a.gitCommit
}

// GetBackendURL 获取后端 URL
func (a *App) GetBackendURL() string {
	return "http://localhost:38080"
}

func main() {
	// 设置应用信息
	app := NewApp(Version, BuildTime, GitCommit)
	
	// 创建 Wails 应用
	err := wails.Run(&options.App{
		Title:     "Git Manage Service",
		Width:     1280,
		Height:    800,
		MinWidth:  1024,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		// 启用调试模式（生产环境可关闭）
		Debug: options.Debug{
			OpenInspectorOnStartup: false,
		},
	})

	if err != nil {
		log.Fatal("Error starting application:", err)
	}
}
