package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/sergidb/wt/internal/config"
	"github.com/sergidb/wt/internal/git"
	"github.com/sergidb/wt/internal/runner"
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

		if result == "config" {
			return tui.RunConfig(repoRoot)
		}

		if strings.HasPrefix(result, "run:") {
			wtPath := strings.TrimPrefix(result, "run:")
			cfg, err := config.Load(repoRoot)
			if err != nil {
				return err
			}
			return runner.Run(cfg.Services, wtPath)
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
