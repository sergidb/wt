package cmd

import (
	"fmt"

	"github.com/sergidb/wt/internal/git"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pruneCmd)
}

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Clean up stale worktree entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := git.RepoRoot()
		if err != nil {
			return err
		}

		if err := git.WorktreePrune(repoRoot); err != nil {
			return fmt.Errorf("failed to prune: %w", err)
		}

		fmt.Println("Pruned stale worktrees.")
		return nil
	},
}
