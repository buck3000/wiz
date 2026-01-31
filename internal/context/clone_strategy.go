package context

import (
	gocontext "context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/firewood-buck-3000/wiz/internal/config"
	"github.com/firewood-buck-3000/wiz/internal/gitx"
)

type cloneProvisioner struct {
	repo *gitx.Repo
}

func (c *cloneProvisioner) Strategy() Strategy {
	return StrategyClone
}

func (c *cloneProvisioner) Create(ctx gocontext.Context, opts CreateOpts) (string, error) {
	clonesDir := config.ClonesDir(opts.Repo)
	if err := os.MkdirAll(clonesDir, 0o755); err != nil {
		return "", fmt.Errorf("create clones dir: %w", err)
	}

	dirName := SafeDirName(opts.Name)
	clonePath := filepath.Join(clonesDir, dirName)

	// git clone --shared <repo-workdir> <clone-path>
	cmd := exec.CommandContext(ctx, "git", "clone", "--shared", opts.Repo.WorkDir, clonePath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git clone --shared: %w\n%s", err, out)
	}

	// Checkout the target branch.
	cloneRepo, err := gitx.Discover(clonePath)
	if err != nil {
		os.RemoveAll(clonePath)
		return "", fmt.Errorf("discover clone: %w", err)
	}

	if cloneRepo.BranchExists(ctx, opts.Branch) {
		if _, err := cloneRepo.Run(ctx, "checkout", opts.Branch); err != nil {
			os.RemoveAll(clonePath)
			return "", fmt.Errorf("checkout branch: %w", err)
		}
	} else {
		base := opts.BaseBranch
		if base == "" {
			base = "HEAD"
		}
		if _, err := cloneRepo.Run(ctx, "checkout", "-b", opts.Branch, base); err != nil {
			os.RemoveAll(clonePath)
			return "", fmt.Errorf("create branch: %w", err)
		}
	}

	return clonePath, nil
}

func (c *cloneProvisioner) Destroy(_ gocontext.Context, path string, force bool) error {
	if !force {
		// Check for uncommitted changes.
		repo, err := gitx.Discover(path)
		if err == nil {
			st, err := repo.Status(gocontext.Background())
			if err == nil && st.Dirty {
				return fmt.Errorf("context has uncommitted changes; use --force to delete anyway")
			}
		}
	}
	return os.RemoveAll(path)
}
