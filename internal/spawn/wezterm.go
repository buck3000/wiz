package spawn

import (
	"fmt"
	"os/exec"
)

type weztermTerminal struct{}

func (t *weztermTerminal) Name() string { return "WezTerm" }

func (t *weztermTerminal) OpenTab(dir, shellCmd, title string) error {
	cmd := exec.Command("wezterm", "cli", "spawn",
		"--cwd", dir,
		"--", "sh", "-c", shellCmd,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wezterm: %w\n%s", err, out)
	}
	return nil
}
