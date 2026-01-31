package gitx_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/firewood-buck-3000/wiz/internal/gitx"
	"github.com/firewood-buck-3000/wiz/testutil"
)

func TestDiscoverFromRoot(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, err := gitx.Discover(tr.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if repo.WorkDir != tr.Dir {
		t.Errorf("WorkDir = %q, want %q", repo.WorkDir, tr.Dir)
	}
	if repo.GitDir != filepath.Join(tr.Dir, ".git") {
		t.Errorf("GitDir = %q, want %q", repo.GitDir, filepath.Join(tr.Dir, ".git"))
	}
	if repo.CommonDir != filepath.Join(tr.Dir, ".git") {
		t.Errorf("CommonDir = %q, want %q", repo.CommonDir, filepath.Join(tr.Dir, ".git"))
	}
}

func TestDiscoverFromSubdir(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	tr.AddFile("sub/dir/file.txt", "hello")
	tr.Commit("add subdir")

	subdir := filepath.Join(tr.Dir, "sub", "dir")
	repo, err := gitx.Discover(subdir)
	if err != nil {
		t.Fatal(err)
	}
	if repo.WorkDir != tr.Dir {
		t.Errorf("WorkDir = %q, want %q", repo.WorkDir, tr.Dir)
	}
}

func TestDiscoverNotARepo(t *testing.T) {
	dir := t.TempDir()
	_, err := gitx.Discover(dir)
	if err == nil {
		t.Fatal("expected error for non-repo dir")
	}
}

func TestRun(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, _ := gitx.Discover(tr.Dir)

	out, err := repo.Run(context.Background(), "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	// Default branch is either "main" or "master".
	if out != "main" && out != "master" {
		t.Errorf("unexpected branch: %q", out)
	}
}

func TestRepoName(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, _ := gitx.Discover(tr.Dir)
	name := repo.RepoName()
	if name == "" {
		t.Fatal("RepoName is empty")
	}
}
