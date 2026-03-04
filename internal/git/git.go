package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RawWorktree is the direct parse of `git worktree list --porcelain`.
type RawWorktree struct {
	Path   string
	HEAD   string
	Branch string // e.g. "refs/heads/main", or empty if detached
	Bare   bool
}

// BranchShort returns the short branch name (e.g. "main" from "refs/heads/main").
func (w RawWorktree) BranchShort() string {
	return strings.TrimPrefix(w.Branch, "refs/heads/")
}

// RepoRoot returns the main repository root directory.
// Works from inside any worktree by using --git-common-dir.
// If the current directory no longer exists (deleted worktree),
// it walks up the path to find the repo root.
func RepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--git-common-dir").Output()
	if err != nil {
		// Current dir may not exist (deleted worktree). Try walking up.
		root, fallbackErr := repoRootFromPath()
		if fallbackErr != nil {
			return "", fmt.Errorf("not inside a git repository: %w", err)
		}
		return root, nil
	}

	gitDir := strings.TrimSpace(string(out))

	// Make absolute if relative
	if !filepath.IsAbs(gitDir) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		gitDir = filepath.Join(cwd, gitDir)
	}

	// Clean and resolve
	gitDir = filepath.Clean(gitDir)

	// The repo root is the parent of the .git directory
	root := filepath.Dir(gitDir)

	return root, nil
}

// repoRootFromPath walks up from the current working directory path
// looking for a .git directory. Useful when the cwd has been deleted
// (e.g. a removed worktree) but the path still contains the repo root.
func repoRootFromPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		gitPath := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("no git repository found in parent directories of %s", cwd)
}

// WorktreeRoot returns the top-level directory of the current worktree.
// Unlike RepoRoot, this returns the worktree's own root, not the main repo root.
func WorktreeRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// WorktreeList runs `git worktree list --porcelain` and parses the output.
func WorktreeList(repoRoot string) ([]RawWorktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed: %w", err)
	}

	return parsePorcelain(string(out)), nil
}

func parsePorcelain(output string) []RawWorktree {
	var worktrees []RawWorktree
	var current *RawWorktree

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			if current != nil {
				worktrees = append(worktrees, *current)
				current = nil
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			current = &RawWorktree{
				Path: strings.TrimPrefix(line, "worktree "),
			}
		} else if current != nil {
			switch {
			case strings.HasPrefix(line, "HEAD "):
				current.HEAD = strings.TrimPrefix(line, "HEAD ")
			case strings.HasPrefix(line, "branch "):
				current.Branch = strings.TrimPrefix(line, "branch ")
			case line == "bare":
				current.Bare = true
			}
		}
	}

	// Handle last entry if no trailing newline
	if current != nil {
		worktrees = append(worktrees, *current)
	}

	return worktrees
}

// WorktreeAdd creates a new worktree.
// If createBranch is true, creates a new branch with -b flag.
func WorktreeAdd(repoRoot, path, branch string, createBranch bool) error {
	args := []string{"worktree", "add"}
	if createBranch {
		args = append(args, "-b", branch, path)
	} else {
		args = append(args, path, branch)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// WorktreeRemove removes a worktree by path.
func WorktreeRemove(path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// WorktreePrune runs `git worktree prune` to clean up stale entries.
func WorktreePrune(repoRoot string) error {
	cmd := exec.Command("git", "worktree", "prune")
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// BranchExists checks if a branch exists in the repository.
func BranchExists(repoRoot, branch string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+branch)
	cmd.Dir = repoRoot
	return cmd.Run() == nil
}
