package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParsePorcelain(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []RawWorktree
	}{
		{
			name: "single worktree",
			input: "worktree /home/user/repo\nHEAD abc123\nbranch refs/heads/main\n\n",
			expected: []RawWorktree{
				{Path: "/home/user/repo", HEAD: "abc123", Branch: "refs/heads/main"},
			},
		},
		{
			name: "multiple worktrees",
			input: "worktree /home/user/repo\nHEAD abc123\nbranch refs/heads/main\n\nworktree /home/user/repo-feat\nHEAD def456\nbranch refs/heads/feature\n\n",
			expected: []RawWorktree{
				{Path: "/home/user/repo", HEAD: "abc123", Branch: "refs/heads/main"},
				{Path: "/home/user/repo-feat", HEAD: "def456", Branch: "refs/heads/feature"},
			},
		},
		{
			name: "detached HEAD",
			input: "worktree /home/user/repo\nHEAD abc123\ndetached\n\n",
			expected: []RawWorktree{
				{Path: "/home/user/repo", HEAD: "abc123"},
			},
		},
		{
			name: "bare repo",
			input: "worktree /home/user/repo.git\nHEAD abc123\nbare\n\n",
			expected: []RawWorktree{
				{Path: "/home/user/repo.git", HEAD: "abc123", Bare: true},
			},
		},
		{
			name:  "no trailing newline",
			input: "worktree /home/user/repo\nHEAD abc123\nbranch refs/heads/main",
			expected: []RawWorktree{
				{Path: "/home/user/repo", HEAD: "abc123", Branch: "refs/heads/main"},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePorcelain(tt.input)

			if len(result) != len(tt.expected) {
				t.Fatalf("got %d worktrees, want %d", len(result), len(tt.expected))
			}

			for i, got := range result {
				want := tt.expected[i]
				if got.Path != want.Path {
					t.Errorf("[%d] Path = %q, want %q", i, got.Path, want.Path)
				}
				if got.HEAD != want.HEAD {
					t.Errorf("[%d] HEAD = %q, want %q", i, got.HEAD, want.HEAD)
				}
				if got.Branch != want.Branch {
					t.Errorf("[%d] Branch = %q, want %q", i, got.Branch, want.Branch)
				}
				if got.Bare != want.Bare {
					t.Errorf("[%d] Bare = %v, want %v", i, got.Bare, want.Bare)
				}
			}
		})
	}
}

func TestBranchShort(t *testing.T) {
	tests := []struct {
		branch string
		want   string
	}{
		{"refs/heads/main", "main"},
		{"refs/heads/feature/foo", "feature/foo"},
		{"main", "main"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			w := RawWorktree{Branch: tt.branch}
			if got := w.BranchShort(); got != tt.want {
				t.Errorf("BranchShort() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRepoRootFromPath(t *testing.T) {
	// Create a temp dir with a .git directory
	tmp := t.TempDir()
	gitDir := filepath.Join(tmp, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a subdirectory to test walking up
	subDir := filepath.Join(tmp, "sub", "deep")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Change to the subdirectory
	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })

	if err := os.Chdir(subDir); err != nil {
		t.Fatal(err)
	}

	root, err := repoRootFromPath()
	if err != nil {
		t.Fatalf("repoRootFromPath() error: %v", err)
	}

	// Resolve symlinks for macOS (/tmp -> /private/tmp)
	wantRoot, _ := filepath.EvalSymlinks(tmp)
	gotRoot, _ := filepath.EvalSymlinks(root)

	if gotRoot != wantRoot {
		t.Errorf("repoRootFromPath() = %q, want %q", gotRoot, wantRoot)
	}
}

func TestRepoRoot(t *testing.T) {
	tmp := t.TempDir()

	// git init
	cmd := exec.Command("git", "init", tmp)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}

	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })

	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	root, err := RepoRoot()
	if err != nil {
		t.Fatalf("RepoRoot() error: %v", err)
	}

	// Resolve symlinks for comparison (macOS /tmp -> /private/tmp)
	wantRoot, _ := filepath.EvalSymlinks(tmp)
	gotRoot, _ := filepath.EvalSymlinks(root)

	if gotRoot != wantRoot {
		t.Errorf("RepoRoot() = %q, want %q", gotRoot, wantRoot)
	}
}

