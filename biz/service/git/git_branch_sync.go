package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// GetBranchSyncStatus returns ahead/behind counts against upstream
func (s *GitService) GetBranchSyncStatus(path, branch, upstream string) (int, int, error) {
	if upstream == "" {
		return 0, 0, nil
	}
	r, err := s.openRepo(path)
	if err != nil {
		return 0, 0, err
	}

	// Resolve refs
	// branch is usually local name, upstream is remote/branch
	hBranch, err := r.ResolveRevision(plumbing.Revision(branch))
	if err != nil {
		// Try refs/heads/branch
		hBranch, err = r.ResolveRevision(plumbing.Revision("refs/heads/" + branch))
		if err != nil {
			return 0, 0, nil
		}
	}

	hUpstream, err := r.ResolveRevision(plumbing.Revision(upstream))
	if err != nil {
		// Try refs/remotes/upstream
		hUpstream, err = r.ResolveRevision(plumbing.Revision("refs/remotes/" + upstream))
		if err != nil {
			return 0, 0, nil
		}
	}

	cBranch, err := r.CommitObject(*hBranch)
	if err != nil {
		return 0, 0, err
	}
	cUpstream, err := r.CommitObject(*hUpstream)
	if err != nil {
		return 0, 0, err
	}

	bases, err := cBranch.MergeBase(cUpstream)
	if err != nil || len(bases) == 0 {
		return 0, 0, nil
	}
	base := bases[0]

	// Count ahead
	ahead := 0
	iter, err := r.Log(&git.LogOptions{From: *hBranch})
	if err == nil {
		iter.ForEach(func(c *object.Commit) error {
			if c.Hash == base.Hash {
				return fmt.Errorf("stop")
			}
			ahead++
			return nil
		})
	}

	// Count behind
	behind := 0
	iter, err = r.Log(&git.LogOptions{From: *hUpstream})
	if err == nil {
		iter.ForEach(func(c *object.Commit) error {
			if c.Hash == base.Hash {
				return fmt.Errorf("stop")
			}
			behind++
			return nil
		})
	}

	return ahead, behind, nil
}

// PushBranch pushes local branch to remote
func (s *GitService) PushBranch(path, remote, branch string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	// git push remote branch
	// Default refspec: refs/heads/branch:refs/heads/branch
	refSpec := config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch))

	// Detect Auth
	var auth transport.AuthMethod
	rem, err := r.Remote(remote)
	if err == nil {
		urls := rem.Config().URLs
		if len(urls) > 0 {
			auth = s.detectSSHAuth(urls[0])
		}
	}

	err = r.Push(&git.PushOptions{
		RemoteName: remote,
		RefSpecs:   []config.RefSpec{refSpec},
		Auth:       auth,
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

// PullBranch pulls changes from upstream
func (s *GitService) PullBranch(path, remote, branch string) error {
	// Replaces git pull --rebase with Worktree.Pull (Merge)
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	// Detect Auth
	var auth transport.AuthMethod
	rem, err := r.Remote(remote)
	if err == nil {
		urls := rem.Config().URLs
		if len(urls) > 0 {
			auth = s.detectSSHAuth(urls[0])
		}
	}

	err = w.Pull(&git.PullOptions{
		RemoteName:    remote,
		ReferenceName: plumbing.ReferenceName("refs/heads/" + branch),
		Auth:          auth,
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

// UpdateBranchFastForward fetches remote branch and updates local branch if fast-forward possible.
// Used for updating non-current branches.
func (s *GitService) UpdateBranchFastForward(path, remote, branch, remoteBranch string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	// Detect Auth
	var auth transport.AuthMethod
	rem, err := r.Remote(remote)
	if err == nil {
		urls := rem.Config().URLs
		if len(urls) > 0 {
			auth = s.detectSSHAuth(urls[0])
		}
	}

	// git fetch remote remoteBranch:branch
	// e.g. git fetch origin main:main
	refSpec := config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", remoteBranch, branch))

	err = rem.Fetch(&git.FetchOptions{
		RemoteName: remote,
		RefSpecs:   []config.RefSpec{refSpec},
		Auth:       auth,
	})

	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

// FetchAll fetches all remotes
func (s *GitService) FetchAll(path string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	remotes, err := r.Remotes()
	if err != nil {
		return err
	}

	for _, remote := range remotes {
		var auth transport.AuthMethod
		urls := remote.Config().URLs
		if len(urls) > 0 {
			auth = s.detectSSHAuth(urls[0])
		}

		err := remote.Fetch(&git.FetchOptions{
			Auth: auth,
			RefSpecs: []config.RefSpec{
				config.RefSpec("+refs/heads/*:refs/remotes/" + remote.Config().Name + "/*"),
				config.RefSpec("+refs/tags/*:refs/tags/*"),
			},
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			// Log error but continue?
		}
	}
	return nil
}
