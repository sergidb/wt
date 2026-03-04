package cmd

import (
	"github.com/sergidb/wt/internal/git"
	"github.com/sergidb/wt/internal/shell"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(homeCmd)
}

var homeCmd = &cobra.Command{
	Use:   "home",
	Short: "Navigate to the main repository root",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := git.RepoRoot()
		if err != nil {
			return err
		}

		shell.PrintCdPath(repoRoot)
		return nil
	},
}
