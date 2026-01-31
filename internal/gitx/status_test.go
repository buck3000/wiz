package gitx_test

import (
	"context"
	"testing"

	"github.com/buck3000/wiz/internal/gitx"
	"github.com/buck3000/wiz/testutil"
)

func TestParseStatusPorcelainV2_Clean(t *testing.T) {
	input := `# branch.oid abc123
# branch.head main
# branch.upstream origin/main
# branch.ab +0 -0
`
	s, err := gitx.ParseStatusPorcelainV2(input)
	if err != nil {
		t.Fatal(err)
	}
	if s.Branch != "main" {
		t.Errorf("Branch = %q, want main", s.Branch)
	}
	if s.OID != "abc123" {
		t.Errorf("OID = %q, want abc123", s.OID)
	}
	if s.Upstream != "origin/main" {
		t.Errorf("Upstream = %q, want origin/main", s.Upstream)
	}
	if s.Ahead != 0 || s.Behind != 0 {
		t.Errorf("Ahead=%d Behind=%d, want 0 0", s.Ahead, s.Behind)
	}
	if s.Dirty {
		t.Error("Dirty = true, want false")
	}
}

func TestParseStatusPorcelainV2_Dirty(t *testing.T) {
	input := `# branch.oid abc123
# branch.head feature/foo
# branch.ab +2 -1
1 .M N... 100644 100644 100644 hash1 hash2 modified.go
1 A. N... 000000 100644 100644 hash1 hash2 new.go
? untracked.txt
`
	s, err := gitx.ParseStatusPorcelainV2(input)
	if err != nil {
		t.Fatal(err)
	}
	if s.Branch != "feature/foo" {
		t.Errorf("Branch = %q", s.Branch)
	}
	if s.Ahead != 2 || s.Behind != 1 {
		t.Errorf("Ahead=%d Behind=%d, want 2 1", s.Ahead, s.Behind)
	}
	if !s.Dirty {
		t.Error("Dirty = false, want true")
	}
	if s.Staged != 1 {
		t.Errorf("Staged = %d, want 1", s.Staged)
	}
	if s.Unstaged != 1 {
		t.Errorf("Unstaged = %d, want 1", s.Unstaged)
	}
	if s.Untracked != 1 {
		t.Errorf("Untracked = %d, want 1", s.Untracked)
	}
}

func TestParseStatusPorcelainV2_Conflicts(t *testing.T) {
	input := `# branch.oid abc123
# branch.head main
u UU N... 100644 100644 100644 100644 hash1 hash2 hash3 conflicted.go
`
	s, err := gitx.ParseStatusPorcelainV2(input)
	if err != nil {
		t.Fatal(err)
	}
	if s.Conflicted != 1 {
		t.Errorf("Conflicted = %d, want 1", s.Conflicted)
	}
	if !s.Dirty {
		t.Error("Dirty = false, want true")
	}
}

func TestStatus_LiveRepo(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, _ := gitx.Discover(tr.Dir)

	s, err := repo.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if s.Dirty {
		t.Error("clean repo reported as dirty")
	}
	if s.Branch == "" {
		t.Error("branch is empty")
	}

	// Make it dirty.
	tr.AddFile("dirty.txt", "uncommitted")
	s, err = repo.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !s.Dirty {
		t.Error("dirty repo reported as clean")
	}
	if s.Untracked != 1 {
		t.Errorf("Untracked = %d, want 1", s.Untracked)
	}
}
