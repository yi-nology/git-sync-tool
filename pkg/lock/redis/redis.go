package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLock Redis 分布式锁实现
type RedisLock struct {
	client *redis.Client
}

// NewRedisLock 创建 Redis 锁实例
func NewRedisLock(addr, password string, db int) (*RedisLock, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisLock{client: client}, nil
}

// Up 尝试获取锁（非阻塞）
// 使用 SET NX EX 命令实现
func (r *RedisLock) Up(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	success, err := r.client.SetNX(ctx, key, 1, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}
	return success, nil
}

// Down 释放锁
func (r *RedisLock) Down(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	return nil
}

// UpWait 等待获取锁（阻塞）
func (r *RedisLock) UpWait(ctx context.Context, key string, ttl time.Duration, waitTimeout time.Duration) error {
	deadline := time.Time{}
	if waitTimeout > 0 {
		deadline = time.Now().Add(waitTimeout)
	}

	for {
		// 检查 context 是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 尝试获取锁
		success, err := r.Up(ctx, key, ttl)
		if err != nil {
			return err
		}
		if success {
			return nil
		}

		// 检查是否超时
		if !deadline.IsZero() && time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for lock: %s", key)
		}

		// 等待一段时间后重试
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}

// Close 关闭 Redis 连接
func (r *RedisLock) Close() error {
	return r.client.Close()
}
