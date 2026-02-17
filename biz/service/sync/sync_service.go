package sync

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/auth"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	notificationSvc "github.com/yi-nology/git-manage-service/biz/service/notification"
	"github.com/yi-nology/git-manage-service/pkg/lock"
)

type SyncService struct {
	git         *git.GitService
	authSvc     *auth.AuthService
	syncTaskDAO *db.SyncTaskDAO
	syncRunDAO  *db.SyncRunDAO
	lockSvc     lock.DistLock
}

func NewSyncService() *SyncService {
	return &SyncService{
		git:         git.NewGitService(),
		authSvc:     auth.NewAuthService(),
		syncTaskDAO: db.NewSyncTaskDAO(),
		syncRunDAO:  db.NewSyncRunDAO(),
	}
}

// SetLockService 设置锁服务（用于依赖注入）
func (s *SyncService) SetLockService(lockSvc lock.DistLock) {
	s.lockSvc = lockSvc
}

func (s *SyncService) RunTask(taskKey string) error {
	return s.RunTaskWithTrigger(taskKey, po.TriggerSourceManual)
}

func (s *SyncService) RunTaskWithTrigger(taskKey string, triggerSource string) error {
	task, err := s.syncTaskDAO.FindByKey(taskKey)
	if err != nil {
		return err
	}
	return s.ExecuteSyncWithTrigger(task, triggerSource)
}

func (s *SyncService) ExecuteSync(task *po.SyncTask) error {
	return s.ExecuteSyncWithTrigger(task, po.TriggerSourceManual)
}

func (s *SyncService) ExecuteSyncWithTrigger(task *po.SyncTask, triggerSource string) error {
	ctx := context.Background()

	// 获取分布式锁保护同步任务
	if s.lockSvc != nil {
		lockKey := fmt.Sprintf("sync:task:%s", task.Key)
		if err := s.lockSvc.UpWait(ctx, lockKey, 5*time.Minute, 30*time.Second); err != nil {
			return fmt.Errorf("failed to acquire lock for task %s: %w", task.Key, err)
		}
		defer s.lockSvc.Down(ctx, lockKey)
	}

	run := po.SyncRun{
		TaskKey:       task.Key,
		TriggerSource: triggerSource,
		StartTime:     time.Now(),
		Status:        "running",
	}
	s.syncRunDAO.Create(&run)

	repoPath := task.SourceRepo.Path

	// Capture logs
	var logs strings.Builder
	logf := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		logs.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05"), msg))
	}

	// 根据同步模式选择执行方法
	syncMode := task.SyncMode
	if syncMode == "" {
		syncMode = "single"
	}

	var commitRange string
	var err error
	if syncMode == "all-branch" {
		commitRange, err = s.doSyncAllBranches(repoPath, task, logf)
	} else {
		commitRange, err = s.doSyncSingleBranch(repoPath, task, logf)
	}

	run.CommitRange = commitRange
	run.Details = logs.String()
	run.EndTime = time.Now()

	if err != nil {
		run.Status = "failed"
		// Check if it was conflict
		if err.Error() == "conflict" {
			run.Status = "conflict"
		}
		run.ErrorMessage = err.Error()
		logf("Sync failed: %v", err)
	} else {
		run.Status = "success"
		logf("Sync completed successfully")
	}
	// Save final details
	run.Details = logs.String()
	s.syncRunDAO.Save(&run)

	// 发送通知
	s.sendNotification(task, &run)

	return err
}

// getAuthInfoForRemote 获取指定远程的认证信息
func getAuthInfoForRemote(repo po.Repo, remoteName string) domain.AuthInfo {
	if repo.RemoteAuths != nil {
		if authInfo, ok := repo.RemoteAuths[remoteName]; ok {
			return authInfo
		}
	}
	return domain.AuthInfo{
		Type:   repo.AuthType,
		Key:    repo.AuthKey,
		Secret: repo.AuthSecret,
		Source: "local",
	}
}

// resolveAuthForRemote 解析指定远程的认证方法
func (s *SyncService) resolveAuthForRemote(repo po.Repo, remoteName string) (transport.AuthMethod, error) {
	authInfo := getAuthInfoForRemote(repo, remoteName)
	return s.authSvc.ResolveAuth(authInfo)
}

