package terminfo

import "os"

// Kind identifies a terminal emulator.
type Kind int

const (
	Unknown Kind = iota
	ITerm2
	Kitty
	WezTerm
	Alacritty
	Tmux
	VSCode
)

// Detect returns the terminal emulator kind based on environment variables.
func Detect() Kind {
	switch os.Getenv("TERM_PROGRAM") {
	case "iTerm.app":
		return ITerm2
	case "WezTerm":
		return WezTerm
	case "vscode":
		return VSCode
	}
	if os.Getenv("KITTY_PID") != "" {
		return Kitty
	}
	if os.Getenv("ALACRITTY_WINDOW_ID") != "" {
		return Alacritty
	}
	if os.Getenv("TMUX") != "" {
		return Tmux
	}
	return Unknown
}

// Name returns a human-readable name for the terminal kind.
func (k Kind) Name() string {
	switch k {
	case ITerm2:
		return "iTerm2"
	case Kitty:
		return "Kitty"
	case WezTerm:
		return "WezTerm"
	case Alacritty:
		return "Alacritty"
	case Tmux:
		return "tmux"
	case VSCode:
		return "VS Code"
	default:
		return "Unknown"
	}
}

// SupportsOSCTitle returns whether the terminal supports OSC 0 title setting.
func (k Kind) SupportsOSCTitle() bool {
	return k != Unknown
}

// SupportsBadge returns whether the terminal supports badge text.
func (k Kind) SupportsBadge() bool {
	return k == ITerm2
}

// SupportsTabColor returns whether the terminal supports tab color changes.
func (k Kind) SupportsTabColor() bool {
	return k == ITerm2
}
