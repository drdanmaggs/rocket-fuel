package testutil

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
)

// InitTestRepo creates a real git repo in a temp directory with an initial commit.
// Returns the repo path. Cleanup is automatic via t.TempDir().
func InitTestRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	ctx := context.Background()

	commands := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@rocket-fuel.dev"},
		{"git", "config", "user.name", "Rocket Fuel Test"},
		{"git", "commit", "--allow-empty", "-m", "initial commit"},
	}

	for _, args := range commands {
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Dir = dir

		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s failed: %v\n%s", args[0], err, out)
		}
	}

	return dir
}

// InitTestRepoWithWorktree creates a repo and adds a worktree.
// Returns (repo path, worktree path). Both cleaned up automatically.
func InitTestRepoWithWorktree(t *testing.T, worktreeName string) (string, string) {
	t.Helper()

	repoDir := InitTestRepo(t)
	ctx := context.Background()
	worktreeDir := filepath.Join(t.TempDir(), worktreeName)
	branchName := "rf/" + worktreeName

	cmd := exec.CommandContext(ctx, "git", "worktree", "add", "-b", branchName, worktreeDir)
	cmd.Dir = repoDir

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add failed: %v\n%s", err, out)
	}

	t.Cleanup(func() {
		remove := exec.CommandContext(context.Background(), "git", "worktree", "remove", "--force", worktreeDir)
		remove.Dir = repoDir
		_ = remove.Run() //nolint:errcheck // best-effort cleanup
	})

	return repoDir, worktreeDir
}

// GitRun executes a git command in the given directory and returns stdout.
func GitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.CommandContext(context.Background(), "git", args...)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed in %s: %v\n%s", args, dir, err, out)
	}

	return string(out)
}