func (s *SyncService) doSyncSingleBranch(path string, task *po.SyncTask, logf func(string, ...interface{})) (string, error) {
	logf("Starting sync for task %s (Repo: %s)", task.Key, path)

	// 1. Fetch Source
	sourceRemote := task.SourceRemote
	if sourceRemote == "" {
		sourceRemote = "origin"
	}

	// Check if source is local
	isLocalSource := (sourceRemote == "local")

	var sourceHash string

	// Helper for progress logging
	progressWriter := &logWriter{logf: logf}

	if !isLocalSource {
		sourceURL, _ := s.git.GetRemoteURL(path, sourceRemote)
		if sourceURL == "" && sourceRemote == "origin" {
			sourceURL = task.SourceRepo.RemoteURL
		}

		// 解析源仓库认证
		sourceAuth, err := s.resolveAuthForRemote(task.SourceRepo, sourceRemote)
		if err != nil {
			logf("Warning: failed to resolve source auth: %v", err)
		}
		sourceAuthInfo := getAuthInfoForRemote(task.SourceRepo, sourceRemote)
		sRefSpec := fmt.Sprintf("+refs/heads/%s:refs/remotes/%s/%s", task.SourceBranch, sourceRemote, task.SourceBranch)

		// Log Fetch Command (Approximate)
		fetchCmd := fmt.Sprintf("git fetch %s %s", sourceRemote, sRefSpec)
		logf("Command: %s", fetchCmd)

		if sourceURL != "" && sourceAuthInfo.Type != "" && sourceAuthInfo.Type != "none" {
			logf("Fetching source %s (Auth: %s, Source: %s)...", sourceRemote, sourceAuthInfo.Type, sourceAuthInfo.Source)
			// 如果使用数据库SSH密钥，使用原生git命令（更可靠）
			if sourceAuthInfo.Source == "database" && sourceAuthInfo.SSHKeyID > 0 {
				privateKey, passphrase, keyErr := s.authSvc.GetDBSSHKeyContent(sourceAuthInfo.SSHKeyID)
				if keyErr != nil {
					return "", fmt.Errorf("failed to load source SSH key: %v", keyErr)
				}
				if err := s.git.FetchWithDBKey(path, sourceURL, privateKey, passphrase, progressWriter, sRefSpec); err != nil {
					return "", fmt.Errorf("fetch source failed: %v", err)
				}
			} else if err := s.git.FetchWithAuthMethod(path, sourceURL, sourceAuth, progressWriter, sRefSpec); err != nil {
				return "", fmt.Errorf("fetch source failed: %v", err)
			}
		} else {
			logf("Fetching source %s...", sourceRemote)
			if err := s.git.Fetch(path, sourceRemote, progressWriter); err != nil {
				return "", fmt.Errorf("fetch source failed: %v", err)
			}
		}

		// Get Hash from Remote Ref
		h, err := s.git.GetCommitHash(path, task.SourceRemote, task.SourceBranch)
		if err != nil {
			return "", fmt.Errorf("get source hash failed: %v", err)
		}
		sourceHash = h
	} else {
		// Local Source
		// Get Hash from Local Head
		logf("Using local branch: %s", task.SourceBranch)
		h, err := s.git.ResolveRevision(path, task.SourceBranch)
		if err != nil {
			return "", fmt.Errorf("get local source hash failed: %v", err)
		}
		sourceHash = h
	}

	logf("Source hash (%s/%s): %s", task.SourceRemote, task.SourceBranch, sourceHash)

	// 2. Fetch Target
	targetRemote := task.TargetRemote
	if targetRemote == "" {
		targetRemote = "origin"
	}

	targetURL, _ := s.git.GetRemoteURL(path, targetRemote)
	if targetURL == "" && targetRemote == "origin" {
		targetURL = task.TargetRepo.RemoteURL
	}

	// 解析目标仓库认证
	targetAuth, err := s.resolveAuthForRemote(task.TargetRepo, targetRemote)
	if err != nil {
		logf("Warning: failed to resolve target auth: %v", err)
	}
	targetAuthInfo := getAuthInfoForRemote(task.TargetRepo, targetRemote)
	tRefSpec := fmt.Sprintf("+refs/heads/%s:refs/remotes/%s/%s", task.TargetBranch, targetRemote, task.TargetBranch)

	// Log Fetch Target Command
	fetchTgtCmd := fmt.Sprintf("git fetch %s %s", targetRemote, tRefSpec)
	logf("Command: %s", fetchTgtCmd)

	if targetURL != "" && targetAuthInfo.Type != "" && targetAuthInfo.Type != "none" {
		logf("Fetching target %s (Auth: %s, Source: %s)...", targetRemote, targetAuthInfo.Type, targetAuthInfo.Source)
		// 如果使用数据库SSH密钥，使用原生git命令（更可靠）
		if targetAuthInfo.Source == "database" && targetAuthInfo.SSHKeyID > 0 {
			privateKey, passphrase, keyErr := s.authSvc.GetDBSSHKeyContent(targetAuthInfo.SSHKeyID)
			if keyErr != nil {
				return "", fmt.Errorf("failed to load target SSH key: %v", keyErr)
			}
			if err := s.git.FetchWithDBKey(path, targetURL, privateKey, passphrase, progressWriter, tRefSpec); err != nil {
				return "", fmt.Errorf("fetch target failed: %v", err)
			}
		} else if err := s.git.FetchWithAuthMethod(path, targetURL, targetAuth, progressWriter, tRefSpec); err != nil {
			return "", fmt.Errorf("fetch target failed: %v", err)
		}
	} else {
		logf("Fetching target %s...", targetRemote)
		if err := s.git.Fetch(path, targetRemote, progressWriter); err != nil {
			return "", fmt.Errorf("fetch target failed: %v", err)
		}
	}

	// 3. Get Hashes
	// sourceHash already got

	targetHash, err := s.git.GetCommitHash(path, task.TargetRemote, task.TargetBranch)
	// Target branch might not exist yet (first sync).
	targetExists := err == nil

	if targetExists {
		logf("Target hash (%s/%s): %s", task.TargetRemote, task.TargetBranch, targetHash)
	} else {
		logf("Target branch does not exist yet")
	}

	var commitRange string
	if targetExists {
		commitRange = fmt.Sprintf("%s..%s", targetHash, sourceHash)
	} else {
		commitRange = sourceHash // New branch
	}

	if targetExists {
		if sourceHash == targetHash {
			logf("Source and Target are at the same commit. No sync needed.")
			return "", nil // Already synced
		}

		// 4. Check Fast-Forward
		// Is Target an ancestor of Source?
		isAncestor, err := s.git.IsAncestor(path, targetHash, sourceHash)
		if err != nil {
			return "", fmt.Errorf("check ancestor failed: %v", err)
		}

		if !isAncestor {
			logf("Not a fast-forward update. Checking divergence...")
			// Check if diverged or Source is behind
			// If Source is ancestor of Target, Source is behind.
			isSourceBehind, _ := s.git.IsAncestor(path, sourceHash, targetHash)
			if isSourceBehind {
				return "", fmt.Errorf("source is behind target")
			}
			return "", fmt.Errorf("conflict")
		}
		logf("Fast-forward check passed.")
	}

	// 5. Push
	var pushOpts []string
	if task.PushOptions != "" {
		pushOpts = strings.Fields(task.PushOptions)
	}

	// Construct command for logging
	cmdStr := fmt.Sprintf("git push %s %s:refs/heads/%s", task.TargetRemote, sourceHash, task.TargetBranch)
	if len(pushOpts) > 0 {
		cmdStr += " " + strings.Join(pushOpts, " ")
	}
	logf("Command: %s", cmdStr)
	logf("Pushing to %s/%s with options: %v", task.TargetRemote, task.TargetBranch, pushOpts)

	if targetURL != "" && targetAuthInfo.Type != "" && targetAuthInfo.Type != "none" {
		logf("Pushing target (Auth: %s, Source: %s)...", targetAuthInfo.Type, targetAuthInfo.Source)
		// 如果使用数据库SSH密钥，使用原生git命令（更可靠）
		if targetAuthInfo.Source == "database" && targetAuthInfo.SSHKeyID > 0 {
			privateKey, passphrase, keyErr := s.authSvc.GetDBSSHKeyContent(targetAuthInfo.SSHKeyID)
			if keyErr != nil {
				return "", fmt.Errorf("failed to load target SSH key for push: %v", keyErr)
			}
			if err := s.git.PushWithDBKey(path, targetURL, sourceHash, task.TargetBranch, privateKey, passphrase, pushOpts, progressWriter); err != nil {
				return "", fmt.Errorf("push failed: %v", err)
			}
		} else {
			err := s.git.PushWithAuthMethod(path, targetURL, sourceHash, task.TargetBranch, targetAuth, pushOpts, progressWriter)
			if err != nil {
				return "", fmt.Errorf("push failed: %v", err)
			}
		}
	} else {
		if err := s.git.Push(path, task.TargetRemote, sourceHash, task.TargetBranch, pushOpts, progressWriter); err != nil {
			return "", fmt.Errorf("push failed: %v", err)
		}
	}

	return commitRange, nil
}

