package spawn

import (
	"fmt"
	"os/exec"
)

type kittyTerminal struct{}

func (t *kittyTerminal) Name() string { return "Kitty" }

func (t *kittyTerminal) OpenTab(dir, shellCmd, title string) error {
	args := []string{
		"@", "launch",
		"--type=tab",
		"--tab-title", title,
		"--cwd", dir,
		"--", "sh", "-c", shellCmd,
	}
	cmd := exec.Command("kitty", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kitty: %w\n%s", err, out)
	}
	return nil
}
