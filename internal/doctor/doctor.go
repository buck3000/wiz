package doctor

import (
	"os"
	"os/exec"
	"strings"

	"github.com/firewood-buck-3000/wiz/internal/terminfo"
)

// Status represents the result of a diagnostic check.
type Status int

const (
	OK   Status = iota
	Warn
	Fail
)

// CheckResult is a single diagnostic result.
type CheckResult struct {
	Name    string
	Status  Status
	Message string
}

// RunAll executes all diagnostic checks.
func RunAll() []CheckResult {
	var results []CheckResult
	results = append(results, checkGitVersion())
	results = append(results, checkGhCLI())
	results = append(results, checkTerminal())
	results = append(results, checkShellIntegration())
	results = append(results, checkActiveContext())
	return results
}

func checkGhCLI() CheckResult {
	out, err := exec.Command("gh", "--version").Output()
	if err != nil {
		return CheckResult{
			Name:    "GitHub CLI",
			Status:  Warn,
			Message: "gh not found in PATH; 'wiz finish' requires it. Install: https://cli.github.com",
		}
	}
	version := strings.TrimSpace(strings.Split(string(out), "\n")[0])
	return CheckResult{
		Name:    "GitHub CLI",
		Status:  OK,
		Message: version,
	}
}

func checkGitVersion() CheckResult {
	out, err := exec.Command("git", "--version").Output()
	if err != nil {
		return CheckResult{
			Name:    "Git",
			Status:  Fail,
			Message: "git not found in PATH",
		}
	}
	version := strings.TrimSpace(string(out))
	// git worktree was introduced in git 2.5, improved in 2.15+
	return CheckResult{
		Name:    "Git",
		Status:  OK,
		Message: version,
	}
}

func checkTerminal() CheckResult {
	kind := terminfo.Detect()
	name := kind.Name()

	features := []string{}
	if kind.SupportsOSCTitle() {
		features = append(features, "title")
	}
	if kind.SupportsBadge() {
		features = append(features, "badge")
	}
	if kind.SupportsTabColor() {
		features = append(features, "tab-color")
	}

	if kind == terminfo.Unknown {
		return CheckResult{
			Name:    "Terminal",
			Status:  Warn,
			Message: "Unknown terminal; title/badge enhancements unavailable",
		}
	}

	msg := name
	if len(features) > 0 {
		msg += " (features: " + strings.Join(features, ", ") + ")"
	}
	return CheckResult{
		Name:    "Terminal",
		Status:  OK,
		Message: msg,
	}
}

func checkShellIntegration() CheckResult {
	// Check if the wiz shell function is available by looking at env hint.
	if os.Getenv("WIZ_CTX") != "" || os.Getenv("WIZ_BRANCH") != "" {
		return CheckResult{
			Name:    "Shell integration",
			Status:  OK,
			Message: "Active (WIZ_CTX set)",
		}
	}
	return CheckResult{
		Name:    "Shell integration",
		Status:  Warn,
		Message: "Not detected. Add to your shell rc: eval \"$(wiz init <shell>)\"",
	}
}

func checkActiveContext() CheckResult {
	ctx := os.Getenv("WIZ_CTX")
	if ctx == "" {
		return CheckResult{
			Name:    "Active context",
			Status:  Warn,
			Message: "None. Use: wiz enter <name>",
		}
	}
	return CheckResult{
		Name:    "Active context",
		Status:  OK,
		Message: ctx,
	}
}