// doSyncAllBranches 全分支同步：自动检测源 remote 所有分支，逐一同步到目标 remote
func (s *SyncService) doSyncAllBranches(path string, task *po.SyncTask, logf func(string, ...interface{})) (string, error) {
	logf("Starting all-branch sync for task %s (Repo: %s)", task.Key, path)

	sourceRemote := task.SourceRemote
	if sourceRemote == "" {
		sourceRemote = "origin"
	}
	targetRemote := task.TargetRemote
	if targetRemote == "" {
		targetRemote = "origin"
	}

	progressWriter := &logWriter{logf: logf}

	// 1. Fetch all branches from source remote
	isLocalSource := (sourceRemote == "local")
	if !isLocalSource {
		sourceURL, _ := s.git.GetRemoteURL(path, sourceRemote)
		if sourceURL == "" && sourceRemote == "origin" {
			sourceURL = task.SourceRepo.RemoteURL
		}

		sourceAuth, err := s.resolveAuthForRemote(task.SourceRepo, sourceRemote)
		if err != nil {
			logf("Warning: failed to resolve source auth: %v", err)
		}
		sourceAuthInfo := getAuthInfoForRemote(task.SourceRepo, sourceRemote)
		allRefSpec := fmt.Sprintf("+refs/heads/*:refs/remotes/%s/*", sourceRemote)

		logf("Command: git fetch %s %s", sourceRemote, allRefSpec)

		if sourceURL != "" && sourceAuthInfo.Type != "" && sourceAuthInfo.Type != "none" {
			logf("Fetching all branches from source %s (Auth: %s, Source: %s)...", sourceRemote, sourceAuthInfo.Type, sourceAuthInfo.Source)
			if sourceAuthInfo.Source == "database" && sourceAuthInfo.SSHKeyID > 0 {
				privateKey, passphrase, keyErr := s.authSvc.GetDBSSHKeyContent(sourceAuthInfo.SSHKeyID)
				if keyErr != nil {
					return "", fmt.Errorf("failed to load source SSH key: %v", keyErr)
				}
				if err := s.git.FetchWithDBKey(path, sourceURL, privateKey, passphrase, progressWriter, allRefSpec); err != nil {
					return "", fmt.Errorf("fetch source (all branches) failed: %v", err)
				}
			} else if err := s.git.FetchWithAuthMethod(path, sourceURL, sourceAuth, progressWriter, allRefSpec); err != nil {
				return "", fmt.Errorf("fetch source (all branches) failed: %v", err)
			}
		} else {
			logf("Fetching all branches from source %s...", sourceRemote)
			if err := s.git.Fetch(path, sourceRemote, progressWriter); err != nil {
				return "", fmt.Errorf("fetch source (all branches) failed: %v", err)
			}
		}
	}

	// 2. List all branches from source remote
	var branches []string
	if isLocalSource {
		// 本地分支：获取所有本地分支名
		allBranches, err := s.git.GetBranches(path)
		if err != nil {
			return "", fmt.Errorf("list local branches failed: %v", err)
		}
		// GetBranches 返回本地和远程分支，只保留本地分支（不含 /）
		for _, b := range allBranches {
			if !strings.Contains(b, "/") {
				branches = append(branches, b)
			}
		}
	} else {
		var err error
		branches, err = s.git.ListRemoteBranches(path, sourceRemote)
		if err != nil {
			return "", fmt.Errorf("list remote branches failed: %v", err)
		}
	}

	if len(branches) == 0 {
		logf("No branches found on source remote %s", sourceRemote)
		return "", nil
	}
	logf("Found %d branches on source: %v", len(branches), branches)

	// 3. Fetch all branches from target remote
	targetURL, _ := s.git.GetRemoteURL(path, targetRemote)
	if targetURL == "" && targetRemote == "origin" {
		targetURL = task.TargetRepo.RemoteURL
	}
	targetAuth, err := s.resolveAuthForRemote(task.TargetRepo, targetRemote)
	if err != nil {
		logf("Warning: failed to resolve target auth: %v", err)
	}
	targetAuthInfo := getAuthInfoForRemote(task.TargetRepo, targetRemote)

	tAllRefSpec := fmt.Sprintf("+refs/heads/*:refs/remotes/%s/*", targetRemote)
	logf("Command: git fetch %s %s", targetRemote, tAllRefSpec)

	if targetURL != "" && targetAuthInfo.Type != "" && targetAuthInfo.Type != "none" {
		logf("Fetching all branches from target %s (Auth: %s, Source: %s)...", targetRemote, targetAuthInfo.Type, targetAuthInfo.Source)
		if targetAuthInfo.Source == "database" && targetAuthInfo.SSHKeyID > 0 {
			privateKey, passphrase, keyErr := s.authSvc.GetDBSSHKeyContent(targetAuthInfo.SSHKeyID)
			if keyErr != nil {
				return "", fmt.Errorf("failed to load target SSH key: %v", keyErr)
			}
			if err := s.git.FetchWithDBKey(path, targetURL, privateKey, passphrase, progressWriter, tAllRefSpec); err != nil {
				return "", fmt.Errorf("fetch target (all branches) failed: %v", err)
			}
		} else if err := s.git.FetchWithAuthMethod(path, targetURL, targetAuth, progressWriter, tAllRefSpec); err != nil {
			return "", fmt.Errorf("fetch target (all branches) failed: %v", err)
		}
	} else {
		logf("Fetching all branches from target %s...", targetRemote)
		if err := s.git.Fetch(path, targetRemote, progressWriter); err != nil {
			return "", fmt.Errorf("fetch target (all branches) failed: %v", err)
		}
	}

	// 4. Sync each branch
	var pushOpts []string
	if task.PushOptions != "" {
		pushOpts = strings.Fields(task.PushOptions)
	}

	successCount := 0
	failedCount := 0
	skippedCount := 0
	var allCommitRanges []string
	var lastErr error

	for _, branch := range branches {
		logf("--- Syncing branch: %s ---", branch)

		// Get source hash
		var sourceHash string
		if isLocalSource {
			h, err := s.git.ResolveRevision(path, branch)
			if err != nil {
				logf("  Skip branch %s: cannot resolve local ref: %v", branch, err)
				failedCount++
				lastErr = err
				continue
			}
			sourceHash = h
		} else {
			h, err := s.git.GetCommitHash(path, sourceRemote, branch)
			if err != nil {
				logf("  Skip branch %s: cannot get source hash: %v", branch, err)
				failedCount++
				lastErr = err
				continue
			}
			sourceHash = h
		}
		logf("  Source hash: %s", sourceHash)

		// Get target hash
		targetHash, err := s.git.GetCommitHash(path, targetRemote, branch)
		targetExists := err == nil

		if targetExists {
			logf("  Target hash: %s", targetHash)
			if sourceHash == targetHash {
				logf("  Already in sync, skipping")
				skippedCount++
				continue
			}

			// Fast-forward check
			isAncestor, err := s.git.IsAncestor(path, targetHash, sourceHash)
			if err != nil {
				logf("  Branch %s: ancestor check failed: %v", branch, err)
				failedCount++
				lastErr = err
				continue
			}
			if !isAncestor {
				isSourceBehind, _ := s.git.IsAncestor(path, sourceHash, targetHash)
				if isSourceBehind {
					logf("  Branch %s: source is behind target, skipping", branch)
					failedCount++
					lastErr = fmt.Errorf("branch %s: source is behind target", branch)
					continue
				}
				logf("  Branch %s: conflict (not fast-forward)", branch)
				failedCount++
				lastErr = fmt.Errorf("branch %s: conflict", branch)
				continue
			}
			logf("  Fast-forward check passed")
			allCommitRanges = append(allCommitRanges, fmt.Sprintf("%s: %s..%s", branch, targetHash[:8], sourceHash[:8]))
		} else {
			logf("  Target branch does not exist yet (new branch)")
			allCommitRanges = append(allCommitRanges, fmt.Sprintf("%s: (new) %s", branch, sourceHash[:8]))
		}

		// Push
		logf("  Pushing %s to %s/%s...", sourceHash[:8], targetRemote, branch)
		if targetURL != "" && targetAuthInfo.Type != "" && targetAuthInfo.Type != "none" {
			if targetAuthInfo.Source == "database" && targetAuthInfo.SSHKeyID > 0 {
				privateKey, passphrase, keyErr := s.authSvc.GetDBSSHKeyContent(targetAuthInfo.SSHKeyID)
				if keyErr != nil {
					logf("  Branch %s: failed to load SSH key: %v", branch, keyErr)
					failedCount++
					lastErr = keyErr
					continue
				}
				if err := s.git.PushWithDBKey(path, targetURL, sourceHash, branch, privateKey, passphrase, pushOpts, progressWriter); err != nil {
					logf("  Branch %s: push failed: %v", branch, err)
					failedCount++
					lastErr = err
					continue
				}
			} else {
				if err := s.git.PushWithAuthMethod(path, targetURL, sourceHash, branch, targetAuth, pushOpts, progressWriter); err != nil {
					logf("  Branch %s: push failed: %v", branch, err)
					failedCount++
					lastErr = err
					continue
				}
			}
		} else {
			if err := s.git.Push(path, targetRemote, sourceHash, branch, pushOpts, progressWriter); err != nil {
				logf("  Branch %s: push failed: %v", branch, err)
				failedCount++
				lastErr = err
				continue
			}
		}

		logf("  Branch %s synced successfully", branch)
		successCount++
	}

	// 5. Summary
	logf("=== All-branch sync summary ===")
	logf("Total: %d, Success: %d, Failed: %d, Skipped (up-to-date): %d", len(branches), successCount, failedCount, skippedCount)

	commitRange := strings.Join(allCommitRanges, "; ")

	if failedCount > 0 && successCount == 0 {
		return commitRange, fmt.Errorf("all branches failed, last error: %v", lastErr)
	}
	if failedCount > 0 {
		return commitRange, fmt.Errorf("%d/%d branches failed, last error: %v", failedCount, len(branches), lastErr)
	}
	return commitRange, nil
}

