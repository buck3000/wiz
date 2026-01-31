package context

import (
	gocontext "context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/buck3000/wiz/internal/config"
	"github.com/buck3000/wiz/internal/gitx"
	"github.com/buck3000/wiz/internal/lock"
)

// State is the on-disk JSON structure.
type State struct {
	Version  int       `json:"version"`
	Contexts []Context `json:"contexts"`
}

// Store manages the persistent collection of contexts for a repo.
type Store struct {
	repo     *gitx.Repo
	stateDir string
	lk       *lock.Lock
}

// NewStore creates a store for the given repo.
func NewStore(repo *gitx.Repo) *Store {
	dir := config.WizDir(repo)
	return &Store{
		repo:     repo,
		stateDir: dir,
		lk:       lock.New(config.LockFilePath(repo)),
	}
}

// List returns all contexts.
func (s *Store) List() ([]Context, error) {
	st, err := s.readState()
	if err != nil {
		return nil, err
	}
	return st.Contexts, nil
}

// Get returns the context with the given name, or an error if not found.
func (s *Store) Get(name string) (*Context, error) {
	st, err := s.readState()
	if err != nil {
		return nil, err
	}
	for i := range st.Contexts {
		if st.Contexts[i].Name == name {
			return &st.Contexts[i], nil
		}
	}
	return nil, fmt.Errorf("context %q not found", name)
}

// Add adds a context to the store. Acquires a file lock.
func (s *Store) Add(ctx gocontext.Context, c Context) error {
	return s.lk.WithLock(ctx, func() error {
		st, err := s.readState()
		if err != nil {
			return err
		}
		for _, existing := range st.Contexts {
			if existing.Name == c.Name {
				return fmt.Errorf("context %q already exists", c.Name)
			}
		}
		st.Contexts = append(st.Contexts, c)
		return s.writeState(st)
	})
}

// Remove removes a context by name. Acquires a file lock.
func (s *Store) Remove(ctx gocontext.Context, name string) error {
	return s.lk.WithLock(ctx, func() error {
		st, err := s.readState()
		if err != nil {
			return err
		}
		found := false
		filtered := st.Contexts[:0]
		for _, c := range st.Contexts {
			if c.Name == name {
				found = true
				continue
			}
			filtered = append(filtered, c)
		}
		if !found {
			return fmt.Errorf("context %q not found", name)
		}
		st.Contexts = filtered
		return s.writeState(st)
	})
}

// Rename renames a context. Acquires a file lock.
func (s *Store) Rename(ctx gocontext.Context, oldName, newName string) error {
	if err := ValidateName(newName); err != nil {
		return err
	}
	return s.lk.WithLock(ctx, func() error {
		st, err := s.readState()
		if err != nil {
			return err
		}
		for _, c := range st.Contexts {
			if c.Name == newName {
				return fmt.Errorf("context %q already exists", newName)
			}
		}
		found := false
		for i := range st.Contexts {
			if st.Contexts[i].Name == oldName {
				st.Contexts[i].Name = newName
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("context %q not found", oldName)
		}
		return s.writeState(st)
	})
}

// Update modifies a context by name using the given function. Acquires a file lock.
func (s *Store) Update(ctx gocontext.Context, name string, fn func(*Context)) error {
	return s.lk.WithLock(ctx, func() error {
		st, err := s.readState()
		if err != nil {
			return err
		}
		for i := range st.Contexts {
			if st.Contexts[i].Name == name {
				fn(&st.Contexts[i])
				return s.writeState(st)
			}
		}
		return fmt.Errorf("context %q not found", name)
	})
}

// Repo returns the underlying git repo.
func (s *Store) Repo() *gitx.Repo {
	return s.repo
}

func (s *Store) readState() (*State, error) {
	path := config.StateFile(s.repo)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{Version: 1}, nil
		}
		return nil, fmt.Errorf("read state: %w", err)
	}
	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("parse state: %w", err)
	}
	return &st, nil
}

func (s *Store) writeState(st *State) error {
	path := config.StateFile(s.repo)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	st.Version = 1
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	// Atomic write: write to temp file, then rename.
	tmp, err := os.CreateTemp(dir, ".state-*.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}
