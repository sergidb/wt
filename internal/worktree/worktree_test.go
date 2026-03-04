package worktree

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sergidb/wt/internal/git"
)

func TestFindByName(t *testing.T) {
	worktrees := []Worktree{
		{Name: "main", Path: "/repo", Branch: "main", IsMain: true},
		{Name: "feature-auth", Path: "/repo-auth", Branch: "feature/auth"},
		{Name: "feature-api", Path: "/repo-api", Branch: "feature/api"},
		{Name: "bugfix", Path: "/repo-bugfix", Branch: "bugfix/login"},
	}

	t.Run("exact match by name", func(t *testing.T) {
		wt, err := FindByName(worktrees, "main")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if wt.Name != "main" {
			t.Errorf("got Name=%q, want %q", wt.Name, "main")
		}
	})

	t.Run("exact match by branch", func(t *testing.T) {
		wt, err := FindByName(worktrees, "feature/auth")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if wt.Name != "feature-auth" {
			t.Errorf("got Name=%q, want %q", wt.Name, "feature-auth")
		}
	})

	t.Run("substring match unique", func(t *testing.T) {
		wt, err := FindByName(worktrees, "bugfix")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if wt.Name != "bugfix" {
			t.Errorf("got Name=%q, want %q", wt.Name, "bugfix")
		}
	})

	t.Run("ambiguous match", func(t *testing.T) {
		_, err := FindByName(worktrees, "feature")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var ambErr *AmbiguousError
		if !errors.As(err, &ambErr) {
			t.Fatalf("expected AmbiguousError, got %T: %v", err, err)
		}
		if len(ambErr.Matches) != 2 {
			t.Errorf("got %d matches, want 2", len(ambErr.Matches))
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := FindByName(worktrees, "nonexistent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var nfErr *NotFoundError
		if !errors.As(err, &nfErr) {
			t.Fatalf("expected NotFoundError, got %T: %v", err, err)
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		wt, err := FindByName(worktrees, "MAIN")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if wt.Name != "main" {
			t.Errorf("got Name=%q, want %q", wt.Name, "main")
		}
	})
}

func TestSortWorktrees(t *testing.T) {
	wts := []Worktree{
		{Name: "zebra"},
		{Name: "Alpha"},
		{Name: "main", IsMain: true},
		{Name: "beta"},
	}

	sortWorktrees(wts)

	expected := []string{"main", "Alpha", "beta", "zebra"}
	for i, want := range expected {
		if wts[i].Name != want {
			t.Errorf("position %d: got %q, want %q", i, wts[i].Name, want)
		}
	}
}

func TestSourceString(t *testing.T) {
	tests := []struct {
		source Source
		want   string
	}{
		{SourceGit, "git"},
		{SourceClaude, "claude"},
		{Source(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.source.String(); got != tt.want {
			t.Errorf("Source(%d).String() = %q, want %q", tt.source, got, tt.want)
		}
	}
}

func TestClaudeWorktreesDir(t *testing.T) {
	got := ClaudeWorktreesDir("/home/user/repo")
	want := filepath.Join("/home/user/repo", ".claude", "worktrees")
	if got != want {
		t.Errorf("ClaudeWorktreesDir() = %q, want %q", got, want)
	}
}

// mockOps implements git.Ops for testing.
type mockOps struct {
	worktrees []git.RawWorktree
	err       error
}

func (m *mockOps) WorktreeList(repoRoot string) ([]git.RawWorktree, error) {
	return m.worktrees, m.err
}

func TestList(t *testing.T) {
	t.Run("git worktrees only", func(t *testing.T) {
		tmp := t.TempDir()
		ops := &mockOps{
			worktrees: []git.RawWorktree{
				{Path: tmp, HEAD: "abc123", Branch: "refs/heads/main"},
			},
		}

		wts, err := List(ops, tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(wts) != 1 {
			t.Fatalf("got %d worktrees, want 1", len(wts))
		}
		if wts[0].Name != "main" {
			t.Errorf("Name = %q, want %q", wts[0].Name, "main")
		}
		if !wts[0].IsMain {
			t.Error("expected IsMain = true")
		}
		if wts[0].Source != SourceGit {
			t.Errorf("Source = %v, want SourceGit", wts[0].Source)
		}
		if wts[0].Branch != "main" {
			t.Errorf("Branch = %q, want %q", wts[0].Branch, "main")
		}
	})

	t.Run("claude worktrees discovered from filesystem", func(t *testing.T) {
		tmp := t.TempDir()

		// Create a .claude/worktrees/feat-x directory with a .git file
		claudeWT := filepath.Join(tmp, ".claude", "worktrees", "feat-x")
		if err := os.MkdirAll(claudeWT, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(claudeWT, ".git"), []byte("gitdir: ..."), 0644); err != nil {
			t.Fatal(err)
		}

		// Git only returns the main worktree
		ops := &mockOps{
			worktrees: []git.RawWorktree{
				{Path: tmp, HEAD: "abc", Branch: "refs/heads/main"},
			},
		}

		wts, err := List(ops, tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(wts) != 2 {
			t.Fatalf("got %d worktrees, want 2", len(wts))
		}

		// main should be first
		if wts[0].Name != "main" {
			t.Errorf("first worktree Name = %q, want %q", wts[0].Name, "main")
		}

		// feat-x should be discovered as SourceClaude
		if wts[1].Name != "feat-x" {
			t.Errorf("second worktree Name = %q, want %q", wts[1].Name, "feat-x")
		}
		if wts[1].Source != SourceClaude {
			t.Errorf("Source = %v, want SourceClaude", wts[1].Source)
		}
		if wts[1].Branch != "" {
			t.Errorf("Branch = %q, want empty (unknown)", wts[1].Branch)
		}
	})

	t.Run("duplicate worktree seen by git and claude dir", func(t *testing.T) {
		tmp := t.TempDir()

		// Create .claude/worktrees/feat-y with .git file
		claudeWT := filepath.Join(tmp, ".claude", "worktrees", "feat-y")
		if err := os.MkdirAll(claudeWT, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(claudeWT, ".git"), []byte("gitdir: ..."), 0644); err != nil {
			t.Fatal(err)
		}

		// Git already reports this worktree
		ops := &mockOps{
			worktrees: []git.RawWorktree{
				{Path: tmp, HEAD: "abc", Branch: "refs/heads/main"},
				{Path: claudeWT, HEAD: "def", Branch: "refs/heads/feat-y"},
			},
		}

		wts, err := List(ops, tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should not duplicate: main + feat-y = 2
		if len(wts) != 2 {
			t.Fatalf("got %d worktrees, want 2 (no duplicates)", len(wts))
		}
	})

	t.Run("claude dir without .git file is skipped", func(t *testing.T) {
		tmp := t.TempDir()

		// Create a directory in .claude/worktrees/ but without .git
		claudeWT := filepath.Join(tmp, ".claude", "worktrees", "broken")
		if err := os.MkdirAll(claudeWT, 0755); err != nil {
			t.Fatal(err)
		}

		ops := &mockOps{
			worktrees: []git.RawWorktree{
				{Path: tmp, HEAD: "abc", Branch: "refs/heads/main"},
			},
		}

		wts, err := List(ops, tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(wts) != 1 {
			t.Fatalf("got %d worktrees, want 1 (broken skipped)", len(wts))
		}
	})

	t.Run("no claude worktrees dir", func(t *testing.T) {
		tmp := t.TempDir()
		ops := &mockOps{
			worktrees: []git.RawWorktree{
				{Path: tmp, HEAD: "abc", Branch: "refs/heads/main"},
			},
		}

		wts, err := List(ops, tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(wts) != 1 {
			t.Fatalf("got %d worktrees, want 1", len(wts))
		}
	})

	t.Run("git ops error", func(t *testing.T) {
		ops := &mockOps{err: errors.New("git failed")}
		_, err := List(ops, "/nonexistent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("worktree in claude path detected as SourceClaude via git", func(t *testing.T) {
		tmp := t.TempDir()

		claudeWT := filepath.Join(tmp, ".claude", "worktrees", "feat-z")
		if err := os.MkdirAll(claudeWT, 0755); err != nil {
			t.Fatal(err)
		}

		// Git reports this worktree (it knows about it)
		ops := &mockOps{
			worktrees: []git.RawWorktree{
				{Path: tmp, HEAD: "abc", Branch: "refs/heads/main"},
				{Path: claudeWT, HEAD: "def", Branch: "refs/heads/feat-z"},
			},
		}

		wts, err := List(ops, tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Find feat-z
		var found *Worktree
		for i := range wts {
			if wts[i].Name == "feat-z" {
				found = &wts[i]
				break
			}
		}
		if found == nil {
			t.Fatal("feat-z not found")
		}
		if found.Source != SourceClaude {
			t.Errorf("Source = %v, want SourceClaude", found.Source)
		}
	})

	t.Run("sorting main first then alphabetical", func(t *testing.T) {
		tmp := t.TempDir()

		// Create claude worktrees
		for _, name := range []string{"zebra", "alpha"} {
			dir := filepath.Join(tmp, ".claude", "worktrees", name)
			os.MkdirAll(dir, 0755)
			os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: ..."), 0644)
		}

		ops := &mockOps{
			worktrees: []git.RawWorktree{
				{Path: tmp, HEAD: "abc", Branch: "refs/heads/main"},
			},
		}

		wts, err := List(ops, tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(wts) != 3 {
			t.Fatalf("got %d worktrees, want 3", len(wts))
		}

		expected := []string{"main", "alpha", "zebra"}
		for i, want := range expected {
			if wts[i].Name != want {
				t.Errorf("position %d: got %q, want %q", i, wts[i].Name, want)
			}
		}
	})
}

func TestErrorMessages(t *testing.T) {
	t.Run("AmbiguousError message", func(t *testing.T) {
		err := &AmbiguousError{Name: "feat", Matches: []string{"feat-a", "feat-b"}}
		got := err.Error()
		if got != "ambiguous worktree name 'feat', matches: feat-a, feat-b" {
			t.Errorf("Error() = %q", got)
		}
	})

	t.Run("NotFoundError message", func(t *testing.T) {
		err := &NotFoundError{Name: "xyz"}
		got := err.Error()
		if got != "worktree not found: xyz" {
			t.Errorf("Error() = %q", got)
		}
	})
}
