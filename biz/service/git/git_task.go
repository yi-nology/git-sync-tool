package git

import (
	"log"
	"sync"
	"time"
)

// Task Manager for Async Clones
type TaskManager struct {
	tasks         sync.Map
	maxConcurrent int
	runningTasks  int
	mutex         sync.Mutex
	taskQueue     chan *Task
	cleanupTicker *time.Ticker
}

type Task struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"` // running, success, failed, queued
	Progress  []string  `json:"progress"`
	Error     string    `json:"error"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

var GlobalTaskManager = &TaskManager{
	maxConcurrent: 100,                    // 支持并发处理100个任务
	taskQueue:     make(chan *Task, 1000), // 任务队列容量
}

// Init 初始化任务管理器
func (tm *TaskManager) Init() {
	// 启动任务处理器
	go tm.processTaskQueue()

	// 启动清理定时器，每小时清理一次已完成的任务
	tm.cleanupTicker = time.NewTicker(time.Hour)
	go tm.cleanupTasks()
}

// processTaskQueue 处理任务队列
func (tm *TaskManager) processTaskQueue() {
	for task := range tm.taskQueue {
		tm.mutex.Lock()
		if tm.runningTasks >= tm.maxConcurrent {
			tm.mutex.Unlock()
			// 等待有任务完成
			time.Sleep(100 * time.Millisecond)
			tm.taskQueue <- task // 重新入队
			continue
		}
		tm.runningTasks++
		tm.mutex.Unlock()

		// 任务已经在AddTask中标记为running，这里不需要再修改状态
		log.Printf("[INFO] Starting task: %s", task.ID)
	}
}

// AddTask 添加任务
func (tm *TaskManager) AddTask(id string) *Task {
	t := &Task{
		ID:        id,
		Status:    "queued",
		Progress:  []string{},
		StartTime: time.Now(),
	}
	tm.tasks.Store(id, t)

	// 将任务加入队列
	tm.taskQueue <- t

	// 立即更新状态为running
	t.Status = "running"

	return t
}

// GetTask 获取任务
func (tm *TaskManager) GetTask(id string) (*Task, bool) {
	v, ok := tm.tasks.Load(id)
	if !ok {
		return nil, false
	}
	return v.(*Task), true
}

// AppendLog 追加任务日志
func (tm *TaskManager) AppendLog(id string, log string) {
	if v, ok := tm.tasks.Load(id); ok {
		t := v.(*Task)
		t.Progress = append(t.Progress, log)
	}
}

// UpdateStatus 更新任务状态
func (tm *TaskManager) UpdateStatus(id string, status string, errStr string) {
	if v, ok := tm.tasks.Load(id); ok {
		t := v.(*Task)
		t.Status = status
		t.Error = errStr
		t.EndTime = time.Now()

		// 如果任务完成，减少运行任务数
		if status == "success" || status == "failed" {
			tm.mutex.Lock()
			tm.runningTasks--
			tm.mutex.Unlock()
			log.Printf("[INFO] Task %s completed with status: %s", id, status)
		}
	}
}

// cleanupTasks 清理已完成的任务
func (tm *TaskManager) cleanupTasks() {
	for range tm.cleanupTicker.C {
		log.Printf("[INFO] Starting task cleanup")
		count := 0
		tm.tasks.Range(func(key, value interface{}) bool {
			task := value.(*Task)
			if task.Status == "success" || task.Status == "failed" {
				// 清理24小时前完成的任务
				if time.Since(task.EndTime) > 24*time.Hour {
					tm.tasks.Delete(key)
					count++
				}
			}
			return true
		})
		log.Printf("[INFO] Cleaned up %d completed tasks", count)
	}
}

// GetRunningTasksCount 获取当前运行的任务数
func (tm *TaskManager) GetRunningTasksCount() int {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	return tm.runningTasks
}

// GetQueueLength 获取任务队列长度
func (tm *TaskManager) GetQueueLength() int {
	return len(tm.taskQueue)
}
