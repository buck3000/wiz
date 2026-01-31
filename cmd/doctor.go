package cmd

import (
	"fmt"

	"github.com/firewood-buck-3000/wiz/internal/doctor"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check environment and show active enhancements",
	RunE: func(cmd *cobra.Command, args []string) error {
		results := doctor.RunAll()
		for _, r := range results {
			var icon string
			switch r.Status {
			case doctor.OK:
				icon = "\033[32m\u2713\033[0m"
			case doctor.Warn:
				icon = "\033[33m!\033[0m"
			case doctor.Fail:
				icon = "\033[31m\u2717\033[0m"
			}
			fmt.Fprintf(cmd.OutOrStdout(), " %s %s: %s\n", icon, r.Name, r.Message)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