// LogWriter implements io.Writer
type logWriter struct {
	logf func(string, ...interface{})
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	str := strings.TrimSpace(string(p))
	if str != "" {
		w.logf("[Git] %s", str)
	}
	return len(p), nil
}

// sendNotification 根据同步结果发送通知
func (s *SyncService) sendNotification(task *po.SyncTask, run *po.SyncRun) {
	var triggerEvent string
	var status string

	switch run.Status {
	case "success":
		triggerEvent = po.TriggerSyncSuccess
		status = "success"
	case "failed":
		triggerEvent = po.TriggerSyncFailure
		status = "failure"
	case "conflict":
		triggerEvent = po.TriggerSyncConflict
		status = "failure"
	default:
		return
	}

	// 计算耗时
	duration := ""
	if !run.EndTime.IsZero() && !run.StartTime.IsZero() {
		d := run.EndTime.Sub(run.StartTime)
		duration = d.Round(time.Millisecond).String()
	}

	data := &notificationSvc.TemplateData{
		TaskKey:      task.Key,
		Status:       run.Status,
		EventType:    triggerEvent,
		SourceRemote: task.SourceRemote,
		SourceBranch: task.SourceBranch,
		TargetRemote: task.TargetRemote,
		TargetBranch: task.TargetBranch,
		RepoKey:      task.SourceRepoKey,
		ErrorMessage: run.ErrorMessage,
		CommitRange:  run.CommitRange,
		Duration:     duration,
		SyncMode:     task.SyncMode,
	}
	if task.Cron != "" {
		data.CronExpression = task.Cron
	}

	// 使用模板渲染的默认标题和内容作为 fallback
	title, content := notificationSvc.RenderTitleAndContent("", "", data)

	notificationSvc.NotifySvc.Send(&notificationSvc.NotificationMessage{
		Title:        title,
		Content:      content,
		Status:       status,
		TriggerEvent: triggerEvent,
		TaskKey:      task.Key,
		RepoKey:      task.SourceRepoKey,
		Data:         data,
	})
}
