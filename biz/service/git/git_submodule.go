package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yi-nology/git-manage-service/pkg/logger"
)

// SubmoduleInfo Submodule 信息
type SubmoduleInfo struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	URL    string `json:"url"`
	Branch string `json:"branch"`
	Commit string `json:"commit"`
	Status string `json:"status"` // initialized, uninitialized, modified
}

// SubmoduleStatusItem Submodule 状态项
type SubmoduleStatusItem struct {
	Path        string `json:"path"`
	Commit      string `json:"commit"`
	Status      string `json:"status"`      // +, -, U, 空
	Description string `json:"description"` // 状态描述
}

// SubmoduleList 列出所有 submodule
func (s *GitService) SubmoduleList(path string) ([]SubmoduleInfo, error) {
	logger.Debug("Listing submodules", logrus.Fields{"path": path})

	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	submodules, err := w.Submodules()
	if err != nil {
		return nil, err
	}

	var result []SubmoduleInfo
	for _, sm := range submodules {
		cfg := sm.Config()
		status, err := sm.Status()

		info := SubmoduleInfo{
			Name:   cfg.Name,
			Path:   cfg.Path,
			URL:    cfg.URL,
			Branch: cfg.Branch,
		}

		if err == nil && status != nil {
			info.Commit = status.Current.String()
			if !status.IsClean() {
				info.Status = "modified"
			} else if status.Current.IsZero() {
				info.Status = "uninitialized"
			} else {
				info.Status = "initialized"
			}
		} else {
			info.Status = "unknown"
		}

		result = append(result, info)
	}

	logger.Debug("Submodules listed", logrus.Fields{"path": path, "count": len(result)})
	return result, nil
}

// SubmoduleAdd 添加 submodule
func (s *GitService) SubmoduleAdd(path, url, subPath, branch string) error {
	logger.Info("Adding submodule", logrus.Fields{
		"path":    path,
		"url":     url,
		"subPath": subPath,
		"branch":  branch,
	})

	args := []string{"submodule", "add"}
	if branch != "" {
		args = append(args, "-b", branch)
	}
	args = append(args, url, subPath)

	_, err := s.RunCommand(path, args...)
	if err != nil {
		logger.ErrorWithErr("Failed to add submodule", err, logrus.Fields{"subPath": subPath})
		return err
	}

	logger.Info("Submodule added successfully", logrus.Fields{"subPath": subPath})
	return nil
}

// SubmoduleInit 初始化 submodule
func (s *GitService) SubmoduleInit(path, subPath string) error {
	logger.Info("Initializing submodule", logrus.Fields{"path": path, "subPath": subPath})

	args := []string{"submodule", "init"}
	if subPath != "" {
		args = append(args, subPath)
	}

	_, err := s.RunCommand(path, args...)
	if err != nil {
		logger.ErrorWithErr("Failed to init submodule", err, logrus.Fields{"subPath": subPath})
		return err
	}

	logger.Info("Submodule initialized", logrus.Fields{"subPath": subPath})
	return nil
}

// SubmoduleUpdate 更新 submodule
func (s *GitService) SubmoduleUpdate(path, subPath string, init, recursive, remote bool) error {
	logger.Info("Updating submodule", logrus.Fields{
		"path":      path,
		"subPath":   subPath,
		"init":      init,
		"recursive": recursive,
		"remote":    remote,
	})

	args := []string{"submodule", "update"}
	if init {
		args = append(args, "--init")
	}
	if recursive {
		args = append(args, "--recursive")
	}
	if remote {
		args = append(args, "--remote")
	}
	if subPath != "" {
		args = append(args, "--", subPath)
	}

	_, err := s.RunCommand(path, args...)
	if err != nil {
		logger.ErrorWithErr("Failed to update submodule", err, logrus.Fields{"subPath": subPath})
		return err
	}

	logger.Info("Submodule updated", logrus.Fields{"subPath": subPath})
	return nil
}

// SubmoduleSync 同步 submodule URL
func (s *GitService) SubmoduleSync(path, subPath string, recursive bool) error {
	logger.Info("Syncing submodule", logrus.Fields{
		"path":      path,
		"subPath":   subPath,
		"recursive": recursive,
	})

	args := []string{"submodule", "sync"}
	if recursive {
		args = append(args, "--recursive")
	}
	if subPath != "" {
		args = append(args, "--", subPath)
	}

	_, err := s.RunCommand(path, args...)
	if err != nil {
		logger.ErrorWithErr("Failed to sync submodule", err, logrus.Fields{"subPath": subPath})
		return err
	}

	logger.Info("Submodule synced", logrus.Fields{"subPath": subPath})
	return nil
}

// SubmoduleRemove 移除 submodule
func (s *GitService) SubmoduleRemove(path, subPath string, force bool) error {
	logger.Info("Removing submodule", logrus.Fields{
		"path":    path,
		"subPath": subPath,
		"force":   force,
	})

	// git submodule deinit
	args := []string{"submodule", "deinit"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, subPath)

	if _, err := s.RunCommand(path, args...); err != nil {
		logger.ErrorWithErr("Failed to deinit submodule", err, logrus.Fields{"subPath": subPath})
		return fmt.Errorf("deinit failed: %w", err)
	}

	// git rm
	rmArgs := []string{"rm"}
	if force {
		rmArgs = append(rmArgs, "-f")
	}
	rmArgs = append(rmArgs, subPath)

	if _, err := s.RunCommand(path, rmArgs...); err != nil {
		logger.ErrorWithErr("Failed to remove submodule from git", err, logrus.Fields{"subPath": subPath})
		return fmt.Errorf("rm failed: %w", err)
	}

	// 删除 .git/modules 中的缓存目录
	modulesDir := filepath.Join(path, ".git", "modules", subPath)
	if _, err := os.Stat(modulesDir); err == nil {
		if err := os.RemoveAll(modulesDir); err != nil {
			logger.ErrorWithErr("Failed to remove modules cache", err, logrus.Fields{"dir": modulesDir})
			return fmt.Errorf("remove modules cache failed: %w", err)
		}
	}

	logger.Info("Submodule removed successfully", logrus.Fields{"subPath": subPath})
	return nil
}

// SubmoduleStatus 获取 submodule 状态
func (s *GitService) SubmoduleStatus(path string, recursive bool) ([]SubmoduleStatusItem, error) {
	logger.Debug("Getting submodule status", logrus.Fields{"path": path, "recursive": recursive})

	args := []string{"submodule", "status"}
	if recursive {
		args = append(args, "--recursive")
	}

	output, err := s.RunCommand(path, args...)
	if err != nil {
		return nil, err
	}

	if output == "" {
		return []SubmoduleStatusItem{}, nil
	}

	var items []SubmoduleStatusItem
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		item := SubmoduleStatusItem{}

		// 解析状态标记
		if len(line) > 0 {
			switch line[0] {
			case '+':
				item.Status = "+"
				item.Description = "has new commits"
				line = line[1:]
			case '-':
				item.Status = "-"
				item.Description = "not initialized"
				line = line[1:]
			case 'U':
				item.Status = "U"
				item.Description = "has conflicts"
				line = line[1:]
			case ' ':
				item.Status = ""
				item.Description = "up to date"
				line = line[1:]
			}
		}

		// 解析 commit 和 path
		parts := strings.Fields(strings.TrimSpace(line))
		if len(parts) >= 2 {
			item.Commit = parts[0]
			item.Path = parts[1]
		}

		items = append(items, item)
	}

	logger.Debug("Submodule status retrieved", logrus.Fields{"path": path, "count": len(items)})
	return items, nil
}
