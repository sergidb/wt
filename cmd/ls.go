package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/sergidb/wt/internal/git"
	"github.com/sergidb/wt/internal/worktree"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(lsCmd)
}

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all worktrees",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := git.RepoRoot()
		if err != nil {
			return err
		}

		wts, err := worktree.List(git.ExecOps{}, repoRoot)
		if err != nil {
			return err
		}

		if len(wts) == 0 {
			fmt.Println("No worktrees found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tBRANCH\tSOURCE\tPATH")
		for _, wt := range wts {
			branch := wt.Branch
			if branch == "" {
				branch = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", wt.Name, branch, wt.Source, wt.Path)
		}
		return w.Flush()
	},
}
