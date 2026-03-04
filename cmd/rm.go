package cmd

import (
	"fmt"

	"github.com/sergidb/wt/internal/config"
	"github.com/sergidb/wt/internal/git"
	"github.com/sergidb/wt/internal/worktree"
	"github.com/spf13/cobra"
)

var rmForce bool

func init() {
	rmCmd.Flags().BoolVarP(&rmForce, "force", "f", false, "Force removal even with uncommitted changes")
	rootCmd.AddCommand(rmCmd)
}

var rmCmd = &cobra.Command{
	Use:   "rm <name>",
	Short: "Remove a worktree",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := git.RepoRoot()
		if err != nil {
			return err
		}

		cfg := config.LoadOrEmpty(repoRoot)
		wts, err := worktree.List(repoRoot, cfg.GetWorktreesDir(repoRoot))
		if err != nil {
			return err
		}

		wt, err := worktree.FindByName(wts, args[0])
		if err != nil {
			return err
		}

		if wt.IsMain {
			return fmt.Errorf("cannot remove the main worktree")
		}

		if err := git.WorktreeRemove(wt.Path, rmForce); err != nil {
			return fmt.Errorf("failed to remove worktree: %w", err)
		}

		fmt.Printf("Removed worktree '%s'\n", wt.Name)
		return nil
	},
}
