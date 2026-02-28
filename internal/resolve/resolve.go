package resolve

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	wizctx "github.com/buck3000/wiz/internal/context"
	"github.com/buck3000/wiz/internal/gitx"
)

// Result holds the resolved context and associated repo/store.
type Result struct {
	Context *wizctx.Context
	Store   *wizctx.Store
	Repo    *gitx.Repo
}

// Options controls how context resolution behaves.
type Options struct {
	// ExplicitName is set via --ctx flag.
	ExplicitName string
	// AllowBase permits running in the base worktree when true.
	AllowBase bool
}

// Context resolves which wiz context to operate on.
//
// Resolution order:
//  1. Explicit --ctx flag (Options.ExplicitName)
//  2. WIZ_CTX environment variable
//  3. Infer from CWD if inside a context path
//  4. Error with a helpful hint
func Context(repo *gitx.Repo, opts Options) (*Result, error) {
	store := wizctx.NewStore(repo)

	// 1. Explicit flag.
	if opts.ExplicitName != "" {
		ctx, err := store.Get(opts.ExplicitName)
		if err != nil {
			return nil, fmt.Errorf("context %q not found; run 'wiz list' to see available contexts", opts.ExplicitName)
		}
		return &Result{Context: ctx, Store: store, Repo: repo}, nil
	}

	// 2. WIZ_CTX env var.
	if envCtx := os.Getenv("WIZ_CTX"); envCtx != "" {
		ctx, err := store.Get(envCtx)
		if err != nil {
			return nil, fmt.Errorf("WIZ_CTX=%q but context not found; run 'wiz list' to see available contexts", envCtx)
		}
		return &Result{Context: ctx, Store: store, Repo: repo}, nil
	}

	// 3. Infer from CWD.
	cwd, err := os.Getwd()
	if err == nil {
		cwd, _ = filepath.EvalSymlinks(cwd)
		contexts, err := store.List()
		if err == nil {
			for i := range contexts {
				ctxPath, _ := filepath.EvalSymlinks(contexts[i].Path)
				if ctxPath != "" && (cwd == ctxPath || strings.HasPrefix(cwd, ctxPath+string(os.PathSeparator))) {
					return &Result{Context: &contexts[i], Store: store, Repo: repo}, nil
				}
			}
		}
	}

	// 4. Not found.
	return nil, fmt.Errorf("no active context; use --ctx <name>, set WIZ_CTX, or cd into a context directory\n  Run 'wiz list' to see available contexts")
}

// LastContextFile returns the path to the file that stores the previous context name.
func LastContextFile(repo *gitx.Repo) string {
	return filepath.Join(repo.CommonDir, "wiz", "last_context")
}

// SaveLastContext writes the current context name so "wiz checkout -" can return to it.
func SaveLastContext(repo *gitx.Repo, name string) error {
	path := LastContextFile(repo)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(name), 0o644)
}

// LoadLastContext reads the previously active context name.
func LoadLastContext(repo *gitx.Repo) (string, error) {
	data, err := os.ReadFile(LastContextFile(repo))
	if err != nil {
		return "", fmt.Errorf("no previous context")
	}
	name := strings.TrimSpace(string(data))
	if name == "" {
		return "", fmt.Errorf("no previous context")
	}
	return name, nil
}
