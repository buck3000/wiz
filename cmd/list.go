package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	wizctx "github.com/firewood-buck-3000/wiz/internal/context"
	"github.com/firewood-buck-3000/wiz/internal/gitx"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all contexts",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := gitx.Discover(".")
		if err != nil {
			return err
		}

		store := wizctx.NewStore(repo)
		contexts, err := store.List()
		if err != nil {
			return err
		}

		asJSON, _ := cmd.Flags().GetBool("json")
		if asJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(contexts)
		}

		if len(contexts) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No contexts. Create one with: wiz create <name>")
			return nil
		}

		showTasks, _ := cmd.Flags().GetBool("tasks")
		current := os.Getenv("WIZ_CTX")
		for _, c := range contexts {
			marker := "  "
			if c.Name == current {
				marker = "\u25b8 "
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\033[1;35m%s\033[0m (%s)\n", marker, c.Name, c.Strategy)
			fmt.Fprintf(cmd.OutOrStdout(), "    branch: %s\n", c.Branch)
			fmt.Fprintf(cmd.OutOrStdout(), "    path:   %s\n", c.Path)
			if showTasks {
				if c.Task != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "    task:   %s\n", c.Task)
				}
				if c.Agent != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "    agent:  %s\n", c.Agent)
				}
			}
		}
		return nil
	},
}

func init() {
	listCmd.Flags().Bool("json", false, "Output as JSON")
	listCmd.Flags().Bool("tasks", false, "Show task and agent info")
	rootCmd.AddCommand(listCmd)
}
