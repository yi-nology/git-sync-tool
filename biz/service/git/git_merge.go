package git

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/diff"
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
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	cBase, cTarget, err := s.resolveCommitPair(r, base, target)
	if err != nil {
		return nil, err
	}

	patch, err := cBase.Patch(cTarget)
	if err != nil {
		return nil, err
	}

	stats := patch.Stats()
	ds := &DiffStat{}
	for _, fs := range stats {
		ds.FilesChanged++
		ds.Insertions += fs.Addition
		ds.Deletions += fs.Deletion
	}
	return ds, nil
}

// GetDiffFiles returns list of changed files with status
func (s *GitService) GetDiffFiles(path, base, target string) ([]FileDiffStatus, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	cBase, cTarget, err := s.resolveCommitPair(r, base, target)
	if err != nil {
		return nil, err
	}

	patch, err := cBase.Patch(cTarget)
	if err != nil {
		return nil, err
	}

	var files []FileDiffStatus
	for _, fp := range patch.FilePatches() {
		from, to := fp.Files()
		status := "M"
		p := ""
		if from == nil && to != nil {
			status = "A"
			p = to.Path()
		} else if from != nil && to == nil {
			status = "D"
			p = from.Path()
		} else if from != nil && to != nil {
			if from.Path() != to.Path() {
				status = "R" // Treat as rename if paths differ
				p = to.Path()
			} else {
				status = "M"
				p = to.Path()
			}
		}

		if p != "" {
			files = append(files, FileDiffStatus{
				Path:   p,
				Status: status,
			})
		}
	}
	return files, nil
}

// simplePatch wrapper for filtering
type simplePatch struct {
	filePatches []diff.FilePatch
}

func (p *simplePatch) FilePatches() []diff.FilePatch {
	return p.filePatches
}

func (p *simplePatch) Message() string {
	return ""
}

// GetRawDiff returns the full diff content for a specific file or all
func (s *GitService) GetRawDiff(path, base, target, file string) (string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return "", err
	}

	cBase, cTarget, err := s.resolveCommitPair(r, base, target)
	if err != nil {
		return "", err
	}

	patch, err := cBase.Patch(cTarget)
	if err != nil {
		return "", err
	}

	if file == "" {
		return patch.String(), nil
	}

	var filtered []diff.FilePatch
	for _, fp := range patch.FilePatches() {
		from, to := fp.Files()
		if (from != nil && from.Path() == file) || (to != nil && to.Path() == file) {
			filtered = append(filtered, fp)
		}
	}

	if len(filtered) == 0 {
		return "", nil
	}

	var sb strings.Builder
	ue := diff.NewUnifiedEncoder(&sb, diff.DefaultContextLines)
	if err := ue.Encode(&simplePatch{filePatches: filtered}); err != nil {
		return "", err
	}
	return sb.String(), nil
}

type MergeResult struct {
	Success   bool     `json:"success"`
	Conflicts []string `json:"conflicts"` // List of conflicting files
	Output    string   `json:"output"`
	MergeID   string   `json:"merge_id"` // Transaction ID if needed
}

// MergeDryRun checks for conflicts without committing
func (s *GitService) MergeDryRun(path, source, target string) (*MergeResult, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	cSource, cTarget, err := s.resolveCommitPair(r, source, target)
	if err != nil {
		return nil, err
	}

	bases, err := cSource.MergeBase(cTarget)
	if err != nil || len(bases) == 0 {
		return nil, fmt.Errorf("no merge base found")
	}
	cBase := bases[0]

	patchSource, err := cBase.Patch(cSource)
	if err != nil {
		return nil, err
	}

	patchTarget, err := cBase.Patch(cTarget)
	if err != nil {
		return nil, err
	}

	// Check overlap
	changedSource := make(map[string]bool)
	for _, fp := range patchSource.FilePatches() {
		from, to := fp.Files()
		if from != nil {
			changedSource[from.Path()] = true
		}
		if to != nil {
			changedSource[to.Path()] = true
		}
	}

	var conflicts []string
	for _, fp := range patchTarget.FilePatches() {
		from, to := fp.Files()
		p := ""
		if from != nil {
			p = from.Path()
		} else if to != nil {
			p = to.Path()
		}

		if p != "" && changedSource[p] {
			// Potential conflict
			conflicts = append(conflicts, p)
		}
	}

	return &MergeResult{
		Success:   len(conflicts) == 0,
		Conflicts: conflicts,
	}, nil
}

// Merge performs the actual merge with optional no-ff and squash strategies
func (s *GitService) Merge(path, source, target, message string, noFF, squash bool) error {
	// 1. Checkout target
	if err := s.CheckoutBranch(path, target); err != nil {
		return fmt.Errorf("checkout target failed: %v", err)
	}

	// 2. Merge source
	// We use RunCommand because go-git does not support full merge logic yet
	args := []string{"merge", source}
	if noFF {
		args = append(args, "--no-ff")
	}
	if squash {
		args = append(args, "--squash")
	}
	if message != "" {
		args = append(args, "-m", message)
	}

	out, err := s.RunCommand(path, args...)
	if err != nil {
		// If failed, it's likely a conflict (since we did checkout).
		// We should abort to restore state
		s.RunCommand(path, "merge", "--abort")
		return fmt.Errorf("merge failed (aborted): %v. Output: %s", err, out)
	}

	// If squash, we need to commit separately
	if squash {
		commitArgs := []string{"commit"}
		if message != "" {
			commitArgs = append(commitArgs, "-m", message)
		} else {
			commitArgs = append(commitArgs, "-m", fmt.Sprintf("Squash merge %s into %s", source, target))
		}
		if out, err := s.RunCommand(path, commitArgs...); err != nil {
			return fmt.Errorf("squash commit failed: %v. Output: %s", err, out)
		}
	}

	return nil
}

// GetPatch generates a patch file content
func (s *GitService) GetPatch(path, base, target string) (string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return "", err
	}

	cBase, cTarget, err := s.resolveCommitPair(r, base, target)
	if err != nil {
		return "", err
	}

	patch, err := cBase.Patch(cTarget)
	if err != nil {
		return "", err
	}

	return patch.String(), nil
}
