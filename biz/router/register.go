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
	"github.com/yi-nology/git-manage-service/biz/router/credential"
	"github.com/yi-nology/git-manage-service/biz/router/file"
	"github.com/yi-nology/git-manage-service/biz/router/notification"
	"github.com/yi-nology/git-manage-service/biz/router/patch"
	"github.com/yi-nology/git-manage-service/biz/router/repo"
	"github.com/yi-nology/git-manage-service/biz/router/spec"
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

// 嵌入的文件系统变量
var (
	embeddedPublicFS fs.FS
	embeddedDocsFS   fs.FS
)

// SetEmbedFS 设置嵌入的文件系统（由 main.go 调用）
func SetEmbedFS(public, docs fs.FS) {
	embeddedPublicFS = public
	embeddedDocsFS = docs
}

// GeneratedRegister registers all routes
func GeneratedRegister(h *server.Hertz) {
	// 注册 Hz 生成的路由
	audit.Register(h)
	branch.Register(h)
	commit.Register(h)
	credential.Register(h)
	file.Register(h)
	notification.Register(h)
	repo.Register(h)
	sshkey.Register(h)
	stash.Register(h)
	submodule.Register(h)
	sync.Register(h)
	system.Register(h)
	tag.Register(h)
	version.Register(h)
	webhook.Register(h)
	patch.Register(h)
	spec.Register(h)
	stats.Register(h)

	// 前端 SPA - 从嵌入的 FS 读取
	// 注意：这个路由应该在所有 API 路由之后注册，并且只匹配非 API 路径
	h.GET("/*filepath", func(ctx context.Context, c *app.RequestContext) {
		fp := c.Param("filepath")

		// 跳过 API 路径，让 API 路由处理
		if strings.HasPrefix(fp, "api/") {
			c.Next(ctx)
			return
		}

		if fp == "" {
			fp = "index.html"
		}
		// 尝试读取文件，失败则回退到 index.html (SPA)
		if embeddedPublicFS != nil {
			if !serveEmbedFile(c, embeddedPublicFS, fp, "") {
				serveEmbedFile(c, embeddedPublicFS, "index.html", "text/html; charset=utf-8")
			}
		} else {
			c.String(http.StatusInternalServerError, "Embedded file system not initialized")
		}
	})

	h.HEAD("/*filepath", func(ctx context.Context, c *app.RequestContext) {
		fp := c.Param("filepath")

		// 跳过 API 路径，让 API 路由处理
		if strings.HasPrefix(fp, "api/") {
			c.Next(ctx)
			return
		}

		if fp == "" {
			fp = "index.html"
		}
		if embeddedPublicFS != nil {
			if !serveEmbedFile(c, embeddedPublicFS, fp, "") {
				serveEmbedFile(c, embeddedPublicFS, "index.html", "text/html; charset=utf-8")
			}
		} else {
			c.String(http.StatusInternalServerError, "Embedded file system not initialized")
		}
	})

	// 根路径重定向到 index.html
	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		if embeddedPublicFS != nil {
			serveEmbedFile(c, embeddedPublicFS, "index.html", "text/html; charset=utf-8")
		} else {
			c.String(http.StatusInternalServerError, "Embedded file system not initialized")
		}
	})
}

// serveEmbedFile 从嵌入的文件系统提供文件
func serveEmbedFile(c *app.RequestContext, fsys fs.FS, path string, contentType string) bool {
	// 清理路径
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		path = "index.html"
	}

	// 检查文件系统是否为 nil
	if fsys == nil {
		return false
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
