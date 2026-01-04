package main

import (
	"context"
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
	"github.com/yi-nology/git-manage-service/biz/router"
	"github.com/yi-nology/git-manage-service/biz/rpc_handler"
	"github.com/yi-nology/git-manage-service/biz/service/audit"
	"github.com/yi-nology/git-manage-service/biz/service/stats"
	"github.com/yi-nology/git-manage-service/biz/service/sync"
	"github.com/yi-nology/git-manage-service/biz/utils"
	"github.com/yi-nology/git-manage-service/biz/kitex_gen/git/gitservice"
	"github.com/yi-nology/git-manage-service/pkg/configs"

	_ "github.com/yi-nology/git-manage-service/docs"
)

// @title Branch Management Tool API
// @version 1.1
// @description API documentation for Branch Management Tool.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.basic BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

func main() {
	// 1. 初始化共享资源 (Initialize Shared Resources)
	log.Println("Initializing resources...")
	configs.Init()
	db.Init()
	utils.InitEncryption()

	// 初始化业务服务 (Initialize Services)
	sync.InitCronService()
	stats.InitStatsService()
	audit.InitAuditService()

	// 2. 启动 gRPC Server (Kitex)
	rpcAddr := fmt.Sprintf(":%d", configs.GlobalConfig.Rpc.Port)
	addr, _ := net.ResolveTCPAddr("tcp", rpcAddr)
	svr := gitservice.NewServer(new(rpc_handler.GitServiceImpl), kserver.WithServiceAddr(addr))

	go func() {
		log.Printf("RPC Server starting on %s\n", rpcAddr)
		if err := svr.Run(); err != nil {
			log.Printf("RPC Server stopped with error: %v\n", err)
		}
	}()

	// 3. 启动 HTTP Server (Hertz)
	hAddr := fmt.Sprintf(":%d", configs.GlobalConfig.Server.Port)
	// 使用 WithHostPorts 配置监听地址
	h := hserver.Default(hserver.WithHostPorts(hAddr))
	router.GeneratedRegister(h)

	go func() {
		log.Printf("HTTP Server starting on %s\n", hAddr)
		// Hertz 的 Spin 会阻塞并处理信号，但在 goroutine 中我们让主线程控制退出
		if err := h.Run(); err != nil {
			log.Printf("HTTP Server stopped with error: %v\n", err)
		}
	}()

	// 4. 等待中断信号 (Wait for Shutdown Signal)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received, shutting down servers...")

	// 5. 优雅关闭 (Graceful Shutdown)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭 Kitex Server
	if err := svr.Stop(); err != nil {
		log.Printf("Kitex Shutdown error: %v\n", err)
	} else {
		log.Println("Kitex Server stopped")
	}

	// 关闭 Hertz Server
	if err := h.Shutdown(ctx); err != nil {
		log.Printf("Hertz Shutdown error: %v\n", err)
	} else {
		log.Println("Hertz Server stopped")
	}

	log.Println("All servers stopped. Exiting.")
}
