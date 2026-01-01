package service

import (
	"fmt"
	"strings"
)

type DiffStat struct {
	FilesChanged int
	Insertions   int
	Deletions    int
}

type FileDiffStatus struct {
	Path   string `json:"path"`
	Status string `json:"status"` // A, M, D, R, C, U, etc.
}

// GetDiffStat returns the shortstat of diff between base and target
func (s *GitService) GetDiffStat(path, base, target string) (*DiffStat, error) {
	// git diff --shortstat base target
	out, err := s.RunCommand(path, "diff", "--shortstat", base, target)
	if err != nil {
		return nil, err
	}
	
	stat := &DiffStat{}
	if strings.TrimSpace(out) == "" {
		return stat, nil
	}
	
	// Example output: " 3 files changed, 15 insertions(+), 5 deletions(-)"
	// Note: Sometimes parts are missing if only insertions or deletions
	
	// Simple parsing
	fmt.Sscanf(out, "%d files changed, %d insertions(+), %d deletions(-)", &stat.FilesChanged, &stat.Insertions, &stat.Deletions)
	
	// fmt.Sscanf might fail if format varies (e.g. only insertions). Let's use regex or manual parse if needed.
	// But for quick impl, let's try manual split.
	parts := strings.Split(out, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		var val int
		if strings.Contains(p, "files changed") || strings.Contains(p, "file changed") {
			fmt.Sscanf(p, "%d", &val)
			stat.FilesChanged = val
		} else if strings.Contains(p, "insertions") || strings.Contains(p, "insertion") {
			fmt.Sscanf(p, "%d", &val)
			stat.Insertions = val
		} else if strings.Contains(p, "deletions") || strings.Contains(p, "deletion") {
			fmt.Sscanf(p, "%d", &val)
			stat.Deletions = val
		}
	}
	
	return stat, nil
}

// GetDiffFiles returns list of changed files with status
func (s *GitService) GetDiffFiles(path, base, target string) ([]FileDiffStatus, error) {
	// git diff --name-status base target
	out, err := s.RunCommand(path, "diff", "--name-status", base, target)
	if err != nil {
		return nil, err
	}
	
	var files []FileDiffStatus
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			files = append(files, FileDiffStatus{
				Status: parts[0],
				Path:   parts[1],
			})
		}
	}
	return files, nil
}

// GetRawDiff returns the full diff content for a specific file or all
func (s *GitService) GetRawDiff(path, base, target, file string) (string, error) {
	args := []string{"diff", base, target}
	if file != "" {
		args = append(args, "--", file)
	}
	return s.RunCommand(path, args...)
}

// MergeResult holds the result of a merge attempt
type MergeResult struct {
	Success   bool     `json:"success"`
	Conflicts []string `json:"conflicts"` // List of conflicting files
	Output    string   `json:"output"`
	MergeID   string   `json:"merge_id"` // Transaction ID if needed
}

// MergeDryRun checks for conflicts without committing
func (s *GitService) MergeDryRun(path, source, target string) (*MergeResult, error) {
	// We use git merge-tree if available (git >= 2.38 for best results with --write-tree, but older versions output text)
	// Or we can try a merge in memory.
	// Let's try `git merge-tree <base-hash> <source-hash> <target-hash>` which is the old low-level command.
	// It outputs the diff to stdout.
	// New way: `git merge-tree --write-tree --name-only <branch1> <branch2>` (shows conflicts)
	
	// First, get the common ancestor (merge base)
	base, err := s.RunCommand(path, "merge-base", target, source)
	if err != nil {
		return nil, fmt.Errorf("failed to find merge base: %v", err)
	}
	base = strings.TrimSpace(base)
	
	// Old school safe way: 
	// 1. Check if we are on target branch? 
	//    We shouldn't assume we can switch branches freely if the repo is in use.
	//    But for this tool, let's assume we own the repo dir.
	
	// For "System Requirements: Support Git standard workflow", we should probably use a temporary worktree 
	// to avoid messing up the main working directory if possible.
	// However, `git merge-tree` avoids touching the worktree.
	
	// Check if git merge-tree supports --write-tree (Git 2.38+)
	// If not, we fall back to logic.
	// Let's assume a modern git or just parse `git merge-tree <base> <target> <source>` output.
	// The output contains "changed in both" which indicates conflict.
	
	out, err := s.RunCommand(path, "merge-tree", base, target, source)
	if err != nil {
		return nil, err
	}
	
	result := &MergeResult{Success: true}
	
	// Parse output for conflicts
	// Format of merge-tree (old style):
	// changed in both
	//   base   100644 ... file
	//   our    100644 ... file
	//   their  100644 ... file
	
	if strings.Contains(out, "changed in both") {
		result.Success = false
		lines := strings.Split(out, "\n")
		var conflicts []string
		capture := false
		for _, line := range lines {
			if strings.TrimSpace(line) == "changed in both" {
				capture = true
				continue
			}
			if capture {
				if strings.TrimSpace(line) == "" {
					capture = false
					continue
				}
				// line example: "  base   100644 4e5... README.md"
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					// Add file path
					conflicts = append(conflicts, parts[3])
				}
			}
		}
		// Deduplicate
		seen := make(map[string]bool)
		for _, c := range conflicts {
			if !seen[c] {
				result.Conflicts = append(result.Conflicts, c)
				seen[c] = true
			}
		}
	}
	
	return result, nil
}

// Merge performs the actual merge
func (s *GitService) Merge(path, source, target, message string) error {
	// 1. Checkout target
	_, err := s.RunCommand(path, "checkout", target)
	if err != nil {
		return fmt.Errorf("checkout target failed: %v", err)
	}
	
	// 2. Merge source
	args := []string{"merge", source}
	if message != "" {
		args = append(args, "-m", message)
	}
	
	// We allow it to fail if conflict
	out, err := s.RunCommand(path, args...)
	if err != nil {
		// If failed, it's likely a conflict (since we did checkout).
		// We should abort to restore state
		s.RunCommand(path, "merge", "--abort")
		return fmt.Errorf("merge failed (aborted): %v. Output: %s", err, out)
	}
	
	return nil
}

// GetPatch generates a patch file content
func (s *GitService) GetPatch(path, base, target string) (string, error) {
	return s.RunCommand(path, "format-patch", "--stdout", base+"..."+target)
}
