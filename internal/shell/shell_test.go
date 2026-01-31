package shell_test

import (
	"strings"
	"testing"

	"github.com/buck3000/wiz/internal/shell"
)

func TestInitScriptBash(t *testing.T) {
	s, err := shell.InitScript("bash")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "wiz()") {
		t.Error("bash script missing wiz() function")
	}
	if !strings.Contains(s, "PROMPT_COMMAND") {
		t.Error("bash script missing PROMPT_COMMAND")
	}
}

func TestInitScriptZsh(t *testing.T) {
	s, err := shell.InitScript("zsh")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "wiz()") {
		t.Error("zsh script missing wiz() function")
	}
	if !strings.Contains(s, "precmd_functions") {
		t.Error("zsh script missing precmd_functions")
	}
}

func TestInitScriptFish(t *testing.T) {
	s, err := shell.InitScript("fish")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "function wiz") {
		t.Error("fish script missing wiz function")
	}
}

func TestInitScriptUnsupported(t *testing.T) {
	_, err := shell.InitScript("powershell")
	if err == nil {
		t.Error("expected error for unsupported shell")
	}
}
