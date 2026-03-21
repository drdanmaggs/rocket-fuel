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

// Reap finds completed workers and cleans up their worktrees and tmux windows.
// A worker is considered complete when its tmux window no longer exists
// in the main session (Claude Code session ended).
// If dryRun is true, reports what would be reaped without deleting.
func Reap(tm tmux.Runner, sessionName, repoDir string, dryRun ...bool) ([]ReapResult, error) {
	isDryRun := len(dryRun) > 0 && dryRun[0]
	_ = isDryRun // used below
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

		// Check if a window for this worker exists. Window names use "#N: title"
		// format, so match by issue number prefix.
		issueNum := strings.TrimPrefix(name, "worker-")
		windowPrefix := "#" + issueNum + ":"
		windowExists := tm.HasSession(sessionName) && hasWorkerWindow(tm, sessionName, windowPrefix)

		if windowExists {
			results = append(results, ReapResult{
				WindowName:  name,
				WorktreeDir: worktreeDir,
				Reaped:      false,
				Reason:      "window still active",
			})
			continue
		}

		if isDryRun {
			results = append(results, ReapResult{
				WindowName:  name,
				WorktreeDir: worktreeDir,
				Reaped:      false,
				Reason:      "would reap (dry-run)",
			})
			continue
		}

		// Window is gone — clean up the worktree.
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

	if !isDryRun {
		_ = pruneWorktrees(repoDir)
	}

	return results, nil
}

// hasWorkerWindow checks if any window in the session starts with the given prefix.
func hasWorkerWindow(tm tmux.Runner, session, prefix string) bool {
	names, err := tm.ListWindowNames(session)
	if err != nil {
		return false
	}
	for _, name := range names {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
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
