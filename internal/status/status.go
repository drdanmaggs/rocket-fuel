// Package status provides a summary of the current Rocket Fuel state.
package status

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/drdanmaggs/rocket-fuel/internal/tmux"
)

// Summary holds the current state of Rocket Fuel.
type Summary struct {
	SessionActive bool
	Workers       []WorkerStatus
	RepoDir       string
}

// WorkerStatus describes a single worker's state.
type WorkerStatus struct {
	Name       string
	WindowOpen bool
	Branch     string
	HasPR      bool
}

// Gather collects the current state from tmux, git, and GitHub.
func Gather(tm tmux.Runner, sessionName, repoDir string) (*Summary, error) {
	s := &Summary{
		SessionActive: tm.HasSession(sessionName),
		RepoDir:       repoDir,
	}

	worktreesDir := filepath.Join(repoDir, ".worktrees")
	entries, err := os.ReadDir(worktreesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("read worktrees: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		worktreeDir := filepath.Join(worktreesDir, name)

		// Worker windows use "#N: title" format. Extract issue number
		// from worktree dir name (worker-N) to match.
		issueNum := strings.TrimPrefix(name, "worker-")
		windowPrefix := "#" + issueNum + ":"
		windowOpen := s.SessionActive && (tm.HasWindow(sessionName, name) || hasWindowWithPrefix(tm, sessionName, windowPrefix))

		ws := WorkerStatus{
			Name:       name,
			WindowOpen: windowOpen,
			Branch:     worktreeBranch(worktreeDir),
		}

		if ws.Branch != "" {
			ws.HasPR = branchHasPR(ws.Branch)
		}

		s.Workers = append(s.Workers, ws)
	}

	return s, nil
}

// Format renders the summary as a human-readable string.
func Format(s *Summary) string {
	var b strings.Builder

	_, _ = fmt.Fprintln(&b, "=== Rocket Fuel Status ===")
	_, _ = fmt.Fprintln(&b)

	if s.SessionActive {
		_, _ = fmt.Fprintln(&b, "Session: ACTIVE")
	} else {
		_, _ = fmt.Fprintln(&b, "Session: INACTIVE")
	}

	_, _ = fmt.Fprintln(&b)

	if len(s.Workers) == 0 {
		_, _ = fmt.Fprintln(&b, "Workers: none")
	} else {
		_, _ = fmt.Fprintf(&b, "Workers: %d\n", len(s.Workers))
		for _, w := range s.Workers {
			state := "done"
			if w.WindowOpen {
				state = "active"
			}

			pr := ""
			if w.HasPR {
				pr = " [PR open]"
			}

			_, _ = fmt.Fprintf(&b, "  %s  %-10s  %s%s\n", w.Name, "("+state+")", w.Branch, pr)
		}
	}

	return b.String()
}

func hasWindowWithPrefix(tm tmux.Runner, session, prefix string) bool {
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

func worktreeBranch(dir string) string {
	cmd := exec.CommandContext(context.Background(), "git", "branch", "--show-current")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func branchHasPR(branch string) bool {
	cmd := exec.CommandContext(context.Background(),
		"gh", "pr", "list", "--head", branch, "--json", "number", "--limit", "1",
	)
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	var prs []json.RawMessage
	if err := json.Unmarshal(out, &prs); err != nil {
		return false
	}
	return len(prs) > 0
}
