package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitTestRepoCreatesGitRepo(t *testing.T) {
	t.Parallel()

	repoDir := InitTestRepo(t)

	// .git directory should exist
	gitDir := filepath.Join(repoDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Fatalf("expected .git directory at %s", gitDir)
	}

	// Should have at least one commit
	log := GitRun(t, repoDir, "log", "--oneline")
	if !strings.Contains(log, "initial commit") {
		t.Errorf("expected initial commit in log, got: %s", log)
	}
}

func TestInitTestRepoWithWorktree(t *testing.T) {
	t.Parallel()

	repoDir, worktreeDir := InitTestRepoWithWorktree(t, "worker-1")

	// Worktree directory should exist
	if _, err := os.Stat(worktreeDir); os.IsNotExist(err) {
		t.Fatalf("expected worktree directory at %s", worktreeDir)
	}

	// Worktree should be on the correct branch
	branch := strings.TrimSpace(GitRun(t, worktreeDir, "branch", "--show-current"))
	if branch != "rf/worker-1" {
		t.Errorf("expected branch 'rf/worker-1', got %q", branch)
	}

	// Main repo should list the worktree
	worktrees := GitRun(t, repoDir, "worktree", "list")
	if !strings.Contains(worktrees, worktreeDir) {
		t.Errorf("expected worktree list to contain %s, got: %s", worktreeDir, worktrees)
	}
}

func TestGitRunExecutesCommands(t *testing.T) {
	t.Parallel()

	repoDir := InitTestRepo(t)

	status := GitRun(t, repoDir, "status", "--porcelain")
	if strings.TrimSpace(status) != "" {
		t.Errorf("expected clean status, got: %q", status)
	}
}
