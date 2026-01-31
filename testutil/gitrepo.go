package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestRepo is a temporary git repository for testing.
type TestRepo struct {
	Dir string
	t   testing.TB
}

// NewTestRepo creates a fresh git repo in a temp directory with one initial commit.
func NewTestRepo(t testing.TB) *TestRepo {
	t.Helper()
	dir := t.TempDir()
	// Resolve symlinks (macOS /var -> /private/var).
	dir, _ = filepath.EvalSymlinks(dir)

	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@wiz.dev")
	run(t, dir, "git", "config", "user.name", "wiz-test")

	// Create initial commit so HEAD exists.
	writeFile(t, filepath.Join(dir, "README.md"), "# test repo\n")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "initial commit")

	return &TestRepo{Dir: dir, t: t}
}

// AddFile writes content to a file relative to the repo root.
func (r *TestRepo) AddFile(rel, content string) {
	r.t.Helper()
	writeFile(r.t, filepath.Join(r.Dir, rel), content)
}

// Commit stages all changes and commits with the given message.
func (r *TestRepo) Commit(msg string) {
	r.t.Helper()
	run(r.t, r.Dir, "git", "add", ".")
	run(r.t, r.Dir, "git", "commit", "-m", msg)
}

// CreateBranch creates a new branch at the current HEAD.
func (r *TestRepo) CreateBranch(name string) {
	r.t.Helper()
	run(r.t, r.Dir, "git", "branch", name)
}

// Checkout switches to the given branch.
func (r *TestRepo) Checkout(name string) {
	r.t.Helper()
	run(r.t, r.Dir, "git", "checkout", name)
}

// GitDir returns the path to the .git directory.
func (r *TestRepo) GitDir() string {
	return filepath.Join(r.Dir, ".git")
}

// CurrentBranch returns the current branch name.
func (r *TestRepo) CurrentBranch() string {
	r.t.Helper()
	out := runOutput(r.t, r.Dir, "git", "rev-parse", "--abbrev-ref", "HEAD")
	return out
}

func run(t testing.TB, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_DATE=2024-01-01T00:00:00+00:00",
		"GIT_COMMITTER_DATE=2024-01-01T00:00:00+00:00",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
}

func runOutput(t testing.TB, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s %v failed: %v", name, args, err)
	}
	// trim trailing newline
	s := string(out)
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	return s
}

func writeFile(t testing.TB, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
