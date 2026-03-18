package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestBranchCRUD(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "git-test-branch")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	s := NewGitService()

	// Init Repo
	r, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatal(err)
	}

	// Create a commit so we have a HEAD
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	w, err := r.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Add("."); err != nil {
		t.Fatalf("w.Add failed: %v", err)
	}
	if _, err := w.Commit("initial", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	}); err != nil {
		t.Fatalf("w.Commit failed: %v", err)
	}

	// Determine default branch name (master or main)
	branches, err := s.ListBranchesWithInfo(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(branches) == 0 {
		t.Fatal("No branches found after init")
	}
	defaultBranch := branches[0].Name

	// 1. Test Create
	newBranch := "feature-test"
	if err := s.CreateBranch(tmpDir, newBranch, defaultBranch); err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}

	// Verify Created
	branches, _ = s.ListBranchesWithInfo(tmpDir)
	found := false
	for _, b := range branches {
		if b.Name == newBranch {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Branch %s not found after creation", newBranch)
	}

	// 2. Test Rename
	renamedBranch := "feature-renamed"
	if err := s.RenameBranch(tmpDir, newBranch, renamedBranch); err != nil {
		t.Fatalf("RenameBranch failed: %v", err)
	}

	branches, _ = s.ListBranchesWithInfo(tmpDir)
	foundOld := false
	foundNew := false
	for _, b := range branches {
		if b.Name == newBranch {
			foundOld = true
		}
		if b.Name == renamedBranch {
			foundNew = true
		}
	}
	if foundOld {
		t.Error("Old branch name still exists")
	}
	if !foundNew {
		t.Error("New branch name not found")
	}

	// 3. Test Description
	desc := "This is a test branch"
	if err := s.SetBranchDescription(tmpDir, renamedBranch, desc); err != nil {
		t.Fatalf("SetBranchDescription failed: %v", err)
	}
	gotDesc, err := s.GetBranchDescription(tmpDir, renamedBranch)
	if err != nil {
		t.Fatalf("GetBranchDescription failed: %v", err)
	}
	if strings.TrimSpace(gotDesc) != desc {
		t.Errorf("Description mismatch. Got '%s', want '%s'", gotDesc, desc)
	}

	// 4. Test Delete
	if err := s.DeleteBranch(tmpDir, renamedBranch, true); err != nil {
		t.Fatalf("DeleteBranch failed: %v", err)
	}

	branches, _ = s.ListBranchesWithInfo(tmpDir)
	for _, b := range branches {
		if b.Name == renamedBranch {
			t.Error("Branch still exists after delete")
		}
	}
}

func TestGetBranchMetrics(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "git-test-metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	s := NewGitService()

	// Init Repo
	r, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatal(err)
	}

	w, err := r.Worktree()
	if err != nil {
		t.Fatal(err)
	}

	// Create 3 commits
	for i := 0; i < 3; i++ {
		filename := filepath.Join(tmpDir, fmt.Sprintf("file%d.txt", i))
		if err := os.WriteFile(filename, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := w.Add(filepath.Base(filename)); err != nil {
			t.Fatalf("w.Add failed: %v", err)
		}
		if _, err := w.Commit(fmt.Sprintf("commit %d", i), &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test",
				Email: "test@example.com",
				When:  time.Now(),
			},
		}); err != nil {
			t.Fatalf("w.Commit failed: %v", err)
		}
	}

	// Test Metrics
	metrics, err := s.GetBranchMetrics(tmpDir, "master")
	if err != nil {
		t.Fatalf("GetBranchMetrics failed: %v", err)
	}

	if count, ok := metrics["commit_count"]; !ok || count != 3 {
		t.Errorf("Expected commit_count 3, got %v", metrics["commit_count"])
	}
}
