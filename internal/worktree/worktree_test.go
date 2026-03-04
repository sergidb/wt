package worktree

import (
	"errors"
	"path/filepath"
	"testing"
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
