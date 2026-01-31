package shell

import (
	"embed"
	"fmt"
)

//go:embed scripts/wiz.bash
var bashScript string

//go:embed scripts/wiz.zsh
var zshScript string

//go:embed scripts/wiz.fish
var fishScript string

// InitScript returns the shell integration script for the given shell.
func InitScript(shellName string) (string, error) {
	switch shellName {
	case "bash":
		return bashScript, nil
	case "zsh":
		return zshScript, nil
	case "fish":
		return fishScript, nil
	default:
		return "", fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shellName)
	}
}

// Ensure the embed import is used.
var _ embed.FS
