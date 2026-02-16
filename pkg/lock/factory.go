package lock

import (
	"fmt"

	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-manage-service/pkg/lock/memory"
	"github.com/yi-nology/git-manage-service/pkg/lock/redis"
)

// NewDistLock 根据配置创建分布式锁实例
func NewDistLock(cfg configs.LockConfig) (DistLock, error) {
	switch cfg.Type {
	case "redis":
		if cfg.RedisAddr == "" {
			return nil, fmt.Errorf("redis_addr is required for redis lock")
		}
		return redis.NewRedisLock(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	case "memory", "":
		return memory.NewMemoryLock(), nil
	default:
		return nil, fmt.Errorf("unsupported lock type: %s", cfg.Type)
	}
}
