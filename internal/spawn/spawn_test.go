package spawn_test

import (
	"testing"

	"github.com/firewood-buck-3000/wiz/internal/spawn"
)

// mockTerminal records the OpenTab call for testing.
type mockTerminal struct {
	OpenTabCalls []openTabCall
}

type openTabCall struct {
	Dir      string
	ShellCmd string
	Title    string
}

func (m *mockTerminal) Name() string { return "mock" }
func (m *mockTerminal) OpenTab(dir, shellCmd, title string) error {
	m.OpenTabCalls = append(m.OpenTabCalls, openTabCall{dir, shellCmd, title})
	return nil
}

func TestDetect(t *testing.T) {
	// Detect should not panic regardless of environment.
	term := spawn.Detect()
	if term.Name() == "" {
		t.Error("Name() returned empty")
	}
}

func TestDetectWithOverride(t *testing.T) {
	mock := &mockTerminal{}
	term := spawn.DetectWithOverride(mock)
	if term.Name() != "mock" {
		t.Errorf("Name() = %q, want mock", term.Name())
	}

	err := term.OpenTab("/tmp", "echo hello", "test")
	if err != nil {
		t.Fatal(err)
	}
	if len(mock.OpenTabCalls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.OpenTabCalls))
	}
	call := mock.OpenTabCalls[0]
	if call.Dir != "/tmp" {
		t.Errorf("Dir = %q", call.Dir)
	}
	if call.Title != "test" {
		t.Errorf("Title = %q", call.Title)
	}
}
