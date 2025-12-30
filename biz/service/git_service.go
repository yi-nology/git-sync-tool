package service

import (
	"fmt"
	"github.com/yi-nology/git-sync-tool/biz/config"
	"log"
	"os/exec"
	"strings"
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
