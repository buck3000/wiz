package agent

import (
	"strings"
	"testing"
)

func TestBuildCommand(t *testing.T) {
	a := &Agent{Name: "claude", Command: "claude", Args: nil}

	// No prompt.
	got := a.BuildCommand("")
	if got != "claude" {
		t.Errorf("BuildCommand('') = %q, want %q", got, "claude")
	}

	// Simple prompt.
	got = a.BuildCommand("Fix the bug")
	if !strings.Contains(got, "claude") || !strings.Contains(got, "Fix the bug") {
		t.Errorf("BuildCommand simple = %q", got)
	}

	// Prompt with single quotes.
	got = a.BuildCommand("Fix it's bug")
	if !strings.Contains(got, `'"'"'`) {
		t.Errorf("BuildCommand with quote = %q, expected escaped single quote", got)
	}
}

func TestBuildCommandWithArgs(t *testing.T) {
	a := &Agent{Name: "custom", Command: "my-agent", Args: []string{"--verbose", "--mode=fast"}}
	got := a.BuildCommand("do stuff")
	if !strings.HasPrefix(got, "my-agent --verbose --mode=fast") {
		t.Errorf("BuildCommand with args = %q", got)
	}
}

func TestBuildExecArgs(t *testing.T) {
	a := &Agent{Name: "claude", Command: "claude", Args: []string{"--flag"}}

	bin, args := a.BuildExecArgs("hello")
	if bin != "claude" {
		t.Errorf("bin = %q", bin)
	}
	if len(args) != 2 || args[0] != "--flag" || args[1] != "hello" {
		t.Errorf("args = %v", args)
	}

	// No prompt.
	bin, args = a.BuildExecArgs("")
	if bin != "claude" {
		t.Errorf("bin = %q", bin)
	}
	if len(args) != 1 || args[0] != "--flag" {
		t.Errorf("args without prompt = %v", args)
	}
}

func TestBuildExecArgsNoMutation(t *testing.T) {
	a := &Agent{Name: "claude", Command: "claude", Args: []string{"--flag"}}
	a.BuildExecArgs("prompt1")
	_, args := a.BuildExecArgs("prompt2")
	if len(args) != 2 {
		t.Errorf("mutation detected: args = %v", args)
	}
}

func TestShellEscape(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "'hello'"},
		{"it's here", `'it'"'"'s here'`},
		{"$HOME", "'$HOME'"},
		{`back\slash`, `'back\slash'`},
	}
	for _, tt := range tests {
		got := shellEscape(tt.input)
		if got != tt.want {
			t.Errorf("shellEscape(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestListReturnsKnownAgents(t *testing.T) {
	names := List()
	if len(names) < 3 {
		t.Errorf("List() returned %d agents, want >= 3", len(names))
	}
	// Should be sorted.
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("List() not sorted: %v", names)
			break
		}
	}
}
