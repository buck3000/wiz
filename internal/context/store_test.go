package context_test

import (
	gocontext "context"
	"sync"
	"testing"
	"time"

	wizctx "github.com/buck3000/wiz/internal/context"
	"github.com/buck3000/wiz/internal/gitx"
	"github.com/buck3000/wiz/testutil"
)

func setupStore(t *testing.T) (*wizctx.Store, *gitx.Repo) {
	t.Helper()
	tr := testutil.NewTestRepo(t)
	repo, err := gitx.Discover(tr.Dir)
	if err != nil {
		t.Fatal(err)
	}
	return wizctx.NewStore(repo), repo
}

func TestStoreAddAndList(t *testing.T) {
	store, _ := setupStore(t)
	ctx := gocontext.Background()

	err := store.Add(ctx, wizctx.Context{
		Name:      "feat-x",
		Branch:    "feat-x",
		Path:      "/tmp/fake",
		Strategy:  wizctx.StrategyWorktree,
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	list, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("len = %d, want 1", len(list))
	}
	if list[0].Name != "feat-x" {
		t.Errorf("Name = %q", list[0].Name)
	}
}

func TestStoreGet(t *testing.T) {
	store, _ := setupStore(t)
	ctx := gocontext.Background()

	store.Add(ctx, wizctx.Context{Name: "alpha", Branch: "alpha", Path: "/a"})
	store.Add(ctx, wizctx.Context{Name: "beta", Branch: "beta", Path: "/b"})

	c, err := store.Get("beta")
	if err != nil {
		t.Fatal(err)
	}
	if c.Path != "/b" {
		t.Errorf("Path = %q", c.Path)
	}

	_, err = store.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent context")
	}
}

func TestStoreDuplicateName(t *testing.T) {
	store, _ := setupStore(t)
	ctx := gocontext.Background()

	store.Add(ctx, wizctx.Context{Name: "dup", Branch: "dup"})
	err := store.Add(ctx, wizctx.Context{Name: "dup", Branch: "dup2"})
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}
}

func TestStoreRemove(t *testing.T) {
	store, _ := setupStore(t)
	ctx := gocontext.Background()

	store.Add(ctx, wizctx.Context{Name: "gone", Branch: "gone"})
	err := store.Remove(ctx, "gone")
	if err != nil {
		t.Fatal(err)
	}

	list, _ := store.List()
	if len(list) != 0 {
		t.Fatalf("len = %d after remove", len(list))
	}

	err = store.Remove(ctx, "gone")
	if err == nil {
		t.Fatal("expected error removing nonexistent")
	}
}

func TestStoreRename(t *testing.T) {
	store, _ := setupStore(t)
	ctx := gocontext.Background()

	store.Add(ctx, wizctx.Context{Name: "old-name", Branch: "old-name"})
	err := store.Rename(ctx, "old-name", "new-name")
	if err != nil {
		t.Fatal(err)
	}

	c, err := store.Get("new-name")
	if err != nil {
		t.Fatal(err)
	}
	if c.Name != "new-name" {
		t.Errorf("Name = %q", c.Name)
	}

	_, err = store.Get("old-name")
	if err == nil {
		t.Fatal("old name should not exist")
	}
}

func TestStoreRenameDuplicate(t *testing.T) {
	store, _ := setupStore(t)
	ctx := gocontext.Background()

	store.Add(ctx, wizctx.Context{Name: "a"})
	store.Add(ctx, wizctx.Context{Name: "b"})

	err := store.Rename(ctx, "a", "b")
	if err == nil {
		t.Fatal("expected error renaming to existing name")
	}
}

func TestStoreConcurrentAdd(t *testing.T) {
	store, _ := setupStore(t)

	var wg sync.WaitGroup
	errs := make([]error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := "ctx-" + time.Now().Format("150405.000000000") + "-" + string(rune('a'+idx))
			errs[idx] = store.Add(gocontext.Background(), wizctx.Context{
				Name:   name,
				Branch: name,
				Path:   "/tmp/" + name,
			})
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: %v", i, err)
		}
	}

	list, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 20 {
		t.Errorf("len = %d, want 20", len(list))
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"feat-x", true},
		{"feature/auth", true},
		{"v1.0", true},
		{"a", true},
		{"", false},
		{".hidden", false},
		{"-dash", false},
		{"has space", false},
	}
	for _, tc := range tests {
		err := wizctx.ValidateName(tc.name)
		if tc.valid && err != nil {
			t.Errorf("ValidateName(%q) = %v, want valid", tc.name, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("ValidateName(%q) = nil, want error", tc.name)
		}
	}
}
