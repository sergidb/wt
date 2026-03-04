package cmd

import (
	"fmt"
	"os"

	"github.com/sergidb/wt/internal/config"
	"github.com/sergidb/wt/internal/git"
	"github.com/sergidb/wt/internal/runner"
	"github.com/sergidb/wt/internal/worktree"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run [worktree] [service...]",
	Short: "Run services for a worktree",
	Long: `Start services defined in .wt.yaml for a specific worktree.

Examples:
  wt run fix-auth          Run all services in the fix-auth worktree
  wt run fix-auth backend  Run only the backend service
  wt run                   Run all services in the current worktree`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := git.RepoRoot()
		if err != nil {
			return err
		}

		// Load config
		cfg, err := config.Load(repoRoot)
		if err != nil {
			return err
		}

		// Determine worktree path and which services to run
		var wtPath string
		var serviceNames []string

		if len(args) == 0 {
			// No args: use current worktree
			wtPath, err = git.WorktreeRoot()
			if err != nil {
				return err
			}
		} else {
			// First arg might be a worktree name or a service name
			wts, err := worktree.List(git.ExecOps{}, repoRoot)
			if err != nil {
				return err
			}

			wt, wtErr := worktree.FindByName(wts, args[0])
			if wtErr == nil {
				// First arg is a worktree
				wtPath = wt.Path
				serviceNames = args[1:]
			} else {
				// First arg is not a worktree — treat all args as service names
				// and use current worktree
				wtPath, err = git.WorktreeRoot()
				if err != nil {
					return err
				}
				serviceNames = args
			}
		}

		// Filter services
		services := cfg.Services
		if len(serviceNames) > 0 {
			filtered := make(map[string]config.Service)
			for _, name := range serviceNames {
				svc, ok := cfg.Services[name]
				if !ok {
					return fmt.Errorf("unknown service '%s' (available: %s)", name, availableServices(cfg))
				}
				filtered[name] = svc
			}
			services = filtered
		}

		return runner.Run(services, wtPath, os.Stderr)
	},
}

func availableServices(cfg *config.Config) string {
	names := make([]string, 0, len(cfg.Services))
	for name := range cfg.Services {
		names = append(names, name)
	}
	result := ""
	for i, name := range names {
		if i > 0 {
			result += ", "
		}
		result += name
	}
	return result
}
