package config_test

import (
	"path/filepath"
	"testing"

	"github.com/buck3000/wiz/internal/config"
	"github.com/buck3000/wiz/internal/gitx"
	"github.com/buck3000/wiz/testutil"
)

func TestWizDir(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, err := gitx.Discover(tr.Dir)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(repo.CommonDir, "wiz")
	got := config.WizDir(repo)
	if got != want {
		t.Errorf("WizDir = %q, want %q", got, want)
	}
}

func TestPaths(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, _ := gitx.Discover(tr.Dir)

	state := config.StateFile(repo)
	if filepath.Base(state) != "state.json" {
		t.Errorf("StateFile base = %q", filepath.Base(state))
	}

	lockPath := config.LockFilePath(repo)
	if filepath.Base(lockPath) != "wiz.lock" {
		t.Errorf("LockFilePath base = %q", filepath.Base(lockPath))
	}

	trees := config.TreesDir(repo)
	if filepath.Base(trees) != "trees" {
		t.Errorf("TreesDir base = %q", filepath.Base(trees))
	}
}

func TestLoadDefaults(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, _ := gitx.Discover(tr.Dir)

	cfg := config.Load(repo)
	if cfg.DefaultStrategy != "auto" {
		t.Errorf("DefaultStrategy = %q", cfg.DefaultStrategy)
	}
	if cfg.PromptEmoji == "" {
		t.Error("PromptEmoji is empty")
	}
}
