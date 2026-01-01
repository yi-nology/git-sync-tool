package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal"
	"github.com/yi-nology/git-manage-service/biz/model"
)

type SyncService struct {
	git *GitService
}

func NewSyncService() *SyncService {
	return &SyncService{
		git: NewGitService(),
	}
}

func (s *SyncService) RunTask(taskID uint) error {
	var task model.SyncTask
	if err := dal.DB.Preload("SourceRepo").Preload("TargetRepo").First(&task, taskID).Error; err != nil {
		return err
	}
	return s.ExecuteSync(&task)
}

func (s *SyncService) ExecuteSync(task *model.SyncTask) error {
	run := model.SyncRun{
		TaskKey:   task.Key,
		StartTime: time.Now(),
		Status:    "running",
	}
	dal.DB.Create(&run)

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
	dal.DB.Save(&run)
	return err
}

func getAuthForRemote(repo model.Repo, remoteName string) (string, string, string) {
	if repo.RemoteAuths != nil {
		if auth, ok := repo.RemoteAuths[remoteName]; ok {
			return auth.Type, auth.Key, auth.Secret
		}
	}
	return repo.AuthType, repo.AuthKey, repo.AuthSecret
}

func (s *SyncService) doSync(path string, task *model.SyncTask, logf func(string, ...interface{})) (string, error) {
	logf("Starting sync for task %s (Repo: %s)", task.Key, path)

	// 1. Fetch Source
	sourceRemote := task.SourceRemote
	if sourceRemote == "" {
		sourceRemote = "origin"
	}

	// Check if source is local
	isLocalSource := (sourceRemote == "local")

	var sourceHash string

	if !isLocalSource {
		sourceURL, _ := s.git.GetRemoteURL(path, sourceRemote)
		if sourceURL == "" && sourceRemote == "origin" {
			sourceURL = task.SourceRepo.RemoteURL
		}

		sType, sKey, sSecret := getAuthForRemote(task.SourceRepo, sourceRemote)
		sRefSpec := fmt.Sprintf("+refs/heads/%s:refs/remotes/%s/%s", task.SourceBranch, sourceRemote, task.SourceBranch)

		if sourceURL != "" && sType != "" && sType != "none" {
			logf("Fetching source %s (Auth: %s)...", sourceRemote, sType)
			if err := s.git.FetchWithAuth(path, sourceURL, sType, sKey, sSecret, sRefSpec); err != nil {
				return "", fmt.Errorf("fetch source failed: %v", err)
			}
		} else {
			logf("Fetching source %s...", sourceRemote)
			if err := s.git.Fetch(path, sourceRemote); err != nil {
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
		h, err := s.git.RunCommand(path, "rev-parse", task.SourceBranch)
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

	tType, tKey, tSecret := getAuthForRemote(task.TargetRepo, targetRemote)
	tRefSpec := fmt.Sprintf("+refs/heads/%s:refs/remotes/%s/%s", task.TargetBranch, targetRemote, task.TargetBranch)

	if targetURL != "" && tType != "" && tType != "none" {
		logf("Fetching target %s (Auth: %s)...", targetRemote, tType)
		if err := s.git.FetchWithAuth(path, targetURL, tType, tKey, tSecret, tRefSpec); err != nil {
			return "", fmt.Errorf("fetch target failed: %v", err)
		}
	} else {
		logf("Fetching target %s...", targetRemote)
		if err := s.git.Fetch(path, targetRemote); err != nil {
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
	logf("Pushing to %s/%s with options: %v", task.TargetRemote, task.TargetBranch, pushOpts)

	if targetURL != "" && tType != "" && tType != "none" {
		logf("Pushing target (Auth: %s)...", tType)
		err := s.git.PushWithAuth(path, targetURL, sourceHash, task.TargetBranch, tType, tKey, tSecret, pushOpts)
		if err != nil {
			return "", fmt.Errorf("push failed: %v", err)
		}
	} else {
		if err := s.git.Push(path, task.TargetRemote, sourceHash, task.TargetBranch, pushOpts); err != nil {
			return "", fmt.Errorf("push failed: %v", err)
		}
	}

	return commitRange, nil
}
