package memory

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// lockInfo 锁信息
type lockInfo struct {
	expireAt time.Time
}

// MemoryLock 本地内存锁实现
type MemoryLock struct {
	mu     sync.Mutex
	locks  map[string]*lockInfo
	stopCh chan struct{}
}

// NewMemoryLock 创建内存锁实例
func NewMemoryLock() *MemoryLock {
	m := &MemoryLock{
		locks:  make(map[string]*lockInfo),
		stopCh: make(chan struct{}),
	}
	// 启动后台清理过期锁的 goroutine
	go m.cleanupExpiredLocks()
	return m
}

// cleanupExpiredLocks 定期清理过期的锁
func (m *MemoryLock) cleanupExpiredLocks() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for key, info := range m.locks {
				if now.After(info.expireAt) {
					delete(m.locks, key)
				}
			}
			m.mu.Unlock()
		case <-m.stopCh:
			return
		}
	}
}

// Up 尝试获取锁（非阻塞）
func (m *MemoryLock) Up(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// 检查锁是否存在且未过期
	if info, exists := m.locks[key]; exists {
		if now.Before(info.expireAt) {
			return false, nil // 锁被占用
		}
		// 锁已过期，可以获取
	}

	// 设置新锁
	m.locks[key] = &lockInfo{
		expireAt: now.Add(ttl),
	}
	return true, nil
}

// Down 释放锁
func (m *MemoryLock) Down(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.locks, key)
	return nil
}

// UpWait 等待获取锁（阻塞）
func (m *MemoryLock) UpWait(ctx context.Context, key string, ttl time.Duration, waitTimeout time.Duration) error {
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
		success, err := m.Up(ctx, key, ttl)
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

// Close 关闭锁服务
func (m *MemoryLock) Close() error {
	close(m.stopCh)
	return nil
}
