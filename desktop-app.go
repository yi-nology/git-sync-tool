//go:build desktop

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/router"
	"github.com/yi-nology/git-manage-service/biz/service/audit"
)

// App struct
type App struct {
	ctx       context.Context
	serverCtx context.CancelFunc
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// 启动后端服务
	go a.startBackend()
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	if a.serverCtx != nil {
		a.serverCtx()
	}
}

// startBackend 启动后端 HTTP 服务
func (a *App) startBackend() {
	// 创建取消上下文
	serverCtx, cancel := context.WithCancel(context.Background())
	a.serverCtx = cancel
	
	// 初始化资源
	fmt.Println("Initializing resources...")
	if err := initResources(); err != nil {
		fmt.Printf("Failed to initialize resources: %v\n", err)
		return
	}
	
	// 启动 HTTP 服务器
	fmt.Println("Starting HTTP server on :38080...")
	hServer := router.NewServer()
	
	// 优雅关闭
	go func() {
		<-serverCtx.Done()
		fmt.Println("Shutting down server...")
		hServer.Shutdown(nil)
	}()
	
	// 启动服务
	if err := hServer.Run(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

// GetVersion returns the application version
func (a *App) GetVersion() string {
	return Version
}

// GetBackendURL returns the backend API URL
func (a *App) GetBackendURL() string {
	return "http://localhost:38080"
}

// OpenInBrowser opens the backend URL in system browser
func (a *App) OpenInBrowser() {
	// 这个方法会被前端调用
	fmt.Println("Backend is running at:", a.GetBackendURL())
}
