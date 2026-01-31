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

func TestCloneProvisionerCreateAndDestroy(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, _ := gitx.Discover(tr.Dir)
	ctx := gocontext.Background()

	prov := wizctx.NewProvisioner(wizctx.StrategyClone, repo)
	if prov.Strategy() != wizctx.StrategyClone {
		t.Fatalf("Strategy = %q", prov.Strategy())
	}

	path, err := prov.Create(ctx, wizctx.CreateOpts{
		Name:   "clone-feat",
		Branch: "clone-feat",
		Repo:   repo,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify the clone directory exists and has files.
	if _, err := os.Stat(filepath.Join(path, "README.md")); err != nil {
		t.Fatal("README.md not found in clone")
	}

	// Verify correct branch.
	cloneRepo, err := gitx.Discover(path)
	if err != nil {
		t.Fatal(err)
	}
	branch, err := cloneRepo.CurrentBranch(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if branch != "clone-feat" {
		t.Errorf("branch = %q, want clone-feat", branch)
	}

	// Destroy.
	if err := prov.Destroy(ctx, path, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("clone dir should not exist after destroy")
	}
}

func TestCloneProvisionerDirtyRefuseDelete(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, _ := gitx.Discover(tr.Dir)
	ctx := gocontext.Background()

	prov := wizctx.NewProvisioner(wizctx.StrategyClone, repo)
	path, err := prov.Create(ctx, wizctx.CreateOpts{
		Name:   "dirty-clone",
		Branch: "dirty-clone",
		Repo:   repo,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Make it dirty.
	os.WriteFile(filepath.Join(path, "untracked.txt"), []byte("dirty"), 0o644)

	// Should refuse without --force.
	err = prov.Destroy(ctx, path, false)
	if err == nil {
		t.Fatal("expected error destroying dirty clone without force")
	}

	// Should succeed with --force.
	err = prov.Destroy(ctx, path, true)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCloneProvisionerWithBase(t *testing.T) {
	tr := testutil.NewTestRepo(t)

	// Create a base branch with content.
	tr.CreateBranch("base-branch")
	tr.Checkout("base-branch")
	tr.AddFile("base.txt", "base content")
	tr.Commit("base commit")
	tr.Checkout("main")

	repo, _ := gitx.Discover(tr.Dir)
	ctx := gocontext.Background()

	// In a clone --shared, local branches appear as origin/<branch>.
	// The clone provisioner uses the base branch ref, so use origin/base-branch.
	prov := wizctx.NewProvisioner(wizctx.StrategyClone, repo)
	path, err := prov.Create(ctx, wizctx.CreateOpts{
		Name:       "from-base",
		Branch:     "from-base",
		BaseBranch: "origin/base-branch",
		Repo:       repo,
	})
	if err != nil {
		t.Fatalf("create with base: %v", err)
	}
	defer prov.Destroy(ctx, path, true)

	// Verify base file exists.
	data, err := os.ReadFile(filepath.Join(path, "base.txt"))
	if err != nil {
		t.Fatalf("read base.txt: %v", err)
	}
	if string(data) != "base content" {
		t.Errorf("base.txt = %q", data)
	}
}

func TestCloneProvisionerExistingBranch(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, _ := gitx.Discover(tr.Dir)
	ctx := gocontext.Background()

	// Create a branch in the source repo.
	tr.CreateBranch("existing-branch")
	tr.Checkout("main")

	prov := wizctx.NewProvisioner(wizctx.StrategyClone, repo)
	path, err := prov.Create(ctx, wizctx.CreateOpts{
		Name:   "use-existing",
		Branch: "existing-branch",
		Repo:   repo,
	})
	if err != nil {
		t.Fatalf("create with existing branch: %v", err)
	}
	defer prov.Destroy(ctx, path, true)

	cloneRepo, err := gitx.Discover(path)
	if err != nil {
		t.Fatal(err)
	}
	branch, _ := cloneRepo.CurrentBranch(ctx)
	if branch != "existing-branch" {
		t.Errorf("branch = %q, want existing-branch", branch)
	}
}

func TestAutoStrategyDefaultsToWorktree(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, _ := gitx.Discover(tr.Dir)

	prov := wizctx.NewProvisioner(wizctx.StrategyAuto, repo)
	if prov.Strategy() != wizctx.StrategyWorktree {
		t.Errorf("auto strategy resolved to %q, want worktree", prov.Strategy())
	}
}
