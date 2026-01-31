package context

import (
	"fmt"
	"regexp"
	"time"
)

// Strategy identifies how a context is backed.
type Strategy string

const (
	StrategyAuto     Strategy = "auto"
	StrategyWorktree Strategy = "worktree"
	StrategyClone    Strategy = "clone"
)

// ParseStrategy parses a strategy string, defaulting to auto.
func ParseStrategy(s string) Strategy {
	switch s {
	case "worktree":
		return StrategyWorktree
	case "clone":
		return StrategyClone
	default:
		return StrategyAuto
	}
}

// Context represents a wiz context — an isolated working directory tied to a git branch.
type Context struct {
	Name       string   `json:"name"`
	Branch     string   `json:"branch"`
	Path       string   `json:"path"`
	Strategy   Strategy `json:"strategy"`
	CreatedAt  time.Time `json:"created_at"`
	BaseBranch string   `json:"base_branch,omitempty"`
	Task       string   `json:"task,omitempty"`
	Agent      string   `json:"agent,omitempty"`
}

var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/-]*$`)

// ValidateName checks that a context name is safe for use as a directory name.
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("context name cannot be empty")
	}
	if len(name) > 128 {
		return fmt.Errorf("context name too long (max 128 chars)")
	}
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid context name %q: must start with alphanumeric and contain only [a-zA-Z0-9._/-]", name)
	}
	return nil
}

// SafeDirName converts a context name to a safe directory component.
func SafeDirName(name string) string {
	return regexp.MustCompile(`[/\\]`).ReplaceAllString(name, "__")
}
