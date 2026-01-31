package orchestra

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPlanValid(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "tasks.yaml")
	os.WriteFile(f, []byte(`tasks:
  - name: fix-auth
    branch: fix/auth
    prompt: "Fix the auth bug"
    agent: claude
  - name: add-tests
    prompt: "Add tests"
    agent: gemini
`), 0o644)

	plan, err := LoadPlan(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(plan.Tasks))
	}
	if plan.Tasks[0].Name != "fix-auth" {
		t.Errorf("task 0 name = %q", plan.Tasks[0].Name)
	}
	if plan.Tasks[0].Branch != "fix/auth" {
		t.Errorf("task 0 branch = %q", plan.Tasks[0].Branch)
	}
	if plan.Tasks[1].Branch != "" {
		t.Errorf("task 1 branch should be empty, got %q", plan.Tasks[1].Branch)
	}
}

func TestLoadPlanEmptyTasks(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "empty.yaml")
	os.WriteFile(f, []byte(`tasks: []`), 0o644)

	_, err := LoadPlan(f)
	if err == nil {
		t.Fatal("expected error for empty tasks")
	}
}

func TestLoadPlanMissingName(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "no-name.yaml")
	os.WriteFile(f, []byte(`tasks:
  - prompt: "do stuff"
    agent: claude
`), 0o644)

	_, err := LoadPlan(f)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestLoadPlanMissingAgent(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "no-agent.yaml")
	os.WriteFile(f, []byte(`tasks:
  - name: test
    prompt: "do stuff"
`), 0o644)

	_, err := LoadPlan(f)
	if err == nil {
		t.Fatal("expected error for missing agent")
	}
}

func TestLoadPlanFileNotFound(t *testing.T) {
	_, err := LoadPlan("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadPlanInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "bad.yaml")
	os.WriteFile(f, []byte(`{{{not yaml`), 0o644)

	_, err := LoadPlan(f)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
