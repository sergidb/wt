package tui

// ActionKind identifies what the TUI wants the caller to do.
type ActionKind int

const (
	ActionNone   ActionKind = iota
	ActionCd                // Navigate to a worktree
	ActionRun               // Start services for a worktree
	ActionRm                // Remove a worktree
	ActionConfig            // Open config editor
)

// Result is the typed outcome of a TUI session.
type Result struct {
	Kind ActionKind
	Path string // Worktree path (Cd, Run)
	Name string // Worktree name (Rm)
}
