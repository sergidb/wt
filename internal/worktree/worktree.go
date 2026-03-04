package worktree

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sergidb/wt/internal/git"
)

// Source indicates where the worktree was discovered.
type Source int

const (
	SourceGit    Source = iota // Standard git worktree
	SourceClaude              // Found in .claude/worktrees/
)

func (s Source) String() string {
	switch s {
	case SourceGit:
		return "git"
	case SourceClaude:
		return "claude"
	default:
		return "unknown"
	}
}

// Worktree represents a single worktree.
type Worktree struct {
	Name    string
	Path    string
	Branch  string
	Source  Source
	IsMain  bool
	ModTime time.Time
}

// ClaudeWorktreesDir returns the path to the .claude/worktrees directory.
func ClaudeWorktreesDir(repoRoot string) string {
	return filepath.Join(repoRoot, ".claude", "worktrees")
}

// List returns all discovered worktrees, merging git worktrees and .claude/worktrees.
func List(repoRoot string) ([]Worktree, error) {
	seen := make(map[string]*Worktree) // keyed by absolute path

	// 1. Parse git worktree list
	rawList, err := git.WorktreeList(repoRoot)
	if err != nil {
		return nil, err
	}

	for _, raw := range rawList {
		absPath, _ := filepath.Abs(raw.Path)
		absPath = filepath.Clean(absPath)

		isMain := absPath == filepath.Clean(repoRoot)

		name := filepath.Base(absPath)
		if isMain {
			name = "main"
		}

		source := SourceGit
		if strings.Contains(absPath, filepath.Join(".claude", "worktrees")) {
			source = SourceClaude
		}

		modTime := dirModTime(absPath)

		seen[absPath] = &Worktree{
			Name:    name,
			Path:    absPath,
			Branch:  raw.BranchShort(),
			Source:  source,
			IsMain:  isMain,
			ModTime: modTime,
		}
	}

	// 2. Scan .claude/worktrees/ for any not already discovered
	claudeDir := ClaudeWorktreesDir(repoRoot)
	entries, err := os.ReadDir(claudeDir)
	if err == nil { // directory might not exist
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			absPath := filepath.Join(claudeDir, entry.Name())
			absPath = filepath.Clean(absPath)

			// Skip if already found via git worktree list
			if _, exists := seen[absPath]; exists {
				continue
			}

			// Check if it looks like a worktree (has .git file)
			gitFile := filepath.Join(absPath, ".git")
			if _, err := os.Stat(gitFile); err != nil {
				continue
			}

			modTime := dirModTime(absPath)

			seen[absPath] = &Worktree{
				Name:    entry.Name(),
				Path:    absPath,
				Branch:  "", // Unknown without git metadata
				Source:  SourceClaude,
				IsMain:  false,
				ModTime: modTime,
			}
		}
	}

	// Convert map to sorted slice (main first, then by name)
	result := make([]Worktree, 0, len(seen))
	for _, wt := range seen {
		result = append(result, *wt)
	}

	sortWorktrees(result)
	return result, nil
}

// FindByName finds a worktree by name or branch, with fuzzy matching.
// Returns exact match first, then substring match.
func FindByName(worktrees []Worktree, name string) (*Worktree, error) {
	name = strings.ToLower(name)

	// Exact match on name or branch
	for i := range worktrees {
		if strings.ToLower(worktrees[i].Name) == name ||
			strings.ToLower(worktrees[i].Branch) == name {
			return &worktrees[i], nil
		}
	}

	// Substring match
	var matches []*Worktree
	for i := range worktrees {
		if strings.Contains(strings.ToLower(worktrees[i].Name), name) ||
			strings.Contains(strings.ToLower(worktrees[i].Branch), name) {
			matches = append(matches, &worktrees[i])
		}
	}

	if len(matches) == 1 {
		return matches[0], nil
	}

	if len(matches) > 1 {
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = m.Name
		}
		return nil, &AmbiguousError{Name: name, Matches: names}
	}

	return nil, &NotFoundError{Name: name}
}

// AmbiguousError is returned when multiple worktrees match a name.
type AmbiguousError struct {
	Name    string
	Matches []string
}

func (e *AmbiguousError) Error() string {
	return "ambiguous worktree name '" + e.Name + "', matches: " + strings.Join(e.Matches, ", ")
}

// NotFoundError is returned when no worktree matches a name.
type NotFoundError struct {
	Name string
}

func (e *NotFoundError) Error() string {
	return "worktree not found: " + e.Name
}

func dirModTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

func sortWorktrees(wts []Worktree) {
	// Simple insertion sort (lists are small)
	for i := 1; i < len(wts); i++ {
		for j := i; j > 0 && less(wts[j], wts[j-1]); j-- {
			wts[j], wts[j-1] = wts[j-1], wts[j]
		}
	}
}

func less(a, b Worktree) bool {
	// Main worktree always first
	if a.IsMain != b.IsMain {
		return a.IsMain
	}
	return strings.ToLower(a.Name) < strings.ToLower(b.Name)
}
