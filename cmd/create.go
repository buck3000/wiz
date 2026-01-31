package cmd

import (
	"fmt"
	"time"

	wizctx "github.com/buck3000/wiz/internal/context"
	"github.com/buck3000/wiz/internal/gitx"
	"github.com/buck3000/wiz/internal/license"
	"github.com/buck3000/wiz/internal/template"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := wizctx.ValidateName(name); err != nil {
			return err
		}

		base, _ := cmd.Flags().GetString("base")
		strategyStr, _ := cmd.Flags().GetString("strategy")
		task, _ := cmd.Flags().GetString("task")
		agent, _ := cmd.Flags().GetString("agent")
		tmplName, _ := cmd.Flags().GetString("template")

		repo, err := gitx.Discover(".")
		if err != nil {
			return err
		}

		// Apply template defaults (explicit flags override).
		if tmplName != "" {
			tmplStore := template.NewStore(repo)
			tmpl, err := tmplStore.Get(tmplName)
			if err != nil {
				return err
			}
			if base == "" {
				base = tmpl.Base
			}
			if strategyStr == "" || strategyStr == "auto" {
				if tmpl.Strategy != "" {
					strategyStr = tmpl.Strategy
				}
			}
			if agent == "" {
				agent = tmpl.Agent
			}
		}

		if !repo.HasCommits(cmd.Context()) {
			return fmt.Errorf("repository has no commits; create an initial commit before using wiz")
		}

		store := wizctx.NewStore(repo)

		// Enforce context limit based on license tier.
		tier, _ := license.CheckLicense()
		existing, err := store.List()
		if err != nil {
			return err
		}
		if err := license.CheckContextLimit(tier, len(existing)); err != nil {
			limErr := err.(*license.ContextLimitErr)
			return fmt.Errorf("%s\n\n  Upgrade to Wiz Pro for unlimited contexts:\n    https://wiz.dev/pro\n\n  Or free up a slot:\n    wiz delete <name>\n    wiz gc --merged\n\n  Current contexts: %d/%d",
				limErr, limErr.Current, limErr.Max)
		}

		strategy := wizctx.ParseStrategy(strategyStr)
		prov := wizctx.NewProvisioner(strategy, repo)

		branch := name
		path, err := prov.Create(cmd.Context(), wizctx.CreateOpts{
			Name:       name,
			Branch:     branch,
			BaseBranch: base,
			Repo:       repo,
		})
		if err != nil {
			return err
		}

		err = store.Add(cmd.Context(), wizctx.Context{
			Name:       name,
			Branch:     branch,
			Path:       path,
			Strategy:   prov.Strategy(),
			CreatedAt:  time.Now(),
			BaseBranch: base,
			Task:       task,
			Agent:      agent,
		})
		if err != nil {
			// Clean up on store failure.
			prov.Destroy(cmd.Context(), path, true)
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\U0001f9d9 Created context: %s\n", name)
		return nil
	},
}

func init() {
	createCmd.Flags().String("base", "", "Base branch (default: current HEAD)")
	createCmd.Flags().String("strategy", "auto", "Strategy: auto, worktree, clone")
	createCmd.Flags().String("task", "", "Task description for this context")
	createCmd.Flags().String("agent", "", "Agent to associate (e.g. claude, gemini, codex)")
	createCmd.Flags().String("template", "", "Apply a saved template")
	rootCmd.AddCommand(createCmd)
}
