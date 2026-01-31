package cmd

import (
	"fmt"

	wizctx "github.com/firewood-buck-3000/wiz/internal/context"
	"github.com/firewood-buck-3000/wiz/internal/gitx"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename <old> <new>",
	Short: "Rename a context",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldName, newName := args[0], args[1]

		repo, err := gitx.Discover(".")
		if err != nil {
			return err
		}

		store := wizctx.NewStore(repo)
		if err := store.Rename(cmd.Context(), oldName, newName); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\U0001f9d9 Renamed: %s \u2192 %s\n", oldName, newName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)
}
