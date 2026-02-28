package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/buck3000/wiz/internal/gitx"
	"github.com/buck3000/wiz/internal/resolve"
	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:                "commit [flags] [args...]",
	Short:              "Commit in the active context (delegates to git commit)",
	Long:               "Pass-through to 'git commit' executed in the resolved context directory.\nAll arguments are forwarded to git commit exactly as-is.\nRefuses to run in the base worktree for safety.",
	DisableFlagParsing: true,
	SilenceUsage:       true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Extract --ctx from our args; never allow --base-ok for commit.
		var ctxName string
		var ignored bool
		gitCommitArgs := parseGitFlags(args, &ctxName, &ignored)

		repo, err := gitx.Discover(".")
		if err != nil {
			return err
		}

		res, err := resolve.Context(repo, resolve.Options{
			ExplicitName: ctxName,
		})
		if err != nil {
			return fmt.Errorf("wiz commit: %w\n  wiz commit never runs in the base worktree for safety", err)
		}

		// Delegate to git commit in the context directory.
		fullArgs := append([]string{"commit"}, gitCommitArgs...)
		c := exec.CommandContext(cmd.Context(), "git", fullArgs...)
		c.Dir = res.Context.Path
		c.Stdin = os.Stdin
		c.Stdout = cmd.OutOrStdout()
		c.Stderr = cmd.ErrOrStderr()
		c.Env = os.Environ()
		return c.Run()
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
}
