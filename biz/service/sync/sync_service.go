package sync

import (
	"context"
	"fmt"
	"io"
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

// getAuthInfoForRemote 获取指定远程的认证信息（旧系统）
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

// resolveAuthForRemote 解析指定远程的认证方法（支持新凭证系统 + 旧系统回退）
func (s *SyncService) resolveAuthForRemote(repo po.Repo, remoteName string) (transport.AuthMethod, bool, error) {
	return s.authSvc.ResolveCredentialForRemote(
		repo.RemoteCredentials,
		repo.DefaultCredentialID,
		repo.RemoteAuths,
		remoteName,
		repo.AuthType, repo.AuthKey, repo.AuthSecret,
	)
}

// fetchRemote 统一的 fetch 操作，自动处理认证方式选择
func (s *SyncService) fetchRemote(path string, repo po.Repo, remoteName, remoteURL string, refSpecs string, progressWriter io.Writer, logf func(string, ...interface{})) error {
	authMethod, isDBKey, err := s.resolveAuthForRemote(repo, remoteName)
	if err != nil {
		logf("Warning: failed to resolve auth for %s: %v", remoteName, err)
	}

	hasAuth := authMethod != nil || isDBKey

	if remoteURL != "" && hasAuth {
		if isDBKey {
			// 获取凭证 ID 或旧系统的 SSHKeyID
			credID := auth.GetCredentialIDForRemote(repo.RemoteCredentials, repo.DefaultCredentialID, remoteName)
			if credID > 0 {
				privateKey, passphrase, keyErr := s.authSvc.GetCredentialKeyContent(credID)
				if keyErr != nil {
					// 回退到旧系统
					authInfo := getAuthInfoForRemote(repo, remoteName)
					if authInfo.SSHKeyID > 0 {
						privateKey, passphrase, keyErr = s.authSvc.GetDBSSHKeyContent(authInfo.SSHKeyID)
					}
				}
				if keyErr == nil && privateKey != "" {
					logf("Fetching %s using DB SSH key...", remoteName)
					return s.git.FetchWithDBKey(path, remoteURL, privateKey, passphrase, progressWriter, refSpecs)
				}
			}
			// 回退到旧系统
			authInfo := getAuthInfoForRemote(repo, remoteName)
			if authInfo.SSHKeyID > 0 {
				privateKey, passphrase, keyErr := s.authSvc.GetDBSSHKeyContent(authInfo.SSHKeyID)
				if keyErr != nil {
					return fmt.Errorf("failed to load SSH key: %v", keyErr)
				}
				logf("Fetching %s using DB SSH key (legacy)...", remoteName)
				return s.git.FetchWithDBKey(path, remoteURL, privateKey, passphrase, progressWriter, refSpecs)
			}
		}
		logf("Fetching %s with auth...", remoteName)
		return s.git.FetchWithAuthMethod(path, remoteURL, authMethod, progressWriter, refSpecs)
	}

	logf("Fetching %s (no auth)...", remoteName)
	return s.git.Fetch(path, remoteName, progressWriter)
}

// pushRemote 统一的 push 操作，自动处理认证方式选择
func (s *SyncService) pushRemote(path string, repo po.Repo, remoteName, remoteURL, sourceHash, targetBranch string, pushOpts []string, progressWriter io.Writer, logf func(string, ...interface{})) error {
	authMethod, isDBKey, err := s.resolveAuthForRemote(repo, remoteName)
	if err != nil {
		logf("Warning: failed to resolve auth for push to %s: %v", remoteName, err)
	}

	hasAuth := authMethod != nil || isDBKey

	if remoteURL != "" && hasAuth {
		if isDBKey {
			credID := auth.GetCredentialIDForRemote(repo.RemoteCredentials, repo.DefaultCredentialID, remoteName)
			if credID > 0 {
				privateKey, passphrase, keyErr := s.authSvc.GetCredentialKeyContent(credID)
				if keyErr != nil {
					authInfo := getAuthInfoForRemote(repo, remoteName)
					if authInfo.SSHKeyID > 0 {
						privateKey, passphrase, keyErr = s.authSvc.GetDBSSHKeyContent(authInfo.SSHKeyID)
					}
				}
				if keyErr == nil && privateKey != "" {
					logf("Pushing to %s using DB SSH key...", remoteName)
					return s.git.PushWithDBKey(path, remoteURL, sourceHash, targetBranch, privateKey, passphrase, pushOpts, progressWriter)
				}
			}
			authInfo := getAuthInfoForRemote(repo, remoteName)
			if authInfo.SSHKeyID > 0 {
				privateKey, passphrase, keyErr := s.authSvc.GetDBSSHKeyContent(authInfo.SSHKeyID)
				if keyErr != nil {
					return fmt.Errorf("failed to load SSH key for push: %v", keyErr)
				}
				logf("Pushing to %s using DB SSH key (legacy)...", remoteName)
				return s.git.PushWithDBKey(path, remoteURL, sourceHash, targetBranch, privateKey, passphrase, pushOpts, progressWriter)
			}
		}
		logf("Pushing to %s with auth...", remoteName)
		return s.git.PushWithAuthMethod(path, remoteURL, sourceHash, targetBranch, authMethod, pushOpts, progressWriter)
	}

	logf("Pushing to %s (no auth)...", remoteName)
	return s.git.Push(path, remoteName, sourceHash, targetBranch, pushOpts, progressWriter)
}

func (s *SyncService) doSyncSingleBranch(path string, task *po.SyncTask, logf func(string, ...interface{})) (string, error) {
	logf("Starting sync for task %s (Repo: %s)", task.Key, path)

	// 1. Fetch Source
	sourceRemote := task.SourceRemote
	if sourceRemote == "" {
		sourceRemote = "origin"
	}

	isLocalSource := (sourceRemote == "local")
	var sourceHash string
	progressWriter := &logWriter{logf: logf}

	if !isLocalSource {
		sourceURL, _ := s.git.GetRemoteURL(path, sourceRemote)
		if sourceURL == "" && sourceRemote == "origin" {
			sourceURL = task.SourceRepo.RemoteURL
		}

		sRefSpec := fmt.Sprintf("+refs/heads/%s:refs/remotes/%s/%s", task.SourceBranch, sourceRemote, task.SourceBranch)
		logf("Command: git fetch %s %s", sourceRemote, sRefSpec)

		if err := s.fetchRemote(path, task.SourceRepo, sourceRemote, sourceURL, sRefSpec, progressWriter, logf); err != nil {
			return "", fmt.Errorf("fetch source failed: %v", err)
		}

		h, err := s.git.GetCommitHash(path, task.SourceRemote, task.SourceBranch)
		if err != nil {
			return "", fmt.Errorf("get source hash failed: %v", err)
		}
		sourceHash = h
	} else {
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

	tRefSpec := fmt.Sprintf("+refs/heads/%s:refs/remotes/%s/%s", task.TargetBranch, targetRemote, task.TargetBranch)
	logf("Command: git fetch %s %s", targetRemote, tRefSpec)

	if err := s.fetchRemote(path, task.TargetRepo, targetRemote, targetURL, tRefSpec, progressWriter, logf); err != nil {
		return "", fmt.Errorf("fetch target failed: %v", err)
	}

	// 3. Get Hashes
	targetHash, err := s.git.GetCommitHash(path, task.TargetRemote, task.TargetBranch)
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
		commitRange = sourceHash
	}

	if targetExists {
		if sourceHash == targetHash {
			logf("Source and Target are at the same commit. No sync needed.")
			return "", nil
		}

		// 4. Check Fast-Forward
		isAncestor, err := s.git.IsAncestor(path, targetHash, sourceHash)
		if err != nil {
			return "", fmt.Errorf("check ancestor failed: %v", err)
		}

		if !isAncestor {
			logf("Not a fast-forward update. Checking divergence...")
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

	cmdStr := fmt.Sprintf("git push %s %s:refs/heads/%s", task.TargetRemote, sourceHash, task.TargetBranch)
	if len(pushOpts) > 0 {
		cmdStr += " " + strings.Join(pushOpts, " ")
	}
	logf("Command: %s", cmdStr)
	logf("Pushing to %s/%s with options: %v", task.TargetRemote, task.TargetBranch, pushOpts)

	if err := s.pushRemote(path, task.TargetRepo, targetRemote, targetURL, sourceHash, task.TargetBranch, pushOpts, progressWriter, logf); err != nil {
		return "", fmt.Errorf("push failed: %v", err)
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

		allRefSpec := fmt.Sprintf("+refs/heads/*:refs/remotes/%s/*", sourceRemote)
		logf("Command: git fetch %s %s", sourceRemote, allRefSpec)

		if err := s.fetchRemote(path, task.SourceRepo, sourceRemote, sourceURL, allRefSpec, progressWriter, logf); err != nil {
			return "", fmt.Errorf("fetch source (all branches) failed: %v", err)
		}
	}

	// 2. List all branches from source remote
	var branches []string
	if isLocalSource {
		allBranches, err := s.git.GetBranches(path)
		if err != nil {
			return "", fmt.Errorf("list local branches failed: %v", err)
		}
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

	tAllRefSpec := fmt.Sprintf("+refs/heads/*:refs/remotes/%s/*", targetRemote)
	logf("Command: git fetch %s %s", targetRemote, tAllRefSpec)

	if err := s.fetchRemote(path, task.TargetRepo, targetRemote, targetURL, tAllRefSpec, progressWriter, logf); err != nil {
		return "", fmt.Errorf("fetch target (all branches) failed: %v", err)
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
		if err := s.pushRemote(path, task.TargetRepo, targetRemote, targetURL, sourceHash, branch, pushOpts, progressWriter, logf); err != nil {
			logf("  Branch %s: push failed: %v", branch, err)
			failedCount++
			lastErr = err
			continue
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
		TaskName:     task.Key,
		Status:       run.Status,
		EventType:    triggerEvent,
		SourceRemote: task.SourceRemote,
		SourceBranch: task.SourceBranch,
		TargetRemote: task.TargetRemote,
		TargetBranch: task.TargetBranch,
		RepoKey:      task.SourceRepoKey,
		RepoName:     task.SourceRepo.Name,
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
