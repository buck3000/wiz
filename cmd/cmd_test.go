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

// runWizEnv runs wiz with extra environment variables.
func runWizEnv(t *testing.T, bin, dir string, env []string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// ============================================================
// Porcelain + escape hatch tests
// ============================================================

func TestCheckoutCreatesAndEnters(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	// checkout a new context — should create it.
	stdout, stderr, err := runWiz(t, bin, repo, "checkout", "feat-x")
	if err != nil {
		t.Fatalf("checkout create: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
	if !strings.Contains(stderr, "Created context") {
		t.Errorf("expected 'Created context' in stderr: %s", stderr)
	}
	if !strings.Contains(stdout, "cd ") {
		t.Errorf("checkout output missing cd: %s", stdout)
	}
	if !strings.Contains(stdout, "WIZ_CTX") {
		t.Errorf("checkout output missing WIZ_CTX: %s", stdout)
	}

	// checkout again — should just enter (not re-create).
	stdout, stderr, err = runWiz(t, bin, repo, "checkout", "feat-x")
	if err != nil {
		t.Fatalf("checkout enter: %v", err)
	}
	if strings.Contains(stderr, "Created context") {
		t.Errorf("checkout of existing context should not re-create")
	}
	if !strings.Contains(stdout, "cd ") {
		t.Errorf("checkout enter output missing cd: %s", stdout)
	}

	// Cleanup.
	runWiz(t, bin, repo, "delete", "feat-x", "--force")
}

func TestCheckoutDash(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	// Create two contexts.
	runWiz(t, bin, repo, "create", "ctx-a")
	runWiz(t, bin, repo, "create", "ctx-b")

	// "Enter" ctx-a, then checkout ctx-b — ctx-a should be saved as last.
	_, _, err := runWizEnv(t, bin, repo, []string{"WIZ_CTX=ctx-a"}, "checkout", "ctx-b")
	if err != nil {
		t.Fatalf("checkout ctx-b: %v", err)
	}

	// Now checkout - should go back to ctx-a.
	stdout, _, err := runWiz(t, bin, repo, "checkout", "-")
	if err != nil {
		t.Fatalf("checkout -: %v", err)
	}
	if !strings.Contains(stdout, "ctx-a") {
		t.Errorf("checkout - should return to ctx-a, got: %s", stdout)
	}

	// Cleanup.
	runWiz(t, bin, repo, "delete", "ctx-a", "--force")
	runWiz(t, bin, repo, "delete", "ctx-b", "--force")
}

func TestCheckoutWithBFlag(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	stdout, stderr, err := runWiz(t, bin, repo, "checkout", "-b", "new-branch")
	if err != nil {
		t.Fatalf("checkout -b: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
	if !strings.Contains(stderr, "Created context") {
		t.Errorf("expected 'Created context' in stderr: %s", stderr)
	}

	// Cleanup.
	runWiz(t, bin, repo, "delete", "new-branch", "--force")
}

func TestGitPassThrough(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	runWiz(t, bin, repo, "create", "git-test")

	// Get the context path.
	ctxPath, _, _ := runWiz(t, bin, repo, "path", "git-test")
	ctxPath = strings.TrimSpace(ctxPath)

	// wiz git status (with WIZ_CTX set).
	stdout, _, err := runWizEnv(t, bin, repo, []string{"WIZ_CTX=git-test"}, "git", "status", "--short")
	if err != nil {
		t.Fatalf("wiz git status: %v\nout: %s", err, stdout)
	}

	// wiz git log --oneline.
	stdout, _, err = runWizEnv(t, bin, repo, []string{"WIZ_CTX=git-test"}, "git", "log", "--oneline")
	if err != nil {
		t.Fatalf("wiz git log: %v\nout: %s", err, stdout)
	}
	if !strings.Contains(stdout, "init") {
		t.Errorf("git log should show init commit: %s", stdout)
	}

	// wiz git rev-parse --abbrev-ref HEAD (verify branch).
	stdout, _, err = runWizEnv(t, bin, repo, []string{"WIZ_CTX=git-test"}, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		t.Fatalf("wiz git rev-parse: %v", err)
	}
	if strings.TrimSpace(stdout) != "git-test" {
		t.Errorf("wiz git branch = %q, want git-test", strings.TrimSpace(stdout))
	}

	// wiz git with --ctx flag.
	stdout, _, err = runWiz(t, bin, repo, "git", "--ctx", "git-test", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		t.Fatalf("wiz git --ctx: %v", err)
	}
	if strings.TrimSpace(stdout) != "git-test" {
		t.Errorf("wiz git --ctx branch = %q, want git-test", strings.TrimSpace(stdout))
	}

	// Cleanup.
	runWiz(t, bin, repo, "delete", "git-test", "--force")
}

func TestGitNoContextFails(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	// wiz git with no context should fail.
	_, _, err := runWiz(t, bin, repo, "git", "status")
	if err == nil {
		t.Fatal("expected error when no context is active")
	}
}

func TestGitBaseOK(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	// wiz git --base-ok should run in base repo.
	stdout, _, err := runWiz(t, bin, repo, "git", "--base-ok", "rev-parse", "--show-toplevel")
	if err != nil {
		t.Fatalf("wiz git --base-ok: %v", err)
	}
	got := strings.TrimSpace(stdout)
	if got != repo {
		t.Errorf("wiz git --base-ok dir = %q, want %q", got, repo)
	}
}

func TestAddAndCommit(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	runWiz(t, bin, repo, "create", "add-test")

	// Get context path.
	ctxPath, _, _ := runWiz(t, bin, repo, "path", "add-test")
	ctxPath = strings.TrimSpace(ctxPath)

	// Create a new file in the context.
	os.WriteFile(filepath.Join(ctxPath, "new.txt"), []byte("hello"), 0o644)

	// wiz add -A (with WIZ_CTX set).
	env := []string{"WIZ_CTX=add-test"}
	_, _, err := runWizEnv(t, bin, repo, env, "add", "-A")
	if err != nil {
		t.Fatalf("wiz add: %v", err)
	}

	// wiz commit -m "test".
	stdout, stderr, err := runWizEnv(t, bin, repo, env, "commit", "-m", "add new file")
	if err != nil {
		t.Fatalf("wiz commit: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	// Verify commit via wiz git log.
	stdout, _, err = runWizEnv(t, bin, repo, env, "git", "log", "--oneline")
	if err != nil {
		t.Fatalf("wiz git log: %v", err)
	}
	if !strings.Contains(stdout, "add new file") {
		t.Errorf("commit not in log: %s", stdout)
	}

	// Cleanup.
	runWiz(t, bin, repo, "delete", "add-test", "--force")
}

func TestAddRefusesBaseWorktree(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	// wiz add with no context should fail.
	_, _, err := runWiz(t, bin, repo, "add", "-A")
	if err == nil {
		t.Fatal("expected error when running wiz add outside a context")
	}
}

func TestCommitRefusesBaseWorktree(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	// wiz commit with no context should fail.
	_, _, err := runWiz(t, bin, repo, "commit", "-m", "nope")
	if err == nil {
		t.Fatal("expected error when running wiz commit outside a context")
	}
}

func TestGitStyleFullLifecycle(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	// 1. wiz checkout feat-x — creates context.
	stdout, stderr, err := runWiz(t, bin, repo, "checkout", "feat-x")
	if err != nil {
		t.Fatalf("checkout: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	// Get the context path from the cd command.
	ctxPath, _, _ := runWiz(t, bin, repo, "path", "feat-x")
	ctxPath = strings.TrimSpace(ctxPath)

	// 2. Modify a file in the context dir.
	os.WriteFile(filepath.Join(ctxPath, "feature.txt"), []byte("new feature"), 0o644)

	// 3. wiz add -A.
	env := []string{"WIZ_CTX=feat-x"}
	_, _, err = runWizEnv(t, bin, repo, env, "add", "-A")
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	// 4. wiz commit -m "x".
	_, _, err = runWizEnv(t, bin, repo, env, "commit", "-m", "add feature x")
	if err != nil {
		t.Fatalf("commit: %v", err)
	}

	// 5. wiz git log --oneline shows commit.
	stdout, _, err = runWizEnv(t, bin, repo, env, "git", "log", "--oneline")
	if err != nil {
		t.Fatalf("git log: %v", err)
	}
	if !strings.Contains(stdout, "add feature x") {
		t.Errorf("log missing commit: %s", stdout)
	}

	// 6. wiz finish (with test mode — no gh required).
	env = append(env, "WIZ_TEST_NO_GH=1")
	stdout, _, err = runWizEnv(t, bin, repo, env, "finish")
	if err != nil {
		t.Fatalf("finish: %v\n%s", err, stdout)
	}
	if !strings.Contains(stdout, "Finished") {
		t.Errorf("finish output: %s", stdout)
	}

	// 7. Context should be gone.
	stdout, _, _ = runWiz(t, bin, repo, "list")
	if strings.Contains(stdout, "feat-x") {
		t.Errorf("context should be cleaned up: %s", stdout)
	}
}

func TestFinishDefaultsToCurrentContext(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	runWiz(t, bin, repo, "create", "finish-me")

	// finish with WIZ_CTX set (no positional arg).
	env := []string{"WIZ_CTX=finish-me", "WIZ_TEST_NO_GH=1"}
	stdout, _, err := runWizEnv(t, bin, repo, env, "finish")
	if err != nil {
		t.Fatalf("finish: %v\n%s", err, stdout)
	}
	if !strings.Contains(stdout, "Finished: finish-me") {
		t.Errorf("finish output: %s", stdout)
	}
}

func TestGitPreservesComplexArgs(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	runWiz(t, bin, repo, "create", "args-test")

	env := []string{"WIZ_CTX=args-test"}

	// Test that complex git flags are preserved.
	stdout, _, err := runWizEnv(t, bin, repo, env, "git", "log", "--oneline", "--decorate", "--all", "-1")
	if err != nil {
		t.Fatalf("wiz git log with flags: %v", err)
	}
	// Should contain some log output (at least the init commit).
	if len(strings.TrimSpace(stdout)) == 0 {
		t.Error("expected log output")
	}

	// Test git config --local.
	_, _, err = runWizEnv(t, bin, repo, env, "git", "config", "--local", "user.name", "wiz-test-user")
	if err != nil {
		t.Fatalf("wiz git config: %v", err)
	}
	stdout, _, _ = runWizEnv(t, bin, repo, env, "git", "config", "--local", "user.name")
	if strings.TrimSpace(stdout) != "wiz-test-user" {
		t.Errorf("config value = %q", strings.TrimSpace(stdout))
	}

	runWiz(t, bin, repo, "delete", "args-test", "--force")
}

func TestAddWithCtxFlag(t *testing.T) {
	bin := buildWiz(t)
	repo := setupTestRepo(t)

	runWiz(t, bin, repo, "create", "ctx-flag-test")

	ctxPath, _, _ := runWiz(t, bin, repo, "path", "ctx-flag-test")
	ctxPath = strings.TrimSpace(ctxPath)
	os.WriteFile(filepath.Join(ctxPath, "file.txt"), []byte("x"), 0o644)

	// Use --ctx flag instead of WIZ_CTX env.
	_, _, err := runWiz(t, bin, repo, "add", "--ctx", "ctx-flag-test", "-A")
	if err != nil {
		t.Fatalf("wiz add --ctx: %v", err)
	}

	// Verify staged.
	stdout, _, _ := runWiz(t, bin, repo, "git", "--ctx", "ctx-flag-test", "diff", "--cached", "--name-only")
	if !strings.Contains(stdout, "file.txt") {
		t.Errorf("file not staged: %s", stdout)
	}

	runWiz(t, bin, repo, "delete", "ctx-flag-test", "--force")
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
