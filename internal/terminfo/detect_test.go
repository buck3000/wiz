package terminfo_test

import (
	"testing"

	"github.com/buck3000/wiz/internal/terminfo"
)

func TestDetect(t *testing.T) {
	// Just verify it doesn't panic and returns a valid kind.
	k := terminfo.Detect()
	_ = k.Name()
	_ = k.SupportsOSCTitle()
	_ = k.SupportsBadge()
}

func TestKindName(t *testing.T) {
	tests := []struct {
		kind terminfo.Kind
		name string
	}{
		{terminfo.ITerm2, "iTerm2"},
		{terminfo.Kitty, "Kitty"},
		{terminfo.WezTerm, "WezTerm"},
		{terminfo.Tmux, "tmux"},
		{terminfo.Unknown, "Unknown"},
	}
	for _, tc := range tests {
		if tc.kind.Name() != tc.name {
			t.Errorf("Kind(%d).Name() = %q, want %q", tc.kind, tc.kind.Name(), tc.name)
		}
	}
}
