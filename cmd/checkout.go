package cmd

import (
	"fmt"
	"os"
	"time"

	wizctx "github.com/buck3000/wiz/internal/context"
	"github.com/buck3000/wiz/internal/gitx"
	"github.com/buck3000/wiz/internal/license"
	"github.com/buck3000/wiz/internal/resolve"
	"github.com/spf13/cobra"
)

var checkoutCmd = &cobra.Command{
	Use:   "checkout <name>",
	Short: "Switch to a context, creating it if needed",
	Long: `Switch to a context by name. If the context exists, enter it.
If it doesn't exist, create it and then enter it.

Use "wiz checkout -" to return to the previous context.
Use -b <name> as an alias for the positional argument.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branchFlag, _ := cmd.Flags().GetString("b")
		base, _ := cmd.Flags().GetString("base")
		strategyStr, _ := cmd.Flags().GetString("strategy")
		task, _ := cmd.Flags().GetString("task")
		agent, _ := cmd.Flags().GetString("agent")

		// Determine the target name.
		var name string
		switch {
		case len(args) == 1:
			name = args[0]
		case branchFlag != "":
			name = branchFlag
		default:
			return fmt.Errorf("usage: wiz checkout <name> or wiz checkout -b <name>")
		}

		repo, err := gitx.Discover(".")
		if err != nil {
			return err
		}

		// Handle "wiz checkout -" (previous context).
		if name == "-" {
			prevName, err := resolve.LoadLastContext(repo)
			if err != nil {
				return err
			}
			name = prevName
		}

		if err := wizctx.ValidateName(name); err != nil {
			return err
		}

		store := wizctx.NewStore(repo)

		// Save current context as "last" before switching.
		if cur := os.Getenv("WIZ_CTX"); cur != "" && cur != name {
			resolve.SaveLastContext(repo, cur)
		}

		// Try to enter existing context.
		ctx, err := store.Get(name)
		if err == nil {
			return printEnterCommands(cmd, ctx, repo)
		}

		// Context doesn't exist — create it.
		if !repo.HasCommits(cmd.Context()) {
			return fmt.Errorf("repository has no commits; create an initial commit before using wiz")
		}

		// Enforce license limit.
		tier, _ := license.CheckLicense()
		existing, err := store.List()
		if err != nil {
			return err
		}
		if err := license.CheckContextLimit(tier, len(existing)); err != nil {
			return fmt.Errorf("%w\n  Unlock unlimited contexts: https://github.com/sponsors/buck3000", err)
		}

		strategy := wizctx.ParseStrategy(strategyStr)
		prov := wizctx.NewProvisioner(strategy, repo)

		path, err := prov.Create(cmd.Context(), wizctx.CreateOpts{
			Name:       name,
			Branch:     name,
			BaseBranch: base,
			Repo:       repo,
		})
		if err != nil {
			return err
		}

		newCtx := wizctx.Context{
			Name:       name,
			Branch:     name,
			Path:       path,
			Strategy:   prov.Strategy(),
			CreatedAt:  time.Now(),
			BaseBranch: base,
			Task:       task,
			Agent:      agent,
		}

		if err := store.Add(cmd.Context(), newCtx); err != nil {
			prov.Destroy(cmd.Context(), path, true)
			return err
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "\U0001f9d9 Created context: %s\n", name)
		return printEnterCommands(cmd, &newCtx, repo)
	},
}

// printEnterCommands outputs the shell eval commands for entering a context.
func printEnterCommands(cmd *cobra.Command, ctx *wizctx.Context, repo *gitx.Repo) error {
	repoName := repo.RepoName()
	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "cd %q\n", ctx.Path)
	fmt.Fprintf(w, "export WIZ_CTX=%q\n", ctx.Name)
	fmt.Fprintf(w, "export WIZ_REPO=%q\n", repoName)
	fmt.Fprintf(w, "export WIZ_DIR=%q\n", ctx.Path)
	fmt.Fprintf(w, "export WIZ_BRANCH=%q\n", ctx.Branch)
	fmt.Fprintf(w, "printf '\\033]0;\\U0001f9d9 %%s \\u2014 %%s\\007' %q %q\n", ctx.Name, repoName)
	return nil
}

func init() {
	checkoutCmd.Flags().StringP("b", "b", "", "Create a new context with this name (alias for positional arg)")
	checkoutCmd.Flags().String("base", "", "Base branch to create from (default: current HEAD)")
	checkoutCmd.Flags().String("strategy", "auto", "Strategy: auto, worktree, clone")
	checkoutCmd.Flags().String("task", "", "Task description for new context")
	checkoutCmd.Flags().String("agent", "", "Agent to associate (e.g. claude, gemini, codex)")
	rootCmd.AddCommand(checkoutCmd)
}
