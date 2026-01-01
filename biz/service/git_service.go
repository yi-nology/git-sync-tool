package service

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/yi-nology/git-manage-service/biz/config"
	"github.com/yi-nology/git-manage-service/biz/model"
)

type GitService struct{}

func NewGitService() *GitService {
	return &GitService{}
}

func (s *GitService) RunCommand(dir string, args ...string) (string, error) {
	if config.DebugMode {
		log.Printf("[DEBUG] Executing in %s: git %s", dir, strings.Join(args, " "))
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	// Prevent password prompts
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("git command failed: %s, output: %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func (s *GitService) IsGitRepo(path string) bool {
	_, err := s.RunCommand(path, "rev-parse", "--is-inside-work-tree")
	return err == nil
}

func (s *GitService) Fetch(path, remote string) error {
	_, err := s.RunCommand(path, "fetch", remote)
	return err
}

func (s *GitService) FetchWithAuth(path, remoteURL, authType, authKey, authSecret string, extraArgs ...string) error {
	finalURL := remoteURL
	var env []string

	if authType == "http" && authKey != "" {
		u, err := url.Parse(remoteURL)
		if err == nil {
			u.User = url.UserPassword(authKey, authSecret)
			finalURL = u.String()
		}
	} else if authType == "ssh" && authKey != "" {
		// authKey is path to private key
		sshCmd := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=no", authKey)
		env = append(env, "GIT_SSH_COMMAND="+sshCmd)
	}

	args := []string{"fetch", finalURL}
	if len(extraArgs) > 0 {
		args = append(args, extraArgs...)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = path
	cmd.Env = append(os.Environ(), env...)
	cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=0")

	if config.DebugMode {
		log.Printf("[DEBUG] FetchWithAuth in %s: URL=%s Auth=%s", path, finalURL, authType)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fetch failed: %v, output: %s", err, string(out))
	}
	return nil
}

func (s *GitService) Clone(remoteURL, localPath, authType, authKey, authSecret string) error {
	return s.CloneWithProgress(remoteURL, localPath, authType, authKey, authSecret, nil)
}

// CloneWithProgress executes clone and writes output to the provided channel or buffer
// If progressChan is nil, it behaves like normal Clone
func (s *GitService) CloneWithProgress(remoteURL, localPath, authType, authKey, authSecret string, progressChan chan string) error {
	finalURL := remoteURL
	var env []string

	if authType == "http" && authKey != "" {
		u, err := url.Parse(remoteURL)
		if err == nil {
			u.User = url.UserPassword(authKey, authSecret)
			finalURL = u.String()
		}
	} else if authType == "ssh" && authKey != "" {
		sshCmd := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=no", authKey)
		env = append(env, "GIT_SSH_COMMAND="+sshCmd)
	}

	// git clone <url> <path> --progress
	args := []string{"clone", "--progress", finalURL, localPath}
	cmd := exec.Command("git", args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=0")

	if config.DebugMode {
		log.Printf("[DEBUG] Clone: URL=%s Path=%s Auth=%s", finalURL, localPath, authType)
	}

	if progressChan == nil {
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("clone failed: %v, output: %s", err, string(out))
		}
		return nil
	}

	// Capture stderr (where git progress is written)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	// Read output line by line
	scanner := bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text := scanner.Text()
		progressChan <- text
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("clone failed: %v", err)
	}
	return nil
}

func (s *GitService) GetCommitHash(path, remote, branch string) (string, error) {
	// ref: refs/remotes/<remote>/<branch>
	ref := fmt.Sprintf("refs/remotes/%s/%s", remote, branch)
	return s.RunCommand(path, "rev-parse", ref)
}

// IsAncestor checks if ancestor is an ancestor of descendant (fast-forward possible)
func (s *GitService) IsAncestor(path, ancestor, descendant string) (bool, error) {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", ancestor, descendant)
	cmd.Dir = path
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func (s *GitService) Push(path, targetRemote, sourceHash, targetBranch string, options []string) error {
	// git push [options] <remote> <source_hash>:refs/heads/<target_branch>
	args := []string{"push"}
	if len(options) > 0 {
		args = append(args, options...)
	}
	refSpec := fmt.Sprintf("%s:refs/heads/%s", sourceHash, targetBranch)
	args = append(args, targetRemote, refSpec)
	_, err := s.RunCommand(path, args...)
	return err
}

func (s *GitService) GetRemotes(path string) ([]string, error) {
	out, err := s.RunCommand(path, "remote")
	if err != nil {
		return nil, err
	}
	return strings.Split(out, "\n"), nil
}

func (s *GitService) GetRemoteURL(path, remoteName string) (string, error) {
	return s.RunCommand(path, "config", "--local", "--get", fmt.Sprintf("remote.%s.url", remoteName))
}

// GetRepoConfig parses .git/config to get remotes and branches
func (s *GitService) GetRepoConfig(path string) (*model.GitRepoConfig, error) {
	// Use git config --local --list to get all config
	out, err := s.RunCommand(path, "config", "--local", "--list")
	if err != nil {
		return nil, err
	}

	config := &model.GitRepoConfig{
		Remotes:  []model.GitRemote{},
		Branches: []model.GitBranch{},
	}

	lines := strings.Split(out, "\n")
	remotesMap := make(map[string]*model.GitRemote)
	branchesMap := make(map[string]*model.GitBranch)

	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		if strings.HasPrefix(key, "remote.") {
			// remote.<name>.<key>
			subParts := strings.Split(key, ".")
			if len(subParts) < 3 {
				continue
			}
			name := subParts[1]
			prop := subParts[2]

			if _, ok := remotesMap[name]; !ok {
				remotesMap[name] = &model.GitRemote{Name: name}
			}
			remote := remotesMap[name]

			switch prop {
			case "url":
				remote.FetchURL = value
				// Default PushURL to FetchURL if not set (will be handled later or UI can handle it)
			case "pushurl":
				remote.PushURL = value
			case "fetch":
				remote.FetchSpecs = append(remote.FetchSpecs, value)
			case "push":
				remote.PushSpecs = append(remote.PushSpecs, value)
			case "mirror":
				remote.IsMirror = (value == "true")
			}
		} else if strings.HasPrefix(key, "branch.") {
			// branch.<name>.<key>
			subParts := strings.Split(key, ".")
			if len(subParts) < 3 {
				continue
			}
			name := subParts[1]
			prop := subParts[2]

			if _, ok := branchesMap[name]; !ok {
				branchesMap[name] = &model.GitBranch{Name: name}
			}
			branch := branchesMap[name]

			switch prop {
			case "remote":
				branch.Remote = value
			case "merge":
				branch.Merge = value
			}
		}
	}

	for _, r := range remotesMap {
		if r.PushURL == "" {
			r.PushURL = r.FetchURL
		}
		config.Remotes = append(config.Remotes, *r)
	}
	for _, b := range branchesMap {
		// Construct UpstreamRef e.g. origin/main
		if b.Remote != "" && b.Merge != "" {
			// b.Merge is usually refs/heads/main
			shortRef := strings.TrimPrefix(b.Merge, "refs/heads/")
			b.UpstreamRef = fmt.Sprintf("%s/%s", b.Remote, shortRef)
		}
		config.Branches = append(config.Branches, *b)
	}

	return config, nil
}

func (s *GitService) AddRemote(path, name, url string, isMirror bool) error {
	args := []string{"remote", "add"}
	if isMirror {
		args = append(args, "--mirror=fetch")
	}
	args = append(args, name, url)
	_, err := s.RunCommand(path, args...)
	return err
}

func (s *GitService) RemoveRemote(path, name string) error {
	_, err := s.RunCommand(path, "remote", "remove", name)
	return err
}

func (s *GitService) SetRemotePushURL(path, name, url string) error {
	_, err := s.RunCommand(path, "remote", "set-url", "--push", name, url)
	return err
}

// GetBranches returns all local and remote branches
func (s *GitService) GetBranches(path string) ([]string, error) {
	// git branch -a --format="%(refname:short)"
	out, err := s.RunCommand(path, "branch", "-a", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(out, "\n")
	var branches []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, "HEAD") {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

// GetCommits returns commits for a specific branch
func (s *GitService) GetCommits(path, branch, since, until string) (string, error) {
	// git log --pretty=format:"%H|%an|%ae|%ad|%s" --date=iso
	args := []string{"log", "--pretty=format:%H|%an|%ae|%ad|%s", "--date=iso", branch}
	if since != "" {
		args = append(args, "--since="+since)
	}
	if until != "" {
		args = append(args, "--until="+until)
	}
	return s.RunCommand(path, args...)
}

func (s *GitService) PushWithAuth(path, targetRemoteURL, sourceHash, targetBranch, authType, authKey, authSecret string, options []string) error {
	finalURL := targetRemoteURL
	var env []string

	if authType == "http" && authKey != "" {
		u, err := url.Parse(targetRemoteURL)
		if err == nil {
			u.User = url.UserPassword(authKey, authSecret)
			finalURL = u.String()
		}
	} else if authType == "ssh" && authKey != "" {
		sshCmd := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes -o StrictHostKeyChecking=no", authKey)
		env = append(env, "GIT_SSH_COMMAND="+sshCmd)
	}

	args := []string{"push"}
	if len(options) > 0 {
		args = append(args, options...)
	}
	refSpec := fmt.Sprintf("%s:refs/heads/%s", sourceHash, targetBranch)
	args = append(args, finalURL, refSpec)

	cmd := exec.Command("git", args...)
	cmd.Dir = path
	cmd.Env = append(os.Environ(), env...)
	cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=0")

	if config.DebugMode {
		log.Printf("[DEBUG] PushWithAuth in %s: URL=%s Auth=%s", path, finalURL, authType)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("push failed: %v, output: %s", err, string(out))
	}
	return nil
}

// GetRepoFiles returns all files in the current HEAD of the branch
func (s *GitService) GetRepoFiles(path, branch string) ([]string, error) {
	// git ls-tree -r --name-only <branch>
	out, err := s.RunCommand(path, "ls-tree", "-r", "--name-only", branch)
	if err != nil {
		return nil, err
	}
	return strings.Split(out, "\n"), nil
}

// BlameFile returns blame information for a file
func (s *GitService) BlameFile(path, branch, file string) (string, error) {
	// git blame --line-porcelain -w <branch> -- <file>
	// -w ignores whitespace
	return s.RunCommand(path, "blame", "--line-porcelain", "-w", branch, "--", file)
}

// TestRemoteConnection checks if the remote is accessible
func (s *GitService) TestRemoteConnection(url string) error {
	// git ls-remote <url>
	cmd := exec.Command("git", "ls-remote", url)
	// We might need to set timeouts or handle auth prompts (which will fail in non-interactive mode)
	// If it prompts for password, it will hang or fail.
	// Setting GIT_TERMINAL_PROMPT=0 prevents hanging on password prompt
	cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=0")

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("connection failed: %v, output: %s", err, string(out))
	}
	return nil
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
