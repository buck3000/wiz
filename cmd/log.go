package cmd

import (
	"fmt"
	"os"
	"os/exec"

	wizctx "github.com/buck3000/wiz/internal/context"
	"github.com/buck3000/wiz/internal/gitx"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log <name>",
	Short: "Show git log for a context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		n, _ := cmd.Flags().GetInt("number")

		repo, err := gitx.Discover(".")
		if err != nil {
			return err
		}

		store := wizctx.NewStore(repo)
		ctx, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("context %q not found; run 'wiz list' to see available contexts", name)
		}

		gitArgs := []string{"log", "--oneline", fmt.Sprintf("-%d", n)}

		// Show commits since base branch if available.
		if ctx.BaseBranch != "" {
			gitArgs = append(gitArgs, fmt.Sprintf("%s..%s", ctx.BaseBranch, ctx.Branch))
		}

		c := exec.CommandContext(cmd.Context(), "git", gitArgs...)
		c.Dir = ctx.Path
		c.Stdout = cmd.OutOrStdout()
		c.Stderr = cmd.OutOrStderr()
		return c.Run()
	},
}

var logAllCmd = &cobra.Command{
	Use:   "log --all",
	Short: "Show commit summary for all contexts",
}

func init() {
	logCmd.Flags().IntP("number", "n", 10, "Number of commits to show")

	logCmd.RunE = func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		n, _ := cmd.Flags().GetInt("number")

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
				gitArgs := []string{"log", "--oneline", fmt.Sprintf("-%d", n)}
				if ctx.BaseBranch != "" {
					gitArgs = append(gitArgs, fmt.Sprintf("%s..%s", ctx.BaseBranch, ctx.Branch))
				}
				c := exec.CommandContext(cmd.Context(), "git", gitArgs...)
				c.Dir = ctx.Path
				c.Stdout = cmd.OutOrStdout()
				c.Stderr = os.Stderr
				c.Run()
				fmt.Fprintln(cmd.OutOrStdout())
			}
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("usage: wiz log <name> or wiz log --all")
		}

		name := args[0]
		ctx, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("context %q not found; run 'wiz list' to see available contexts", name)
		}

		gitArgs := []string{"log", "--oneline", fmt.Sprintf("-%d", n)}
		if ctx.BaseBranch != "" {
			gitArgs = append(gitArgs, fmt.Sprintf("%s..%s", ctx.BaseBranch, ctx.Branch))
		}

		c := exec.CommandContext(cmd.Context(), "git", gitArgs...)
		c.Dir = ctx.Path
		c.Stdout = cmd.OutOrStdout()
		c.Stderr = cmd.OutOrStderr()
		return c.Run()
	}

	logCmd.Args = cobra.ArbitraryArgs
	logCmd.Flags().Bool("all", false, "Show log for all contexts")
	rootCmd.AddCommand(logCmd)
}
