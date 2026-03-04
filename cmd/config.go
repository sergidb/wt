package cmd

import (
	"github.com/sergidb/wt/internal/git"
	"github.com/sergidb/wt/internal/tui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure services interactively",
	Long:  "Opens an interactive TUI to add, edit, and remove services defined in .wt.yaml.",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := git.RepoRoot()
		if err != nil {
			return err
		}

		return tui.RunConfig(repoRoot)
	},
}
