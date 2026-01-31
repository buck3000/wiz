package gitx_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/firewood-buck-3000/wiz/internal/gitx"
	"github.com/firewood-buck-3000/wiz/testutil"
)

func TestWorktreeAddAndList(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, _ := gitx.Discover(tr.Dir)
	ctx := context.Background()

	tmpDir, _ := filepath.EvalSymlinks(t.TempDir())
	wtPath := filepath.Join(tmpDir, "feature-branch")
	err := repo.WorktreeAdd(ctx, wtPath, "feature-branch", true, "")
	if err != nil {
		t.Fatal(err)
	}

	// Verify directory exists.
	if _, err := os.Stat(wtPath); err != nil {
		t.Fatalf("worktree dir does not exist: %v", err)
	}

	// List should show both the main worktree and the new one.
	wts, err := repo.WorktreeList(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(wts) < 2 {
		t.Fatalf("expected at least 2 worktrees, got %d", len(wts))
	}

	found := false
	for _, wt := range wts {
		if wt.Path == wtPath {
			found = true
			if wt.Branch != "refs/heads/feature-branch" {
				t.Errorf("Branch = %q, want refs/heads/feature-branch", wt.Branch)
			}
		}
	}
	if !found {
		t.Errorf("worktree at %q not found in list", wtPath)
	}
}

func TestWorktreeAddExistingBranch(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	tr.CreateBranch("existing")
	repo, _ := gitx.Discover(tr.Dir)
	ctx := context.Background()

	wtPath := filepath.Join(t.TempDir(), "existing")
	err := repo.WorktreeAdd(ctx, wtPath, "existing", false, "")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(wtPath, "README.md")); err != nil {
		t.Fatal("expected README.md in worktree")
	}
}

func TestWorktreeRemove(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, _ := gitx.Discover(tr.Dir)
	ctx := context.Background()

	wtPath := filepath.Join(t.TempDir(), "to-remove")
	if err := repo.WorktreeAdd(ctx, wtPath, "to-remove", true, ""); err != nil {
		t.Fatal(err)
	}

	if err := repo.WorktreeRemove(ctx, wtPath, false); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Fatal("worktree dir still exists after remove")
	}
}

func TestParseWorktreeListEmpty(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, _ := gitx.Discover(tr.Dir)
	ctx := context.Background()

	wts, err := repo.WorktreeList(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(wts) == 0 {
		t.Fatal("expected at least 1 worktree (the main one)")
	}
	if wts[0].Path != tr.Dir {
		t.Errorf("first worktree path = %q, want %q", wts[0].Path, tr.Dir)
	}
}
