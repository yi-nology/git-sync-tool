// biz/service/git/git_commit.go - Git Commit搜索服务

package git

import (
	"io"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// CommitDetail 提交详情
type CommitDetail struct {
	Hash           string   `json:"hash"`
	ShortHash      string   `json:"short_hash"`
	Message        string   `json:"message"`
	AuthorName     string   `json:"author_name"`
	AuthorEmail    string   `json:"author_email"`
	AuthorDate     string   `json:"author_date"`
	CommitterName  string   `json:"committer_name"`
	CommitterEmail string   `json:"committer_email"`
	CommitterDate  string   `json:"committer_date"`
	ParentHashes   []string `json:"parent_hashes"`
	FilesChanged   int      `json:"files_changed"`
	Additions      int      `json:"additions"`
	Deletions      int      `json:"deletions"`
}

// FileChange 文件变更
type FileChange struct {
	Path      string `json:"path"`
	Status    string `json:"status"` // added, modified, deleted, renamed
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	OldPath   string `json:"old_path,omitempty"`
}

// SearchCommitsOptions 搜索选项
type SearchCommitsOptions struct {
	Ref      string
	Author   string
	Keyword  string
	Since    string
	Until    string
	Path     string
	Page     int
	PageSize int
}

// SearchCommits 搜索提交
func (s *GitService) SearchCommits(repoPath string, opts SearchCommitsOptions) ([]CommitDetail, int64, error) {
	r, err := s.openRepo(repoPath)
	if err != nil {
		return nil, 0, err
	}

	// 解析ref
	ref := opts.Ref
	if ref == "" {
		ref = "HEAD"
	}

	hash, err := r.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return nil, 0, err
	}

	// 构建LogOptions
	logOpts := &git.LogOptions{From: *hash}

	// 文件路径过滤
	if opts.Path != "" {
		path := strings.TrimPrefix(opts.Path, "/")
		logOpts.FileName = &path
	}

	// 时间范围过滤
	if opts.Since != "" {
		if t, err := time.Parse("2006-01-02", opts.Since); err == nil {
			logOpts.Since = &t
		}
	}
	if opts.Until != "" {
		if t, err := time.Parse("2006-01-02", opts.Until); err == nil {
			t = t.Add(24 * time.Hour) // 包含当天
			logOpts.Until = &t
		}
	}

	iter, err := r.Log(logOpts)
	if err != nil {
		return nil, 0, err
	}

	// 分页
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PageSize < 1 {
		opts.PageSize = 50
	}

	var commits []CommitDetail
	var total int64 = 0
	skip := (opts.Page - 1) * opts.PageSize

	err = iter.ForEach(func(c *object.Commit) error {
		// 作者过滤
		if opts.Author != "" {
			if !strings.Contains(strings.ToLower(c.Author.Name), strings.ToLower(opts.Author)) &&
				!strings.Contains(strings.ToLower(c.Author.Email), strings.ToLower(opts.Author)) {
				return nil
			}
		}

		// 关键词过滤
		if opts.Keyword != "" {
			if !strings.Contains(strings.ToLower(c.Message), strings.ToLower(opts.Keyword)) {
				return nil
			}
		}

		total++

		// 跳过前面的
		if total <= int64(skip) {
			return nil
		}

		// 超出页面大小
		if len(commits) >= opts.PageSize {
			return nil
		}

		// 构建CommitDetail
		var parentHashes []string
		for _, p := range c.ParentHashes {
			parentHashes = append(parentHashes, p.String())
		}

		commits = append(commits, CommitDetail{
			Hash:           c.Hash.String(),
			ShortHash:      c.Hash.String()[:7],
			Message:        strings.TrimSpace(c.Message),
			AuthorName:     c.Author.Name,
			AuthorEmail:    c.Author.Email,
			AuthorDate:     c.Author.When.Format("2006-01-02 15:04:05"),
			CommitterName:  c.Committer.Name,
			CommitterEmail: c.Committer.Email,
			CommitterDate:  c.Committer.When.Format("2006-01-02 15:04:05"),
			ParentHashes:   parentHashes,
		})

		return nil
	})

	if err != nil && err != io.EOF {
		return nil, 0, err
	}

	return commits, total, nil
}

