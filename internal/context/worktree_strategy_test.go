package context_test

import (
	gocontext "context"
	"os"
	"path/filepath"
	"testing"

	wizctx "github.com/buck3000/wiz/internal/context"
	"github.com/buck3000/wiz/internal/gitx"
	"github.com/buck3000/wiz/testutil"
)

func TestWorktreeProvisionerCreateAndDestroy(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, _ := gitx.Discover(tr.Dir)
	ctx := gocontext.Background()

	prov := wizctx.NewProvisioner(wizctx.StrategyWorktree, repo)
	if prov.Strategy() != wizctx.StrategyWorktree {
		t.Fatalf("Strategy = %q", prov.Strategy())
	}

	path, err := prov.Create(ctx, wizctx.CreateOpts{
		Name:   "feat-a",
		Branch: "feat-a",
		Repo:   repo,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify the worktree directory exists and has files.
	if _, err := os.Stat(filepath.Join(path, "README.md")); err != nil {
		t.Fatal("README.md not found in worktree")
	}

	// Verify git reports the correct branch.
	wtRepo, err := gitx.Discover(path)
	if err != nil {
		t.Fatal(err)
	}
	branch, err := wtRepo.CurrentBranch(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if branch != "feat-a" {
		t.Errorf("branch = %q, want feat-a", branch)
	}

	// Destroy it.
	if err := prov.Destroy(ctx, path, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("worktree dir should not exist after destroy")
	}
}

func TestWorktreeProvisionerExistingBranch(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	tr.CreateBranch("existing-branch")
	repo, _ := gitx.Discover(tr.Dir)
	ctx := gocontext.Background()

	prov := wizctx.NewProvisioner(wizctx.StrategyWorktree, repo)
	path, err := prov.Create(ctx, wizctx.CreateOpts{
		Name:   "existing-branch",
		Branch: "existing-branch",
		Repo:   repo,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer prov.Destroy(ctx, path, true)

	wtRepo, _ := gitx.Discover(path)
	branch, _ := wtRepo.CurrentBranch(ctx)
	if branch != "existing-branch" {
		t.Errorf("branch = %q", branch)
	}
}

func TestWorktreeProvisionerWithBase(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	tr.AddFile("base.txt", "base content")
	tr.Commit("add base")
	tr.CreateBranch("base-branch")

	repo, _ := gitx.Discover(tr.Dir)
	ctx := gocontext.Background()

	prov := wizctx.NewProvisioner(wizctx.StrategyWorktree, repo)
	path, err := prov.Create(ctx, wizctx.CreateOpts{
		Name:       "from-base",
		Branch:     "from-base",
		BaseBranch: "base-branch",
		Repo:       repo,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer prov.Destroy(ctx, path, true)

	// The worktree should contain the base file.
	if _, err := os.Stat(filepath.Join(path, "base.txt")); err != nil {
		t.Fatal("base.txt not found in worktree")
	}
}
