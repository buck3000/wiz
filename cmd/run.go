package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/buck3000/wiz/internal/agent"
	wizctx "github.com/buck3000/wiz/internal/context"
	"github.com/buck3000/wiz/internal/gitx"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <name> -- <command...>",
	Short: "Run a command inside a context",
	Args:  cobra.MinimumNArgs(1),
	// Disable flag parsing after '--' so the command's flags aren't consumed.
	DisableFlagParsing: false,
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

		env := append(os.Environ(),
			"WIZ_CTX="+ctx.Name,
			"WIZ_REPO="+repo.RepoName(),
			"WIZ_DIR="+ctx.Path,
			"WIZ_BRANCH="+ctx.Branch,
		)

		// Agent mode: resolve agent and exec it.
		if agentName == "" {
			agentName = ctx.Agent
		}
		if agentName != "" && len(args) < 2 {
			ag, err := agent.Resolve(repo, agentName)
			if err != nil {
				return err
			}
			if prompt == "" {
				prompt = ctx.Task
			}
			bin, agentArgs := ag.BuildExecArgs(prompt)
			c := exec.CommandContext(cmd.Context(), bin, agentArgs...)
			c.Dir = ctx.Path
			c.Stdin = os.Stdin
			c.Stdout = cmd.OutOrStdout()
			c.Stderr = cmd.OutOrStderr()
			c.Env = env
			return c.Run()
		}

		// Explicit command mode.
		if len(args) < 2 {
			return fmt.Errorf("usage: wiz run <name> -- <command...> or wiz run <name> --agent <agent>")
		}
		cmdArgs := args[1:]

		c := exec.CommandContext(cmd.Context(), cmdArgs[0], cmdArgs[1:]...)
		c.Dir = ctx.Path
		c.Stdin = os.Stdin
		c.Stdout = cmd.OutOrStdout()
		c.Stderr = cmd.OutOrStderr()
		c.Env = env

		return c.Run()
	},
}

func init() {
	runCmd.Flags().String("agent", "", "Agent to run (claude, gemini, codex, or custom)")
	runCmd.Flags().String("prompt", "", "Prompt to send to the agent")
	rootCmd.AddCommand(runCmd)
}
