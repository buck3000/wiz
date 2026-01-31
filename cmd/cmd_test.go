package cmd_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildWiz builds the wiz binary and returns its path.
func buildWiz(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "wiz")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = filepath.Join(mustFindModRoot(t))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

func mustFindModRoot(t *testing.T) string {
	t.Helper()
	// Walk up from this test file to find go.mod.
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find go.mod")
		}
		dir = parent
	}
}

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	dir, _ = filepath.EvalSymlinks(dir)
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@wiz.dev")
	run(t, dir, "git", "config", "user.name", "wiz-test")
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test\n"), 0o644)
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "init")
	return dir
}

func run(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func runWiz(t *testing.T, bin, dir string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func TestFullLifecycle(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	// Create context.
	stdout, _, err := runWiz(t, bin, repo, "create", "feat-x")
	if err != nil {
		t.Fatalf("create: %v\n%s", err, stdout)
	}
	if !strings.Contains(stdout, "feat-x") {
		t.Errorf("create output missing name: %s", stdout)
	}

	// List.
	stdout, _, err = runWiz(t, bin, repo, "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(stdout, "feat-x") {
		t.Errorf("list output missing feat-x: %s", stdout)
	}

	// Path.
	stdout, _, err = runWiz(t, bin, repo, "path", "feat-x")
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	ctxPath := strings.TrimSpace(stdout)
	if _, err := os.Stat(ctxPath); err != nil {
		t.Errorf("context path doesn't exist: %s", ctxPath)
	}

	// Enter.
	stdout, _, err = runWiz(t, bin, repo, "enter", "feat-x")
	if err != nil {
		t.Fatalf("enter: %v", err)
	}
	if !strings.Contains(stdout, "cd ") {
		t.Errorf("enter output missing cd: %s", stdout)
	}
	if !strings.Contains(stdout, "WIZ_CTX") {
		t.Errorf("enter output missing WIZ_CTX: %s", stdout)
	}

	// Run.
	stdout, _, err = runWiz(t, bin, repo, "run", "feat-x", "--", "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		t.Fatalf("run: %v\n%s", err, stdout)
	}
	if strings.TrimSpace(stdout) != "feat-x" {
		t.Errorf("run branch = %q, want feat-x", strings.TrimSpace(stdout))
	}

	// Rename.
	stdout, _, err = runWiz(t, bin, repo, "rename", "feat-x", "feat-y")
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if !strings.Contains(stdout, "feat-y") {
		t.Errorf("rename output: %s", stdout)
	}

	// List should show new name.
	stdout, _, _ = runWiz(t, bin, repo, "list")
	if !strings.Contains(stdout, "feat-y") {
		t.Errorf("list after rename: %s", stdout)
	}

	// Delete.
	stdout, _, err = runWiz(t, bin, repo, "delete", "feat-y")
	if err != nil {
		t.Fatalf("delete: %v\n%s", err, stdout)
	}

	// List should be empty.
	stdout, _, _ = runWiz(t, bin, repo, "list")
	if strings.Contains(stdout, "feat-y") {
		t.Errorf("list after delete still shows feat-y: %s", stdout)
	}
}

func TestCreateWithBase(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	// Create a branch with content.
	run(t, repo, "git", "checkout", "-b", "base-branch")
	os.WriteFile(filepath.Join(repo, "base.txt"), []byte("base"), 0o644)
	run(t, repo, "git", "add", ".")
	run(t, repo, "git", "commit", "-m", "base commit")
	run(t, repo, "git", "checkout", "-")

	// Create context from base.
	_, _, err := runWiz(t, bin, repo, "create", "from-base", "--base", "base-branch")
	if err != nil {
		t.Fatal(err)
	}

	// Verify the base file exists in the context.
	stdout, _, _ := runWiz(t, bin, repo, "run", "from-base", "--", "cat", "base.txt")
	if strings.TrimSpace(stdout) != "base" {
		t.Errorf("base.txt content = %q", stdout)
	}

	// Clean up.
	runWiz(t, bin, repo, "delete", "from-base", "--force")
}

func TestListJSON(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	runWiz(t, bin, repo, "create", "json-test")
	stdout, _, err := runWiz(t, bin, repo, "list", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, `"name"`) {
		t.Errorf("JSON output missing name field: %s", stdout)
	}
	if !strings.Contains(stdout, "json-test") {
		t.Errorf("JSON output missing context: %s", stdout)
	}

	runWiz(t, bin, repo, "delete", "json-test", "--force")
}

func TestRunExecutesInCorrectDir(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	runWiz(t, bin, repo, "create", "dir-test")

	// pwd should be the context path.
	stdout, _, err := runWiz(t, bin, repo, "run", "dir-test", "--", "pwd")
	if err != nil {
		t.Fatal(err)
	}

	ctxPath, _, _ := runWiz(t, bin, repo, "path", "dir-test")
	ctxPath = strings.TrimSpace(ctxPath)
	got := strings.TrimSpace(stdout)

	// Resolve symlinks for comparison.
	ctxPath, _ = filepath.EvalSymlinks(ctxPath)
	got, _ = filepath.EvalSymlinks(got)

	if got != ctxPath {
		t.Errorf("pwd = %q, want %q", got, ctxPath)
	}

	runWiz(t, bin, repo, "delete", "dir-test", "--force")
}

func TestDoctorRuns(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	stdout, _, err := runWiz(t, bin, repo, "doctor")
	if err != nil {
		t.Fatalf("doctor: %v\n%s", err, stdout)
	}
	if !strings.Contains(stdout, "Git") {
		t.Errorf("doctor output missing Git check: %s", stdout)
	}
}

func TestInitShell(t *testing.T) {
	bin := buildWiz(t)
	dir := t.TempDir()

	for _, sh := range []string{"bash", "zsh", "fish"} {
		stdout, _, err := runWiz(t, bin, dir, "init", sh)
		if err != nil {
			t.Fatalf("init %s: %v", sh, err)
		}
		if !strings.Contains(stdout, "wiz") {
			t.Errorf("init %s output missing wiz: %s", sh, stdout)
		}
	}
}

func TestCreateWithTask(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	_, _, err := runWiz(t, bin, repo, "create", "task-test", "--task", "Fix the OAuth bug", "--agent", "claude")
	if err != nil {
		t.Fatal(err)
	}

	// JSON output should include task and agent fields.
	stdout, _, err := runWiz(t, bin, repo, "list", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, `"task"`) {
		t.Errorf("JSON output missing task field: %s", stdout)
	}
	if !strings.Contains(stdout, "Fix the OAuth bug") {
		t.Errorf("JSON output missing task value: %s", stdout)
	}
	if !strings.Contains(stdout, `"agent"`) {
		t.Errorf("JSON output missing agent field: %s", stdout)
	}
	if !strings.Contains(stdout, "claude") {
		t.Errorf("JSON output missing agent value: %s", stdout)
	}

	// --tasks flag should show task/agent in human-readable output.
	stdout, _, err = runWiz(t, bin, repo, "list", "--tasks")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "task:") {
		t.Errorf("--tasks output missing task line: %s", stdout)
	}
	if !strings.Contains(stdout, "agent:") {
		t.Errorf("--tasks output missing agent line: %s", stdout)
	}

	runWiz(t, bin, repo, "delete", "task-test", "--force")
}

func TestTemplateWorkflow(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	// Save a template.
	stdout, _, err := runWiz(t, bin, repo, "template", "save", "bugfix", "--base", "main", "--agent", "claude")
	if err != nil {
		t.Fatalf("template save: %v\n%s", err, stdout)
	}

	// List templates.
	stdout, _, err = runWiz(t, bin, repo, "template", "list")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "bugfix") {
		t.Errorf("template list missing bugfix: %s", stdout)
	}

	// Create context with template.
	stdout, _, err = runWiz(t, bin, repo, "create", "my-fix", "--template", "bugfix", "--task", "Fix the bug")
	if err != nil {
		t.Fatalf("create with template: %v\n%s", err, stdout)
	}

	// Verify template defaults were applied.
	stdout, _, err = runWiz(t, bin, repo, "list", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, `"agent"`) || !strings.Contains(stdout, "claude") {
		t.Errorf("template agent not applied: %s", stdout)
	}
	if !strings.Contains(stdout, "Fix the bug") {
		t.Errorf("task not preserved: %s", stdout)
	}

	// Delete template.
	stdout, _, err = runWiz(t, bin, repo, "template", "delete", "bugfix")
	if err != nil {
		t.Fatal(err)
	}

	// Cleanup.
	runWiz(t, bin, repo, "delete", "my-fix", "--force")
}

func TestFinishNotFound(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	_, _, err := runWiz(t, bin, repo, "finish", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent context")
	}
}

func TestStatusPorcelain(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	runWiz(t, bin, repo, "create", "status-test")

	// Run status with WIZ env vars set.
	ctxPath, _, _ := runWiz(t, bin, repo, "path", "status-test")
	ctxPath = strings.TrimSpace(ctxPath)

	cmd := exec.Command(bin, "status", "--porcelain")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(),
		"WIZ_CTX=status-test",
		"WIZ_REPO=test",
		"WIZ_DIR="+ctxPath,
		"WIZ_BRANCH=status-test",
	)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	out := strings.TrimSpace(stdout.String())
	if !strings.Contains(out, "status-test") {
		t.Errorf("porcelain output = %q", out)
	}
	if !strings.Contains(out, "clean") {
		t.Errorf("expected clean status: %q", out)
	}

	runWiz(t, bin, repo, "delete", "status-test", "--force")
}
