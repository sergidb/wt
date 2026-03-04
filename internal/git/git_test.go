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
