package template

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/buck3000/wiz/internal/config"
	"github.com/buck3000/wiz/internal/gitx"
)

// Template defines a reusable set of context defaults.
type Template struct {
	Name     string `json:"name"`
	Base     string `json:"base,omitempty"`
	Strategy string `json:"strategy,omitempty"`
	Agent    string `json:"agent,omitempty"`
}

// Store manages templates on disk.
type Store struct {
	path string
}

// NewStore returns a template store for the given repo.
func NewStore(repo *gitx.Repo) *Store {
	return &Store{
		path: filepath.Join(config.WizDir(repo), "templates.json"),
	}
}

func (s *Store) load() ([]Template, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var templates []Template
	return templates, json.Unmarshal(data, &templates)
}

func (s *Store) save(templates []Template) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(templates, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// List returns all templates.
func (s *Store) List() ([]Template, error) {
	return s.load()
}

// Get returns a template by name.
func (s *Store) Get(name string) (*Template, error) {
	templates, err := s.load()
	if err != nil {
		return nil, err
	}
	for i := range templates {
		if templates[i].Name == name {
			return &templates[i], nil
		}
	}
	return nil, fmt.Errorf("template %q not found", name)
}

// Save adds or updates a template.
func (s *Store) Save(t Template) error {
	templates, err := s.load()
	if err != nil {
		return err
	}
	for i := range templates {
		if templates[i].Name == t.Name {
			templates[i] = t
			return s.save(templates)
		}
	}
	templates = append(templates, t)
	return s.save(templates)
}

// Delete removes a template by name.
func (s *Store) Delete(name string) error {
	templates, err := s.load()
	if err != nil {
		return err
	}
	filtered := templates[:0]
	found := false
	for _, t := range templates {
		if t.Name == name {
			found = true
			continue
		}
		filtered = append(filtered, t)
	}
	if !found {
		return fmt.Errorf("template %q not found", name)
	}
	return s.save(filtered)
}
