package spawn

import (
	"fmt"
	"os/exec"
	"runtime"
)

type genericTerminal struct{}

func (t *genericTerminal) Name() string { return "system default" }

func (t *genericTerminal) OpenTab(dir, shellCmd, title string) error {
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(`
tell application "Terminal"
	activate
	do script "cd %q && %s"
end tell`, dir, shellCmd)
		cmd := exec.Command("osascript", "-e", script)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Terminal.app: %w\n%s", err, out)
		}
		return nil

	case "linux":
		// Try common terminal emulators.
		for _, term := range []string{"x-terminal-emulator", "gnome-terminal", "xterm"} {
			if _, err := exec.LookPath(term); err == nil {
				cmd := exec.Command(term, "-e", fmt.Sprintf("sh -c 'cd %q && %s; exec $SHELL'", dir, shellCmd))
				if err := cmd.Start(); err == nil {
					return nil
				}
			}
		}
		return fmt.Errorf("no supported terminal emulator found; install gnome-terminal or xterm")

	default:
		return fmt.Errorf("spawn not supported on %s; use `wiz enter` instead", runtime.GOOS)
	}
}
