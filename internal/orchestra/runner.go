package orchestra

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/firewood-buck-3000/wiz/internal/agent"
	wizctx "github.com/firewood-buck-3000/wiz/internal/context"
	"github.com/firewood-buck-3000/wiz/internal/gitx"
	"github.com/firewood-buck-3000/wiz/internal/spawn"
)

// Result captures the outcome of a single task execution.
type Result struct {
	Name  string
	Error error
}

// Run executes all tasks in the plan: creates contexts sequentially, then spawns agents
// respecting dependency ordering.
func Run(ctx context.Context, repo *gitx.Repo, plan *Plan, term spawn.Terminal) []Result {
	store := wizctx.NewStore(repo)
	results := make([]Result, len(plan.Tasks))

	// Build name-to-index map.
	nameIdx := make(map[string]int, len(plan.Tasks))
	for i, task := range plan.Tasks {
		nameIdx[task.Name] = i
	}

	// Phase 1: Create contexts sequentially (store file lock).
	for i, task := range plan.Tasks {
		branch := task.Branch
		if branch == "" {
			branch = task.Name
		}
		strategy := wizctx.ParseStrategy(task.Strategy)
		prov := wizctx.NewProvisioner(strategy, repo)

		path, err := prov.Create(ctx, wizctx.CreateOpts{
			Name:       task.Name,
			Branch:     branch,
			BaseBranch: task.Base,
			Repo:       repo,
		})
		if err != nil {
			results[i] = Result{Name: task.Name, Error: fmt.Errorf("create: %w", err)}
			continue
		}

		err = store.Add(ctx, wizctx.Context{
			Name:       task.Name,
			Branch:     branch,
			Path:       path,
			Strategy:   prov.Strategy(),
			CreatedAt:  time.Now(),
			BaseBranch: task.Base,
			Task:       task.Prompt,
			Agent:      task.Agent,
		})
		if err != nil {
			_ = prov.Destroy(ctx, path, true)
			results[i] = Result{Name: task.Name, Error: fmt.Errorf("store: %w", err)}
			continue
		}
	}

	// Phase 2: Spawn agents respecting depends_on ordering.
	// Each task gets a "done" channel that closes when it's spawned.
	done := make([]chan struct{}, len(plan.Tasks))
	for i := range done {
		done[i] = make(chan struct{})
	}

	var wg sync.WaitGroup
	for i, task := range plan.Tasks {
		if results[i].Error != nil {
			close(done[i]) // unblock dependents
			continue
		}
		wg.Add(1)
		go func(idx int, t TaskDef) {
			defer wg.Done()
			defer close(done[idx])

			// Wait for dependencies.
			for _, dep := range t.DependsOn {
				depIdx := nameIdx[dep]
				<-done[depIdx]
				// If the dependency failed, propagate failure.
				if results[depIdx].Error != nil {
					results[idx] = Result{Name: t.Name, Error: fmt.Errorf("dependency %q failed", dep)}
					return
				}
			}

			ag, err := agent.Resolve(repo, t.Agent)
			if err != nil {
				results[idx] = Result{Name: t.Name, Error: err}
				return
			}
			c, _ := store.Get(t.Name)
			shellCmd := ag.BuildCommand(t.Prompt)
			title := fmt.Sprintf("\U0001f9d9 %s [%s]", t.Name, t.Agent)
			if err := term.OpenTab(c.Path, shellCmd, title); err != nil {
				results[idx] = Result{Name: t.Name, Error: fmt.Errorf("spawn: %w", err)}
				return
			}
			results[idx] = Result{Name: t.Name}
		}(i, task)
	}
	wg.Wait()

	return results
}
