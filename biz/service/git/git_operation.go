package git

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

type channelWriter struct {
	ch chan string
}

func (w *channelWriter) Write(p []byte) (n int, err error) {
	if w.ch != nil {
		w.ch <- string(p)
	}
	return len(p), nil
}

func (s *GitService) openRepo(path string) (*git.Repository, error) {
	log.Printf("[DEBUG] Opening repository at: %s", path)
	r, err := git.PlainOpen(path)
	if err != nil {
		log.Printf("[ERROR] Failed to open repository at %s: %v", path, err)
		return nil, fmt.Errorf("failed to open repository at %s: %v", path, err)
	}
	log.Printf("[DEBUG] Repository opened successfully: %s", path)
	return r, nil
}

func (s *GitService) IsGitRepo(path string) bool {
	_, err := git.PlainOpen(path)
	return err == nil
}

func (s *GitService) Fetch(path, remote string, progress io.Writer) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	// Get remote URL to detect auth
	var auth transport.AuthMethod
	rem, err := r.Remote(remote)
	if err == nil {
		urls := rem.Config().URLs
		if len(urls) > 0 {
			auth = s.detectSSHAuth(urls[0])
		}
	}

	// 只在auth不为nil时传递认证信息
	fetchOptions := &git.FetchOptions{
		RemoteName: remote,
		Progress:   progress,
		RefSpecs: []config.RefSpec{
			config.RefSpec("+refs/heads/*:refs/remotes/" + remote + "/*"),
			config.RefSpec("+refs/tags/*:refs/tags/*"),
		},
	}
	if auth != nil {
		fetchOptions.Auth = auth
	}

	err = r.Fetch(fetchOptions)
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (s *GitService) FetchWithAuth(path, remoteURL, authType, authKey, authSecret string, progress io.Writer, extraArgs ...string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	auth, err := s.getAuth(authType, authKey, authSecret)
	if err != nil {
		return err
	}

	// Create a temporary remote to fetch from the URL
	remote := git.NewRemote(r.Storer, &config.RemoteConfig{
		Name: "origin",
		URLs: []string{remoteURL},
	})

	err = remote.Fetch(&git.FetchOptions{
		Auth:       auth,
		RemoteName: "origin",
		Progress:   progress,
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

// FetchWithAuthMethod 使用已解析的认证方法进行 fetch
func (s *GitService) FetchWithAuthMethod(path, remoteURL string, auth transport.AuthMethod, progress io.Writer, extraArgs ...string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	// Create a temporary remote to fetch from the URL
	remote := git.NewRemote(r.Storer, &config.RemoteConfig{
		Name: "origin",
		URLs: []string{remoteURL},
	})

	err = remote.Fetch(&git.FetchOptions{
		Auth:       auth,
		RemoteName: "origin",
		Progress:   progress,
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (s *GitService) Clone(remoteURL, localPath, authType, authKey, authSecret string) error {
	return s.CloneWithProgress(remoteURL, localPath, authType, authKey, authSecret, nil)
}

func (s *GitService) CloneWithProgress(remoteURL, localPath, authType, authKey, authSecret string, progressChan chan string) error {
	auth, err := s.getAuth(authType, authKey, authSecret)
	if err != nil {
		return err
	}

	if auth == nil {
		auth = s.detectSSHAuth(remoteURL)
	}

	var progress io.Writer
	if progressChan != nil {
		progress = &channelWriter{ch: progressChan}
	}

	_, err = git.PlainClone(localPath, false, &git.CloneOptions{
		URL:      remoteURL,
		Auth:     auth,
		Progress: progress,
	})
	return err
}

// CloneWithAuthMethod 使用已解析的认证方法进行克隆
func (s *GitService) CloneWithAuthMethod(remoteURL, localPath string, auth transport.AuthMethod, progressChan chan string) error {
	if auth == nil {
		auth = s.detectSSHAuth(remoteURL)
	}

	var progress io.Writer
	if progressChan != nil {
		progress = &channelWriter{ch: progressChan}
	}

	_, err := git.PlainClone(localPath, false, &git.CloneOptions{
		URL:      remoteURL,
		Auth:     auth,
		Progress: progress,
	})
	return err
}

func (s *GitService) GetCommitHash(path, remote, branch string) (string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return "", err
	}
	// Only resolve as remote reference (do NOT fallback to local branch)
	refName := plumbing.ReferenceName(fmt.Sprintf("refs/remotes/%s/%s", remote, branch))
	ref, err := r.Reference(refName, true)
	if err != nil {
		return "", fmt.Errorf("remote branch %s/%s not found: %v", remote, branch, err)
	}
	return ref.Hash().String(), nil
}

func (s *GitService) IsAncestor(path, ancestor, descendant string) (bool, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return false, err
	}

	h1 := plumbing.NewHash(ancestor)
	h2 := plumbing.NewHash(descendant)

	c1, err := r.CommitObject(h1)
	if err != nil {
		return false, err
	}
	c2, err := r.CommitObject(h2)
	if err != nil {
		return false, err
	}

	return c1.IsAncestor(c2)
}

func parsePushOptions(options []string) *git.PushOptions {
	opts := &git.PushOptions{}
	opts.Options = make(map[string]string)

	for _, o := range options {
		if o == "-f" || o == "--force" {
			opts.Force = true
		} else if o == "--prune" {
			opts.Prune = true
		} else if strings.HasPrefix(o, "--push-option=") {
			// --push-option=key=value
			kv := strings.TrimPrefix(o, "--push-option=")
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) == 2 {
				opts.Options[parts[0]] = parts[1]
			}
		}
	}
	return opts
}

func (s *GitService) Push(path, targetRemote, sourceHash, targetBranch string, options []string, progress io.Writer) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	// sourceHash might be a commit hash. We need to map it to the remote branch.
	// git push remote hash:refs/heads/branch
	refSpec := config.RefSpec(fmt.Sprintf("%s:refs/heads/%s", sourceHash, targetBranch))

	// Detect Auth
	var auth transport.AuthMethod
	rem, err := r.Remote(targetRemote)
	if err == nil {
		urls := rem.Config().URLs
		if len(urls) > 0 {
			auth = s.detectSSHAuth(urls[0])
		}
	}

	pushOpts := parsePushOptions(options)
	pushOpts.RemoteName = targetRemote
	pushOpts.RefSpecs = []config.RefSpec{refSpec}
	if auth != nil {
		pushOpts.Auth = auth
	}
	pushOpts.Progress = progress

	err = r.Push(pushOpts)
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (s *GitService) GetBranches(path string) ([]string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}
	iter, err := r.References()
	if err != nil {
		return nil, err
	}
	var branches []string
	iter.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() || ref.Name().IsRemote() {
			branches = append(branches, ref.Name().Short())
		}
		return nil
	})
	return branches, nil
}

