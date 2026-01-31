package cmd

import (
	"fmt"
	"os"

	wizctx "github.com/firewood-buck-3000/wiz/internal/context"
	"github.com/firewood-buck-3000/wiz/internal/gitx"
	"github.com/firewood-buck-3000/wiz/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wiz",
	Short: "\U0001f9d9 Magical git branch contexts",
	Long:  "wiz provides seamless branch contexts so multiple terminal windows can work concurrently on different branches.",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := gitx.Discover(".")
		if err != nil {
			cmd.Help()
			return nil
		}

		store := wizctx.NewStore(repo)
		contexts, err := store.List()
		if err != nil {
			return err
		}

		// Non-interactive check.
		fi, _ := os.Stdin.Stat()
		if fi.Mode()&os.ModeCharDevice == 0 {
			cmd.Help()
			return nil
		}

		result, err := tui.Run(contexts)
		if err != nil {
			return err
		}

		switch result.Action {
		case tui.ActionEnter:
			if result.Context != nil {
				// Print the enter command for the shell to eval.
				fmt.Fprintf(os.Stderr, "\U0001f9d9 Entering: %s\n", result.Context.Name)
				fmt.Printf("cd %q\n", result.Context.Path)
				fmt.Printf("export WIZ_CTX=%q\n", result.Context.Name)
				fmt.Printf("export WIZ_REPO=%q\n", repo.RepoName())
				fmt.Printf("export WIZ_DIR=%q\n", result.Context.Path)
				fmt.Printf("export WIZ_BRANCH=%q\n", result.Context.Branch)
			}
		case tui.ActionSpawn:
			if result.Context != nil {
				spawnCmd.RunE(cmd, []string{result.Context.Name})
			}
		case tui.ActionDelete:
			if result.Context != nil {
				deleteCmd.RunE(cmd, []string{result.Context.Name})
			}
		case tui.ActionCreate:
			fmt.Fprintln(os.Stderr, "Use: wiz create <name>")
		}

		return nil
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
