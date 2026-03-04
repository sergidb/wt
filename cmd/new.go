package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/sergidb/wt/internal/git"
	"github.com/sergidb/wt/internal/worktree"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use:   "new <branch>",
	Short: "Create a new worktree",
	Long:  "Creates a new worktree in .claude/worktrees/<branch>. Creates the branch if it doesn't exist.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := args[0]

		repoRoot, err := git.RepoRoot()
		if err != nil {
			return err
		}

		wtPath := filepath.Join(worktree.ClaudeWorktreesDir(repoRoot), branch)
		createBranch := !git.BranchExists(repoRoot, branch)

		if err := git.WorktreeAdd(repoRoot, wtPath, branch, createBranch); err != nil {
			return fmt.Errorf("failed to create worktree: %w", err)
		}

		fmt.Printf("Created worktree '%s' at %s\n", branch, wtPath)
		return nil
	},
}
