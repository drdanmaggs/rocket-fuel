package worker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
)

// ReapResult describes what was cleaned up for a single worker.
type ReapResult struct {
	WindowName  string
	WorktreeDir string
	Reaped      bool
	Reason      string
}

// Reap finds completed workers and cleans up their worktrees and tmux sessions.
// A worker is considered complete when its tmux session no longer exists
// (Claude Code session ended).
func Reap(tm tmux.Runner, _, repoDir string) ([]ReapResult, error) {
	worktreesDir := filepath.Join(repoDir, ".worktrees")

	entries, err := os.ReadDir(worktreesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read worktrees dir: %w", err)
	}

	var results []ReapResult

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name() // e.g. "worker-42"
		worktreeDir := filepath.Join(worktreesDir, name)

		// Extract issue number from worker name to find the session.
		issueNum := strings.TrimPrefix(name, "worker-")
		workerSession := "rf-worker-" + issueNum

		// Check if the worker's tmux session still exists.
		sessionExists := tm.HasSession(workerSession)

		if sessionExists {
			results = append(results, ReapResult{
				WindowName:  name,
				WorktreeDir: worktreeDir,
				Reaped:      false,
				Reason:      "session still active",
			})
			continue
		}

		// Session is gone — clean up the worktree.
		if err := removeWorktree(repoDir, worktreeDir); err != nil {
			results = append(results, ReapResult{
				WindowName:  name,
				WorktreeDir: worktreeDir,
				Reaped:      false,
				Reason:      fmt.Sprintf("cleanup failed: %v", err),
			})
			continue
		}

		results = append(results, ReapResult{
			WindowName:  name,
			WorktreeDir: worktreeDir,
			Reaped:      true,
			Reason:      "cleaned up",
		})
	}

	_ = pruneWorktrees(repoDir)

	return results, nil
}

func removeWorktree(repoDir, worktreeDir string) error {
	cmd := exec.CommandContext(context.Background(),
		"git", "worktree", "remove", "--force", worktreeDir,
	)
	cmd.Dir = repoDir

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w\n%s", err, out)
	}
	return nil
}

func pruneWorktrees(repoDir string) error {
	cmd := exec.CommandContext(context.Background(),
		"git", "worktree", "prune",
	)
	cmd.Dir = repoDir

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w\n%s", err, out)
	}
	return nil
}
