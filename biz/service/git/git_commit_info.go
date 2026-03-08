package git

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type CommitInfo struct {
	Author         string    `json:"author"`
	AuthorEmail    string    `json:"authorEmail"`
	Committer      string    `json:"committer"`
	CommitterEmail string    `json:"committerEmail"`
	Message        string    `json:"message"`
	CommitTime     time.Time `json:"commitTime"`
}

// GetCommitInfo 获取提交信息
func (s *GitService) GetCommitInfo(repoPath, hashStr string) (*CommitInfo, error) {
	r, err := s.openRepo(repoPath)
	if err != nil {
		return nil, err
	}

	commit, err := r.CommitObject(plumbing.NewHash(hashStr))
	if err != nil {
		return nil, err
	}

	return &CommitInfo{
		Author:         commit.Author.Name,
		AuthorEmail:    commit.Author.Email,
		Committer:      commit.Committer.Name,
		CommitterEmail: commit.Committer.Email,
		Message:        commit.Message,
		CommitTime:     commit.Author.When,
	}, nil
}

// GetRecentCommits 获取最近的提交历史
func (s *GitService) GetRecentCommits(repoPath string, limit int) ([]string, error) {
	r, err := s.openRepo(repoPath)
	if err != nil {
		return nil, err
	}

	head, err := r.Head()
	if err != nil {
		return nil, err
	}

	commitIter, err := r.Log(&git.LogOptions{
		From: head.Hash(),
	})
	if err != nil {
		return nil, err
	}

	var commits []string
	count := 0
	err = commitIter.ForEach(func(commit *object.Commit) error {
		if count >= limit {
			return fmt.Errorf("limit reached")
		}
		commits = append(commits, commit.Hash.String())
		count++
		return nil
	})
	// 忽略limit reached错误
	if err != nil && !strings.Contains(err.Error(), "limit reached") {
		return nil, err
	}

	return commits, nil
}

// GetCommitDiffSimple 获取提交的diff（简化版，用于分析）
func (s *GitService) GetCommitDiffSimple(repoPath, hashStr string) (string, error) {
	r, err := s.openRepo(repoPath)
	if err != nil {
		return "", err
	}

	commit, err := r.CommitObject(plumbing.NewHash(hashStr))
	if err != nil {
		return "", err
	}

	var parentTree *object.Tree
	if len(commit.ParentHashes) > 0 {
		parent, err := commit.Parent(0)
		if err == nil {
			parentTree, _ = parent.Tree()
		}
	}

	commitTree, err := commit.Tree()
	if err != nil {
		return "", err
	}

	if parentTree == nil {
		// 创建空树
		parentTree = &object.Tree{}
	}

	diff, err := parentTree.Diff(commitTree)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	for _, d := range diff {
		diffStr, err := d.Patch()
		if err != nil {
			continue
		}
		result.WriteString(diffStr.String())
	}

	return result.String(), nil
}
