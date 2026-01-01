package service

import (
	"fmt"
	"github.com/yi-nology/git-manage-service/biz/dal"
	"github.com/yi-nology/git-manage-service/biz/model"
	"log"
	"sync"

	"github.com/robfig/cron/v3"
)

type CronService struct {
	cron    *cron.Cron
	entries map[uint]cron.EntryID
	mu      sync.Mutex
	syncSvc *SyncService
}

var CronSvc *CronService

func InitCronService() {
	CronSvc = &CronService{
		cron:    cron.New(),
		entries: make(map[uint]cron.EntryID),
		syncSvc: NewSyncService(),
	}
	CronSvc.cron.Start()
	CronSvc.Reload()
}

func (s *CronService) Reload() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear existing
	for _, id := range s.entries {
		s.cron.Remove(id)
	}
	s.entries = make(map[uint]cron.EntryID)

	var tasks []model.SyncTask
	if err := dal.DB.Where("enabled = ?", true).Find(&tasks).Error; err != nil {
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

func (s *CronService) UpdateTask(task model.SyncTask) {
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

func (s *CronService) addTask(task model.SyncTask) {
	taskID := task.ID
	entryID, err := s.cron.AddFunc(task.Cron, func() {
		log.Printf("Executing Cron Task %d", taskID)
		err := s.syncSvc.RunTask(taskID)
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
