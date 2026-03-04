package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/sergidb/wt/internal/config"
	"github.com/sergidb/wt/internal/git"
	"github.com/sergidb/wt/internal/runner"
	"github.com/sergidb/wt/internal/tui"
	"github.com/sergidb/wt/internal/worktree"
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

		if strings.HasPrefix(result, "rm:") {
			name := strings.TrimPrefix(result, "rm:")
			if err := git.WorktreeRemove(
				findWorktreePath(repoRoot, name), false,
			); err != nil {
				return fmt.Errorf("failed to remove worktree '%s': %w", name, err)
			}
			fmt.Fprintf(os.Stderr, "Removed worktree '%s'\n", name)
			return nil
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

func findWorktreePath(repoRoot, name string) string {
	wts, err := worktree.List(repoRoot)
	if err != nil {
		return name
	}
	wt, err := worktree.FindByName(wts, name)
	if err != nil {
		return name
	}
	return wt.Path
}
