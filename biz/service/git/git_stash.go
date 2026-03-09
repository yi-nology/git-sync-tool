package git

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yi-nology/git-manage-service/pkg/logger"
)

// StashEntry Stash 条目
type StashEntry struct {
	Index   int    `json:"index"`
	Ref     string `json:"ref"`
	Message string `json:"message"`
	Branch  string `json:"branch"`
	Date    string `json:"date"`
}

// StashList 列出所有 stash
func (s *GitService) StashList(path string) ([]StashEntry, error) {
	logger.Debug("Listing stash entries", logrus.Fields{"path": path})

	output, err := s.RunCommand(path, "stash", "list", "--format=%gd|%gs|%ci")
	if err != nil {
		logger.ErrorWithErr("Failed to list stash", err, logrus.Fields{"path": path})
		return nil, err
	}

	if output == "" {
		return []StashEntry{}, nil
	}

	var entries []StashEntry
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		entry := StashEntry{
			Index: i,
			Ref:   fmt.Sprintf("stash@{%d}", i),
		}
		if len(parts) >= 1 {
			entry.Ref = parts[0]
		}
		if len(parts) >= 2 {
			entry.Message = parts[1]
		}
		if len(parts) >= 3 {
			entry.Date = parts[2]
		}
		entries = append(entries, entry)
	}

	logger.Debug("Stash entries retrieved", logrus.Fields{"path": path, "count": len(entries)})
	return entries, nil
}

// StashSave 保存当前更改到 stash
func (s *GitService) StashSave(path, message string, includeUntracked bool) error {
	logger.Info("Saving stash", logrus.Fields{
		"path":             path,
		"message":          message,
		"includeUntracked": includeUntracked,
	})

	// 检查是否有可暂存的更改
	status, err := s.RunCommand(path, "status", "--porcelain")
	if err != nil {
		logger.ErrorWithErr("Failed to check status", err, logrus.Fields{"path": path})
		return err
	}

	// 如果启用了 includeUntracked，检查所有文件；否则只检查已跟踪的修改
	hasChanges := false
	if includeUntracked {
		hasChanges = status != ""
	} else {
		// 只检查已修改的文件（M, A, D, R, C 状态，排除 ??）
		lines := strings.Split(status, "\n")
		for _, line := range lines {
			if len(line) > 2 && line[:2] != "??" {
				hasChanges = true
				break
			}
		}
	}

	if !hasChanges {
		logger.Warn("No changes to stash", logrus.Fields{"path": path})
		return fmt.Errorf("no changes to stash")
	}

	// 记录执行前的 stash 数量
	beforeList, _ := s.RunCommand(path, "stash", "list")
	beforeCount := len(strings.Split(beforeList, "\n"))
	if beforeList == "" {
		beforeCount = 0
	}

	args := []string{"stash", "push"}
	if message != "" {
		args = append(args, "-m", message)
	}
	if includeUntracked {
		args = append(args, "-u")
	}

	_, err = s.RunCommand(path, args...)
	if err != nil {
		logger.ErrorWithErr("Failed to save stash", err, logrus.Fields{"path": path})
		return err
	}

	// 验证 stash 是否真的被创建
	afterList, _ := s.RunCommand(path, "stash", "list")
	afterCount := len(strings.Split(afterList, "\n"))
	if afterList == "" {
		afterCount = 0
	}

	if afterCount <= beforeCount {
		logger.Error("Stash was not created", logrus.Fields{"path": path, "before": beforeCount, "after": afterCount})
		return fmt.Errorf("failed to create stash: no stash entry was added")
	}

	logger.Info("Stash saved successfully", logrus.Fields{"path": path, "new_count": afterCount})
	return nil
}

// StashApply 应用 stash（不删除）
func (s *GitService) StashApply(path string, index int) error {
	ref := fmt.Sprintf("stash@{%d}", index)
	logger.Info("Applying stash", logrus.Fields{"path": path, "ref": ref})

	_, err := s.RunCommand(path, "stash", "apply", ref)
	if err != nil {
		logger.ErrorWithErr("Failed to apply stash", err, logrus.Fields{"ref": ref})
		return err
	}

	logger.Info("Stash applied successfully", logrus.Fields{"ref": ref})
	return nil
}

// StashPop 弹出 stash（应用并删除）
func (s *GitService) StashPop(path string, index int) error {
	ref := fmt.Sprintf("stash@{%d}", index)
	logger.Info("Popping stash", logrus.Fields{"path": path, "ref": ref})

	_, err := s.RunCommand(path, "stash", "pop", ref)
	if err != nil {
		logger.ErrorWithErr("Failed to pop stash", err, logrus.Fields{"ref": ref})
		return err
	}

	logger.Info("Stash popped successfully", logrus.Fields{"ref": ref})
	return nil
}

// StashDrop 删除指定 stash
func (s *GitService) StashDrop(path string, index int) error {
	ref := fmt.Sprintf("stash@{%d}", index)
	logger.Info("Dropping stash", logrus.Fields{"path": path, "ref": ref})

	_, err := s.RunCommand(path, "stash", "drop", ref)
	if err != nil {
		logger.ErrorWithErr("Failed to drop stash", err, logrus.Fields{"ref": ref})
		return err
	}

	logger.Info("Stash dropped successfully", logrus.Fields{"ref": ref})
	return nil
}

// StashClear 清空所有 stash
func (s *GitService) StashClear(path string) error {
	logger.Info("Clearing all stash entries", logrus.Fields{"path": path})

	_, err := s.RunCommand(path, "stash", "clear")
	if err != nil {
		logger.ErrorWithErr("Failed to clear stash", err, logrus.Fields{"path": path})
		return err
	}

	logger.Info("All stash entries cleared", logrus.Fields{"path": path})
	return nil
}
