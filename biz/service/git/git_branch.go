package git

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
)

// ListBranchesWithInfo returns detailed information for all branches
func (s *GitService) ListBranchesWithInfo(path string) ([]domain.BranchInfo, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	iter, err := r.References()
	if err != nil {
		return nil, err
	}

	headRef, err := r.Head()
	var headHash plumbing.Hash
	if err == nil {
		headHash = headRef.Hash()
	}

	cfg, err := r.Config()
	if err != nil {
		return nil, err
	}

	var branches []domain.BranchInfo

	err = iter.ForEach(func(ref *plumbing.Reference) error {
		// 只处理本地分支 (refs/heads/*) 和远程跟踪分支 (refs/remotes/*)
		isBranch := ref.Name().IsBranch()
		isRemote := ref.Name().IsRemote()

		if !isBranch && !isRemote {
			return nil
		}

		name := ref.Name().Short()
		hash := ref.Hash()

		b := domain.BranchInfo{
			Name: name,
			Hash: hash.String(),
		}

		// 明确判断分支类型
		if isBranch {
			b.Type = "local"
		} else if isRemote {
			b.Type = "remote"
		} else {
			// 不应该到这里，但保险起见跳过
			return nil
		}

		if hash == headHash && ref.Name().IsBranch() {
			b.IsCurrent = true
		}

		// Commit Info
		commit, err := r.CommitObject(hash)
		if err == nil {
			b.Author = commit.Author.Name
			b.AuthorEmail = commit.Author.Email
			b.Date = commit.Author.When
			b.Message = strings.TrimSpace(strings.Split(commit.Message, "\n")[0])
		}

		// Upstream Info (only for local branches)
		if ref.Name().IsBranch() {
			if branchCfg, ok := cfg.Branches[name]; ok {
				if branchCfg.Remote != "" && branchCfg.Merge != "" {
					shortMerge := branchCfg.Merge.Short()
					b.Upstream = fmt.Sprintf("%s/%s", branchCfg.Remote, shortMerge)
				}
			}
		}

		branches = append(branches, b)
		return nil
	})

	return branches, nil
}

func (s *GitService) CreateBranch(path, name, base string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	var hash plumbing.Hash
	if base != "" {
		// Resolve base
		h, err := r.ResolveRevision(plumbing.Revision(base))
		if err != nil {
			return err
		}
		hash = *h
	} else {
		// Default to HEAD
		head, err := r.Head()
		if err != nil {
			return err
		}
		hash = head.Hash()
	}

	refName := plumbing.ReferenceName("refs/heads/" + name)
	return r.Storer.SetReference(plumbing.NewHashReference(refName, hash))
}

func (s *GitService) DeleteBranch(path, name string, force bool) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	refName := plumbing.ReferenceName("refs/heads/" + name)
	return r.Storer.RemoveReference(refName)
}

func (s *GitService) RenameBranch(path, oldName, newName string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	oldRefName := plumbing.ReferenceName("refs/heads/" + oldName)
	newRefName := plumbing.ReferenceName("refs/heads/" + newName)

	ref, err := r.Reference(oldRefName, true)
	if err != nil {
		return err
	}

	// Create new
	err = r.Storer.SetReference(plumbing.NewHashReference(newRefName, ref.Hash()))
	if err != nil {
		return err
	}

	// Delete old
	return r.Storer.RemoveReference(oldRefName)
}

func (s *GitService) GetBranchDescription(path, branch string) (string, error) {
	out, err := s.RunCommand(path, "config", fmt.Sprintf("branch.%s.description", branch))
	if err != nil {
		// Config key might not exist
		return "", nil
	}
	return out, nil
}

func (s *GitService) SetBranchDescription(path, branch, desc string) error {
	_, err := s.RunCommand(path, "config", fmt.Sprintf("branch.%s.description", branch), desc)
	return err
}

// GetBranchMetrics returns simple metrics: commit count, lines of code (approx)
// This is expensive, use sparingly
func (s *GitService) GetBranchMetrics(path, branch string) (map[string]int, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	commit, err := s.resolveCommit(r, branch)
	if err != nil {
		return nil, err
	}

	// Commit count
	// Efficient counting is hard.
	// We'll use Log and count. Limit to avoiding timeout?
	cIter, err := r.Log(&git.LogOptions{From: commit.Hash})
	if err != nil {
		return nil, err
	}

	count := 0
	err = cIter.ForEach(func(c *object.Commit) error {
		count++
		return nil
	})

	return map[string]int{
		"commit_count": count,
	}, nil
}
