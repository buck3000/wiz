package gitx

import (
	"context"
	"strings"
)

// BranchInfo represents a local git branch.
type BranchInfo struct {
	Name      string
	Ref       string
	Upstream  string
	IsHead    bool
	CommitMsg string
}

// Branches lists local branches using git for-each-ref.
func (r *Repo) Branches(ctx context.Context) ([]BranchInfo, error) {
	format := "%(refname:short)%09%(objectname:short)%09%(upstream:short)%09%(HEAD)%09%(subject)"
	lines, err := r.RunLines(ctx, "for-each-ref", "--format="+format, "refs/heads/")
	if err != nil {
		return nil, err
	}

	var branches []BranchInfo
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 5)
		if len(parts) < 5 {
			continue
		}
		branches = append(branches, BranchInfo{
			Name:      parts[0],
			Ref:       parts[1],
			Upstream:  parts[2],
			IsHead:    parts[3] == "*",
			CommitMsg: parts[4],
		})
	}
	return branches, nil
}

// BranchExists checks whether a branch with the given name exists.
func (r *Repo) BranchExists(ctx context.Context, name string) bool {
	_, err := r.Run(ctx, "rev-parse", "--verify", "refs/heads/"+name)
	return err == nil
}

// CurrentBranch returns the current branch name, or "(detached)" if detached.
func (r *Repo) CurrentBranch(ctx context.Context) (string, error) {
	out, err := r.Run(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return out, nil
}
