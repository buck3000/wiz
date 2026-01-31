package context

import (
	gocontext "context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/firewood-buck-3000/wiz/internal/config"
	"github.com/firewood-buck-3000/wiz/internal/gitx"
)

type worktreeProvisioner struct {
	repo *gitx.Repo
}

func (w *worktreeProvisioner) Strategy() Strategy {
	return StrategyWorktree
}

func (w *worktreeProvisioner) Create(ctx gocontext.Context, opts CreateOpts) (string, error) {
	treesDir := config.TreesDir(opts.Repo)
	if err := os.MkdirAll(treesDir, 0o755); err != nil {
		return "", fmt.Errorf("create trees dir: %w", err)
	}

	dirName := SafeDirName(opts.Name)
	wtPath := filepath.Join(treesDir, dirName)

	// Check if branch already exists.
	createBranch := !opts.Repo.BranchExists(ctx, opts.Branch)
	if err := opts.Repo.WorktreeAdd(ctx, wtPath, opts.Branch, createBranch, opts.BaseBranch); err != nil {
		return "", fmt.Errorf("create context: %w", err)
	}

	return wtPath, nil
}

func (w *worktreeProvisioner) Destroy(ctx gocontext.Context, path string, force bool) error {
	err := w.repo.WorktreeRemove(ctx, path, force)
	if err != nil {
		// If git worktree remove fails, try cleaning up manually.
		if force {
			os.RemoveAll(path)
		}
		return err
	}
	return nil
}
