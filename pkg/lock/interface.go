package lock

import (
	"context"
	"time"
)

// DistLock 分布式锁接口
type DistLock interface {
	// Up 尝试获取锁（非阻塞）
	// 返回 true 表示成功获取锁，false 表示锁已被其他持有者占用
	Up(ctx context.Context, key string, ttl time.Duration) (bool, error)

	// Down 释放锁
	Down(ctx context.Context, key string) error

	// UpWait 等待获取锁（阻塞）
	// 如果锁被其他持有者占用，会等待直到获取成功或超时
	// waitTimeout 为等待超时时间，0 表示一直等待
	UpWait(ctx context.Context, key string, ttl time.Duration, waitTimeout time.Duration) error

	// Close 关闭锁服务
	Close() error
}
