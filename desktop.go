//go:build desktop

package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

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
	
	// 在后台异步启动后端服务（延迟 1 秒以确保 Wails 完成初始化）
	time.AfterFunc(1*time.Second, func() {
		go a.startBackend()
	})
}

// shutdown 应用关闭时调用
func (a *App) shutdown(ctx context.Context) {
	log.Println("Application shutting down...")
}

// isPortInUse 检测端口是否被占用
func isPortInUse(port int) bool {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return true
	}
	listener.Close()
	return false
}

// startBackend 启动后端服务
func (a *App) startBackend() {
	log.Println("Initializing backend services...")
	
	// 设置桌面应用的数据目录
	if err := setupDesktopDataDir(); err != nil {
		log.Printf("Failed to setup data directory: %v\n", err)
		return
	}
	
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
	
	// 设置嵌入的文件系统（供 API 路由使用）
	router.SetEmbedFS(GetPublicFS(), GetDocsFS())
	
	// 检测端口是否被占用
	port := 38080
	if isPortInUse(port) {
		log.Printf("Port %d is already in use", port)
		log.Println("Possible reasons:")
		log.Println("  1. Another instance is already running")
		log.Println("  2. Previous instance didn't exit properly")
		log.Println("\nPlease check and close the process using port 38080:")
		log.Println("  lsof -i :38080")
		log.Println("  kill -9 $(lsof -t -i :38080)")
		return
	}
	
	// 启动 HTTP 服务器
	log.Printf("Starting HTTP server on :%d...\n", port)
	hServer := hserver.Default(hserver.WithHostPorts(":38080"))
	
	// 注册路由
	router.GeneratedRegister(hServer)
	
	log.Println("Backend services started successfully")
	
	if err := hServer.Run(); err != nil {
		log.Printf("HTTP server error: %v\n", err)
	}
}

// setupDesktopDataDir 设置桌面应用的数据目录
func setupDesktopDataDir() error {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	
	// 设置应用数据目录（macOS: ~/Library/Application Support/Git Manage Service/）
	dataDir := filepath.Join(homeDir, "Library", "Application Support", "Git Manage Service")
	
	// 确保目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	
	// 切换工作目录到数据目录
	if err := os.Chdir(dataDir); err != nil {
		return fmt.Errorf("failed to change working directory: %w", err)
	}
	
	log.Printf("Application data directory: %s\n", dataDir)
	log.Printf("Database will be stored at: %s/git_sync.db\n", dataDir)
	log.Printf("Config file will be stored at: %s/config.yaml\n", dataDir)
	
	return nil
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
