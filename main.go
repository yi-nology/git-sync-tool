package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	hserver "github.com/cloudwego/hertz/pkg/app/server"
	kserver "github.com/cloudwego/kitex/server"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/kitex_gen/git/gitservice"
	"github.com/yi-nology/git-manage-service/biz/router"
	"github.com/yi-nology/git-manage-service/biz/rpc_handler"
	"github.com/yi-nology/git-manage-service/biz/service/audit"
	"github.com/yi-nology/git-manage-service/biz/service/stats"
	"github.com/yi-nology/git-manage-service/biz/service/sync"
	"github.com/yi-nology/git-manage-service/biz/utils"
	"github.com/yi-nology/git-manage-service/pkg/appinfo"
	"github.com/yi-nology/git-manage-service/pkg/configs"

	_ "github.com/yi-nology/git-manage-service/docs"
)

// @title Git Manage Service API
// @version 2.0
// @description 轻量级多仓库、多分支自动化同步管理系统 API 文档

// @contact.name API Support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api

var (
	mode    = flag.String("mode", "all", "启动模式: http, rpc, all")
	version = flag.Bool("version", false, "显示版本信息")
)

var (
	// 这些变量在编译时通过 -ldflags 注入
	Version   = "dev"     // 版本号，如 v1.0.0
	BuildTime = "unknown" // 构建时间
	GitCommit = "unknown" // Git commit hash
)

const (
	AppName = "git-manage-service"
)

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("%s version %s\n", AppName, Version)
		fmt.Printf("Build time: %s\n", BuildTime)
		fmt.Printf("Git commit: %s\n", GitCommit)
		return
	}

	log.Printf("[%s] Starting in '%s' mode...\n", AppName, *mode)

	// 设置应用信息（供 API 使用）
	appinfo.Set(Version, BuildTime, GitCommit)

	// 初始化共享资源
	initResources()

	// 创建全局上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 信号处理
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	var httpServer *hserver.Hertz
	var rpcServer kserver.Server

	switch *mode {
	case "http":
		httpServer = startHTTPServer()
	case "rpc":
		rpcServer = startRPCServer()
	case "all":
		rpcServer = startRPCServer()
		httpServer = startHTTPServer()
	default:
		log.Fatalf("Unknown mode: %s. Available modes: http, rpc, all", *mode)
	}

	// 等待退出信号
	<-quit
	log.Println("Shutdown signal received, shutting down servers...")

	// 优雅关闭
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	if rpcServer != nil {
		if err := rpcServer.Stop(); err != nil {
			log.Printf("RPC Server shutdown error: %v\n", err)
		} else {
			log.Println("RPC Server stopped")
		}
	}

	if httpServer != nil {
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP Server shutdown error: %v\n", err)
		} else {
			log.Println("HTTP Server stopped")
		}
	}

	log.Println("All servers stopped. Exiting.")
}

// initResources 初始化共享资源
func initResources() {
	log.Println("Initializing resources...")

	// 加载配置
	configs.Init()

	// 初始化数据库
	db.Init()

	// 初始化加密工具
	utils.InitEncryption()

	// 初始化业务服务
	sync.InitCronService()
	stats.InitStatsService()
	audit.InitAuditService()

	log.Println("Resources initialized successfully")
}

// startHTTPServer 启动 HTTP 服务器
func startHTTPServer() *hserver.Hertz {
	// 注入嵌入的静态资源
	router.SetEmbedFS(GetPublicFS(), GetDocsFS())

	addr := fmt.Sprintf(":%d", configs.GlobalConfig.Server.Port)
	h := hserver.Default(hserver.WithHostPorts(addr))

	// 注册路由
	router.GeneratedRegister(h)

	go func() {
		log.Printf("HTTP Server starting on %s\n", addr)
		if err := h.Run(); err != nil {
			log.Printf("HTTP Server stopped with error: %v\n", err)
		}
	}()

	return h
}

// startRPCServer 启动 RPC 服务器
func startRPCServer() kserver.Server {
	addr := fmt.Sprintf(":%d", configs.GlobalConfig.Rpc.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to resolve RPC address: %v", err)
	}

	svr := gitservice.NewServer(
		new(rpc_handler.GitServiceImpl),
		kserver.WithServiceAddr(tcpAddr),
	)

	go func() {
		log.Printf("RPC Server starting on %s\n", addr)
		if err := svr.Run(); err != nil {
			log.Printf("RPC Server stopped with error: %v\n", err)
		}
	}()

	return svr
}
