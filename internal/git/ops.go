package git

// Ops abstracts git operations for testability.
type Ops interface {
	WorktreeList(repoRoot string) ([]RawWorktree, error)
}

// ExecOps implements Ops by executing real git commands.
type ExecOps struct{}

func (ExecOps) WorktreeList(repoRoot string) ([]RawWorktree, error) {
	return WorktreeList(repoRoot)
}
