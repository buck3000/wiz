package orchestra

import (
	"context"
	"sync"
	"testing"

	"github.com/firewood-buck-3000/wiz/internal/gitx"
	"github.com/firewood-buck-3000/wiz/testutil"
)

// mockTerminal records OpenTab calls.
type mockTerminal struct {
	mu    sync.Mutex
	calls []openTabCall
}

type openTabCall struct {
	Dir      string
	ShellCmd string
	Title    string
}

func (m *mockTerminal) Name() string { return "mock" }

func (m *mockTerminal) OpenTab(dir, shellCmd, title string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, openTabCall{Dir: dir, ShellCmd: shellCmd, Title: title})
	return nil
}

func TestRunCreatesContextsAndSpawns(t *testing.T) {
	tr := testutil.NewTestRepo(t)

	repo, err := gitx.Discover(tr.Dir)
	if err != nil {
		t.Fatal(err)
	}

	plan := &Plan{
		Tasks: []TaskDef{
			{Name: "task-a", Prompt: "Do A", Agent: "claude"},
			{Name: "task-b", Prompt: "Do B", Agent: "claude"},
		},
	}

	term := &mockTerminal{}
	results := Run(context.Background(), repo, plan, term)

	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	for _, r := range results {
		if r.Error != nil {
			t.Errorf("task %s failed: %v", r.Name, r.Error)
		}
	}

	term.mu.Lock()
	defer term.mu.Unlock()
	if len(term.calls) != 2 {
		t.Errorf("got %d OpenTab calls, want 2", len(term.calls))
	}
}

func TestRunHandlesDuplicateNames(t *testing.T) {
	tr := testutil.NewTestRepo(t)

	repo, err := gitx.Discover(tr.Dir)
	if err != nil {
		t.Fatal(err)
	}

	plan := &Plan{
		Tasks: []TaskDef{
			{Name: "dup", Prompt: "First", Agent: "claude"},
			{Name: "dup", Prompt: "Second", Agent: "claude"},
		},
	}

	term := &mockTerminal{}
	results := Run(context.Background(), repo, plan, term)

	// Second task should fail because name already exists.
	if results[0].Error != nil {
		t.Errorf("first task should succeed: %v", results[0].Error)
	}
	if results[1].Error == nil {
		t.Error("second task with duplicate name should fail")
	}
}
