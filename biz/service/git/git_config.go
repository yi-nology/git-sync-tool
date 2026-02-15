package git

import (
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5/config"
	"github.com/sirupsen/logrus"
	"github.com/yi-nology/git-manage-service/pkg/logger"
)

// GetGitUser 获取仓库的 git 用户配置
func (s *GitService) GetGitUser(path string) (string, string, error) {
	logger.Debug("Getting git user", logrus.Fields{"path": path})

	var name, email string

	// 1. 尝试本地配置
	r, err := s.openRepo(path)
	if err == nil {
		if cfg, err := r.Config(); err == nil {
			name = cfg.User.Name
			email = cfg.User.Email
		}
	}

	if name != "" && email != "" {
		return name, email, nil
	}

	// 2. 尝试全局配置 (~/.gitconfig)
	home, err := os.UserHomeDir()
	if err == nil {
		globalConfigPath := filepath.Join(home, ".gitconfig")
		content, err := os.ReadFile(globalConfigPath)
		if err == nil {
			cfg := config.NewConfig()
			if err := cfg.Unmarshal(content); err == nil {
				if name == "" {
					name = cfg.User.Name
				}
				if email == "" {
					email = cfg.User.Email
				}
			}
		}
	}

	logger.Debug("Git user retrieved", logrus.Fields{
		"path":  path,
		"name":  name,
		"email": email,
	})
	return name, email, nil
}

// GetGlobalGitUser 获取全局 git 用户配置
func (s *GitService) GetGlobalGitUser() (string, string, error) {
	logger.Debug("Getting global git user")

	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}

	globalConfigPath := filepath.Join(home, ".gitconfig")
	content, err := os.ReadFile(globalConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", nil
		}
		return "", "", err
	}

	cfg := config.NewConfig()
	if err := cfg.Unmarshal(content); err != nil {
		return "", "", err
	}

	logger.Debug("Global git user retrieved", logrus.Fields{
		"name":  cfg.User.Name,
		"email": cfg.User.Email,
	})
	return cfg.User.Name, cfg.User.Email, nil
}

// SetGlobalGitUser 设置全局 git 用户配置
func (s *GitService) SetGlobalGitUser(name, email string) error {
	logger.Info("Setting global git user", logrus.Fields{
		"name":  name,
		"email": email,
	})

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	globalConfigPath := filepath.Join(home, ".gitconfig")

	// 读取现有配置
	cfg := config.NewConfig()
	content, err := os.ReadFile(globalConfigPath)
	if err == nil {
		if err := cfg.Unmarshal(content); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	// 更新用户信息
	cfg.User.Name = name
	cfg.User.Email = email

	// 写回配置
	data, err := cfg.Marshal()
	if err != nil {
		return err
	}

	err = os.WriteFile(globalConfigPath, data, 0644)
	if err != nil {
		logger.ErrorWithErr("Failed to write git config", err, logrus.Fields{"path": globalConfigPath})
		return err
	}

	logger.Info("Global git user set successfully", logrus.Fields{
		"name":  name,
		"email": email,
	})
	return nil
}
