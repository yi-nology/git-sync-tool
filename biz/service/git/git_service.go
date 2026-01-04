package git

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	ssh2 "golang.org/x/crypto/ssh"

	"github.com/yi-nology/git-manage-service/biz/model/domain"
	conf "github.com/yi-nology/git-manage-service/pkg/configs"
)

type GitService struct{}

func NewGitService() *GitService {
	return &GitService{}
}

// RunCommand executes a raw git command.
// Deprecated: Ideally use go-git methods. However, kept for operations not fully supported by go-git (e.g. Merge logic, Config branch description).
func (s *GitService) RunCommand(dir string, args ...string) (string, error) {
	if conf.DebugMode {
		log.Printf("[DEBUG] Executing in %s: git %s", dir, strings.Join(args, " "))
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	// Prevent password prompts and force English output
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "LC_ALL=C")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("git command failed: %s, output: %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func (s *GitService) getAuth(authType, authKey, authSecret string) (transport.AuthMethod, error) {
	if authType == "http" && authKey != "" {
		return &http.BasicAuth{
			Username: authKey,
			Password: authSecret,
		}, nil
	} else if authType == "ssh" && authKey != "" {
		publicKeys, err := ssh.NewPublicKeysFromFile("git", authKey, "")
		if err != nil {
			return nil, err
		}
		publicKeys.HostKeyCallback = ssh2.InsecureIgnoreHostKey()
		return publicKeys, nil
	}
	return nil, nil
}

func (s *GitService) detectSSHAuth(urlStr string) transport.AuthMethod {
	// Simple check for SSH
	// git@... or ssh://...
	if !strings.HasPrefix(urlStr, "git@") && !strings.HasPrefix(urlStr, "ssh://") && !strings.Contains(urlStr, "git") {
		// Try parsing endpoint to be sure
		ep, err := transport.NewEndpoint(urlStr)
		if err != nil || ep.Protocol != "ssh" {
			return nil
		}
	}

	user := "git"
	ep, err := transport.NewEndpoint(urlStr)
	if err == nil && ep.User != "" {
		user = ep.User
	}

	if conf.DebugMode {
		log.Printf("[DEBUG] detectSSHAuth for %s (user: %s)", urlStr, user)
	}

	// 1. Try common key paths first (if they are unencrypted)
	home, err := os.UserHomeDir()
	if err == nil {
		keyPaths := []string{
			filepath.Join(home, ".ssh", "id_rsa"),
			filepath.Join(home, ".ssh", "id_ed25519"),
			filepath.Join(home, ".ssh", "id_ecdsa"),
		}

		for _, path := range keyPaths {
			if _, err := os.Stat(path); err == nil {
				// Try to load with empty password
				auth, err := ssh.NewPublicKeysFromFile(user, path, "")
				if err == nil {
					auth.HostKeyCallback = ssh2.InsecureIgnoreHostKey()
					if conf.DebugMode {
						log.Printf("[DEBUG] Using SSH Key: %s", path)
					}
					return auth
				} else if conf.DebugMode {
					log.Printf("[DEBUG] Failed to load key %s (maybe encrypted?): %v", path, err)
				}
			}
		}
	}

	// 2. Try SSH Agent
	if auth, err := ssh.NewSSHAgentAuth(user); err == nil {
		auth.HostKeyCallback = ssh2.InsecureIgnoreHostKey()
		if conf.DebugMode {
			log.Printf("[DEBUG] Using SSH Agent Auth")
		}
		return auth
	}

	if conf.DebugMode {
		log.Printf("[DEBUG] No SSH auth found")
	}
	return nil
}

func (s *GitService) openRepo(path string) (*git.Repository, error) {
	return git.PlainOpen(path)
}

func (s *GitService) IsGitRepo(path string) bool {
	_, err := git.PlainOpen(path)
	return err == nil
}

func (s *GitService) Fetch(path, remote string) error {
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

	err = r.Fetch(&git.FetchOptions{
		RemoteName: remote,
		Auth:       auth,
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (s *GitService) FetchWithAuth(path, remoteURL, authType, authKey, authSecret string, extraArgs ...string) error {
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

type channelWriter struct {
	ch chan string
}

func (w *channelWriter) Write(p []byte) (n int, err error) {
	if w.ch != nil {
		w.ch <- string(p)
	}
	return len(p), nil
}

func (s *GitService) GetCommitHash(path, remote, branch string) (string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return "", err
	}
	// Try to resolve as remote reference
	refName := plumbing.ReferenceName(fmt.Sprintf("refs/remotes/%s/%s", remote, branch))
	ref, err := r.Reference(refName, true)
	if err != nil {
		// Try local branch
		refName = plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch))
		ref, err = r.Reference(refName, true)
		if err != nil {
			return "", err
		}
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

func (s *GitService) Push(path, targetRemote, sourceHash, targetBranch string, options []string) error {
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

	err = r.Push(&git.PushOptions{
		RemoteName: targetRemote,
		RefSpecs:   []config.RefSpec{refSpec},
		Auth:       auth,
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

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
	return names, nil
}

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
		r := &domain.GitRemote{
			Name:       remote.Name,
			FetchURL:   "",
			PushURL:    "",
			FetchSpecs: []string{},
			PushSpecs:  []string{},
			IsMirror:   remote.Mirror,
		}
		if len(remote.URLs) > 0 {
			r.FetchURL = remote.URLs[0]
			// Default PushURL
			r.PushURL = remote.URLs[0]
		}
		for _, u := range remote.URLs {
			r.FetchSpecs = append(r.FetchSpecs, u) // This is wrong, FetchSpecs are refspecs.
		}
		for _, spec := range remote.Fetch {
			r.FetchSpecs = append(r.FetchSpecs, spec.String())
		}
		repoConfig.Remotes = append(repoConfig.Remotes, *r)
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

	return repoConfig, nil
}

func (s *GitService) AddRemote(path, name, url string, isMirror bool) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}
	_, err = r.CreateRemote(&config.RemoteConfig{
		Name:   name,
		URLs:   []string{url},
		Mirror: isMirror,
	})
	return err
}

func (s *GitService) RemoveRemote(path, name string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}
	return r.DeleteRemote(name)
}

func (s *GitService) SetRemotePushURL(path, name, url string) error {
	// go-git doesn't have a direct SetURL method on Remote, need to edit Config.
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}
	cfg, err := r.Config()
	if err != nil {
		return err
	}
	if remote, ok := cfg.Remotes[name]; ok {
		// remote.URLs is a slice.
		remote.URLs = []string{url} // This sets fetch URL.
		// If we want to set push URL, go-git Config struct doesn't expose it well in simple map?
		// Actually cfg.Remotes is map[string]*RemoteConfig.
		// RemoteConfig has URLs.
		return r.Storer.SetConfig(cfg)
	}
	return fmt.Errorf("remote %s not found", name)
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

func (s *GitService) PushWithAuth(path, targetRemoteURL, sourceHash, targetBranch, authType, authKey, authSecret string, options []string) error {
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

	err = remote.Push(&git.PushOptions{
		Auth:     auth,
		RefSpecs: []config.RefSpec{refSpec},
	})
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

func (s *GitService) TestRemoteConnection(url string) error {
	// Create a temporary remote
	// memory.NewStorage() would be better but nil is accepted for non-persistent remote
	remote := git.NewRemote(nil, &config.RemoteConfig{
		Name: "anonymous",
		URLs: []string{url},
	})

	auth := s.detectSSHAuth(url)

	_, err := remote.List(&git.ListOptions{
		Auth: auth,
	})
	return err
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

func (s *GitService) GetGitUser(path string) (string, string, error) {
	// 1. Try Local Config
	var name, email string
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

	// 2. Try Global Config (~/.gitconfig)
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

	return name, email, nil
}

func (s *GitService) SetGlobalGitUser(name, email string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	globalConfigPath := filepath.Join(home, ".gitconfig")

	// Read existing
	cfg := config.NewConfig()
	content, err := os.ReadFile(globalConfigPath)
	if err == nil {
		if err := cfg.Unmarshal(content); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	cfg.User.Name = name
	cfg.User.Email = email

	// Write back
	data, err := cfg.Marshal()
	if err != nil {
		return err
	}
	return os.WriteFile(globalConfigPath, data, 0644)
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

func (s *GitService) GetGlobalGitUser() (string, string, error) {
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
	return cfg.User.Name, cfg.User.Email, nil
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

// Task Manager for Async Clones
type TaskManager struct {
	tasks sync.Map
}

type Task struct {
	ID       string   `json:"id"`
	Status   string   `json:"status"` // running, success, failed
	Progress []string `json:"progress"`
	Error    string   `json:"error"`
}

var GlobalTaskManager = &TaskManager{}

func (tm *TaskManager) AddTask(id string) *Task {
	t := &Task{ID: id, Status: "running", Progress: []string{}}
	tm.tasks.Store(id, t)
	return t
}

func (tm *TaskManager) GetTask(id string) (*Task, bool) {
	v, ok := tm.tasks.Load(id)
	if !ok {
		return nil, false
	}
	return v.(*Task), true
}

func (tm *TaskManager) AppendLog(id string, log string) {
	if v, ok := tm.tasks.Load(id); ok {
		t := v.(*Task)
		t.Progress = append(t.Progress, log)
	}
}

func (tm *TaskManager) UpdateStatus(id string, status string, errStr string) {
	if v, ok := tm.tasks.Load(id); ok {
		t := v.(*Task)
		t.Status = status
		t.Error = errStr
	}
}

func (s *GitService) CreateTag(path, tagName, ref, message, authorName, authorEmail string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	hash, err := r.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
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
	return err
}

func (s *GitService) PushTag(path, remoteName, tagName, authType, authKey, authSecret string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}

	// Detect Auth
	auth, err := s.getAuth(authType, authKey, authSecret)
	if err != nil {
		return err
	}

	if auth == nil {
		// Try to detect from remote URL
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
		return nil
	}
	return err
}

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
	return tags, nil
}

type TagInfo struct {
	Name    string    `json:"name"`
	Hash    string    `json:"hash"`
	Message string    `json:"message"`
	Tagger  string    `json:"tagger"`
	Date    time.Time `json:"date"`
}

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
	return tags, nil
}

func (s *GitService) GetDescribe(path string) (string, error) {
	// git describe --tags --always --long
	// Note: go-git does not implement describe yet.
	// We use shell command for robustness and simplicity here.
	return s.RunCommand(path, "describe", "--tags", "--always", "--long")
}

func (s *GitService) GetLatestVersion(path string) (string, error) {
	// Simple strategy: use git describe --tags --abbrev=0 to get the latest tag reachable from HEAD
	// This respects semver ordering if git is configured or tags are standard.
	// However, git describe only sorts by topological reachability (most recent on branch).
	// To get strictly highest semver, we need to list all tags and sort.
	// For now, let's trust git describe --tags --abbrev=0 which is "latest tag on this branch".
	out, err := s.RunCommand(path, "describe", "--tags", "--abbrev=0")
	if err != nil {
		// If no tags found, might fail. Return empty or error.
		return "", err
	}
	return out, nil
}

type NextVersionInfo struct {
	Current   string `json:"current"`
	NextMajor string `json:"next_major"`
	NextMinor string `json:"next_minor"`
	NextPatch string `json:"next_patch"`
}

func (s *GitService) GetNextVersions(path string) (*NextVersionInfo, error) {
	// 1. Get latest version
	latest, err := s.GetLatestVersion(path)
	// If error or empty, default to v0.0.0
	if err != nil || latest == "" {
		latest = "v0.0.0"
	}

	// 2. Parse SemVer
	// Handle v prefix
	version := latest
	hasV := false
	if strings.HasPrefix(version, "v") {
		hasV = true
		version = version[1:]
	}

	parts := strings.Split(version, ".")
	major, minor, patch := 0, 0, 0

	// Parse robustly
	if len(parts) >= 1 {
		fmt.Sscanf(parts[0], "%d", &major)
	}
	if len(parts) >= 2 {
		fmt.Sscanf(parts[1], "%d", &minor)
	}
	if len(parts) >= 3 {
		fmt.Sscanf(parts[2], "%d", &patch)
	}

	// 3. Calculate next versions
	// Major: +1.0.0
	nextMajor := fmt.Sprintf("%d.0.0", major+1)
	// Minor: +0.1.0
	nextMinor := fmt.Sprintf("%d.%d.0", major, minor+1)
	// Patch: +0.0.1
	nextPatch := fmt.Sprintf("%d.%d.%d", major, minor, patch+1)

	if hasV {
		nextMajor = "v" + nextMajor
		nextMinor = "v" + nextMinor
		nextPatch = "v" + nextPatch
	}

	return &NextVersionInfo{
		Current:   latest,
		NextMajor: nextMajor,
		NextMinor: nextMinor,
		NextPatch: nextPatch,
	}, nil
}
