package sync

import (
	"context"
	"fmt"
	"log"
	stdsync "sync"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/lock"

	"github.com/robfig/cron/v3"
)

type CronService struct {
	cron    *cron.Cron
	entries map[uint]cron.EntryID
	mu      stdsync.Mutex
	syncSvc *SyncService
	taskDAO *db.SyncTaskDAO
	lockSvc lock.DistLock
}

var CronSvc *CronService

func InitCronService() {
	CronSvc = &CronService{
		cron:    cron.New(),
		entries: make(map[uint]cron.EntryID),
		syncSvc: NewSyncService(),
		taskDAO: db.NewSyncTaskDAO(),
	}
	CronSvc.cron.Start()
	CronSvc.Reload()
}

// SetLockService 设置锁服务（用于依赖注入）
func (s *CronService) SetLockService(lockSvc lock.DistLock) {
	s.lockSvc = lockSvc
	// 同时设置给 syncSvc
	if s.syncSvc != nil {
		s.syncSvc.SetLockService(lockSvc)
	}
}

func (s *CronService) Reload() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear existing
	for _, id := range s.entries {
		s.cron.Remove(id)
	}
	s.entries = make(map[uint]cron.EntryID)

	tasks, err := s.taskDAO.FindEnabledWithCron()
	if err != nil {
		log.Println("Failed to load tasks:", err)
		return
	}

	for _, task := range tasks {
		if task.Cron == "" {
			continue
		}
		s.addTask(task)
	}
}

func (s *CronService) UpdateTask(task po.SyncTask) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove existing if any
	if id, ok := s.entries[task.ID]; ok {
		s.cron.Remove(id)
		delete(s.entries, task.ID)
	}

	if task.Enabled && task.Cron != "" {
		s.addTask(task)
	}
}

func (s *CronService) RemoveTask(taskID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, ok := s.entries[taskID]; ok {
		s.cron.Remove(id)
		delete(s.entries, taskID)
	}
}

func (s *CronService) addTask(task po.SyncTask) {
	taskID := task.ID
	taskKey := task.Key
	entryID, err := s.cron.AddFunc(task.Cron, func() {
		ctx := context.Background()

		// 使用分布式锁防止多实例重复执行
		if s.lockSvc != nil {
			lockKey := fmt.Sprintf("cron:task:%s", taskKey)
			success, lockErr := s.lockSvc.Up(ctx, lockKey, 10*time.Minute)
			if lockErr != nil {
				log.Printf("Cron Task %d lock error: %v", taskID, lockErr)
				return
			}
			if !success {
				log.Printf("Cron Task %d (Key: %s) skipped: another instance is running", taskID, taskKey)
				return
			}
			defer s.lockSvc.Down(ctx, lockKey)
		}

		log.Printf("Executing Cron Task %d (Key: %s)", taskID, taskKey)
		err := s.syncSvc.RunTaskWithTrigger(taskKey, po.TriggerSourceCron)
		if err != nil {
			log.Printf("Cron Task %d failed: %v", taskID, err)
		}
	})
	if err != nil {
		log.Printf("Failed to add cron for task %d: %v", task.ID, err)
		return
	}
	s.entries[task.ID] = entryID
	fmt.Printf("Added cron task %d: %s\n", task.ID, task.Cron)
}
