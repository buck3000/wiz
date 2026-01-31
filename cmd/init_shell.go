package cmd

import (
	"fmt"

	"github.com/firewood-buck-3000/wiz/internal/shell"
	"github.com/spf13/cobra"
)

var initShellCmd = &cobra.Command{
	Use:   "init <bash|zsh|fish>",
	Short: "Print shell integration script",
	Long:  "Print shell integration to be eval'd in your shell rc file.\nAdd to your .zshrc: eval \"$(wiz init zsh)\"",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		script, err := shell.InitScript(args[0])
		if err != nil {
			return err
		}
		fmt.Fprint(cmd.OutOrStdout(), script)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initShellCmd)
}