func (s *GitService) GetCommits(path, branch, since, until string) (string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return "", err
	}

	// Resolve branch
	commit, err := s.resolveCommit(r, branch)
	if err != nil {
		return "", err
	}

	cIter, err := r.Log(&git.LogOptions{From: commit.Hash})
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	// Format: %H|%an|%ae|%ad|%s (Hash|AuthorName|AuthorEmail|Date|Subject)
	err = cIter.ForEach(func(c *object.Commit) error {
		// Filter by since/until if needed (parsing dates is annoying)
		// For now, skip date filtering or implement it.
		// since/until are strings like "2023-01-01".

		line := fmt.Sprintf("%s|%s|%s|%s|%s\n",
			c.Hash.String(),
			c.Author.Name,
			c.Author.Email,
			c.Author.When.Format("2006-01-02 15:04:05 -0700"),    // ISO-ish
			strings.TrimSpace(strings.Split(c.Message, "\n")[0]), // Subject
		)
		sb.WriteString(line)
		return nil
	})

	return sb.String(), nil
}

func (s *GitService) PushWithAuth(path, targetRemoteURL, sourceHash, targetBranch, authType, authKey, authSecret string, options []string, progress io.Writer) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	auth, err := s.getAuth(authType, authKey, authSecret)
	if err != nil {
		return err
	}

	remote := git.NewRemote(r.Storer, &config.RemoteConfig{
		Name: "anonymous",
		URLs: []string{targetRemoteURL},
	})

	refSpec := config.RefSpec(fmt.Sprintf("%s:refs/heads/%s", sourceHash, targetBranch))

	pushOpts := parsePushOptions(options)
	pushOpts.Auth = auth
	pushOpts.RefSpecs = []config.RefSpec{refSpec}
	pushOpts.Progress = progress

	err = remote.Push(pushOpts)
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

// PushWithAuthMethod 使用已解析的认证方法进行 push
func (s *GitService) PushWithAuthMethod(path, targetRemoteURL, sourceHash, targetBranch string, auth transport.AuthMethod, options []string, progress io.Writer) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	remote := git.NewRemote(r.Storer, &config.RemoteConfig{
		Name: "anonymous",
		URLs: []string{targetRemoteURL},
	})

	refSpec := config.RefSpec(fmt.Sprintf("%s:refs/heads/%s", sourceHash, targetBranch))

	pushOpts := parsePushOptions(options)
	pushOpts.Auth = auth
	pushOpts.RefSpecs = []config.RefSpec{refSpec}
	pushOpts.Progress = progress

	err = remote.Push(pushOpts)
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (s *GitService) GetRepoFiles(path, branch string) ([]string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	// Resolve branch to commit -> tree
	commit, err := s.resolveCommit(r, branch)
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	var files []string
	tree.Files().ForEach(func(f *object.File) error {
		files = append(files, f.Name)
		return nil
	})
	return files, nil
}

