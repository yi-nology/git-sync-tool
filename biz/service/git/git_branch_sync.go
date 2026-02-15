package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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

// PushBranchWithDBKey pushes local branch to remote using database SSH key
func (s *GitService) PushBranchWithDBKey(path, remote, branch, privateKey, passphrase string) error {
	// Create temp private key file
	tmpFile, err := os.CreateTemp("", "git_ssh_key_*")
	if err != nil {
		return fmt.Errorf("failed to create temp key file: %v", err)
	}
	tmpKeyPath := tmpFile.Name()
	defer os.Remove(tmpKeyPath)

	// 确保私钥以换行符结尾
	keyContent := privateKey
	if !strings.HasSuffix(keyContent, "\n") {
		keyContent += "\n"
	}

	if _, err := tmpFile.WriteString(keyContent); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write key file: %v", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpKeyPath, 0600); err != nil {
		return fmt.Errorf("failed to set key file permissions: %v", err)
	}

	// 获取 remote URL 直接推送，避免 mirror 配置冲突
	urlCmd := exec.Command("git", "remote", "get-url", remote)
	urlCmd.Dir = path
	urlOutput, err := urlCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get remote URL: %v", err)
	}
	remoteURL := strings.TrimSpace(string(urlOutput))

	sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o IdentitiesOnly=yes", tmpKeyPath)

	// git push <url> refs/heads/branch:refs/heads/branch
	refSpec := fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch)
	cmd := exec.Command("git", "push", remoteURL, refSpec)
	cmd.Dir = path
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %v, output: %s", err, string(output))
	}

	return nil
}

// PullBranchWithDBKey pulls changes from remote using database SSH key
func (s *GitService) PullBranchWithDBKey(path, remote, branch, privateKey, passphrase string) error {
	// Create temp private key file
	tmpFile, err := os.CreateTemp("", "git_ssh_key_*")
	if err != nil {
		return fmt.Errorf("failed to create temp key file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(privateKey); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write key file: %v", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0600); err != nil {
		return fmt.Errorf("failed to set key file permissions: %v", err)
	}

	sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null", tmpFile.Name())

	// git pull remote branch
	cmd := exec.Command("git", "pull", remote, branch)
	cmd.Dir = path
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %v, output: %s", err, string(output))
	}

	return nil
}

// FetchAllWithDBKey fetches all remotes using database SSH key (native git command)
func (s *GitService) FetchAllWithDBKey(path, privateKey, passphrase string) error {
	tmpFile, err := os.CreateTemp("", "git_ssh_key_*")
	if err != nil {
		return fmt.Errorf("failed to create temp key file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	keyContent := privateKey
	if !strings.HasSuffix(keyContent, "\n") {
		keyContent += "\n"
	}

	if _, err := tmpFile.WriteString(keyContent); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write key file: %v", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0600); err != nil {
		return fmt.Errorf("failed to set key file permissions: %v", err)
	}

	sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o IdentitiesOnly=yes", tmpFile.Name())

	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = path
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch --all failed: %v, output: %s", err, string(output))
	}

	return nil
}

// FetchBranchWithDBKey fetches a specific branch from remote using database SSH key
func (s *GitService) FetchBranchWithDBKey(path, remote, branch, privateKey, passphrase string) error {
	// Create temp private key file
	tmpFile, err := os.CreateTemp("", "git_ssh_key_*")
	if err != nil {
		return fmt.Errorf("failed to create temp key file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(privateKey); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write key file: %v", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0600); err != nil {
		return fmt.Errorf("failed to set key file permissions: %v", err)
	}

	sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null", tmpFile.Name())

	// git fetch remote branch
	cmd := exec.Command("git", "fetch", remote, branch)
	cmd.Dir = path
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch failed: %v, output: %s", err, string(output))
	}

	return nil
}
