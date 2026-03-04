package cmd

import (
	"fmt"
	"os"

	"github.com/sergidb/wt/internal/git"
	"github.com/sergidb/wt/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wt",
	Short: "Git worktree manager with interactive TUI",
	Long:  "A tool to navigate and manage git worktrees, optimized for Claude Code workflows.",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := git.RepoRoot()
		if err != nil {
			return err
		}

		result, err := tui.Run(repoRoot)
		if err != nil {
			return err
		}

		if result != "" {
			fmt.Print(result)
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