// initGitRepo creates a git repo with an initial commit and returns its resolved path.
func initGitRepo(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()

	cmds := [][]string{
		{"git", "init", tmp},
		{"git", "-C", tmp, "config", "user.email", "test@test.com"},
		{"git", "-C", tmp, "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		if out, err := exec.Command(args[0], args[1:]...).CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %v\n%s", args, err, out)
		}
	}

	// Create initial commit so branches work
	f := filepath.Join(tmp, "README.md")
	os.WriteFile(f, []byte("# test"), 0644)
	exec.Command("git", "-C", tmp, "add", ".").Run()
	exec.Command("git", "-C", tmp, "commit", "-m", "init").Run()

	resolved, _ := filepath.EvalSymlinks(tmp)
	return resolved
}

func TestWorktreeList(t *testing.T) {
	repo := initGitRepo(t)

	wts, err := WorktreeList(repo)
	if err != nil {
		t.Fatalf("WorktreeList() error: %v", err)
	}
	if len(wts) < 1 {
		t.Fatal("expected at least 1 worktree")
	}

	resolved, _ := filepath.EvalSymlinks(wts[0].Path)
	if resolved != repo {
		t.Errorf("first worktree path = %q, want %q", resolved, repo)
	}
}

func TestWorktreeAddAndRemove(t *testing.T) {
	repo := initGitRepo(t)

	wtPath := filepath.Join(t.TempDir(), "feat-test")

	// Add worktree with new branch
	if err := WorktreeAdd(repo, wtPath, "feat-test", true); err != nil {
		t.Fatalf("WorktreeAdd() error: %v", err)
	}

	// Verify it appears in the list
	wts, err := WorktreeList(repo)
	if err != nil {
		t.Fatalf("WorktreeList() error: %v", err)
	}

	found := false
	for _, wt := range wts {
		resolved, _ := filepath.EvalSymlinks(wt.Path)
		resolvedWT, _ := filepath.EvalSymlinks(wtPath)
		if resolved == resolvedWT {
			found = true
			if wt.BranchShort() != "feat-test" {
				t.Errorf("branch = %q, want %q", wt.BranchShort(), "feat-test")
			}
			break
		}
	}
	if !found {
		t.Error("added worktree not found in list")
	}

	// Remove it — WorktreeRemove runs git from cwd, so chdir to repo
	origDir, _ := os.Getwd()
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	if err := WorktreeRemove(wtPath, false); err != nil {
		t.Fatalf("WorktreeRemove() error: %v", err)
	}

	// Verify it's gone
	wts, err = WorktreeList(repo)
	if err != nil {
		t.Fatalf("WorktreeList() after remove error: %v", err)
	}
	for _, wt := range wts {
		resolved, _ := filepath.EvalSymlinks(wt.Path)
		resolvedWT, _ := filepath.EvalSymlinks(wtPath)
		if resolved == resolvedWT {
			t.Error("removed worktree still appears in list")
		}
	}
}

func TestBranchExists(t *testing.T) {
	repo := initGitRepo(t)

	// Default branch should exist
	// Find the default branch name (could be main or master)
	wts, _ := WorktreeList(repo)
	defaultBranch := wts[0].BranchShort()

	if !BranchExists(repo, defaultBranch) {
		t.Errorf("BranchExists(%q) = false, want true", defaultBranch)
	}

	if BranchExists(repo, "nonexistent-branch-xyz") {
		t.Error("BranchExists(nonexistent) = true, want false")
	}
}

func TestWorktreePrune(t *testing.T) {
	repo := initGitRepo(t)

	// Prune on a clean repo should succeed
	if err := WorktreePrune(repo); err != nil {
		t.Fatalf("WorktreePrune() error: %v", err)
	}
}

func TestWorktreeRoot(t *testing.T) {
	repo := initGitRepo(t)

	origDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origDir) })

	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}

	root, err := WorktreeRoot()
	if err != nil {
		t.Fatalf("WorktreeRoot() error: %v", err)
	}

	gotRoot, _ := filepath.EvalSymlinks(root)
	if gotRoot != repo {
		t.Errorf("WorktreeRoot() = %q, want %q", gotRoot, repo)
	}
}
