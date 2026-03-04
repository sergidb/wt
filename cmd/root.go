package cmd

import (
	"fmt"
	"os"

	"github.com/sergidb/wt/internal/config"
	"github.com/sergidb/wt/internal/git"
	"github.com/sergidb/wt/internal/runner"
	"github.com/sergidb/wt/internal/shell"
	"github.com/sergidb/wt/internal/tui"
	"github.com/sergidb/wt/internal/worktree"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wt",
	Short: "Git worktree manager with interactive TUI",
	Long:  "A tool to navigate and manage git worktrees, optimized for AI-assisted development workflows.",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := git.RepoRoot()
		if err != nil {
			return err
		}

		result, err := tui.Run(repoRoot)
		if err != nil {
			return err
		}

		switch result.Kind {
		case tui.ActionNone:
			return nil
		case tui.ActionConfig:
			return tui.RunConfig(repoRoot)
		case tui.ActionRun:
			cfg, err := config.Load(repoRoot)
			if err != nil {
				return err
			}
			return runner.Run(cfg.Services, result.Path, os.Stderr)
		case tui.ActionRm:
			if err := git.WorktreeRemove(
				findWorktreePath(repoRoot, result.Name), false,
			); err != nil {
				return fmt.Errorf("failed to remove worktree '%s': %w", result.Name, err)
			}
			fmt.Fprintf(os.Stderr, "Removed worktree '%s'\n", result.Name)
			return nil
		case tui.ActionCd:
			shell.PrintCdPath(os.Stdout, result.Path)
			return nil
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
	cfg := config.LoadOrEmpty(repoRoot)
	wts, err := worktree.List(repoRoot, cfg.GetWorktreesDir(repoRoot))
	if err != nil {
		return name
	}
	wt, err := worktree.FindByName(wts, name)
	if err != nil {
		return name
	}
	return wt.Path
}
