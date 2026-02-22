package cmd

import (
	"fmt"

	"github.com/buck3000/wiz/internal/gitx"
	"github.com/buck3000/wiz/internal/orchestra"
	"github.com/buck3000/wiz/internal/spawn"
	"github.com/spf13/cobra"
)

var orchestraCmd = &cobra.Command{
	Use:   "orchestra <file.yaml>",
	Short: "Create contexts and spawn agents from a task file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		planFile := args[0]

		plan, err := orchestra.LoadPlan(planFile)
		if err != nil {
			return err
		}

		repo, err := gitx.Discover(".")
		if err != nil {
			return err
		}

		term := spawn.Detect()
		results := orchestra.Run(cmd.Context(), repo, plan, term)

		var anyErr bool
		for _, r := range results {
			if r.Error != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "FAIL %s: %v\n", r.Name, r.Error)
				anyErr = true
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "\U0001f9d9 %s: spawned\n", r.Name)
			}
		}
		if anyErr {
			return fmt.Errorf("some tasks failed")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(orchestraCmd)
}
