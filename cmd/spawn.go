package cmd

import (
	"fmt"

	"github.com/firewood-buck-3000/wiz/internal/agent"
	wizctx "github.com/firewood-buck-3000/wiz/internal/context"
	"github.com/firewood-buck-3000/wiz/internal/gitx"
	"github.com/firewood-buck-3000/wiz/internal/spawn"
	"github.com/spf13/cobra"
)

var spawnCmd = &cobra.Command{
	Use:   "spawn <name>",
	Short: "Open a new terminal window/tab in a context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		agentName, _ := cmd.Flags().GetString("agent")
		prompt, _ := cmd.Flags().GetString("prompt")

		repo, err := gitx.Discover(".")
		if err != nil {
			return err
		}

		store := wizctx.NewStore(repo)
		ctx, err := store.Get(name)
		if err != nil {
			return err
		}

		// Fall back to context's stored agent.
		if agentName == "" {
			agentName = ctx.Agent
		}

		term := spawn.Detect()
		var shellCmd string

		if agentName != "" {
			ag, err := agent.Resolve(repo, agentName)
			if err != nil {
				return err
			}
			// Fall back to context's task as prompt.
			if prompt == "" {
				prompt = ctx.Task
			}
			shellCmd = ag.BuildCommand(prompt)
		} else {
			shellCmd = fmt.Sprintf(`eval "$(wiz init ${SHELL##*/})"; eval "$(wiz enter %s)"`, name)
		}

		title := fmt.Sprintf("\U0001f9d9 %s \u2014 %s", name, repo.RepoName())
		if err := term.OpenTab(ctx.Path, shellCmd, title); err != nil {
			return fmt.Errorf("spawn: %w", err)
		}

		label := name
		if agentName != "" {
			label = fmt.Sprintf("%s [%s]", name, agentName)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\U0001f9d9 Spawned: %s (%s)\n", label, term.Name())
		return nil
	},
}

func init() {
	spawnCmd.Flags().String("agent", "", "Agent to run (claude, gemini, codex, or custom)")
	spawnCmd.Flags().String("prompt", "", "Prompt to send to the agent")
	rootCmd.AddCommand(spawnCmd)
}
