package router

import (
	"context"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/yi-nology/git-manage-service/biz/router/audit"
	"github.com/yi-nology/git-manage-service/biz/router/branch"
	"github.com/yi-nology/git-manage-service/biz/router/commit"
	"github.com/yi-nology/git-manage-service/biz/router/file"
	"github.com/yi-nology/git-manage-service/biz/router/notification"
	"github.com/yi-nology/git-manage-service/biz/router/repo"
	"github.com/yi-nology/git-manage-service/biz/router/sshkey"
	"github.com/yi-nology/git-manage-service/biz/router/stash"
	"github.com/yi-nology/git-manage-service/biz/router/stats"
	"github.com/yi-nology/git-manage-service/biz/router/submodule"
	"github.com/yi-nology/git-manage-service/biz/router/sync"
	"github.com/yi-nology/git-manage-service/biz/router/system"
	"github.com/yi-nology/git-manage-service/biz/router/tag"
	"github.com/yi-nology/git-manage-service/biz/router/version"
	"github.com/yi-nology/git-manage-service/biz/router/webhook"
)

var (
	embeddedPublicFS fs.FS // 嵌入的前端资源
	embeddedDocsFS   fs.FS // 嵌入的文档资源
)

// SetEmbedFS 设置嵌入的文件系统（由 main.go 调用）
func SetEmbedFS(public, docs fs.FS) {
	embeddedPublicFS = public
	embeddedDocsFS = docs
}

// GeneratedRegister registers all routes
func GeneratedRegister(h *server.Hertz) {
	// 注册各模块路由（/api/v1）
	repo.Register(h)
	branch.Register(h)
	branch.RegisterCustomRoutes(h) // 注册自定义分支路由（cherry-pick, rebase等）
	tag.Register(h)
	version.Register(h)
	system.Register(h)
	system.RegisterCustomRoutes(h) // 注册自定义系统路由（app-info等）
	sync.Register(h)
	stats.Register(h)
	audit.Register(h)
	webhook.Register(h)
	file.Register(h)
	commit.Register(h)
	notification.Register(h)
	stash.Register(h)
	submodule.Register(h)
	sshkey.Register(h) // SSH密钥管理路由

	// Swagger 文档 - 从嵌入的 FS 读取
	h.GET("/docs/swagger.json", func(ctx context.Context, c *app.RequestContext) {
		serveEmbedFile(c, embeddedDocsFS, "swagger.json", "application/json")
	})
	h.GET("/docs/*filepath", func(ctx context.Context, c *app.RequestContext) {
		fp := c.Param("filepath")
		serveEmbedFile(c, embeddedDocsFS, fp, "")
	})

	// 前端 SPA - 从嵌入的 FS 读取
	h.GET("/*filepath", func(ctx context.Context, c *app.RequestContext) {
		fp := c.Param("filepath")
		if fp == "" {
			fp = "index.html"
		}
		// 尝试读取文件，失败则回退到 index.html (SPA)
		if !serveEmbedFile(c, embeddedPublicFS, fp, "") {
			serveEmbedFile(c, embeddedPublicFS, "index.html", "text/html; charset=utf-8")
		}
	})

	h.HEAD("/*filepath", func(ctx context.Context, c *app.RequestContext) {
		fp := c.Param("filepath")
		if fp == "" {
			fp = "index.html"
		}
		if !serveEmbedFile(c, embeddedPublicFS, fp, "") {
			serveEmbedFile(c, embeddedPublicFS, "index.html", "text/html; charset=utf-8")
		}
	})

	// 根路径重定向到 index.html
	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		serveEmbedFile(c, embeddedPublicFS, "index.html", "text/html; charset=utf-8")
	})
}

// serveEmbedFile 从嵌入的文件系统提供文件
func serveEmbedFile(c *app.RequestContext, fsys fs.FS, path string, contentType string) bool {
	// 清理路径
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		path = "index.html"
	}

	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return false
	}

	if contentType == "" {
		contentType = getContentType(path)
	}
	c.Data(http.StatusOK, contentType, data)
	return true
}

// getContentType 根据文件扩展名返回 MIME 类型
func getContentType(path string) string {
	ext := filepath.Ext(path)
	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		return mimeType
	}
	// 常用类型的后备
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}
