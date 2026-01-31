package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/buck3000/wiz/internal/gitx"
	"github.com/buck3000/wiz/internal/template"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage context templates",
}

var templateSaveCmd = &cobra.Command{
	Use:   "save <name>",
	Short: "Save a template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		base, _ := cmd.Flags().GetString("base")
		agent, _ := cmd.Flags().GetString("agent")
		strategy, _ := cmd.Flags().GetString("strategy")

		repo, err := gitx.Discover(".")
		if err != nil {
			return err
		}

		store := template.NewStore(repo)
		t := template.Template{
			Name:     name,
			Base:     base,
			Strategy: strategy,
			Agent:    agent,
		}
		if err := store.Save(t); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\U0001f9d9 Template saved: %s\n", name)
		return nil
	},
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := gitx.Discover(".")
		if err != nil {
			return err
		}

		store := template.NewStore(repo)
		templates, err := store.List()
		if err != nil {
			return err
		}

		asJSON, _ := cmd.Flags().GetBool("json")
		if asJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(templates)
		}

		if len(templates) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No templates.")
			return nil
		}
		for _, t := range templates {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s", t.Name)
			if t.Base != "" {
				fmt.Fprintf(cmd.OutOrStdout(), " (base: %s)", t.Base)
			}
			if t.Agent != "" {
				fmt.Fprintf(cmd.OutOrStdout(), " [%s]", t.Agent)
			}
			if t.Strategy != "" {
				fmt.Fprintf(cmd.OutOrStdout(), " (%s)", t.Strategy)
			}
			fmt.Fprintln(cmd.OutOrStdout())
		}
		return nil
	},
}

var templateDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := gitx.Discover(".")
		if err != nil {
			return err
		}
		store := template.NewStore(repo)
		if err := store.Delete(args[0]); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\U0001f9d9 Template deleted: %s\n", args[0])
		return nil
	},
}

func init() {
	templateSaveCmd.Flags().String("base", "", "Default base branch")
	templateSaveCmd.Flags().String("agent", "", "Default agent")
	templateSaveCmd.Flags().String("strategy", "", "Default strategy")

	templateListCmd.Flags().Bool("json", false, "Output as JSON")

	templateCmd.AddCommand(templateSaveCmd)
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateDeleteCmd)
	rootCmd.AddCommand(templateCmd)
}
