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

// getAuthFromInfo 从AuthInfo结构获取认证方法，支持本地密钥和数据库密钥
func (s *GitService) getAuthFromInfo(authInfo domain.AuthInfo) (transport.AuthMethod, error) {
	if authInfo.Type == "ssh" {
		if authInfo.Source == "database" && authInfo.SSHKeyID > 0 {
			// 从数据库加载密钥 - 需要在调用方提供私钥内容
			return nil, fmt.Errorf("database key loading should be handled by caller with GetAuthFromDBKey")
		}
		// Source == "local" 或为空，使用文件路径
		if authInfo.Key != "" {
			publicKeys, err := ssh.NewPublicKeysFromFile("git", authInfo.Key, authInfo.Secret)
			if err != nil {
				return nil, err
			}
			publicKeys.HostKeyCallback = ssh2.InsecureIgnoreHostKey()
			return publicKeys, nil
		}
	} else if authInfo.Type == "http" && authInfo.Key != "" {
		return &http.BasicAuth{
			Username: authInfo.Key,
			Password: authInfo.Secret,
		}, nil
	}
	return nil, nil
}

// GetAuthFromDBKey 从数据库密钥内容创建认证方法
func (s *GitService) GetAuthFromDBKey(privateKey, passphrase string) (transport.AuthMethod, error) {
	publicKeys, err := ssh.NewPublicKeys("git", []byte(privateKey), passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}
	publicKeys.HostKeyCallback = ssh2.InsecureIgnoreHostKey()

	// 配置更广泛的 SSH 算法支持以提高兼容性
	publicKeys.HostKeyCallbackHelper = ssh.HostKeyCallbackHelper{
		HostKeyCallback: ssh2.InsecureIgnoreHostKey(),
	}

	return publicKeys, nil
}

// TestRemoteConnectionWithDBKey 使用数据库密钥测试远程连接
func (s *GitService) TestRemoteConnectionWithDBKey(url, privateKey, passphrase string) error {
	// 方案1: 先尝试使用原生 git 命令（更可靠）
	err := s.testConnectionWithGitCommand(url, privateKey, passphrase)
	if err == nil {
		return nil
	}

	// 方案2: 回退到 go-git（作为备选）
	auth, err := s.GetAuthFromDBKey(privateKey, passphrase)
	if err != nil {
		return err
	}

	ep, err := transport.NewEndpoint(url)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	r, err := git.Init(nil, nil)
	if err != nil {
		return fmt.Errorf("failed to init memory repo: %v", err)
	}

	remote, err := r.CreateRemote(&config.RemoteConfig{
		Name: "test",
		URLs: []string{ep.String()},
	})
	if err != nil {
		return fmt.Errorf("failed to create remote: %v", err)
	}

	_, err = remote.List(&git.ListOptions{
		Auth: auth,
	})
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}

	return nil
}

// testConnectionWithGitCommand 使用原生 git 命令测试连接（更可靠）
func (s *GitService) testConnectionWithGitCommand(url, privateKey, passphrase string) error {
	// 创建临时私钥文件
	tmpFile, err := os.CreateTemp("", "git_ssh_key_*")
	if err != nil {
		return fmt.Errorf("failed to create temp key file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// 写入私钥内容
	if _, err := tmpFile.WriteString(privateKey); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write key file: %v", err)
	}
	tmpFile.Close()

	// 设置文件权限为 600
	if err := os.Chmod(tmpFile.Name(), 0600); err != nil {
		return fmt.Errorf("failed to set key file permissions: %v", err)
	}

	// 构建 GIT_SSH_COMMAND
	sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null", tmpFile.Name())

	// 执行 git ls-remote
	cmd := exec.Command("git", "ls-remote", "--heads", url)
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git ls-remote failed: %v, output: %s", err, string(output))
	}

	return nil
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

	err = r.Fetch(&git.FetchOptions{
		RemoteName: remote,
		Auth:       auth,
		Progress:   progress,
		RefSpecs: []config.RefSpec{
			config.RefSpec("+refs/heads/*:refs/remotes/" + remote + "/*"),
			config.RefSpec("+refs/tags/*:refs/tags/*"),
		},
	})
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

// FetchWithDBKey 使用数据库SSH密钥进行fetch（使用原生git命令，更可靠）
func (s *GitService) FetchWithDBKey(path, remoteURL, privateKey, passphrase string, progress io.Writer, refSpecs ...string) error {
	// 创建临时私钥文件
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

	args := []string{"fetch", remoteURL}
	args = append(args, refSpecs...)

	cmd := exec.Command("git", args...)
	cmd.Dir = path
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	if progress != nil {
		cmd.Stdout = progress
		cmd.Stderr = progress
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git fetch failed: %v", err)
		}
	} else {
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git fetch failed: %v, output: %s", err, string(output))
		}
	}

	return nil
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

// CloneWithDBKey 使用数据库SSH密钥进行克隆（使用原生git命令，更可靠）
func (s *GitService) CloneWithDBKey(remoteURL, localPath, privateKey, passphrase string, progressChan chan string) error {
	// 创建临时私钥文件
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

	cmd := exec.Command("git", "clone", remoteURL, localPath)
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	if progressChan != nil {
		cmd.Stdout = &channelWriter{ch: progressChan}
		cmd.Stderr = &channelWriter{ch: progressChan}
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %v", err)
	}

	return nil
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
	pushOpts.Auth = auth
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

// PushWithDBKey 使用数据库SSH密钥进行push（使用原生git命令，更可靠）
func (s *GitService) PushWithDBKey(path, targetRemoteURL, sourceHash, targetBranch, privateKey, passphrase string, options []string, progress io.Writer) error {
	// 创建临时私钥文件
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

	refSpec := fmt.Sprintf("%s:refs/heads/%s", sourceHash, targetBranch)
	args := []string{"push", targetRemoteURL, refSpec}
	args = append(args, options...)

	cmd := exec.Command("git", args...)
	cmd.Dir = path
	cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)

	if progress != nil {
		cmd.Stdout = progress
		cmd.Stderr = progress
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git push failed: %v", err)
		}
	} else {
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git push failed: %v, output: %s", err, string(output))
		}
	}

	return nil
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
