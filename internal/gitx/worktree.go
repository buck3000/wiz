package gitx

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// WorktreeInfo represents a git worktree entry.
type WorktreeInfo struct {
	Path     string
	HEAD     string
	Branch   string
	Bare     bool
	Detached bool
}

// WorktreeList returns all worktrees via git worktree list --porcelain.
func (r *Repo) WorktreeList(ctx context.Context) ([]WorktreeInfo, error) {
	cmd := exec.CommandContext(ctx, "git", "worktree", "list", "--porcelain")
	cmd.Dir = r.WorkDir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w", err)
	}
	return parseWorktreeList(string(out)), nil
}

// WorktreeAdd creates a new worktree. If createBranch is true, it creates a
// new branch with the given name at baseBranch (or HEAD if empty).
func (r *Repo) WorktreeAdd(ctx context.Context, path, branch string, createBranch bool, baseBranch string) error {
	args := []string{"worktree", "add"}
	if createBranch {
		args = append(args, "-b", branch, path)
		if baseBranch != "" {
			args = append(args, baseBranch)
		}
	} else {
		args = append(args, path, branch)
	}
	_, err := r.Run(ctx, args...)
	return err
}

// WorktreeRemove removes a worktree at the given path.
func (r *Repo) WorktreeRemove(ctx context.Context, path string, force bool) error {
	args := []string{"worktree", "remove", path}
	if force {
		args = append(args, "--force")
	}
	_, err := r.Run(ctx, args...)
	return err
}

func parseWorktreeList(output string) []WorktreeInfo {
	var result []WorktreeInfo
	var current *WorktreeInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "worktree "):
			if current != nil {
				result = append(result, *current)
			}
			current = &WorktreeInfo{
				Path: strings.TrimPrefix(line, "worktree "),
			}
		case strings.HasPrefix(line, "HEAD "):
			if current != nil {
				current.HEAD = strings.TrimPrefix(line, "HEAD ")
			}
		case strings.HasPrefix(line, "branch "):
			if current != nil {
				current.Branch = strings.TrimPrefix(line, "branch ")
			}
		case line == "bare":
			if current != nil {
				current.Bare = true
			}
		case line == "detached":
			if current != nil {
				current.Detached = true
			}
		case line == "":
			// separator between entries
		}
	}
	if current != nil {
		result = append(result, *current)
	}
	return result
}
