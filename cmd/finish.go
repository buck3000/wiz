package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	wizctx "github.com/firewood-buck-3000/wiz/internal/context"
	"github.com/firewood-buck-3000/wiz/internal/gitx"
	"github.com/spf13/cobra"
)

var finishCmd = &cobra.Command{
	Use:   "finish <name>",
	Short: "Create a PR, optionally merge, then delete the context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		merge, _ := cmd.Flags().GetBool("merge")
		title, _ := cmd.Flags().GetString("title")
		body, _ := cmd.Flags().GetString("body")

		repo, err := gitx.Discover(".")
		if err != nil {
			return err
		}

		// Check for gh CLI before doing any work.
		if _, err := exec.LookPath("gh"); err != nil {
			return fmt.Errorf("GitHub CLI (gh) not found in PATH; install it from https://cli.github.com")
		}

		store := wizctx.NewStore(repo)
		ctx, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("context %q not found; run 'wiz list' to see available contexts", name)
		}

		if title == "" {
			title = name
		}
		if body == "" && ctx.Task != "" {
			body = ctx.Task
		}

		// Step 1: Push the branch.
		fmt.Fprintf(cmd.OutOrStdout(), "\U0001f9d9 Pushing %s...\n", ctx.Branch)
		pushCmd := exec.CommandContext(cmd.Context(), "git", "push", "-u", "origin", ctx.Branch)
		pushCmd.Dir = ctx.Path
		if out, err := pushCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git push: %w\n%s", err, out)
		}

		// Step 2: Create PR.
		fmt.Fprintf(cmd.OutOrStdout(), "\U0001f9d9 Creating PR...\n")
		ghArgs := []string{"pr", "create",
			"--title", title,
			"--head", ctx.Branch,
		}
		if ctx.BaseBranch != "" {
			ghArgs = append(ghArgs, "--base", ctx.BaseBranch)
		}
		if body != "" {
			ghArgs = append(ghArgs, "--body", body)
		}
		ghCmd := exec.CommandContext(cmd.Context(), "gh", ghArgs...)
		ghCmd.Dir = ctx.Path
		prOut, err := ghCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("gh pr create: %w\n%s", err, prOut)
		}
		prURL := strings.TrimSpace(string(prOut))
		fmt.Fprintf(cmd.OutOrStdout(), "\U0001f9d9 PR created: %s\n", prURL)

		// Step 3: Optionally merge.
		if merge {
			fmt.Fprintf(cmd.OutOrStdout(), "\U0001f9d9 Merging PR...\n")
			mergeCmd := exec.CommandContext(cmd.Context(), "gh", "pr", "merge", prURL, "--merge", "--delete-branch")
			mergeCmd.Dir = ctx.Path
			if out, err := mergeCmd.CombinedOutput(); err != nil {
				return fmt.Errorf("gh pr merge: %w\n%s", err, out)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\U0001f9d9 PR merged.\n")
		}

		// Step 4: Delete the context.
		prov := wizctx.NewProvisioner(ctx.Strategy, repo)
		if err := prov.Destroy(cmd.Context(), ctx.Path, true); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: destroy context: %v\n", err)
		}
		if err := store.Remove(cmd.Context(), name); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\U0001f9d9 Finished: %s\n", name)
		return nil
	},
}

func init() {
	finishCmd.Flags().Bool("merge", false, "Also merge the PR after creation")
	finishCmd.Flags().String("title", "", "PR title (default: context name)")
	finishCmd.Flags().String("body", "", "PR body (default: task description)")
	rootCmd.AddCommand(finishCmd)
}
