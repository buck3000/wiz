package spawn

import (
	"fmt"
	"os/exec"
)

type tmuxTerminal struct{}

func (t *tmuxTerminal) Name() string { return "tmux" }

func (t *tmuxTerminal) OpenTab(dir, shellCmd, title string) error {
	cmd := exec.Command("tmux", "new-window",
		"-c", dir,
		"-n", title,
		"sh", "-c", shellCmd+"; exec $SHELL",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux: %w\n%s", err, out)
	}
	return nil
}
