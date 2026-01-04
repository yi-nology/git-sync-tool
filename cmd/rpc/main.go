package main

import (
	"fmt"
	"log"
	"net"

	"github.com/cloudwego/kitex/server"
	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/rpc_handler"
	"github.com/yi-nology/git-manage-service/biz/utils"
	"github.com/yi-nology/git-manage-service/biz/kitex_gen/git/gitservice"
)

func main() {
	// 0. Init Config
	configs.Init()

	// 1. Init DB
	db.Init()

	// 2. Init Utils
	utils.InitEncryption()

	// 3. Start RPC Server
	rpcAddr := fmt.Sprintf(":%d", configs.GlobalConfig.Rpc.Port)
	addr, _ := net.ResolveTCPAddr("tcp", rpcAddr)
	svr := gitservice.NewServer(new(rpc_handler.GitServiceImpl), server.WithServiceAddr(addr))

	log.Printf("RPC Server starting on %s\n", rpcAddr)
	if err := svr.Run(); err != nil {
		log.Println("server stopped with error:", err)
	} else {
		log.Println("server stopped")
	}
}
