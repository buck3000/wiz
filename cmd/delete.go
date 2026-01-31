package cmd

import (
	"fmt"

	wizctx "github.com/firewood-buck-3000/wiz/internal/context"
	"github.com/firewood-buck-3000/wiz/internal/gitx"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a context",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		all, _ := cmd.Flags().GetBool("all")

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
			if len(contexts) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No contexts to delete.")
				return nil
			}
			var anyErr bool
			for _, ctx := range contexts {
				if err := deleteContext(cmd, store, repo, ctx.Name, force); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "FAIL %s: %v\n", ctx.Name, err)
					anyErr = true
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "\U0001f9d9 Deleted context: %s\n", ctx.Name)
				}
			}
			if anyErr {
				return fmt.Errorf("some deletes failed; use --force to override")
			}
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("usage: wiz delete <name> or wiz delete --all")
		}
		name := args[0]

		if err := deleteContext(cmd, store, repo, name, force); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\U0001f9d9 Deleted context: %s\n", name)
		return nil
	},
}

func deleteContext(cmd *cobra.Command, store *wizctx.Store, repo *gitx.Repo, name string, force bool) error {
	ctx, err := store.Get(name)
	if err != nil {
		return fmt.Errorf("context %q not found; run 'wiz list' to see available contexts", name)
	}

	if !force {
		st, err := gitx.StatusAt(cmd.Context(), ctx.Path)
		if err == nil && st.Dirty {
			return fmt.Errorf("context %q has uncommitted changes; use --force to delete anyway", name)
		}
	}

	prov := wizctx.NewProvisioner(ctx.Strategy, repo)
	if err := prov.Destroy(cmd.Context(), ctx.Path, force); err != nil {
		return fmt.Errorf("destroy context: %w", err)
	}

	return store.Remove(cmd.Context(), name)
}

func init() {
	deleteCmd.Flags().Bool("force", false, "Force delete even with uncommitted changes")
	deleteCmd.Flags().Bool("all", false, "Delete all contexts")
	rootCmd.AddCommand(deleteCmd)
}
