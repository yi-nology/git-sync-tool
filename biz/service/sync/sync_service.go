package sync

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/auth"
	"github.com/yi-nology/git-manage-service/biz/service/git"
)

type SyncService struct {
	git         *git.GitService
	authSvc     *auth.AuthService
	syncTaskDAO *db.SyncTaskDAO
	syncRunDAO  *db.SyncRunDAO
}

func NewSyncService() *SyncService {
	return &SyncService{
		git:         git.NewGitService(),
		authSvc:     auth.NewAuthService(),
		syncTaskDAO: db.NewSyncTaskDAO(),
		syncRunDAO:  db.NewSyncRunDAO(),
	}
}

func (s *SyncService) RunTask(taskKey string) error {
	task, err := s.syncTaskDAO.FindByKey(taskKey)
	if err != nil {
		return err
	}
	return s.ExecuteSync(task)
}

func (s *SyncService) ExecuteSync(task *po.SyncTask) error {
	run := po.SyncRun{
		TaskKey:   task.Key,
		StartTime: time.Now(),
		Status:    "running",
	}
	s.syncRunDAO.Create(&run)

	repoPath := task.SourceRepo.Path

	// Capture logs
	var logs strings.Builder
	logf := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		logs.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05"), msg))
	}

	commitRange, err := s.doSync(repoPath, task, logf)

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

func (s *SyncService) doSync(path string, task *po.SyncTask, logf func(string, ...interface{})) (string, error) {
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
