package spawn

import (
	"fmt"
	"os/exec"
)

type iTerm2Terminal struct{}

func (t *iTerm2Terminal) Name() string { return "iTerm2" }

func (t *iTerm2Terminal) OpenTab(dir, shellCmd, title string) error {
	script := fmt.Sprintf(`
tell application "iTerm2"
	tell current window
		create tab with default profile
		tell current session
			write text "cd %q && %s"
			set name to %q
		end tell
	end tell
end tell`, dir, shellCmd, title)

	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("iTerm2 AppleScript: %w\n%s", err, out)
	}
	return nil
}
