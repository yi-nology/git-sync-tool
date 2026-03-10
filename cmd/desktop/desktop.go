package desktop

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	hserver "github.com/cloudwego/hertz/pkg/app/server"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/router"
	"github.com/yi-nology/git-manage-service/biz/service/audit"
	"github.com/yi-nology/git-manage-service/biz/service/stats"
	"github.com/yi-nology/git-manage-service/biz/service/sync"
	"github.com/yi-nology/git-manage-service/biz/utils"
	"github.com/yi-nology/git-manage-service/pkg/appinfo"
	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-manage-service/pkg/embed"
)

// App 应用结构
type App struct {
	ctx       context.Context
	version   string
	buildTime string
	gitCommit string
	hServer   *hserver.Hertz
}

var appInstance *App

// NewApp 创建新的应用实例
func NewApp(version, buildTime, gitCommit string) *App {
	if appInstance == nil {
		appInstance = &App{
			version:   version,
			buildTime: buildTime,
			gitCommit: gitCommit,
		}
	}
	return appInstance
}

// GetApp 获取应用实例
func GetApp() *App {
	if appInstance == nil {
		appInstance = NewApp(appinfo.Version, appinfo.BuildTime, appinfo.GitCommit)
	}
	return appInstance
}

// Startup 应用启动时调用
func Startup(ctx context.Context) {
	app := GetApp()
	app.ctx = ctx

	// 在后台异步启动后端服务（延迟 1 秒以确保 Wails 完成初始化）
	time.AfterFunc(1*time.Second, func() {
		go app.startBackend()
	})
}

// Shutdown 应用关闭时调用
func Shutdown(ctx context.Context) {
	log.Println("Application shutting down...")

	// 停止定时任务服务
	log.Println("Stopping cron service...")
	sync.StopCronService()

	// 停止 HTTP 服务器
	app := GetApp()
	if app.hServer != nil {
		log.Println("Stopping HTTP server...")
		if err := app.hServer.Close(); err != nil {
			log.Printf("Error stopping HTTP server: %v\n", err)
		} else {
			log.Println("HTTP server stopped successfully")
		}
	}
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

	// 先加载配置（此时 CWD 还在项目目录，能找到 conf/config.yaml）
	configs.Init()

	// 设置桌面应用的数据目录（会切换 CWD）
	if err := setupDesktopDataDir(); err != nil {
		log.Printf("Failed to setup data directory: %v\n", err)
		return
	}

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
	router.SetEmbedFS(embed.GetPublicFS(), embed.GetDocsFS())

	// 从配置文件读取端口号
	port := configs.GlobalConfig.Server.Port
	if isPortInUse(port) {
		log.Printf("Port %d is already in use", port)
		log.Println("Possible reasons:")
		log.Println("  1. Another instance is already running")
		log.Println("  2. Previous instance didn't exit properly")
		log.Printf("\nPlease check and close the process using port %d:\n", port)
		log.Printf("  lsof -i :%d\n", port)
		log.Printf("  kill -9 $(lsof -t -i :%d)\n", port)
		return
	}

	// 启动 HTTP 服务器
	log.Printf("Starting HTTP server on :%d...\n", port)
	a.hServer = hserver.Default(hserver.WithHostPorts(":" + strconv.Itoa(port)))

	// 注册路由
	router.GeneratedRegister(a.hServer)

	log.Println("Backend services started successfully")

	if err := a.hServer.Run(); err != nil {
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
	dataDir := filepath.Join(homeDir, "Library", "Application Support", appinfo.AppName)

	// 确保目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// 确保数据库文件的父目录存在
	dbPath := configs.GlobalConfig.Database.Path
	if dbPath == "" {
		dbPath = "git_sync.db"
	}
	dbDir := filepath.Dir(dbPath)
	if dbDir != "." && dbDir != "" {
		if err := os.MkdirAll(filepath.Join(dataDir, dbDir), 0755); err != nil {
			return fmt.Errorf("failed to create database directory: %w", err)
		}
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
	return "http://localhost:" + strconv.Itoa(configs.GlobalConfig.Server.Port)
}
