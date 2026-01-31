package cmd

import (
	"time"

	wizctx "github.com/firewood-buck-3000/wiz/internal/context"
	"github.com/firewood-buck-3000/wiz/internal/gitx"
	"github.com/firewood-buck-3000/wiz/internal/tui"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Live dashboard of all contexts",
	RunE: func(cmd *cobra.Command, args []string) error {
		interval, _ := cmd.Flags().GetDuration("interval")

		repo, err := gitx.Discover(".")
		if err != nil {
			return err
		}

		store := wizctx.NewStore(repo)
		return tui.RunDashboard(store, interval)
	},
}

func init() {
	watchCmd.Flags().Duration("interval", 2*time.Second, "Refresh interval")
	rootCmd.AddCommand(watchCmd)
}
