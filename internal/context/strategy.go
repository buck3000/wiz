package context

import (
	gocontext "context"

	"github.com/buck3000/wiz/internal/gitx"
)

// CreateOpts holds options for provisioning a new context.
type CreateOpts struct {
	Name       string
	Branch     string
	BaseBranch string
	Repo       *gitx.Repo
}

// Provisioner creates and destroys the filesystem backing for a context.
type Provisioner interface {
	// Create provisions a new context directory and returns its absolute path.
	Create(ctx gocontext.Context, opts CreateOpts) (string, error)
	// Destroy removes the context directory.
	Destroy(ctx gocontext.Context, path string, force bool) error
	// Strategy returns the strategy name.
	Strategy() Strategy
}

// NewProvisioner returns a provisioner for the given strategy.
// "auto" tries worktree first.
func NewProvisioner(strategy Strategy, repo *gitx.Repo) Provisioner {
	switch strategy {
	case StrategyClone:
		return &cloneProvisioner{repo: repo}
	default:
		return &worktreeProvisioner{repo: repo}
	}
}