func (s *GitService) BlameFile(path, branch, file string) (*git.BlameResult, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	commit, err := s.resolveCommit(r, branch)
	if err != nil {
		return nil, err
	}

	return git.Blame(commit, file)
}

func (s *GitService) CheckoutBranch(path, branch string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	return w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + branch),
		Force:  true,
	})
}

func (s *GitService) GetStatus(path string) (string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return "", err
	}
	w, err := r.Worktree()
	if err != nil {
		return "", err
	}
	status, err := w.Status()
	if err != nil {
		return "", err
	}
	return status.String(), nil
}

func (s *GitService) AddAll(path string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	// Add(".") in go-git
	_, err = w.Add(".")
	return err
}

// AddFiles stages specific files for commit
func (s *GitService) AddFiles(path string, files []string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	for _, f := range files {
		if _, err := w.Add(f); err != nil {
			return fmt.Errorf("failed to add %s: %w", f, err)
		}
	}
	return nil
}

func (s *GitService) Commit(path, message, authorName, authorEmail string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	// Use provided author info, or fallback to default
	if authorName == "" {
		authorName = "Git Manage Service"
	}
	if authorEmail == "" {
		authorEmail = "git-manage@example.com"
	}

	_, err = w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  time.Now(),
		},
	})
	return err
}

func (s *GitService) PushCurrent(path string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	// Detect Auth for default remote (origin)
	var auth transport.AuthMethod
	rem, err := r.Remote("origin")
	if err == nil {
		urls := rem.Config().URLs
		if len(urls) > 0 {
			auth = s.detectSSHAuth(urls[0])
		}
	}

	err = r.Push(&git.PushOptions{
		Auth: auth,
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (s *GitService) Reset(path string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	return w.Reset(&git.ResetOptions{Mode: git.MixedReset})
}

func (s *GitService) GetLogIterator(path, branch string) (object.CommitIter, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}
	hash, err := r.ResolveRevision(plumbing.Revision(branch))
	if err != nil {
		return nil, err
	}
	return r.Log(&git.LogOptions{From: *hash})
}

func (s *GitService) GetLogStats(path, branch string) (string, error) {
	// git log --numstat --no-merges --pretty=format:"COMMIT|%H|%aN|%aE|%at" <branch>
	return s.RunCommand(path, "log", "--numstat", "--no-merges", "--pretty=format:COMMIT|%H|%aN|%aE|%at", branch)
}

func (s *GitService) GetLogStatsStream(path, branch string) (io.ReadCloser, error) {
	// git log --numstat --no-merges --pretty=format:"COMMIT|%H|%aN|%aE|%at" <branch>
	cmd := exec.Command("git", "log", "--numstat", "--no-merges", "--pretty=format:COMMIT|%H|%aN|%aE|%at", branch)
	cmd.Dir = path
	// Prevent password prompts and force English output
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "LC_ALL=C")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &cmdStream{
		cmd:    cmd,
		stdout: stdout,
	}, nil
}

type cmdStream struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

func (c *cmdStream) Read(p []byte) (n int, err error) {
	return c.stdout.Read(p)
}

func (c *cmdStream) Close() error {
	_ = c.stdout.Close()
	return c.cmd.Wait()
}

func (s *GitService) GetCommit(path, hashStr string) (*object.Commit, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}
	return r.CommitObject(plumbing.NewHash(hashStr))
}

func (s *GitService) resolveCommit(r *git.Repository, rev string) (*object.Commit, error) {
	hash, err := r.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		// Try adding refs/heads/ if simple name failed and it doesn't already look like a ref
		if !strings.HasPrefix(rev, "refs/") {
			h, err2 := r.ResolveRevision(plumbing.Revision("refs/heads/" + rev))
			if err2 == nil {
				hash = h
				err = nil
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return r.CommitObject(*hash)
}

func (s *GitService) resolveCommitPair(r *git.Repository, base, target string) (*object.Commit, *object.Commit, error) {
	cBase, err := s.resolveCommit(r, base)
	if err != nil {
		return nil, nil, err
	}
	cTarget, err := s.resolveCommit(r, target)
	if err != nil {
		return nil, nil, err
	}
	return cBase, cTarget, nil
}

func (s *GitService) ResolveRevision(path, rev string) (string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return "", err
	}
	hash, err := r.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return "", err
	}
	return hash.String(), nil
}

func (s *GitService) GetHeadBranch(path string) (string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return "", err
	}
	head, err := r.Head()
	if err != nil {
		return "", err
	}
	// refs/heads/master -> master
	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}
	// Detached HEAD or other state
	return head.Hash().String(), nil
}
