package git

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
	"github.com/yi-nology/git-manage-service/pkg/logger"
)

// TagInfo Tag 信息结构
type TagInfo struct {
	Name    string    `json:"name"`
	Hash    string    `json:"hash"`
	Message string    `json:"message"`
	Tagger  string    `json:"tagger"`
	Date    time.Time `json:"date"`
}

// CreateTag 创建标签
func (s *GitService) CreateTag(path, tagName, ref, message, authorName, authorEmail string) error {
	logger.Info("Creating tag", logrus.Fields{
		"path":    path,
		"tag":     tagName,
		"ref":     ref,
		"message": message,
	})

	r, err := s.openRepo(path)
	if err != nil {
		logger.ErrorWithErr("Failed to open repo for tag creation", err, logrus.Fields{"path": path})
		return err
	}

	hash, err := r.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		logger.ErrorWithErr("Invalid reference for tag", err, logrus.Fields{"ref": ref})
		return fmt.Errorf("invalid reference '%s': %v", ref, err)
	}

	if authorName == "" {
		authorName = "Git Manage Service"
	}
	if authorEmail == "" {
		authorEmail = "git-manage@example.com"
	}

	_, err = r.CreateTag(tagName, *hash, &git.CreateTagOptions{
		Tagger: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  time.Now(),
		},
		Message: message,
	})

	if err != nil {
		logger.ErrorWithErr("Failed to create tag", err, logrus.Fields{"tag": tagName})
		return err
	}

	logger.Info("Tag created successfully", logrus.Fields{"tag": tagName})
	return nil
}

// PushTag 推送标签到远程
func (s *GitService) PushTag(path, remoteName, tagName, authType, authKey, authSecret string) error {
	logger.Info("Pushing tag", logrus.Fields{
		"path":   path,
		"remote": remoteName,
		"tag":    tagName,
	})

	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	auth, err := s.getAuth(authType, authKey, authSecret)
	if err != nil {
		return err
	}

	if auth == nil {
		rem, err := r.Remote(remoteName)
		if err == nil {
			urls := rem.Config().URLs
			if len(urls) > 0 {
				auth = s.detectSSHAuth(urls[0])
			}
		}
	}

	refSpec := config.RefSpec(fmt.Sprintf("refs/tags/%s:refs/tags/%s", tagName, tagName))

	err = r.Push(&git.PushOptions{
		RemoteName: remoteName,
		RefSpecs:   []config.RefSpec{refSpec},
		Auth:       auth,
	})
	if err == git.NoErrAlreadyUpToDate {
		logger.Debug("Tag already up to date", logrus.Fields{"tag": tagName})
		return nil
	}
	if err != nil {
		logger.ErrorWithErr("Failed to push tag", err, logrus.Fields{"tag": tagName})
		return err
	}

	logger.Info("Tag pushed successfully", logrus.Fields{"tag": tagName, "remote": remoteName})
	return nil
}

// DeleteTag 删除本地标签
func (s *GitService) DeleteTag(path, tagName string) error {
	logger.Info("Deleting tag", logrus.Fields{"path": path, "tag": tagName})

	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	err = r.DeleteTag(tagName)
	if err != nil {
		logger.ErrorWithErr("Failed to delete tag", err, logrus.Fields{"tag": tagName})
		return err
	}

	logger.Info("Tag deleted successfully", logrus.Fields{"tag": tagName})
	return nil
}

// DeleteRemoteTag 删除远程标签
func (s *GitService) DeleteRemoteTag(path, remoteName, tagName, authType, authKey, authSecret string) error {
	logger.Info("Deleting remote tag", logrus.Fields{
		"path":   path,
		"remote": remoteName,
		"tag":    tagName,
	})

	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	auth, _ := s.getAuth(authType, authKey, authSecret)
	if auth == nil {
		rem, err := r.Remote(remoteName)
		if err == nil {
			urls := rem.Config().URLs
			if len(urls) > 0 {
				auth = s.detectSSHAuth(urls[0])
			}
		}
	}

	refSpec := config.RefSpec(fmt.Sprintf(":refs/tags/%s", tagName))
	err = r.Push(&git.PushOptions{
		RemoteName: remoteName,
		RefSpecs:   []config.RefSpec{refSpec},
		Auth:       auth,
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	if err != nil {
		logger.ErrorWithErr("Failed to delete remote tag", err, logrus.Fields{"tag": tagName})
		return err
	}

	logger.Info("Remote tag deleted successfully", logrus.Fields{"tag": tagName, "remote": remoteName})
	return nil
}

// GetTags 获取所有标签名
func (s *GitService) GetTags(path string) ([]string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	iter, err := r.Tags()
	if err != nil {
		return nil, err
	}

	var tags []string
	iter.ForEach(func(ref *plumbing.Reference) error {
		tags = append(tags, ref.Name().Short())
		return nil
	})

	logger.Debug("Tags retrieved", logrus.Fields{"path": path, "count": len(tags)})
	return tags, nil
}

// GetTagList 获取详细的标签列表
func (s *GitService) GetTagList(path string) ([]TagInfo, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	iter, err := r.Tags()
	if err != nil {
		return nil, err
	}

	var tags []TagInfo
	iter.ForEach(func(ref *plumbing.Reference) error {
		tagObj, err := r.TagObject(ref.Hash())
		if err == nil {
			// Annotated Tag
			tags = append(tags, TagInfo{
				Name:    ref.Name().Short(),
				Hash:    ref.Hash().String(),
				Message: tagObj.Message,
				Tagger:  tagObj.Tagger.Name,
				Date:    tagObj.Tagger.When,
			})
		} else {
			// Lightweight Tag (commit)
			commit, err := r.CommitObject(ref.Hash())
			if err == nil {
				tags = append(tags, TagInfo{
					Name:    ref.Name().Short(),
					Hash:    ref.Hash().String(),
					Message: commit.Message,
					Tagger:  commit.Author.Name,
					Date:    commit.Author.When,
				})
			}
		}
		return nil
	})

	logger.Debug("Tag list retrieved", logrus.Fields{"path": path, "count": len(tags)})
	return tags, nil
}
