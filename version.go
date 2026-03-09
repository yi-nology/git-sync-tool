package main

// 版本信息变量，通过编译时 -ldflags 注入
var (
	Version   = "dev"     // 版本号，如 v1.0.0
	BuildTime = "unknown" // 构建时间
	GitCommit = "unknown" // Git commit hash
)
