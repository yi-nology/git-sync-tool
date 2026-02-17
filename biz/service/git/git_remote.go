package git

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sirupsen/logrus"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/pkg/logger"
)

// GetRemotes 获取所有远程仓库名称
func (s *GitService) GetRemotes(path string) ([]string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	remotes, err := r.Remotes()
	if err != nil {
		return nil, err
	}

	var names []string
	for _, remote := range remotes {
		names = append(names, remote.Config().Name)
	}

	logger.Debug("Remotes retrieved", logrus.Fields{"path": path, "count": len(names)})
	return names, nil
}

// GetRemoteURL 获取远程仓库 URL
func (s *GitService) GetRemoteURL(path, remoteName string) (string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return "", err
	}

	remote, err := r.Remote(remoteName)
	if err != nil {
		return "", err
	}

	urls := remote.Config().URLs
	if len(urls) > 0 {
		return urls[0], nil
	}
	return "", fmt.Errorf("no URL for remote %s", remoteName)
}

// AddRemote 添加远程仓库
func (s *GitService) AddRemote(path, name, url string, isMirror bool) error {
	logger.Info("Adding remote", logrus.Fields{
		"path":     path,
		"name":     name,
		"url":      url,
		"isMirror": isMirror,
	})

	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	_, err = r.CreateRemote(&config.RemoteConfig{
		Name:   name,
		URLs:   []string{url},
		Mirror: isMirror,
	})

	if err != nil {
		logger.ErrorWithErr("Failed to add remote", err, logrus.Fields{"name": name})
		return err
	}

	logger.Info("Remote added successfully", logrus.Fields{"name": name})
	return nil
}

// RemoveRemote 删除远程仓库
func (s *GitService) RemoveRemote(path, name string) error {
	logger.Info("Removing remote", logrus.Fields{"path": path, "name": name})

	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	err = r.DeleteRemote(name)
	if err != nil {
		logger.ErrorWithErr("Failed to remove remote", err, logrus.Fields{"name": name})
		return err
	}

	logger.Info("Remote removed successfully", logrus.Fields{"name": name})
	return nil
}

// SetRemotePushURL 设置远程仓库的推送 URL
func (s *GitService) SetRemotePushURL(path, name, url string) error {
	logger.Info("Setting remote push URL", logrus.Fields{
		"path": path,
		"name": name,
		"url":  url,
	})

	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	cfg, err := r.Config()
	if err != nil {
		return err
	}

	if remote, ok := cfg.Remotes[name]; ok {
		remote.URLs = []string{url}
		return r.Storer.SetConfig(cfg)
	}

	return fmt.Errorf("remote %s not found", name)
}

// GetRepoConfig 获取仓库配置信息
func (s *GitService) GetRepoConfig(path string) (*domain.GitRepoConfig, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	cfg, err := r.Config()
	if err != nil {
		return nil, err
	}

	repoConfig := &domain.GitRepoConfig{
		Remotes:  []domain.GitRemote{},
		Branches: []domain.GitBranch{},
	}

	for _, remote := range cfg.Remotes {
		gitRemote := &domain.GitRemote{
			Name:       remote.Name,
			FetchURL:   "",
			PushURL:    "",
			FetchSpecs: []string{},
			PushSpecs:  []string{},
			IsMirror:   remote.Mirror,
		}
		if len(remote.URLs) > 0 {
			gitRemote.FetchURL = remote.URLs[0]
			gitRemote.PushURL = remote.URLs[0]
		}
		for _, u := range remote.URLs {
			gitRemote.FetchSpecs = append(gitRemote.FetchSpecs, u)
		}
		for _, spec := range remote.Fetch {
			gitRemote.FetchSpecs = append(gitRemote.FetchSpecs, spec.String())
		}
		repoConfig.Remotes = append(repoConfig.Remotes, *gitRemote)
	}

	for _, branch := range cfg.Branches {
		b := &domain.GitBranch{
			Name:   branch.Name,
			Remote: branch.Remote,
			Merge:  branch.Merge.String(),
		}
		if branch.Remote != "" && branch.Merge != "" {
			shortRef := branch.Merge.Short()
			b.UpstreamRef = fmt.Sprintf("%s/%s", branch.Remote, shortRef)
		}
		repoConfig.Branches = append(repoConfig.Branches, *b)
	}

	logger.Debug("Repo config retrieved", logrus.Fields{
		"path":     path,
		"remotes":  len(repoConfig.Remotes),
		"branches": len(repoConfig.Branches),
	})
	return repoConfig, nil
}

// ListRemoteBranches 获取指定远程的所有分支名（基于本地 remote-tracking refs）
func (s *GitService) ListRemoteBranches(path, remoteName string) ([]string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	prefix := "refs/remotes/" + remoteName + "/"
	iter, err := r.References()
	if err != nil {
		return nil, err
	}

	var branches []string
	iter.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().String()
		if strings.HasPrefix(name, prefix) {
			branch := strings.TrimPrefix(name, prefix)
			if branch != "HEAD" {
				branches = append(branches, branch)
			}
		}
		return nil
	})

	logger.Debug("Remote branches listed", logrus.Fields{"path": path, "remote": remoteName, "count": len(branches)})
	return branches, nil
}

// TestRemoteConnection 测试远程连接
func (s *GitService) TestRemoteConnection(url string) error {
	logger.Info("Testing remote connection", logrus.Fields{"url": url})

	remote := git.NewRemote(nil, &config.RemoteConfig{
		Name: "anonymous",
		URLs: []string{url},
	})

	auth := s.detectSSHAuth(url)

	_, err := remote.List(&git.ListOptions{
		Auth: auth,
	})

	if err != nil {
		logger.ErrorWithErr("Remote connection test failed", err, logrus.Fields{"url": url})
		return err
	}

	logger.Info("Remote connection test successful", logrus.Fields{"url": url})
	return nil
}
