package cmd

import (
	"fmt"

	wizctx "github.com/buck3000/wiz/internal/context"
	"github.com/buck3000/wiz/internal/gitx"
	"github.com/spf13/cobra"
)

var enterCmd = &cobra.Command{
	Use:   "enter <name>",
	Short: "Activate a context in the current shell",
	Long:  "Prints shell commands to stdout. Use with: eval \"$(wiz enter <name>)\"",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		repo, err := gitx.Discover(".")
		if err != nil {
			return err
		}

		store := wizctx.NewStore(repo)
		ctx, err := store.Get(name)
		if err != nil {
			return err
		}

		repoName := repo.RepoName()

		// Output shell commands to be eval'd.
		fmt.Fprintf(cmd.OutOrStdout(), "cd %q\n", ctx.Path)
		fmt.Fprintf(cmd.OutOrStdout(), "export WIZ_CTX=%q\n", ctx.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "export WIZ_REPO=%q\n", repoName)
		fmt.Fprintf(cmd.OutOrStdout(), "export WIZ_DIR=%q\n", ctx.Path)
		fmt.Fprintf(cmd.OutOrStdout(), "export WIZ_BRANCH=%q\n", ctx.Branch)
		// Set terminal title.
		fmt.Fprintf(cmd.OutOrStdout(), "printf '\\033]0;\\U0001f9d9 %%s \\u2014 %%s\\007' %q %q\n", ctx.Name, repoName)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(enterCmd)
}