// GetCommitDetail 获取提交详情
func (s *GitService) GetCommitDetail(repoPath, hashStr string) (*CommitDetail, []FileChange, error) {
	r, err := s.openRepo(repoPath)
	if err != nil {
		return nil, nil, err
	}

	commit, err := r.CommitObject(plumbing.NewHash(hashStr))
	if err != nil {
		return nil, nil, err
	}

	// 构建CommitDetail
	var parentHashes []string
	for _, p := range commit.ParentHashes {
		parentHashes = append(parentHashes, p.String())
	}

	detail := &CommitDetail{
		Hash:           commit.Hash.String(),
		ShortHash:      commit.Hash.String()[:7],
		Message:        strings.TrimSpace(commit.Message),
		AuthorName:     commit.Author.Name,
		AuthorEmail:    commit.Author.Email,
		AuthorDate:     commit.Author.When.Format("2006-01-02 15:04:05"),
		CommitterName:  commit.Committer.Name,
		CommitterEmail: commit.Committer.Email,
		CommitterDate:  commit.Committer.When.Format("2006-01-02 15:04:05"),
		ParentHashes:   parentHashes,
	}

	// 获取文件变更
	var changes []FileChange
	var parentTree *object.Tree

	if len(commit.ParentHashes) > 0 {
		parent, err := commit.Parent(0)
		if err == nil {
			parentTree, _ = parent.Tree()
		}
	}

	commitTree, err := commit.Tree()
	if err != nil {
		return detail, nil, nil
	}

	if parentTree != nil {
		diff, err := parentTree.Diff(commitTree)
		if err == nil {
			for _, d := range diff {
				change := FileChange{}

				from, to, err := d.Files()
				if err != nil {
					continue
				}

				if from == nil && to != nil {
					change.Status = "added"
					change.Path = to.Name
				} else if from != nil && to == nil {
					change.Status = "deleted"
					change.Path = from.Name
				} else if from != nil && to != nil {
					if from.Name != to.Name {
						change.Status = "renamed"
						change.OldPath = from.Name
						change.Path = to.Name
					} else {
						change.Status = "modified"
						change.Path = to.Name
					}
				}

				// 统计行数变更
				patch, err := d.Patch()
				if err == nil {
					for _, fp := range patch.FilePatches() {
						for _, chunk := range fp.Chunks() {
							content := chunk.Content()
							lines := strings.Split(content, "\n")
							switch chunk.Type() {
							case 1: // Add
								change.Additions += len(lines)
								detail.Additions += len(lines)
							case 2: // Delete
								change.Deletions += len(lines)
								detail.Deletions += len(lines)
							}
						}
					}
				}

				changes = append(changes, change)
			}
		}
	} else {
		// 第一个commit，所有文件都是added
		commitTree.Files().ForEach(func(f *object.File) error {
			lines, _ := f.Lines()
			changes = append(changes, FileChange{
				Path:      f.Name,
				Status:    "added",
				Additions: len(lines),
			})
			detail.Additions += len(lines)
			return nil
		})
	}

	detail.FilesChanged = len(changes)

	return detail, changes, nil
}

// GetCommitDiff 获取提交的diff
func (s *GitService) GetCommitDiff(repoPath, hashStr, filePath string) (string, error) {
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
		// 如果指定了文件路径，只返回该文件的diff
		if filePath != "" {
			from, to, _ := d.Files()
			var matchPath string
			if to != nil {
				matchPath = to.Name
			} else if from != nil {
				matchPath = from.Name
			}
			if matchPath != strings.TrimPrefix(filePath, "/") {
				continue
			}
		}

		patch, err := d.Patch()
		if err != nil {
			continue
		}
		result.WriteString(patch.String())
	}

	return result.String(), nil
}
