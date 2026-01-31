package orchestra

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// TaskDef is a single task in an orchestra file.
type TaskDef struct {
	Name      string   `yaml:"name"`
	Branch    string   `yaml:"branch,omitempty"`
	Base      string   `yaml:"base,omitempty"`
	Prompt    string   `yaml:"prompt"`
	Agent     string   `yaml:"agent"`
	Strategy  string   `yaml:"strategy,omitempty"`
	DependsOn []string `yaml:"depends_on,omitempty"`
}

// Plan is the top-level orchestra YAML structure.
type Plan struct {
	Tasks []TaskDef `yaml:"tasks"`
}

// LoadPlan reads and parses an orchestra YAML file.
func LoadPlan(path string) (*Plan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read orchestra file: %w", err)
	}
	var p Plan
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse orchestra file: %w", err)
	}
	if len(p.Tasks) == 0 {
		return nil, fmt.Errorf("orchestra file contains no tasks")
	}
	names := make(map[string]bool, len(p.Tasks))
	for i, t := range p.Tasks {
		if t.Name == "" {
			return nil, fmt.Errorf("task %d: name is required", i)
		}
		if t.Agent == "" {
			return nil, fmt.Errorf("task %d (%s): agent is required", i, t.Name)
		}
		names[t.Name] = true
	}
	// Validate depends_on references.
	for _, t := range p.Tasks {
		for _, dep := range t.DependsOn {
			if !names[dep] {
				return nil, fmt.Errorf("task %q depends on unknown task %q", t.Name, dep)
			}
		}
	}
	return &p, nil
}
