package template

import (
	"testing"

	"github.com/firewood-buck-3000/wiz/internal/gitx"
	"github.com/firewood-buck-3000/wiz/testutil"
)

func TestTemplateCRUD(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, err := gitx.Discover(tr.Dir)
	if err != nil {
		t.Fatal(err)
	}

	store := NewStore(repo)

	// List empty.
	templates, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) != 0 {
		t.Fatalf("expected 0 templates, got %d", len(templates))
	}

	// Save.
	err = store.Save(Template{Name: "bugfix", Base: "main", Agent: "claude"})
	if err != nil {
		t.Fatal(err)
	}

	// Get.
	tmpl, err := store.Get("bugfix")
	if err != nil {
		t.Fatal(err)
	}
	if tmpl.Base != "main" {
		t.Errorf("base = %q, want main", tmpl.Base)
	}
	if tmpl.Agent != "claude" {
		t.Errorf("agent = %q, want claude", tmpl.Agent)
	}

	// Overwrite.
	err = store.Save(Template{Name: "bugfix", Base: "develop", Agent: "gemini"})
	if err != nil {
		t.Fatal(err)
	}
	tmpl, _ = store.Get("bugfix")
	if tmpl.Base != "develop" {
		t.Errorf("after overwrite base = %q, want develop", tmpl.Base)
	}

	// List.
	templates, _ = store.List()
	if len(templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(templates))
	}

	// Delete.
	err = store.Delete("bugfix")
	if err != nil {
		t.Fatal(err)
	}
	templates, _ = store.List()
	if len(templates) != 0 {
		t.Fatalf("after delete expected 0, got %d", len(templates))
	}
}

func TestGetNotFound(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, err := gitx.Discover(tr.Dir)
	if err != nil {
		t.Fatal(err)
	}
	store := NewStore(repo)
	_, err = store.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent template")
	}
}

func TestDeleteNotFound(t *testing.T) {
	tr := testutil.NewTestRepo(t)
	repo, err := gitx.Discover(tr.Dir)
	if err != nil {
		t.Fatal(err)
	}
	store := NewStore(repo)
	err = store.Delete("nonexistent")
	if err == nil {
		t.Fatal("expected error for deleting nonexistent template")
	}
}
