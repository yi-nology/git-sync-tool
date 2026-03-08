package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// AuthorInfo 作者信息
type AuthorInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetAuthors 获取仓库的所有提交作者列表
func (s *GitService) GetAuthors(path string) ([]AuthorInfo, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	// 获取所有提交
	iter, err := r.Log(&git.LogOptions{All: true})
	if err != nil {
		return nil, err
	}

	// 使用 map 去重
	authorMap := make(map[string]AuthorInfo)

	iter.ForEach(func(c *object.Commit) error {
		key := c.Author.Name + "|" + c.Author.Email
		if _, exists := authorMap[key]; !exists {
			authorMap[key] = AuthorInfo{
				Name:  c.Author.Name,
				Email: c.Author.Email,
			}
		}
		return nil
	})

	// 转换为切片
	authors := make([]AuthorInfo, 0, len(authorMap))
	for _, author := range authorMap {
		authors = append(authors, author)
	}

	return authors, nil
}

// CherryPick 执行cherry-pick操作
func (s *GitService) CherryPick(path, commitHash string, noCommit bool) (string, []string, error) {
	args := []string{"cherry-pick"}
	if noCommit {
		args = append(args, "-n")
	}
	args = append(args, commitHash)

	output, err := s.RunCommand(path, args...)
	if err != nil {
		// 检查是否是冲突
		if strings.Contains(output, "conflict") || strings.Contains(output, "CONFLICT") {
			// 获取冲突文件列表
			conflicts := s.getConflictFiles(path)
			return "", conflicts, fmt.Errorf("cherry-pick conflict")
		}
		return "", nil, err
	}

	// 获取新的commit hash
	if !noCommit {
		newHash, _ := s.RunCommand(path, "rev-parse", "HEAD")
		return newHash, nil, nil
	}
	return "", nil, nil
}

// CherryPickAbort 中止cherry-pick
func (s *GitService) CherryPickAbort(path string) error {
	_, err := s.RunCommand(path, "cherry-pick", "--abort")
	return err
}

// Rebase 执行rebase操作
func (s *GitService) Rebase(path, upstream, onto string) (bool, []string, error) {
	args := []string{"rebase"}
	if onto != "" {
		args = append(args, "--onto", onto)
	}
	args = append(args, upstream)

	output, err := s.RunCommand(path, args...)
	if err != nil {
		// 检查是否是冲突
		if strings.Contains(output, "conflict") || strings.Contains(output, "CONFLICT") {
			conflicts := s.getConflictFiles(path)
			return false, conflicts, nil
		}
		return false, nil, err
	}
	return true, nil, nil
}

// RebaseAbort 中止rebase
func (s *GitService) RebaseAbort(path string) error {
	_, err := s.RunCommand(path, "rebase", "--abort")
	return err
}

// RebaseContinue 继续rebase
func (s *GitService) RebaseContinue(path string) (bool, []string, error) {
	output, err := s.RunCommand(path, "rebase", "--continue")
	if err != nil {
		if strings.Contains(output, "conflict") || strings.Contains(output, "CONFLICT") {
			conflicts := s.getConflictFiles(path)
			return false, conflicts, nil
		}
		return false, nil, err
	}
	return true, nil, nil
}

// RebaseSkip 跳过当前commit
func (s *GitService) RebaseSkip(path string) error {
	_, err := s.RunCommand(path, "rebase", "--skip")
	return err
}

// IsRebaseInProgress 检查是否有进行中的rebase
func (s *GitService) IsRebaseInProgress(path string) bool {
	// 检查 .git/rebase-merge 或 .git/rebase-apply 目录
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(filepath.Join(gitDir, "rebase-merge")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(gitDir, "rebase-apply")); err == nil {
		return true
	}
	return false
}

// getConflictFiles 获取冲突文件列表
func (s *GitService) getConflictFiles(path string) []string {
	output, err := s.RunCommand(path, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil
	}
	if output == "" {
		return nil
	}
	return strings.Split(strings.TrimSpace(output), "\n")
}
