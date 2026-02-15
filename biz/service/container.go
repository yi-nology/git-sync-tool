package service

import (
	stdsync "sync"

	"github.com/yi-nology/git-manage-service/biz/service/branch"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	syncsvc "github.com/yi-nology/git-manage-service/biz/service/sync"
)

// Container 服务容器，用于依赖注入
type Container struct {
	GitService    *git.GitService
	BranchService *branch.BranchService
	SyncService   *syncsvc.SyncService
}

var (
	container *Container
	once      stdsync.Once
)

// GetContainer 获取服务容器单例
func GetContainer() *Container {
	once.Do(func() {
		container = &Container{
			GitService:    git.NewGitService(),
			BranchService: branch.NewBranchService(),
			SyncService:   syncsvc.NewSyncService(),
		}
	})
	return container
}

// Git 获取 Git 服务
func Git() *git.GitService {
	return GetContainer().GitService
}

// Branch 获取分支服务
func Branch() *branch.BranchService {
	return GetContainer().BranchService
}

// Sync 获取同步服务
func Sync() *syncsvc.SyncService {
	return GetContainer().SyncService
}
