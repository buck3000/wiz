package cmd

import (
	"fmt"

	wizctx "github.com/buck3000/wiz/internal/context"
	"github.com/buck3000/wiz/internal/gitx"
	"github.com/spf13/cobra"
)

var pathCmd = &cobra.Command{
	Use:   "path <name>",
	Short: "Print the filesystem path for a context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := gitx.Discover(".")
		if err != nil {
			return err
		}
		store := wizctx.NewStore(repo)
		ctx, err := store.Get(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), ctx.Path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pathCmd)
}
