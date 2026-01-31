package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/buck3000/wiz/internal/gitx"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current context status",
	RunE: func(cmd *cobra.Command, args []string) error {
		porcelain, _ := cmd.Flags().GetBool("porcelain")
		asJSON, _ := cmd.Flags().GetBool("json")

		ctxName := os.Getenv("WIZ_CTX")
		repoName := os.Getenv("WIZ_REPO")
		wizDir := os.Getenv("WIZ_DIR")

		if ctxName == "" {
			if porcelain || asJSON {
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Not in a wiz context. Use: wiz enter <name>")
			return nil
		}

		if asJSON {
			return printJSONStatus(cmd, ctxName, repoName, wizDir)
		}

		if porcelain {
			return printPorcelainStatus(cmd, ctxName, repoName, wizDir)
		}

		// Human-readable.
		fmt.Fprintf(cmd.OutOrStdout(), "\U0001f9d9 Context: %s\n", ctxName)
		fmt.Fprintf(cmd.OutOrStdout(), "   Repo:    %s\n", repoName)
		fmt.Fprintf(cmd.OutOrStdout(), "   Dir:     %s\n", wizDir)

		if wizDir != "" {
			st, err := gitx.StatusAt(cmd.Context(), wizDir)
			if err == nil {
				state := "clean"
				if st.Dirty {
					state = "dirty"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "   Branch:  %s\n", st.Branch)
				fmt.Fprintf(cmd.OutOrStdout(), "   State:   %s\n", state)
				if st.Ahead > 0 || st.Behind > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "   Ahead:   %d  Behind: %d\n", st.Ahead, st.Behind)
				}
			}
		}
		return nil
	},
}

type statusJSON struct {
	Context   string `json:"context"`
	Repo      string `json:"repo"`
	Dir       string `json:"dir"`
	Branch    string `json:"branch"`
	State     string `json:"state"`
	Ahead     int    `json:"ahead"`
	Behind    int    `json:"behind"`
	Staged    int    `json:"staged"`
	Unstaged  int    `json:"unstaged"`
	Untracked int    `json:"untracked"`
}

func printJSONStatus(cmd *cobra.Command, ctxName, repoName, dir string) error {
	s := statusJSON{
		Context: ctxName,
		Repo:    repoName,
		Dir:     dir,
		Branch:  os.Getenv("WIZ_BRANCH"),
		State:   "clean",
	}

	if dir != "" {
		st, err := gitx.StatusAt(cmd.Context(), dir)
		if err == nil {
			s.Branch = st.Branch
			if st.Dirty {
				s.State = "dirty"
			}
			s.Ahead = st.Ahead
			s.Behind = st.Behind
			s.Staged = st.Staged
			s.Unstaged = st.Unstaged
			s.Untracked = st.Untracked
		}
	}

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

func printPorcelainStatus(cmd *cobra.Command, ctxName, repoName, dir string) error {
	branch := os.Getenv("WIZ_BRANCH")
	state := "clean"

	if dir != "" {
		st, err := gitx.StatusAt(cmd.Context(), dir)
		if err == nil {
			branch = st.Branch
			if st.Dirty {
				state = "dirty"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s %s %s %d %d\n",
				ctxName, repoName, branch, state, st.Ahead, st.Behind)
			return nil
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%s %s %s %s 0 0\n",
		ctxName, repoName, branch, state)
	return nil
}

func init() {
	statusCmd.Flags().Bool("porcelain", false, "Machine-readable output for prompt hooks")
	statusCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(statusCmd)
}
