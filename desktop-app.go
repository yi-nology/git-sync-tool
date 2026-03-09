// +build desktop

package main

import (
	"context"
	"fmt"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// 启动后台服务
	go func() {
		if err := startBackendServices(); err != nil {
			fmt.Printf("Failed to start backend services: %v\n", err)
		}
	}()
}

// GetVersion returns the application version
func (a *App) GetVersion() string {
	return Version
}

// GetBackendURL returns the backend API URL
func (a *App) GetBackendURL() string {
	return "http://localhost:38080"
}
