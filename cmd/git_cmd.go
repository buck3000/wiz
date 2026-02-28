package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/buck3000/wiz/internal/gitx"
	"github.com/buck3000/wiz/internal/resolve"
	"github.com/spf13/cobra"
)

var gitCmd = &cobra.Command{
	Use:                "git [flags] <git-args...>",
	Short:              "Run git in the active context directory",
	Long:               "Transparent pass-through to git, executed in the resolved context directory.\nAll arguments after 'wiz git' are forwarded to git exactly as-is.",
	DisableFlagParsing: true,
	SilenceUsage:       true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse our own flags (--ctx, --base-ok) from the front of args,
		// leaving everything else for git.
		var ctxName string
		var baseOK bool
		gitArgs := parseGitFlags(args, &ctxName, &baseOK)

		repo, err := gitx.Discover(".")
		if err != nil {
			return err
		}

		res, err := resolve.Context(repo, resolve.Options{
			ExplicitName: ctxName,
			AllowBase:    baseOK,
		})
		if err != nil {
			if baseOK {
				// Run in the base worktree.
				return runGitInDir(cmd, repo.WorkDir, gitArgs)
			}
			return err
		}

		dir := res.Context.Path

		// Warn if user runs "worktree list" through wiz.
		if len(gitArgs) >= 2 && gitArgs[0] == "worktree" && gitArgs[1] == "list" {
			fmt.Fprintln(cmd.ErrOrStderr(), "Note: wiz manages worktrees; use 'wiz list' for context info")
		}

		return runGitInDir(cmd, dir, gitArgs)
	},
}

// parseGitFlags extracts --ctx and --base-ok from the beginning of args.
// Everything else is returned as git arguments, preserving order exactly.
func parseGitFlags(args []string, ctxName *string, baseOK *bool) []string {
	var gitArgs []string
	i := 0
	for i < len(args) {
		switch {
		case args[i] == "--ctx" && i+1 < len(args):
			*ctxName = args[i+1]
			i += 2
		case len(args[i]) > 6 && args[i][:6] == "--ctx=":
			*ctxName = args[i][6:]
			i++
		case args[i] == "--base-ok":
			*baseOK = true
			i++
		default:
			gitArgs = append(gitArgs, args[i])
			i++
		}
	}
	return gitArgs
}

func runGitInDir(cmd *cobra.Command, dir string, gitArgs []string) error {
	c := exec.CommandContext(cmd.Context(), "git", gitArgs...)
	c.Dir = dir
	c.Stdin = os.Stdin
	c.Stdout = cmd.OutOrStdout()
	c.Stderr = cmd.ErrOrStderr()
	c.Env = os.Environ()
	return c.Run()
}

func init() {
	rootCmd.AddCommand(gitCmd)
}
