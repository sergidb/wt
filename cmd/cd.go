package cmd

import (
	"github.com/sergidb/wt/internal/git"
	"github.com/sergidb/wt/internal/shell"
	"github.com/sergidb/wt/internal/worktree"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cdCmd)
}

var cdCmd = &cobra.Command{
	Use:   "cd <name>",
	Short: "Navigate to a worktree",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := git.RepoRoot()
		if err != nil {
			return err
		}

		wts, err := worktree.List(repoRoot)
		if err != nil {
			return err
		}

		wt, err := worktree.FindByName(wts, args[0])
		if err != nil {
			return err
		}

		shell.PrintCdPath(wt.Path)
		return nil
	},
}
