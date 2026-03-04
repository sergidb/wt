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
		{SourceManaged, "managed"},
		{Source(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.source.String(); got != tt.want {
			t.Errorf("Source(%d).String() = %q, want %q", tt.source, got, tt.want)
		}
	}
}

// mockGitWorktreeList replaces gitWorktreeListFunc for the duration of a test.
func mockGitWorktreeList(t *testing.T, worktrees []git.RawWorktree, err error) {
	t.Helper()
	orig := gitWorktreeListFunc
	gitWorktreeListFunc = func(repoRoot string) ([]git.RawWorktree, error) {
		return worktrees, err
	}
	t.Cleanup(func() { gitWorktreeListFunc = orig })
}

func TestList(t *testing.T) {
	t.Run("git worktrees only", func(t *testing.T) {
		tmp := t.TempDir()
		mockGitWorktreeList(t, []git.RawWorktree{
			{Path: tmp, HEAD: "abc123", Branch: "refs/heads/main"},
		}, nil)

		worktreesDir := filepath.Join(tmp, ".managed", "worktrees")
		wts, err := List(tmp, worktreesDir)
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

	t.Run("managed worktrees discovered from filesystem", func(t *testing.T) {
		tmp := t.TempDir()

		// Create a managed worktrees directory with feat-x
		worktreesDir := filepath.Join(tmp, ".managed", "worktrees")
		managedWT := filepath.Join(worktreesDir, "feat-x")
		if err := os.MkdirAll(managedWT, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(managedWT, ".git"), []byte("gitdir: ..."), 0644); err != nil {
			t.Fatal(err)
		}

		// Git only returns the main worktree
		mockGitWorktreeList(t, []git.RawWorktree{
			{Path: tmp, HEAD: "abc", Branch: "refs/heads/main"},
		}, nil)

		wts, err := List(tmp, worktreesDir)
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

		// feat-x should be discovered as SourceManaged
		if wts[1].Name != "feat-x" {
			t.Errorf("second worktree Name = %q, want %q", wts[1].Name, "feat-x")
		}
		if wts[1].Source != SourceManaged {
			t.Errorf("Source = %v, want SourceManaged", wts[1].Source)
		}
		if wts[1].Branch != "" {
			t.Errorf("Branch = %q, want empty (unknown)", wts[1].Branch)
		}
	})

	t.Run("duplicate worktree seen by git and managed dir", func(t *testing.T) {
		tmp := t.TempDir()

		worktreesDir := filepath.Join(tmp, ".managed", "worktrees")
		managedWT := filepath.Join(worktreesDir, "feat-y")
		if err := os.MkdirAll(managedWT, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(managedWT, ".git"), []byte("gitdir: ..."), 0644); err != nil {
			t.Fatal(err)
		}

		// Git already reports this worktree
		mockGitWorktreeList(t, []git.RawWorktree{
			{Path: tmp, HEAD: "abc", Branch: "refs/heads/main"},
			{Path: managedWT, HEAD: "def", Branch: "refs/heads/feat-y"},
		}, nil)

		wts, err := List(tmp, worktreesDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should not duplicate: main + feat-y = 2
		if len(wts) != 2 {
			t.Fatalf("got %d worktrees, want 2 (no duplicates)", len(wts))
		}
	})

	t.Run("managed dir without .git file is skipped", func(t *testing.T) {
		tmp := t.TempDir()

		worktreesDir := filepath.Join(tmp, ".managed", "worktrees")
		managedWT := filepath.Join(worktreesDir, "broken")
		if err := os.MkdirAll(managedWT, 0755); err != nil {
			t.Fatal(err)
		}

		mockGitWorktreeList(t, []git.RawWorktree{
			{Path: tmp, HEAD: "abc", Branch: "refs/heads/main"},
		}, nil)

		wts, err := List(tmp, worktreesDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(wts) != 1 {
			t.Fatalf("got %d worktrees, want 1 (broken skipped)", len(wts))
		}
	})

	t.Run("no managed worktrees dir", func(t *testing.T) {
		tmp := t.TempDir()
		mockGitWorktreeList(t, []git.RawWorktree{
			{Path: tmp, HEAD: "abc", Branch: "refs/heads/main"},
		}, nil)

		worktreesDir := filepath.Join(tmp, ".managed", "worktrees")
		wts, err := List(tmp, worktreesDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(wts) != 1 {
			t.Fatalf("got %d worktrees, want 1", len(wts))
		}
	})

	t.Run("git ops error", func(t *testing.T) {
		mockGitWorktreeList(t, nil, errors.New("git failed"))
		_, err := List("/nonexistent", "/nonexistent/.managed/worktrees")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("worktree in managed path detected as SourceManaged via git", func(t *testing.T) {
		tmp := t.TempDir()

		worktreesDir := filepath.Join(tmp, ".managed", "worktrees")
		managedWT := filepath.Join(worktreesDir, "feat-z")
		if err := os.MkdirAll(managedWT, 0755); err != nil {
			t.Fatal(err)
		}

		// Git reports this worktree (it knows about it)
		mockGitWorktreeList(t, []git.RawWorktree{
			{Path: tmp, HEAD: "abc", Branch: "refs/heads/main"},
			{Path: managedWT, HEAD: "def", Branch: "refs/heads/feat-z"},
		}, nil)

		wts, err := List(tmp, worktreesDir)
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
		if found.Source != SourceManaged {
			t.Errorf("Source = %v, want SourceManaged", found.Source)
		}
	})

	t.Run("sorting main first then alphabetical", func(t *testing.T) {
		tmp := t.TempDir()

		worktreesDir := filepath.Join(tmp, ".managed", "worktrees")
		for _, name := range []string{"zebra", "alpha"} {
			dir := filepath.Join(worktreesDir, name)
			os.MkdirAll(dir, 0755)
			os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: ..."), 0644)
		}

		mockGitWorktreeList(t, []git.RawWorktree{
			{Path: tmp, HEAD: "abc", Branch: "refs/heads/main"},
		}, nil)

		wts, err := List(tmp, worktreesDir)
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
