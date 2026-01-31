package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	wizctx "github.com/buck3000/wiz/internal/context"
	"github.com/buck3000/wiz/internal/gitx"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff <name>",
	Short: "Show diff for a context vs its base branch",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		stat, _ := cmd.Flags().GetBool("stat")

		repo, err := gitx.Discover(".")
		if err != nil {
			return err
		}

		store := wizctx.NewStore(repo)

		if all {
			contexts, err := store.List()
			if err != nil {
				return err
			}
			for _, ctx := range contexts {
				fmt.Fprintf(cmd.OutOrStdout(), "\033[1;35m%s\033[0m (%s)\n", ctx.Name, ctx.Branch)
				printDiffStat(cmd, &ctx, repo)
				fmt.Fprintln(cmd.OutOrStdout())
			}
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("usage: wiz diff <name> or wiz diff --all")
		}

		name := args[0]
		ctx, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("context %q not found; run 'wiz list' to see available contexts", name)
		}

		gitArgs := []string{"diff"}
		if stat {
			gitArgs = append(gitArgs, "--stat")
		}
		if ctx.BaseBranch != "" {
			gitArgs = append(gitArgs, fmt.Sprintf("%s...%s", ctx.BaseBranch, ctx.Branch))
		}

		c := exec.CommandContext(cmd.Context(), "git", gitArgs...)
		c.Dir = ctx.Path
		c.Stdout = cmd.OutOrStdout()
		c.Stderr = cmd.OutOrStderr()
		return c.Run()
	},
}

func printDiffStat(cmd *cobra.Command, ctx *wizctx.Context, repo *gitx.Repo) {
	gitArgs := []string{"diff", "--stat"}
	if ctx.BaseBranch != "" {
		gitArgs = append(gitArgs, fmt.Sprintf("%s...%s", ctx.BaseBranch, ctx.Branch))
	}

	c := exec.CommandContext(cmd.Context(), "git", gitArgs...)
	c.Dir = ctx.Path
	out, err := c.Output()
	if err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "  (no diff available)\n")
		return
	}

	output := strings.TrimSpace(string(out))
	if output == "" {
		fmt.Fprintf(cmd.OutOrStdout(), "  no changes\n")
		return
	}

	// Indent each line.
	for _, line := range strings.Split(output, "\n") {
		fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", line)
	}
	_ = os.Stderr // suppress unused import
}

func init() {
	diffCmd.Flags().Bool("all", false, "Show diff summary for all contexts")
	diffCmd.Flags().Bool("stat", false, "Show diffstat instead of full diff")
	rootCmd.AddCommand(diffCmd)
}
