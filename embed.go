package main

import (
	"embed"
	"io/fs"
)

//go:embed public
var publicFS embed.FS

//go:embed docs
var docsFS embed.FS

// GetPublicFS 返回嵌入的前端资源（去除 "public" 前缀）
func GetPublicFS() fs.FS {
	sub, _ := fs.Sub(publicFS, "public")
	return sub
}

// GetDocsFS 返回嵌入的文档资源（去除 "docs" 前缀）
func GetDocsFS() fs.FS {
	sub, _ := fs.Sub(docsFS, "docs")
	return sub
}
