package agent

import (
	"fmt"
	"os/exec"
	"sort"

	"github.com/buck3000/wiz/internal/config"
	"github.com/buck3000/wiz/internal/gitx"
)

var builtins = map[string]Agent{
	"claude": {Name: "claude", Command: "claude", Args: nil},
	"gemini": {Name: "gemini", Command: "gemini", Args: nil},
	"codex":  {Name: "codex", Command: "codex", Args: nil},
}

// Resolve looks up an agent by name: config custom agents first, then builtins.
func Resolve(repo *gitx.Repo, name string) (*Agent, error) {
	cfg := config.Load(repo)
	if custom, ok := cfg.Agents[name]; ok {
		a := &Agent{
			Name:    name,
			Command: custom.Command,
			Args:    custom.Args,
		}
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return a, nil
	}
	if a, ok := builtins[name]; ok {
		if err := a.Validate(); err != nil {
			return nil, err
		}
		return &a, nil
	}
	return nil, fmt.Errorf("unknown agent %q; known agents: claude, gemini, codex", name)
}

// Validate checks that the agent's command binary exists in PATH.
func (a *Agent) Validate() error {
	_, err := exec.LookPath(a.Command)
	if err != nil {
		return fmt.Errorf("agent %q: command %q not found in PATH; install it first", a.Name, a.Command)
	}
	return nil
}

// List returns all built-in agent names sorted alphabetically.
func List() []string {
	names := make([]string, 0, len(builtins))
	for k := range builtins {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
