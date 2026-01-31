package spawn

import (
	"github.com/firewood-buck-3000/wiz/internal/terminfo"
)

// Terminal represents a detected terminal emulator that can open new tabs/windows.
type Terminal interface {
	Name() string
	OpenTab(dir string, shellCmd string, title string) error
}

// Detect returns the best Terminal implementation for the current environment.
func Detect() Terminal {
	kind := terminfo.Detect()
	switch kind {
	case terminfo.ITerm2:
		return &iTerm2Terminal{}
	case terminfo.Kitty:
		return &kittyTerminal{}
	case terminfo.WezTerm:
		return &weztermTerminal{}
	case terminfo.Tmux:
		return &tmuxTerminal{}
	default:
		return &genericTerminal{}
	}
}

// DetectWithOverride returns a Terminal, allowing DI for testing.
func DetectWithOverride(t Terminal) Terminal {
	if t != nil {
		return t
	}
	return Detect()
}
