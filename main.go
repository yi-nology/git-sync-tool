package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/yi-nology/git-manage-service/cmd/desktop"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
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
		OnStartup:        desktop.Startup,
		OnShutdown:       desktop.Shutdown,
		Bind: []interface{}{
			desktop.GetApp(),
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
