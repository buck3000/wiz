package agent

import "strings"

// Agent describes how to invoke a coding agent CLI.
type Agent struct {
	Name    string
	Command string
	Args    []string
}

// BuildCommand returns a full shell command string for spawning this agent with a prompt.
func (a *Agent) BuildCommand(prompt string) string {
	parts := []string{a.Command}
	parts = append(parts, a.Args...)
	if prompt != "" {
		parts = append(parts, shellEscape(prompt))
	}
	return strings.Join(parts, " ")
}

// BuildExecArgs returns the executable and arguments for exec.Command (no shell).
func (a *Agent) BuildExecArgs(prompt string) (string, []string) {
	args := make([]string, len(a.Args))
	copy(args, a.Args)
	if prompt != "" {
		args = append(args, prompt)
	}
	return a.Command, args
}

// shellEscape wraps a string in single quotes, escaping internal single quotes.
func shellEscape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}
