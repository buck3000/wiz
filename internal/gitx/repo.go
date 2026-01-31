package gitx

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repo represents a discovered git repository.
type Repo struct {
	// WorkDir is the working directory where git was discovered.
	WorkDir string
	// GitDir is the per-worktree .git directory (git rev-parse --git-dir).
	GitDir string
	// CommonDir is the shared .git directory (git rev-parse --git-common-dir).
	CommonDir string
}

// Discover finds the git repository from the given path.
func Discover(fromPath string) (*Repo, error) {
	absPath, err := filepath.Abs(fromPath)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}
	// Resolve symlinks (macOS /var -> /private/var).
	absPath, err = filepath.EvalSymlinks(absPath)
	if err != nil {
		return nil, fmt.Errorf("resolve symlinks: %w", err)
	}

	workDir, err := gitOutput(absPath, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("not a git repository (or any parent): %w", err)
	}

	gitDir, err := gitOutput(absPath, "rev-parse", "--git-dir")
	if err != nil {
		return nil, fmt.Errorf("could not determine git dir: %w", err)
	}
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(absPath, gitDir)
	}
	gitDir = filepath.Clean(gitDir)

	commonDir, err := gitOutput(absPath, "rev-parse", "--git-common-dir")
	if err != nil {
		return nil, fmt.Errorf("could not determine git common dir: %w", err)
	}
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(absPath, commonDir)
	}
	commonDir = filepath.Clean(commonDir)

	return &Repo{
		WorkDir:   workDir,
		GitDir:    gitDir,
		CommonDir: commonDir,
	}, nil
}

// Run executes a git command in the repo and returns combined stdout.
func (r *Repo) Run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.WorkDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimRight(string(out), "\n"), nil
}

// RunLines executes a git command and returns output split by newlines.
func (r *Repo) RunLines(ctx context.Context, args ...string) ([]string, error) {
	out, err := r.Run(ctx, args...)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

// HasCommits returns true if the repository has at least one commit.
func (r *Repo) HasCommits(ctx context.Context) bool {
	_, err := r.Run(ctx, "rev-parse", "--verify", "HEAD")
	return err == nil
}

// RepoName returns the base name of the working directory.
func (r *Repo) RepoName() string {
	return filepath.Base(r.WorkDir)
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(out), "\n"), nil
}
