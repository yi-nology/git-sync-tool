package service

import (
	"log"
	stdsync "sync"

	"github.com/yi-nology/git-manage-service/biz/service/branch"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	syncsvc "github.com/yi-nology/git-manage-service/biz/service/sync"
	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-manage-service/pkg/lock"
	"github.com/yi-nology/git-manage-service/pkg/storage"
)

// Container 服务容器，用于依赖注入
type Container struct {
	GitService     *git.GitService
	BranchService  *branch.BranchService
	SyncService    *syncsvc.SyncService
	StorageService storage.Storage
	LockService    lock.DistLock
}

var (
	container *Container
	once      stdsync.Once
)

// GetContainer 获取服务容器单例
func GetContainer() *Container {
	once.Do(func() {
		// 初始化存储服务
		storageSvc, err := storage.NewStorage(configs.GlobalConfig.Storage)
		if err != nil {
			log.Printf("Warning: Failed to initialize storage service: %v, using local storage", err)
			storageSvc, _ = storage.NewStorage(configs.StorageConfig{Type: "local", LocalPath: "./storage"})
		}

		// 初始化锁服务
		lockSvc, err := lock.NewDistLock(configs.GlobalConfig.Lock)
		if err != nil {
			log.Printf("Warning: Failed to initialize lock service: %v, using memory lock", err)
			lockSvc, _ = lock.NewDistLock(configs.LockConfig{Type: "memory"})
		}

		container = &Container{
			GitService:     git.NewGitService(),
			BranchService:  branch.NewBranchService(),
			SyncService:    syncsvc.NewSyncService(),
			StorageService: storageSvc,
			LockService:    lockSvc,
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

// Storage 获取存储服务
func Storage() storage.Storage {
	return GetContainer().StorageService
}

// Lock 获取分布式锁服务
func Lock() lock.DistLock {
	return GetContainer().LockService
}
